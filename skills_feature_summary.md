# Skills Feature Implementation

I have successfully added the "Skills" feature to the application.

## Changes

1.  **Backend (`app.go`)**:
    *   Added `Skill` struct to define skill data (Name, Description, Type, Value).
    *   Added `GetSkillsDir()` to resolve `~/.cceasy/skills`.
    *   Added `AddSkill()` to save skills to `metadata.json` and copy zip files to the skills directory.
    *   Added `DeleteSkill()` to remove skills from metadata and delete associated zip files.
    *   Added `ListSkills()` to retrieve the list of skills.
    *   Added `SelectSkillFile()` to open a system file dialog for selecting .zip files.

2.  **Frontend (`frontend/src/App.tsx`)**:
    *   Added a "Skills" (🛠️) button to the sidebar below "Tutorial".
    *   Implemented the Skills view:
        *   Lists existing skills with their details.
        *   "Add Skill" button opens a modal.
        *   Modal supports two types: "Skill Address" (text input) and "Zip Package" (file selection).
        *   Includes fields for Name and Description.
        *   Delete button (🗑️) for each skill.
    *   Added translations for English, Simplified Chinese, and Traditional Chinese.

3.  **Wails Bindings (`frontend/wailsjs/`)**:
    *   Updated `go/models.ts` to include the `Skill` class.
    *   Updated `go/main/App.js` and `go/main/App.d.ts` to expose the new backend methods to the frontend.

## Usage

1.  Click the "Skills" (🛠️) icon in the left sidebar.
2.  Click "Add Skill" to open the creation modal.
3.  Choose "Skill Address" to enter a URL-like address, or "Zip Package" to browse for a local `.zip` file.
4.  Enter a Name and Description.
5.  Click "Confirm" to save. The skill will appear in the list.
6.  Zip files are automatically copied to your user directory under `.cceasy/skills`.
