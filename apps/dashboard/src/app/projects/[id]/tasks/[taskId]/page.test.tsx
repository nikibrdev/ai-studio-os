import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import TaskPage from "./page";
import * as api from "@/lib/api";

vi.mock("@/lib/api", () => ({
  getTask: vi.fn(),
  getArtifact: vi.fn(),
}));

describe("TaskPage", () => {
  it("renders the task's fields", async () => {
    vi.mocked(api.getTask).mockResolvedValue({
      id: "TASK-001",
      projectId: "proj-a",
      state: "in-progress",
      updatedAt: "2026-07-22T00:00:00Z",
      title: "Заголовок задачи",
      type: "feature",
      scope: "Описание области работ",
      acceptanceCriteria: ["критерий раз", "критерий два"],
      awaitingDecision: "",
      repository: "github.com/org/repo",
      pullRequestId: "42",
      qaReportId: "",
    });

    render(
      await TaskPage({
        params: Promise.resolve({ id: "proj-a", taskId: "TASK-001" }),
      }),
    );

    expect(api.getTask).toHaveBeenCalledWith("proj-a", "TASK-001");
    expect(screen.getByText(/Заголовок задачи/)).toBeInTheDocument();
    expect(screen.getByText("feature")).toBeInTheDocument();
    expect(screen.getByText("in-progress")).toBeInTheDocument();
    expect(screen.getByText("Описание области работ")).toBeInTheDocument();
    expect(screen.getByText("критерий раз")).toBeInTheDocument();
    expect(screen.getByText("критерий два")).toBeInTheDocument();
  });

  it("shows a dash when scope and acceptance criteria are empty", async () => {
    vi.mocked(api.getTask).mockResolvedValue({
      id: "TASK-002",
      projectId: "proj-a",
      state: "backlog",
      updatedAt: "2026-07-22T00:00:00Z",
      title: "Без scope",
      type: "bugfix",
      scope: "",
      acceptanceCriteria: [],
      awaitingDecision: "",
      repository: "",
      pullRequestId: "",
      qaReportId: "",
    });

    render(
      await TaskPage({
        params: Promise.resolve({ id: "proj-a", taskId: "TASK-002" }),
      }),
    );

    expect(screen.getAllByText("—")).toHaveLength(2);
  });
});

// The QA report is what a human reads before an irreversible acceptance
// decision, so its absence must be stated, never left as blank space (TASK-100).
describe("TaskPage QA report", () => {
  // Resets call counts as well as implementations: without it a previous test's
  // getArtifact call leaks in and "was never called" cannot be asserted.
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const base = {
    id: "TASK-001",
    projectId: "proj-a",
    state: "testing",
    updatedAt: "2026-07-27T00:00:00Z",
    title: "Задача",
    type: "feature",
    scope: "",
    acceptanceCriteria: [],
    awaitingDecision: "acceptance" as const,
    repository: "github.com/org/repo",
    pullRequestId: "42",
  };

  it("shows the report when one exists", async () => {
    vi.mocked(api.getTask).mockResolvedValue({
      ...base,
      qaReportId: "report-1",
    });
    vi.mocked(api.getArtifact).mockResolvedValue({
      id: "report-1",
      projectId: "proj-a",
      type: "TestReport",
      origin: "produced",
      author: "claude-code",
      state: "published",
      payload: "Критерии сходятся, замечаний нет.",
      createdAt: "2026-07-27T00:00:00Z",
    });

    render(
      await TaskPage({
        params: Promise.resolve({ id: "proj-a", taskId: "TASK-001" }),
      }),
    );

    expect(api.getArtifact).toHaveBeenCalledWith("report-1");
    expect(screen.getByText(/Критерии сходятся/)).toBeInTheDocument();
    expect(screen.getByText(/claude-code/)).toBeInTheDocument();
  });

  it("says a check has not happened rather than showing nothing", async () => {
    vi.mocked(api.getTask).mockResolvedValue({ ...base, qaReportId: "" });

    render(
      await TaskPage({
        params: Promise.resolve({ id: "proj-a", taskId: "TASK-001" }),
      }),
    );

    expect(screen.getByText(/Проверка ещё не выполнялась/)).toBeInTheDocument();
    expect(api.getArtifact).not.toHaveBeenCalled();
  });
});
