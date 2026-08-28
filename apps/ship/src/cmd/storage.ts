import { Command } from "commander";
import { OBJECT_PREFIX } from "../common/const.js";
import { encryptInline } from "../config/inline-crypto.js";
import {
  type StorageInput,
  type StorageProvider,
  addStorage,
  decryptStorage,
  loadUserConfig,
  removeStorage,
  saveUserConfig,
  storageProviderSchema,
} from "../config/user-config.js";
import { loadConfig } from "../protocol/load.js";
import { type ResolvedStorage, resolveStorage } from "../protocol/schema.js";
import { newStorage } from "../storage/factory.js";
import type { ObjectStorage } from "../storage/storage.js";
import { byteSize } from "../utils/byte-size.js";
import { ask } from "../utils/prompt.js";

interface AddFlags {
  provider: string;
  endpoint: string;
  region: string;
  bucket: string;
  accessKeyId: string;
  accessKeySecret: string;
}

const out = (s: string): void => {
  process.stdout.write(s + "\n");
};

export function storageCommand(getConfigFile: () => string): Command {
  const cmd = new Command("storage").description("Manage cloud storage configurations (OSS / S3)");

  cmd
    .command("add")
    .description("Add a cloud storage configuration (interactive or flag-based)")
    .argument("[name]", "config name", "")
    .option("--provider <provider>", "Storage provider: oss or s3", "")
    .option("--endpoint <endpoint>", "Service endpoint (required for OSS)", "")
    .option("--region <region>", "Region (required for S3)", "")
    .option("--bucket <bucket>", "Bucket name", "")
    .option("--access-key-id <id>", "Access Key ID", "")
    .option("--access-key-secret <secret>", "Access Key Secret", "")
    .action(async (nameArg: string, flags: AddFlags) => {
      await runAdd(nameArg, flags);
    });

  cmd
    .command("encrypt")
    .description("Encrypt credentials with the tool's built-in key and print a ship.yaml 'storages:' block to paste (anyone with the file can decrypt it)")
    .argument("[alias]", "alias to use in ship.yaml", "oss")
    .option("--provider <provider>", "Storage provider: oss or s3", "")
    .option("--endpoint <endpoint>", "Service endpoint (required for OSS)", "")
    .option("--region <region>", "Region (required for S3)", "")
    .option("--bucket <bucket>", "Bucket name", "")
    .option("--access-key-id <id>", "Access Key ID", "")
    .option("--access-key-secret <secret>", "Access Key Secret", "")
    .action(async (alias: string, flags: AddFlags) => {
      await runEncrypt(alias, flags);
    });

  cmd
    .command("list")
    .description("List all configured storages with provider, bucket, and endpoint")
    .action(() => {
      runList();
    });

  cmd
    .command("remove")
    .description("Remove a storage configuration by name")
    .argument("<name>")
    .action((name: string) => {
      const cfg = removeStorage(loadUserConfig(), name);
      saveUserConfig(cfg);
      out(`✅ Storage '${name}' removed`);
    });

  cmd
    .command("usage")
    .description("Show total size and object count of ship artifacts in the bucket")
    .argument("[name]", "storage alias or global config name", "")
    .action(async (name: string) => {
      await runUsage(getConfigFile(), name);
    });

  cmd
    .command("clean")
    .description("Delete all ship-managed objects from the bucket (with confirmation)")
    .argument("<name>")
    .action(async (name: string) => {
      await runClean(getConfigFile(), name);
    });

  return cmd;
}

async function runAdd(nameArg: string, flags: AddFlags): Promise<void> {
  let name = nameArg;
  if (name === "" && flags.provider !== "" && flags.bucket !== "") {
    name = flags.bucket;
  } else if (name === "") {
    out("☁️  Add Cloud Storage");
    out("──────────────────────────");
    out("");
    name = await ask("Config name (e.g. oss-prod, s3-us)", "");
    if (name === "") {
      throw new Error("config name is required");
    }
  }
  const input = await collectStorageInput(name, flags);
  const cfg = addStorage(loadUserConfig(), input);
  saveUserConfig(cfg);
  out("");
  out(`✅ Storage '${name}' saved (credentials encrypted in ~/.ship/config.json)`);
}

async function runEncrypt(alias: string, flags: AddFlags): Promise<void> {
  const input = await collectStorageInput(alias, flags);
  const lines = [
    "storages:",
    `  ${alias}:`,
    `    provider: ${input.provider}`,
    ...(input.endpoint !== "" ? [`    endpoint: ${input.endpoint}`] : []),
    ...(input.region !== "" ? [`    region: ${input.region}`] : []),
    `    bucket: ${input.bucket}`,
    `    access_key_id: ${encryptInline(input.accessKeyId)}`,
    `    access_key_secret: ${encryptInline(input.accessKeySecret)}`,
  ];
  out("");
  out("# Paste into ship.yaml:");
  out(lines.join("\n"));
  out("");
  out("⚠️  The key is built into the ship binary: anyone holding this ship.yaml can access the bucket.");
  out(`    Treat the bucket as public, never keep secrets in it, and clean it regularly: ship storage clean ${alias}`);
}

/** Fill missing fields interactively; validates provider and required fields. */
async function collectStorageInput(name: string, flags: AddFlags): Promise<StorageInput> {
  let providerInput = flags.provider;
  if (providerInput === "") {
    out("");
    out("Select provider:");
    out("  1) oss  — Alibaba Cloud OSS");
    out("  2) s3   — AWS S3 (or S3-compatible)");
    out("");
    const choice = await ask("Provider (1/2 or oss/s3)", "");
    if (choice === "1") {
      providerInput = "oss";
    } else if (choice === "2") {
      providerInput = "s3";
    } else {
      providerInput = choice;
    }
  }
  const parsed = storageProviderSchema.safeParse(providerInput);
  if (!parsed.success) {
    throw new Error(`provider must be 'oss' or 's3', got '${providerInput}'`);
  }
  const provider: StorageProvider = parsed.data;

  out("");
  let endpoint = flags.endpoint;
  let region = flags.region;
  if (provider === "oss") {
    if (endpoint === "") {
      endpoint = await ask("OSS Endpoint (e.g. oss-cn-hangzhou.aliyuncs.com)", "");
      if (endpoint === "") {
        throw new Error("endpoint is required for OSS");
      }
    }
  } else {
    if (region === "") {
      region = await ask("AWS Region (e.g. us-east-1)", "");
      if (region === "") {
        throw new Error("region is required for S3");
      }
    }
    if (endpoint === "") {
      endpoint = await ask("Custom endpoint (leave empty for AWS)", "");
    }
  }

  let bucket = flags.bucket;
  if (bucket === "") {
    bucket = await ask("Bucket name", "");
    if (bucket === "") {
      throw new Error("bucket name is required");
    }
  }

  out("");
  let accessKeyId = flags.accessKeyId;
  if (accessKeyId === "") {
    accessKeyId = await ask("Access Key ID", "");
    if (accessKeyId === "") {
      throw new Error("Access Key ID is required");
    }
  }
  let accessKeySecret = flags.accessKeySecret;
  if (accessKeySecret === "") {
    accessKeySecret = await ask("Access Key Secret", "");
    if (accessKeySecret === "") {
      throw new Error("Access Key Secret is required");
    }
  }

  return { name, provider, endpoint, region, bucket, accessKeyId, accessKeySecret };
}

function runList(): void {
  const cfg = loadUserConfig();
  const entries = Object.entries(cfg.storages);
  if (entries.length === 0) {
    out("No storage configs found.");
    out("");
    out("Add one with:");
    out("  ship storage add");
    return;
  }
  out("Configured storages:");
  out("");
  for (const [name, s] of entries) {
    out(`  📦 ${name}`);
    out(`     provider : ${s.provider}`);
    out(`     bucket   : ${s.bucket}`);
    if (s.endpoint !== "") {
      out(`     endpoint : ${s.endpoint}`);
    }
    if (s.region !== "") {
      out(`     region   : ${s.region}`);
    }
    out(`     keys     : ${maskSecret(s.access_key_id)} (encrypted)`);
    out("");
  }
}

/** Accepts a ship.yaml alias (global or inline) or a global config name. */
function newStorageClient(configFile: string, name: string): ObjectStorage {
  let resolved: ResolvedStorage = { kind: "global", name };
  try {
    resolved = resolveStorage(loadConfig(configFile), name);
  } catch {
    // not a ship.yaml alias (or no ship.yaml): treat name as a global config name
  }
  if (resolved.kind === "inline") {
    return newStorage(resolved.credential);
  }
  return newStorage(decryptStorage(loadUserConfig(), resolved.name));
}

async function runUsage(configFile: string, name: string): Promise<void> {
  if (name !== "") {
    await printStorageUsage(configFile, name);
    return;
  }
  const cfg = loadUserConfig();
  const names = Object.keys(cfg.storages);
  if (names.length === 0) {
    out("No storage configs found.");
    out("");
    out("Add one with:");
    out("  ship storage add");
    return;
  }
  for (const n of names) {
    try {
      await printStorageUsage(configFile, n);
    } catch (e) {
      out(`⚠️  Storage '${n}': ${e instanceof Error ? e.message : String(e)}\n`);
    }
  }
}

async function printStorageUsage(configFile: string, name: string): Promise<void> {
  const store = newStorageClient(configFile, name);
  const objects = await store.listObjects(OBJECT_PREFIX);
  if (objects.length === 0) {
    out(`📦 Storage '${name}': empty (no ${OBJECT_PREFIX} objects)\n`);
    return;
  }
  const total = objects.reduce((sum, o) => sum + o.size, 0);
  out(`📦 Storage '${name}': ${objects.length} objects, ${byteSize(total)}`);
}

async function runClean(configFile: string, name: string): Promise<void> {
  const store = newStorageClient(configFile, name);
  const objects = await store.listObjects(OBJECT_PREFIX);
  if (objects.length === 0) {
    out(`📦 Storage '${name}': already empty`);
    return;
  }
  const total = objects.reduce((sum, o) => sum + o.size, 0);
  out(`⚠️  About to delete ${objects.length} objects (${byteSize(total)}) from storage '${name}'`);
  out("");
  for (const o of objects) {
    out(`   🗑  ${byteSize(o.size)}  ${o.key}`);
  }
  out("");
  const confirm = await ask("Type 'yes' to confirm", "");
  if (confirm !== "yes") {
    out("Cancelled.");
    return;
  }
  await store.deleteObjects(objects.map((o) => o.key));
  out(`✅ Deleted ${objects.length} objects (${byteSize(total)})`);
}

function maskSecret(s: string): string {
  if (s.length <= 8) {
    return "****";
  }
  return `${s.slice(0, 4)}****${s.slice(-4)}`;
}
