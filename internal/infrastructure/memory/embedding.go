package memory

import (
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

// embeddingDim is the fixed vector dimensionality (ADR-018).
const embeddingDim = 256

// embed computes a naive, deterministic embedding for text via feature
// hashing (the "hashing trick"): each token is hashed into one of
// embeddingDim buckets, incremented or decremented with a sign derived
// from a separate hash bit (the signed variant reduces collision bias
// versus plain unsigned hashing), and the result is L2-normalized.
//
// Not semantic — a documented limitation (ADR-018), accepted because no
// embedding provider is available (no key, no new dependency) and the
// rest of the Memory System infrastructure is real and worth proving end
// to end now; swapping in a real model later only touches this function.
func embed(text string) []float32 {
	vector := make([]float32, embeddingDim)

	for _, token := range tokenize(text) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(token))
		sum := h.Sum32()

		index := int(sum % uint32(embeddingDim))
		if sum&(1<<31) != 0 {
			vector[index]++
		} else {
			vector[index]--
		}
	}

	normalize(vector)
	return vector
}

// tokenize lowercases text and splits it on runs of characters that are
// neither letters nor digits.
func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

// normalize scales vector to unit length (L2), in place. A vector of all
// zeros (empty text) is left as-is.
func normalize(vector []float32) {
	var sumSquares float64
	for _, v := range vector {
		sumSquares += float64(v) * float64(v)
	}
	if sumSquares == 0 {
		return
	}

	norm := float32(math.Sqrt(sumSquares))
	for i := range vector {
		vector[i] /= norm
	}
}
