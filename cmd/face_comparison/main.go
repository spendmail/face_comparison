package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	internalApp "github.com/spendmail/face_comparison/internal/app"
	awsClient "github.com/spendmail/face_comparison/internal/aws"
	internalConfig "github.com/spendmail/face_comparison/internal/config"
	internalLogger "github.com/spendmail/face_comparison/internal/logger"
	internalServer "github.com/spendmail/face_comparison/internal/server/http"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/etc/face_comparison/face_comparison.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := internalConfig.New(configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := internalLogger.New(config)
	if err != nil {
		log.Fatal(err)
	}

	recognitionClient, err := awsClient.NewRecognitionClient(config, logger)
	if err != nil {
		log.Fatal(err)
	}

	app, err := internalApp.New(logger, config, recognitionClient)
	if err != nil {
		log.Fatal(err)
	}

	server := internalServer.New(config, logger, app)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Locking until OS signal is sent or context cancel func is called.
		<-ctx.Done()

		// Stopping http server.
		stopHTTPCtx, stopHTTPCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer stopHTTPCancel()
		if err := server.Stop(stopHTTPCtx); err != nil {
			logger.Error(err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Info("starting http server...")

		// Locking over here until server is stopped.
		if err := server.Start(); err != nil {
			logger.Error(err.Error())
			cancel()
		}
	}()

	wg.Wait()
}
