package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	case "configs":
		fmt.Println("rendering the test result page")
		return makeConfigsPage(ctx, navChannel)
	// Add more cases for different pages
	default:
		return widget.NewLabel("Page not found")
	}
}

func makePageHeader(title string, leftToolbar *widget.Toolbar, rightToolbar *widget.Toolbar) *fyne.Container {
	// Create the header label
	headerLabel := widget.NewLabel(title)
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}
	headerLabel.Alignment = fyne.TextAlignCenter

	// Create the header using HBox layout
	header := container.NewHBox(
		leftToolbar,
		layout.NewSpacer(), // Spacer pushes the toolbar to the right
		rightToolbar,
	)
	return container.NewStack(header, headerLabel)
}
