package version

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

const updateLockFile = "/tmp/ocsession-update.lock"

func GetVersion() string {
	return Version
}

func GetFullVersion() string {
	return fmt.Sprintf("ocsession %s (commit: %s, built: %s)",
		Version, GitCommit, BuildDate)
}

func GetUserAgent() string {
	return fmt.Sprintf("ocsession/%s (%s/%s)",
		Version, runtime.GOOS, runtime.GOARCH)
}

func GetPlatform() string {
	os := runtime.GOOS
	if os == "darwin" {
		os = "macos"
	}
	return fmt.Sprintf("%s-%s", os, runtime.GOARCH)
}

func GetCurrentBinaryPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取二进制路径失败: %w", err)
	}
	return filepath.EvalSymlinks(path)
}

func GetInstallDir() (string, error) {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(binaryPath), nil
}

func fetchLatestRelease() (*ReleaseInfo, error) {
	url := "https://api.github.com/repos/zhaiyz/ocsession/releases/latest"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", GetUserAgent())

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回状态码: %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &release, nil
}

func getAssetURL(release *ReleaseInfo, filename string) string {
	for _, asset := range release.Assets {
		if asset.Name == filename {
			return asset.DownloadURL
		}
	}
	return ""
}

func CheckUpdate() (currentVersion, latestVersion, releaseURL string, err error) {
	currentVersion = Version

	release, err := fetchLatestRelease()
	if err != nil {
		return currentVersion, "", "", err
	}

	latestVersion = release.TagName
	releaseURL = release.HTMLURL

	return currentVersion, latestVersion, releaseURL, nil
}

func NeedsUpdate() (bool, string, error) {
	current, latest, _, err := CheckUpdate()
	if err != nil {
		return false, "", err
	}

	if current == "dev" {
		return false, latest, nil
	}

	if current == latest {
		return false, latest, nil
	}

	return true, latest, nil
}

func CanUpdate() bool {
	installDir, err := GetInstallDir()
	if err != nil {
		return false
	}

	testFile := filepath.Join(installDir, ".write-test-"+fmt.Sprintf("%d", time.Now().UnixNano()))
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)

	return true
}

func acquireUpdateLock() (*os.File, error) {
	lockFile, err := os.OpenFile(updateLockFile,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("另一个更新进程正在运行")
		}
		return nil, err
	}

	lockFile.WriteString(fmt.Sprintf("%d", os.Getpid()))

	return lockFile, nil
}

func releaseUpdateLock(lockFile *os.File) {
	if lockFile != nil {
		lockFile.Close()
		os.Remove(updateLockFile)
	}
}

func downloadFile(url, destPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", GetUserAgent())

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func verifySHA256(filePath, expectedHash string) error {
	actualHash, err := calculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("计算 SHA256 失败: %w", err)
	}

	expectedHash = strings.TrimSpace(expectedHash)
	parts := strings.Fields(expectedHash)
	if len(parts) >= 1 {
		expectedHash = parts[0]
	}

	if actualHash != expectedHash {
		return fmt.Errorf("SHA256 不匹配: 期望 %s, 实际 %s", expectedHash, actualHash)
	}

	return nil
}

func BackupCurrentVersion() (string, error) {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return "", err
	}

	backupPath := binaryPath + ".backup"

	err = copyFile(binaryPath, backupPath)
	if err != nil {
		return "", fmt.Errorf("备份失败: %w", err)
	}

	return backupPath, nil
}

func RestoreBackup() error {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return err
	}

	backupPath := binaryPath + ".backup"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在")
	}

	err = copyFile(backupPath, binaryPath)
	if err != nil {
		return fmt.Errorf("恢复失败: %w", err)
	}

	os.Chmod(binaryPath, 0755)

	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		os.Chmod(dst, os.FileMode(stat.Mode))
	} else {
		os.Chmod(dst, info.Mode())
	}

	return nil
}

func SelfUpdate() error {
	lockFile, err := acquireUpdateLock()
	if err != nil {
		return err
	}
	defer releaseUpdateLock(lockFile)

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("获取版本信息失败: %w", err)
	}

	platform := GetPlatform()
	tarFilename := fmt.Sprintf("ocsession-%s.tar.gz", platform)
	shaFilename := fmt.Sprintf("ocsession-%s.sha256", platform)

	tarURL := getAssetURL(release, tarFilename)
	shaURL := getAssetURL(release, shaFilename)

	if tarURL == "" {
		return fmt.Errorf("未找到 %s 的下载链接", tarFilename)
	}

	fmt.Printf("下载: %s\n", tarURL)

	tmpDir, err := os.MkdirTemp("", "ocsession-update-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, tarFilename)
	shaPath := filepath.Join(tmpDir, shaFilename)
	extractedBinary := filepath.Join(tmpDir, "ocsession")

	if err := downloadFile(tarURL, tarPath); err != nil {
		return fmt.Errorf("下载二进制失败: %w", err)
	}

	fmt.Println("验证 SHA256...")
	if err := downloadFile(shaURL, shaPath); err != nil {
		fmt.Println("警告: 无法下载 SHA256 校验文件")
	} else {
		shaContent, err := os.ReadFile(shaPath)
		if err != nil {
			return fmt.Errorf("读取 SHA256 文件失败: %w", err)
		}

		if err := verifySHA256(tarPath, string(shaContent)); err != nil {
			return fmt.Errorf("校验失败: %w", err)
		}
	}

	fmt.Println("解压...")
	cmd := exec.Command("tar", "-xzf", tarPath, "-C", tmpDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	fmt.Println("备份当前版本...")
	backupPath, err := BackupCurrentVersion()
	if err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}
	fmt.Printf("备份位置: %s\n", backupPath)

	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return err
	}

	fmt.Println("替换二进制...")
	if err := copyFile(extractedBinary, binaryPath); err != nil {
		fmt.Println("替换失败，尝试恢复备份...")
		RestoreBackup()
		return fmt.Errorf("替换失败: %w", err)
	}

	os.Chmod(binaryPath, 0755)

	// macOS: 立即清除 quarantine 属性（在签名和验证之前）
	if runtime.GOOS == "darwin" {
		fmt.Println("清除 quarantine 属性...")
		// 使用 -cr 清除所有扩展属性（更彻底）
		if err := exec.Command("xattr", "-cr", binaryPath).Run(); err != nil {
			fmt.Printf("警告: 清除 quarantine 属性失败: %v\n", err)
		}
	}

	// macOS: 处理代码签名
	if runtime.GOOS == "darwin" {
		fmt.Println("处理代码签名...")
		// 移除可能存在的无效签名
		exec.Command("codesign", "--remove-signature", binaryPath).Run()
		// 添加 ad-hoc 签名
		if err := exec.Command("codesign", "--force", "--sign", "-", binaryPath).Run(); err != nil {
			fmt.Printf("警告: 代码签名失败: %v\n", err)
		}
	}

	fmt.Println("验证新版本...")
	cmd = exec.Command(binaryPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		// 添加详细错误信息
		fmt.Printf("命令执行失败: %v\n", err)
		if len(output) > 0 {
			fmt.Printf("输出: %s\n", string(output))
		}
		fmt.Println("验证失败，尝试恢复备份...")
		RestoreBackup()
		return fmt.Errorf("验证失败: %w", err)
	}

	fmt.Printf("\n新版本信息:\n%s\n", strings.TrimSpace(string(output)))

	return nil
}
