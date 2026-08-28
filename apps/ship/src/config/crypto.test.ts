import { describe, expect, it } from "vitest";
import { decrypt, encrypt, generateEncryptKey } from "./crypto.js";

describe("crypto", () => {
  it("round-trips", () => {
    const key = generateEncryptKey();
    const enc = encrypt(key, "hello world");
    expect(enc).not.toBe("hello world");
    expect(decrypt(key, enc)).toBe("hello world");
  });

  it("nonce(12) + ciphertext + tag(16) layout", () => {
    const key = generateEncryptKey();
    const raw = Buffer.from(encrypt(key, "abc"), "base64");
    expect(raw.length).toBe(12 + 3 + 16);
  });

  it("rejects wrong key", () => {
    const enc = encrypt(generateEncryptKey(), "abc");
    expect(() => decrypt(generateEncryptKey(), enc)).toThrow("decrypt error");
  });

  it("rejects short ciphertext", () => {
    expect(() => decrypt(generateEncryptKey(), Buffer.from("short").toString("base64"))).toThrow("too short");
  });
});
