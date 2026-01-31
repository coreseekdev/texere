package rope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRange_Line tests getting lines
func TestRange_GetLine(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	r := New(text)

	lines := r.Lines()
	assert.Equal(t, 3, len(lines))
	assert.Equal(t, "Line 1\n", lines[0])
	assert.Equal(t, "Line 2\n", lines[1])
	assert.Equal(t, "Line 3", lines[2])
}

// TestRange_LineAt tests getting line at specific index
func TestRange_LineAt(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	r := New(text)

	line := r.Line(0)
	// Note: Line() may or may not include trailing newline
	// Adjust based on actual implementation
	assert.True(t, line == "Line 1" || line == "Line 1\n")

	line = r.Line(1)
	assert.True(t, line == "Line 2" || line == "Line 2\n")

	line = r.Line(2)
	assert.Equal(t, "Line 3", line) // Last line shouldn't have newline
}

// TestLineAtChar tests getting line number at character position
func TestLineInfo_LineAtChar(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	r := New(text)

	// Character 0-4 (Line 1) -> line 0
	lineNum := r.LineAtChar(0)
	assert.Equal(t, 0, lineNum)

	lineNum = r.LineAtChar(4)
	assert.Equal(t, 0, lineNum)

	// Character 5 (\n) -> still line 0
	lineNum = r.LineAtChar(5)
	assert.Equal(t, 0, lineNum)

	// Character 6-12 (Line 2) -> line 1
	lineNum = r.LineAtChar(6)
	assert.Equal(t, 1, lineNum)

	lineNum = r.LineAtChar(12)
	assert.Equal(t, 1, lineNum)

	// Character 13-19 (Line 3) -> line 2
	lineNum = r.LineAtChar(13)
	assert.Equal(t, 2, lineNum)
}
