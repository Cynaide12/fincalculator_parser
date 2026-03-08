package main

import (
	"fincalparser/internal/config"
	"fincalparser/internal/infrastructure/logger"
	"fincalparser/internal/infrastructure/parser"
	response "fincalparser/pkg/api"
	"fincalparser/pkg/logger/sl"
	"fmt"
	"log/slog"

	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.MustLoad()

	log, _, rotate, err := logger.SetupLogger(cfg.Env, cfg.LogFilePath)
	if err != nil {
		panic(err)
	}

	log.Info("starting fincalparser", slog.String("env", cfg.Env))

	log.Debug("debug messages are enabled")

	setupLogRotation(rotate)

	log.Info("logs rotation are enabled")

	parser, err := parser.New(parser.ParserResource{
		Type: &parser.Bskrt,
		Year: 2026,
	}, log, cfg.DataDir)
	if err != nil {
		panic(fmt.Sprintf("unable to init parser: %s", err))
	}

	go parser.Start()

	setupRouter(cfg, log, *parser)
}

func setupLogRotation(rotate func()) {
	//запускаем ротацию логов каждые сутки
	c := cron.New(cron.WithLocation(time.Local))

	c.AddFunc("@every 1d", func() {
		rotate()
	})

	c.Start()
}

func setupRouter(cfg *config.Config, log *slog.Logger, p parser.Parser) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(logger.New(log, cfg))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true,
	}))


	r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
		data, err := p.LoadData()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("не удалось получить данные с жсон файла")
			render.JSON(w,r, response.Error("internal error"))
			return
		}
		render.JSON(w,r, data)
	})

	log.Info("starting server", slog.String("address", srv.Addr))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))

		os.Exit(1)
	}

	log.Error("server stopped")

}
