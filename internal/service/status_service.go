package service

import (
	"github.com/jame-developer/aeontrac/internal/appcore"
)

// IsTracking returns true if a time tracking is currently running.
func IsTracking() bool {
	_, vault, _, err := appcore.LoadApp()
	if err != nil {
		// Consider how to handle errors. For now, assume not tracking on error.
		return false
	}
	return vault.CurrentRunningUnit != nil
}
