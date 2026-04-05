# OpenCode Session Manager

一个终端界面(TUI)工具，帮助快速管理和切换OpenCode会话。

## 功能特性

- **会话列表浏览** - 查看所有主会话（自动过滤 sub-agent 会话）
- **实时模糊搜索** - 快速查找会话（支持中文标题）
- **会话预览** - 显示会话详情（ID、目录、时间、标签）
- **智能时间显示** - 统一英文格式（今天/本周/更早）
- **标签管理** - 通过配置文件添加标签分类
- **别名管理** - 设置快速访问别名
- **一键启动** - 按 Enter 直接切换到 OpenCode 会话

## 特性亮点

✅ **界面美观** - 固定布局，边框稳定，不受选中行影响  
✅ **时间准确** - 正确处理毫秒时间戳，显示清晰  
✅ **启动可靠** - OpenCode 直接接管终端，无缝切换  
✅ **中文字符支持** - 正确处理中文宽字符，显示无乱码

## 安装

### 方式1: 从源码编译

```bash
git clone https://github.com/opencode-session-manager/ocsession
cd ocsession
make install
```

### 方式2: 直接编译

```bash
cd ocsession
make build
sudo mv bin/ocsession /usr/local/bin/
```

## 使用

```bash
ocsession
```

## 快捷键

| 快捷键 | 功能 |
|--------|------|
| `/` | 进入搜索模式 |
| `Enter` | 继续选中的会话 |
| `↑/↓` 或 `j/k` | 在列表中导航 |
| `q` | 退出程序 |
| `Esc` | 取消当前操作 |

## 配置

配置文件位于 `~/.config/ocsession/config.toml`

### 配置示例

```toml
[general]
default_sort = "updated"      # 默认排序方式
preview_lines = 10            # 预览显示行数
max_sessions_display = 50     # 最大显示会话数

[aliases]
voice-input = "ses_2a725bdbbffeP9irDnInRMc2yQ"

[session_tags.ses_2a725bdbbffeP9irDnInRMc2yQ]
tags = ["voice-input", "active-project"]
notes = "语音输入功能开发"
```

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
