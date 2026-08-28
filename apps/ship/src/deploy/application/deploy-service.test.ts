import { describe, expect, it } from "vitest";
import type { ConflictDecision } from "../domain/interfaces.js";
import { FakeExec, FakeFS, FakePrompter, FakeSnapshotRepo } from "../test-fakes.js";
import { DeployService } from "./deploy-service.js";

const NOW = new Date("2026-01-01T00:00:00Z");

function setup(decision: ConflictDecision) {
  const fs = new FakeFS();
  const exec = new FakeExec(fs);
  const repo = new FakeSnapshotRepo();
  const svc = new DeployService(fs, exec, repo, new FakePrompter(decision), () => NOW);
  return { fs, exec, repo, svc };
}
const backup: ConflictDecision = { action: "backup", suffix: ".bak" };
const bundle = (name: string, hash: string) => ({ bundleRemotePath: "/target/bundle.tar.gz", targetDir: "/target", bundleName: name, bundleHash: hash });

describe("DeployService", () => {
  it("empty dir → snapshot created with 1 entry", async () => {
    const { fs, repo, svc } = setup(backup);
    fs.files.set("/target/.ship/tmp/app.js", "content");
    await svc.deploy(bundle("app.tar.gz", "hash123"));
    expect(repo.snapshots.get("/target")?.entries).toHaveLength(1);
    expect(fs.files.get("/target/app.js")).toBe("content");
  });

  it("managed conflict → replaced silently, no backup", async () => {
    const { fs, repo, svc } = setup(backup);
    fs.files.set("/target/app.js", "old");
    repo.snapshots.set("/target", { entries: [{ bundle_name: "", bundle_hash: "", deployed_at: "", files: [{ path: "app.js", hash: "", mod_time: "" }] }] });
    fs.files.set("/target/.ship/tmp/app.js", "new");
    await svc.deploy(bundle("app.tar.gz", "hash123"));
    expect(await fs.stat("/target/app.js.bak")).toBeNull();
    expect(fs.files.get("/target/app.js")).toBe("new");
  });

  it("unmanaged conflict with snapshot → backed up", async () => {
    const { fs, repo, svc } = setup(backup);
    fs.files.set("/target/config.yml", "secret");
    repo.snapshots.set("/target", { entries: [{ bundle_name: "", bundle_hash: "", deployed_at: "", files: [{ path: "app.js", hash: "", mod_time: "" }] }] });
    fs.files.set("/target/.ship/tmp/config.yml", "new config");
    await svc.deploy(bundle("app.tar.gz", "hash123"));
    expect(fs.files.get("/target/config.yml.bak")).toBe("secret");
  });

  it("no snapshot with conflict → backed up; second time → .bak.2", async () => {
    const { fs, svc } = setup(backup);
    fs.files.set("/target/app.js", "old");
    fs.files.set("/target/.ship/tmp/app.js", "new");
    await svc.deploy(bundle("app.tar.gz", "hash123"));
    expect(fs.files.get("/target/app.js.bak")).toBe("old");
    // simulate an unmanaged file reappearing (snapshot now tracks app.js, so plant a different name)
    fs.files.set("/target/x.txt", "1");
    fs.files.set("/target/x.txt.bak", "0");
    fs.files.set("/target/.ship/tmp/x.txt", "2");
    await svc.deploy(bundle("x.tar.gz", "h"));
    expect(fs.files.get("/target/x.txt.bak.2")).toBe("1");
  });

  it("user chooses remove → no backup", async () => {
    const { fs, svc } = setup({ action: "remove", suffix: "" });
    fs.files.set("/target/old.txt", "data");
    fs.files.set("/target/.ship/tmp/old.txt", "new data");
    await svc.deploy(bundle("app.tar.gz", "hash123"));
    expect(await fs.stat("/target/old.txt.bak")).toBeNull();
    expect(fs.files.get("/target/old.txt")).toBe("new data");
  });

  it("snapshot written with hash and file records", async () => {
    const { fs, repo, svc } = setup(backup);
    fs.files.set("/target/.ship/tmp/index.html", "<html>");
    await svc.deploy(bundle("bundle.tar.gz", "hash999"));
    const entry = repo.snapshots.get("/target")?.entries[0];
    expect(entry?.bundle_hash).toBe("hash999");
    expect(entry?.files.map((f) => f.path)).toEqual(["index.html"]);
    expect(entry?.deployed_at).toBe(NOW.toISOString());
  });

  it("snapshot preserves history across bundle names", async () => {
    const { fs, repo, svc } = setup(backup);
    fs.files.set("/target/.ship/tmp/v1.js", "v1");
    await svc.deploy(bundle("v1.tar.gz", "hash1"));
    fs.files.set("/target/.ship/tmp/v2.js", "v2");
    await svc.deploy(bundle("v2.tar.gz", "hash2"));
    expect(repo.snapshots.get("/target")?.entries).toHaveLength(2);
  });

  it("tmp dir cleaned even on failure", async () => {
    const fs = new FakeFS();
    class FailingExec extends FakeExec {
      override async exec(command: string): Promise<void> {
        await super.exec(command);
        if (command.startsWith("mv ")) {
          throw new Error("mv failed");
        }
      }
    }
    const exec = new FailingExec(fs);
    const svc = new DeployService(fs, exec, new FakeSnapshotRepo(), new FakePrompter(backup), () => NOW);
    fs.files.set("/target/.ship/tmp/a", "1");
    await expect(svc.deploy(bundle("a.tar.gz", "h"))).rejects.toThrow("mv failed");
    expect(exec.commands.at(-1)).toBe("rm -rf /target/.ship/tmp");
  });
});
