import { existsSync } from "node:fs";
import { basename } from "node:path";
import { OBJECT_PREFIX } from "../common/const.js";
import { decryptStorage, loadUserConfig } from "../config/user-config.js";
import { type Config, resolveStorage } from "../protocol/schema.js";
import { newStorage } from "../storage/factory.js";
import type { ObjectStorage } from "../storage/storage.js";
import { byteSize } from "../utils/byte-size.js";
import { digestFile } from "../utils/hash.js";
import { compress } from "../utils/tar.js";
import { cleanShipTmpDir, shipTmpDir } from "./tmp-dir.js";

export interface UploadedObject {
  objectKey: string;
  storageAlias: string;
}

export interface UploadResult {
  objectKey: string;
  /** SHA256 of the bundle; "" when the upload was deduped from an earlier call in this run. */
  bundleHash: string;
}

export interface AliasStorage {
  store: ObjectStorage;
  /** true when credentials are embedded in ship.yaml (anyone holding the file can access the bucket). */
  inline: boolean;
}

/** Resolves a ship.yaml alias → (global config | inline credentials) → client. */
export function storageForAlias(config: Config, alias: string): AliasStorage {
  const resolved = resolveStorage(config, alias);
  if (resolved.kind === "inline") {
    try {
      return { store: newStorage(resolved.credential), inline: true };
    } catch (e) {
      throw new Error(`storage '${alias}' (inline): ${e instanceof Error ? e.message : String(e)}`);
    }
  }
  try {
    return { store: newStorage(decryptStorage(loadUserConfig(), resolved.name)), inline: false };
  } catch (e) {
    throw new Error(`storage '${alias}' (config '${resolved.name}'): ${e instanceof Error ? e.message : String(e)}`);
  }
}

export function inlineStorageWarning(alias: string): string[] {
  return [
    `storage '${alias}' credentials are embedded in ship.yaml — anyone holding this file can read/write the bucket, treat it as PUBLIC.`,
    `Do not keep secrets there and clean it regularly: ship storage clean ${alias}`,
  ];
}

/**
 * Compress → stream sha1 (object key) + sha256 (bundle hash) → Head compare by size → upload.
 * Dedups by local path within one task run and records objects for later cleanup.
 */
export class CloudUploader {
  private readonly dedupKeys = new Map<string, string>();
  readonly uploaded: UploadedObject[] = [];

  constructor(private readonly log: (emoji: string, args: string[]) => void) {}

  async upload(store: ObjectStorage, alias: string, localPath: string): Promise<UploadResult> {
    const sourceName = basename(localPath);
    const known = this.dedupKeys.get(localPath);
    if (known !== undefined) {
      this.log("⏭️", [`[${alias}] ${sourceName} already uploaded`]);
      return { objectKey: known, bundleHash: "" };
    }
    if (!existsSync(localPath)) {
      throw new Error(`source not found: ${localPath}`);
    }

    const bundlePath = `${shipTmpDir()}/${sourceName}.tar.gz`;
    try {
      await compress(localPath, bundlePath);
      const digest = await digestFile(bundlePath);
      const objectKey = `${OBJECT_PREFIX}${digest.sha1}.tar.gz`;
      this.log("☁️", [`[${alias}]`, `${localPath} → ${objectKey} (${byteSize(digest.size)})`]);

      const remoteSize = await store.head(objectKey);
      if (remoteSize === digest.size) {
        this.log("⏭️", ["skipped (already exists)"]);
      } else {
        await store.upload(objectKey, bundlePath);
        this.log("✅", ["uploaded"]);
      }
      this.dedupKeys.set(localPath, objectKey);
      this.uploaded.push({ objectKey, storageAlias: alias });
      return { objectKey, bundleHash: digest.sha256 };
    } finally {
      cleanShipTmpDir();
    }
  }

  /** Delete everything uploaded so far, grouped by alias. Errors are logged, not thrown. */
  async cleanup(config: Config): Promise<void> {
    if (this.uploaded.length === 0) {
      return;
    }
    const groups = new Map<string, string[]>();
    const seen = new Set<string>();
    for (const o of this.uploaded) {
      if (seen.has(o.objectKey)) {
        continue;
      }
      seen.add(o.objectKey);
      groups.set(o.storageAlias, [...(groups.get(o.storageAlias) ?? []), o.objectKey]);
    }
    for (const [alias, keys] of groups) {
      try {
        await storageForAlias(config, alias).store.deleteObjects(keys);
        this.log("🧹", [`cleaned ${keys.length} cloud object(s) from ${alias}`]);
      } catch (e) {
        this.log("⚠️", [`clean: delete objects from '${alias}' error: ${e instanceof Error ? e.message : String(e)}`]);
      }
    }
    this.uploaded.length = 0;
  }
}
