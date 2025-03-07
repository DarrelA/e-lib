package logger

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	errMsgCreateLogFileError = "failed to create log file"
)

type ZeroLogger struct{}

func NewZeroLogger(logFile *os.File) *ZeroLogger {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	// Configure logger to write to both file and console
	log.Logger = zerolog.
		New(zerolog.
			MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr}, logFile)).
		With().
		Caller().
		Timestamp().
		Logger()

	return &ZeroLogger{}
}

func CreateAppLog(logFilePath string) *os.File {
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Error().Err(err).Str("path", logFilePath).Msg(errMsgCreateLogFileError)
		return nil // return nil if creation fails to avoid panics later.
	}

	err = os.Chmod(logFilePath, 0644)
	if err != nil {
		log.Error().Err(err).Str("path", logFilePath).Msg("Failed to set log file permissions")
	}

	return logFile
}
