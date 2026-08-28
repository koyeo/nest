import { type Snapshot, isManaged } from "./snapshot.js";

export type ConflictAction = "backup" | "remove";

export interface ConflictResult {
  /** Tracked in the snapshot — safe to replace directly. */
  managedFiles: string[];
  /** Not in the snapshot — require a user decision. */
  unmanagedFiles: string[];
}

export function classifyConflicts(conflicts: string[], snapshot: Snapshot | null): ConflictResult {
  const result: ConflictResult = { managedFiles: [], unmanagedFiles: [] };
  for (const f of conflicts) {
    if (isManaged(snapshot, f)) {
      result.managedFiles.push(f);
    } else {
      result.unmanagedFiles.push(f);
    }
  }
  return result;
}
