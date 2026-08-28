import { existsSync, readFileSync } from "node:fs";
import { parse } from "yaml";
import { type Config, configSchema } from "./schema.js";

export function loadConfig(path: string): Config {
  if (!existsSync(path)) {
    throw new Error(`${path} not exist`);
  }
  const content = readFileSync(path, "utf8");
  const result = configSchema.safeParse(parse(content));
  if (!result.success) {
    throw new Error(`unmarshal yml error: ${result.error.message}`);
  }
  return result.data;
}
