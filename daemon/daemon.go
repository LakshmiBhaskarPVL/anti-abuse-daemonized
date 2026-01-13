package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"anti-abuse-go/logger"
)

const (
	PidFile = "/var/run/sentinel/sentinel.pid"
	LogFile = "/var/log/sentinel/sentinel.log"
)

func StartDaemon(binaryPath, configPath, logLevel string) error {
	// Check if already running
	if isRunning() {
		return fmt.Errorf("daemon already running")
	}

	// Fork process
	cmd := exec.Command(binaryPath, "--daemon", "--config", configPath, "--log-level", logLevel)
	cmd.Stdout = nil
	cmd.Stderr = nil
	// Note: Setsid not available on Windows, use for Linux builds
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Write PID
	pidFile := getPidFile()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		cmd.Process.Kill()
		return err
	}

	logger.Log.Infof("Daemon started with PID %d", cmd.Process.Pid)
	return nil
}

func StopDaemon() error {
	pid, err := readPid()
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Kill(); err != nil {
		return err
	}

	os.Remove(getPidFile())
	logger.Log.Info("Daemon stopped")
	return nil
}

func RestartDaemon(binaryPath, configPath, logLevel string) error {
	StopDaemon()
	return StartDaemon(binaryPath, configPath, logLevel)
}

func Status() error {
	if isRunning() {
		pid, _ := readPid()
		fmt.Printf("Daemon is running with PID %d\n", pid)
	} else {
		fmt.Println("Daemon is not running")
	}
	return nil
}

func isRunning() bool {
	pid, err := readPid()
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return process.Signal(syscall.Signal(0)) == nil
}

func readPid() (int, error) {
	pidFile := getPidFile()
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

func getPidFile() string {
	return filepath.Join(".", PidFile)
}