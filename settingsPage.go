package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// makeSettingsPageContent creates the settings page content
func makeSettingsPageContent(ctx *AppContext, navChannel chan NavEvent) fyne.CanvasObject {
	settings := ctx.Settings
	domainEntry := widget.NewEntry()
	domainEntry.Text = settings.Domain

	domainLabel := widget.NewRichTextFromMarkdown("**Domain**")
	dnsEntry := widget.NewMultiLineEntry()
	dnsEntry.Wrapping = fyne.TextWrapBreak
	dnsEntry.Text = settings.ResolverHost

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
	// Validator function
	addressEntry.Validator = func(s string) error {
		host, port, err := net.SplitHostPort(s)
		if err != nil {
			return fmt.Errorf("input must be in hostname:port format")
		}

		// Optionally validate the hostname and port
		if _, err := net.LookupHost(host); err != nil {
			return fmt.Errorf("invalid hostname")
		}

		if p, err := strconv.Atoi(port); err != nil || p < 1 || p > 65535 {
			return fmt.Errorf("invalid port")
		}

		return nil
	}
	addressEntry.SetPlaceHolder("Enter proxy local address")
	addressEntry.Text = settings.LocalAddress

	saveButton := widget.NewButton("Save", func() {
		ctx.Settings.Domain = domainEntry.Text
		ctx.Settings.ResolverHost = dnsEntry.Text
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

	advancedSettings := widget.NewAccordionItem("Advanced",
		container.NewVBox(domainLabel,
			domainEntry,
			dnsLabel,
			dnsEntry,
			protocolSelect,
			reporterLabel,
			reporterEntry,
		))

	accordion := widget.NewAccordion(advancedSettings)

	return container.NewVBox(
		header,
		addressEntryLabel,
		addressEntry,
		accordion,
		layout.NewSpacer(),
		saveButton,
	)
}
