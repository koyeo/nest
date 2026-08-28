import { readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import { Client, type ConnectConfig, type SFTPWrapper } from "ssh2";
import { TailBuffer, withTail } from "../utils/tail-buffer.js";

export interface SshTarget {
  key: string;
  host: string;
  port: number;
  user: string;
  password: string;
  identityFile: string;
}

const CONNECT_TIMEOUT_MS = 60_000;

/** Wraps a command so /etc/profile, ~/.bash_profile, ~/.bashrc are loaded (PATH, nvm, pyenv...). */
export function wrapLoginShell(command: string): string {
  const escaped = command.replaceAll("'", `'"'"'`);
  return `bash -l -c '${escaped}'`;
}

export function expandHome(path: string): string {
  return path.startsWith("~") ? join(homedir(), path.slice(1)) : path;
}

function connectConfig(t: SshTarget): ConnectConfig {
  const base: ConnectConfig = { host: t.host, port: t.port, username: t.user, readyTimeout: CONNECT_TIMEOUT_MS };
  const withPassword: ConnectConfig = t.password !== "" ? { ...base, password: t.password } : base;
  return t.identityFile !== "" ? { ...withPassword, privateKey: readFileSync(expandHome(t.identityFile)) } : withPassword;
}

/** One SSH connection (+ lazily opened SFTP channel) to a server. */
export class SshServer {
  private client: Client | null = null;
  private sftpWrapper: SFTPWrapper | null = null;

  constructor(readonly target: SshTarget) {}

  async connect(): Promise<Client> {
    if (this.client !== null) {
      return this.client;
    }
    const client = new Client();
    await new Promise<void>((resolve, reject) => {
      client
        .once("ready", () => resolve())
        .once("error", (e) => reject(new Error(`connect server error: ${e.message}`)))
        .connect(connectConfig(this.target));
    });
    this.client = client;
    return client;
  }

  async sftp(): Promise<SFTPWrapper> {
    if (this.sftpWrapper !== null) {
      return this.sftpWrapper;
    }
    const client = await this.connect();
    const sftp = await new Promise<SFTPWrapper>((resolve, reject) => {
      client.sftp((err, s) => {
        if (err) {
          reject(new Error(`init sftp client error: ${err.message}`));
          return;
        }
        resolve(s);
      });
    });
    this.sftpWrapper = sftp;
    return sftp;
  }

  /** Run in a login shell, buffer combined output; on non-zero exit reject with the trimmed output. */
  async combinedExec(command: string): Promise<string> {
    const client = await this.connect();
    return new Promise((resolve, reject) => {
      client.exec(wrapLoginShell(command), (err, stream) => {
        if (err) {
          reject(err);
          return;
        }
        const chunks: Buffer[] = [];
        stream.on("data", (c: Buffer) => chunks.push(c));
        stream.stderr.on("data", (c: Buffer) => chunks.push(c));
        stream.on("close", (code: number) => {
          const out = Buffer.concat(chunks).toString("utf8").trim();
          if (code === 0) {
            resolve(out);
          } else {
            reject(new Error(out !== "" ? out : `exit status ${code}`));
          }
        });
      });
    });
  }

  /** Run in a login shell with stdout/stderr streamed to the terminal. */
  async pipeExec(command: string): Promise<void> {
    const client = await this.connect();
    return new Promise((resolve, reject) => {
      client.exec(wrapLoginShell(command), (err, stream) => {
        if (err) {
          reject(err);
          return;
        }
        const tail = new TailBuffer(4096);
        stream.on("data", (c: Buffer) => {
          process.stdout.write(c);
        });
        stream.stderr.on("data", (c: Buffer) => {
          process.stderr.write(c);
          tail.write(c);
        });
        stream.on("close", (code: number) => {
          if (code === 0) {
            resolve();
          } else {
            reject(new Error(withTail(`session run command error: exit status ${code}`, tail)));
          }
        });
      });
    });
  }

  close(): void {
    this.sftpWrapper?.end();
    this.client?.end();
    this.sftpWrapper = null;
    this.client = null;
  }
}
