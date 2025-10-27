//go:build windows

package permissions

import (
	"golang.org/x/sys/windows"
)

func IsAdmin() bool {
	token, err := windows.GetCurrentProcessToken()
	if err != nil {
		return false
	}
	defer token.Close()

	isElevated, err := token.IsElevated()
	if err != nil {
		return false
	}

	return isElevated
}