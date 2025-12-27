# FAQ - 常见问题解答

## 1. 为什么系统托盘图标点击无反应？
在较早版本中，如果后台操作（如文件读写）阻塞了主线程，可能会导致托盘图标暂时失去响应。当前版本已通过异步处理和线程锁定优化了此问题。如果仍遇到此类情况，请尝试重启程序。

## 2. 如何使用自定义模型 (Custom Model)？
1. 点击“模型设置”。
2. 选择“Custom”标签。
3. 输入您的模型名称（例如 `claude-3-5-sonnet-20241022`）。
4. 输入兼容 Anthropic 协议的 API 端点地址（Endpoint）。
5. 输入 API Key 并保存。

## 3. 我的 API Key 无法工作？
程序中预设的 GLM、Kimi、豆包、MiniMax 快捷选择**仅支持各厂商提供的 "Coding Plan" 专用 API Key**。
如果您使用的是通用型 API Key，请使用 **“Custom”** 模式进行配置，并手动输入对应的模型名称和 API 端点地址。

## 4. 配置文件保存在哪里？
程序配置文件保存在您的用户主目录下，文件名为 `.claude_model_config.json`。
Claude Code 的原生设置保存在 `~/.claude/settings.json`。

## 5. 如何更新 Claude Code？
每次启动程序时，工具会自动检查 `@anthropic-ai/claude-code` 的版本。如果有新版本，它会自动为您运行 `npm install -g @anthropic-ai/claude-code`。您也可以在安装进度日志中查看具体的执行命令。

## 6. 环境检查失败怎么办？
如果 Node.js 安装失败，请检查您的网络连接。在中国大陆地区，程序会自动尝试使用清华大学镜像源以加快下载速度。如果自动安装持续失败，建议手动从 [nodejs.org](https://nodejs.org/) 下载并安装 v22.14.0 或更高版本。

## 8. 什么是“恢复CC”？什么时候需要使用它？
“恢复CC”（Recover CC）功能旨在将 `claude-code` 的运行环境重置为出厂状态。
*   **适用场景**：如果您在手动修改过 Claude 的官方配置、由于 API Key 冲突导致无法登录、或者遇到本程序无法自动修复的环境报错时，建议使用此功能。
*   **操作影响**：它将永久删除 `~/.claude/` 目录下的所有本地配置和认证令牌。
*   **后续操作**：重置成功后，请**不要**立即点击本程序的“启动 Claude Code”，而应先手动打开一个新的终端（CMD 或 PowerShell），输入 `claude` 并按照官方指引重新完成一次基础设置。

---
*更多问题请访问 GitHub Issues：[RapidAI/cceasy/issues](https://github.com/RapidAI/cceasy/issues)*
