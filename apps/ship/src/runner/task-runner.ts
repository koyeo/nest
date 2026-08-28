import { existsSync } from "node:fs";
import { localExec, mergeEnv } from "../execer/local-runner.js";
import { ServerPool } from "../execer/server-pool.js";
import * as logger from "../logger.js";
import { type Config, type Deploy, type Server, type Task, type Upload, serverName } from "../protocol/schema.js";
import { CloudUploader, storageForAlias } from "./cloud.js";
import { ServerRunner } from "./server-runner.js";

export class TaskRunner {
  private readonly uploader: CloudUploader;

  constructor(
    private readonly conf: Config,
    private readonly task: Task,
    private readonly key: string,
    /** Ancestor task keys, used to detect `use` cycles. */
    private readonly parents: ReadonlySet<string>,
  ) {
    this.uploader = new CloudUploader((emoji, args) => this.log(emoji, args));
  }

  static root(conf: Config, task: Task, key: string): TaskRunner {
    return new TaskRunner(conf, task, key, new Set());
  }

  private log(emoji: string, args: string[]): void {
    logger.step(this.key, this.task.comment, emoji, args);
  }

  async exec(): Promise<void> {
    for (const cmd of this.task.commands) {
      if ("use" in cmd) {
        await this.use(cmd.use);
      } else if ("upload" in cmd) {
        await this.upload(cmd.upload);
      } else if ("deploy" in cmd) {
        await this.deploy(cmd.deploy);
      } else {
        await this.execute(cmd.run);
      }
    }
  }

  printStart(): void {
    this.log("🕘", ["start"]);
  }
  printSuccess(): void {
    this.log("🎉", ["success"]);
  }
  printFailed(): void {
    this.log("❌️", ["failed"]);
  }

  private async use(key: string): Promise<void> {
    if (this.parents.has(key) || key === this.key) {
      throw new Error(`task: ${this.key} depend task: ${key} circlely`);
    }
    const task = this.conf.tasks[key];
    if (task === undefined) {
      throw new Error(`use task: '${key}' not found`);
    }
    const child = new TaskRunner(this.conf, task, key, new Set([...this.parents, this.key]));
    this.log("👉", [task.comment !== "" ? task.comment : key]);
    await child.exec();
    this.log("👈", []);
  }

  private async execute(run: string): Promise<void> {
    this.log("🏃", [run]);
    try {
      await localExec({ command: run, cwd: this.task.workspace, env: mergeEnv(this.conf.envs, this.task.envs) });
    } catch (e) {
      throw new Error(`runner pipe exec error: ${e instanceof Error ? e.message : String(e)}`);
    }
  }

  private async upload(u: Upload): Promise<void> {
    const store = storageForAlias(this.conf, u.storage);
    if (!existsSync(u.source)) {
      throw new Error(`upload source not found: ${u.source}`);
    }
    await this.uploader.upload(store, u.storage, u.source);
  }

  private resolveServers(deploy: Deploy): Map<string, Server> {
    const servers = new Map<string, Server>();
    for (const s of deploy.servers) {
      if (s.use !== "") {
        const found = this.conf.servers[s.use];
        if (found === undefined) {
          throw new Error(`deploy use server: '${s.use}' not exists`);
        }
        servers.set(found.host, found);
      } else {
        servers.set(s.host, s);
      }
    }
    for (const [, s] of servers) {
      if (s.host === "") {
        throw new Error("deploy server host is empty");
      }
    }
    return servers;
  }

  private async deploy(deploy: Deploy): Promise<void> {
    const servers = this.resolveServers(deploy);
    const pool = new ServerPool();
    try {
      const runners = new Map<string, ServerRunner>();
      for (const [key, server] of servers) {
        runners.set(key, new ServerRunner(server, key, pool, (emoji, args) => this.log(emoji, args)));
      }

      for (const [key] of servers) {
        const runner = runners.get(key);
        if (runner === undefined) {
          continue;
        }
        for (const file of deploy.files) {
          if (file.storage !== "") {
            const store = storageForAlias(this.conf, file.storage);
            await runner.deployViaStorage(this.uploader, store, file.storage, file.source, file.target);
          } else {
            await runner.upload(file.source, file.target);
          }
        }
      }

      await this.uploader.cleanup(this.conf);

      for (const [key, server] of servers) {
        const runner = runners.get(key);
        if (runner === undefined) {
          continue;
        }
        for (const command of deploy.commands) {
          let cmd = command.run;
          if (deploy.shell_init !== "") {
            cmd = `${deploy.shell_init} && ${cmd}`;
          }
          if (deploy.cwd !== "") {
            cmd = `cd ${deploy.cwd} && ${cmd}`;
          }
          this.log("🏃", [`[${serverName(server)}]`, command.run]);
          try {
            await runner.pipeExec(cmd);
          } catch (e) {
            throw new Error(`server execute error: ${e instanceof Error ? e.message : String(e)}`);
          }
        }
      }
    } finally {
      pool.close();
    }
  }
}

