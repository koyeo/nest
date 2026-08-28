import { describe, expect, it } from "vitest";
import { nextBackupName } from "./backup.js";
import { classifyConflicts } from "./conflict.js";
import { type Snapshot, addEntry, emptySnapshot, isManaged, newSnapshotEntry } from "./snapshot.js";

const snap = (files: string[][]): Snapshot => ({
  entries: files.map((fs, i) => ({
    bundle_name: `b${i}`,
    bundle_hash: "",
    deployed_at: "",
    files: fs.map((path) => ({ path, hash: "", mod_time: "" })),
  })),
});

describe("nextBackupName", () => {
  const taken = (names: string[]) => (n: string) => names.includes(n);
  it("no conflict", () => expect(nextBackupName("foo", ".bak", taken([]))).toBe("foo.bak"));
  it("first conflict", () => expect(nextBackupName("foo", ".bak", taken(["foo.bak"]))).toBe("foo.bak.2"));
  it("multi conflict", () => expect(nextBackupName("foo", ".bak", taken(["foo.bak", "foo.bak.2"]))).toBe("foo.bak.3"));
  it("custom suffix", () => expect(nextBackupName("foo", ".old", taken([]))).toBe("foo.old"));
  it("high sequence", () => {
    const names = ["data.bak", ...Array.from({ length: 9 }, (_, i) => `data.bak.${i + 2}`)];
    expect(nextBackupName("data", ".bak", taken(names))).toBe("data.bak.11");
  });
  it("gap in sequence", () => expect(nextBackupName("config", ".bak", taken(["config.bak", "config.bak.3"]))).toBe("config.bak.2"));
});

describe("classifyConflicts", () => {
  it("no snapshot", () => {
    const r = classifyConflicts(["a.js", "b.js"], null);
    expect(r.managedFiles).toEqual([]);
    expect(r.unmanagedFiles).toEqual(["a.js", "b.js"]);
  });
  it("all managed", () => {
    const r = classifyConflicts(["a.js", "b.js"], snap([["a.js", "b.js"]]));
    expect(r.managedFiles).toHaveLength(2);
    expect(r.unmanagedFiles).toHaveLength(0);
  });
  it("all unmanaged", () => {
    const r = classifyConflicts(["a.js", "b.js"], snap([["x.js"]]));
    expect(r.managedFiles).toHaveLength(0);
    expect(r.unmanagedFiles).toHaveLength(2);
  });
  it("mixed", () => {
    const r = classifyConflicts(["a.js", "b.js", "c.js"], snap([["a.js"]]));
    expect(r.managedFiles).toEqual(["a.js"]);
    expect(r.unmanagedFiles).toEqual(["b.js", "c.js"]);
  });
  it("empty conflicts", () => {
    const r = classifyConflicts([], snap([["a.js"]]));
    expect(r.managedFiles).toEqual([]);
    expect(r.unmanagedFiles).toEqual([]);
  });
});

describe("snapshot", () => {
  it("isManaged file in entry", () => expect(isManaged(snap([["app.js", "index.html"]]), "app.js")).toBe(true));
  it("isManaged file not in entry", () => expect(isManaged(snap([["app.js"]]), "other.js")).toBe(false));
  it("isManaged null snapshot", () => expect(isManaged(null, "anything")).toBe(false));
  it("isManaged multiple entries", () => {
    const s = snap([["old.js"], ["new.js"]]);
    expect(isManaged(s, "old.js")).toBe(true);
    expect(isManaged(s, "new.js")).toBe(true);
  });
  it("addEntry appends then replaces by bundle_name", () => {
    const now = new Date("2026-01-01T00:00:00Z");
    let s = emptySnapshot();
    expect(s.entries).toHaveLength(0);
    s = addEntry(s, newSnapshotEntry("app.tar.gz", "h1", [], now));
    expect(s.entries).toHaveLength(1);
    s = addEntry(s, newSnapshotEntry("other.tar.gz", "h2", [], now));
    expect(s.entries).toHaveLength(2);
    s = addEntry(s, newSnapshotEntry("app.tar.gz", "h3", [], now));
    expect(s.entries).toHaveLength(2);
    expect(s.entries[0]?.bundle_hash).toBe("h3");
    expect(s.entries[0]?.deployed_at).toBe("2026-01-01T00:00:00.000Z");
  });
});
