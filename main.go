package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	done := make(chan bool, 1)
	safeClose(done)
	log.Println("Setting the context")
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

	fmt.Println("awaiting signal")
	// The program blocks here waiting for the signal
	<-done
	fmt.Println("exiting")
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

func safeClose(done chan bool) {
	// Setting up a channel to listen for signals
	sigs := make(chan os.Signal, 1)
	// Channel to indicate the program can stop
	// Register the channel to receive notifications of the specified signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Start a goroutine to handle the signals
	// This goroutine executes a blocking receive for signals
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		// Here you can call your cleanup or exit function
		systemProxy, err := GetSystemProxy()
		if err != nil {
			fmt.Println(err)
		}
		if err := systemProxy.UnsetProxy(); err != nil {
			fmt.Println("Error setting up proxy:", err)
		} else {
			fmt.Println("Proxy unset successful")
		}
		done <- true
	}()
}
