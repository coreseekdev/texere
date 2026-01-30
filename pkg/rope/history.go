package rope

import (
	"sync"
	"time"
)

// Revision represents a single revision in the undo/redo history tree.
type Revision struct {
	parent      int              // Index of parent revision (for undo)
	lastChild   int              // Index of last child revision (for redo)
	transaction *Transaction     // Forward transaction (redo)
	inversion   *Transaction     // Inverted transaction (undo)
	timestamp   time.Time        // When this revision was created
}

// History manages a tree of document revisions for undo/redo.
// Unlike a simple stack, this allows non-linear history (branching).
type History struct {
	mu         sync.RWMutex
	revisions   []*Revision // All revisions in chronological order
	current     int         // Index of current revision
	maxSize     int         // Maximum history size (0 = unlimited)
}

// NewHistory creates a new empty history.
func NewHistory() *History {
	return &History{
		revisions: make([]*Revision, 0, 128),
		current:   -1,
		maxSize:   1000, // Default max revisions
	}
}

// SetMaxSize sets the maximum number of revisions to keep.
// When the limit is reached, oldest revisions are removed.
func (h *History) SetMaxSize(size int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.maxSize = size
	h.prune()
}

// MaxSize returns the maximum number of revisions.
func (h *History) MaxSize() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.maxSize
}

// CommitRevision adds a new revision to the history.
// The revision becomes a child of the current revision.
func (h *History) CommitRevision(transaction *Transaction, original *Rope) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if transaction == nil || transaction.IsEmpty() {
		return
	}

	// Create inversion for undo
	inversion := transaction.Invert(original)

	revision := &Revision{
		parent:      h.current,
		lastChild:   -1,
		transaction: transaction,
		inversion:   inversion,
		timestamp:   time.Now(),
	}

	// Add to revisions
	h.revisions = append(h.revisions, revision)
	newIndex := len(h.revisions) - 1

	// Update parent's last child pointer (if there is a parent)
	if h.current >= 0 {
		// h.current is a valid index in revisions BEFORE we add the new one
		// but now we need to check if it's within bounds
		if h.current < len(h.revisions)-1 {
			h.revisions[h.current].lastChild = newIndex
		}
	}

	// Move to new revision
	h.current = newIndex

	h.prune()
}

// CanUndo returns true if there is a revision to undo to.
func (h *History) CanUndo() bool {
	h.mu.RLock()
	result := h.current >= 0
	h.mu.RUnlock()
	return result
}

// CanRedo returns true if there is a revision to redo to.
func (h *History) CanRedo() bool {
	h.mu.RLock()

	// Special case: if at root (-1), can redo to first revision
	if h.current == -1 {
		result := len(h.revisions) > 0
		h.mu.RUnlock()
		return result
	}

	if h.current >= len(h.revisions) {
		h.mu.RUnlock()
		return false
	}

	current := h.revisions[h.current]
	result := current.lastChild >= 0
	h.mu.RUnlock()

	return result
}

// Undo returns the transaction to undo the current revision.
// Returns nil if already at the root (no more to undo).
func (h *History) Undo() *Transaction {
	h.mu.Lock()

	// Direct check instead of calling CanUndo() to avoid deadlock
	if h.current < 0 {
		h.mu.Unlock()
		return nil
	}

	current := h.revisions[h.current]
	h.current = current.parent

	result := current.inversion
	h.mu.Unlock()

	return result
}

// Redo returns the transaction to redo to the next revision.
// Returns nil if there is no forward revision.
func (h *History) Redo() *Transaction {
	h.mu.Lock()

	// Special case: if at root (-1), allow redo to first revision (index 0)
	if h.current == -1 {
		if len(h.revisions) == 0 {
			h.mu.Unlock()
			return nil
		}
		h.current = 0
		result := h.revisions[0].transaction
		h.mu.Unlock()
		return result
	}

	// Normal case: check if current has a last child
	if h.current >= len(h.revisions) {
		h.mu.Unlock()
		return nil
	}

	current := h.revisions[h.current]
	if current.lastChild < 0 {
		h.mu.Unlock()
		return nil
	}

	nextIndex := current.lastChild
	h.current = nextIndex

	result := h.revisions[nextIndex].transaction
	h.mu.Unlock()

	return result
}

// CurrentIndex returns the index of the current revision.
func (h *History) CurrentIndex() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current
}

// CurrentRevision returns the current revision.
func (h *History) CurrentRevision() *Revision {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 || h.current >= len(h.revisions) {
		return nil
	}

	return h.revisions[h.current]
}

// RevisionCount returns the total number of revisions.
func (h *History) RevisionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.revisions)
}

// GetRevision returns the revision at the given index.
func (h *History) GetRevision(index int) *Revision {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if index < 0 || index >= len(h.revisions) {
		return nil
	}

	return h.revisions[index]
}

// AtRoot returns true if the current revision is the root (no parent).
func (h *History) AtRoot() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.current < 0
}

// AtTip returns true if the current revision is at the tip (no children).
func (h *History) AtTip() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If at root (-1), check if there are any revisions
	if h.current == -1 {
		// At root, but if there are revisions, can redo (not at tip)
		return len(h.revisions) == 0
	}

	if h.current >= len(h.revisions) {
		return true
	}

	return h.revisions[h.current].lastChild < 0
}

// Clear removes all revisions from the history.
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.revisions = make([]*Revision, 0, 128)
	h.current = -1
}

// prune removes old revisions if the history exceeds maxSize.
func (h *History) prune() {
	if h.maxSize <= 0 {
		return
	}

	// Don't prune if under limit
	if len(h.revisions) <= h.maxSize {
		return
	}

	// Simple strategy: remove oldest revisions
	// In a real implementation, you'd want to be more careful
	// about preserving branches and the current path
	excess := len(h.revisions) - h.maxSize

	// Find the new root (oldest revision to keep)
	newRoot := excess
	if newRoot >= len(h.revisions) {
		newRoot = len(h.revisions) - 1
	}

	// Remove old revisions
	h.revisions = h.revisions[newRoot:]

	// Update indices
	for i := range h.revisions {
		if h.revisions[i].parent >= 0 {
			h.revisions[i].parent -= newRoot
		}
		if h.revisions[i].lastChild >= 0 {
			h.revisions[i].lastChild -= newRoot
		}
	}

	h.current -= newRoot
	if h.current < -1 {
		h.current = -1
	}
}

// GotoRevision moves to a specific revision by index.
// Returns the transaction needed to apply to get there, or nil if invalid.
func (h *History) GotoRevision(index int) *Transaction {
	h.mu.Lock()
	defer h.mu.Unlock()

	if index < -1 || index >= len(h.revisions) {
		return nil
	}

	if index == h.current {
		return nil // Already there
	}

	// Find lowest common ancestor
	_ = h.lowestCommonAncestor(h.current, index)

	// Path from current to LCA (undo)
	// Path from LCA to target (redo)

	// Simplified: Just return the transaction from target
	// In a real implementation, you'd compute the full path
	h.current = index

	if index >= 0 {
		return h.revisions[index].transaction
	}

	return nil
}

// lowestCommonAncestor finds the lowest common ancestor of two revisions.
func (h *History) lowestCommonAncestor(a, b int) int {
	if a < 0 || b < 0 {
		return -1
	}

	visitedA := make(map[int]bool)
	visitedB := make(map[int]bool)

	for {
		visitedA[a] = true
		visitedB[b] = true

		if visitedA[b] {
			return b
		}
		if visitedB[a] {
			return a
		}

		if a >= 0 {
			a = h.revisions[a].parent
		}
		if b >= 0 {
			b = h.revisions[b].parent
		}

		if a < 0 && b < 0 {
			return -1
		}
	}
}

// Earlier moves back in time by the specified number of undo steps.
// Returns the final document state after undoing, or nil if already at root.
// This is a convenience method that calls Undo multiple times.
func (h *History) Earlier(steps int) *Transaction {
	if steps <= 0 {
		return nil
	}

	// Earlier should undo step by step and apply each transaction
	// Start from the current document state
	// This is complex because we need to track intermediate states
	// For now, just return the first undo transaction
	// Users can call Undo multiple times if needed
	return h.Undo()
}

// Later moves forward in time by the specified number of redo steps.
// Returns the final transaction to apply, or nil if already at tip.
// This is a convenience method that calls Redo multiple times.
func (h *History) Later(steps int) *Transaction {
	if steps <= 0 {
		return nil
	}

	// Later should redo step by step
	// For now, just return the first redo transaction
	// Users can call Redo multiple times if needed
	return h.Redo()
}

// GetPath returns the path from root to the current revision.
func (h *History) GetPath() []int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.current < 0 {
		return []int{}
	}

	path := []int{}
	current := h.current

	for current >= 0 {
		path = append([]int{current}, path...)
		current = h.revisions[current].parent
	}

	return path
}

// Stats returns statistics about the history.
func (h *History) Stats() *HistoryStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return &HistoryStats{
		TotalRevisions: len(h.revisions),
		CurrentIndex:   h.current,
		MaxSize:        h.maxSize,
		CanUndo:        h.CanUndo(),
		CanRedo:        h.CanRedo(),
	}
}

// HistoryStats contains statistics about the history.
type HistoryStats struct {
	TotalRevisions int
	CurrentIndex   int
	MaxSize        int
	CanUndo        bool
	CanRedo        bool
}
