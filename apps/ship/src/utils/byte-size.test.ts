import { describe, expect, it } from "vitest";
import { byteSize } from "./byte-size.js";

describe("byteSize", () => {
  it("formats bytes", () => {
    expect(byteSize(0)).toBe("0B");
    expect(byteSize(512)).toBe("512B");
    expect(byteSize(1024)).toBe("1K");
    expect(byteSize(1536)).toBe("1.5K");
    expect(byteSize(10 * 1024 * 1024)).toBe("10M");
    expect(byteSize(3 * 2 ** 30)).toBe("3G");
  });
});
