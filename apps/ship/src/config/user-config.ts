import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { homedir } from "node:os";
import { dirname, join } from "node:path";
import { z } from "zod";
import { USER_CONFIG_DIR, USER_CONFIG_FILE } from "../common/const.js";
import { decrypt, encrypt, generateEncryptKey } from "./crypto.js";

import { type StorageCredential, type StorageProvider, storageProviderSchema } from "../storage/provider.js";

export { storageProviderSchema, type StorageProvider, type StorageCredential };

const storageCredentialSchema = z.object({
  provider: storageProviderSchema,
  endpoint: z.string().default(""),
  region: z.string().default(""),
  bucket: z.string(),
  access_key_id: z.string(),
  access_key_secret: z.string(),
});

const userConfigSchema = z.object({
  lang: z.enum(["zh", "en"]).default("zh"),
  encrypt_key: z.string().default(""),
  storages: z.record(z.string(), storageCredentialSchema).default({}),
});
export type UserConfig = z.infer<typeof userConfigSchema>;

export function userConfigPath(): string {
  return join(homedir(), USER_CONFIG_DIR, USER_CONFIG_FILE);
}

function defaultUserConfig(): UserConfig {
  return { lang: "zh", encrypt_key: "", storages: {} };
}

/** Missing file → defaults. Existing but unparsable file → error (decided in proposal 未决 2). */
export function loadUserConfig(): UserConfig {
  const p = userConfigPath();
  if (!existsSync(p)) {
    return defaultUserConfig();
  }
  const result = userConfigSchema.safeParse(JSON.parse(readFileSync(p, "utf8")));
  if (!result.success) {
    throw new Error(`invalid user config ${p}: ${result.error.message}`);
  }
  return result.data;
}

export function saveUserConfig(cfg: UserConfig): void {
  const p = userConfigPath();
  mkdirSync(dirname(p), { recursive: true });
  writeFileSync(p, JSON.stringify(cfg, null, 2));
}

export interface StorageInput {
  name: string;
  provider: StorageProvider;
  endpoint: string;
  region: string;
  bucket: string;
  accessKeyId: string;
  accessKeySecret: string;
}

/** Encrypts credentials and upserts a storage config. Generates the encrypt key on first use. */
export function addStorage(cfg: UserConfig, input: StorageInput): UserConfig {
  const encryptKey = cfg.encrypt_key !== "" ? cfg.encrypt_key : generateEncryptKey();
  return {
    ...cfg,
    encrypt_key: encryptKey,
    storages: {
      ...cfg.storages,
      [input.name]: {
        provider: input.provider,
        endpoint: input.endpoint,
        region: input.region,
        bucket: input.bucket,
        access_key_id: encrypt(encryptKey, input.accessKeyId),
        access_key_secret: encrypt(encryptKey, input.accessKeySecret),
      },
    },
  };
}

export function removeStorage(cfg: UserConfig, name: string): UserConfig {
  if (cfg.storages[name] === undefined) {
    throw new Error(`storage '${name}' not found`);
  }
  const storages = { ...cfg.storages };
  delete storages[name];
  return { ...cfg, storages };
}

/** Returns the credential with plaintext access keys. */
export function decryptStorage(cfg: UserConfig, name: string): StorageCredential {
  const stored = cfg.storages[name];
  if (stored === undefined) {
    throw new Error(`storage '${name}' not found`);
  }
  if (cfg.encrypt_key === "") {
    throw new Error("encrypt_key not set");
  }
  return {
    ...stored,
    access_key_id: decrypt(cfg.encrypt_key, stored.access_key_id),
    access_key_secret: decrypt(cfg.encrypt_key, stored.access_key_secret),
  };
}
