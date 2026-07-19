// Package shared defines the ubiquitous language of the AI Studio OS
// domain: vocabulary types used by every domain module.
//
// Identifier types, domain errors and further value objects are added here
// as the corresponding decisions are made (identifiers — ADR-011).
// The package contains types and constants only — no logic.
package shared

// Role is a set of responsibilities in the development process, deliberately
// separated from its executor: any role can be performed by a human or by an
// AI agent (docs/architecture/agents.md). Role descriptions live in
// .claude/agents/.
type Role string

// The fixed role vocabulary of the MVP (docs/architecture/domain-model.md).
const (
	RoleProjectManager Role = "project-manager"
	RoleDeveloper      Role = "developer"
	RoleQA             Role = "qa"
	RoleReviewer       Role = "reviewer"
	RoleArchitect      Role = "architect"
)

// TaskState is a state of the canonical task lifecycle. The full state
// machine — transitions, guards and invariants — is defined in
// docs/architecture/state-machine.md; validation logic is implemented by the
// workflow module (Domain Layer), never by callers.
type TaskState string

// The nine canonical task states (docs/architecture/state-machine.md).
const (
	StateBacklog    TaskState = "backlog"
	StateReady      TaskState = "ready"
	StateInProgress TaskState = "in-progress"
	StateReview     TaskState = "review"
	StateTesting    TaskState = "testing"
	StateDone       TaskState = "done"
	StateBlocked    TaskState = "blocked"
	StateCancelled  TaskState = "cancelled"
	StateArchived   TaskState = "archived"
)
