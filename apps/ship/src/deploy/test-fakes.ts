import type { ConflictDecision, RemoteExec, RemoteFS, RemoteFileInfo, SnapshotRepository, UserPrompter } from "./domain/interfaces.js";
import type { Snapshot } from "./domain/snapshot.js";

/** In-memory RemoteFS: files map path → content, dirs set. */
export class FakeFS implements RemoteFS {
  readonly files = new Map<string, string>();
  readonly dirs = new Set<string>();

  async stat(path: string): Promise<RemoteFileInfo | null> {
    const content = this.files.get(path);
    if (content !== undefined) {
      return { name: path, size: content.length, isDir: false, modTime: new Date(0) };
    }
    if (this.dirs.has(path)) {
      return { name: path, size: 0, isDir: true, modTime: new Date(0) };
    }
    return null;
  }
  async mkdirAll(path: string): Promise<void> {
    this.dirs.add(path);
  }
  async readDir(path: string): Promise<string[]> {
    const prefix = `${path}/`;
    const names: string[] = [];
    for (const p of this.files.keys()) {
      if (p.startsWith(prefix)) {
        const rest = p.slice(prefix.length);
        if (!rest.includes("/")) {
          names.push(rest);
        }
      }
    }
    return names;
  }
  async readFile(path: string): Promise<string> {
    const c = this.files.get(path);
    if (c === undefined) {
      throw new Error(`not found: ${path}`);
    }
    return c;
  }
  async writeFile(path: string, data: string): Promise<void> {
    this.files.set(path, data);
  }
  async remove(path: string): Promise<void> {
    this.files.delete(path);
    this.dirs.delete(path);
  }
  async rename(src: string, dst: string): Promise<void> {
    const c = this.files.get(src);
    if (c === undefined) {
      throw new Error(`not found: ${src}`);
    }
    this.files.set(dst, c);
    this.files.delete(src);
  }
  async fileHash(): Promise<string> {
    return "mockhash";
  }
  async fileModTime(): Promise<Date> {
    return new Date(0);
  }
}

/** Records commands; emulates `mv a b` against a FakeFS so deploy tests see files land. */
export class FakeExec implements RemoteExec {
  readonly commands: string[] = [];
  constructor(private readonly fs: FakeFS) {}

  async exec(command: string): Promise<void> {
    this.commands.push(command);
    const mv = /^mv (\S+) (\S+)$/.exec(command);
    if (mv !== null && mv[1] !== undefined && mv[2] !== undefined) {
      await this.fs.rename(mv[1], mv[2]);
    }
  }
  async execPipe(command: string): Promise<void> {
    this.commands.push(command);
  }
}

export class FakeSnapshotRepo implements SnapshotRepository {
  readonly snapshots = new Map<string, Snapshot>();
  async read(targetDir: string): Promise<Snapshot | null> {
    return this.snapshots.get(targetDir) ?? null;
  }
  async write(targetDir: string, snapshot: Snapshot): Promise<void> {
    this.snapshots.set(targetDir, snapshot);
  }
}

export class FakePrompter implements UserPrompter {
  constructor(private readonly decision: ConflictDecision) {}
  async askConflictAction(): Promise<ConflictDecision> {
    return this.decision;
  }
}
