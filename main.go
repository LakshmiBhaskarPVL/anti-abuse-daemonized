package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"anti-abuse-go/banner"
	"anti-abuse-go/config"
	"anti-abuse-go/daemon"
	"anti-abuse-go/logger"
	"anti-abuse-go/plugins"
	"anti-abuse-go/scanner"
	"anti-abuse-go/watcher"
)

var (
	configPath = flag.String("config", config.GetConfigPath(), "Path to config file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	daemonMode = flag.Bool("daemon", false, "Run as daemon")
	action     = flag.String("action", "", "Action: start, stop, restart, status")
)

func main() {
	flag.Parse()

	logger.SetLogLevel(*logLevel)

	// Print banner on startup unless daemonized
	if !*daemonMode {
		banner.PrintBanner()
	}

	if *daemonMode {
		runDaemon()
		return
	}

	switch *action {
	case "start":
		binary, err := os.Executable()
		if err != nil {
			logger.Log.WithError(err).Fatal("Failed to get executable path")
		}
		if err := daemon.StartDaemon(binary, *configPath, *logLevel); err != nil {
			logger.Log.Fatal(err)
		}
	case "stop":
		if err := daemon.StopDaemon(); err != nil {
			logger.Log.Fatal(err)
		}
	case "restart":
		binary, err := os.Executable()
		if err != nil {
			logger.Log.WithError(err).Fatal("Failed to get executable path")
		}
		if err := daemon.RestartDaemon(binary, *configPath, *logLevel); err != nil {
			logger.Log.Fatal(err)
		}
	case "status":
		if err := daemon.Status(); err != nil {
			logger.Log.Fatal(err)
		}
	default:
		runForeground()
	}
}

func runForeground() {
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to load config")
	}

	logger.Log.Infof("Starting %s v%s by %s", config.AppName, config.GetVersion(), config.Company)

	// Initialize plugins
	if err := plugins.InitPlugins(cfg); err != nil {
		logger.Log.WithError(err).Fatal("Failed to initialize plugins")
	}

	// Initialize scanner
	scan, err := scanner.NewScanner(cfg.Detection.SignaturePath)
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to initialize scanner")
	}

	// Initialize watcher
	watch, err := watcher.NewWatcher(cfg, scan)
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to initialize watcher")
	}

	// Print system tuning info
	workers, buffer, cpuCount, ramGB := watch.GetTuningInfo()
	banner.PrintSystemInfo(workers, buffer, cpuCount, ramGB)

	// Start watcher
	if err := watch.Start(); err != nil {
		logger.Log.WithError(err).Fatal("Failed to start watcher")
	}

	// Wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Log.Info("Anti-Abuse is running. Press Ctrl+C to stop.")

	<-sigChan
	logger.Log.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = shutdownCtx

	// Stop watcher
	watch.Stop()

	logger.Log.Info("Shutdown complete")
}

func runDaemon() {
	// Daemon mode - redirect logs to file
	logFile := "/var/log/sentinel/sentinel.log"
	if err := logger.SetLogFile(logFile); err != nil {
		logger.Log.WithError(err).Fatal("Failed to set log file")
	}

	runForeground()
}
