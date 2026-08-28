import { describe, expect, it } from "vitest";
import { checkTargetPath, targetDirOf } from "./server-runner.js";

describe("target path rules", () => {
  it("rejects ~ and shallow absolute paths", () => {
    expect(() => checkTargetPath("~/app")).toThrow("invalid target path");
    expect(() => checkTargetPath("/data")).toThrow("too short");
    expect(() => checkTargetPath("/data/app")).not.toThrow();
  });
  it("derives targetDir", () => {
    expect(targetDirOf("/data/app/")).toBe("/data/app");
    expect(targetDirOf("/data/app")).toBe("/data");
  });
});
