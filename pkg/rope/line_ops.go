package rope

import (
	"strings"
)

// Line operations provide editor-friendly functionality for working with lines.
// All line numbers are 0-indexed (first line is line 0).

// Line returns the text of the specified line (without line ending).
// Panics if lineNum is out of bounds.
func (r *Rope) Line(lineNum int) (string, error) {
	lineCount := r.LineCount()
	if lineNum < 0 || lineNum >= lineCount {
		return "", &ErrOutOfBounds{
			Operation: "Line",
			Position:  lineNum,
			Min:       0,
			Max:       lineCount,
		}
	}

	start := r.LineStart(lineNum)
	end, err := r.LineEnd(lineNum)
	if err != nil {
		return "", err
	}
	return r.Slice(start, end)
}

// LineWithEnding returns the text of the specified line including the line ending.
// Panics if lineNum is out of bounds.
func (r *Rope) LineWithEnding(lineNum int) (string, error) {
	lineCount := r.LineCount()
	if lineNum < 0 || lineNum >= lineCount {
		return "", &ErrOutOfBounds{
			Operation: "LineWithEnding",
			Position:  lineNum,
			Min:       0,
			Max:       lineCount,
		}
	}

	start := r.LineStart(lineNum)
	end := start + r.LineLength(lineNum)
	return r.Slice(start, end)
}

// LineCount returns the total number of lines in the rope.
// An empty rope has 0 lines. A rope with content has at least 1 line.
func (r *Rope) LineCount() int {
	if r.length == 0 {
		return 0
	}

	content := r.String()
	count := strings.Count(content, "\n")

	// If content doesn't end with newline, add 1 for the last line
	if !strings.HasSuffix(content, "\n") {
		return count + 1
	}

	return count
}

// LineStart returns the character position where the specified line starts.
// Panics if lineNum is out of bounds.
func (r *Rope) LineStart(lineNum int) int {
	if lineNum < 0 || lineNum >= r.LineCount() {
		panic("line number out of bounds")
	}

	if lineNum == 0 {
		return 0
	}

	it := r.NewIterator()
	currentLine := 0

	for it.Next() {
		if it.Current() == '\n' {
			currentLine++
			if currentLine == lineNum {
				// Return position AFTER the newline
				// Position() returns charPos + 1, which is after the newline
				return it.Position()
			}
		}
	}

	// Should not reach here
	return r.Length()
}

// LineEnd returns the character position where the specified line ends (exclusive).
// This does not include the line ending character.
// Panics if lineNum is out of bounds.
func (r *Rope) LineEnd(lineNum int) (int, error) {
	if lineNum < 0 || lineNum >= r.LineCount() {
		return 0, &ErrOutOfBounds{
			Operation: "LineEnd",
			Position:  lineNum,
			Min:       0,
			Max:       r.LineCount(),
		}
	}

	start := r.LineStart(lineNum)

	// Find the next newline after start
	for i := start; i < r.Length(); i++ {
		ch, err := r.CharAt(i)
		if err != nil {
			return 0, err
		}
		if ch == '\n' {
			return i, nil
		}
	}

	// No newline found, this is the last line
	return r.Length(), nil
}

// LineLength returns the length of the specified line in characters (excluding line ending).
// Panics if lineNum is out of bounds.
func (r *Rope) LineLength(lineNum int) int {
	start := r.LineStart(lineNum)
	end, _ := r.LineEnd(lineNum)
	return end - start
}

// LineWithEndingLength returns the length of the specified line including the line ending.
// Panics if lineNum is out of bounds.
func (r *Rope) LineWithEndingLength(lineNum int) (int, error) {
	if lineNum < 0 || lineNum >= r.LineCount() {
		return 0, &ErrOutOfBounds{
			Operation: "LineWithEndingLength",
			Position:  lineNum,
			Min:       0,
			Max:       r.LineCount(),
		}
	}

	start := r.LineStart(lineNum)
	end := start + r.LineLength(lineNum)

	// Add 1 for the newline if it exists
	if end < r.Length() {
		ch, err := r.CharAt(end)
		if err != nil {
			return 0, err
		}
		if ch == '\n' {
			return (end - start) + 1, nil
		}
	}

	return end - start, nil
}

// InsertLine inserts text at the beginning of the specified line.
// Returns a new Rope, leaving the original unchanged.
// Panics if lineNum is out of bounds.
func (r *Rope) InsertLine(lineNum int, text string) (*Rope, error) {
	pos := r.LineStart(lineNum)
	return r.Insert(pos, text)
}

// DeleteLine removes the specified line.
// Returns a new Rope, leaving the original unchanged.
// Panics if lineNum is out of bounds.
func (r *Rope) DeleteLine(lineNum int) (*Rope, error) {
	if lineNum < 0 || lineNum >= r.LineCount() {
		return nil, &ErrOutOfBounds{
			Operation: "DeleteLine",
			Position:  lineNum,
			Min:       0,
			Max:       r.LineCount(),
		}
	}

	start := r.LineStart(lineNum)
	end, err := r.LineEnd(lineNum)
	if err != nil {
		return nil, err
	}

	// Check if there's a newline after the line
	if end < r.Length() {
		ch, err := r.CharAt(end)
		if err != nil {
			return nil, err
		}
		if ch == '\n' {
			end++ // Include the newline in deletion
		}
	}

	return r.Delete(start, end)
}

// ReplaceLine replaces the content of the specified line with the given text.
// Returns a new Rope, leaving the original unchanged.
// Panics if lineNum is out of bounds.
func (r *Rope) ReplaceLine(lineNum int, text string) (*Rope, error) {
	start := r.LineStart(lineNum)
	end := r.LineStart(lineNum) + r.LineLength(lineNum)
	return r.Replace(start, end, text)
}

// AppendLine appends a new line to the end of the rope.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) AppendLine(text string) (*Rope, error) {
	if r.Length() == 0 {
		return r.Insert(0, text)
	}

	// Insert after the last character
	return r.Insert(r.Length(), "\n"+text)
}

// PrependLine prepends a new line at the beginning of the rope.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) PrependLine(text string) (*Rope, error) {
	if r.Length() == 0 {
		return r.Insert(0, text)
	}

	return r.Insert(0, text+"\n")
}

// LinesIterator creates an iterator that yields one line at a time.
func (r *Rope) LinesIterator() *LinesIterator {
	return &LinesIterator{
		rope:       r,
		lineNum:    0,
		totalLines: r.LineCount(),
	}
}

// LinesIterator iterates through lines of a rope.
type LinesIterator struct {
	rope       *Rope
	lineNum    int
	totalLines int
}

// Next advances to the next line and returns true if there are more lines.
func (it *LinesIterator) Next() bool {
	it.lineNum++
	return it.lineNum < it.totalLines
}

// Current returns the current line (without line ending).
func (it *LinesIterator) Current() (string, error) {
	if it.lineNum < 0 || it.lineNum >= it.totalLines {
		return "", &ErrOutOfBounds{
			Operation: "LinesIterator.Current",
			Position:  it.lineNum,
			Min:       0,
			Max:       it.totalLines,
		}
	}
	return it.rope.Line(it.lineNum)
}

// CurrentWithEnding returns the current line including the line ending.
func (it *LinesIterator) CurrentWithEnding() (string, error) {
	if it.lineNum < 0 || it.lineNum >= it.totalLines {
		return "", &ErrOutOfBounds{
			Operation: "LinesIterator.CurrentWithEnding",
			Position:  it.lineNum,
			Min:       0,
			Max:       it.totalLines,
		}
	}
	return it.rope.LineWithEnding(it.lineNum)
}

// LineNumber returns the current line number.
func (it *LinesIterator) LineNumber() int {
	return it.lineNum
}

// Reset resets the iterator to the beginning.
func (it *LinesIterator) Reset() {
	it.lineNum = -1
}

// ToSlice collects all lines into a slice (without line endings).
func (it *LinesIterator) ToSlice() ([]string, error) {
	lines := make([]string, 0, it.totalLines)
	it.Reset()
	for it.Next() {
		line, err := it.Current()
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// ========== Line-based Editing Operations ==========

// LineAtChar returns the line number containing the given character position.
func (r *Rope) LineAtChar(pos int) int {
	if pos < 0 || pos > r.Length() {
		panic("character position out of bounds")
	}

	if pos == 0 {
		return 0
	}

	// Use iterator for efficient traversal (avoids expensive CharAt calls)
	lineNum := 0
	it := r.NewIterator()
	for i := 0; i <= pos && it.Next(); i++ {
		if it.Current() == '\n' {
			lineNum++
		}
	}

	return lineNum
}

// ColumnAtChar returns the column number (0-indexed) within the line
// for the given character position.
func (r *Rope) ColumnAtChar(pos int) int {
	if pos < 0 || pos > r.Length() {
		panic("character position out of bounds")
	}

	lineStart := r.LineStart(r.LineAtChar(pos))
	return pos - lineStart
}

// PositionAtLineCol returns the character position for the given line and column.
// Panics if lineNum or colNum is out of bounds.
func (r *Rope) PositionAtLineCol(lineNum, colNum int) int {
	lineStart := r.LineStart(lineNum)
	lineEnd, _ := r.LineEnd(lineNum)

	if colNum < 0 || colNum > (lineEnd-lineStart) {
		panic("column number out of bounds")
	}

	return lineStart + colNum
}

// InsertAtLineCol inserts text at the specified line and column.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) InsertAtLineCol(lineNum, colNum int, text string) (*Rope, error) {
	pos := r.PositionAtLineCol(lineNum, colNum)
	return r.Insert(pos, text)
}

// DeleteAtLineCol deletes characters from (lineNum, colNum) to (lineNum2, colNum2).
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) DeleteAtLineCol(lineNum, colNum, lineNum2, colNum2 int) (*Rope, error) {
	start := r.PositionAtLineCol(lineNum, colNum)
	end := r.PositionAtLineCol(lineNum2, colNum2)
	return r.Delete(start, end)
}

// ========== Line Information ==========

// HasTrailingNewline returns true if the rope ends with a newline character.
func (r *Rope) HasTrailingNewline() (bool, error) {
	if r.Length() == 0 {
		return false, nil
	}
	ch, err := r.CharAt(r.Length() - 1)
	if err != nil {
		return false, err
	}
	return ch == '\n', nil
}

// LineEnding returns the line ending style used in the rope.
// Returns "\n", "\r\n", "\r", or "" if no line endings.
func (r *Rope) LineEnding() string {
	content := r.String()

	// Check for Windows-style (CRLF)
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}

	// Check for Unix-style (LF)
	if strings.Contains(content, "\n") {
		return "\n"
	}

	// Check for Mac Classic-style (CR)
	if strings.Contains(content, "\r") {
		return "\r"
	}

	return ""
}

// NormalizeLineEndings converts all line endings to the specified style.
// Valid styles are "\n" (Unix), "\r\n" (Windows), or "\r" (Mac Classic).
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) NormalizeLineEndings(style string) (*Rope, error) {
	if style != "\n" && style != "\r\n" && style != "\r" {
		return nil, &ErrInvalidInput{
			Parameter: "style",
			Value:     style,
			Reason:    "must be \\n, \\r\\n, or \\r",
		}
	}

	content := r.String()

	// First normalize to \n
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Then convert to desired style
	if style == "\n" {
		return New(content), nil
	}

	// Convert \n to desired style
	if style == "\r\n" {
		content = strings.ReplaceAll(content, "\n", "\r\n")
	} else if style == "\r" {
		content = strings.ReplaceAll(content, "\n", "\r")
	}

	return New(content), nil
}

// TrimTrailingNewlines removes all trailing newline characters.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) TrimTrailingNewlines() (*Rope, error) {
	content := r.String()
	trimmed := strings.TrimRight(content, "\n\r")
	return New(trimmed), nil
}

// TrimLeadingNewlines removes all leading newline characters.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) TrimLeadingNewlines() (*Rope, error) {
	content := r.String()
	trimmed := strings.TrimLeft(content, "\n\r")
	return New(trimmed), nil
}

// JoinLines concatenates all lines into a single line.
// Removes all line endings.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) JoinLines() (*Rope, error) {
	content := r.String()
	joined := strings.ReplaceAll(content, "\n", "")
	joined = strings.ReplaceAll(joined, "\r", "")
	return New(joined), nil
}

// SplitLines splits the rope into lines (without line endings).
// Returns a slice of strings.
func (r *Rope) SplitLines() ([]string, error) {
	it := r.LinesIterator()
	return it.ToSlice()
}

// IndentLines adds indentation to all lines.
// prefix is added to the beginning of each line.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) IndentLines(prefix string) (*Rope, error) {
	builder := NewBuilder()
	it := r.LinesIterator()
	it.Reset()

	for it.Next() {
		builder.Append(prefix)
		lineWithEnding, err := it.CurrentWithEnding()
		if err != nil {
			return nil, err
		}
		builder.Append(lineWithEnding)
	}

	return builder.Build()
}

// DedentLines removes common leading whitespace from all lines.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) DedentLines() (*Rope, error) {
	lines, err := r.SplitLines()
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return r, nil
	}

	// Find minimum leading whitespace
	minIndent := -1
	for _, line := range lines {
		if line == "" {
			continue
		}
		indent := leadingWhitespaceCount(line)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return r, nil
	}

	// Remove minIndent from each line
	b := NewBuilder()
	for i, line := range lines {
		if len(line) >= minIndent {
			b.Append(line[minIndent:])
		}
		if i < len(lines)-1 {
			b.Append("\n")
		}
	}

	return b.Build()
}

// leadingWhitespaceCount returns the number of leading whitespace characters.
func leadingWhitespaceCount(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ' ' || ch == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}

// ========== Paragraph Operations ==========

// ParagraphCount returns the number of paragraphs (separated by blank lines).
func (r *Rope) ParagraphCount() int {
	content := strings.Trim(r.String(), "\n\r")
	if content == "" {
		return 0
	}

	// Split by double newlines
	paragraphs := strings.Split(content, "\n\n")
	return len(paragraphs)
}

// Paragraph returns the text of the specified paragraph.
// Panics if paraNum is out of bounds.
func (r *Rope) Paragraph(paraNum int) string {
	content := strings.Trim(r.String(), "\n\r")
	paragraphs := strings.Split(content, "\n\n")

	if paraNum < 0 || paraNum >= len(paragraphs) {
		panic("paragraph number out of bounds")
	}

	return paragraphs[paraNum]
}
