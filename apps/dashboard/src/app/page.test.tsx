import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import HomePage from "./page";

describe("HomePage", () => {
  it("renders without crashing", () => {
    render(<HomePage />);
    expect(screen.getByText(/TASK-075/)).toBeInTheDocument();
  });
});
