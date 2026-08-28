/** Keeps the last `max` bytes written. Used to attach a stderr tail to errors. */
export class TailBuffer {
  private buf: Buffer = Buffer.alloc(0);
  constructor(private readonly max: number) {}

  write(chunk: Buffer): void {
    this.buf = Buffer.concat([this.buf, chunk]);
    if (this.buf.length > this.max) {
      this.buf = this.buf.subarray(this.buf.length - this.max);
    }
  }

  toString(): string {
    return this.buf.toString("utf8");
  }
}

export function withTail(message: string, tail: TailBuffer): string {
  const t = tail.toString().trim();
  return t !== "" ? `${message}\n${t}` : message;
}
