package logging

import (
	"log/syslog"

	"github.com/hibare/GoS3Backup/internal/constants"
	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
)

func SetupLogger() {
	// Create a new SyslogHook with the appropriate tag
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, constants.ProgramIdentifier)
	if err != nil {
		log.Fatalf("Unable to create syslog hook: %s", err)
	}

	writer := hook.Writer

	// Set the global logger instance
	log.SetOutput(writer)
	log.SetLevel(log.InfoLevel)
}
