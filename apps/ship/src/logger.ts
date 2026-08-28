import pc from "picocolors";
import { APP_NAME } from "./common/const.js";

function timestamp(): string {
  const d = new Date();
  const p = (n: number): string => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

export function step(taskKey: string, taskComment: string, emoji: string, args: string[]): void {
  const label = taskComment !== "" ? taskComment : taskKey;
  const body = args.filter((a) => a !== "").join(" ");
  process.stdout.write(`[${APP_NAME}] ${pc.white(timestamp())} ${pc.bgWhite(pc.black(`[${label}]`))} ${emoji} ${body}\n`);
}

export function error(err: Error): void {
  process.stdout.write(`[${APP_NAME}] ${pc.white(timestamp())} ${pc.red(err.message)}\n`);
}
