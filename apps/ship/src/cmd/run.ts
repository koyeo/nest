import { Command } from "commander";
import * as logger from "../logger.js";
import { loadConfig } from "../protocol/load.js";
import { TaskRunner } from "../runner/task-runner.js";

export function runCommand(getConfigFile: () => string): Command {
  return new Command("run")
    .description("Execute one or more tasks defined in ship.yaml")
    .argument("<tasks...>", "task names, executed in order")
    .action(async (tasks: string[]) => {
      const conf = loadConfig(getConfigFile());
      for (const name of tasks) {
        const task = conf.tasks[name];
        if (task === undefined) {
          throw new Error(`task: ${name} not found`);
        }
        const runner = TaskRunner.root(conf, task, name);
        runner.printStart();
        try {
          await runner.exec();
        } catch (e) {
          runner.printFailed();
          logger.error(e instanceof Error ? e : new Error(String(e)));
          process.exit(1);
        }
        runner.printSuccess();
      }
    });
}
