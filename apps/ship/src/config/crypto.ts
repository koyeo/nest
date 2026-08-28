import { createCipheriv, createDecipheriv, randomBytes } from "node:crypto";

const NONCE_SIZE = 12;
const TAG_SIZE = 16;

/** Random 32-byte AES-256 key, base64 encoded. */
export function generateEncryptKey(): string {
  return randomBytes(32).toString("base64");
}

/** AES-256-GCM. Output = base64(nonce ‖ ciphertext ‖ tag) — byte-compatible with nest's Go implementation. */
export function encrypt(keyBase64: string, plaintext: string): string {
  const key = Buffer.from(keyBase64, "base64");
  const nonce = randomBytes(NONCE_SIZE);
  const cipher = createCipheriv("aes-256-gcm", key, nonce);
  const body = Buffer.concat([cipher.update(plaintext, "utf8"), cipher.final()]);
  const tag = cipher.getAuthTag();
  return Buffer.concat([nonce, body, tag]).toString("base64");
}

export function decrypt(keyBase64: string, encryptedBase64: string): string {
  const key = Buffer.from(keyBase64, "base64");
  const data = Buffer.from(encryptedBase64, "base64");
  if (data.length < NONCE_SIZE + TAG_SIZE) {
    throw new Error("ciphertext too short");
  }
  const nonce = data.subarray(0, NONCE_SIZE);
  const body = data.subarray(NONCE_SIZE, data.length - TAG_SIZE);
  const tag = data.subarray(data.length - TAG_SIZE);
  const decipher = createDecipheriv("aes-256-gcm", key, nonce);
  decipher.setAuthTag(tag);
  try {
    return Buffer.concat([decipher.update(body), decipher.final()]).toString("utf8");
  } catch (e) {
    throw new Error(`decrypt error: ${e instanceof Error ? e.message : String(e)}`);
  }
}
