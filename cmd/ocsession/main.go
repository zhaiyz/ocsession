package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/zhaiyz/ocsession/internal/agent"
	"github.com/zhaiyz/ocsession/internal/config"
	"github.com/zhaiyz/ocsession/internal/service"
	"github.com/zhaiyz/ocsession/internal/store"
	"github.com/zhaiyz/ocsession/internal/tui"
	"github.com/zhaiyz/ocsession/internal/version"
)

var (
	showVersion = flag.Bool("v", false, "显示版本信息")
	showHelp    = flag.Bool("h", false, "显示帮助信息")
	autoConfirm = flag.Bool("y", false, "自动确认更新（用于 update 命令）")
)

func main() {
	flag.BoolVar(showVersion, "version", false, "显示版本信息")
	flag.BoolVar(showHelp, "help", false, "显示帮助信息")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.GetFullVersion())
		os.Exit(0)
	}

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "update":
			handleUpdate()
			os.Exit(0)
		case "version":
			fmt.Println(version.GetFullVersion())
			os.Exit(0)
		case "help":
			printHelp()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
			printHelp()
			os.Exit(1)
		}
	}

	agentCfgPath := getAgentConfigPath()

	agentCfg, err := loadOrDetectAgentConfig(agentCfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	cfgPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadOrCreateConfig(cfgPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	dbPath := agentCfg.GetDBPath()
	sqliteStore, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer sqliteStore.Close()

	sessionSvc := service.NewSessionService(sqliteStore, cfg)
	err = sessionSvc.LoadSessions()
	if err != nil {
		fmt.Printf("Error loading sessions: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(sessionSvc, agentCfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(tui.Model); ok {
		if m.SessionToStart != nil {
			if err := tui.RunAgent(m.SessionToStart, agentCfg); err != nil {
				fmt.Fprintf(os.Stderr, "\n错误: 无法启动会话: %v\n", err)
				fmt.Fprintf(os.Stderr, "请确保 %s 已正确安装\n", agentCfg.GetCommand())
				os.Exit(1)
			}
		}
	}
}

func getAgentConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ocsession", "agent.toml")
}

func loadOrDetectAgentConfig(path string) (*agent.AgentConfig, error) {
	cfg, err := agent.LoadAgentConfig(path)
	if err == nil {
		if err := agent.ValidateAgent(cfg); err == nil {
			return cfg, nil
		}
		fmt.Printf("配置无效: %v\n", err)
	}

	return detectOrPromptAgentConfig(path)
}

func detectOrPromptAgentConfig(path string) (*agent.AgentConfig, error) {
	defaultCfg := &agent.AgentConfig{AgentName: "opencode"}

	dbPath := agent.DetectAgentDBPath("opencode")
	if dbPath != "" {
		defaultCfg.DBPath = dbPath
	}

	if err := agent.ValidateAgent(defaultCfg); err == nil {
		fmt.Printf("使用默认配置: opencode\n")
		if err := agent.SaveAgentConfig(path, defaultCfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}
		return defaultCfg, nil
	}

	fmt.Println("未找到有效配置，请输入 AI Agent 名称（如 opencode、codewiz）:")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("读取输入失败: %w", err)
		}

		agentName := strings.TrimSpace(input)
		if agentName == "" {
			fmt.Println("Agent 名称不能为空，请重新输入")
			continue
		}

		cfg := &agent.AgentConfig{AgentName: agentName}

		dbPath := agent.DetectAgentDBPath(agentName)
		if dbPath == "" {
			fmt.Printf("未找到 %s 的数据库，请输入数据库完整路径:\n", agentName)
			fmt.Printf("> ")
			dbInput, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("读取输入失败: %w", err)
			}
			cfg.DBPath = strings.TrimSpace(dbInput)
		} else {
			cfg.DBPath = dbPath
		}

		if err := agent.ValidateAgent(cfg); err != nil {
			fmt.Printf("验证失败: %v\n请重新输入 Agent 名称:\n", err)
			continue
		}

		if err := agent.SaveAgentConfig(path, cfg); err != nil {
			return nil, fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Printf("配置已保存，使用 %s\n", agentName)
		return cfg, nil
	}
}

func printHelp() {
	fmt.Println(`ocsession - OpenCode Session Manager

用法:
  ocsession              启动会话管理界面
  ocsession update       检查并更新到最新版本
  ocsession update -y    自动确认更新（无需手动确认）
  ocsession version      显示版本信息
  ocsession -v           显示版本信息
  ocsession --version    显示版本信息
  ocsession -h           显示帮助信息
  ocsession --help       显示帮助信息

快捷键（在 TUI 中）:
  ↑/k     向上移动
  ↓/j     向下移动
  /       搜索模式
  Enter   继续会话
  r       刷新列表
  q       退出

更多信息: https://github.com/zhaiyz/ocsession`)
}

func handleUpdate() {
	fmt.Println("检查更新...")

	current, latest, releaseURL, err := version.CheckUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "检查更新失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("当前版本: %s\n", current)
	fmt.Printf("最新版本: %s\n", latest)

	if current == "dev" {
		fmt.Println("\n您正在运行开发版本，无法自动更新。")
		fmt.Println("请使用以下命令重新安装:")
		fmt.Println("  make install")
		fmt.Println("\n或:")
		fmt.Println("  curl -sSL https://raw.githubusercontent.com/zhaiyz/ocsession/main/install.sh | bash")
		os.Exit(0)
	}

	if current == latest {
		fmt.Println("\n✓ 已是最新版本!")
		os.Exit(0)
	}

	fmt.Println("\n发现新版本!")
	fmt.Printf("发布页面: %s\n", releaseURL)

	if !version.CanUpdate() {
		fmt.Println("\n✗ 没有更新权限")
		fmt.Println("\n当前二进制安装在无写权限的目录。")
		fmt.Println("请使用以下命令手动更新:")
		fmt.Println("  curl -sSL https://raw.githubusercontent.com/zhaiyz/ocsession/main/install.sh | bash")
		os.Exit(1)
	}

	if !*autoConfirm {
		fmt.Print("\n是否更新？[y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		if strings.TrimSpace(strings.ToLower(input)) != "y" {
			fmt.Println("取消更新")
			os.Exit(0)
		}
	}

	fmt.Println("\n开始更新...")

	if err := version.SelfUpdate(); err != nil {
		fmt.Fprintf(os.Stderr, "\n✗ 更新失败: %v\n", err)
		fmt.Println("\n正在尝试恢复备份...")

		if err := version.RestoreBackup(); err != nil {
			fmt.Fprintf(os.Stderr, "✗ 恢复备份失败: %v\n", err)
			fmt.Println("\n请手动重新安装:")
			fmt.Println("  curl -sSL https://raw.githubusercontent.com/zhaiyz/ocsession/main/install.sh | bash")
		} else {
			fmt.Println("✓ 已恢复备份版本")
			fmt.Println("\n请稍后重试或手动下载更新")
		}

		os.Exit(1)
	}

	fmt.Println("\n✓ 更新成功!")
	fmt.Printf("新版本: %s\n", latest)
	fmt.Println("\n下次运行 'ocsession' 时将使用新版本")
}
