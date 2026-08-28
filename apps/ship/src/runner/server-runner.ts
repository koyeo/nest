import { createReadStream, existsSync, statSync } from "node:fs";
import { basename, posix } from "node:path";
import pc from "picocolors";
import { REMOTE_TMP_BUNDLE_PREFIX } from "../common/const.js";
import { DeployService } from "../deploy/application/deploy-service.js";
import { SnapshotRepo } from "../deploy/infrastructure/snapshot-repo.js";
import { SshRemoteExec } from "../deploy/infrastructure/ssh-remote-exec.js";
import { SshRemoteFS } from "../deploy/infrastructure/ssh-remote-fs.js";
import { StdinPrompter } from "../deploy/infrastructure/stdin-prompter.js";
import type { ServerPool } from "../execer/server-pool.js";
import type { SshServer } from "../execer/ssh-server.js";
import { type Server, serverName } from "../protocol/schema.js";
import type { ObjectStorage } from "../storage/storage.js";
import { byteSize } from "../utils/byte-size.js";
import { digestFile } from "../utils/hash.js";
import { compress } from "../utils/tar.js";
import type { CloudUploader } from "./cloud.js";
import { cleanShipTmpDir, shipTmpDir } from "./tmp-dir.js";

const PRESIGN_EXPIRES_SECONDS = 3600;
const PROGRESS_CHUNK = 1024 * 1024;

/** "~" targets and absolute paths shallower than 2 segments are rejected (e.g. "/data"). */
export function checkTargetPath(target: string): void {
  if (target.startsWith("~")) {
    throw new Error(`invalid target path: '${target}'`);
  }
  if (target.startsWith("/") && target.replace(/^\//, "").split("/").length < 2) {
    throw new Error(`target path: ${target} too short`);
  }
}

/** Trailing "/" means target is a directory; otherwise the parent is the directory. */
export function targetDirOf(target: string): string {
  return target.endsWith("/") ? target.replace(/\/+$/, "") : posix.dirname(target);
}

export class ServerRunner {
  private readonly ssh: SshServer;

  constructor(
    private readonly server: Server,
    private readonly key: string,
    pool: ServerPool,
    private readonly log: (emoji: string, args: string[]) => void,
  ) {
    const identityFile = server.password === "" && server.identity_file === "" ? "~/.ssh/id_rsa" : server.identity_file;
    this.ssh = pool.get({ key, host: server.host, port: server.port, user: server.user, password: server.password, identityFile });
  }

  pipeExec(command: string): Promise<void> {
    return this.ssh.pipeExec(command);
  }

  combinedExec(command: string): Promise<string> {
    return this.ssh.combinedExec(command);
  }

  /** Direct SFTP path: compress → stream upload with progress → deployBundle. */
  async upload(source: string, target: string): Promise<void> {
    checkTargetPath(target);
    if (!existsSync(source)) {
      throw new Error(`upload source: ${source} not exists`);
    }
    const targetDir = targetDirOf(target);
    const remoteFs = new SshRemoteFS(this.ssh);
    await remoteFs.mkdirAll(targetDir);

    const sourceName = basename(source);
    const targetName = target.endsWith("/") ? sourceName : posix.basename(target);
    const bundleName = `${sourceName}.tar.gz`;
    const bundleLocalPath = `${shipTmpDir()}/${bundleName}`;
    const bundleRemoteTmpPath = `${targetDir}/bundle-${bundleName}~`;

    try {
      await compress(source, bundleLocalPath);
      const digest = await digestFile(bundleLocalPath);
      this.log("🚀", [pc.cyan(`[${serverName(this.server)}]`), pc.bold(pc.magenta(`${source} ===> ${posix.join(targetDir, targetName)}`))]);
      await this.sftpUploadWithProgress(bundleLocalPath, bundleRemoteTmpPath, statSync(bundleLocalPath).size);
      try {
        await this.deployBundle(bundleRemoteTmpPath, targetDir, bundleName, digest.sha256);
      } finally {
        await remoteFs.remove(bundleRemoteTmpPath);
      }
    } finally {
      cleanShipTmpDir();
    }
  }

  private async sftpUploadWithProgress(localPath: string, remotePath: string, total: number): Promise<void> {
    const sftp = await this.ssh.sftp();
    await new Promise<void>((resolve, reject) => {
      const ws = sftp.createWriteStream(remotePath);
      let uploaded = 0;
      let lastPrinted = -PROGRESS_CHUNK;
      ws.on("error", (e: Error) => reject(new Error(`upload write remote bundle error: ${e.message}`)));
      ws.on("close", () => {
        process.stdout.write("\n");
        resolve();
      });
      createReadStream(localPath, { highWaterMark: PROGRESS_CHUNK })
        .on("data", (chunk: Buffer | string) => {
          uploaded += chunk.length;
          if (uploaded - lastPrinted >= PROGRESS_CHUNK) {
            lastPrinted = uploaded;
            process.stdout.write(`\rTotal: ${byteSize(total)} Uploaded: ${byteSize(uploaded)}`);
          }
        })
        .on("error", (e) => reject(new Error(`read local bundle error: ${e.message}`)))
        .pipe(ws);
    });
  }

  /** Cloud relay path: upload once → presigned URL → remote curl → deployBundle → rm. */
  async deployViaStorage(uploader: CloudUploader, store: ObjectStorage, alias: string, source: string, target: string): Promise<void> {
    checkTargetPath(target);
    const { objectKey, bundleHash } = await uploader.upload(store, alias, source);
    const url = await store.presignedUrl(objectKey, PRESIGN_EXPIRES_SECONDS);

    const sourceName = basename(source);
    const bundleName = `${sourceName}.tar.gz`;
    const bundleRemotePath = `${REMOTE_TMP_BUNDLE_PREFIX}${bundleName}`;
    this.log("⬇️", [`[${serverName(this.server)}]`, `downloading ${sourceName} via ${alias}`]);
    await this.ssh.pipeExec(`curl -fsSL '${url}' -o ${bundleRemotePath}`);

    const targetDir = targetDirOf(target);
    try {
      await this.deployBundle(bundleRemotePath, targetDir, bundleName, bundleHash);
    } finally {
      await this.ssh.combinedExec(`rm -f ${bundleRemotePath}`);
    }
    this.log("✅", [`deployed ${sourceName} → ${targetDir}`]);
  }

  private deployBundle(bundleRemotePath: string, targetDir: string, bundleName: string, bundleHash: string): Promise<void> {
    const remoteFs = new SshRemoteFS(this.ssh);
    const svc = new DeployService(remoteFs, new SshRemoteExec(this.ssh), new SnapshotRepo(remoteFs), new StdinPrompter(), () => new Date());
    return svc.deploy({ bundleRemotePath, targetDir, bundleName, bundleHash });
  }
}
