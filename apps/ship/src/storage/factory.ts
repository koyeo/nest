import type { StorageCredential } from "./provider.js";
import { OssStorage } from "./oss.js";
import { S3Storage } from "./s3.js";
import type { ObjectStorage } from "./storage.js";

/** Build a client from a decrypted credential. */
export function newStorage(cred: StorageCredential): ObjectStorage {
  switch (cred.provider) {
    case "oss":
      if (cred.endpoint === "") {
        throw new Error("OSS requires 'endpoint'");
      }
      return new OssStorage(cred.endpoint, cred.access_key_id, cred.access_key_secret, cred.bucket);
    case "s3":
      if (cred.region === "") {
        throw new Error("S3 requires 'region'");
      }
      return new S3Storage(cred.region, cred.access_key_id, cred.access_key_secret, cred.bucket, cred.endpoint);
  }
}
