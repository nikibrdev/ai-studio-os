import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DecisionActions from "./DecisionActions";
import * as actions from "@/lib/actions";
import type { AwaitingDecision, DecisionKind, TaskView } from "@/lib/api";

vi.mock("@/lib/actions", () => ({
  acceptDefinitionOfReady: vi.fn(),
  acceptTask: vi.fn(),
  returnForRework: vi.fn(),
}));

function awaiting(
  decision: Exclude<DecisionKind, "">,
  overrides: Partial<TaskView> = {},
): AwaitingDecision {
  const task: TaskView = {
    id: "TASK-001",
    projectId: "proj-1",
    state: decision === "acceptance" ? "testing" : "backlog",
    updatedAt: "2026-07-26T00:00:00Z",
    title: "Задача",
    type: "feature",
    scope: "",
    acceptanceCriteria: [],
    awaitingDecision: decision,
    repository: "",
    pullRequestId: "",
    ...overrides,
  };
  return { decision, task };
}

beforeEach(() => {
  // Сбрасывает и реализацию, и счётчики вызовов: без этого счётчик копится
  // между тестами и любая проверка «вызвано N раз» становится ненадёжной.
  vi.clearAllMocks();
  vi.mocked(actions.acceptDefinitionOfReady).mockResolvedValue({ ok: true });
  vi.mocked(actions.acceptTask).mockResolvedValue({ ok: true });
  vi.mocked(actions.returnForRework).mockResolvedValue({ ok: true });
});

describe("DecisionActions", () => {
  it("accepts Definition of Ready for the task it was given", async () => {
    render(<DecisionActions awaiting={awaiting("definition-of-ready")} />);

    await userEvent.click(screen.getByRole("button", { name: "Принять" }));

    expect(actions.acceptDefinitionOfReady).toHaveBeenCalledWith(
      "proj-1",
      "TASK-001",
    );
  });

  // The requirement is normative (workflow.md): the acceptance decision merges a
  // pull request and is irreversible within the task, so the consequence must be
  // stated before the click, not after.
  it("warns that accepting merges, naming the pull request, before any click", () => {
    render(
      <DecisionActions
        awaiting={awaiting("acceptance", {
          repository: "github.com/org/repo",
          pullRequestId: "42",
        })}
      />,
    );

    expect(screen.getByText(/сольёт/)).toBeInTheDocument();
    expect(
      screen.getByText(/Pull Request 42 в github\.com\/org\/repo/),
    ).toBeInTheDocument();
    expect(screen.getByText(/необратимо/)).toBeInTheDocument();
  });

  it("offers both outcomes of the acceptance decision", async () => {
    render(<DecisionActions awaiting={awaiting("acceptance")} />);

    await userEvent.click(
      screen.getByRole("button", { name: "Вернуть на доработку" }),
    );

    expect(actions.returnForRework).toHaveBeenCalledWith("proj-1", "TASK-001");
    expect(actions.acceptTask).not.toHaveBeenCalled();
  });

  // A failed action must not replace the page (error.tsx is for render failures):
  // the person keeps context, sees what went wrong, and can retry.
  it("shows a failure next to the button and allows retrying", async () => {
    vi.mocked(actions.acceptDefinitionOfReady)
      .mockResolvedValueOnce({ ok: false, error: "переход недопустим" })
      .mockResolvedValueOnce({ ok: true });

    render(<DecisionActions awaiting={awaiting("definition-of-ready")} />);
    const button = screen.getByRole("button", { name: "Принять" });

    await userEvent.click(button);

    const alert = await waitFor(() => screen.getByRole("alert"));
    expect(alert).toHaveTextContent("переход недопустим");
    // Still on screen and usable — nothing was replaced.
    expect(button).toBeEnabled();

    await userEvent.click(button);

    await waitFor(() =>
      expect(screen.queryByRole("alert")).not.toBeInTheDocument(),
    );
    expect(actions.acceptDefinitionOfReady).toHaveBeenCalledTimes(2);
  });

  it("reports a network failure rather than staying silent", async () => {
    vi.mocked(actions.acceptTask).mockResolvedValue({
      ok: false,
      error: "fetch failed",
    });

    render(<DecisionActions awaiting={awaiting("acceptance")} />);
    await userEvent.click(
      screen.getByRole("button", { name: "Принять и слить" }),
    );

    const alert = await waitFor(() => screen.getByRole("alert"));
    expect(alert).toHaveTextContent("fetch failed");
  });
});
