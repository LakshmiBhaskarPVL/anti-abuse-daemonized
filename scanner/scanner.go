package scanner

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"anti-abuse-go/logger"
	"github.com/hillu/go-yara/v4"
	"github.com/nwaples/rardecode"
)

// Match represents a YARA match
type Match struct {
	Rule string
	Tags string
}

type MatchRules []Match

type Scanner struct {
	rules []*yara.Rules
	mu    sync.RWMutex
}

func NewScanner(signaturePath string) (*Scanner, error) {
	scanner := &Scanner{
		rules: make([]*yara.Rules, 0),
	}
	if err := scanner.loadRules(signaturePath); err != nil {
		return nil, err
	}
	return scanner, nil
}

func (s *Scanner) loadRules(signaturePath string) error {
	logger.Log.Infof("Loading YARA rules from %s", signaturePath)

	// Check if path is a file or directory
	fileInfo, err := os.Stat(signaturePath)
	if err != nil {
		logger.Log.Warnf("Signature path not found: %s - no YARA rules will be applied", signaturePath)
		// Return empty scanner instead of error - allow first-time startup
		s.rules = make([]*yara.Rules, 0)
		return nil
	}

	var filesToCompile []string

	if fileInfo.IsDir() {
		// Load all .yar and .yara files from directory
		files, err := os.ReadDir(signaturePath)
		if err != nil {
			logger.Log.Warnf("Failed to read signature directory %s: %v - no YARA rules will be applied", signaturePath, err)
			s.rules = make([]*yara.Rules, 0)
			return nil
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filename := file.Name()
			if !(strings.HasSuffix(filename, ".yar") || strings.HasSuffix(filename, ".yara")) {
				continue
			}

			rulePath := filepath.Join(signaturePath, filename)
			filesToCompile = append(filesToCompile, rulePath)
		}

		if len(filesToCompile) == 0 {
			logger.Log.Warnf("No YARA rules found in directory %s - scanner will not detect anything", signaturePath)
			s.rules = make([]*yara.Rules, 0)
			return nil
		}
	} else {
		// Single file
		if !(strings.HasSuffix(signaturePath, ".yar") || strings.HasSuffix(signaturePath, ".yara")) {
			return fmt.Errorf("file must have .yar or .yara extension: %s", signaturePath)
		}
		filesToCompile = append(filesToCompile, signaturePath)
	}

	// Compile all rules together
	compiler, err := yara.NewCompiler()
	if err != nil {
		return fmt.Errorf("failed to create YARA compiler: %w", err)
	}

	var compiledCount int
	for _, rulePath := range filesToCompile {
		file, err := os.Open(rulePath)
		if err != nil {
			logger.Log.Warnf("Failed to open YARA file %s: %v", rulePath, err)
			continue
		}

		err = compiler.AddFile(file, rulePath)
		file.Close()
		if err != nil {
			logger.Log.Warnf("Failed to compile YARA rules from %s: %v", rulePath, err)
			continue
		}

		logger.Log.Debugf("Added YARA rules from %s", rulePath)
		compiledCount++
	}

	if compiledCount == 0 {
		logger.Log.Warnf("Failed to compile any YARA rules from %d files - scanner will not detect anything", len(filesToCompile))
		s.rules = make([]*yara.Rules, 0)
		return nil
	}

	rules, err := compiler.GetRules()
	if err != nil {
		return fmt.Errorf("failed to get compiled rules: %w", err)
	}

	s.mu.Lock()
	s.rules = append(s.rules, rules)
	s.mu.Unlock()

	logger.Log.Infof("YARA rules loaded successfully (%d files)", len(filesToCompile))

	return nil
}

func (s *Scanner) Scan(data []byte, filePath string) (MatchRules, error) {
	s.mu.RLock()
	rulesList := s.rules
	s.mu.RUnlock()

	if len(rulesList) == 0 {
		return nil, fmt.Errorf("scanner not initialized - no rules loaded")
	}

	if isArchiveFile(filePath) {
		if isJarFile(filePath) {
			return s.scanJar(data)
		} else if isRarFile(filePath) {
			return s.scanRar(data)
		}
	}

	// Collect matches from all rule files
	var allMatches MatchRules

	for _, rules := range rulesList {
		var matches yara.MatchRules

		// Scan the data with timeout
		err := rules.ScanMem(data, 0, 30*time.Second, &matches)
		if err != nil {
			logger.Log.Warnf("Scan failed with ruleset: %v", err)
			continue
		}

		// Convert from yara.MatchRules to our Match type
		for _, matchRule := range matches {
			allMatches = append(allMatches, Match{
				Rule: matchRule.Rule,
				Tags: strings.Join(matchRule.Tags, ","),
			})
		}
	}

	return allMatches, nil
}

func isArchiveFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".jar" || ext == ".zip" || ext == ".rar"
}

func isJarFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".jar" || ext == ".zip"
}

func isRarFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".rar"
}

func (s *Scanner) scanJar(data []byte) (MatchRules, error) {
	reader, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open JAR: %w", err)
	}

	var allMatches MatchRules
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if file.UncompressedSize64 > 10*1024*1024 { // 10MB limit per file in JAR
			logger.Log.Debugf("Skipping %s in JAR (size > 10MB)", file.Name)
			continue
		}

		rc, err := file.Open()
		if err != nil {
			logger.Log.Warnf("Failed to open %s in JAR: %v", file.Name, err)
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			logger.Log.Warnf("Failed to read %s in JAR: %v", file.Name, err)
			continue
		}

		// Scan the extracted file content
		matches, err := s.Scan(content, file.Name)
		if err != nil {
			logger.Log.Warnf("Error scanning %s in JAR: %v", file.Name, err)
			continue
		}

		allMatches = append(allMatches, matches...)
	}

	return allMatches, nil
}

func (s *Scanner) scanRar(data []byte) (MatchRules, error) {
	reader, err := rardecode.NewReader(strings.NewReader(string(data)), "")
	if err != nil {
		return nil, fmt.Errorf("failed to open RAR: %w", err)
	}

	var allMatches MatchRules
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Log.Warnf("Error reading RAR entry: %v", err)
			continue
		}

		if header.IsDir {
			continue
		}
		if header.UnpackedSize > 10*1024*1024 { // 10MB limit per file in RAR
			logger.Log.Debugf("Skipping %s in RAR (size > 10MB)", header.Name)
			continue
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			logger.Log.Warnf("Failed to read %s in RAR: %v", header.Name, err)
			continue
		}

		// Scan the extracted file content
		matches, err := s.Scan(content, header.Name)
		if err != nil {
			logger.Log.Warnf("Error scanning %s in RAR: %v", header.Name, err)
			continue
		}

		allMatches = append(allMatches, matches...)
	}

	return allMatches, nil
}

func (s *Scanner) ReloadRules(signaturePath string) error {
	return s.loadRules(signaturePath)
}
