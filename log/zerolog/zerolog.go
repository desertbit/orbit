package zerolog

import (
	"log"

	zlog "github.com/rs/zerolog/log"
)

// Logger returns a log.Logger for orbit using zerolog.
func Logger() *log.Logger {
	return log.New(zlog.With().Logger(), "orbit: ", 0)
}
