package application

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID generates a random hex identifier using only the standard
// library: no UUID dependency is listed in .claude/context/stack.md.
// Used for entities the Application Layer spawns as a side effect rather
// than one an external command names explicitly (Task/Project/Executor
// identifiers are supplied by their creating command; a spawned Execution
// has no such external caller — see internal/application/work.go).
func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand.Read does not fail on any platform this project
		// targets; a zero-value ID would silently violate identity
		// uniqueness, so panic is the honest response to an impossible
		// condition rather than returning a bad ID.
		panic("application: failed to generate id: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}
