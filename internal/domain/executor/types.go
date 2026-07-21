package executor

// State is the Executor lifecycle state (spec Lifecycle:
// Registered -> Active <-> Disabled -> Retired, plus the direct
// Registered -> Retired path).
type State string

// The four Lifecycle states (spec Lifecycle).
const (
	StateRegistered State = "registered"
	StateActive     State = "active"
	StateDisabled   State = "disabled"
	StateRetired    State = "retired"
)

// Terminal reports whether the state is Retired — the only terminal state.
// A backend that becomes available again is registered as a NEW Executor,
// never by reviving a retired one (spec Behavioral Invariant 1).
func (s State) Terminal() bool { return s == StateRetired }
