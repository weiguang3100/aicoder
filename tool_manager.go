package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type ToolStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Path      string `json:"path"`
}

type ToolManager struct {
	app *App
}

func NewToolManager(app *App) *ToolManager {
	return &ToolManager{app: app}
}

func (tm *ToolManager) GetToolStatus(name string) ToolStatus {
	status := ToolStatus{Name: name}
	
	binaryName := name
	// Check for specific binary names if different from tool name
	// currently claude -> claude, codex -> codex, gemini -> gemini

	path, err := exec.LookPath(binaryName)
	if err != nil {
		return status
	}

	status.Installed = true
	status.Path = path
	
	version, err := tm.getToolVersion(binaryName, path)
	if err == nil {
		status.Version = version
	}

	return status
}

func (tm *ToolManager) getToolVersion(name, path string) (string, error) {
	var cmd *exec.Cmd
	// Use --version for all tools
	cmd = exec.Command(path, "--version")

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	output := strings.TrimSpace(string(out))
	// Parse version based on tool output format
	if strings.Contains(name, "claude") {
		// claude-code/0.2.29 darwin-arm64 node-v22.12.0
		parts := strings.Split(output, " ")
		if len(parts) > 0 {
			verParts := strings.Split(parts[0], "/")
			if len(verParts) == 2 {
				return verParts[1], nil
			}
		}
	}

	return output, nil
}

func (tm *ToolManager) InstallTool(name string) error {
	npmPath := tm.getNpmPath()
	if npmPath == "" {
		return fmt.Errorf("npm not found. Please ensure Node.js is installed.")
	}

	var cmd *exec.Cmd
	switch name {
	case "claude":
		cmd = exec.Command(npmPath, "install", "-g", "@anthropic-ai/claude-code")
	case "gemini":
		cmd = exec.Command(npmPath, "install", "-g", "@google/gemini-cli")
	case "codex":
		cmd = exec.Command(npmPath, "install", "-g", "@openai/codex")
	default:
		return fmt.Errorf("unknown tool: %s", name)
	}

	// Set environment to include local node bin for the installation process
	home, _ := os.UserHomeDir()
	localBinDir := filepath.Join(home, ".cceasy", "node", "bin")
	if runtime.GOOS == "windows" {
		localBinDir = filepath.Join(home, ".cceasy", "node")
	}

	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			env[i] = fmt.Sprintf("PATH=%s%c%s", localBinDir, os.PathListSeparator, e[5:])
			pathFound = true
			break
		}
	}
	if !pathFound {
		env = append(env, "PATH="+localBinDir)
	}
	cmd.Env = env

	// For Windows, handle .cmd extension and shell execution
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(strings.ToLower(npmPath), ".cmd") && !strings.HasSuffix(strings.ToLower(npmPath), ".exe") {
			cmd.Args = append([]string{"/c", npmPath}, cmd.Args[1:]...)
			cmd.Path = "cmd"
		}
	}

	tm.app.log(fmt.Sprintf("Running installation: %s %s", npmPath, strings.Join(cmd.Args[1:], " ")))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %v\nOutput: %s", name, err, string(out))
	}
	return nil
}

func (tm *ToolManager) getNpmPath() string {
	// 1. Check local node environment first
	home, _ := os.UserHomeDir()
	var localNpm string
	if runtime.GOOS == "windows" {
		localNpm = filepath.Join(home, ".cceasy", "node", "npm.cmd")
	} else {
		localNpm = filepath.Join(home, ".cceasy", "node", "bin", "npm")
	}

	if _, err := os.Stat(localNpm); err == nil {
		return localNpm
	}

	// 2. Fallback to system npm
	path, err := exec.LookPath("npm")
	if err == nil {
		return path
	}

	return ""
}

func (a *App) InstallTool(name string) error {
	tm := NewToolManager(a)
	return tm.InstallTool(name)
}

func (a *App) CheckToolsStatus() []ToolStatus {
	tm := NewToolManager(a)
	tools := []string{"claude", "gemini", "codex"}
	statuses := make([]ToolStatus, len(tools))
	for i, name := range tools {
		statuses[i] = tm.GetToolStatus(name)
	}
	return statuses
}
