import { describe, expect, it } from "vitest";
import { decryptInline, encryptInline, isInlineCiphertext } from "./inline-crypto.js";

describe("inline crypto", () => {
  it("round-trips with enc: prefix", () => {
    const v = encryptInline("LTAI5tSECRET");
    expect(isInlineCiphertext(v)).toBe(true);
    expect(decryptInline(v)).toBe("LTAI5tSECRET");
  });
  it("rejects plaintext", () => {
    expect(() => decryptInline("LTAI5tSECRET")).toThrow("ship storage encrypt");
  });
});
