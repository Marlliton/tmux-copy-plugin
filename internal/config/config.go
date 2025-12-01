package config

import "fmt"

type NotificationStyle string

const (
	StylePreview NotificationStyle = "preview"
	StyleMessage NotificationStyle = "msg"
	StyleSystem  NotificationStyle = "system"
	StyleNone    NotificationStyle = "none"
)

type Config struct {
	NotificationStyle NotificationStyle
}

func (c *Config) Validate() error {
	switch c.NotificationStyle {
	case StyleMessage, StylePreview, StyleSystem, StyleNone:
		return nil
	default:
		return fmt.Errorf("invalid notification style: '%s'. Alloewd values: preview, msg, system, none", c.NotificationStyle)
	}
}
