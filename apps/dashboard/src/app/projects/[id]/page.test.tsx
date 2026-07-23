import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ProjectPage from "./page";
import * as api from "@/lib/api";

vi.mock("@/lib/api", () => ({
  listProjectTasks: vi.fn(),
}));

describe("ProjectPage", () => {
  it("renders each task as a link to its detail page", async () => {
    vi.mocked(api.listProjectTasks).mockResolvedValue([
      {
        id: "TASK-001",
        projectId: "proj-a",
        state: "backlog",
        updatedAt: "2026-07-22T00:00:00Z",
        title: "Первая задача",
        type: "feature",
        scope: "",
        acceptanceCriteria: [],
      },
    ]);

    render(await ProjectPage({ params: Promise.resolve({ id: "proj-a" }) }));

    expect(api.listProjectTasks).toHaveBeenCalledWith("proj-a");
    const link = screen.getByRole("link", { name: /TASK-001/ });
    expect(link).toHaveAttribute("href", "/projects/proj-a/tasks/TASK-001");
    expect(screen.getByText(/Первая задача/)).toBeInTheDocument();
  });

  it("shows a message when the project has no tasks", async () => {
    vi.mocked(api.listProjectTasks).mockResolvedValue([]);

    render(await ProjectPage({ params: Promise.resolve({ id: "proj-a" }) }));

    expect(screen.getByText(/задач пока нет/i)).toBeInTheDocument();
  });
});
