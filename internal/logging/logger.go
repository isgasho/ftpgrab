package logging

import (
	"io"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/model"
	"github.com/ilya1st/rotatewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Configure configures logger
func Configure(cli *model.Cli, location *time.Location) {
	var err error
	var w io.Writer

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(location)
	}

	if !cli.LogJSON {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC1123,
		}
	} else {
		w = os.Stdout
	}

	if cli.LogFile != "" {
		logFile := path.Clean(cli.LogFile)
		if err := os.MkdirAll(path.Dir(logFile), os.ModePerm); err != nil {
			log.Fatal().Err(err).Msgf("Cannot create log folder")
		}
		rwriter, err := rotatewriter.NewRotateWriter(logFile, 5)
		if err != nil {
			log.Fatal().Err(err).Msgf("Cannot create log file writer")
		}
		sighupChan := make(chan os.Signal, 1)
		signal.Notify(sighupChan, syscall.SIGHUP)
		go func() {
			for {
				_, ok := <-sighupChan
				if !ok {
					return
				}
				rwriter.Rotate(nil)
			}
		}()
		w = zerolog.MultiLevelWriter(w, rwriter)
	}

	log.Logger = zerolog.New(w).With().Timestamp().Logger()

	logLevel, err := zerolog.ParseLevel(cli.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unknown log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
	}
}
