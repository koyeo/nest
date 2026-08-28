const UNITS: ReadonlyArray<readonly [string, number]> = [
  ["E", 2 ** 60],
  ["P", 2 ** 50],
  ["T", 2 ** 40],
  ["G", 2 ** 30],
  ["M", 2 ** 20],
  ["K", 2 ** 10],
];

/** Human-readable byte string like "10M" or "12.5K" (matches nest's unit.ByteSize). */
export function byteSize(bytes: number): string {
  for (const [unit, size] of UNITS) {
    if (bytes >= size) {
      return trimZeros((bytes / size).toFixed(1)) + unit;
    }
  }
  return `${bytes}B`;
}

function trimZeros(s: string): string {
  return s.replace(/\.0$/, "");
}
