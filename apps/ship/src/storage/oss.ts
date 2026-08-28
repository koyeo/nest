import { statSync } from "node:fs";
import OSS from "ali-oss";
import { z } from "zod";
import { DELETE_BATCH, type ObjectInfo, type ObjectStorage, chunk } from "./storage.js";

/** Above this size an upload is split into parts (Initiate/Complete round trips cost more below it). */
const MULTIPART_THRESHOLD = 32 * 1024 * 1024;
const BASE_PART_SIZE = 4 * 1024 * 1024;
const MAX_PARTS = 10000;
const UPLOAD_PARALLEL = 5;

const headerSchema = z.object({ "content-length": z.string() });

/** Part size that keeps the part count within MAX_PARTS. */
export function ossPartSize(size: number): number {
  let part = BASE_PART_SIZE;
  while (size / part > MAX_PARTS) {
    part *= 2;
  }
  return part;
}

export class OssStorage implements ObjectStorage {
  private readonly client: OSS;

  constructor(endpoint: string, accessKeyId: string, accessKeySecret: string, bucket: string) {
    this.client = new OSS({ endpoint, accessKeyId, accessKeySecret, bucket });
  }

  async upload(key: string, filePath: string): Promise<void> {
    const size = statSync(filePath).size;
    if (size < MULTIPART_THRESHOLD) {
      await this.client.put(key, filePath);
      return;
    }
    await this.client.multipartUpload(key, filePath, {
      partSize: ossPartSize(size),
      parallel: UPLOAD_PARALLEL,
    });
  }

  async head(key: string): Promise<number> {
    try {
      const res = await this.client.head(key);
      const headers = headerSchema.safeParse(res.res.headers);
      return headers.success ? Number.parseInt(headers.data["content-length"], 10) : -1;
    } catch (e) {
      if (typeof e === "object" && e !== null && "status" in e && e.status === 404) {
        return -1;
      }
      throw e;
    }
  }

  async listObjects(prefix: string): Promise<ObjectInfo[]> {
    const objects: ObjectInfo[] = [];
    let marker = "";
    for (;;) {
      const res = await this.client.list({ prefix, marker, "max-keys": 1000 }, {});
      for (const obj of res.objects) {
        objects.push({ key: obj.name, size: obj.size });
      }
      if (!res.isTruncated || res.nextMarker === null) {
        break;
      }
      marker = res.nextMarker;
    }
    return objects;
  }

  async deleteObjects(keys: string[]): Promise<void> {
    for (const batch of chunk(keys, DELETE_BATCH)) {
      await this.client.deleteMulti(batch, { quiet: true });
    }
  }

  async presignedUrl(key: string, expiresSeconds: number): Promise<string> {
    return this.client.signatureUrl(key, { expires: expiresSeconds, method: "GET" });
  }
}

