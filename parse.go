package main

import (
	"fmt"
	"net/url"
)

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
