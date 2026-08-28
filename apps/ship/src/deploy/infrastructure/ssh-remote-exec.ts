import type { RemoteExec } from "../domain/interfaces.js";
import type { SshServer } from "../../execer/ssh-server.js";

export class SshRemoteExec implements RemoteExec {
  constructor(private readonly server: SshServer) {}

  async exec(command: string): Promise<void> {
    await this.server.combinedExec(command);
  }

  execPipe(command: string): Promise<void> {
    return this.server.pipeExec(command);
  }
}
