import { posix } from "node:path";
import type { SFTPWrapper, Stats } from "ssh2";
import { REMOTE_META_DIR } from "../../common/const.js";
import type { SshServer } from "../../execer/ssh-server.js";
import type { RemoteFS, RemoteFileInfo } from "../domain/interfaces.js";

function sftpStat(sftp: SFTPWrapper, path: string): Promise<Stats | null> {
  return new Promise((resolve, reject) => {
    sftp.stat(path, (err, stats) => {
      if (err) {
        // ssh2 reports SFTP status codes on the error; 2 = NO_SUCH_FILE
        if ("code" in err && err.code === 2) {
          resolve(null);
          return;
        }
        reject(err);
        return;
      }
      resolve(stats);
    });
  });
}

function sftpMkdir(sftp: SFTPWrapper, path: string): Promise<void> {
  return new Promise((resolve, reject) => {
    sftp.mkdir(path, (err) => (err ? reject(err) : resolve()));
  });
}

export class SshRemoteFS implements RemoteFS {
  constructor(private readonly server: SshServer) {}

  async stat(path: string): Promise<RemoteFileInfo | null> {
    const stats = await sftpStat(await this.server.sftp(), path);
    if (stats === null) {
      return null;
    }
    return { name: posix.basename(path), size: stats.size, isDir: stats.isDirectory(), modTime: new Date(stats.mtime * 1000) };
  }

  async mkdirAll(path: string): Promise<void> {
    const sftp = await this.server.sftp();
    const parts = path.split("/").filter((p) => p !== "");
    let current = path.startsWith("/") ? "" : ".";
    for (const part of parts) {
      current = `${current}/${part}`;
      if ((await sftpStat(sftp, current)) === null) {
        await sftpMkdir(sftp, current);
      }
    }
  }

  async readDir(path: string): Promise<string[]> {
    const sftp = await this.server.sftp();
    return new Promise((resolve, reject) => {
      sftp.readdir(path, (err, list) => {
        if (err) {
          reject(new Error(`read remote dir error: ${err.message}`));
          return;
        }
        resolve(list.map((e) => e.filename).filter((n) => !n.startsWith(REMOTE_META_DIR)));
      });
    });
  }

  async readFile(path: string): Promise<string> {
    const sftp = await this.server.sftp();
    return new Promise((resolve, reject) => {
      const chunks: Buffer[] = [];
      sftp
        .createReadStream(path)
        .on("data", (c: Buffer | string) => chunks.push(typeof c === "string" ? Buffer.from(c) : c))
        .on("error", reject)
        .on("end", () => resolve(Buffer.concat(chunks).toString("utf8")));
    });
  }

  async writeFile(path: string, data: string): Promise<void> {
    const sftp = await this.server.sftp();
    return new Promise((resolve, reject) => {
      const ws = sftp.createWriteStream(path);
      ws.on("error", (e: Error) => reject(new Error(`create remote file error: ${e.message}`)));
      ws.on("close", () => resolve());
      ws.end(data);
    });
  }

  async remove(path: string): Promise<void> {
    const sftp = await this.server.sftp();
    const stats = await sftpStat(sftp, path);
    if (stats === null) {
      return;
    }
    if (stats.isDirectory()) {
      await this.server.combinedExec(`rm -rf ${path}`);
      return;
    }
    await new Promise<void>((resolve, reject) => {
      sftp.unlink(path, (err) => (err ? reject(err) : resolve()));
    });
  }

  async rename(src: string, dst: string): Promise<void> {
    const sftp = await this.server.sftp();
    await new Promise<void>((resolve, reject) => {
      sftp.rename(src, dst, (err) => (err ? reject(err) : resolve()));
    });
  }

  async fileHash(path: string): Promise<string> {
    const out = await this.server.combinedExec(`sha256sum ${path} 2>/dev/null || shasum -a 256 ${path}`);
    const first = out.trim().split(/\s+/)[0];
    if (first === undefined || first === "") {
      throw new Error(`unexpected hash output: ${out}`);
    }
    return first;
  }

  async fileModTime(path: string): Promise<Date> {
    const info = await this.stat(path);
    if (info === null) {
      throw new Error(`stat remote file error: ${path} not found`);
    }
    return info.modTime;
  }
}
