//go:build windows

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) platformStartup() {
	// No console to hide if built with -H windowsgui
}

func (a *App) updatePathForNode() {
	nodePath := `C:\Program Files\nodejs`
	npmPath := filepath.Join(os.Getenv("AppData"), "npm")
	home, _ := os.UserHomeDir()
	localToolPath := filepath.Join(home, ".cceasy", "tools")
	oldToolPath := filepath.Join(home, ".cceasy", "node")

	currentPath := os.Getenv("PATH")
	// Remove old path if present to prevent conflicts
	if strings.Contains(strings.ToLower(currentPath), strings.ToLower(oldToolPath)) {
		// Simple removal - split by separator, filter, join
		parts := strings.Split(currentPath, string(os.PathListSeparator))
		var newParts []string
		for _, part := range parts {
			if !strings.EqualFold(part, oldToolPath) {
				newParts = append(newParts, part)
			}
		}
		currentPath = strings.Join(newParts, string(os.PathListSeparator))
	}

	newPath := currentPath
	
	// Check and add Node.js path
	if _, err := os.Stat(nodePath); err == nil {
		if !strings.Contains(strings.ToLower(currentPath), strings.ToLower(nodePath)) {
			newPath = nodePath + string(os.PathListSeparator) + newPath
		}
	}
	
	// Check and add npm global bin path
	if _, err := os.Stat(npmPath); err == nil {
		if !strings.Contains(strings.ToLower(currentPath), strings.ToLower(npmPath)) {
			newPath = npmPath + string(os.PathListSeparator) + newPath
		}
	}

	// Check and add local tool path
	if _, err := os.Stat(localToolPath); err == nil {
		if !strings.Contains(strings.ToLower(currentPath), strings.ToLower(localToolPath)) {
			newPath = localToolPath + string(os.PathListSeparator) + newPath
		}
	}

	if newPath != currentPath {
		os.Setenv("PATH", newPath)
		a.log("Updated PATH environment variable: " + newPath)
	}
}

func (a *App) CheckEnvironment() {
	go func() {
		a.log(a.tr("Checking Node.js installation..."))

		// Check for node
		nodeCmd := exec.Command("node", "--version")
		nodeCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := nodeCmd.Run(); err != nil {
			a.log(a.tr("Node.js not found. Downloading and installing..."))
			if err := a.installNodeJS(); err != nil {
				a.log(a.tr("Failed to install Node.js: ") + err.Error())
				return
			}
			a.log(a.tr("Node.js installed successfully."))
		} else {
			a.log(a.tr("Node.js is installed."))
		}

		// Update path for the current process anyway to ensure npm is found
		a.updatePathForNode()

		// Check for Git
		a.log(a.tr("Checking Git installation..."))
		if _, err := exec.LookPath("git"); err != nil {
			// Check common locations before giving up
			gitFound := false
			if _, err := os.Stat(`C:\Program Files\Git\cmd\git.exe`); err == nil {
				gitFound = true
			}
			
			if gitFound {
				a.updatePathForGit()
				a.log(a.tr("Git found in standard location."))
			} else {
				a.log(a.tr("Git not found. Downloading and installing..."))
				if err := a.installGitBash(); err != nil {
					a.log("Failed to install Git: " + err.Error())
				} else {
					a.log(a.tr("Git installed successfully."))
					a.updatePathForGit()
				}
			}
		} else {
			a.log(a.tr("Git is installed."))
		}

		// Ensure node.exe is in local tool path for npm wrappers
		a.ensureLocalNodeBinary()

		// 5. Check and Install AI Tools
		tm := NewToolManager(a)


		// Search for npm
		npmExec, err := exec.LookPath("npm")
		if err != nil {
			npmExec, err = exec.LookPath("npm.cmd")
		}
		if npmExec == "" {
			npmExec = "npm"
		}

		tools := []string{"claude", "gemini", "codex", "opencode", "codebuddy", "qoder", "iflow"}
		
		for _, tool := range tools {
			a.log(a.tr("Checking %s...", tool))
			status := tm.GetToolStatus(tool)
			
			if !status.Installed {
				a.log(a.tr("%s not found. Attempting automatic installation...", tool))
				if err := tm.InstallTool(tool); err != nil {
					a.log(a.tr("ERROR: Failed to install %s: %v", tool, err))
					// We continue to other tools even if one fails, allowing manual intervention later
				} else {
					a.log(a.tr("%s installed successfully.", tool))
					a.updatePathForNode() // Refresh path after install
				}
			} else {
				a.log(a.tr("%s found at %s (version: %s).", tool, status.Path, status.Version))
				// Check for updates for all tools
				if tool == "codex" || tool == "opencode" || tool == "codebuddy" || tool == "qoder" || tool == "iflow" || tool == "gemini" || tool == "claude" {
					a.log(a.tr("Checking for %s updates...", tool))
					latest, err := a.getLatestNpmVersion(npmExec, tm.GetPackageName(tool))
					if err == nil && latest != "" && latest != status.Version {
						a.log(a.tr("New version available for %s: %s (current: %s). Updating...", tool, latest, status.Version))
						if err := tm.InstallTool(tool); err != nil {
							a.log(a.tr("ERROR: Failed to update %s: %v", tool, err))
						} else {
							a.log(a.tr("%s updated successfully to %s.", tool, latest))
						}
					}
				}
			}
		}

		a.log(a.tr("Environment check complete."))
		runtime.EventsEmit(a.ctx, "env-check-done")
	}()
}

func (a *App) installNodeJS() error {
	arch := os.Getenv("PROCESSOR_ARCHITECTURE")
	nodeArch := "x64"
	if arch == "ARM64" || os.Getenv("PROCESSOR_ARCHITEW6432") == "ARM64" {
		nodeArch = "arm64"
	}

	// Using a more recent version
	nodeVersion := RequiredNodeVersion
	fileName := fmt.Sprintf("node-v%s-%s.msi", nodeVersion, nodeArch)
	
	downloadURL := fmt.Sprintf("https://nodejs.org/dist/v%s/%s", nodeVersion, fileName)
	if strings.HasPrefix(strings.ToLower(a.CurrentLanguage), "zh") && nodeArch != "arm64" {
		// Use a mirror in China for faster download (only for x64 as arm64 might not be synced)
		downloadURL = fmt.Sprintf("https://mirrors.tuna.tsinghua.edu.cn/nodejs-release/v%s/%s", nodeVersion, fileName)
	}

	a.log(a.tr("Downloading Node.js %s for %s...", nodeVersion, nodeArch))

	// Pre-check if the file exists and is accessible
	client := &http.Client{Timeout: 10 * time.Second}
	headReq, _ := http.NewRequest("HEAD", downloadURL, nil)
	headReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	headResp, err := client.Do(headReq)
	if err != nil || headResp.StatusCode != http.StatusOK {
		status := "Unknown"
		if headResp != nil {
			status = headResp.Status
		}
		return fmt.Errorf(a.tr("Node.js installer is not accessible (Status: %s). Please check your internet connection or mirror availability.", status))
	}
	headResp.Body.Close()

	tempDir := os.TempDir()
	msiPath := filepath.Join(tempDir, fileName)

	if err := a.downloadFile(msiPath, downloadURL); err != nil {
		return fmt.Errorf("error downloading Node.js installer: %w", err)
	}
	defer os.Remove(msiPath)

	a.log(a.tr("Installing Node.js (this may take a moment, please grant administrator permission if prompted)..."))
	// Use /passive for basic UI or /qn for completely silent.
	// Adding ALLUSERS=1 to ensure it's in the standard path.
	cmd := exec.Command("msiexec", "/i", msiPath, "/passive", "ALLUSERS=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error installing Node.js: %s\n%s", err, string(output))
	}

	// Wait a bit for the system to finalize the installation
	time.Sleep(2 * time.Second)

	return nil
}

func (a *App) updatePathForGit() {
	// Common git paths
	gitPaths := []string{
		`C:\Program Files\Git\cmd`,
		`C:\Program Files\Git\bin`,
	}
	
	currentPath := os.Getenv("PATH")
	newPath := currentPath
	
	for _, path := range gitPaths {
		if _, err := os.Stat(path); err == nil {
			if !strings.Contains(strings.ToLower(currentPath), strings.ToLower(path)) {
				newPath = path + string(os.PathListSeparator) + newPath
			}
		}
	}

	if newPath != currentPath {
		os.Setenv("PATH", newPath)
		a.log("Updated PATH environment variable for Git.")
	}
}

func (a *App) installGitBash() error {
	gitVersion := "2.47.1"
	// git-for-windows versioning can be tricky. v2.47.1.windows.1
	fullVersion := "v2.47.1.windows.1"
	fileName := fmt.Sprintf("Git-%s-64-bit.exe", gitVersion)
	
	downloadURL := fmt.Sprintf("https://github.com/git-for-windows/git/releases/download/%s/%s", fullVersion, fileName)
	if strings.HasPrefix(strings.ToLower(a.CurrentLanguage), "zh") {
		downloadURL = fmt.Sprintf("https://npmmirror.com/mirrors/git-for-windows/%s/%s", fullVersion, fileName)
	}
	
	a.log(a.tr("Downloading Git %s...", gitVersion))

	tempDir := os.TempDir()
	exePath := filepath.Join(tempDir, fileName)

	if err := a.downloadFile(exePath, downloadURL); err != nil {
		return fmt.Errorf("error downloading Git installer: %w", err)
	}
	defer os.Remove(exePath)

	a.log(a.tr("Installing Git (this may take a moment, please grant administrator permission if prompted)..."))
	// Silent installation
	cmd := exec.Command(exePath, "/VERYSILENT", "/NORESTART", "/NOCANCEL", "/SP-", "/CLOSEAPPLICATIONS", "/RESTARTAPPLICATIONS")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error installing Git: %s\n%s", err, string(output))
	}

	// Wait a bit for the system to finalize the installation
	time.Sleep(2 * time.Second)

	return nil
}

func (a *App) downloadFile(filepath string, url string) error {
	a.log(fmt.Sprintf("Requesting URL: %s", url))
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	// Add User-Agent to avoid 403 Forbidden from some mirrors/CDNs
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Use a client with timeout for the connection phase
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error during download: %v. Please check your internet connection or firewall settings.", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s. The file might not be available on this server.", resp.Status)
	}

	size := resp.ContentLength
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	var downloaded int64
	buffer := make([]byte, 32768)
	lastReport := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			out.Write(buffer[:n])
			downloaded += int64(n)
			if size > 0 && time.Since(lastReport) > 500*time.Millisecond {
				percent := float64(downloaded) / float64(size) * 100
				a.log(a.tr("Downloading Node.js (%.1f%%): %d/%d bytes", percent, downloaded, size))
				lastReport = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("interrupted download: %v. The connection was lost during data transfer.", err)
		}
	}

	return nil
}

func (a *App) restartApp() {
	executable, err := os.Executable()
	if err != nil {
		a.log("Failed to get executable path: " + err.Error())
		return
	}

	cmdLine := fmt.Sprintf(`cmd /c start "" "%s"`, executable)
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:    cmdLine,
		HideWindow: true,
	}
	if err := cmd.Start(); err != nil {
		a.log("Failed to restart: " + err.Error())
	} else {
		runtime.Quit(a.ctx)
	}
}

func (a *App) findSh() string {
	// Try standard path lookup
	path, err := exec.LookPath("sh")
	if err == nil {
		return path
	}

	// Try common locations for Git Bash
	commonPaths := []string{
		`C:\Program Files\Git\bin\sh.exe`,
		`C:\Program Files\Git\usr\bin\sh.exe`,
		`C:\Program Files (x86)\Git\bin\sh.exe`,
		`C:\Program Files (x86)\Git\usr\bin\sh.exe`,
	}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "sh" // Fallback
}

func (a *App) platformLaunch(binaryName string, yoloMode bool, adminMode bool, pythonEnv string, projectDir string, env map[string]string, modelId string) {
	tm := NewToolManager(a)
	status := tm.GetToolStatus(binaryName)
	
	binaryPath := ""
	if status.Installed {
		binaryPath = status.Path
	}

	if binaryPath == "" {
		msg := fmt.Sprintf("Tool %s not found. Please ensure it is installed.", binaryName)
		a.log(msg)
		a.ShowMessage("Launch Error", msg)
		return
	}
	a.log("Using binary at: " + binaryPath)

	projectDir = filepath.Clean(projectDir)
	binaryPath = filepath.Clean(binaryPath)

	// Build the command with arguments
	cmdArgs := ""
	if binaryName == "codebuddy" && modelId != "" {
		cmdArgs += fmt.Sprintf(" --model %s", modelId)
	}

	if yoloMode {
		var flag string
		switch binaryName {
		case "claude":
			flag = "--dangerously-skip-permissions"
		case "gemini":
			flag = "--yolo"
		case "codex":
			flag = "--full-auto"
		case "codebuddy":
			flag = "-y"
		case "iflow":
			flag = "-y"
		case "qodercli", "qoder":
			flag = "--yolo"
		}
		if flag != "" {
			cmdArgs += " " + flag
		}
	}

	// Create a unified batch file for launching
	batchContent := "@echo off\r\n"
	batchContent += "chcp 65001 > nul\r\n" // Use UTF-8
	batchContent += fmt.Sprintf("cd /d \"%s\"\r\n", projectDir)

	// Set environment variables in the batch file
	for k, v := range env {
		batchContent += fmt.Sprintf("set %s=%s\r\n", k, v)
	}

	// Add local tools directory to PATH
	home, _ := os.UserHomeDir()
	localToolPath := filepath.Join(home, ".cceasy", "tools")
	batchContent += fmt.Sprintf("set PATH=%s;%%PATH%%\r\n", localToolPath)

	// Activate Python environment if specified
	if pythonEnv != "" && pythonEnv != "None (Default)" {
		// Find conda root directory
		condaRoot := a.getCondaRoot()
		if condaRoot != "" {
			// Initialize conda first by calling activate.bat
			activateScript := filepath.Join(condaRoot, "Scripts", "activate.bat")
			batchContent += fmt.Sprintf("echo Initializing Conda from: %s\r\n", condaRoot)
			batchContent += fmt.Sprintf("call \"%s\"\r\n", activateScript)

			// Now activate the specific environment
			batchContent += fmt.Sprintf("echo Activating Python environment: %s\r\n", pythonEnv)
			batchContent += fmt.Sprintf("call conda activate \"%s\"\r\n", pythonEnv)
			batchContent += "if errorlevel 1 (\r\n"
			batchContent += fmt.Sprintf("  echo Warning: Failed to activate conda environment '%s'. Continuing with base environment.\r\n", pythonEnv)
			batchContent += ")\r\n"
		} else {
			batchContent += "echo Warning: Conda installation not found. Cannot activate environment.\r\n"
		}
	}

	// Launch the tool
	batchContent += fmt.Sprintf("echo Launching %s...\r\n", binaryName)
	
	ext := strings.ToLower(filepath.Ext(binaryPath))
	if ext == ".ps1" {
		batchContent += fmt.Sprintf("powershell -ExecutionPolicy Bypass -File \"%s\"%s\r\n", binaryPath, cmdArgs)
	} else if ext == ".js" {
		batchContent += fmt.Sprintf("node \"%s\"%s\r\n", binaryPath, cmdArgs)
	} else if ext == "" {
		// Assume shell script (extensionless). Try to run with sh (Git Bash)
		// Find sh executable explicitly
		shPath := a.findSh()
		batchContent += fmt.Sprintf("\"%s\" \"%s\"%s\r\n", shPath, binaryPath, cmdArgs)
	} else {
		batchContent += fmt.Sprintf("\"%s\"%s\r\n", binaryPath, cmdArgs)
	}
	
	// Pause on error
	batchContent += "if errorlevel 1 (\r\n"
	batchContent += "  echo.\r\n"
	batchContent += "  echo Process exited with error code %errorlevel%.\r\n"
	batchContent += "  pause\r\n"
	batchContent += "  exit /b %errorlevel%\r\n"
	batchContent += ")\r\n"
	batchContent += "exit /b 0\r\n"

	// Create a temporary batch file
	tempBatchPath := filepath.Join(os.TempDir(), fmt.Sprintf("aicoder_launch_%d.bat", time.Now().UnixNano()))
	err := os.WriteFile(tempBatchPath, []byte(batchContent), 0644)
	if err != nil {
		a.log("Error creating batch file: " + err.Error())
		a.ShowMessage("Launch Error", "Failed to create temporary batch file")
		return
	}

	a.log(fmt.Sprintf("Created launch script: %s", tempBatchPath))

	// Clean up the batch file after a delay
	go func() {
		time.Sleep(10 * time.Second)
		os.Remove(tempBatchPath)
	}()

	if adminMode {
		// Launch with administrator privileges
		shell32 := syscall.NewLazyDLL("shell32.dll")
		shellExecute := shell32.NewProc("ShellExecuteW")

		verb := syscall.StringToUTF16Ptr("runas")
		file := syscall.StringToUTF16Ptr("cmd.exe")
		params := syscall.StringToUTF16Ptr(fmt.Sprintf("/c \"%s\"", tempBatchPath))
		dir := syscall.StringToUTF16Ptr(projectDir)

		ret, _, _ := shellExecute.Call(
			0,
			uintptr(unsafe.Pointer(verb)),
			uintptr(unsafe.Pointer(file)),
			uintptr(unsafe.Pointer(params)),
			uintptr(unsafe.Pointer(dir)),
			uintptr(syscall.SW_SHOW),
		)

		if ret <= 32 {
			a.log(fmt.Sprintf("ShellExecute failed with return value: %d", ret))
			a.ShowMessage("Launch Error", "Failed to launch with admin privileges.")
		}
	} else {
		// Normal launch
		cmdLine := fmt.Sprintf(`cmd /c start "AICoder - %s" /d "%s" cmd /k "%s"`, binaryName, projectDir, tempBatchPath)
		
		cmd := exec.Command("cmd")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CmdLine:    cmdLine,
			HideWindow: true,
		}

		if err := cmd.Start(); err != nil {
			a.log("Error launching tool: " + err.Error())
			a.ShowMessage("Launch Error", "Failed to start process: "+err.Error())
		}
	}
}

func (a *App) syncToSystemEnv(config AppConfig) {
	toolName := strings.ToLower(config.ActiveTool)
	var toolCfg ToolConfig
	var envKey, envBaseUrl string

	switch toolName {
	case "claude":
		toolCfg = config.Claude
		envKey = "ANTHROPIC_AUTH_TOKEN"
		envBaseUrl = "ANTHROPIC_BASE_URL"
	case "gemini":
		toolCfg = config.Gemini
		envKey = "GEMINI_API_KEY"
		envBaseUrl = "GOOGLE_GEMINI_BASE_URL"
	case "codex":
		toolCfg = config.Codex
		envKey = "OPENAI_API_KEY"
		envBaseUrl = "OPENAI_BASE_URL"
	default:
		return
	}

	var selectedModel *ModelConfig
	for _, m := range toolCfg.Models {
		if m.ModelName == toolCfg.CurrentModel {
			selectedModel = &m
			break
		}
	}

	if selectedModel == nil {
		return
	}

	if strings.ToLower(selectedModel.ModelName) == "original" {
		// Clear environment variables for the current process
		os.Unsetenv(envKey)
		os.Unsetenv(envBaseUrl)
		if toolName == "codex" {
			os.Unsetenv("WIRE_API")
		}

		// Clear persistent environment variables on Windows
		go func() {
			cmd1 := exec.Command("setx", envKey, "")
			cmd1.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			cmd1.Run()

			cmd2 := exec.Command("setx", envBaseUrl, "")
			cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			cmd2.Run()

			if toolName == "claude" {
				cmd3 := exec.Command("setx", "ANTHROPIC_API_KEY", "")
				cmd3.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				cmd3.Run()
			}
			if toolName == "codex" {
				cmd4 := exec.Command("setx", "WIRE_API", "")
				cmd4.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				cmd4.Run()
			}
		}()
		return
	}

	baseUrl := selectedModel.ModelUrl
	if toolName == "claude" {
		baseUrl = getBaseUrl(selectedModel)
	}

	// Set environment variables for the current process immediately
	os.Setenv(envKey, selectedModel.ApiKey)
	if baseUrl != "" {
		os.Setenv(envBaseUrl, baseUrl)
	}


	if toolName == "codex" {
		os.Setenv("WIRE_API", "responses")
	}

	// Set persistent environment variables on Windows in a goroutine
	go func() {
		cmd1 := exec.Command("setx", envKey, selectedModel.ApiKey)
		cmd1.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd1.Run()

		if baseUrl != "" {
			cmd2 := exec.Command("setx", envBaseUrl, baseUrl)
			cmd2.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			cmd2.Run()
		}
		if toolName == "claude" {
			// Explicitly clear API_KEY in system to avoid conflict with AUTH_TOKEN
			cmd3 := exec.Command("setx", "ANTHROPIC_API_KEY", "")
			cmd3.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			cmd3.Run()
		}
		if toolName == "codex" {
			cmd4 := exec.Command("setx", "WIRE_API", "responses")
			cmd4.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			cmd4.Run()
		}
	}()
}

func createVersionCmd(path string) *exec.Cmd {
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:    fmt.Sprintf(`cmd /c ""%s" --version"`, path),
		HideWindow: true,
	}
	return cmd
}



func createNpmInstallCmd(npmPath string, args []string) *exec.Cmd {
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " &^") {
			quotedArgs[i] = fmt.Sprintf(`"%s"`, arg)
		} else {
			quotedArgs[i] = arg
		}
	}
	cmdLine := fmt.Sprintf(`cmd /c ""%s" %s"`, npmPath, strings.Join(quotedArgs, " "))
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:    cmdLine,
		HideWindow: true,
	}
	return cmd
}

func createCondaEnvListCmd(condaCmd string) *exec.Cmd {
	cmd := exec.Command("cmd", "/c", condaCmd, "env", "list")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

func (a *App) ensureLocalNodeBinary() {
	home, _ := os.UserHomeDir()
	localNodeDir := filepath.Join(home, ".cceasy", "tools")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(localNodeDir, 0755); err != nil {
		a.log("Failed to create local tools dir: " + err.Error())
		return
	}

	localNodeExe := filepath.Join(localNodeDir, "node.exe")

	if _, err := os.Stat(localNodeExe); err == nil {
		// node.exe already exists
		return
	}

	// Find system node.exe
	systemNode, err := exec.LookPath("node")
	if err != nil {
		// Try common paths
		commonPaths := []string{
			`C:\Program Files\nodejs\node.exe`,
			filepath.Join(os.Getenv("AppData"), "npm", "node.exe"),
		}
		for _, p := range commonPaths {
			if _, err := os.Stat(p); err == nil {
				systemNode = p
				break
			}
		}
	}

	if systemNode == "" {
		a.log("Warning: Could not find system node.exe to copy to local tool dir.")
		return
	}

	a.log(fmt.Sprintf("Copying node.exe from %s to %s to ensure wrapper compatibility...", systemNode, localNodeExe))

	// Copy the file
	input, err := os.ReadFile(systemNode)
	if err != nil {
		a.log("Failed to read system node.exe: " + err.Error())
		return
	}

	if err := os.WriteFile(localNodeExe, input, 0755); err != nil {
		a.log("Failed to write local node.exe: " + err.Error())
		return
	}
	
	a.log("Successfully copied node.exe to local directory.")
}

