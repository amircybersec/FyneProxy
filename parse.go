package main

import (
	"fmt"
	"net/url"
	"strings"
)

func parseInputText(clipboardContent string) ([]Config, error) {
	u, err := url.Parse(strings.TrimSpace(clipboardContent))
	// if parse is successful, check the schema
	if err == nil {
		switch u.Scheme {
		case "ss", "socks5", "tls", "split":
			// try to parse ss url
			return []Config{{Transport: u.String()}}, nil
		case "https":
			// fetch list from remote config
			var c []Config
			configStrings, err := getDynamicConfig(u.String())
			if err != nil {
				return []Config{}, err
			}
			for _, configString := range configStrings {
				c = append(c, Config{Transport: configString, Health: 0, TestReports: []*connectivityReport{}})
			}
			fmt.Printf("Parsed %d configs from remote url\n", len(c))
			return c, nil
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
