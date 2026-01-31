package rope

// Selection represents a collection of selection ranges.
// It always contains at least one range.
type Selection struct {
	ranges       []Range
	primaryIndex int
}

// NewSelection creates a new Selection with a single Range.
func NewSelection(ranges ...Range) *Selection {
	if len(ranges) == 0 {
		// A selection must have at least one range
		ranges = []Range{Point(0)}
	}
	return &Selection{
		ranges:       ranges,
		primaryIndex: 0,
	}
}

// NewSelectionWithPrimary creates a new Selection with the specified primary index.
func NewSelectionWithPrimary(ranges []Range, primaryIndex int) *Selection {
	if len(ranges) == 0 {
		ranges = []Range{Point(0)}
	}
	if primaryIndex < 0 || primaryIndex >= len(ranges) {
		primaryIndex = 0
	}
	return &Selection{
		ranges:       ranges,
		primaryIndex: primaryIndex,
	}
}

// Primary returns the primary (active) selection range.
func (s *Selection) Primary() Range {
	if s.primaryIndex >= 0 && s.primaryIndex < len(s.ranges) {
		return s.ranges[s.primaryIndex]
	}
	if len(s.ranges) > 0 {
		return s.ranges[0]
	}
	return Point(0)
}

// PrimaryIndex returns the index of the primary selection.
func (s *Selection) PrimaryIndex() int {
	return s.primaryIndex
}

// Len returns the number of ranges in the selection.
func (s *Selection) Len() int {
	return len(s.ranges)
}

// Iter returns an iterator over the selection ranges.
func (s *Selection) Iter() []Range {
	return s.ranges
}

// Add adds a range to the selection.
func (s *Selection) Add(r Range) {
	s.ranges = append(s.ranges, r)
}

// SetPrimary sets the primary selection index.
func (s *Selection) SetPrimary(index int) {
	if index >= 0 && index < len(s.ranges) {
		s.primaryIndex = index
	}
}

// ========== Position Mapping Integration ==========

// MapPositions maps all cursor positions in the selection through a changeset.
// Returns a new selection with all positions mapped.
func (s *Selection) MapPositions(cs *ChangeSet) *Selection {
	if s == nil || len(s.ranges) == 0 {
		return s
	}

	positions := s.GetPositions()
	assocs := s.GetAssociations()

	mapped := MapPositionsOptimized(cs, positions, assocs)

	return s.FromPositions(mapped)
}

// GetPositions returns all cursor positions from the selection ranges.
// Uses the Cursor() method for each range, which represents the actual cursor position.
func (s *Selection) GetPositions() []int {
	positions := make([]int, len(s.ranges))
	for i, r := range s.ranges {
		positions[i] = r.Cursor()
	}
	return positions
}

// GetAssociations returns default associations for all positions.
// Currently returns AssocBefore for all positions.
func (s *Selection) GetAssociations() []Assoc {
	assocs := make([]Assoc, len(s.ranges))
	for i := range assocs {
		assocs[i] = AssocBefore
	}
	return assocs
}

// FromPositions creates a new selection from a slice of positions.
// Each position becomes a single-point cursor (Range with Anchor == Head).
// Preserves the primary index from the original selection.
func (s *Selection) FromPositions(positions []int) *Selection {
	if len(positions) == 0 {
		return NewSelection(Point(0))
	}

	newRanges := make([]Range, len(positions))
	for i, pos := range positions {
		newRanges[i] = Point(pos)
	}

	primaryIdx := s.primaryIndex
	if primaryIdx >= len(newRanges) {
		primaryIdx = 0
	}

	return &Selection{
		ranges:       newRanges,
		primaryIndex: primaryIdx,
	}
}
