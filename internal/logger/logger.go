package logger

import (
	"os"
	"time"

	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func init() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05", // Direct time format here
		NoColor:    false,
		// Indent JSON for better readability
		FormatMessage: func(i interface{}) string {
			return "  " + i.(string) // Add two spaces before message
		},
		// Format level with consistent padding
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i)) // Pad level to 6 chars
		},
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	Log = zerolog.New(output).With().Timestamp().Caller().Logger()
}
