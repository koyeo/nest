import { mkdirSync, rmSync } from "node:fs";
import { LOCAL_TMP_DIR } from "../common/const.js";

export function shipTmpDir(): string {
  mkdirSync(LOCAL_TMP_DIR, { recursive: true });
  return LOCAL_TMP_DIR;
}

export function cleanShipTmpDir(): void {
  rmSync(LOCAL_TMP_DIR, { recursive: true, force: true });
}
