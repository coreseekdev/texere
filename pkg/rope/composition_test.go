package rope

import (
	"testing"
)

// TestCompose_Basic tests basic composition following Helix's approach
func TestCompose_Basic(t *testing.T) {
	// Create a document
	text := "hello xz"  // 8 chars
	doc := New(text)

	// First changeset: insert " test!" after "hello", delete "xz"
	cs1 := NewChangeSet(doc.Length()).
		Retain(5).               // "hello"
	Insert(" test!").        // 6 chars
	Retain(1).               // " "
	Delete(2).               // "xz"
	Insert("abc")             // 3 chars

	// Expected: lenBefore=8, lenAfter=15
	if cs1.LenBefore() != 8 {
		t.Errorf("Expected lenBefore 8, got %d", cs1.LenBefore())
	}
	if cs1.LenAfter() != 15 {
		t.Errorf("Expected lenAfter 15, got %d", cs1.LenAfter())
	}

	// Apply cs1 to verify
	result1 := cs1.Apply(doc)
	expected1 := "hello test! abc"
	if result1.String() != expected1 {
		t.Errorf("cs1.Apply: expected %q, got %q", expected1, result1.String())
	}

	// Second changeset: delete " test!" and insert "世"
	// Note: cs2.len_before must equal cs1.len_after (15)
	cs2 := NewChangeSet(cs1.LenAfter()).
		Delete(10).              // " test!" (6 chars + 2 chars = 8, but need to account)
		Insert("世")             // Just insert
		// Wait, let me recalculate...

	// Actually, let me use a simpler example
	cs2 = NewChangeSet(cs1.LenAfter()).
		Delete(8).               // " test!ab" (8 chars)
		Insert("world")          // 5 chars

	// Compose
	composed := cs1.Compose(cs2)

	// Apply composed to original document
	result := composed.Apply(doc)

	// The result should be equivalent to applying cs1 then cs2
	expected := "helloworld c"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestCompose_InsertInsert tests composition of two insert operations
func TestCompose_InsertInsert(t *testing.T) {
	doc := New("ab")
	cs1 := NewChangeSet(doc.Length()).Retain(1).Insert("x")
	cs2 := NewChangeSet(cs1.LenAfter()).Retain(2).Insert("y")

	composed := cs1.Compose(cs2)
	result := composed.Apply(doc)

	expected := "axyb"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestCompose_DeleteDelete tests composition of delete operations
func TestCompose_DeleteDelete(t *testing.T) {
	doc := New("abcd")
	cs1 := NewChangeSet(doc.Length()).Retain(1).Delete(1)
	cs2 := NewChangeSet(cs1.LenAfter()).Retain(1).Delete(1)

	composed := cs1.Compose(cs2)
	result := composed.Apply(doc)

	expected := "ad"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}

// TestCompose_Empty tests composition with empty changesets
func TestCompose_Empty(t *testing.T) {
	doc := New("hello")
	cs1 := NewChangeSet(doc.Length()).Retain(5).Insert(" world")
	cs2 := NewChangeSet(cs1.LenAfter()) // Empty

	composed := cs1.Compose(cs2)
	result := composed.Apply(doc)

	if result.String() != "hello world" {
		t.Errorf("Expected 'hello world', got %q", result.String())
	}
}

// TestInvert_Basic tests basic invert functionality
func TestInvert_Basic(t *testing.T) {
	doc := New("hello world")
	cs := NewChangeSet(doc.Length()).
		Retain(6).
		Delete(5).
		Insert("gophers")

	// Apply changeset
	modified := cs.Apply(doc)
	expectedModified := "hello gophers"
	if modified.String() != expectedModified {
		t.Fatalf("Apply: expected %q, got %q", expectedModified, modified.String())
	}

	// Invert and apply to get back original
	inverted := cs.Invert(doc)
	restored := inverted.Apply(modified)

	if restored.String() != doc.String() {
		t.Errorf("Invert: expected %q, got %q", doc.String(), restored.String())
	}
}

// TestCompose_Optimization tests if composition optimizes operations
func TestCompose_Optimization(t *testing.T) {
	doc := New("hi")

	// Multiple small insert operations
	cs1 := NewChangeSet(doc.Length()).Insert("a")
	cs2 := NewChangeSet(cs1.LenAfter()).Insert("b")
	cs3 := NewChangeSet(cs2.LenAfter()).Insert("c")

	composed := cs1.Compose(cs2).Compose(cs3)

	// Should optimize to a single Insert("abc")
	// (depending on fusion implementation)
	result := composed.Apply(doc)

	expected := "abc"
	if result.String() != expected {
		t.Errorf("Expected %q, got %q", expected, result.String())
	}
}
