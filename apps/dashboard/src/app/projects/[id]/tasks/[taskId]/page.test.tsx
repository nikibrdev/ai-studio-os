import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import TaskPage from "./page";
import * as api from "@/lib/api";

vi.mock("@/lib/api", () => ({
  getTask: vi.fn(),
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
    });

    render(
      await TaskPage({
        params: Promise.resolve({ id: "proj-a", taskId: "TASK-002" }),
      }),
    );

    expect(screen.getAllByText("—")).toHaveLength(2);
  });
});
