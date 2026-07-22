package memory

import (
	"math"
	"testing"
)

func TestEmbed_IsDeterministic(t *testing.T) {
	a := embed("привет мир")
	b := embed("привет мир")
	if len(a) != embeddingDim {
		t.Fatalf("embed() length = %d, want %d", len(a), embeddingDim)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("embed() is not deterministic at index %d: %v vs %v", i, a[i], b[i])
		}
	}
}

func TestEmbed_DifferentTextsdifferentVectors(t *testing.T) {
	a := embed("постгрес и pgx")
	b := embed("совершенно другой текст про докер")
	same := true
	for i := range a {
		if a[i] != b[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("embed() produced identical vectors for different texts")
	}
}

func TestEmbed_IsUnitNormalized(t *testing.T) {
	v := embed("любой непустой текст с несколькими словами")
	var sumSquares float64
	for _, x := range v {
		sumSquares += float64(x) * float64(x)
	}
	norm := math.Sqrt(sumSquares)
	if math.Abs(norm-1) > 1e-6 {
		t.Errorf("embed() norm = %v, want ~1", norm)
	}
}

func TestEmbed_EmptyTextIsZeroVector(t *testing.T) {
	v := embed("")
	for i, x := range v {
		if x != 0 {
			t.Fatalf("embed(\"\")[%d] = %v, want 0", i, x)
		}
	}
}

func TestTokenize_LowercasesAndSplitsOnPunctuation(t *testing.T) {
	got := tokenize("Postgres, pgx/v5 — драйвер!")
	want := []string{"postgres", "pgx", "v5", "драйвер"}
	if len(got) != len(want) {
		t.Fatalf("tokenize() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tokenize()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
