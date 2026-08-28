/**
 * Non-conflicting backup name: filename+suffix, then filename+suffix.2, .3, ...
 * (sequence 1 is implicit).
 */
export function nextBackupName(filename: string, suffix: string, exists: (candidate: string) => boolean): string {
  let candidate = filename + suffix;
  if (!exists(candidate)) {
    return candidate;
  }
  for (let i = 2; i < 100000; i++) {
    candidate = `${filename}${suffix}.${i}`;
    if (!exists(candidate)) {
      return candidate;
    }
  }
  return candidate;
}
