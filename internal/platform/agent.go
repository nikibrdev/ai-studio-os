package platform

import "context"

// Request is the input of one agent execution.
//
// Deliberately abstract: the exchange format between the platform and agents
// (what an assignment carries, how context is delivered) is fixed by ADR-005
// (Decision Required). Introducing fields or methods here before that
// decision would pre-empt it.
type Request any

// Response is the output of one agent execution.
//
// Deliberately abstract until ADR-005 fixes the exchange format (report
// structure, artifacts, progress reporting).
type Response any

// Agent is the contract every AI-provider adapter in agents/ implements
// (docs/architecture/interfaces.md, "Agent"). The platform core knows only
// this contract and never a concrete provider.
//
// Contract constraints (ADR-014):
//   - an agent never accesses platform storage (Agent -> Database is
//     forbidden); all effects go through tools and the platform process;
//   - an agent must stay within the scope of its request;
//   - an adapter contains no platform domain logic.
type Agent interface {
	// Execute performs the request and returns its response. The call
	// blocks until the execution finishes; cancellation is requested
	// through ctx. The concrete shapes of Request and Response are fixed
	// by ADR-005.
	Execute(ctx context.Context, req Request) (Response, error)
}
