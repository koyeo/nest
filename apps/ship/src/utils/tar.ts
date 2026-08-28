import { lstatSync, statSync } from "node:fs";
import { basename, dirname, join } from "node:path";
import { create } from "tar";

/**
 * Compress a file or directory into a gzip tarball at dest.
 * The archive's top-level entry is basename(source); symlinks are dereferenced,
 * broken symlinks are skipped with a warning (matches nest's utils/_tar).
 */
export async function compress(source: string, dest: string): Promise<void> {
  const cwd = dirname(source);
  await create(
    {
      gzip: true,
      file: dest,
      cwd,
      follow: true,
      filter: (path: string): boolean => {
        const full = join(cwd, path);
        if (!lstatSync(full).isSymbolicLink()) {
          return true;
        }
        try {
          statSync(full);
          return true;
        } catch {
          process.stderr.write(`⚠ skipping broken symlink: ${full}\n`);
          return false;
        }
      },
    },
    [basename(source)],
  );
}
