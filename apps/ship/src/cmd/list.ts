import { Command } from "commander";
import pc from "picocolors";
import { loadConfig } from "../protocol/load.js";

function title(s: string): string {
  return pc.bold(pc.green(s));
}

export function listCommand(getConfigFile: () => string): Command {
  return new Command("list")
    .description("Display tasks, servers, and environment variables from the config")
    .action(() => {
      const conf = loadConfig(getConfigFile());
      const out: string[] = [`${title("version:")} ${conf.version}`];
      const tasks = Object.entries(conf.tasks);
      if (tasks.length > 0) {
        out.push(`${title("tasks:")} `);
        for (const [key, task] of tasks) {
          out.push(`  ${pc.cyan(key.padEnd(25))} ${pc.white(task.comment)}`);
        }
      }
      const envs = Object.entries(conf.envs);
      if (envs.length > 0) {
        out.push(`${title("envs:")} `);
        for (const [key, value] of envs) {
          out.push(`  ${pc.cyan(key.padEnd(25))} ${value}`);
        }
      }
      const servers = Object.entries(conf.servers);
      if (servers.length > 0) {
        out.push(`${title("servers:")} `);
        for (const [key, server] of servers) {
          out.push(`  ${pc.cyan(`${key}(${server.host})`.padEnd(35))} ${pc.white(server.comment)}`);
        }
      }
      process.stdout.write(out.join("\n") + "\n");
    });
}
