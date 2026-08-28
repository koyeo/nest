import type { ConflictDecision, UserPrompter } from "../domain/interfaces.js";
import { ask } from "../../utils/prompt.js";

/** Interactive conflict resolution on the terminal. Enter = backup, empty suffix = ".bak". */
export class StdinPrompter implements UserPrompter {
  async askConflictAction(files: string[]): Promise<ConflictDecision> {
    process.stdout.write(`⚠ Unmanaged file conflicts in target directory: ${files.join(", ")}\n`);
    process.stdout.write("  [1] Backup (default)\n  [2] Remove\n");
    const choice = await ask("Choose", "1");
    if (choice === "2") {
      return { action: "remove", suffix: "" };
    }
    const suffix = await ask("Backup suffix", ".bak");
    return { action: "backup", suffix };
  }
}
