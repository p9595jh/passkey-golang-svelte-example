package logger

import (
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		panic(err)
	}
	time.Local = loc

	zerolog.TimeFieldFormat = time.RFC3339
}

func Debug() *zerolog.Event { return logger.Debug() }
func Info() *zerolog.Event  { return logger.Info() }
func Warn() *zerolog.Event  { return logger.Warn() }
func Error() *zerolog.Event { return logger.Error() }
