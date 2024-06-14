package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func InitLogger(logFile string) (zerolog.Logger, error) {

	fileWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return zerolog.Logger{}, err
	}

	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		io.Writer(fileWriter),
	)

	logger := zerolog.New(multiWriter).With().Timestamp().Logger()

	logger.Info().Str("message", "This message is written to the log file").Msg("")

	return logger, nil
}
