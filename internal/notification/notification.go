package notification

import "github.com/gen2brain/beeep"

// Send sends a system notification using the beeep library.
func Send(title, message string) error {
	return beeep.Notify(title, message, "")
}