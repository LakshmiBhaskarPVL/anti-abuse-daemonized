package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"anti-abuse-go/config"
	"anti-abuse-go/integrations"
	"anti-abuse-go/logger"
	"anti-abuse-go/plugins"
	"anti-abuse-go/scanner"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher    *fsnotify.Watcher
	scanner    *scanner.Scanner
	config     *config.Config
	workChan   chan FileEvent
	workerPool int
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

type FileEvent struct {
	Path    string
	Op      fsnotify.Op
	Content []byte
}

func NewWatcher(cfg *config.Config, scan *scanner.Scanner) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Auto-tune based on system resources
	workerPool, bufferSize := autoTuneResources()

	watch := &Watcher{
		watcher:    w,
		scanner:    scan,
		config:     cfg,
		workChan:   make(chan FileEvent, bufferSize),
		workerPool: workerPool,
		ctx:        ctx,
		cancel:     cancel,
	}

	return watch, nil
}

func autoTuneResources() (int, int) {
	numCPU := runtime.NumCPU()
	memGB := getSystemMemoryGB()

	// Worker pool: based on CPU, capped
	workerPool := numCPU * 2
	if workerPool > 32 {
		workerPool = 32
	}
	if workerPool < 2 {
		workerPool = 2
	}

	// Buffer size: based on memory
	bufferSize := 2048
	if memGB >= 8 {
		bufferSize = 4096
	}
	if memGB >= 16 {
		bufferSize = 8192
	}

	logger.Log.Infof("Auto-tuned: %d workers, %d buffer size (CPU: %d, RAM: %dGB)", workerPool, bufferSize, numCPU, memGB)
	return workerPool, bufferSize
}

func getSystemMemoryGB() int {
	// Simple memory detection (in GB)
	// In a real implementation, read /proc/meminfo or use syscall
	// For now, assume based on runtime
	return 8 // Placeholder; in production, detect properly
}

func (w *Watcher) Start() error {
	// Add watch paths
	for _, path := range w.config.Detection.WatchdogPath {
		if err := w.addWatchRecursive(path); err != nil {
			logger.Log.WithError(err).Warnf("Failed to watch path: %s", path)
		}
	}

	// Start workers
	for i := 0; i < w.workerPool; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	// Start event loop
	go w.eventLoop()

	logger.Log.Infof("Watcher started with %d workers", w.workerPool)
	return nil
}

func (w *Watcher) Stop() {
	w.cancel()
	w.watcher.Close()
	close(w.workChan)
	w.wg.Wait()
	logger.Log.Info("Watcher stopped")
}

func (w *Watcher) addWatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && !w.shouldIgnore(path) {
			return w.watcher.Add(path)
		}
		return nil
	})
}

func (w *Watcher) shouldIgnore(path string) bool {
	for _, ignore := range w.config.Detection.WatchdogIgnorePath {
		if matched, _ := filepath.Match(ignore, path); matched {
			return true
		}
	}
	return false
}

func (w *Watcher) eventLoop() {
	ticker := time.NewTicker(1 * time.Second) // Batch events
	defer ticker.Stop()

	var events []fsnotify.Event

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if w.shouldProcessEvent(event) {
				events = append(events, event)
			}
		case <-ticker.C:
			w.processBatch(events)
			events = nil
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			logger.Log.WithError(err).Error("Watcher error")
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Watcher) shouldProcessEvent(event fsnotify.Event) bool {
	if w.shouldIgnore(event.Name) {
		return false
	}
	for _, ignore := range w.config.Detection.WatchdogIgnoreFile {
		if matched, _ := filepath.Match(ignore, filepath.Base(event.Name)); matched {
			return false
		}
	}
	return event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0
}

func (w *Watcher) processBatch(events []fsnotify.Event) {
	for _, event := range events {
		content, err := w.readFileContent(event.Name)
		if err != nil {
			logger.Log.WithError(err).Debugf("Failed to read file: %s", event.Name)
			continue
		}

		select {
		case w.workChan <- FileEvent{Path: event.Name, Op: event.Op, Content: content}:
		case <-w.ctx.Done():
			return
		default:
			logger.Log.Warn("Work channel full, dropping event")
		}
	}
}

func (w *Watcher) readFileContent(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	maxSize := int64(100 * 1024 * 1024) // Default 100MB
	if w.config.Detection.MaxFileSizeMB > 0 {
		maxSize = int64(w.config.Detection.MaxFileSizeMB) * 1024 * 1024
	}

	if stat.Size() > maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", stat.Size(), maxSize)
	}

	buf := make([]byte, stat.Size())
	_, err = file.Read(buf)
	return buf, err
}

func (w *Watcher) worker(id int) {
	defer w.wg.Done()
	logger.Log.Debugf("Worker %d started", id)

	for {
		select {
		case event, ok := <-w.workChan:
			if !ok {
				return
			}
			w.processFile(event)
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Watcher) processFile(event FileEvent) {
	matches, err := w.scanner.Scan(event.Content, event.Path)
	if err != nil {
		logger.Log.WithError(err).Debugf("Scan failed for %s", event.Path)
		return
	}

	if len(matches) > 0 {
		logger.Log.WithField("matches", len(matches)).Infof("Flagged: %s", event.Path)

		// Trigger AI analysis if enabled
		var aiAnalysis string
		if w.config.Integration.AI.Enabled {
			analysis, err := integrations.AnalyzeWithAI(w.config, string(event.Content))
			if err != nil {
				logger.Log.WithError(err).Warnf("AI analysis failed for %s", event.Path)
				aiAnalysis = "AI analysis failed"
			} else if analysis != nil {
				aiAnalysis = analysis.Content
			}
		}

		// Send Discord webhook if enabled
		if w.config.Integration.Discord.Enabled {
			fields := make([]integrations.DiscordField, 0)
			for _, match := range matches {
				fields = append(fields, integrations.DiscordField{
					Name:   match.Rule,
					Value:  match.Tags,
					Inline: true,
				})
			}
			if err := integrations.SendDiscordWebhook(w.config, w.config.MachineID, "⚠️ Abuse Detection Alert", event.Path, fields, aiAnalysis); err != nil {
				logger.Log.WithError(err).Warnf("Discord webhook failed for %s", event.Path)
			}
		}

		// Trigger plugins
		for _, plugin := range plugins.GetPlugins() {
			if err := plugin.OnDetected(event.Path, matches); err != nil {
				logger.Log.WithError(err).Warnf("Plugin %s failed for %s", plugin.Name(), event.Path)
			}
		}
	} else if w.config.Logs.FileModified || w.config.Logs.FileCreated {
		logger.Log.Debugf("Processed: %s", event.Path)
	}
}
