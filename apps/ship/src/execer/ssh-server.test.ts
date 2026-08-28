import { describe, expect, it } from "vitest";
import { wrapLoginShell } from "./ssh-server.js";

describe("wrapLoginShell", () => {
  it("escapes single quotes and preserves newlines", () => {
    expect(wrapLoginShell("echo 'hi'\nls")).toBe(`bash -l -c 'echo '"'"'hi'"'"'\nls'`);
  });
});
