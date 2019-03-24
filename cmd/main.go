package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"github.com/go-kit/kit/log"

	"github.com/gorilla/mux"
	"github.com/jozuenoon/biblia2y/messenger"
	"github.com/jozuenoon/biblia2y/pages/privacyPolicy"
	"github.com/stevenroose/gonfig"
	validator "gopkg.in/go-playground/validator.v9"
)

var config = struct {
	ServerPort string `id:"server_port" validate:"required"`

	VerifyToken     string `id:"verify_token" validate:"required"`
	PageAccessToken string `id:"page_access_token" validate:"required"`
	TLSCert         string `id:"tls_cert" validate:"required"`
	TLSKey          string `id:"tls_key" validate:"required"`
	FaceBookAPI     string `id:"facebook_api" validate:"required"`

	DatabasePath string `id:"database_path" validate:"required"`
	BooksPath    string `id:"books_path" validate:"required"`
	TextPath     string `id:"text_path" validate:"required"`
	PlanPath     string `id:"plan_path" validate:"required"`

	ConfigFile string `id:"config_file"`
}{
	ServerPort:   ":443",
	DatabasePath: "db",
}

func main() {
	// Configure logger...
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// Load and validate config.
	gonf := gonfig.Conf{
		ConfigFileVariable:  "config_file",
		FileDefaultFilename: "config/config.toml",
		FileDecoder:         gonfig.DecoderTOML,
	}
	if err := gonfig.Load(&config, gonf); err != nil {
		panicf("error loading config: %s", err)
	}
	if err := validator.New().Struct(config); err != nil {
		panicf("invalid config: %s", err)
	}

	done := make(chan struct{})

	bs, err := messenger.New(config.DatabasePath,
		config.PageAccessToken,
		config.FaceBookAPI,
		config.BooksPath,
		config.TextPath,
		config.PlanPath,
		logger,
		done)
	if err != nil {
		panicf("failed to create service %s", err)
	}

	mux := mux.NewRouter()

	mux.PathPrefix("/webhook").Handler(messenger.MakeHandler(bs, logger, config.VerifyToken))
	mux.HandleFunc("/privacyPolicy", privacyPolicy.Handler)

	h := handlers.LoggingHandler(os.Stderr, mux)

	l, err := net.Listen("tcp4", config.ServerPort)
	if err != nil {
		panicf("can't bind to port  %s", err)
	}
	logger.Log("msg", "server started", "port", config.ServerPort)
	err = http.ServeTLS(l, h, config.TLSCert, config.TLSKey)
	logger.Log("terminated", err)
}

func panicf(s string, i ...interface{}) {
	panic(fmt.Sprintf(s, i...))
}
