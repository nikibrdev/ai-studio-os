import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, listProjects } from "./api";

describe("listProjects", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns the parsed project list on success", async () => {
    const projects = [
      {
        id: "proj-a",
        name: "A",
        state: "active",
        createdAt: "2026-07-22T00:00:00Z",
      },
    ];
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(projects),
      }),
    );

    await expect(listProjects()).resolves.toEqual(projects);
  });

  it("throws ApiError with the response status on failure", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
      }),
    );

    await expect(listProjects()).rejects.toBeInstanceOf(ApiError);
    await expect(listProjects()).rejects.toMatchObject({ status: 500 });
  });
});
