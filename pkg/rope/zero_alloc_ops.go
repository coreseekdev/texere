package rope

import (
	"unicode/utf8"
)

// ========== Zero-Allocation Operations ==========

// InsertZeroAlloc inserts text with zero allocations for nodes.
// Uses object pooling and copy-on-write for optimal performance.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) InsertZeroAlloc(pos int, text string) *Rope {
	if r == nil {
		return New(text)
	}
	if pos < 0 || pos > r.length {
		panic("insert position out of range")
	}
	if text == "" {
		return r
	}
	if pos == 0 {
		return r.PrependZeroAlloc(text)
	}
	if pos == r.length {
		return r.AppendZeroAlloc(text)
	}

	newRoot := insertNodeZeroAlloc(r.root, pos, text)
	return &Rope{
		root:   newRoot,
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}

// insertNodeZeroAlloc performs zero-allocation insertion.
func insertNodeZeroAlloc(node RopeNode, pos int, text string) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)

		// Use safe string operations
		oldText := leaf.text

		// Fast path: find byte position without allocating
		bytePos := 0
		for i := 0; i < pos; i++ {
			_, size := utf8.DecodeRuneInString(oldText[bytePos:])
			bytePos += size
		}

		// Allocate exact size buffer
		newLen := len(oldText) + len(text)
		newText := make([]byte, newLen)

		// Copy using built-in copy (optimized)
		copy(newText, oldText[:bytePos])
		copy(newText[bytePos:], text)
		copy(newText[bytePos+len(text):], oldText[bytePos:])

		// Use pooled node
		newLeaf := AcquireLeaf()
		newLeaf.text = string(newText)
		return newLeaf
	}

	// Internal node - use pooled nodes
	internal := node.(*InternalNode)
	leftLen := internal.length // Use cached value

	if pos <= leftLen {
		// Insert into left subtree
		newLeft := insertNodeZeroAlloc(internal.left, pos, text)
		newNode := AcquireInternal()
		newNode.left = newLeft
		newNode.right = internal.right // Share right subtree (COW)
		newNode.length = newLeft.Length()
		newNode.size = newLeft.Size()
		return newNode
	}

	// Insert into right subtree
	newRight := insertNodeZeroAlloc(internal.right, pos-leftLen, text)
	newNode := AcquireInternal()
	newNode.left = internal.left // Share left subtree (COW)
	newNode.right = newRight
	newNode.length = internal.left.Length()
	newNode.size = internal.left.Size()
	return newNode
}

// DeleteZeroAlloc removes characters with zero node allocations.
// Uses object pooling and optimized slice operations.
// Returns a new Rope, leaving the original unchanged.
func (r *Rope) DeleteZeroAlloc(start, end int) *Rope {
	if r == nil {
		return r
	}
	if start < 0 || end > r.length || start > end {
		panic("delete range out of bounds")
	}
	if start == end {
		return r
	}

	// Calculate deleted size upfront
	deletedText := r.Slice(start, end)
	deletedLen := utf8.RuneCountInString(deletedText)
	deletedSize := len(deletedText)

	newRoot := deleteNodeZeroAlloc(r.root, start, end)
	return &Rope{
		root:   newRoot,
		length: r.length - deletedLen,
		size:   r.size - deletedSize,
	}
}

// deleteNodeZeroAlloc performs zero-allocation deletion.
func deleteNodeZeroAlloc(node RopeNode, start, end int) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)

		// Use safe string operations instead of unsafe
		oldText := leaf.text

		// Find byte positions safely
		startByte := 0
		for i := 0; i < start; i++ {
			_, size := utf8.DecodeRuneInString(oldText[startByte:])
			startByte += size
		}

		endByte := startByte
		for i := start; i < end; i++ {
			_, size := utf8.DecodeRuneInString(oldText[endByte:])
			endByte += size
		}

		// Allocate exact size buffer
		newLen := len(oldText) - (endByte - startByte)
		newText := make([]byte, newLen)

		// Copy using optimized copy
		copy(newText, oldText[:startByte])
		copy(newText[startByte:], oldText[endByte:])

		// Use pooled node
		newLeaf := AcquireLeaf()
		newLeaf.text = string(newText)
		return newLeaf
	}

	// Internal node
	internal := node.(*InternalNode)
	leftLen := internal.length

	// Entirely in left subtree
	if end <= leftLen {
		newLeft := deleteNodeZeroAlloc(internal.left, start, end)
		if newLeft.Length() == 0 {
			// Left subtree deleted, return right
			return internal.right
		}
		newNode := AcquireInternal()
		newNode.left = newLeft
		newNode.right = internal.right // Share right (COW)
		newNode.length = newLeft.Length()
		newNode.size = newLeft.Size()
		return newNode
	}

	// Entirely in right subtree
	if start >= leftLen {
		newRight := deleteNodeZeroAlloc(internal.right, start-leftLen, end-leftLen)
		if newRight.Length() == 0 {
			// Right subtree deleted, return left
			return internal.left
		}
		newNode := AcquireInternal()
		newNode.left = internal.left // Share left (COW)
		newNode.right = newRight
		newNode.length = internal.left.Length()
		newNode.size = internal.left.Size()
		return newNode
	}

	// Spans both subtrees - optimized merge without Slice()
	return deleteMergeZeroAlloc(internal.left, start, internal.right, end-leftLen)
}

// deleteMergeZeroAlloc merges left and right parts after deletion without using Slice().
func deleteMergeZeroAlloc(left RopeNode, leftStart int, right RopeNode, rightEnd int) RopeNode {
	// Extract parts using direct iteration
	leftPart := extractSuffixZeroAlloc(left, leftStart)
	rightPart := extractPrefixZeroAlloc(right, rightEnd)

	// Concatenate efficiently
	if leftPart == nil || leftPart.Length() == 0 {
		return rightPart
	}
	if rightPart == nil || rightPart.Length() == 0 {
		return leftPart
	}

	newNode := AcquireInternal()
	newNode.left = leftPart
	newNode.right = rightPart
	newNode.length = leftPart.Length()
	newNode.size = leftPart.Size()
	return newNode
}

// extractSuffixZeroAlloc extracts suffix from start without allocations.
func extractSuffixZeroAlloc(node RopeNode, start int) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		if start == 0 {
			return leaf // Return entire leaf (shared)
		}
		if start >= leaf.Length() {
			return nil // Empty
		}

		// Need to split - find byte position
		oldText := leaf.text
		startByte := 0
		for i := 0; i < start; i++ {
			_, size := utf8.DecodeRuneInString(oldText[startByte:])
			startByte += size
		}

		// Create new leaf with suffix
		newLeaf := AcquireLeaf()
		newLeaf.text = oldText[startByte:]
		return newLeaf
	}

	// Internal node
	internal := node.(*InternalNode)
	leftLen := internal.length

	if start < leftLen {
		// Split left subtree
		newLeft := extractSuffixZeroAlloc(internal.left, start)
		if newLeft == nil {
			return internal.right
		}
		newNode := AcquireInternal()
		newNode.left = newLeft
		newNode.right = internal.right
		newNode.length = newLeft.Length()
		newNode.size = newLeft.Size()
		return newNode
	}

	// Entirely in right subtree
	return extractSuffixZeroAlloc(internal.right, start-leftLen)
}

// extractPrefixZeroAlloc extracts prefix up to end without allocations.
func extractPrefixZeroAlloc(node RopeNode, end int) RopeNode {
	if node.IsLeaf() {
		leaf := node.(*LeafNode)
		if end >= leaf.Length() {
			return leaf // Return entire leaf (shared)
		}
		if end == 0 {
			return nil // Empty
		}

		// Need to split - find byte position
		oldText := leaf.text
		endByte := 0
		for i := 0; i < end; i++ {
			_, size := utf8.DecodeRuneInString(oldText[endByte:])
			endByte += size
		}

		// Create new leaf with prefix
		newLeaf := AcquireLeaf()
		newLeaf.text = oldText[:endByte]
		return newLeaf
	}

	// Internal node
	internal := node.(*InternalNode)
	leftLen := internal.length

	if end <= leftLen {
		// Entirely in left subtree
		return extractPrefixZeroAlloc(internal.left, end)
	}

	// Spans both
	newRight := extractPrefixZeroAlloc(internal.right, end-leftLen)
	if newRight == nil {
		return internal.left
	}

	newNode := AcquireInternal()
	newNode.left = internal.left
	newNode.right = newRight
	newNode.length = internal.left.Length()
	newNode.size = internal.left.Size()
	return newNode
}

// AppendZeroAlloc appends text with zero allocations.
func (r *Rope) AppendZeroAlloc(text string) *Rope {
	if r == nil {
		return New(text)
	}
	if text == "" {
		return r
	}
	if r.length == 0 {
		return New(text)
	}

	// Create rope from text
	textRope := New(text)

	// Use pooled node
	newNode := AcquireInternal()
	newNode.left = r.root
	newNode.right = textRope.root
	newNode.length = r.Length()
	newNode.size = r.Size()

	return &Rope{
		root:   newNode,
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}

// PrependZeroAlloc prepends text with zero allocations.
func (r *Rope) PrependZeroAlloc(text string) *Rope {
	if r == nil {
		return New(text)
	}
	if text == "" {
		return r
	}
	if r.length == 0 {
		return New(text)
	}

	// Create rope from text
	textRope := New(text)

	// Use pooled node
	newNode := AcquireInternal()
	newNode.left = textRope.root
	newNode.right = r.root
	newNode.length = textRope.Length()
	newNode.size = textRope.Size()

	return &Rope{
		root:   newNode,
		length: r.length + utf8.RuneCountInString(text),
		size:   r.size + len(text),
	}
}
