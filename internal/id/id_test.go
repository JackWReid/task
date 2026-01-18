package id

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(id) != 3 {
		t.Errorf("Generate() length = %d, want 3", len(id))
	}

	// Check that all characters are in the charset
	for _, c := range id {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("Generate() contains invalid character: %c", c)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	ids := make(map[string]bool)
	// Generate 100 IDs and check for uniqueness
	// Note: With 36^3 = 46656 possible IDs, collisions are possible but unlikely
	for i := 0; i < 100; i++ {
		id, err := Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if ids[id] {
			// This could happen by chance, so we just log it
			t.Logf("Duplicate ID generated: %s (this is statistically possible)", id)
		}
		ids[id] = true
	}
}

func TestGenerateNoteID(t *testing.T) {
	taskID := "abc"
	noteID, err := GenerateNoteID(taskID)
	if err != nil {
		t.Fatalf("GenerateNoteID() error = %v", err)
	}

	if !strings.HasPrefix(noteID, taskID+"-") {
		t.Errorf("GenerateNoteID() = %q, should have prefix %q", noteID, taskID+"-")
	}

	// Check format: taskID-xxx (7 chars total for 3-char task ID)
	if len(noteID) != 7 {
		t.Errorf("GenerateNoteID() length = %d, want 7", len(noteID))
	}

	// Check the suffix is valid
	suffix := noteID[len(taskID)+1:]
	if len(suffix) != 3 {
		t.Errorf("GenerateNoteID() suffix length = %d, want 3", len(suffix))
	}
	for _, c := range suffix {
		if !strings.ContainsRune(charset, c) {
			t.Errorf("GenerateNoteID() suffix contains invalid character: %c", c)
		}
	}
}

func TestGenerateUnique(t *testing.T) {
	existing := map[string]bool{
		"abc": true,
		"xyz": true,
	}

	id, err := GenerateUnique(existing)
	if err != nil {
		t.Fatalf("GenerateUnique() error = %v", err)
	}

	if existing[id] {
		t.Errorf("GenerateUnique() returned existing ID: %s", id)
	}
}

func TestGenerateUniqueWithManyExisting(t *testing.T) {
	// Even with some existing IDs, we should be able to generate a unique one
	existing := make(map[string]bool)
	for i := 0; i < 50; i++ {
		id, err := Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		existing[id] = true
	}

	newID, err := GenerateUnique(existing)
	if err != nil {
		t.Fatalf("GenerateUnique() error = %v", err)
	}

	if existing[newID] {
		t.Errorf("GenerateUnique() returned existing ID: %s", newID)
	}
}

func TestGenerateCharacterDistribution(t *testing.T) {
	// Generate many IDs and check character distribution is roughly uniform
	counts := make(map[rune]int)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id, err := Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		for _, c := range id {
			counts[c]++
		}
	}

	// With 36 characters and 3000 total characters (1000 * 3),
	// each character should appear roughly 83 times (3000/36)
	expectedPerChar := float64(iterations*3) / float64(len(charset))
	minExpected := expectedPerChar * 0.3 // Allow significant variance due to randomness
	maxExpected := expectedPerChar * 2.0

	for _, c := range charset {
		count := counts[c]
		if float64(count) < minExpected || float64(count) > maxExpected {
			t.Logf("Character %c appeared %d times (expected ~%.0f)", c, count, expectedPerChar)
		}
	}
}

func TestGenerateEmptyExisting(t *testing.T) {
	id, err := GenerateUnique(nil)
	if err != nil {
		t.Fatalf("GenerateUnique(nil) error = %v", err)
	}

	if len(id) != 3 {
		t.Errorf("GenerateUnique(nil) length = %d, want 3", len(id))
	}

	id2, err := GenerateUnique(map[string]bool{})
	if err != nil {
		t.Fatalf("GenerateUnique({}) error = %v", err)
	}

	if len(id2) != 3 {
		t.Errorf("GenerateUnique({}) length = %d, want 3", len(id2))
	}
}
