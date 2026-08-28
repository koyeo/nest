import { describe, expect, it } from "vitest";
import { FakeFS } from "../test-fakes.js";
import { SnapshotRepo } from "./snapshot-repo.js";

describe("SnapshotRepo", () => {
  it("read: no file → null", async () => {
    expect(await new SnapshotRepo(new FakeFS()).read("/target")).toBeNull();
  });

  it("read: valid json", async () => {
    const fs = new FakeFS();
    fs.files.set("/target/.ship/snapshot.json", JSON.stringify({ entries: [{ bundle_name: "app.tar.gz", bundle_hash: "abc123", deployed_at: "", files: [] }] }));
    const snap = await new SnapshotRepo(fs).read("/target");
    expect(snap?.entries).toHaveLength(1);
    expect(snap?.entries[0]?.bundle_name).toBe("app.tar.gz");
  });

  it("read: invalid json → error", async () => {
    const fs = new FakeFS();
    fs.files.set("/target/.ship/snapshot.json", "{invalid json");
    await expect(new SnapshotRepo(fs).read("/target")).rejects.toThrow("decode snapshot error");
  });

  it("write: creates .ship dir", async () => {
    const fs = new FakeFS();
    await new SnapshotRepo(fs).write("/target", { entries: [] });
    expect(fs.dirs.has("/target/.ship")).toBe(true);
  });

  it("write/read round trip", async () => {
    const fs = new FakeFS();
    const repo = new SnapshotRepo(fs);
    await repo.write("/target", {
      entries: [{ bundle_name: "test.tar.gz", bundle_hash: "deadbeef", deployed_at: "", files: [{ path: "main.go", hash: "abc123", mod_time: "" }] }],
    });
    const back = await repo.read("/target");
    expect(back?.entries[0]?.bundle_name).toBe("test.tar.gz");
    expect(back?.entries[0]?.files[0]?.path).toBe("main.go");
  });
});
