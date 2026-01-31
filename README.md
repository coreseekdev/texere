# Texere

> **Weave Knowledge Together** - 编织知识，连接智慧

Texere 是一个基于 Operational Transformation (OT) 和 Rope 数据结构的文本编辑核心库。

## 🎯 项目概述

Texere 提供了构建实时协作编辑器和文本编辑器所需的核心组件：

- **Operational Transformation (OT)** - 通过 `pkg/ot` 包实现高效的 OT 算法，兼容 JavaScript ot.js
- **Rope 数据结构** - 通过 `pkg/rope` 包实现高性能的文本操作
- **文档抽象** - 通过 `pkg/concordia` 包提供统一的文档接口
- **传输层** - 通过 `pkg/transport` 包支持 WebSocket/SSE/TCP 实时通信
- **会话管理** - 通过 `pkg/session` 包提供用户认证和会话管理

## ✨ 核心特性

### OT (Operational Transformation)
- ✅ 完整的操作转换实现（Insert, Delete, Retain）
- ✅ 操作组合 (Compose)
- ✅ 操作转换 (Transform) - 支持并发编辑冲突解决
- ✅ 操作反转 (Invert) - 支持 Undo/Redo
- ✅ 客户端同步 (Client) - 支持客户端-服务器架构
- ✅ 撤销管理器 (UndoManager) - 带时间戳的撤销/重做

### Rope 数据结构
- ✅ 不可变二叉树结构 - 高效的文本操作
- ✅ 快速插入/删除 - O(log n) 时间复杂度
- ✅ 零拷贝切片 - 高效的文本访问
- ✅ UTF-8/UTF-16 支持 - 完整的 Unicode 支持，兼容 JavaScript
- ✅ 字节/字符迭代器 - 灵活的文本遍历
- ✅ 性能优化 - InsertOptimized/DeleteOptimized (比标准实现快 17-35%)
- ✅ 事务支持 - 支持原子操作和位置映射
- ✅ 接口隔离 - 小而专注的接口设计

### 传输与协作
- ✅ WebSocket/SSE/TCP 支持 - 多种实时通信方式
- ✅ Redis/Memory 历史服务 - 灵活的版本历史存储
- ✅ 补丁压缩 - Delta 压缩减少网络传输
- ✅ 会话管理 - Token 认证和用户会话
- ✅ 多文档支持 - 单连接管理多个文档

### 性能
- **插入操作**: InsertOptimized 比 ZeroAlloc 快 **17%**
- **删除操作**: DeleteOptimized 与 ZeroAlloc 相当或更快
- **单叶优化**: InsertFast/DeleteFast 快 **4-16x**
- **内存优化**: 移除了 ZeroAlloc (内存开销减少 97%)

## 📦 包结构

```
texere/
├── pkg/ot/           # OT 核心算法
│   ├── operation.go         # 操作定义和实现
│   ├── builder.go           # 操作构建器
│   ├── transform.go         # 操作转换
│   ├── compose.go           # 操作组合
│   ├── string_document.go   # String 文档实现
│   └── undoable_document.go # Undo/Redo 支持
├── pkg/rope/          # Rope 数据结构
│   ├── rope.go              # 核心 Rope 实现
│   ├── insert_optimized.go  # 优化的插入操作
│   ├── delete_optimized.go  # 优化的删除操作
│   ├── text_utf16.go        # UTF-16 支持（JS 兼容）
│   └── interfaces.go        # 文档接口定义
├── pkg/concordia/     # 文档集成层
│   ├── document.go          # Document 接口
│   └── rope_document.go     # Rope 文档实现
├── pkg/session/       # 会话管理
│   ├── session.go           # 会话管理
│   └── manager.go           # 会话管理器
├── pkg/transport/     # 传输层
│   ├── websocket.go         # WebSocket 传输
│   ├── interfaces.go        # 传输接口定义
│   └── session_manager.go   # 多用户会话管理
├── e2e/               # 端到端测试
│   └── transport_test.go    # 集成测试
├── docs/              # 文档
│   └── LENGTH_CALCULATION.md # UTF-16 长度计算说明
├── QUICKSTART.md      # OT 快速入门
└── ROPE_QUICKSTART.md # Rope 快速入门
```

## 🚀 快速开始

### 安装

```bash
go get github.com/coreseekdev/texere
```

### OT 基础使用

```go
package main

import (
    "fmt"
    "github.com/coreseekdev/texere/pkg/ot"
)

func main() {
    // 创建文档
    doc := ot.NewStringDocument("Hello")

    // 创建插入操作
    op := ot.NewBuilder().
        Retain(5).
        Insert(" World").
        Build()

    // 应用操作
    result, err := op.ApplyToDocument(doc)
    if err != nil {
        panic(err)
    }

    fmt.Println(result.String()) // "Hello World"
}
```

### Rope 基础使用

```go
package main

import (
    "fmt"
    "github.com/coreseekdev/texere/pkg/rope"
)

func main() {
    // 创建 Rope
    r := rope.New("Hello, World!")

    // 插入文本
    r = r.InsertFast(7, "Beautiful ")

    // 删除文本
    r = r.DeleteFast(16, 25)

    // 获取结果
    fmt.Println(r.String()) // "Hello, Beautiful!"
}
```

## 📚 文档

- **[OT 快速入门](QUICKSTART.md)** - 5 分钟上手 OT 库
- **[Rope 快速入门](pkg/rope/QUICKSTART.md)** - Rope 数据结构使用指南
- **[OT API 文档](pkg/ot/README.md)** - OT API 参考
- **[Rope API 文档](pkg/rope/README.md)** - Rope API 参考
- **[UTF-16 长度计算](docs/LENGTH_CALCULATION.md)** - JavaScript 兼容性和字符编码
- **[传输层协议](docs/PROTOCOL.md)** - 实时通信协议说明

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行 OT 测试
go test ./pkg/ot/... -v

# 运行 Rope 测试
go test ./pkg/rope/... -v

# 带覆盖率
go test ./... -cover
```

## 🔧 构建

项目使用 [just](https://github.com/casey/just) 作为构建工具：

```bash
# 安装 just
cargo install just

# 查看可用命令
just --list

# 运行测试
just test

# 构建项目
just build

# 清理
just clean
```

## 🏗️ 分支结构

- **master** - 主分支，包含最新的 OT、Rope、Transport 功能
- **feature/transport** - 传输层功能分支（已合并到 master）

## 🌟 亮点特性

- **JavaScript 兼容**: OT 操作使用 UTF-16 code units，与 `ot.js` 完全兼容
- **高性能**: Rope 优化实现比标准快 17-35%
- **类型安全**: Go 的静态类型系统提供编译时检查
- **完整测试**: 覆盖率高的单元测试和端到端测试
- **生产就绪**: 支持分布式部署、Redis 存储、负载均衡

## 📊 性能基准

### Insert 操作
| 实现 | 速度 (ns/op) | 内存 (B/op) |
|------|-------------|-------------|
| InsertFast | 144 | 72 |
| InsertOptimized | 1952 | 2864 |
| Insert (Standard) | 2991 | 880 |

### Delete 操作
| 实现 | 速度 (ns/op) | 内存 (B/op) |
|------|-------------|-------------|
| DeleteFast | 174 | 56 |
| DeleteOptimized | 672 | 2864 |
| Delete (Standard) | 922 | 1456 |

## 🤝 贡献

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

Copyright (c) 2025 Texere Contributors

---

**Texere - Weave Knowledge Together** 🧵✨
