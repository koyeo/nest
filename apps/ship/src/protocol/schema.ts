import { z } from "zod";

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

export const configSchema = z.object({
  version: z.union([z.string(), z.number()]).transform((v) => String(v)),
  servers: z.record(z.string(), serverSchema).default({}),
  storages: z.record(z.string(), z.string()).default({}),
  envs: z.record(z.string(), z.string()).default({}),
  tasks: z.record(z.string(), taskSchema).default({}),
});
export type Config = z.infer<typeof configSchema>;

/** Display name of a server: "comment:host" or "host". */
export function serverName(server: Server): string {
  return server.comment !== "" ? `${server.comment}:${server.host}` : server.host;
}

/** Resolve a ship.yaml storage alias to the global storage config name. */
export function resolveStorage(config: Config, alias: string): string {
  const entries = Object.keys(config.storages);
  if (entries.length === 0) {
    throw new Error("no storage declared in ship.yaml, add a 'storages' section first");
  }
  const name = config.storages[alias];
  if (name === undefined) {
    throw new Error(`storage '${alias}' not declared in ship.yaml`);
  }
  return name;
}
