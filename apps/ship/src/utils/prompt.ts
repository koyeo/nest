import { type Interface, createInterface } from "node:readline";
import { stdin, stdout } from "node:process";

interface LineReader {
  rl: Interface;
  queue: string[];
  waiting: ((line: string | null) => void) | null;
  closed: boolean;
}

let shared: LineReader | null = null;

/**
 * One readline for the whole process with a line queue: per-question interfaces
 * would swallow stdin buffered between questions (e.g. `printf '1\n\n' | ship ...`).
 */
function reader(): LineReader {
  if (shared !== null) {
    return shared;
  }
  const state: LineReader = { rl: createInterface({ input: stdin, terminal: false }), queue: [], waiting: null, closed: false };
  state.rl.on("line", (line) => {
    if (state.waiting !== null) {
      const w = state.waiting;
      state.waiting = null;
      w(line);
    } else {
      state.queue.push(line);
    }
  });
  state.rl.on("close", () => {
    state.closed = true;
    if (state.waiting !== null) {
      const w = state.waiting;
      state.waiting = null;
      w(null);
    }
  });
  state.rl.pause();
  shared = state;
  return state;
}

function nextLine(r: LineReader): Promise<string | null> {
  const queued = r.queue.shift();
  if (queued !== undefined) {
    return Promise.resolve(queued);
  }
  if (r.closed) {
    return Promise.resolve(null);
  }
  return new Promise((resolve) => {
    r.waiting = resolve;
    r.rl.resume();
  });
}

/** Ask one line on stdin. Empty input returns defaultValue. Throws when stdin is closed. */
export async function ask(label: string, defaultValue: string): Promise<string> {
  const r = reader();
  stdout.write(defaultValue !== "" ? `${label} [${defaultValue}]: ` : `${label}: `);
  const line = await nextLine(r);
  if (!r.closed) {
    r.rl.pause();
  }
  if (line === null) {
    stdout.write("\n");
    throw new Error("read input error: stdin closed");
  }
  const input = line.trim();
  return input === "" ? defaultValue : input;
}
