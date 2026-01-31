# Length Calculation and JavaScript Compatibility

## Overview

This document explains why and how OT (Operational Transformation) and Rope use UTF-16 code unit counting for JavaScript compatibility, and how this affects Chinese characters and Emoji handling.

## Background: UTF-16 Code Units vs. Runes

### JavaScript String Length

In JavaScript, strings are UTF-16 encoded. The `String.length` property returns the number of **UTF-16 code units**, not the number of characters (Unicode code points).

```javascript
// JavaScript
const text = "Hello";              // 5 UTF-16 code units
console.log(text.length);          // 5

const emoji = "";                  // 1 character, 2 UTF-16 code units (surrogate pair)
console.log(emoji.length);         // 2

const chinese = "中文";            // 2 characters, 2 UTF-16 code units
console.log(chinese.length);       // 2

const mixed = "Hi";               // 3 characters, 4 UTF-16 code units
console.log(mixed.length);         // 4 (H=1, i=1, =2)
```

### Go String Handling

In Go, strings are UTF-8 encoded. There are multiple ways to measure "length":

```go
// Go
text := "Hello"
fmt.Println(len(text))           // 5 (bytes)
fmt.Println(utf8.RuneCountInString(text))  // 5 (runes/characters)

emoji := ""
fmt.Println(len(emoji))          // 4 (bytes in UTF-8)
fmt.Println(utf8.RuneCountInString(emoji))  // 1 (rune/character)

chinese := "中文"
fmt.Println(len(chinese))        // 6 (bytes in UTF-8: 3 bytes per character)
fmt.Println(utf8.RuneCountInString(chinese))  // 2 (runes/characters)

// The OT Length() method returns UTF-16 code units to match JavaScript
```

## Why OT Uses UTF-16 Code Units

The OT implementation is designed for compatibility with [ot.js](https://github.com/Operational-Transformation/ot.js), a popular JavaScript library for operational transformation. To ensure operations can be serialized and shared between Go and JavaScript clients:

1. **Position indexing**: All positions (retain, delete) are in UTF-16 code units
2. **Base length**: The expected document length before operation
3. **Target length**: The resulting document length after operation

### Example: Character Outside BMP (Basic Multilingual Plane)

```go
// String with emoji (U+1F600 - 😀)
text := "Hi"

// In JavaScript:
// text.length === 4 (H=1, i=1, 😀=2)

// In Go OT:
doc := NewStringDocument(text)
fmt.Println(doc.Length())        // 4 (UTF-16 code units)
fmt.Println(doc.LengthChars())   // 3 (Unicode code points/runes)
fmt.Println(doc.LengthBytes())   // 4 (alias for Length)
```

## OT Length Calculation Implementation

In `pkg/ot/string_document.go`:

```go
// Length returns the length of the document in UTF-16 code units.
// This matches JavaScript's string.length behavior.
func (d *StringDocument) Length() int {
    // Count UTF-16 code units (not runes, not bytes)
    count := 0
    for _, r := range d.content {
        if r >= 0x10000 {
            // Characters outside BMP need 2 UTF-16 code units (surrogate pair)
            count += 2
        } else {
            count += 1
        }
    }
    return count
}

// LengthBytes returns the length of the document in bytes.
// This is an alias for Length() for explicit intent.
func (d *StringDocument) LengthBytes() int {
    return d.Length()
}

// LengthChars returns the length of the document in characters (code points).
func (d *StringDocument) LengthChars() int {
    return utf8.RuneCountInString(d.content)
}
```

### Key Rules

1. **Characters U+0000 to U+FFFF** (BMP):
   - 1 UTF-16 code unit
   - Includes: ASCII, Latin, Greek, Cyrillic, Chinese, Japanese, Korean (CJK)
   - Examples: `'A'`, `'中'`, `'日'`, `''`

2. **Characters U+10000 to U+10FFFF** (outside BMP):
   - 2 UTF-16 code units (surrogate pair)
   - Most emoji, some historic scripts, rare CJK characters
   - Examples: `''` (U+1F600), `''` (U+1F308)

## Rope Length Calculation

The Rope package provides multiple length methods for different use cases:

### UTF-16 Code Units (JavaScript Compatible)

In `pkg/rope/text_utf16.go`:

```go
// LenUTF16 returns the number of UTF-16 code units needed to represent the rope.
// This is important for interoperability with JavaScript and Windows APIs,
// which use UTF-16 encoding.
func (r *Rope) LenUTF16() int {
    if r == nil || r.Length() == 0 {
        return 0
    }

    count := 0
    it := r.NewIterator()
    for it.Next() {
        r := it.Current()
        if r <= 0xFFFF {
            count++ // BMP character
        } else {
            count += 2 // Surrogate pair
        }
    }
    return count
}
```

### Character Count (Runes)

In `pkg/rope/rope.go`:

```go
// Length returns the number of characters (runes) in the rope.
func (r *Rope) Length() int {
    return r.length
}
```

### Position Conversion Methods

```go
// CharToUTF16Offset converts a character index to a UTF-16 code unit offset.
func (r *Rope) CharToUTF16Offset(charIdx int) int

// UTF16OffsetToChar converts a UTF-16 code unit offset to a character index.
func (r *Rope) UTF16OffsetToChar(utf16Offset int) int

// SliceUTF16 returns a substring from UTF-16 code unit start to end.
func (r *Rope) SliceUTF16(startUTF16, endUTF16 int) (string, error)
```

## Chinese Characters (Hanzi)

Most Chinese characters fall within the BMP (U+4E00 to U+9FFF), so they count as **1 UTF-16 code unit** each:

```go
text := "汉字测试"

// JavaScript: text.length === 4
// Go OT Length(): 4
// Go OT LengthChars(): 4
// Go len() (bytes): 12 (3 bytes per character in UTF-8)
```

However, some rare CJK characters (CJK Extension B, C, D, E, F, etc.) are in U+20000+ range and require surrogate pairs:

```go
// Rare CJK character (U+20000 - 𠀀)
text := "𠀀"

// JavaScript: text.length === 2
// Go OT Length(): 2
// Go OT LengthChars(): 1
```

## Emoji Characters

Most emoji are outside the BMP and require **2 UTF-16 code units**:

```go
// Common emoji
text := "😀🎉🚀"

// JavaScript: text.length === 6
// Go OT Length(): 6
// Go OT LengthChars(): 3
```

### Complex Emoji (Sequences)

Some emoji are actually sequences of multiple code points:

```go
// Family emoji (man + ZWJ + woman + ZWJ + child)
text := "👨‍👩‍👧"

// This is:
// - 👨 (1 code point, 2 UTF-16 code units)
// - ZWJ (1 code point, 1 UTF-16 code unit)
// - 👩 (1 code point, 2 UTF-16 code units)
// - ZWJ (1 code point, 1 UTF-16 code unit)
// - 👧 (1 code point, 2 UTF-16 code units)
//
// JavaScript: text.length === 8
// Go OT Length(): 8
// Go OT LengthChars(): 5
```

## OT Operation Application

When applying OT operations, the system converts UTF-16 positions to rune positions:

From `pkg/ot/operation.go` (ApplyToDocument):

```go
// Build mapping from UTF-16 position to rune position
utf16ToRunePos := make([]int, 0, len(runes)*2)
runePos := 0
utf16Pos := 0

for _, r := range runes {
    utf16ToRunePos = append(utf16ToRunePos, runePos)
    if r >= 0x10000 {
        // Surrogate pair: 2 UTF-16 code units for 1 rune
        utf16ToRunePos = append(utf16ToRunePos, runePos)
        utf16Pos += 2
    } else {
        // BMP character: 1 UTF-16 code unit
        utf16Pos += 1
    }
    runePos++
}
```

This ensures that operations like `Retain(2)` work correctly regardless of character types.

## Best Practices

### When to Use Each Length Method

| Method | Use Case |
|--------|----------|
| `Length()` / `LengthBytes()` | OT operations, JS compatibility, cursor positions |
| `LengthChars()` / `Length()` (Rope) | Character counting, loop iterations |
| `LenUTF16()` (Rope) | Explicit UTF-16 code unit counting |
| `len(string)` (Go) | Byte-level operations, buffer sizing |

### Cross-Language Interoperability

When sending operations between Go and JavaScript:

1. **Positions**: Always use UTF-16 code units
2. **Base Length**: Must match in both languages
3. **String Serialization**: JSON handles UTF-8/UTF-16 conversion automatically

### Testing

```go
// Test that Go OT matches JavaScript behavior
func TestUTF16Compatibility(t *testing.T) {
    testCases := []struct {
        text      string
        expected  int // UTF-16 code units
    }{
        {"Hello", 5},
        {"中文", 2},
        {"", 2},          // Emoji (outside BMP)
        {"Hi!", 5},       // Mixed: H(1) + i(1) +(2) + !(1)
    }

    for _, tc := range testCases {
        doc := NewStringDocument(tc.text)
        if doc.Length() != tc.expected {
            t.Errorf("Length(%q) = %d, want %d", tc.text, doc.Length(), tc.expected)
        }
    }
}
```

## References

- [Unicode UTF-16 FAQ](https://www.unicode.org/faq/utf_bom.html#UTF16)
- [MDN: String.length](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/length)
- [ot.js Documentation](https://github.com/Operational-Transformation/ot.js)
- [Unicode Supplementary Planes](https://en.wikipedia.org/wiki/Plane_(Unicode)#Supplementary_Multilingual_Plane)

## Summary

- **OT operations** use UTF-16 code units for JavaScript compatibility
- **Chinese characters** (CJK) are usually 1 UTF-16 code unit (in BMP)
- **Emoji** are typically 2 UTF-16 code units (outside BMP, require surrogate pairs)
- **Rope** provides both character count and UTF-16 code unit count methods
- Always use the correct length method for your use case
