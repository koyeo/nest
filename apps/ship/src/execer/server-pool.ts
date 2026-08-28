import { SshServer, type SshTarget } from "./ssh-server.js";

/** Reuses one SshServer per target key for the lifetime of a deploy step. */
export class ServerPool {
  private readonly servers = new Map<string, SshServer>();

  get(target: SshTarget): SshServer {
    const existing = this.servers.get(target.key);
    if (existing !== undefined) {
      return existing;
    }
    const created = new SshServer(target);
    this.servers.set(target.key, created);
    return created;
  }

  close(): void {
    for (const s of this.servers.values()) {
      s.close();
    }
    this.servers.clear();
  }
}
