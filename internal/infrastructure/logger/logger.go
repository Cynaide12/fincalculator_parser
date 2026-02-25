package logger

import (
	"fincalparser/internal/config"
	slogpretty "fincalparser/pkg/logger/handlers"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env, logDir string) (*slog.Logger, func(), func(), error) {
	var writer io.Writer = os.Stdout
	var cleanup func() = func() {}
	var rotate func() = func() {}

	if env == envProd && logDir != "" {
		lj := &lumberjack.Logger{
			Filename:   logDir,
			MaxSize:    5,     // Макс. размер файла (МБ) перед ротацией
			MaxBackups: 10,    // Сколько старых логов хранить
			MaxAge:     10,    // Сколько дней хранить логи
			Compress:   false, // Сжатие старых логов (gzip)
			LocalTime:  true,  // Использовать локальное время вместо UTC
		}
		writer = lj
		cleanup = func() { _ = lj.Close() }
		rotate = func() { lj.Rotate() }
	}

	var handler slog.Handler
	switch env {
	case envLocal, envDev:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}
		handler = opts.NewPrettyHandler(writer)
	case envProd:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler), cleanup, rotate, nil
}

func New(log *slog.Logger, cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()

			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
