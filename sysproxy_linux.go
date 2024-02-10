//go:build linux

package main

import (
	"os"
	"os/exec"
)

func SetProxy(ip string, port string) error {
	// Execute Linux specific commands to set proxy
	// Example: exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", ip).Run()
	execCommand("gsettings", "set", "org.gnome.system.proxy", "mode", "manual")
	execCommand("gsettings", "set", "org.gnome.system.proxy.http", "host", ip)
	execCommand("gsettings", "set", "org.gnome.system.proxy.http", "port", port)
	execCommand("gsettings", "set", "org.gnome.system.proxy.https", "host", ip)
	execCommand("gsettings", "set", "org.gnome.system.proxy.https", "port", port)
	execCommand("gsettings", "set", "org.gnome.system.proxy.ftp", "host", ip)
	execCommand("gsettings", "set", "org.gnome.system.proxy.ftp", "port", port)
	return nil
}

func UnsetProxy() error {
	// Execute Linux specific commands to unset proxy
	// gsettings set org.gnome.system.proxy mode none
	execCommand("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	return nil
}

func execCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	err := cmd.Run()
	if err != nil {
		println("Failed to execute command:", err)
		os.Exit(1)
	}
}
