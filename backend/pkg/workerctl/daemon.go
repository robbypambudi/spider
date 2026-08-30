package workerctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// StateDir returns (creating if needed) the directory SPIDER uses to track
// locally-started background workers: PID files and their log output.
func StateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	dir := filepath.Join(home, ".spider", "worker")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create state dir: %w", err)
	}
	return dir, nil
}

func pidFilePath(workerID string) (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, workerID+".pid"), nil
}

// DefaultLogFilePath is where a detached worker's stdout/stderr is captured
// unless the caller supplies an explicit --log-file.
func DefaultLogFilePath(workerID string) (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, workerID+".log"), nil
}

func WritePIDFile(workerID string, pid int) error {
	path, err := pidFilePath(workerID)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o600)
}

func ReadPIDFile(workerID string) (int, error) {
	path, err := pidFilePath(workerID)
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func RemovePIDFile(workerID string) error {
	path, err := pidFilePath(workerID)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsRunning reports whether pid refers to a live process, cross-platform.
func IsRunning(pid int) bool {
	return isProcessRunning(pid)
}

// Detach re-launches the current executable with args as a background,
// detached child process (own process group, no console window on Windows),
// redirects its stdout/stderr to logPath, and records its PID under
// workerID so `spider worker stop` / `spider worker ps` can find it later.
func Detach(workerID string, args []string, logPath string) (pid int, err error) {
	if logPath == "" {
		logPath, err = DefaultLogFilePath(workerID)
		if err != nil {
			return 0, err
		}
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = sysProcAttrDetached()

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start background worker: %w", err)
	}
	if err := WritePIDFile(workerID, cmd.Process.Pid); err != nil {
		_ = cmd.Process.Kill()
		return 0, fmt.Errorf("write pid file: %w", err)
	}
	// Detach: don't hold onto the child, the parent CLI process exits right after this.
	_ = cmd.Process.Release()
	return cmd.Process.Pid, nil
}

// Stop kills a locally-started background worker and clears its PID file.
func Stop(workerID string) error {
	pid, err := ReadPIDFile(workerID)
	if err != nil {
		return fmt.Errorf("no background worker found for id %q", workerID)
	}
	proc, err := os.FindProcess(pid)
	if err == nil {
		_ = proc.Kill()
	}
	return RemovePIDFile(workerID)
}

// LocalWorker describes a worker previously started via `spider worker join`.
type LocalWorker struct {
	WorkerID string
	PID      int
	Running  bool
	LogFile  string
}

// ListLocal enumerates workers this machine has started (background or
// foreground) via the CLI, based on recorded PID files.
func ListLocal() ([]LocalWorker, error) {
	dir, err := StateDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []LocalWorker
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pid") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".pid")
		pid, err := ReadPIDFile(id)
		if err != nil {
			continue
		}
		logPath, _ := DefaultLogFilePath(id)
		out = append(out, LocalWorker{WorkerID: id, PID: pid, Running: IsRunning(pid), LogFile: logPath})
	}
	return out, nil
}
