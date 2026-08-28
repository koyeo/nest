import { createRequire } from "node:module";

const pkg: { version: string } = createRequire(import.meta.url)("../package.json");
export const version: string = pkg.version;
