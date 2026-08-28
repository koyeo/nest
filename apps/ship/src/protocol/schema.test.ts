import { describe, expect, it } from "vitest";
import { parse } from "yaml";
import { encryptInline } from "../config/inline-crypto.js";
import { configSchema, resolveStorage } from "./schema.js";

const sample = `
version: 1.0
servers:
  web:
    host: 10.0.0.1
    user: root
storages:
  oss: my-oss
tasks:
  build:
    comment: build it
    commands:
      - run: echo hi
      - use: other
      - upload:
          storage: oss
          source: ./dist
      - deploy:
          servers:
            - use: web
          files:
            - source: ./dist
              target: /data/app
              storage: oss
          commands:
            - run: ls
          cwd: /data/app
`;

describe("configSchema", () => {
  it("parses a full config with defaults", () => {
    const cfg = configSchema.parse(parse(sample));
    expect(cfg.version).toBe("1");
    expect(cfg.servers["web"]?.port).toBe(22);
    expect(cfg.tasks["build"]?.commands).toHaveLength(4);
    const deploy = cfg.tasks["build"]?.commands[3];
    expect(deploy !== undefined && "deploy" in deploy && deploy.deploy.shell_init).toBe("");
    expect(resolveStorage(cfg, "oss")).toEqual({ kind: "global", name: "my-oss" });
  });

  it("rejects a command mixing run and use", () => {
    const bad = parse("version: '1'\ntasks:\n  t:\n    commands:\n      - run: a\n        use: b\n");
    expect(configSchema.safeParse(bad).success).toBe(false);
  });

  it("rejects undeclared storage alias", () => {
    const cfg = configSchema.parse(parse("version: '1'\nstorages:\n  a: b\n"));
    expect(() => resolveStorage(cfg, "x")).toThrow("not declared");
  });

  it("parses inline encrypted storage", () => {
    const yamlText = `version: '1'\nstorages:\n  oss:\n    provider: oss\n    endpoint: e\n    bucket: b\n    access_key_id: ${encryptInline("ID")}\n    access_key_secret: ${encryptInline("SECRET")}\n`;
    const cfg = configSchema.parse(parse(yamlText));
    const r = resolveStorage(cfg, "oss");
    expect(r.kind).toBe("inline");
    expect(r.kind === "inline" && r.credential.access_key_secret).toBe("SECRET");
  });

  it("rejects inline storage with plaintext keys", () => {
    const bad = parse("version: '1'\nstorages:\n  oss:\n    provider: oss\n    bucket: b\n    access_key_id: ID\n    access_key_secret: S\n");
    expect(configSchema.safeParse(bad).success).toBe(false);
  });
});
