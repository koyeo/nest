import { REMOTE_META_DIR, REMOTE_SNAPSHOT_FILE } from "../../common/const.js";
import type { RemoteFS, SnapshotRepository } from "../domain/interfaces.js";
import { type Snapshot, snapshotSchema } from "../domain/snapshot.js";

export class SnapshotRepo implements SnapshotRepository {
  constructor(private readonly fs: RemoteFS) {}

  async read(targetDir: string): Promise<Snapshot | null> {
    const path = `${targetDir}/${REMOTE_SNAPSHOT_FILE}`;
    if ((await this.fs.stat(path)) === null) {
      return null;
    }
    const text = await this.fs.readFile(path);
    let parsed;
    try {
      parsed = JSON.parse(text);
    } catch (e) {
      throw new Error(`decode snapshot error: ${e instanceof Error ? e.message : String(e)}`);
    }
    const result = snapshotSchema.safeParse(parsed);
    if (!result.success) {
      throw new Error(`decode snapshot error: ${result.error.message}`);
    }
    return result.data;
  }

  async write(targetDir: string, snapshot: Snapshot): Promise<void> {
    await this.fs.mkdirAll(`${targetDir}/${REMOTE_META_DIR}`);
    await this.fs.writeFile(`${targetDir}/${REMOTE_SNAPSHOT_FILE}`, JSON.stringify(snapshot, null, 2));
  }
}
