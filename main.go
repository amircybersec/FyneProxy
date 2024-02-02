package main

import (
	"encoding/json"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var selectedItemID int

type AppState struct {
	CurrentPage string
}

type NavEvent struct {
	TargetPage string
}

type AppSettings struct {
	Domain       string   `json:"domain"`
	DnsList      string   `json:"dnsList"`
	Tcp          bool     `json:"tcp"`
	Udp          bool     `json:"udp"`
	ReporterURL  string   `json:"reporter"`
	LocalAddress string   `json:"localAddress"`
	Configs      []Config `json:"configs"`
}

type Config struct {
	Transport   string                `json:"transport"`
	TestReports []*connectivityReport `json:"testReport"`
	Health      int                   `json:"health"`
}

// 0: healthly, 1: some tests failed, 2: all tests failed
type AppContext struct {
	Window      fyne.Window
	Preferences fyne.Preferences
	Settings    *AppSettings
}

func main() {
	log.Println("Setting the context")
	ProxyApp := app.NewWithID("FyneProxyApp")
	if meta := ProxyApp.Metadata(); meta.Name == "" {
		// App not packaged, probably from `go run`.
		meta.Name = "Proxy App"
		app.SetMetadata(meta)
	}
	ProxyApp.Settings().SetTheme(newAppTheme())
	mainWin := ProxyApp.NewWindow(ProxyApp.Metadata().Name)
	mainWin.Resize(fyne.NewSize(200, 300))

	log.Println("Setting the context")

	ctx := &AppContext{
		Window:      mainWin,
		Preferences: ProxyApp.Preferences(),
		Settings:    &AppSettings{},
	}
	// Load settings from preferences
	loadSettings(ctx)

	// State variable
	state := &AppState{CurrentPage: "main"}

	// Channel for navigation events
	navChannel := make(chan NavEvent)

	log.Println("Starting the app")

	// Start listening to navigation events
	go func() {
		for event := range navChannel {
			state.CurrentPage = event.TargetPage
			mainWin.SetContent(makePageContent(ctx, state, navChannel))
		}
	}()

	// Set initial content
	mainWin.SetContent(makePageContent(ctx, state, navChannel))
	mainWin.ShowAndRun()
}

func loadSettings(ctx *AppContext) {
	// Create your settings content here
	var settings AppSettings
	settingsJSON := ctx.Preferences.String("AppSettings")
	if settingsJSON != "" {
		err := json.Unmarshal([]byte(settingsJSON), &settings)
		if err != nil {
			log.Println("Error loading settings:", err)
		}
		ctx.Settings = &settings
	}
}

func updateSettings(ctx *AppContext) {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(ctx.Settings)
	if err != nil {
		log.Println("Error saving settings:", err)
		return
	}
	// Save JSON string to app preferences
	ctx.Preferences.SetString("AppSettings", string(settingsJSON))
}
