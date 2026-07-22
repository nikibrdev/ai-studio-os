// Package memory implements platform.MemoryProvider (v0.7, EPIC-007):
// project knowledge stored as durable, human-readable files under
// memory/<projectID>/<id>.md (the source of truth, per memory.md's
// "transparency" principle) with a rebuildable Qdrant index on top for
// search (naive local embedding — ADR-018).
//
// memory/ at the repository root is data, not code
// (module-boundaries.md) — this package is the code that reads and
// writes it.
package memory
