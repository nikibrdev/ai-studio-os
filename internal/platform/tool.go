package platform

import "context"

// ToolDescriptor declares what a tool is: its unique name and purpose.
// The full declaration format (parameter schema, constraints, idempotency
// marker) is designed in v0.8 (docs/architecture/tools.md) and will extend
// this contract, not change it.
type ToolDescriptor interface {
	// Name returns the unique tool name used for explicit registration.
	Name() string

	// Description returns the human-readable purpose of the tool.
	Description() string
}

// Tool is one action in the external environment (git, files, checks)
// available to agents through the Tool Layer
// (docs/architecture/interfaces.md, "Tool").
//
// Contract constraints:
//   - a tool performs exactly one action and holds no state between calls;
//   - a tool contains no platform domain logic and never calls Core
//     (Tool -> Core is forbidden, ADR-014); the action result is returned
//     to the caller, state changes happen only through the platform process;
//   - tools are registered explicitly — no discovery by convention;
//   - every invocation is journaled as an event by the platform.
type Tool interface {
	// Descriptor returns the declaration of this tool.
	Descriptor() ToolDescriptor

	// Invoke executes the action with the given parameters and returns its
	// result. The parameter and result shapes are placeholders until the
	// v0.8 Tool Layer design fixes the declaration format.
	Invoke(ctx context.Context, params map[string]any) (any, error)
}
