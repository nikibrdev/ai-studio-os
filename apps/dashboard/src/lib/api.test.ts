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

    // TEMPORARY: intentionally wrong expectation to prove CI's dashboard
    // test step actually fails red, not silently passes (TASK-077
    // acceptance criterion). Reverted before merge — see the task's Отчёт.
    await expect(listProjects()).resolves.toEqual([]);
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
