import { spawn } from "node:child_process";
import { TailBuffer, withTail } from "../utils/tail-buffer.js";

export interface LocalExecOptions {
  command: string;
  cwd: string;
  env: Record<string, string>;
}

/** Run `bash -c command` with stdin inherited and stdout/stderr streamed live. */
export function localExec(opts: LocalExecOptions): Promise<void> {
  return new Promise((resolve, reject) => {
    const child = spawn("bash", ["-c", opts.command], {
      cwd: opts.cwd !== "" ? opts.cwd : process.cwd(),
      env: opts.env,
      stdio: ["inherit", "pipe", "pipe"],
    });
    const tail = new TailBuffer(4096);
    child.stdout.on("data", (chunk: Buffer) => {
      process.stdout.write(chunk);
    });
    child.stderr.on("data", (chunk: Buffer) => {
      process.stderr.write(chunk);
      tail.write(chunk);
    });
    child.on("error", (e) => reject(new Error(`spawn error: ${e.message}`)));
    child.on("close", (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }
      const reason = code !== null ? `exit status ${code}` : `signal ${signal ?? "unknown"}`;
      reject(new Error(withTail(reason, tail)));
    });
  });
}

/** process.env ⊕ global envs ⊕ task envs (later wins). */
export function mergeEnv(globalEnvs: Record<string, string>, taskEnvs: Record<string, string>): Record<string, string> {
  const env: Record<string, string> = {};
  for (const [k, v] of Object.entries(process.env)) {
    if (v !== undefined) {
      env[k] = v;
    }
  }
  return { ...env, ...globalEnvs, ...taskEnvs };
}
