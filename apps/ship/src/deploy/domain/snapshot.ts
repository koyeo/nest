import { z } from "zod";

export const fileRecordSchema = z.object({
  path: z.string(),
  hash: z.string(),
  mod_time: z.string(),
});
export type FileRecord = z.infer<typeof fileRecordSchema>;

export const snapshotEntrySchema = z.object({
  bundle_name: z.string(),
  bundle_hash: z.string(),
  deployed_at: z.string(),
  files: z.array(fileRecordSchema).default([]),
});
export type SnapshotEntry = z.infer<typeof snapshotEntrySchema>;

/** Deployment metadata stored at <targetDir>/.ship/snapshot.json on the remote server. */
export const snapshotSchema = z.object({
  entries: z.array(snapshotEntrySchema).default([]),
});
export type Snapshot = z.infer<typeof snapshotSchema>;

export function emptySnapshot(): Snapshot {
  return { entries: [] };
}

/** True when filename appears in any entry of the snapshot. */
export function isManaged(snapshot: Snapshot | null, filename: string): boolean {
  if (snapshot === null) {
    return false;
  }
  return snapshot.entries.some((e) => e.files.some((f) => f.path === filename));
}

/** Upsert by bundle_name: replace the existing entry or append. Returns a new snapshot. */
export function addEntry(snapshot: Snapshot, entry: SnapshotEntry): Snapshot {
  const idx = snapshot.entries.findIndex((e) => e.bundle_name === entry.bundle_name);
  if (idx === -1) {
    return { entries: [...snapshot.entries, entry] };
  }
  const entries = [...snapshot.entries];
  entries[idx] = entry;
  return { entries };
}

export function newSnapshotEntry(bundleName: string, bundleHash: string, files: FileRecord[], now: Date): SnapshotEntry {
  return { bundle_name: bundleName, bundle_hash: bundleHash, deployed_at: now.toISOString(), files };
}
