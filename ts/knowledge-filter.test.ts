import { describe, it, expect } from "vitest";
import { filterKnowledgeByStatus, buildKnowledgeFilterParams } from "./knowledge-filter";

describe("knowledge-filter", () => {
  it("filters by status", () => {
    const items = [{ status: "rumor" }, { status: "revealed" }, { status: "rumor" }];
    expect(filterKnowledgeByStatus(items, "rumor")).toHaveLength(2);
    expect(filterKnowledgeByStatus(items, "")).toHaveLength(3);
  });
  it("builds params", () => {
    expect(buildKnowledgeFilterParams("false")).toBe("?status=false");
    expect(buildKnowledgeFilterParams("")).toBe("");
  });
});
