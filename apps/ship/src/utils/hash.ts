import { createHash } from "node:crypto";
import { createReadStream } from "node:fs";

export interface BundleDigest {
  sha1: string;
  sha256: string;
  size: number;
}

/** Single streaming pass over a file computing sha1 (object key), sha256 (snapshot hash) and size. */
export function digestFile(path: string): Promise<BundleDigest> {
  return new Promise((resolve, reject) => {
    const h1 = createHash("sha1");
    const h256 = createHash("sha256");
    let size = 0;
    createReadStream(path)
      .on("data", (chunk: Buffer | string) => {
        const b = typeof chunk === "string" ? Buffer.from(chunk) : chunk;
        h1.update(b);
        h256.update(b);
        size += b.length;
      })
      .on("error", reject)
      .on("end", () => resolve({ sha1: h1.digest("hex"), sha256: h256.digest("hex"), size }));
  });
}
