package domain

import "fmt"

// NextBackupName returns a non-conflicting backup filename by appending
// a suffix and optionally a sequence number.
//
// Rules:
//   - First try: filename + suffix (e.g., "foo.bak")
//   - If that exists, try: filename + suffix + ".2", ".3", ...
//   - Sequence number 1 is implicit (never shown)
//
// The exists function is injected to check whether a candidate name
// is already taken, keeping this function pure and testable.
func NextBackupName(filename, suffix string, exists func(string) bool) string {
	// Try without sequence number (equivalent to .1)
	candidate := filename + suffix
	if !exists(candidate) {
		return candidate
	}

	// Try with sequence numbers starting from 2
	for i := 2; i < 100000; i++ {
		candidate = fmt.Sprintf("%s%s.%d", filename, suffix, i)
		if !exists(candidate) {
			return candidate
		}
	}

	// Fallback (should never reach here in practice)
	return candidate
}
