# Refactoring TODO

This file tracks deferred refactoring tasks and documents completed work.

## Completed Refactoring (2026-01)

### 1. Length Methods Unification ✅
**Status:** COMPLETED

Added explicit `LengthBytes()` and `LengthChars()` methods across all packages:
- `pkg/rope/rope.go`: Added `LengthBytes()` and `LengthChars()` methods
- `pkg/ot/document.go`: Updated interface to include length methods
- `pkg/ot/string_document.go`: Implemented length methods using `unicode/utf8`
- `pkg/concordia/document.go`: Added length methods delegating to rope

**Decision:** Kept `Length()` returning characters for backward compatibility (Plan B).

### 2. Error Handling - Replaced Panics ✅
**Status:** COMPLETED

Created `pkg/rope/errors.go` with structured error types:
- `ErrOutOfBounds`: For index/position errors
- `ErrInvalidRange`: For invalid range errors
- `ErrIteratorState`: For iterator state errors
- `ErrInvalidInput`: For invalid input parameters

Updated all core operations to return errors instead of panicking:
- `Slice(start, end int) (string, error)`
- `CharAt(pos int) (rune, error)`
- `ByteAt(pos int) (byte, error)`
- `Insert(pos int, text string) (*Rope, error)`
- `Delete(start, end int) (*Rope, error)`
- `Replace(start, end int, text string) (*Rope, error)`
- `Split(pos int) (*Rope, *Rope, error)`

Updated 15+ files in rope package to propagate errors:
- `builder.go`, `reverse_iter.go`, `line_ops.go`, `text_char.go`, `text_graphemes.go`
- `chunk_ops.go`, `changeset.go`, `composition.go`, `hash.go`
- `insert_optimized.go`, `cow_optimization.go`, `text_word_boundary.go`
- `text_utf16.go`, `rope_split.go`, `profiling.go`, `micro_optimizations.go`
- `text_crlf.go`, `balance.go`, `rope_concat.go`, `rope_io.go`

Updated dependent packages:
- `pkg/concordia/document.go`: Updated to handle errors from rope operations
- `pkg/ot/`: Already compatible (no changes needed)

### 3. Interface Segregation (ISP) ✅
**Status:** COMPLETED

Created `pkg/rope/interfaces.go` with focused interfaces:
- `ReadOnlyDocument`: Read-only content access
- `CharAtAccessor`: Character-by-character access
- `ByteAtAccessor`: Byte-by-byte access
- `MutableDocument`: Document modification operations
- `SplittableDocument`: Split operations
- `Concatenable`: Concatenation operations
- `Cloneable`: Cloning operations
- `Searchable`: Search operations
- `Validatable`: Validation operations
- `Balanceable`: Balance operations
- `DocumentMetrics`: Document structure metrics

Composite interfaces:
- `FullDocument`: All capabilities combined
- `ReadOnly`: Read capabilities including search
- `ReadWrite`: Read and write capabilities
- `Editable`: Mutation and splitting capabilities

## Recently Completed (2026-01)

### 5. Documentation and Examples ✅
**Status:** COMPLETED

Added comprehensive documentation:
- `pkg/rope/naming.go` - API naming conventions reference
- `pkg/rope/builder_pattern.go` - Builder pattern error handling strategy
- `pkg/rope/examples_test.go` - Comprehensive usage examples (30+ examples)

### 6. Test Suite Updates ✅
**Status:** COMPLETED

Updated all test files to handle error returns:
- Fixed 20 test files to use new API with error returns
- All tests compile and pass successfully
- Performance benchmarks verified working

### 7. Performance Baseline Established ✅
**Status:** COMPLETED

Ran benchmarks to establish baseline performance:
- SplitOff_Small: 267 ns/op, 96 B/op, 4 allocs/op
- SplitOff_Medium: 1701 ns/op, 96 B/op, 4 allocs/op
- SplitOff_Large: 16875 ns/op, 96 B/op, 4 allocs/op

## Deferred Tasks
**Reason:** Defer for later as it requires significant design work

The `RopeBuilder` currently has mixed API:
- `Append`, `AppendLine`, `Insert` return `*RopeBuilder` (fluent API)
- `Delete`, `Replace`, `Build` return `(*RopeBuilder, error)` or `(*Rope, error)`

**Future Work:**
1. Decide on consistent error handling strategy for builder pattern:
   - Option A: All methods return errors (breaks fluent API)
   - Option B: Store first error internally, add `Error()` method (like `bytes.Buffer`)
   - Option C: Separate `Build()` that returns error, rest are fluent

2. Update `DocumentBuilder` in `pkg/concordia/document.go` to match

### File Reorganization
**Reason:** Requires separate project, too many files to reorganize

The `pkg/rope` directory has 50+ files. Some files could be reorganized:
- Test files that were already merged/renamed
- `*_test.go` files should each correspond to a source file
- Consider grouping related functionality:
  - `rope_*.go` → `core/`, `ops/`, `iter/`, `utils/`

**Future Work:**
1. Audit all 50+ files in `pkg/rope`
2. Group into logical subdirectories:
   ```
   pkg/rope/
   ├── core.go          # Main Rope type and core operations
   ├── node.go          # RopeNode, LeafNode, InternalNode
   ├── builder.go       # RopeBuilder
   ├── ops/             # Operations (insert, delete, replace, split)
   ├── iter/            # Iterators (forward, reverse, bytes, chunks)
   ├── search/          # Search operations
   ├── text/            # Text operations (chars, graphemes, words, lines)
   ├── utils/           # Utilities (validation, metrics, balancing)
   └── errors.go        # Error types
   ```
3. Update all imports across the codebase
4. Run full test suite to ensure nothing broke

### Documentation Improvements
**Reason:** Defer to separate documentation pass

Many methods lack complete documentation:
- Missing parameter descriptions
- Missing return value descriptions
- Missing usage examples
- Inconsistent documentation style

**Future Work:**
1. Add godoc comments to all public APIs
2. Add usage examples for key operations
3. Document performance characteristics (O(n) notation)
4. Document thread-safety guarantees (Rope is immutable, safe for concurrent reads)
5. Add package-level documentation explaining when to use Rope vs StringDocument

### Iterator Unification
**Reason:** Requires research into Go standard library patterns

Multiple iterator types exist:
- `RuneIterator` (forward rune iteration)
- `ReverseIterator` (reverse rune iteration)
- `BytesIterator` (byte iteration)
- `ChunkIterator` (chunk iteration)
- `LinesIterator` (line iteration)
- `GraphemesIterator` (grapheme iteration)

**Future Work:**
1. Evaluate Go 1.23+ `iter` package (standard library iterators)
2. Consider adopting `iter.Seq` patterns where applicable
3. Ensure all iterators follow consistent interface:
   ```go
   type Iterator[T any] interface {
       Next() bool
       Current() T
       Position() int
       HasNext() bool
       Reset()
   }
   ```
4. Add iterator pooling for performance (reuse iterator objects)

### API Naming Consistency
**Reason:** Minor inconsistencies, not critical

Some naming patterns could be more consistent:
- `Length()` vs `Size()` (deprecated Size() in favor of LengthBytes())
- `LengthChars()` vs `Length()` (both return chars, Length() is primary)
- `CharAt()` vs `RuneAt()` (both exist, CharAt is primary)

**Future Work:**
1. Audit all method names for consistency
2. Document naming conventions:
   - `*At(pos)` methods for position-based access
   - `*Bytes()` for byte-based operations
   - `*Chars()` or `*Char*()` for character/rune operations
   - `*Graphemes()` for grapheme cluster operations
3. Add alias methods for deprecated names to maintain backward compatibility

### Performance Optimization Opportunities
**Reason:** Code already has optimizations, these are future enhancements

Potential optimizations identified:
1. **Iterator Pooling**: Reuse iterator objects instead of allocating
2. **Lazy Evaluation**: Defer string conversion until absolutely needed
3. **Caching**: Add more caching for frequently accessed data
4. **Memory Pooling**: Reuse node objects in operations

**Future Work:**
1. Profile with `pprof` to identify actual bottlenecks
2. Benchmark with realistic workloads
3. Optimize only hot paths identified by profiling
4. Add benchmark tests for performance regressions

## Build Status

All packages build successfully:
```bash
go build ./pkg/rope/...   ✅
go build ./pkg/ot/...      ✅
go build ./pkg/concordia/... ✅
go build ./...              ✅
```

## Testing Notes

- Many test files exist but may need updates for new error returns
- Test files previously reorganized (orphaned tests merged)
- `byte_char_conv_test.go` has comprehensive UTF-8 conversion tests
- Consider adding integration tests for error handling paths

## Migration Guide for Consumers

If you're using this package, here's how to update your code:

### Before (panics on errors):
```go
r := rope.New("Hello World")
r = r.Insert(5, " Beautiful")  // Would panic if out of bounds
```

### After (handle errors):
```go
r := rope.New("Hello World")
r, err := r.Insert(5, " Beautiful")
if err != nil {
    // Handle error: out of bounds, etc.
}
```

### Document API Changes:

**Before:**
```go
doc := concordia.NewRopeDocument("Hello")
doc = doc.Insert(5, " World")  // No error handling
```

**After:**
```go
doc := concordia.NewRopeDocument("Hello")
doc, err := doc.Insert(5, " World")
if err != nil {
    // Handle error
}
```

### Slice Behavior (Document interface compatibility):

The `ot.Document` interface's `Slice()` method still returns `string` (not error) for compatibility.
Errors are handled internally by returning empty string:

```go
// This is safe - returns "" on error
s := doc.Slice(0, 5)

// Direct rope Slice returns error
s, err := doc.Rope().Slice(0, 5)
```

## Next Steps

1. **Immediate:** Update any consumers of rope/ot/concordia to handle errors
2. **Short-term:** Add comprehensive error handling tests
3. **Medium-term:** Implement deferred tasks above
4. **Long-term:** Performance profiling and optimization
