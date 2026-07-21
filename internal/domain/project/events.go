package project

import "time"

// Created is the data of the event published when a Project is registered
// (enters Created) (spec Domain Events: ProjectCreated).
type Created struct {
	ID   string
	Name string
	At   time.Time
}

// RepositoryConnected is the data of the event published on every
// repository connection, not only the first — an operational fact
// meaningful beyond the Project itself (spec Domain Events).
type RepositoryConnected struct {
	ProjectID  string
	Repository string
	At         time.Time
}

// Activated is the data of the event published on Created -> Active,
// caused by the explicit Activate command (spec Domain Events:
// ProjectActivated).
type Activated struct {
	ID string
	At time.Time
}

// Archived is the data of the event published on Active -> Archived (spec
// Domain Events: ProjectArchived).
type Archived struct {
	ID string
	At time.Time
}
