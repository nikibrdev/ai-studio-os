// Package project defines the contract of the managed-project registry
// (docs/architecture/domain-model.md, "Project").
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
	// project has at least one repository (docs/architecture/domain-model.md).
	ConnectRepository(ctx context.Context, projectID, repo string) error

	// Archive moves the project to the immutable Archived state.
	Archive(ctx context.Context, projectID string) error
}
