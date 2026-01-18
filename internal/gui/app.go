// Package gui provides the graphical user interface for Multiablo.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	// AppID is the unique identifier for the application
	AppID = "com.github.chenwei791129.multiablo"
	// AppTitle is the window title
	AppTitle = "Multiablo - D2R Multi-Instance Helper"
)

// App wraps the Fyne application
type App struct {
	fyneApp fyne.App
	window  *MainWindow
}

// NewApp creates a new GUI application
func NewApp() *App {
	a := app.NewWithID(AppID)
	return &App{
		fyneApp: a,
	}
}

// Run starts the application
func (a *App) Run() {
	a.window = NewMainWindow(a.fyneApp)
	a.window.Show()
	a.window.StartMonitoringAutomatically()
	a.fyneApp.Run()
}

// Quit closes the application
func (a *App) Quit() {
	a.fyneApp.Quit()
}
