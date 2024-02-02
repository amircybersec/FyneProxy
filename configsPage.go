package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func makeConfigsPage(ctx *AppContext, navChannel chan NavEvent) fyne.CanvasObject {
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
