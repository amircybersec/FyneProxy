package main

import (
	"encoding/json"
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Jigsaw-Code/outline-sdk/x/sysproxy"
)

var selectedItemID int

type AppState struct {
	CurrentPage string
}

type NavEvent struct {
	TargetPage string
}

type AppSettings struct {
	Domain         string   `json:"domain"`
	ResolverHost   string   `json:"resolverHost"`
	Tcp            bool     `json:"tcp"`
	Udp            bool     `json:"udp"`
	ReporterURL    string   `json:"reporter"`
	LocalAddress   string   `json:"localAddress"`
	Configs        []Config `json:"configs"`
	SmartConfig    []byte   `json:"smartConfig"`
	SmartConfigURL url.URL  `json:"smartConfigURL"`
	BlockedDomains []string `json:"blockedDomains"`
}

type Config struct {
	Transport   string                `json:"transport"`
	ConfigFile  []byte                `json:"configFile"`
	TestReports []*connectivityReport `json:"testReport"`
	Health      int                   `json:"health"`
}

// 0: healthly, 1: some tests failed, 2: all tests failed
type AppContext struct {
	Window      fyne.Window
	Preferences fyne.Preferences
	Settings    *AppSettings
}

var proxy *runningProxy

func main() {
	defer sysproxy.DisableWebProxy()
	ProxyApp := app.NewWithID("com.amirgh.fyneproxy")
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
	printSettings(ctx)

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
	} else {
		// Set default settings if no saved settings are found
		ctx.Settings = &AppSettings{
			Domain:       "example.com",
			ResolverHost: "8.8.8.8",
			Tcp:          true,
			Udp:          true,
			LocalAddress: "localhost:8080",
		}
	}
}

func updateSettings(ctx *AppContext) {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(ctx.Settings)
	if err != nil {
		log.Println("Error marshaling settings:", err)
		return
	}
	// Save JSON string to app preferences
	ctx.Preferences.SetString("AppSettings", string(settingsJSON))
}

func printSettings(ctx *AppContext) {
	// Serialize settings to JSON
	settingsJSON, err := json.Marshal(ctx.Settings)
	if err != nil {
		log.Println("Error marshaling settings:", err)
		return
	}
	log.Printf("Settings: %v", string(settingsJSON))
}
