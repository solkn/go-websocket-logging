package RTLogger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func InitRTLogger(logFile string, ctr int) (zerolog.Logger, error) {

	fileWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return zerolog.Logger{}, err
	}

	multiWriter := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		io.Writer(fileWriter),
	)

	logger := zerolog.New(multiWriter).With().Timestamp().Logger()

	logger.Info().Str("message", fmt.Sprintf("This message %d is written to the log file", ctr)).Msg("")

	return logger, nil
}
