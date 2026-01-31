package rope

import "fmt"

// Common error types for rope operations.
// These errors replace panics for better error handling.

// ErrOutOfBounds is returned when an index or position is out of valid range.
type ErrOutOfBounds struct {
	// Operation is the name of the operation that failed (e.g., "Slice", "Insert")
	Operation string
	// Position is the invalid position that was provided
	Position int
	// ValidRange is the valid range [min, max)
	Min int
	Max int
}

func (e *ErrOutOfBounds) Error() string {
	if e.Max == e.Min {
		return fmt.Sprintf("%s: position %d out of bounds (empty rope)", e.Operation, e.Position)
	}
	return fmt.Sprintf("%s: position %d out of bounds (valid range: [%d, %d])", e.Operation, e.Position, e.Min, e.Max)
}

// ErrInvalidRange is returned when a range [start, end) is invalid.
type ErrInvalidRange struct {
	// Operation is the name of the operation that failed
	Operation string
	// Start is the start of the range
	Start int
	// End is the end of the range
	End int
	// ValidMax is the maximum valid end position
	ValidMax int
}

func (e *ErrInvalidRange) Error() string {
	if e.Start > e.End {
		return fmt.Sprintf("%s: invalid range [%d, %d) (start > end)", e.Operation, e.Start, e.End)
	}
	return fmt.Sprintf("%s: range [%d, %d) out of bounds (valid range: [0, %d])", e.Operation, e.Start, e.End, e.ValidMax)
}

// ErrIteratorState is returned when an iterator operation is invalid for its current state.
type ErrIteratorState struct {
	// Operation is the iterator operation that failed
	Operation string
	// Reason describes why the operation failed
	Reason string
}

func (e *ErrIteratorState) Error() string {
	return fmt.Sprintf("iterator %s: %s", e.Operation, e.Reason)
}

// ErrInvalidInput is returned for invalid input parameters.
type ErrInvalidInput struct {
	// Parameter is the parameter name
	Parameter string
	// Value is the invalid value
	Value interface{}
	// Reason describes why it's invalid
	Reason string
}

func (e *ErrInvalidInput) Error() string {
	return fmt.Sprintf("invalid parameter %s: %s (%v)", e.Parameter, e.Reason, e.Value)
}

// Helper functions to create common errors

func errSliceOutOfBounds(start, end, max int) error {
	if start > end {
		return &ErrInvalidRange{
			Operation: "Slice",
			Start:     start,
			End:       end,
			ValidMax:  max,
		}
	}
	if end > max {
		return &ErrInvalidRange{
			Operation: "Slice",
			Start:     start,
			End:       end,
			ValidMax:  max,
		}
	}
	if start < 0 {
		return &ErrInvalidRange{
			Operation: "Slice",
			Start:     start,
			End:       end,
			ValidMax:  max,
		}
	}
	return nil
}

func errCharOutOfBounds(pos, max int) error {
	if pos < 0 || pos >= max {
		return &ErrOutOfBounds{
			Operation: "CharAt",
			Position:  pos,
			Min:       0,
			Max:       max,
		}
	}
	return nil
}

func errByteOutOfBounds(pos, max int) error {
	if pos < 0 || pos >= max {
		return &ErrOutOfBounds{
			Operation: "ByteAt",
			Position:  pos,
			Min:       0,
			Max:       max,
		}
	}
	return nil
}

func errInsertOutOfBounds(pos, max int) error {
	if pos < 0 || pos > max {
		return &ErrOutOfBounds{
			Operation: "Insert",
			Position:  pos,
			Min:       0,
			Max:       max + 1, // Insert allows position at max
		}
	}
	return nil
}

func errDeleteOutOfBounds(start, end, max int) error {
	if start < 0 || end > max || start > end {
		return &ErrInvalidRange{
			Operation: "Delete",
			Start:     start,
			End:       end,
			ValidMax:  max,
		}
	}
	return nil
}

func errSplitOutOfBounds(pos, max int) error {
	if pos < 0 || pos > max {
		return &ErrOutOfBounds{
			Operation: "Split",
			Position:  pos,
			Min:       0,
			Max:       max + 1, // Split allows position at max
		}
	}
	return nil
}

// Common error instances for quick use
var (
	// ErrIteratorExhausted is returned when trying to advance an exhausted iterator
	ErrIteratorExhausted = &ErrIteratorState{
		Operation: "Next",
		Reason:    "iterator exhausted",
	}

	// ErrLengthMismatch is returned when document length doesn't match expected length
	ErrLengthMismatch = &ErrInvalidInput{
		Parameter: "length",
		Reason:    "document length mismatch",
	}
)
