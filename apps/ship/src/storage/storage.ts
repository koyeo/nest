export interface ObjectInfo {
  key: string;
  size: number;
}

export interface ObjectStorage {
  /** Upload the local file at filePath to key. */
  upload(key: string, filePath: string): Promise<void>;
  /** Object size in bytes, or -1 when the object does not exist. */
  head(key: string): Promise<number>;
  listObjects(prefix: string): Promise<ObjectInfo[]>;
  deleteObjects(keys: string[]): Promise<void>;
  presignedUrl(key: string, expiresSeconds: number): Promise<string>;
}

export const DELETE_BATCH = 1000;

export function chunk<T>(items: T[], size: number): T[][] {
  const out: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    out.push(items.slice(i, i + size));
  }
  return out;
}
