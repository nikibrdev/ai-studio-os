import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import DecisionsPage from "./page";
import * as api from "@/lib/api";
import type { AwaitingDecision, DecisionKind, TaskView } from "@/lib/api";

vi.mock("@/lib/api", async () => {
  // decisionLabels is real data, not a dependency to fake: mocking it would
  // let the test pass while the page showed the wrong wording.
  const actual = await vi.importActual<typeof api>("@/lib/api");
  return { ...actual, listAwaitingDecision: vi.fn() };
});

function task(overrides: Partial<TaskView> = {}): TaskView {
  return {
    id: "TASK-001",
    projectId: "proj-1",
    state: "backlog",
    updatedAt: "2026-07-26T00:00:00Z",
    title: "Задача",
    type: "feature",
    scope: "",
    acceptanceCriteria: [],
    awaitingDecision: "definition-of-ready",
    repository: "",
    pullRequestId: "",
    ...overrides,
  };
}

function awaiting(
  decision: Exclude<DecisionKind, "">,
  overrides: Partial<TaskView> = {},
): AwaitingDecision {
  return { decision, task: task({ awaitingDecision: decision, ...overrides }) };
}

describe("DecisionsPage", () => {
  it("lists each task awaiting a decision, linking to the task", async () => {
    vi.mocked(api.listAwaitingDecision).mockResolvedValue([
      awaiting("definition-of-ready", { id: "TASK-001", title: "Первая" }),
    ]);

    render(await DecisionsPage());

    expect(
      screen.getByRole("link", { name: /TASK-001: Первая/ }),
    ).toHaveAttribute("href", "/projects/proj-1/tasks/TASK-001");
  });

  // The two decisions must read differently — otherwise a human cannot tell
  // "accept Definition of Ready" from "make the final acceptance call", and the
  // second one merges a pull request.
  it("distinguishes the two kinds of decision", async () => {
    vi.mocked(api.listAwaitingDecision).mockResolvedValue([
      awaiting("definition-of-ready", { id: "TASK-001" }),
      awaiting("acceptance", { id: "TASK-002", state: "testing" }),
    ]);

    render(await DecisionsPage());

    expect(screen.getByText("Принять Definition of Ready")).toBeInTheDocument();
    expect(screen.getByText("Приёмочное решение")).toBeInTheDocument();
  });

  it("shows an empty list as a normal state, not a blank screen", async () => {
    vi.mocked(api.listAwaitingDecision).mockResolvedValue([]);

    render(await DecisionsPage());

    expect(
      screen.getByText("Задач, ожидающих решения, нет."),
    ).toBeInTheDocument();
  });

  // Tasks from different projects share task ids (TASK-NNN is unique only
  // within a project — ADR-011/BUGFIX-003), so the list must key and label them
  // by the pair, or two projects' TASK-001 would collide.
  it("keeps same-numbered tasks from different projects apart", async () => {
    vi.mocked(api.listAwaitingDecision).mockResolvedValue([
      awaiting("definition-of-ready", { id: "TASK-001", projectId: "proj-a" }),
      awaiting("definition-of-ready", { id: "TASK-001", projectId: "proj-b" }),
    ]);

    render(await DecisionsPage());

    expect(screen.getAllByRole("link", { name: /TASK-001/ })).toHaveLength(2);
    expect(screen.getByText("(proj-a)")).toBeInTheDocument();
    expect(screen.getByText("(proj-b)")).toBeInTheDocument();
  });
});
