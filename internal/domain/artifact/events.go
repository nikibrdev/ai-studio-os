package artifact

import "time"

// Created is the data of the event published when an Artifact is
// registered (enters Draft) (spec Domain Events: ArtifactCreated).
type Created struct {
	ID         string
	ProjectID  string
	Type       Type
	Origin     Origin
	Author     Author
	ProducedBy string // empty if this Artifact was not introduced by an Execution
	At         time.Time
}

// Published is the data of the event published when an Artifact
// transitions Draft -> Published (spec Domain Events: ArtifactPublished).
type Published struct {
	ID         string
	ProducedBy string // empty if none
	At         time.Time
}

// Archived is the data of the event published when an Artifact transitions
// into Archived, from either Draft or Published — one event for both paths,
// since the resulting state is the same (spec Domain Events:
// ArtifactArchived).
type Archived struct {
	ID   string
	From State // the state the Artifact was archived from: Draft or Published
	At   time.Time
}
