package rope

import (
	"unicode/utf8"
)

// ========== Optimized Append/Prepend ==========

// AppendOptimized appends text to the end of the rope.
// Optimized version that directly creates a new node instead of using Insert().
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) AppendOptimized(text string) *Rope {
	if r == nil {
		return New(text)
	}
	if text == "" {
		return r
	}
	if r.length == 0 {
		return New(text)
	}

	// Create rope from text and append it
	textRope := New(text)

	return &Rope{
		root: &InternalNode{
			left:   r.root,
			right:  textRope.root,
			length: r.Length(),
			size:   r.Size(),
		},
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}

// PrependOptimized prepends text to the beginning of the rope.
// Optimized version that directly creates a new node instead of using Insert().
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) PrependOptimized(text string) *Rope {
	if r == nil {
		return New(text)
	}
	if text == "" {
		return r
	}
	if r.length == 0 {
		return New(text)
	}

	// Create rope from text and prepend it
	textRope := New(text)

	return &Rope{
		root: &InternalNode{
			left:   textRope.root,
			right:  r.root,
			length: textRope.Length(),
			size:   textRope.Size(),
		},
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}
