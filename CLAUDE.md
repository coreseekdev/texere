# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Texere is a Go-based collaborative text editing library implementing Operational Transformation (OT) algorithms with high-performance Rope data structures. It provides core components for building real-time collaborative editors like Google Docs or Notion.

**Key Design Philosophy:**
- OT uses **UTF-16 code units** for positions/lengths to match JavaScript's `String.length` behavior
- All operations are **immutable** - they return new instances rather than mutating existing ones
- Rope provides O(log n) insert/delete operations through balanced binary tree structure
- **Interface Segregation Principle** - small, focused interfaces rather than monolithic ones

## Build, Test, and Development Commands

This project uses [Just](https://github.com/casey/just) as the build tool. Install with `cargo install just`.

### Essential Commands

```bash
just build         # Build all packages: go build -v ./...
just test          # Run all tests: go test -v ./...
just dev           # Format and test (quick dev workflow)
just check         # Format and lint
just ci            # Comprehensive check (tidy + vet + test)

# Package-specific tests
just test-ot       # Run OT package tests: ./pkg/concordia/...
just test-rope     # Run Rope package tests: ./pkg/rope/...

# Benchmarks
just bench         # Run all benchmarks
just bench-ot      # OT benchmarks
just bench-rope    # Rope benchmarks

# Code quality
just fmt           # Format code: go fmt ./...
just lint          # Run golangci-lint
just tidy          # Tidy dependencies: go mod tidy
```

### Running Single Tests

```bash
# Run a specific test
go test -v ./pkg/ot -run TestOperation_Apply

# Run tests in a single file
go test -v ./pkg/rope -run TestRope_Split

# Run with race detection
go test -race ./pkg/ot
```

## Package Architecture

### Core Package Relationships

```
┌─────────────────┐
│   pkg/transport │  ← WebSocket/SSE/TCP communication, Redis/memory history
├─────────────────┤
│   pkg/session   │  ← Session management, token auth, content storage
├─────────────────┤
│ pkg/concordia   │  ← Document integration layer (OT + Rope bridge)
├─────────────────┤
│     pkg/ot      │  ← Operational Transformation algorithms
├─────────────────┤
│    pkg/rope     │  ← High-performance text data structure
└─────────────────┘
```

### Package Responsibilities

**pkg/ot** - Operational Transformation
- Core OT operations: Insert, Delete, Retain
- `OperationBuilder` - Fluent API for constructing operations with automatic optimization
- Operation composition (compose multiple operations into one)
- Operation transformation (resolve concurrent edit conflicts)
- Undo/redo with `UndoManager`
- **Important**: All positions and lengths use **UTF-16 code units** (not bytes, not runes)

**pkg/rope** - Immutable Text Data Structure
- Balanced binary tree for O(log n) insert/delete
- Multiple implementations with different performance tradeoffs:
  - `InsertFast`/`DeleteFast` - 4-16x faster for single-node operations
  - `InsertOptimized`/`DeleteOptimized` - General purpose (17-35% faster than standard)
- Focused interfaces (Interface Segregation):
  - `ReadOnlyDocument` - read access
  - `MutableDocument` - modifications
  - `SplittableDocument` - split operations
  - `Concatenable` - concatenation
- Comprehensive Unicode support: UTF-8, grapheme clusters, word boundaries
- **Length methods**:
  - `Length()` - character count (runes)
  - `LenUTF16()` - UTF-16 code units (for JS compatibility)

**pkg/concordia** - Document Integration
- Bridges OT and Rope packages
- `RopeDocument` - Rope-based document implementation
- `StringDocument` - Simple string-based document
- Document builders and change tracking

**pkg/session** - Session Management
- Token-based authentication (Jupyter-style)
- User session management with pub/sub
- Content storage interface (K/V store)
- Structured error handling

**pkg/transport** - Distributed Communication
- Multiple transport backends: WebSocket, SSE, TCP
- History service with Redis/memory backends
- Protocol handler for message routing
- Patch-based delta compression for efficient sync

**e2e** - End-to-End Testing
- Framework for testing collaborative editing scenarios
- Integration tests for multi-user editing

## Key Architectural Patterns

### 1. Immutability Pattern
All operations return new instances:
```go
// Rope operations return new Rope instances
newRope := oldRope.Insert(5, "text")  // oldRope unchanged

// OT operations are immutable
op1 := ot.NewBuilder().Insert("Hello").Build()
op2 := op1.Retain(5)  // op1 unchanged, op2 is new
```

### 2. Builder Pattern with Automatic Optimization
```go
// Builder automatically merges adjacent operations
op := ot.NewBuilder().
    Retain(5).
    Insert("Hello").
    Retain(3).
    Insert("World").
    Build()  // Merges the two Retains automatically
```

### 3. Interface Segregation
Depend on smallest capability needed:
```go
func processDocument(doc rope.ReadOnlyDocument) {  // read-only access
    text := doc.String()
}

func modifyDocument(doc rope.MutableDocument) {  // modification access
    doc.Insert(0, "text")
}
```

### 4. Strategy Pattern
- Transport backends: WebSocket, SSE, TCP (interchangeable)
- History storage: Redis, memory, database (interchangeable)
- Authentication: Multiple providers supported

## Critical Understanding: UTF-16 vs UTF-8 vs Runes

**This is the most important architectural detail:**

OT operations use **UTF-16 code units** for JavaScript compatibility (matching `ot.js`). Go strings are UTF-8 encoded.

| String | `len(string)` (bytes) | `utf8.RuneCount()` (chars) | OT `Length()` (UTF-16) |
|--------|----------------------|---------------------------|------------------------|
| `"Hello"` | 5 | 5 | 5 |
| `"中文"` | 6 | 2 | 2 (Chinese chars in BMP) |
| `""` | 4 | 1 | 2 (emoji needs surrogate pair) |
| `"Hi!"` | 9 | 3 | 5 (H=1, i=1, =2, !=1) |

**Why?** JavaScript strings are UTF-16. `String.length` returns UTF-16 code units. For OT operations to be compatible between Go and JavaScript clients, positions must use the same encoding.

**See also:** `docs/LENGTH_CALCULATION.md` for comprehensive explanation.

## Key File Locations

### Core OT Implementation
- `pkg/ot/operation.go` - Core operation logic with UTF-16 handling
- `pkg/ot/builder.go` - Fluent builder API
- `pkg/ot/types.go` - Op interfaces (RetainOp, InsertOp, DeleteOp)
- `pkg/ot/string_document.go` - Simple document with UTF-16 length calculation

### Rope Implementation
- `pkg/rope/rope.go` - Core Rope data structure
- `pkg/rope/insert_optimized.go` - Optimized insert (17-35% faster)
- `pkg/rope/delete_optimized.go` - Optimized delete
- `pkg/rope/interfaces.go` - Focused document interfaces
- `pkg/rope/text_utf16.go` - UTF-16 support for JS compatibility

### Transport Layer
- `pkg/transport/websocket.go` - WebSocket transport
- `pkg/transport/interfaces.go` - Transport and HistoryService interfaces
- `pkg/transport/session_manager.go` - Multi-user session management

### Documentation
- `README.md` - Project overview and quick start
- `QUICKSTART.md` - OT library 5-minute guide
- `pkg/rope/QUICKSTART.md` - Rope usage guide
- `docs/LENGTH_CALCULATION.md` - UTF-16 compatibility explanation
- `docs/HISTORY_SERVICE_INTERFACE.md` - History service design

## Common Development Patterns

### Creating OT Operations
```go
// Insert text at position
op := ot.NewBuilder().Retain(pos).Insert(text).Build()

// Delete text at position
op := ot.NewBuilder().Retain(pos).Delete(length).Build()

// Replace text
op := ot.NewBuilder().Retain(start).Delete(oldLen).Insert(newText).Build()
```

### Working with Rope
```go
// Create rope from string
r := rope.New("Hello World")

// Fast operations (4-16x faster for single-node)
r = r.InsertFast(5, "Beautiful ")
r = r.DeleteFast(16, 21)

// Optimized operations (general purpose, 17-35% faster)
r, _ = r.InsertOptimized(5, "Beautiful ")
r, _ = r.DeleteOptimized(16, 21)

// UTF-16 length for JS compatibility
utf16Len := r.LenUTF16()
charLen := r.Length()
```

### Testing OT Operations
```go
func TestOperationApply(t *testing.T) {
    op := ot.NewBuilder().Insert("Hello").Build()
    result, err := op.Apply("")
    assert.NoError(t, err)
    assert.Equal(t, "Hello", result)
}
```

## Common Pitfalls

1. **Using byte positions instead of UTF-16 positions in OT**
   - Wrong: `op.Retain(len(str))` - bytes don't match UTF-16
   - Right: `doc.Length()` - returns UTF-16 code units

2. **Mutating operations instead of using new instances**
   - Wrong: `rope.Insert(5, "text")` and expecting original to change
   - Right: `newRope := rope.Insert(5, "text")`

3. **Forgetting Chinese/Emoji need special handling**
   - Chinese characters (in BMP): 1 UTF-16 code unit each
   - Emoji (outside BMP): 2 UTF-16 code units (surrogate pair)

4. **Using wrong Rope method for the use case**
   - Single insert on small rope: Use `InsertFast` (4-16x faster)
   - Multiple operations: Use `InsertOptimized` (17-35% faster)
   - Standard case: Use `Insert` (most readable)

## Development Notes

- The project uses **master** as the default branch (not main)
- All public APIs should handle errors gracefully (no panics)
- Performance is critical - Rope optimizations provide 17-35% speedup
- Test coverage is comprehensive - run tests before committing
- Documentation is in both English and Chinese (README.md)
