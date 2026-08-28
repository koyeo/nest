import { describe, expect, it } from "vitest";
import { parse } from "yaml";
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
    expect(resolveStorage(cfg, "oss")).toBe("my-oss");
  });

  it("rejects a command mixing run and use", () => {
    const bad = parse("version: '1'\ntasks:\n  t:\n    commands:\n      - run: a\n        use: b\n");
    expect(configSchema.safeParse(bad).success).toBe(false);
  });

  it("rejects undeclared storage alias", () => {
    const cfg = configSchema.parse(parse("version: '1'\nstorages:\n  a: b\n"));
    expect(() => resolveStorage(cfg, "x")).toThrow("not declared");
  });
});
