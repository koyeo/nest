import { Command } from "commander";
import { APP_NAME, DEFAULT_CONFIG_FILE } from "./common/const.js";
import { initCommand } from "./cmd/init.js";
import { listCommand } from "./cmd/list.js";
import { runCommand } from "./cmd/run.js";
import { storageCommand } from "./cmd/storage.js";
import { versionCommand } from "./cmd/version.js";
import * as logger from "./logger.js";
import { version } from "./version.js";

const program = new Command(APP_NAME)
  .description("Task runner & deployment CLI for local builds, remote execution, and server deploys")
  .option("-c, --config <path>", `Path to ${APP_NAME} config file`, DEFAULT_CONFIG_FILE)
  .version(version, "-v, --version");

const getConfigFile = (): string => {
  const parsed: { config: string } = program.opts();
  return parsed.config;
};

program.addCommand(initCommand());
program.addCommand(listCommand(getConfigFile));
program.addCommand(runCommand(getConfigFile));
program.addCommand(storageCommand(getConfigFile));
program.addCommand(versionCommand(version));

program.parseAsync(process.argv).catch((e) => {
  logger.error(e instanceof Error ? e : new Error(String(e)));
  process.exit(1);
});
