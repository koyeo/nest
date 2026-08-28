import { Command } from "commander";
import { APP_NAME } from "../common/const.js";

export function versionCommand(version: string): Command {
  return new Command("version")
    .description("Print version")
    .action(() => {
      process.stdout.write(`${APP_NAME} ${version}\n`);
    });
}
