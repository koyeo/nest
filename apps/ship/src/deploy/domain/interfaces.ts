import type { ConflictAction } from "./conflict.js";
import type { Snapshot } from "./snapshot.js";

export interface RemoteFileInfo {
  name: string;
  size: number;
  isDir: boolean;
  modTime: Date;
}

/** Remote filesystem operations (SFTP). */
export interface RemoteFS {
  /** null when the path does not exist. */
  stat(path: string): Promise<RemoteFileInfo | null>;
  mkdirAll(path: string): Promise<void>;
  /** Child names (non-recursive), excluding ship metadata. */
  readDir(path: string): Promise<string[]>;
  readFile(path: string): Promise<string>;
  writeFile(path: string, data: string): Promise<void>;
  /** Removes a file or directory recursively; a missing path is not an error. */
  remove(path: string): Promise<void>;
  rename(src: string, dst: string): Promise<void>;
  /** SHA256 of a remote file. */
  fileHash(path: string): Promise<string>;
  fileModTime(path: string): Promise<Date>;
}

/** Remote command execution (SSH). */
export interface RemoteExec {
  exec(command: string): Promise<void>;
  execPipe(command: string): Promise<void>;
}

export interface ConflictDecision {
  action: ConflictAction;
  /** Only meaningful when action === "backup". */
  suffix: string;
}

export interface UserPrompter {
  askConflictAction(files: string[]): Promise<ConflictDecision>;
}

export interface SnapshotRepository {
  /** null when no snapshot exists. */
  read(targetDir: string): Promise<Snapshot | null>;
  write(targetDir: string, snapshot: Snapshot): Promise<void>;
}
