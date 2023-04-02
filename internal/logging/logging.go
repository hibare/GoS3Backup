package logging

import (
	log "github.com/sirupsen/logrus"
)

func SetupLogger() {
	// Set the global logger instance
	log.SetLevel(log.InfoLevel)
}
