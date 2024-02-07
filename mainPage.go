package main

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"net"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// customProgressBar extends widget.ProgressBar to set a custom minimum size.
type customProgressBar struct {
	widget.ProgressBarInfinite
	minSize fyne.Size
}

// NewCustomProgressBar creates a new instance of customProgressBar with a specified minimum size.
func NewCustomProgressBar(minSize fyne.Size) *customProgressBar {
	progressBar := &customProgressBar{}
	progressBar.minSize = minSize
	progressBar.ExtendBaseWidget(progressBar)
	return progressBar
}

// MinSize returns the custom minimum size for the progress bar.
func (c *customProgressBar) MinSize() fyne.Size {
	// Return the larger of the default minimum size or the specified minimum size.
	return fyne.NewSize(5, 5)
}

// makeMainPageContent creates the main page content
func makeMainPageContent(ctx *AppContext, navChannel chan NavEvent) fyne.CanvasObject {
	var list *widget.List

	list = widget.NewList(
		func() int {
			return len(ctx.Settings.Configs)
		},
		func() fyne.CanvasObject {
			// Create and initialize the toolbar here
			// indicator := canvas.NewRectangle(color.Transparent)
			// indicator.SetMinSize(fyne.NewSize(10, 10))
			selected := canvas.NewRectangle(color.Transparent)
			selected.SetMinSize(fyne.NewSize(10, 10))
			indicator := widget.NewIcon(theme.ViewRefreshIcon())
			label := widget.NewLabel("")
			toolbar := widget.NewToolbar()

			return container.NewHBox(
				selected,
				indicator,
				label,
				layout.NewSpacer(),
				toolbar,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			container := o.(*fyne.Container)
			selected := container.Objects[0].(*canvas.Rectangle)
			indicator := container.Objects[1].(*widget.Icon)
			label := container.Objects[2].(*widget.Label)
			toolbar := container.Objects[4].(*widget.Toolbar)

			switch ctx.Settings.Configs[i].Health {
			case 0:
				indicator.SetResource(theme.ViewRefreshIcon())
			case 1:
				// all passed
				indicator.SetResource(theme.ConfirmIcon())
			case 2:
				// some failed
				indicator.SetResource(theme.WarningIcon())
			case 3:
				// all tests failed
				indicator.SetResource(theme.ErrorIcon())
			}
			u, err := url.Parse(ctx.Settings.Configs[i].Transport)
			if err != nil {
				label.SetText("Parse error")
			} else {
				label.SetText(u.Host)
			}

			if i == selectedItemID {
				// Set the selected item style
				selected.FillColor = theme.PrimaryColor()
			} else {
				// Reset the style for the unselected items
				selected.FillColor = color.Transparent
			}
			// Clear previous toolbar items and add a new delete icon
			toolbar.Items = nil

			arrowIcon := widget.NewToolbarAction(theme.NavigateNextIcon(), func() {
				log.Printf("Next icon clicked for item %v", i)
				// navigate to page result for specific menu item
				navChannel <- NavEvent{TargetPage: "configs"}
				// Define action for the "+" icon
			})

			// Create a new delete icon action for each item
			deleteIcon := widget.NewToolbarAction(theme.DeleteIcon(), func() {
				callback := func(confirm bool) {
					if confirm {
						// Delete the item from the data slice
						ctx.Settings.Configs = append(ctx.Settings.Configs[:i], ctx.Settings.Configs[i+1:]...)
						updateSettings(ctx)
						// Refresh the list to update the view
						list.Refresh()
						log.Printf("Delete icon clicked - Item %d deleted", i)
					}
				}
				dialog := dialog.NewConfirm("Confirm Delete", "Sure to delete config?", callback, ctx.Window)
				dialog.Show()
			})
			toolbar.Append(deleteIcon)
			toolbar.Append(arrowIcon)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		//selectedItem := ctx.Settings.Configs[id].Transport
		log.Printf("Selected Item ID %v", selectedItemID)
		selectedItemID = id
		list.Refresh()
	}
	// list.OnUnselected = func(o fyne.CanvasObject) {
	// }

	// Create the scroll container for the list
	scrollContainer := container.NewScroll(list)
	// Set a minimum size for the scroll container if needed
	scrollContainer.SetMinSize(fyne.NewSize(400, 300)) // Set width and height as needed
	// Use container.Max to allocate as much space as possible to the list
	listWithMaxHeight := container.NewStack(scrollContainer)

	// addConfig := func(ctx *AppContext) {
	// 	log.Println("Add icon clicked")
	// 	// Define action for the "+" icon
	// 	c := ctx.Window.Clipboard().Content()
	// 	// process clipboard content
	// 	configURLs, err := parseInputText(c)
	// 	if err == nil {
	// 		ctx.Settings.Configs = append(ctx.Settings.Configs, configURLs...)
	// 		updateSettings(ctx)
	// 	} else {
	// 		log.Println("Error parsing clipboard content:", err)
	// 	}
	// }

	// Create the toolbar with a "+" icon
	headerToolbarRight := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			//addConfig(ctx)
			//name := widget.NewEntry()
			//name.SetPlaceHolder("Give it a name (optional)")
			inputURL := widget.NewEntry()
			inputURL.SetPlaceHolder("Enter config here")
			paste := widget.NewToolbar(widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
				c := ctx.Window.Clipboard().Content()
				inputURL.SetText(c)
			}))
			content := container.NewVBox(inputURL, paste)
			dialog.ShowCustomConfirm("Add Config", "Add", "Cancel", content, func(confirm bool) {
				if confirm {
					configURLs, err := parseInputText(inputURL.Text)
					if err == nil {
						ctx.Settings.Configs = append(ctx.Settings.Configs, configURLs...)
						updateSettings(ctx)
					} else {
						log.Println("Error parsing clipboard content:", err)
					}
					list.Refresh()
				}
			}, ctx.Window)
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

	minSize := fyne.NewSize(5, 5) // Set your desired minimum width and height
	progressBar := NewCustomProgressBar(minSize)
	progressBar.Hide()

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
		TestSingleConfig(ctx.Settings, selectedItemID)
		list.Refresh()
		var err error
		systemProxy, err := GetSystemProxy()
		if err != nil {
			fmt.Println(err)
		}
		if proxy == nil {
			// Start proxy.
			log.Printf("Starting proxy on %v", ctx.Settings.LocalAddress)
			log.Printf("Using config: %v", ctx.Settings.Configs[selectedItemID].Transport)
			if ctx.Settings.Configs[selectedItemID].Health == 1 {
				proxy, err = runServer(ctx.Settings.LocalAddress, ctx.Settings.Configs[selectedItemID].Transport)
				host, port, err := net.SplitHostPort(ctx.Settings.LocalAddress)
				if err != nil {
					fmt.Println("failed to parse address: %w", err)
				}
				if err := systemProxy.SetProxy(host, port); err != nil {
					fmt.Println("Error setting up proxy:", err)
				} else {
					fmt.Println("Proxy setup successful")
				}
			} else {
				err = errors.New("could not connecto to remote destination")
				proxy = nil
			}
		} else {
			// Stop proxy
			proxy.Close()
			if err := systemProxy.UnsetProxy(); err != nil {
				fmt.Println("Error setting up proxy:", err)
			} else {
				fmt.Println("Proxy unset successful")
			}
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
			updateSettings(ctx)
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
				progressBar.Show()
			} else {
				TestButton.SetText("Test All")
				TestButton.Enable()
				progressBar.Hide()
			}
		}
	}()

	// progressBarContainer := container.New(layout.NewFixedGridLayout(fyne.NewSize(fixedWidth, progressBar.MinSize().Height)), progressBar)

	// Combine the header and the scrollable list in a vertical box layout
	return container.NewVBox(
		header,
		listWithMaxHeight, // The scrollable list with enforced maximum height
		progressBar,
		container.New(layout.NewGridLayoutWithColumns(2), TestButton, ConnectButton),
		statusBox,
	)
}
