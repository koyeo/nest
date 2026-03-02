package execer

import "os"

func HomeDir() (string, error) {
	return os.UserHomeDir()
}
