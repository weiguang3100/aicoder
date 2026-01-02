# FAQ - Frequently Asked Questions

## 1. Why is the system tray icon unresponsive?
In earlier versions, if background operations (such as file I/O) blocked the main thread, the tray icon might temporarily become unresponsive. The current version has optimized this issue through asynchronous processing and OS thread locking. If you still encounter this, please try restarting the program.

## 2. How to use a Custom Model?
1. Select the AI tool (e.g., Claude) in the sidebar.
2. Click "Model Settings".
3. Select the "Custom" tab.
4. Enter your model name (e.g., `claude-3-5-sonnet-20241022`).
5. Enter an API Endpoint compatible with the protocol.
6. Enter your API Key and save.

## 3. My API Key is not working?
The preset shortcuts in AICoder **may only support specific "Coding Plan" API Keys** provided by each vendor.
If you are using a general-purpose API Key, please use the **"Custom"** mode and manually enter the corresponding model name and API endpoint.

## 4. Where is the configuration file saved?
AICoder's configuration is saved in your user home directory with the filename `.aicoder_config.json`.
Native settings for various AI tools (like Claude's `~/.claude/settings.json`) are also automatically synced based on your configuration.

## 5. How to update AI CLI tools?
Each time AICoder starts, it automatically checks the versions of supported tools (like `claude-code`, `codex`, `gemini-cli`). If a new version is available, it will attempt to update it for you. You can see the specific status in the startup progress window.

## 6. What if the environment check fails?
If Node.js or tool installation fails, please check your internet connection. In mainland China, the program automatically attempts to use domestic mirrors to speed up downloads. If automatic installation continues to fail, it is recommended to manually install the environment as prompted.

## 7. How to use the original model services of each tool?
Select **AICoderMirror** as the provider and enter your API Key. This provider acts as a relay for the original services, allowing easy access to native model services.

## 8. How to use the services provided directly by the tools themselves?
Select **"Original"** as the provider in the tool's settings to restore each tool to its default state. In this mode, you can use the tool's built-in authentication and communication methods directly.

## 9. What does Yolo mode mean?
When Yolo mode is enabled, the programming tool will no longer ask for confirmation before every file or system operation, enhancing the coding experience. However, please be aware that this option is risky and carries the potential for accidental file deletion or modification. It is intended for expert users only.

---
*For more issues, please visit GitHub Issues: [RapidAI/cceasy/issues](https://github.com/RapidAI/cceasy/issues)*