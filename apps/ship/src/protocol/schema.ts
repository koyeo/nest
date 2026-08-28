import { z } from "zod";
import { decryptInline, isInlineCiphertext } from "../config/inline-crypto.js";
import { type StorageCredential, storageProviderSchema } from "../storage/provider.js";

export const serverSchema = z.object({
  alias: z.string().default(""),
  comment: z.string().default(""),
  use: z.string().default(""),
  host: z.string().default(""),
  port: z.number().int().default(22),
  user: z.string().default(""),
  password: z.string().default(""),
  identity_file: z.string().default(""),
});
export type Server = z.infer<typeof serverSchema>;

export const fileMappingSchema = z.object({
  source: z.string(),
  target: z.string(),
  storage: z.string().default(""),
});
export type FileMapping = z.infer<typeof fileMappingSchema>;

export const uploadSchema = z.object({
  storage: z.string(),
  source: z.string(),
});
export type Upload = z.infer<typeof uploadSchema>;

export const remoteCommandSchema = z.object({
  comment: z.string().default(""),
  run: z.string(),
});
export type RemoteCommand = z.infer<typeof remoteCommandSchema>;

export const deploySchema = z.object({
  servers: z.array(serverSchema).default([]),
  files: z.array(fileMappingSchema).default([]),
  commands: z.array(remoteCommandSchema).default([]),
  cwd: z.string().default(""),
  shell_init: z.string().default(""),
});
export type Deploy = z.infer<typeof deploySchema>;

const commentField = z.string().default("");

export const commandSchema = z.union([
  z.object({ comment: commentField, run: z.string() }).strict(),
  z.object({ comment: commentField, use: z.string() }).strict(),
  z.object({ comment: commentField, upload: uploadSchema }).strict(),
  z.object({ comment: commentField, deploy: deploySchema }).strict(),
]);
export type Command = z.infer<typeof commandSchema>;

export const taskSchema = z.object({
  comment: z.string().default(""),
  workspace: z.string().default(""),
  branches: z.array(z.string()).default([]),
  envs: z.record(z.string(), z.string()).default({}),
  commands: z.array(commandSchema).default([]),
});
export type Task = z.infer<typeof taskSchema>;

const encValue = z.string().refine(isInlineCiphertext, { message: "must be an 'enc:' value from 'ship storage encrypt'" });

/** Credentials embedded in ship.yaml; access keys are `enc:` ciphertext under the tool's built-in key. */
export const inlineStorageSchema = z
  .object({
    provider: storageProviderSchema,
    endpoint: z.string().default(""),
    region: z.string().default(""),
    bucket: z.string(),
    access_key_id: encValue,
    access_key_secret: encValue,
  })
  .strict();
export type InlineStorage = z.infer<typeof inlineStorageSchema>;

export const storageRefSchema = z.union([z.string(), inlineStorageSchema]);
export type StorageRef = z.infer<typeof storageRefSchema>;

export const configSchema = z.object({
  version: z.union([z.string(), z.number()]).transform((v) => String(v)),
  servers: z.record(z.string(), serverSchema).default({}),
  storages: z.record(z.string(), storageRefSchema).default({}),
  envs: z.record(z.string(), z.string()).default({}),
  tasks: z.record(z.string(), taskSchema).default({}),
});
export type Config = z.infer<typeof configSchema>;

/** Display name of a server: "comment:host" or "host". */
export function serverName(server: Server): string {
  return server.comment !== "" ? `${server.comment}:${server.host}` : server.host;
}

export type ResolvedStorage =
  | { kind: "global"; name: string }
  | { kind: "inline"; credential: StorageCredential };

/** Resolve a ship.yaml storage alias: either a global config name or inline (decrypted) credentials. */
export function resolveStorage(config: Config, alias: string): ResolvedStorage {
  const entries = Object.keys(config.storages);
  if (entries.length === 0) {
    throw new Error("no storage declared in ship.yaml, add a 'storages' section first");
  }
  const ref = config.storages[alias];
  if (ref === undefined) {
    throw new Error(`storage '${alias}' not declared in ship.yaml`);
  }
  if (typeof ref === "string") {
    return { kind: "global", name: ref };
  }
  return {
    kind: "inline",
    credential: {
      provider: ref.provider,
      endpoint: ref.endpoint,
      region: ref.region,
      bucket: ref.bucket,
      access_key_id: decryptInline(ref.access_key_id),
      access_key_secret: decryptInline(ref.access_key_secret),
    },
  };
}
