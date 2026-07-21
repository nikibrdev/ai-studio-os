package execution

// State is the Execution lifecycle state (spec Lifecycle:
// Queued -> Running -> Succeeded | Failed | Aborted, plus the direct
// Queued -> Aborted path).
type State string

// The five Lifecycle states (spec Lifecycle).
const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateAborted   State = "aborted"
)

// Terminal reports whether the state is one of the three mutually
// exclusive terminal states. After entering any of them the Execution
// never changes again (spec Behavioral Invariants 1-2).
func (s State) Terminal() bool {
	return s == StateSucceeded || s == StateFailed || s == StateAborted
}
