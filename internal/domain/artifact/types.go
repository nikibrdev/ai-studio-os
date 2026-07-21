package artifact

// Type classifies what an Artifact represents (spec Structural Invariant 1).
// Fixed at creation; never redefined by another value during the
// Artifact's life.
type Type string

// Origin describes how an Artifact entered the system (spec Structural
// Invariant 3). The specification leaves the full enumeration of values an
// open question; these three are the examples it names explicitly. Origin
// is intentionally left an open string type rather than a closed enum for
// that reason.
type Origin string

// The three Origin values the specification names explicitly (Examples).
const (
	OriginProduced Origin = "produced" // произведён исполнением
	OriginImported Origin = "imported" // импортирован
	OriginUploaded Origin = "uploaded" // загружен человеком
)

// Author identifies who is responsible for creating or last meaningfully
// changing the Artifact (spec Structural Invariant 3). AuthorUnknown is a
// legitimate value of Author, distinct from Origin: "author unknown" does
// not mean "origin unknown."
type Author string

// AuthorUnknown is the explicit "no responsible author" value (spec
// Structural Invariant 3) — never the zero value used implicitly.
const AuthorUnknown Author = "unknown"

// State is the Artifact lifecycle state (spec Lifecycle: Draft -> Published
// -> Archived).
type State string

// The three Lifecycle states (spec Lifecycle).
const (
	StateDraft     State = "draft"
	StatePublished State = "published"
	StateArchived  State = "archived"
)
