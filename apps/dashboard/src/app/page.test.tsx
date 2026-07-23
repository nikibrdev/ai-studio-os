import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import HomePage from "./page";
import * as api from "@/lib/api";

vi.mock("@/lib/api", () => ({
  listProjects: vi.fn(),
}));

describe("HomePage", () => {
  it("renders each project as a link to its page", async () => {
    vi.mocked(api.listProjects).mockResolvedValue([
      {
        id: "proj-a",
        name: "Alpha",
        state: "active",
        createdAt: "2026-07-22T00:00:00Z",
      },
      {
        id: "proj-b",
        name: "Beta",
        state: "created",
        createdAt: "2026-07-22T00:00:00Z",
      },
    ]);

    render(await HomePage());

    expect(screen.getByRole("link", { name: "Alpha" })).toHaveAttribute(
      "href",
      "/projects/proj-a",
    );
    expect(screen.getByRole("link", { name: "Beta" })).toHaveAttribute(
      "href",
      "/projects/proj-b",
    );
  });

  it("shows a message when there are no projects", async () => {
    vi.mocked(api.listProjects).mockResolvedValue([]);

    render(await HomePage());

    expect(screen.getByText(/проектов пока нет/i)).toBeInTheDocument();
  });
});
