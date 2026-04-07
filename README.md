# OpenCode Session Manager

一个终端界面(TUI)工具，帮助快速管理和切换OpenCode会话。

## 功能特性

- **会话列表浏览** - 查看所有主会话（自动过滤 sub-agent 会话）
- **文件夹名显示** - 会话列表显示文件夹名，快速识别项目
- **实时模糊搜索** - 快速查找会话（支持中文标题）
- **会话详情预览** - 显示完整信息（标题、ID、路径、时长、对话记录等）
- **对话记录显示** - 显示用户输入的前5条和后5条消息
- **智能时间显示** - 统一中文格式（今天/周一/2026-04-06）

- **一键启动** - 按 Enter 直接切换到 OpenCode 会话

## 特性亮点

✅ **界面美观** - 固定布局，边框稳定，不受选中行影响  
✅ **时间准确** - 正确处理毫秒时间戳，显示清晰  
✅ **详情丰富** - 显示项目路径、会话时长、对话记录等多维度信息  
✅ **启动可靠** - OpenCode 直接接管终端，无缝切换  
✅ **中文字符支持** - 正确处理中文宽字符，显示无乱码  
✅ **实时搜索** - 输入即搜索，支持中英文，即时过滤  
✅ **文件夹名** - 会话列表显示文件夹名，快速识别项目

## 安装

### 方式1: 一键安装（推荐）

macOS / Linux:

```bash
curl -sSL https://raw.githubusercontent.com/zhaiyz/ocsession/main/install.sh | bash
```

自定义安装路径:

```bash
INSTALL_DIR=~/.local/bin curl -sSL https://raw.githubusercontent.com/zhaiyz/ocsession/main/install.sh | bash
```

安装特定版本:

```bash
VERSION=v1.0.0 curl -sSL https://raw.githubusercontent.com/zhaiyz/ocsession/main/install.sh | bash
```

### 方式2: 手动下载

从 [Releases](https://github.com/zhaiyz/ocsession/releases) 下载对应平台的二进制：

| 平台 | 架构 | 下载 |
|------|------|------|
| macOS | Apple Silicon (M1/M2/M3) | ocsession-macos-arm64.tar.gz |
| macOS | Intel | ocsession-macos-amd64.tar.gz |
| Linux | x86_64 | ocsession-linux-amd64.tar.gz |
| Linux | arm64 | ocsession-linux-arm64.tar.gz |

```bash
# 解压
tar -xzf ocsession-*.tar.gz

# 安装到用户目录（推荐）
mkdir -p ~/.local/bin
mv ocsession ~/.local/bin/
chmod +x ~/.local/bin/ocsession

# 添加到 PATH（如果 ~/.local/bin 不在 PATH 中）
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc  # 或 ~/.bashrc
source ~/.zshrc

# 验证安装
ocsession -v
```

### 方式3: 从源码编译

```bash
git clone https://github.com/zhaiyz/ocsession
cd ocsession
make install
```

**依赖**: Go 1.21+, GCC (用于 SQLite CGO)

## 使用

```bash
# 启动会话管理
ocsession

# 查看版本
ocsession -v
ocsession --version
ocsession version

# 查看帮助
ocsession -h
ocsession --help
ocsession help

# 检查并更新
ocsession update         # 会提示确认
ocsession update -y      # 自动确认更新
```

## 更新

### 自动更新（推荐）

```bash
ocsession update
```

功能：
- 自动检查 GitHub Releases 最新版本
- 下载并验证 SHA256
- 备份当前版本
- 自动替换二进制文件
- 更新失败自动恢复备份

### 手动更新

```bash
curl -sSL https://raw.githubusercontent.com/zhaiyz/ocsession/main/install.sh | bash
```

或从 [Releases](https://github.com/zhaiyz/ocsession/releases) 手动下载。

## 命令行参数

| 命令 | 说明 |
|------|------|
| `ocsession` | 启动 TUI 会话管理 |
| `ocsession -v` | 显示版本信息 |
| `ocsession -h` | 显示帮助信息 |
| `ocsession update` | 检查并更新到最新版本 |
| `ocsession update -y` | 自动确认更新（跳过确认提示） |

## ⌨️ 键盘快捷键

| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `↑` 或 `k` | 向上移动 | 选择上一个会话 |
| `↓` 或 `j` | 向下移动 | 选择下一个会话 |
| `/` | 搜索模式 | 进入搜索模式 |
| `Enter` | 继续会话 | 切换到 OpenCode 继续选中的会话 |
| `r` | 刷新列表 | 重新加载会话列表 |
| `q` 或 `Ctrl+C` | 退出 | 退出程序 |
| `Esc` | 取消搜索 | 退出搜索模式 |

**💡 提示**：
- 会话列表自动过滤 sub-agent 会话，只显示主会话
- 时间列对齐显示，便于快速浏览
- 中文标题正确显示，无乱码
- 按 Enter 后直接切换到 OpenCode 会话

## 🎨 界面布局

```
┌──────────────────────────────────────────────────────────────┐
│  OpenCode Session Manager    [快捷键提示]                    │
├──────────────────────────────────┬───────────────────────────┤
│ 会话列表                         │ 会话详情                  │
│                                  │                           │
│ → 文件夹名    会话标题      时间 │ 标题: xxx                 │
│   ocsession   OpenCode...   今天 │ ID: ses_xxx               │
│   voice-input 查询可用...   今天 │ 路径: /Users/xxx/code     │
│   code        查看并分...   周日 │ 更新: 今天 00:43          │
│                                  │ 创建: 2026-04-05          │
│                                  │ 时长: 2小时30分           │
│                                  │ 消息: 288 条              │
│                                  │                           │
│                                  │ ─ 对话记录 ─              │
│                                  │ 1. 创建session管理程序    │
│                                  │ 2. 使用中文               │
│                                  │ ... 省略 15 条消息 ...    │
│                                  │ 21. 会话详情中的对话记录  │
│                                  │ 22. 显示最开始的5条       │
├──────────────────────────────────┴───────────────────────────┤
└──────────────────────────────────────────────────────────────┘
```

## 📝 功能说明

### 1. 会话列表浏览

- 启动后自动显示最近更新的会话
- **自动过滤 sub-agent 会话**：只显示主会话，不显示 sub-agent 会话
- **显示文件夹名**：从路径自动提取文件夹名，快速识别项目
- 显示会话标题和更新时间
- 使用 `j/k` 或 `↑/↓` 导航
- 选中的会话会有 `→` 标记和高亮显示

**关于 Sub-agent 过滤**：
- 默认隐藏所有 sub-agent 会话（有 parent_id 的会话）
- 只显示您手动创建的主会话
- 搜索时也自动过滤 sub-agent 会话

**关于文件夹名**：
- 自动从会话路径提取文件夹名
- 显示在会话列表第一列（15字符宽）
- 快速识别项目，无需查看完整路径

### 2. 会话搜索

- 按 `/` 进入搜索模式
- 输入关键词进行模糊搜索
- 按 `Enter` 确认搜索
- 按 `Esc` 取消搜索

### 3. 查看会话详情

- 右侧面板显示选中会话的详细信息
- 包含：标题、ID、完整路径、时间、时长、消息数
- **对话记录**：显示用户输入的前5条和后5条消息
  - 少于等于10条：全部显示
  - 超过10条：前5条 + 省略提示 + 后5条
  - 编号左对齐，格式统一

### 4. 继续会话

- 选中会话后按 `Enter`
- 自动调用 `opencode -s <session-id>`
- 切换到 OpenCode 继续开发

### 5. 刷新列表

- 按 `r` 重新加载会话列表
- 获取最新的会话数据

## 🔧 配置文件

配置文件位于 `~/.config/ocsession/config.toml`

配置文件示例：

```toml
[general]
default_sort = "updated"      # 默认排序方式
preview_lines = 10            # 预览显示行数
max_sessions_display = 50     # 最大显示会话数
```

## 📊 时间显示

应用会智能显示时间格式：

- **今天**: 显示为 `今天 15:04`
- **本周**: 显示为 `周一 15:04`
- **更早**: 显示为 `2026-01-02`

## 🎯 使用场景

### 场景1: 快速切换项目

1. 启动 `ocsession`
2. 使用 `j/k` 浏览会话
3. 找到目标会话按 `Enter` 继续

### 场景2: 查找历史会话

1. 按 `/` 进入搜索模式
2. 输入项目名或关键词
3. 选择会话按 `Enter` 继续



## 开发

### 构建

```bash
make build    # 编译项目
make test     # 运行测试
make run      # 直接运行
make clean    # 清理编译产物
```

### 项目结构

```
ocsession/
├── cmd/ocsession/main.go      # 主程序入口
├── internal/
│   ├── config/                # 配置管理
│   ├── store/                 # 数据访问层
│   ├── service/               # 业务逻辑层
│   ├── tui/                   # TUI界面
│   └── fuzzy/                 # 模糊搜索
├── config/                    # 配置文件示例
└── test/                      # 测试文件
```

## 依赖

- Go 1.21+
- OpenCode (已安装并配置)

## 技术栈

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI框架
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 样式库
- [SQLite](https://github.com/mattn/go-sqlite3) - 数据库访问
- [go-toml](https://github.com/pelletier/go-toml) - TOML解析

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
