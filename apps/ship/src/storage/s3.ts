import { createReadStream, statSync } from "node:fs";
import {
  DeleteObjectsCommand,
  GetObjectCommand,
  HeadObjectCommand,
  ListObjectsV2Command,
  type ListObjectsV2CommandOutput,
  NotFound,
  PutObjectCommand,
  S3Client,
} from "@aws-sdk/client-s3";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";
import { DELETE_BATCH, type ObjectInfo, type ObjectStorage, chunk } from "./storage.js";

export class S3Storage implements ObjectStorage {
  private readonly client: S3Client;
  private readonly bucket: string;

  constructor(region: string, accessKeyId: string, accessKeySecret: string, bucket: string, endpoint: string) {
    this.client = new S3Client({
      region,
      credentials: { accessKeyId, secretAccessKey: accessKeySecret },
      ...(endpoint !== "" ? { endpoint } : {}),
    });
    this.bucket = bucket;
  }

  async upload(key: string, filePath: string): Promise<void> {
    const size = statSync(filePath).size;
    await this.client.send(
      new PutObjectCommand({ Bucket: this.bucket, Key: key, Body: createReadStream(filePath), ContentLength: size }),
    );
  }

  async head(key: string): Promise<number> {
    try {
      const out = await this.client.send(new HeadObjectCommand({ Bucket: this.bucket, Key: key }));
      return out.ContentLength ?? -1;
    } catch (e) {
      if (e instanceof NotFound) {
        return -1;
      }
      throw e;
    }
  }

  async listObjects(prefix: string): Promise<ObjectInfo[]> {
    const objects: ObjectInfo[] = [];
    let token: string | undefined = undefined;
    do {
      const page: ListObjectsV2CommandOutput = await this.client.send(
        new ListObjectsV2Command({ Bucket: this.bucket, Prefix: prefix, ...(token !== undefined ? { ContinuationToken: token } : {}) }),
      );
      for (const obj of page.Contents ?? []) {
        if (obj.Key !== undefined) {
          objects.push({ key: obj.Key, size: obj.Size ?? 0 });
        }
      }
      token = page.NextContinuationToken;
    } while (token !== undefined);
    return objects;
  }

  async deleteObjects(keys: string[]): Promise<void> {
    for (const batch of chunk(keys, DELETE_BATCH)) {
      await this.client.send(
        new DeleteObjectsCommand({ Bucket: this.bucket, Delete: { Objects: batch.map((k) => ({ Key: k })) } }),
      );
    }
  }

  async presignedUrl(key: string, expiresSeconds: number): Promise<string> {
    return getSignedUrl(this.client, new GetObjectCommand({ Bucket: this.bucket, Key: key }), { expiresIn: expiresSeconds });
  }
}
