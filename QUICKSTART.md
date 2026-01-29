# Quick Start Guide - Texere OT (Concordia)

This guide will help you get started with the Concordia OT library in 5 minutes.

## Installation

The library is already part of the Texere project at:
```
S:/workspace/texere-ot/pkg/concordia
```

## Your First OT Operation

### 1. Basic Text Insertion

```go
package main

import (
    "fmt"
    "github.com/coreseekdev/texere/pkg/concordia"
)

func main() {
    // Create an operation that inserts "Hello" at position 0
    op := concordia.NewBuilder().
        Insert("Hello").
        Build()

    // Apply to an empty document
    result, err := op.Apply("")
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // Output: Hello
}
```

### 2. Complex Operations

```go
// Create an operation that:
// 1. Skips first 5 characters
// 2. Inserts "Hello"
// 3. Skips next 3 characters
// 4. Deletes 2 characters
op := concordia.NewBuilder().
    Retain(5).
    Insert("Hello").
    Retain(3).
    Delete(2).
    Build()

doc := "_____XYZ_______"
result, _ := op.Apply(doc)
fmt.Println(result) // Output: _____HelloXYZ_____
```

### 3. Collaborative Editing (Transform)

```go
// Two users editing at the same time

// User A inserts "Hello" at position 0
userAOp := concordia.NewBuilder().Insert("Hello").Build()

// User B inserts "World" at position 0
userBOp := concordia.NewBuilder().Insert("World").Build()

// Transform to resolve the conflict
transformedA, transformedB, _ := concordia.Transform(userAOp, userBOp)

// Both users apply the transformed operations
doc := ""
docA, _ := transformedA.Apply(doc)
docB, _ := transformedB.Apply(docA)

// Result is consistent regardless of order
fmt.Println(docB) // Output: HelloWorld or WorldHello (consistent)
```

### 4. Undo/Redo

```go
// Create an undo manager
um := concordia.NewUndoManager(50)

// Apply an operation
op := concordia.NewBuilder().Insert("Hello").Build()
doc, _ := op.Apply(doc)
fmt.Println(doc) // Output: Hello

// Add inverse to undo stack
inverse, _ := op.Invert("")
um.Add(inverse, true)

// Undo
um.PerformUndo(func(op *concordia.Operation) {
    doc, _ = op.Apply(doc)
})
fmt.Println(doc) // Output: (empty)

// Redo
um.PerformRedo(func(op *concordia.Operation) {
    doc, _ = op.Apply(doc)
})
fmt.Println(doc) // Output: Hello
```

## Common Patterns

### Pattern 1: Building Operations Step by Step

```go
builder := concordia.NewBuilder()

// Add operations based on user input
if userInserted {
    builder.Insert(userText)
}
if userDeleted {
    builder.Delete(deleteLength)
}
if userMovedCursor {
    builder.Retain(cursorPos)
}

// Build the final operation
op := builder.Build()
```

### Pattern 2: Applying Operations to Documents

```go
import "github.com/coreseekdev/texere/pkg/document"

// Using string (simple)
doc := "Hello World"
result, _ := op.Apply(doc)

// Using Document interface (flexible)
docImpl := document.NewStringDocument("Hello World")
resultDoc, _ := op.ApplyToDocument(docImpl)
fmt.Println(resultDoc.String())
```

### Pattern 3: Composing Operations

```go
// First operation: Insert "Hello"
op1 := concordia.NewBuilder().Insert("Hello ").Build()

// Second operation: Insert "World"
op2 := concordia.NewBuilder().Retain(6).Insert("World").Build()

// Compose into single operation
composed, _ := concordia.Compose(op1, op2)

// Apply composed operation
result, _ := composed.Apply("")
fmt.Println(result) // Output: Hello World
```

### Pattern 4: Client-Server Synchronization

```go
// Create client
client := concordia.NewClient()

// User makes local edit
localOp := concordia.NewBuilder().Insert("Hello").Build()
doc, _ = client.ApplyClient(localOp)

// Send to server
opToSend := client.OutgoingOperation()
sendToServer(opToSend)

// Receive server acknowledgment
client.ServerAck()

// Receive remote operation from server
remoteOp := receiveFromServer()
doc, _ = client.ApplyServer(revision, remoteOp)
```

## Testing Your Code

The library includes comprehensive tests. Run them:

```bash
# Navigate to project
cd S:/workspace/texere-ot

# Run all tests
go test ./pkg/concordia/... -v

# Run specific test
go test ./pkg/concordia/... -run TestOperation_Apply_Random -v

# Run with coverage
go test -cover ./pkg/concordia/...
```

## Key Concepts

### Operation Types

1. **Retain**: Skip over characters without changing them
   ```go
   builder.Retain(5) // Skip next 5 characters
   ```

2. **Insert**: Insert new text at current position
   ```go
   builder.Insert("Hello") // Insert "Hello"
   ```

3. **Delete**: Remove characters at current position
   ```go
   builder.Delete(3) // Delete next 3 characters
   ```

### Builder Optimization

The builder automatically merges adjacent operations:

```go
op := concordia.NewBuilder().
    Retain(5).
    Retain(3).      // Automatically merged
    Insert("Hello").
    Insert(" World"). // Automatically merged
    Build()

// Result: retain(8).insert("Hello World")
```

### Immutability

Operations are immutable and safe for concurrent use:

```go
op1 := concordia.NewBuilder().Insert("Hello").Build()

// This creates a NEW operation, doesn't modify op1
op2 := concordia.NewBuilder().Insert("World").Build()

// Both are independent
```

## Error Handling

Always check for errors:

```go
result, err := op.Apply(doc)
if err != nil {
    // Handle error
    if err == concordia.ErrInvalidBaseLength {
        fmt.Println("Operation doesn't match document length")
    } else {
        fmt.Println("Error:", err)
    }
}
```

## Best Practices

1. **Use Builder Pattern**: Always use `NewBuilder()` for constructing operations
2. **Check Errors**: Always handle errors from `Apply()`, `Compose()`, `Transform()`
3. **Use Document Interface**: For flexibility, use `ApplyToDocument()` instead of `Apply()`
4. **Test Thoroughly**: Use randomized tests to verify OT properties
5. **Consider Performance**: Large operations may need batching

## Next Steps

1. Read the full documentation: `pkg/concordia/README.md`
2. Explore the examples: `examples/`
3. Review the test files for more usage patterns
4. Integrate into your Texere editor

## Troubleshooting

### "base length does not match document length"

This means you're trying to apply an operation to a document with the wrong length:

```go
// Wrong: operation expects 10 characters, document has 5
op := concordia.NewBuilder().Retain(10).Build()
result, err := op.Apply("short") // Error!

// Right: document length matches operation baseLength
result, err := op.Apply("1234567890") // OK!
```

### Transform Returns Error

Transform requires both operations to have the same baseLength:

```go
// Wrong: different baseLength
op1 := concordia.NewBuilder().Retain(5).Build()
op2 := concordia.NewBuilder().Retain(10).Build()
_, _, err := concordia.Transform(op1, op2) // Error!

// Right: same baseLength
op1 := concordia.NewBuilder().Retain(10).Build()
op2 := concordia.NewBuilder().Insert("Hello").Build()
_, _, err := concordia.Transform(op1, op2) // OK!
```

## Resources

- **API Documentation**: `pkg/concordia/README.md`
- **Implementation Report**: `INITIALIZATION_REPORT.md`
- **Source Code**: `pkg/concordia/*.go`
- **Tests**: `pkg/concordia/*_test.go`

## Support

For issues or questions:
1. Check the documentation
2. Review the test files for examples
3. Examine the source code comments

---

**Happy Coding! 🚀**

The Concordia OT library is ready to help you build real-time collaborative editing features.
