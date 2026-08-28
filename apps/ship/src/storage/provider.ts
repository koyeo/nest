import { z } from "zod";

export const storageProviderSchema = z.enum(["oss", "s3"]);
export type StorageProvider = z.infer<typeof storageProviderSchema>;

/** Plaintext credential used to build a client. */
export interface StorageCredential {
  provider: StorageProvider;
  endpoint: string;
  region: string;
  bucket: string;
  access_key_id: string;
  access_key_secret: string;
}
