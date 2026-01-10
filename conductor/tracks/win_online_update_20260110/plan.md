# Implementation Plan: Windows Online Update Feature

## Phase 1: Backend Implementation (Go)
This phase focuses on adding the necessary logic to handle file downloading with progress and launching the installer on Windows.

- [x] **Task 1: Implement `GetDownloadsFolder`** (c7c63fc)
  - Define a helper function to retrieve the user's "Downloads" folder path using Windows shell APIs or environment variables.
  - *Status:* `[ ]`
- [ ] **Task 2: Implement `DownloadUpdate` with Progress Tracking**
  - Add a method `DownloadUpdate(url string, fileName string)` to `App` struct.
  - Use `http.Get` to fetch the file.
  - Use a custom writer to track progress and emit events to the frontend via `runtime.EventsEmit`.
  - Handle download cancellation.
  - *Status:* `[ ]`
- [ ] **Task 3: Implement `LaunchInstallerAndExit`**
  - Add a method to execute the downloaded `.exe` file using `os/exec`.
  - Ensure the application exits immediately after launching the installer.
  - *Status:* `[ ]`
- [ ] **Task: Conductor - User Manual Verification 'Phase 1: Backend' (Protocol in workflow.md)**
  - *Status:* `[ ]`

## Phase 2: Frontend Implementation (React/TS)
This phase updates the UI to reflect the new update workflow on Windows.

- [ ] **Task 1: Update I18n Labels**
  - Change "检查更新" to "在线更新" for Windows in `frontend/src/App.tsx`.
  - Add new strings for "Downloading...", "Download Cancelled", "Install Now", etc.
  - *Status:* `[ ]`
- [ ] **Task 2: Enhance `UpdateModal` with Download State**
  - Introduce states for `isDownloading`, `downloadProgress`, and `downloadError`.
  - Replace the static download link with a "Download and Update" button for Windows.
  - Implement the progress bar UI.
  - *Status:* `[ ]`
- [ ] **Task 3: Connect Frontend to Backend Update Logic**
  - Listen for download progress events emitted by the backend.
  - Wire the "Download and Update" button to call `DownloadUpdate`.
  - Wire the "Install Now" button to call `LaunchInstallerAndExit`.
  - *Status:* `[ ]`
- [ ] **Task: Conductor - User Manual Verification 'Phase 2: Frontend' (Protocol in workflow.md)**
  - *Status:* `[ ]`

## Phase 3: Final Integration and Verification
- [ ] **Task 1: End-to-End Test on Windows**
  - Verify the entire flow: Check -> Download -> Progress UI -> Launch Installer -> App Exit.
  - *Status:* `[ ]`
- [ ] **Task: Conductor - User Manual Verification 'Phase 3: Final Integration' (Protocol in workflow.md)**
  - *Status:* `[ ]`
