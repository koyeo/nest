import { decrypt, encrypt } from "./crypto.js";

/**
 * Key baked into the tool for credentials embedded in ship.yaml (`enc:` values).
 * This is obfuscation, not secrecy: anyone with the yaml + this tool can recover the keys.
 */
const BUILTIN_KEY = "a296aWxsYS1zaGlwLWlubGluZS1zdG9yYWdlLWtleTE=";
export const INLINE_PREFIX = "enc:";

export function encryptInline(plaintext: string): string {
  return `${INLINE_PREFIX}${encrypt(BUILTIN_KEY, plaintext)}`;
}

export function isInlineCiphertext(value: string): boolean {
  return value.startsWith(INLINE_PREFIX);
}

export function decryptInline(value: string): string {
  if (!isInlineCiphertext(value)) {
    throw new Error(`expected an '${INLINE_PREFIX}' value produced by 'ship storage encrypt'`);
  }
  return decrypt(BUILTIN_KEY, value.slice(INLINE_PREFIX.length));
}
