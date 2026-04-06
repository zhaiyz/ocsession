# OpenCode Session Manager 设计文档

**项目名称**: ocsession  
**设计日期**: 2026-04-05  
**版本**: v1.0.0

---

## 1. 概述

### 1.1 项目目标

创建一个终端界面(TUI)工具，帮助用户快速管理和切换OpenCode会话。通过模糊搜索，提升会话管理效率。

### 1.2 核心功能

- **会话列表浏览**: 显示所有OpenCode会话，包含标题、目录、时间信息
- **实时模糊搜索**: 快速查找会话（搜索标题、目录）
- **会话预览**: 显示最后消息、文件变更、token统计
- **继续会话**: 快速切换到OpenCode继续开发

### 1.3 技术栈

- **语言**: Go 1.21+
- **TUI框架**: Bubbletea (charmbracelet/bubbletea)
- **样式库**: Lipgloss (charmbracelet/lipgloss)
- **配置格式**: TOML (pelletier/go-toml)
- **数据库**: SQLite (mattn/go-sqlite3)
- **模糊搜索**: Sahilm/fuzzy

---

## 2. 系统架构

### 2.1 架构层次

系统采用三层架构设计：

```
┌─────────────────────────────────────┐
│      TUI界面层 (Bubbletea)          │
│  - 主应用循环                        │
│  - 组件渲染                          │
│  - 用户交互处理                      │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│      业务逻辑层 (Service)           │
│  - SessionService                    │
│  - SearchEngine                      │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│      数据访问层 (Store)             │
│  - SQLiteStore (主数据源)            │
│  - CLIStore (备用数据源)             │
│  - ConfigStore (配置管理)            │
└─────────────────────────────────────┘
```

### 2.2 数据流向

**启动流程**:
```
main() → 加载配置 → 连接数据库 → 加载会话列表 → 构建搜索索引 → 显示TUI
```

**搜索流程**:
```
用户输入 → SearchEngine.Search() → 过滤会话列表 → 更新界面显示
```

**继续会话流程**:
```
选择会话 → 执行 opencode -s <session-id> → 切换到OpenCode TUI
```

---

## 3. 核心模块设计

### 3.1 SessionService (会话服务)

**职责**:
- 加载会话列表（从SQLite或CLI）
- 会话排序（按更新时间、创建时间、访问频率）
- 会话过滤（按项目、日期范围）
- 会话搜索（模糊匹配）

**核心方法**:
```go
LoadSessions() []Session
FilterSessions(criteria FilterCriteria) []Session
SearchSessions(query string) []Session
GetSessionDetail(id string) SessionDetail
```

**性能优化**:
- 启动时批量加载，内存缓存
- 搜索使用倒排索引
- 增量更新检测（每分钟检查updated时间戳）

### 3.2 SearchEngine (搜索引擎)

**职责**:
- 模糊搜索实现
- 多字段搜索（标题、目录）
- 搜索结果排序（相关性、时间）
- 搜索历史记录

**核心方法**:
```go
Search(query string, sessions []Session) []SearchResult
BuildIndex(sessions []Session)
HighlightMatch(text string, query string) string
```

**算法**:
- Fuzzy matching（类似fzf）
- 支持前缀匹配、子串匹配、正则匹配
- 搜索结果按匹配度评分排序

---

## 4. TUI界面设计

### 4.1 界面布局

```
┌─────────────────────────────────────────────────────────────────┐
│ OpenCode Session Manager                    [h:帮助] [q:退出]   │
├──────────────────────┬──────────────────────────────────────────┤
│ 会话列表 (12/45)     │ 会话预览                                │
│                      │                                          │
│ 🔍 search: voice     │ 📌 voice-input项目开发                   │
│                      │ 会话ID: ses_2a725bdbbffeP9irDnInRMc2yQ   │
│ ✅ voice-input       │                                          │
│   📁 voice-input     │ 📋 最后消息 (10行)                       │
│                      │                                          │
│                      │                                          │
│ ○ empty-main         │                                          │
│   📁 empty           │                                          │
│                      │                                          │
│                      │                                          │
│                      │                                          │
│ ○ todo-list-dev      │                                          │
│   📁 todo-list       │                                          │
│                      │                                          │
│                      │                                          │
│ ○ old-session        │                                          │
│   📁 empty           │                                          │
│   ⏰ 30天前          │                                          │
│                      │                                          │
├──────────────────────┴──────────────────────────────────────────┤
│ [Enter:继续] [d:删除] [/:搜索] [r:刷新]                          │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 核心交互流程

**启动与列表显示**:
- 启动 `ocsession` → 加载会话列表 → 默认显示最近更新的20个会话
- 列表项显示：状态图标(✅活跃/○非活跃)、标题、目录图标📁、时间⏰

**搜索与过滤**:
- 按 `/` 键进入搜索模式 → 输入关键词 → 实时模糊匹配
- 搜索范围：标题、目录
- 支持正则表达式（高级模式）

**会话选择与预览**:
- 上/下键导航 → Enter键选择 → 右侧显示详细预览
- 预览内容：最后消息、文件变更列表、token统计

**继续会话**:
- Enter键 → 执行 `opencode -s <session-id>` → 自动切换到OpenCode TUI
- 或按 `o` 键 → 在新终端窗口打开会话（保留ocsession窗口）

### 4.3 键盘快捷键

| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `/` | 搜索模式 | 进入搜索输入框 |
| `Esc` | 取消操作 | 退出搜索/管理界面 |
| `Enter` | 继续会话 | 切换到OpenCode会话 |
| `d` | 删除会话 | 删除选中会话（需确认） |
| `r` | 刷新列表 | 重新加载会话数据 |
| `h` | 帮助 | 显示快捷键帮助 |
| `q` | 退出 | 退出ocsession |
| `↑/↓` | 导航 | 在列表中上下移动 |
| `j/k` | 导航 | Vim风格上下移动 |
| `Ctrl+f` | 下一页 | 列表向下翻页 |
| `Ctrl+b` | 上一页 | 列表向上翻页 |

---

## 5. 数据访问层设计

### 5.1 数据源优先级

1. **SQLite数据库** (主数据源)
   - 路径: `~/.local/share/opencode/opencode.db`
   - 表结构: sessions, messages, tools 等
   - 优势: 性能最优，数据完整
   
2. **OpenCode CLI** (备用数据源)
   - 命令: `opencode session list`, `opencode export`
   - 优势: 不依赖数据库结构，兼容性好
   - 降级场景: SQLite访问失败时自动切换

### 5.2 SQLite查询策略

**启动加载**:
```sql
SELECT id, title, updated, created, project_id, directory 
FROM sessions 
ORDER BY updated DESC 
LIMIT 50;
```

**预览查询**:
```sql
SELECT content, timestamp 
FROM messages 
WHERE session_id = ? 
ORDER BY timestamp DESC 
LIMIT 10;
```

**性能优化**:
- 使用索引加速查询（updated字段索引）
- 批量查询避免多次数据库访问
- 结果缓存减少重复查询

### 5.3 配置文件结构

```toml
[general]
default_sort = "updated"
preview_lines = 10
max_sessions_display = 50
theme = "default"
```

### 5.4 缓存数据

**搜索索引缓存**:
- 路径: `~/.cache/ocsession/search_index.json`
- 内容: 倒排索引结构（关键词→会话ID映射）
- 更新策略: 每小时重建或会话变更时重建

**预览缓存**:
- 路径: `~/.cache/ocsession/previews/`
- 内容: 会话预览数据（按会话ID分文件）
- 过期策略: 24小时过期

---

## 6. 项目目录结构

```
ocsession/
├── cmd/
│   └── ocsession/
│       └── main.go              # 主程序入口
│
├── internal/
│   ├── config/
│   │   ├── config.go            # 配置管理结构体定义
│   │   ├── loader.go            # 配置文件加载器（TOML解析）
│   │   └── saver.go             # 配置文件保存器
│   │   └── defaults.go          # 默认配置值
│   │
│   ├── store/
│   │   ├── interface.go         # Store接口定义
│   │   ├── sqlite.go            # SQLite数据库访问实现
│   │   ├── cli.go               # OpenCode CLI备用实现
│   │   ├── session.go           # Session数据结构
│   │   └── message.go           # Message数据结构
│   │
│   ├── service/
│   │   ├── session_service.go   # 会话服务
│   │   └── search_engine.go     # 搜索引擎
│   │
│   ├── tui/
│   │   ├── app.go               # Bubbletea主应用
│   │   ├── components/
│   │   │   ├── list.go          # 会话列表组件
│   │   │   ├── preview.go       # 预览面板组件
│   │   │   └── search.go        # 搜索输入组件
│   │   ├── styles/
│   │   │   └── theme.go         # Lipgloss样式定义
│   │   └── keybinds.go          # 键盘快捷键处理
│   │
│   ├── fuzzy/
│   │   ├── matcher.go           # 模糊匹配算法实现
│   │   ├── scorer.go            # 匹配度评分算法
│   │   └── ranker.go            # 结果排序算法
│   │
│   └── utils/
│       ├── time.go              # 时间格式化工具
│       ├── text.go              # 文本处理工具
│       ├── validator.go         # 输入验证工具
│       └── paths.go             # 路径处理工具
│
├── pkg/
│   ├── api/
│   │   └── client.go            # OpenCode API客户端（可选）
│   └── models/
│       ├── session.go           # Session公开模型
│       └── config.go            # Config公开模型
│
├── config/
│   └── config.example.toml      # 配置文件示例
│
├── scripts/
│   ├── install.sh               # 安装脚本
│   ├── build.sh                 # 编译脚本
│   └── test.sh                  # 测试脚本
│
├── test/
│   ├── integration/
│   │   ├── sqlite_test.go       # SQLite集成测试
│   │   └── cli_test.go          # CLI集成测试
│   └─ unit/
│       ├── service_test.go      # 服务单元测试
│       ├── fuzzy_test.go        # 模糊搜索测试
│       └── suggestion_test.go   # 建议生成测试
│
├── go.mod                       # Go模块依赖
├── go.sum                       # Go依赖校验
├── Makefile                     # Make构建工具
├── README.md                    # 项目文档
└── LICENSE                      # 许可证文件
```

---

## 7. 错误处理策略

### 7.1 数据访问错误

**SQLite数据库损坏或不存在**:
- 检测错误 → 自动降级到CLI备用方案
- 显示警告："⚠️ 数据库访问失败，已切换到CLI模式（性能较低）"
- 记录错误日志到 `~/.cache/ocsession/errors.log`

**OpenCode CLI不可用**:
- 检测 `opencode` 命令是否存在
- 显示错误："❌ OpenCode未安装，请先安装opencode"
- 程序退出（exit code 1）

### 7.2 配置文件错误

**TOML格式错误**:
- 显示警告："⚠️ 配置文件格式错误，已使用默认配置"
- 提示用户检查配置文件语法
- 提供配置修复命令：`ocsession config-fix`

**配置文件不存在**:
- 自动创建默认配置文件
- 显示提示："✓ 已创建默认配置文件"

### 7.4 日志管理

**日志级别**:
- DEBUG: 详细调试信息
- INFO: 正常操作日志
- WARN: 警告信息
- ERROR: 错误信息

**日志输出**:
- 文件：`~/.cache/ocsession/ocsession.log`
- 格式：JSON结构化日志
- 轮转：每日轮转，保留30天

---

## 8. 测试策略

### 8.1 单元测试

**测试重点**:
- fuzzy/matcher.go: 模糊匹配算法准确性
- utils/validator.go: 输入验证逻辑
- config/loader.go: TOML解析正确性

**测试方法**:
- 使用Go标准测试框架
- Mock数据依赖
- 边界条件测试

### 8.2 集成测试

**测试重点**:
- store/sqlite.go: 真实SQLite数据库交互
- service层: 多模块协作流程
- TUI交互: 用户输入→界面响应

**测试环境**:
- 使用测试数据库副本
- 模拟OpenCode CLI输出
- 使用虚假配置文件

### 8.3 性能测试

**测试指标**:
- 启动时间：<500ms（加载1000个会话）
- 搜索响应：<100ms（实时模糊搜索）
- 内存占用：<50MB（运行时）

**测试方法**:
- 模拟大量会话数据（1000+ sessions）
- 使用pprof进行性能分析
- 优化热点函数

### 8.4 兼容性测试

**测试环境**:
- macOS (10.15+)
- Linux (Ubuntu 20.04+, Arch Linux)

**终端兼容性**:
- WezTerm
- Alacritty
- iTerm2
- Terminal.app

**测试内容**:
- TUI渲染正确性（颜色、样式、中文支持）
- 键盘输入响应
- OpenCode数据库路径兼容性

---

## 9. 实现路线图

### 9.1 Phase 1: 核心基础（Week 1-2）

**任务**:
- 项目初始化（目录结构、go.mod、Makefile）
- 数据访问层实现（SQLite连接、Session/Message模型、CLI备用）
- 配置管理实现（TOML加载/保存、默认配置）
- 基础会话服务（加载会话列表、排序过滤）

**验收标准**:
- 能成功加载OpenCode会话列表
- 配置文件读写正常
- SQLite查询返回正确数据

### 9.2 Phase 2: TUI界面框架（Week 3-4）

**任务**:
- Bubbletea主应用框架（主循环、状态管理）
- 会话列表组件（渲染、导航、高亮）
- 预览面板组件（详细信息、消息展示）
- 基础交互实现（Enter继续会话、样式美化）

**验收标准**:
- TUI界面正常显示
- 键盘导航响应正确
- 能成功切换到OpenCode会话

### 9.3 Phase 3: 搜索与过滤（Week 5）

**任务**:
- 搜索引擎实现（模糊匹配、多字段搜索、评分）
- 搜索组件集成（输入框、实时过滤、结果排序）
- 过滤增强（按项目、日期过滤）

**验收标准**:
- 搜索响应速度<100ms
- 模糊匹配准确度高
- 支持中文搜索

### 9.4 Phase 4: 测试与优化（Week 6）

**任务**:
- 单元测试补充（核心服务覆盖、边界测试）
- 集成测试实现（数据库测试、端到端流程）
- 性能优化（启动时间、内存、搜索性能）
- 错误处理完善（全局捕获、降级方案）

**验收标准**:
- 测试覆盖率>70%
- 启动时间<500ms
- 内存占用<50MB

### 9.5 Phase 5: 文档与发布（Week 7）

**任务**:
- 用户文档（README、配置文档、快捷键）
- 开发文档（架构、API、贡献指南）
- 发布准备（GitHub Release、Homebrew formula、安装脚本）
- 示例素材（演示GIF、配置示例）

**验收标准**:
- 文档清晰完整
- 安装流程顺畅
- 发布包可用

---

## 10. 发布计划

### 10.1 版本规划

**v0.1.0 (Alpha) - Phase 2后**:
- 基础会话列表浏览
- 继续会话功能
- 基础预览显示

**v0.2.0 (Beta) - Phase 3后**:
- 完整搜索功能

**v1.0.0 (正式版) - Phase 5后**:
- 全功能实现
- 测试覆盖完整
- 文档完善
- 性能优化到位

### 10.2 安装方式

**编译安装（推荐）**:
```bash
git clone https://github.com/yourname/ocsession
cd ocsession
make install
```

**Homebrew安装（计划）**:
```bash
brew tap yourname/tap
brew install ocsession
```

**直接下载**:
```bash
curl -sL https://github.com/yourname/ocsession/releases/latest/download/ocsession-macos -o /usr/local/bin/ocsession
chmod +x /usr/local/bin/ocsession
```

---

## 11. 技术约束与假设

### 11.1 技术约束

- OpenCode数据库路径固定为 `~/.local/share/opencode/opencode.db`
- OpenCode版本 >= 1.3.0
- Go版本 >= 1.21
- 仅支持macOS和Linux（Windows需WSL）
- SQLite数据库schema可能随OpenCode版本变化（需兼容处理）

### 11.2 设计假设

- 用户会话数量通常 < 1000（超出时性能可能下降）
- 会话数据存储在本地数据库（不支持远程服务器）
- OpenCode CLI已正确安装和配置
- 用户熟悉终端操作和Vim风格快捷键

### 11.3 风险与缓解

**风险1**: OpenCode数据库schema变更导致查询失败
- **缓解**: 使用CLI备用方案，监控OpenCode版本更新

**风险2**: 大量会话导致性能下降
- **缓解**: 分页加载、延迟查询、缓存优化

**风险3**: 终端兼容性问题
- **缓解**: 测试主流终端、提供样式降级选项

**风险4**: 配置文件损坏
- **缓解**: 自动备份、配置验证、恢复机制

---

## 12. 总结

本项目设计了一个功能完整的OpenCode会话管理工具，通过TUI界面和模糊搜索，显著提升会话管理效率。采用Go语言和成熟的开源库，确保性能和可维护性。分阶段实施，确保每个阶段都有明确的交付成果和验收标准。

核心优势：
- **效率提升**: 模糊搜索，快速定位会话
- **无缝集成**: 直接调用OpenCode继续会话，流程顺畅
- **独立维护**: 独立配置文件，不侵入OpenCode数据结构
- **跨平台**: macOS和Linux支持，主流终端兼容

适用场景：
- 多项目开发者（快速切换不同项目会话）
- 长期项目维护（保留历史会话上下文）
- 学习探索（保存重要查询和探索会话）

---

**设计完成日期**: 2026-04-05  
**下一步**: 用户审查设计文档 → 创建实施计划 → 开始开发