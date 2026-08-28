import { REMOTE_META_DIR } from "../../common/const.js";
import { nextBackupName } from "../domain/backup.js";
import { classifyConflicts } from "../domain/conflict.js";
import type { RemoteExec, RemoteFS, SnapshotRepository, UserPrompter } from "../domain/interfaces.js";
import { type FileRecord, type Snapshot, addEntry, emptySnapshot, newSnapshotEntry } from "../domain/snapshot.js";

export interface DeployBundle {
  /** Path of the uploaded tar.gz on the remote server. */
  bundleRemotePath: string;
  targetDir: string;
  bundleName: string;
  /** SHA256 of the bundle. */
  bundleHash: string;
}

/** Extract → conflict resolution → move → snapshot update. Both SFTP and cloud-storage paths converge here. */
export class DeployService {
  constructor(
    private readonly fs: RemoteFS,
    private readonly exec: RemoteExec,
    private readonly snapshots: SnapshotRepository,
    private readonly prompter: UserPrompter,
    private readonly now: () => Date,
  ) {}

  async deploy(bundle: DeployBundle): Promise<void> {
    const { targetDir } = bundle;
    const metaDir = `${targetDir}/${REMOTE_META_DIR}`;
    const tmpDir = `${metaDir}/tmp`;

    await this.fs.mkdirAll(metaDir);
    await this.exec.exec(`rm -rf ${tmpDir} && mkdir -p ${tmpDir} && tar -xzf ${bundle.bundleRemotePath} -C ${tmpDir}`);
    try {
      const extracted = await this.fs.readDir(tmpDir);

      const conflicts: string[] = [];
      for (const f of extracted) {
        if ((await this.fs.stat(`${targetDir}/${f}`)) !== null) {
          conflicts.push(f);
        }
      }

      const existing = await this.snapshots.read(targetDir);
      if (conflicts.length > 0) {
        await this.resolveConflicts(targetDir, conflicts, existing);
      }

      for (const f of extracted) {
        await this.exec.exec(`mv ${tmpDir}/${f} ${targetDir}/${f}`);
      }

      const records: FileRecord[] = [];
      for (const f of extracted) {
        const path = `${targetDir}/${f}`;
        // Directories have no content hash; nest records "" for them (hash, _ := FileHash).
        const info = await this.fs.stat(path);
        const hash = info !== null && !info.isDir ? await this.fs.fileHash(path) : "";
        records.push({ path: f, hash, mod_time: (await this.fs.fileModTime(path)).toISOString() });
      }

      const base: Snapshot = existing ?? emptySnapshot();
      const next = addEntry(base, newSnapshotEntry(bundle.bundleName, bundle.bundleHash, records, this.now()));
      await this.snapshots.write(targetDir, next);
      process.stdout.write("  ✅ Deploy complete\n");
    } finally {
      await this.exec.exec(`rm -rf ${tmpDir}`);
    }
  }

  private async resolveConflicts(targetDir: string, conflicts: string[], snapshot: Snapshot | null): Promise<void> {
    const result = classifyConflicts(conflicts, snapshot);

    for (const f of result.managedFiles) {
      await this.fs.remove(`${targetDir}/${f}`);
    }

    if (result.unmanagedFiles.length === 0) {
      return;
    }
    const decision = await this.prompter.askConflictAction(result.unmanagedFiles);
    for (const f of result.unmanagedFiles) {
      const path = `${targetDir}/${f}`;
      if (decision.action === "backup") {
        const siblings = new Set(await this.fs.readDir(targetDir));
        const backupName = nextBackupName(f, decision.suffix, (c) => siblings.has(c));
        process.stdout.write(`  📦 Backup: ${f} → ${backupName}\n`);
        await this.fs.rename(path, `${targetDir}/${backupName}`);
      } else {
        process.stdout.write(`  🗑  Remove: ${f}\n`);
        await this.fs.remove(path);
      }
    }
  }
}
