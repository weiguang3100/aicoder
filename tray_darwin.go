// +build darwin

package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func setupTray(app *App, appOptions *options.App) {
	// For macOS, we use the Application Menu instead of Tray to avoid AppDelegate conflict
	
	appMenu := menu.NewMenu()
	
	// App specific menu
	appMenu.Append(menu.AppMenu())
	
	// Add standard Edit menu to enable Copy/Paste shortcuts
	appMenu.Append(menu.EditMenu())
	
	// Model switching menu
	modelMenu := appMenu.AddSubmenu("Models")
	modelItems := make(map[string]*menu.MenuItem)
	
	config, _ := app.LoadConfig()
	for _, model := range config.Models {
		modelName := model.ModelName
		m := modelMenu.AddCheckbox(modelName, modelName == config.CurrentModel, nil, func(cd *menu.CallbackData) {
			currentConfig, _ := app.LoadConfig()
			currentConfig.CurrentModel = modelName
			app.SaveConfig(currentConfig)
		})
		modelItems[modelName] = m
	}
	
	// Actions menu
	actionsMenu := appMenu.AddSubmenu("Actions")
	mShow := actionsMenu.AddText("Show Main Window", nil, func(cd *menu.CallbackData) {
		runtime.WindowShow(app.ctx)
	})
	mLaunch := actionsMenu.AddText("Launch Claude Code", nil, func(cd *menu.CallbackData) {
		cfg, _ := app.LoadConfig()
		app.LaunchClaude(false, cfg.ProjectDir)
	})

	appOptions.Menu = appMenu
	
	appOptions.OnStartup = func(ctx context.Context) {
		app.startup(ctx)
		
		// Register update function
		UpdateTrayMenu = func(lang string) {
			t, ok := trayTranslations[lang]
			if !ok {
				t = trayTranslations["en"]
			}
			
			// Update Submenu Labels
			// Note: In Wails v2, MenuItem.Label can be updated directly
			// The submenus themselves are MenuItems in the parent menu
			for _, item := range appMenu.Items {
				if item.SubMenu != nil {
					if item.Label == "Models" || item.Label == trayTranslations["en"]["models"] || containsValue(trayTranslations, "models", item.Label) {
						item.Label = t["models"]
					} else if item.Label == "Actions" || item.Label == trayTranslations["en"]["actions"] || containsValue(trayTranslations, "actions", item.Label) {
						item.Label = t["actions"]
					}
				}
			}
			
			mShow.Label = t["show"]
			mLaunch.Label = t["launch"]
			
			runtime.MenuSetApplicationMenu(ctx, appMenu)
		}
		
		// Register config change listener
		OnConfigChanged = func(cfg AppConfig) {
			for name, item := range modelItems {
				item.Checked = (name == cfg.CurrentModel)
			}
			runtime.MenuSetApplicationMenu(ctx, appMenu)
			runtime.EventsEmit(app.ctx, "config-changed", cfg)
		}

		// Initial language sync
		if app.CurrentLanguage != "" {
			UpdateTrayMenu(app.CurrentLanguage)
		}
	}
}

func containsValue(translations map[string]map[string]string, key string, value string) bool {
	for _, t := range translations {
		if t[key] == value {
			return true
		}
	}
	return false
}