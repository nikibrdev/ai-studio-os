// Package project defines the contract of the managed-project registry and
// implements the Project entity itself
// (docs/specifications/domain/project.md, status: Утверждена;
// docs/architecture/domain-model.md, "Project").
//
// The connection format of repositories and the projects/ directory layout
// are Decision Required (ADR-013); this contract fixes only the operations
// independent of that decision. Identifiers are strings until ADR-011.
package project

import "context"

// Registry manages the lifecycle of projects the platform develops
// (Lifecycle: Created -> Active -> Archived; an archived project is
// immutable).
type Registry interface {
	// Create registers a new project and returns its identifier.
	Create(ctx context.Context, name string) (string, error)

	// ConnectRepository attaches a git repository to the project. Every
	// active project has at least one repository
	// (docs/architecture/domain-model.md). Connecting never transitions
	// state by itself — it only satisfies Activate's guard condition.
	ConnectRepository(ctx context.Context, projectID, repo string) error

	// Activate moves the project from Created to Active. Allowed only
	// when at least one repository is connected (the spec's explicit
	// Activate command — stage-2 contract extension decided by the final
	// architecture review, not an implicit side effect of
	// ConnectRepository).
	Activate(ctx context.Context, projectID string) error

	// Archive moves the project to the immutable Archived state.
	Archive(ctx context.Context, projectID string) error
}
