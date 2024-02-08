//go:build linux

package main

func SetProxy(ip string, port string) error {
	// Execute Linux specific commands to set proxy
	// Example: exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", ip).Run()
	return nil
}

func UnsetProxy() error {
	// Execute Linux specific commands to unset proxy
	return nil
}
