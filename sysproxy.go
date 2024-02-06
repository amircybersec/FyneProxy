package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Define an interface for system proxy settings
type SystemProxy interface {
	SetProxy(ip string, port string) error
	UnsetProxy() error
}

// Implement SystemProxy for Linux
type LinuxSystemProxy struct{}

func (p LinuxSystemProxy) SetProxy(ip string, port string) error {
	// Execute Linux specific commands to set proxy
	// Example: exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", ip).Run()
	return nil
}

func (p LinuxSystemProxy) UnsetProxy() error {
	// Execute Linux specific commands to unset proxy
	return nil
}

// Implement SystemProxy for Darwin (macOS)
type DarwinSystemProxy struct{}

func (p DarwinSystemProxy) SetProxy(ip string, port string) error {
	// Execute macOS specific commands to set proxy
	// Get the active network interface
	activeInterface, err := getActiveNetworkInterfaceMacOS()
	if err != nil {
		return err
	}

	// Set the web proxy and secure web proxy
	if err := setProxyMacOS("web", activeInterface, ip, port); err != nil {
		return err
	}
	if err := setProxyMacOS("secureweb", activeInterface, ip, port); err != nil {
		return err
	}

	return nil
}

func (p DarwinSystemProxy) UnsetProxy() error {
	// Execute macOS specific commands to unset proxy
	// Get the active network interface
	activeInterface, err := getActiveNetworkInterfaceMacOS()
	if err != nil {
		return err
	}

	// Set the web proxy and secure web proxy
	if err := removeProxyMacOS("web", activeInterface); err != nil {
		return err
	}
	if err := removeProxyMacOS("secureweb", activeInterface); err != nil {
		return err
	}

	return nil
}

// Implement SystemProxy for Windows
type WindowsSystemProxy struct{}

func (p WindowsSystemProxy) SetProxy(ip string, port string) error {
	// Execute Windows specific commands or registry operations to set proxy
	return nil
}

func (p WindowsSystemProxy) UnsetProxy() error {
	// Execute Windows specific commands or registry operations to unset proxy
	return nil
}

// getActiveNetworkInterface finds the active network interface using shell commands.
func getActiveNetworkInterfaceMacOS() (string, error) {
	cmd := "sh -c \"networksetup -listnetworkserviceorder | grep `route -n get 0.0.0.0 | grep 'interface' | cut -d ':' -f2` -B 1 | head -n 1 | cut -d ' ' -f2\""
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// setProxyMacOS sets the specified type of proxy on the given network interface.
func setProxyMacOS(proxyType string, interfaceName string, ip string, port string) error {
	var cmdStr string
	if proxyType == "web" {
		cmdStr = fmt.Sprintf("networksetup -setwebproxy \"%s\" %s %s", interfaceName, ip, port)
	} else if proxyType == "secureweb" {
		cmdStr = fmt.Sprintf("networksetup -setsecurewebproxy \"%s\" %s %s", interfaceName, ip, port)
	} else {
		return fmt.Errorf("unknown proxy type: %s", proxyType)
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}
	return nil
}

func removeProxyMacOS(proxyType string, interfaceName string) error {
	var cmdStr string
	if proxyType == "web" {
		cmdStr = fmt.Sprintf("networksetup -setwebproxystate \"%s\" off", interfaceName)
	} else if proxyType == "secureweb" {
		cmdStr = fmt.Sprintf("networksetup -setsecurewebproxystate \"%s\" off", interfaceName)
	} else {
		return fmt.Errorf("unknown proxy type: %s", proxyType)
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}
	return nil
}

// detectPlatform returns a string representing the detected platform.
func detectPlatform() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		if arch == "amd64" {
			return "win64"
		} else if arch == "386" {
			return "win32"
		}
	}

	return "unknown"
}

// Factory function to get the appropriate SystemProxy implementation
func GetSystemProxy() (SystemProxy, error) {
	switch detectPlatform() {
	case "darwin":
		return DarwinSystemProxy{}, nil
	case "linux":
		return LinuxSystemProxy{}, nil
	case "win64":
		return WindowsSystemProxy{}, nil
	default:
		return nil, errors.New("unknown or unsupported OS")
	}
}

// func main() {
//     // Example usage
//     if err := setupProxyMacOS("127.0.0.1", "8080"); err != nil {
//         fmt.Println("Error setting up proxy:", err)
//     } else {
//         fmt.Println("Proxy setup successful")
//     }
// }
