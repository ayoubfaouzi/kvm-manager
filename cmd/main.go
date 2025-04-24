package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ayoubfaouzi/kvm-manager/internal/config"
	"github.com/ayoubfaouzi/kvm-manager/internal/entity"
	"github.com/ayoubfaouzi/kvm-manager/internal/queue"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"

	"github.com/ayoubfaouzi/kvm-manager/internal/server"
	"github.com/ayoubfaouzi/kvm-manager/internal/vmmgr"

	"github.com/ayoubfaouzi/kvm-manager/pkg/log"
)

// Version indicates the current version of the application.
var Version = "0.0.1"

var flagConfig = flag.String("config", "./../configs/", "path to the config file")

func main() {

	flag.Parse()

	// Create root logger tagged with server version.
	logger := log.New().With(context.TODO(), "version", Version)

	if err := run(logger); err != nil {
		logger.Errorf("failed to run the server: %s", err)
		os.Exit(-1)
	}
}

// run was explicitly created to allow main() to receive an error when server
// creation fails.
func run(logger log.Logger) error {

	// Load application configuration.
	cfg, err := config.Load(*flagConfig)
	if err != nil {
		return err
	}
	logger.Info("successfully loaded config")

	// Create a translator for validation error messages.
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New()
	err = en_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return err
	}

	// Create a producer to write messages to stream processing framework.
	producer, err := queue.New(cfg.Broker.Address, cfg.Broker.Topic)
	if err != nil {
		return err
	}

	// Connect to the VM Manager.
	vmManager, err := vmmgr.New(logger, entity.NodeInstance{
		LibVirtURI:      cfg.VMMgr.URI,
		LibVirtImageDir: cfg.VMMgr.ImageDir})
	if err != nil {
		return err
	}

	hs := &http.Server{
		Addr:    cfg.Address,
		Handler: server.BuildHandler(logger, cfg, Version, trans, producer, vmManager),
	}

	// Start server.
	go func() {
		logger.Infof("server is running at %s", cfg.Address)
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err)
			os.Exit(-1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a
	// timeout of 10 seconds. Use a buffered channel to avoid missing
	// signals as recommended for signal.Notify.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := hs.Shutdown(ctx); err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	return nil
}
