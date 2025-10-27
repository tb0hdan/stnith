//go:build unix

package permissions

import "os"

func IsAdmin() bool {
	return os.Geteuid() == 0
}
