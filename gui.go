package main

import (
	"fmt"
	"image/color"
	"log"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func makePageContent(ctx *AppContext, state *AppState, navChannel chan NavEvent) fyne.CanvasObject {
	switch state.CurrentPage {
	case "main":
		fmt.Println("rendering the main page")
		return makeMainPageContent(ctx, navChannel)
	case "settings":
		fmt.Println("rendering the settings page")
		return makeSettingsPageContent(ctx, navChannel)
	case "testResult":
		fmt.Println("rendering the test result page")
		return makeTestResultPage(ctx, navChannel)
	// Add more cases for different pages
	default:
		return widget.NewLabel("Page not found")
	}
}

// makeMainPageContent creates the main page content
func makeMainPageContent(ctx *AppContext, navChannel chan NavEvent) fyne.CanvasObject {
	var selectedItemID int
	var list *widget.List

	list = widget.NewList(
		func() int {
			return len(ctx.Settings.Configs)
		},
		func() fyne.CanvasObject {
			// Create and initialize the toolbar here
			indicator := canvas.NewRectangle(color.Transparent)
			indicator.SetMinSize(fyne.NewSize(10, 10))
			label := widget.NewLabel("")
			toolbar := widget.NewToolbar()

			return container.NewHBox(
				indicator,
				label,
				layout.NewSpacer(),
				toolbar,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			container := o.(*fyne.Container)
			indicator := container.Objects[0].(*canvas.Rectangle)
			label := container.Objects[1].(*widget.Label)
			toolbar := container.Objects[3].(*widget.Toolbar)

			switch ctx.Settings.Configs[i].Health {
			case 0:
				indicator.FillColor = color.Transparent
			case 1:
				// all passed
				indicator.FillColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
			case 2:
				// some failed
				indicator.FillColor = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
			case 3:
				// all tests failed
				indicator.FillColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
			}
			u, err := url.Parse(ctx.Settings.Configs[i].Transport)
			if err != nil {
				label.SetText("Parse error")
			} else {
				label.SetText(u.Host)
			}
			// Clear previous toolbar items and add a new delete icon
			toolbar.Items = nil

			arrowIcon := widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
				log.Printf("Next icon clicked for item %v", i)
				// navigate to page result for specific menu item
				navChannel <- NavEvent{TargetPage: "testResult"}
				// Define action for the "+" icon
			})

			// Create a new delete icon action for each item
			deleteIcon := widget.NewToolbarAction(theme.DeleteIcon(), func() {
				// Delete the item from the data slice
				ctx.Settings.Configs = append(ctx.Settings.Configs[:i], ctx.Settings.Configs[i+1:]...)
				updateSettings(ctx)
				// Refresh the list to update the view
				list.Refresh()

				log.Printf("Delete icon clicked - Item %d deleted", i)
			})
			toolbar.Append(deleteIcon)
			toolbar.Append(arrowIcon)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		//selectedItem := ctx.Settings.Configs[id].Transport
		log.Printf("Selected Item ID %v", selectedItemID)
		selectedItemID = id
	}
	// list.OnUnselected = func(o fyne.CanvasObject) {
	// }

	// Create the scroll container for the list
	scrollContainer := container.NewScroll(list)
	// Set a minimum size for the scroll container if needed
	scrollContainer.SetMinSize(fyne.NewSize(400, 300)) // Set width and height as needed
	// Use container.Max to allocate as much space as possible to the list
	listWithMaxHeight := container.NewStack(scrollContainer)

	// Create the toolbar with a "+" icon
	headerToolbarRight := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			log.Println("Add icon clicked")
			// Define action for the "+" icon
			c := ctx.Window.Clipboard().Content()
			// process clipboard content
			configURLs, err := parseInputText(c)
			if err == nil {
				ctx.Settings.Configs = append(ctx.Settings.Configs, configURLs...)
				updateSettings(ctx)
			} else {
				log.Println("Error parsing clipboard content:", err)
			}
			list.Refresh()
		}),
	)

	// Create the toolbar with settings icon
	headerToolbarLeft := widget.NewToolbar(
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			log.Println("Setting icon clicked")
			navChannel <- NavEvent{TargetPage: "settings"}
			// myWindow.SetContent(makeSettingsPageContent())
			// Define action for the "+" icon
		}),
	)

	header := makePageHeader("Proxy App", headerToolbarLeft, headerToolbarRight)

	ConnectButton := widget.NewButton("Connect", func() {})
	ConnectButton.Importance = widget.HighImportance

	statusBox := widget.NewLabel("")
	statusBox.Wrapping = fyne.TextWrapWord

	setProxyUI := func(proxy *runningProxy, err error) {
		if proxy != nil {
			statusBox.SetText("Proxy listening on " + proxy.Address)
			ConnectButton.SetText("Stop")
			ConnectButton.SetIcon(theme.MediaStopIcon())
			return
		}
		if err != nil {
			statusBox.SetText("❌ ERROR: " + err.Error())
		} else {
			statusBox.SetText("Proxy not running")
		}
		ConnectButton.SetText("Connect")
		ConnectButton.SetIcon(theme.MediaPlayIcon())
	}
	var proxy *runningProxy
	ConnectButton.OnTapped = func() {
		log.Println(ConnectButton.Text)
		var err error
		if proxy == nil {
			// Start proxy.
			log.Printf("Starting proxy on %v", ctx.Settings.LocalAddress)
			log.Printf("Using config: %v", ctx.Settings.Configs[selectedItemID].Transport)
			proxy, err = runServer(ctx.Settings.LocalAddress, ctx.Settings.Configs[selectedItemID].Transport)
		} else {
			// Stop proxy
			proxy.Close()
			proxy = nil
		}
		setProxyUI(proxy, err)
	}
	setProxyUI(proxy, nil)

	buttonState := make(chan bool)
	TestButton := widget.NewButton("Test All", func() {
		go func() {
			// Disable the button and update text in the main goroutine
			buttonState <- true

			// test all configs
			TestConfigs(ctx.Settings)
			log.Printf("Test reports: %v", ctx.Settings.Configs)
			submitReports(ctx.Settings)
			list.Refresh()

			// Re-enable the button and reset text in the main goroutine
			buttonState <- false
		}()
	})
	TestButton.Importance = widget.HighImportance

	go func() {
		for update := range buttonState {
			if update {
				TestButton.Disable()
				TestButton.SetText("Testing...")
			} else {
				TestButton.SetText("Test All")
				TestButton.Enable()
			}
		}
	}()

	// Combine the header and the scrollable list in a vertical box layout
	return container.NewVBox(
		header,
		listWithMaxHeight, // The scrollable list with enforced maximum height
		container.New(layout.NewGridLayoutWithColumns(2), TestButton, ConnectButton),
		statusBox,
	)

}

// makeSettingsPageContent creates the settings page content
func makeSettingsPageContent(ctx *AppContext, navChannel chan NavEvent) fyne.CanvasObject {
	settings := ctx.Settings
	domainEntry := widget.NewEntry()
	if settings.Domain != "" {
		domainEntry.Text = settings.Domain
	} else {
		domainEntry.Text = "example.com"
	}
	domainLabel := widget.NewRichTextFromMarkdown("**Domain**")

	dnsEntry := widget.NewMultiLineEntry()
	dnsEntry.Wrapping = fyne.TextWrapBreak
	if settings.DnsList != "" {
		dnsEntry.Text = settings.DnsList
	} else {
		dnsEntry.Text = "8.8.8.8"
	}

	dnsLabel := widget.NewRichTextFromMarkdown("**Resolvers** ([format](https://pkg.go.dev/github.com/Jigsaw-Code/outline-sdk/x/config#hdr-Config_Format))")

	reporterLabel := widget.NewRichTextFromMarkdown("**Reporter URL** ([format](https://pkg.go.dev/github.com/Jigsaw-Code/outline-sdk/x/config#hdr-Config_Format))")
	reporterEntry := widget.NewEntry()
	reporterEntry.Text = settings.ReporterURL

	checkUDP := widget.NewCheck("UDP", func(value bool) {
		log.Println("Check set to", value)
		settings.Udp = value
	})
	checkUDP.Checked = settings.Udp

	checkTCP := widget.NewCheck("TCP", func(value bool) {
		log.Println("Check set to", value)
		settings.Udp = value
	})
	checkTCP.Checked = settings.Tcp

	protocolSelect := container.NewHBox(
		checkUDP,
		checkTCP,
	)

	addressEntryLabel := widget.NewLabelWithStyle("Local address", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Enter proxy local address")
	if settings.LocalAddress == "" {
		addressEntry.Text = "localhost:8080"
	}

	saveButton := widget.NewButton("Save", func() {
		ctx.Settings.Domain = domainEntry.Text
		ctx.Settings.DnsList = dnsEntry.Text
		ctx.Settings.ReporterURL = reporterEntry.Text
		ctx.Settings.Udp = checkUDP.Checked
		ctx.Settings.Tcp = checkTCP.Checked
		ctx.Settings.LocalAddress = addressEntry.Text
		updateSettings(ctx)
	})
	saveButton.Importance = widget.HighImportance

	// Create the toolbar with back "<-" icon
	headerToolbarLeft := widget.NewToolbar(
		widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
			log.Println("Setting icon clicked")
			navChannel <- NavEvent{TargetPage: "main"}
			// myWindow.SetContent(makeSettingsPageContent())
			// Define action for the "+" icon
		}),
	)
	// Create the toolbar with a "+" icon
	headerToolbarRight := widget.NewToolbar()
	header := makePageHeader("Settings", headerToolbarLeft, headerToolbarRight)
	// Combine the header and the scrollable list in a vertical box layout
	return container.NewVBox(
		header,
		domainLabel,
		domainEntry,
		dnsLabel,
		dnsEntry,
		protocolSelect,
		reporterLabel,
		reporterEntry,
		addressEntryLabel,
		addressEntry,
		saveButton,
	)
}

func makePageHeader(title string, leftToolbar *widget.Toolbar, rightToolbar *widget.Toolbar) *fyne.Container {
	// Create the header label
	headerLabel := widget.NewLabel(title)
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}
	headerLabel.Alignment = fyne.TextAlignCenter

	// Create the header using HBox layout
	header := container.NewHBox(
		leftToolbar,
		layout.NewSpacer(),
		headerLabel,
		layout.NewSpacer(), // Spacer pushes the toolbar to the right
		rightToolbar,
	)
	return header
}

func parseInputText(clipboardContent string) ([]Config, error) {
	u, err := url.Parse(clipboardContent)
	// if parse is successful, check the schema
	if err == nil {
		switch u.Scheme {
		case "ss", "socks5", "tls", "split":
			// try to parse ss url
			return []Config{{Transport: u.String()}}, nil
		case "https":
			// fetch list from remote config
			return []Config{}, fmt.Errorf("not implemented yet")
		case "http":
			// reject url due to security issue
			return []Config{}, fmt.Errorf("not implemented yet")
		default:
			// reject url due to unknown schema
			return []Config{}, fmt.Errorf("not implemented yet")
		}
	}
	return []Config{}, fmt.Errorf("failed to parse input")
}

func makeTestResultPage(ctx *AppContext, navChannel chan NavEvent) fyne.CanvasObject {
	// Create the toolbar with back "<-" icon
	headerToolbarLeft := widget.NewToolbar(
		widget.NewToolbarAction(theme.NavigateBackIcon(), func() {
			log.Println("Setting icon clicked")
			navChannel <- NavEvent{TargetPage: "main"}
			// myWindow.SetContent(makeSettingsPageContent())
			// Define action for the "+" icon
		}),
	)
	// Create empty toolbar
	headerToolbarRight := widget.NewToolbar()
	header := makePageHeader("Test Result", headerToolbarLeft, headerToolbarRight)
	// Combine the header and the accordions in a vertical box layout
	return container.NewVBox(
		header,
	)
}
