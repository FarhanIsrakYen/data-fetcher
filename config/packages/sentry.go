package packages

import (
	"os"
	"github.com/getsentry/sentry-go"
	"log"
)

func SentryInit() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("APP_ENV"),
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Debug:            true,
		AttachStacktrace: true,
		IgnoreErrors: []string{"fs.PathError", "errors.errorString"},
	})

	if err != nil {
		log.Fatalf("sentry init failed: %s", err)
	}
}
