package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// FormatA struct for the first JSON format
type ServerInfo struct {
	ID         string `json:"id,omitempty"`
	Remarks    string `json:"remarks,omitempty"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
	Prefix     string `json:"prefix"`
	Plugin     string `json:"plugin,omitempty"`
	PluginOpts string `json:"plugin_opts,omitempty"`
}

// FormatC struct for the SIP008 JSON format
type SIP008Config struct {
	Version         int                    `json:"version"`
	Servers         []ServerInfo           `json:"servers"`
	BytesUsed       uint64                 `json:"bytes_used,omitempty"`
	BytesRemaining  uint64                 `json:"bytes_remaining,omitempty"`
	AdditionalProps map[string]interface{} // For custom fields
}

func parseDynamicConfig(data []byte) ([]string, error) {
	//Parse if simple JSON format
	server, err := parseSingleJSON(data)
	if err == nil {
		return []string{server}, nil
	}
	// Parse if SIP008 JSON format
	servers, err := parseSIP008(data)
	if err == nil {
		return servers, nil
	} else {
		fmt.Println("parseSIP008 error:", err)
	}
	// Parse if CSV format
	servers, err = parseBase64URLLine(data)
	if err == nil {
		return servers, nil
	} else {
		fmt.Println("parseBase64URLLine error:", err)
	}
	servers, err = parseCSVformat(data)
	if err == nil {
		return servers, nil
	} else {
		fmt.Println("parseCSVformat error:", err)
	}
	return []string{}, fmt.Errorf("unknown format")
	// parse
}

func getDynamicConfig(url string) ([]string, error) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return []string{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return []string{}, err
	}

	conf, err := parseDynamicConfig(body)
	if err != nil {
		fmt.Println("Error detecting format:", err)
		return []string{}, err
	}
	return conf, nil
}

func parseSingleJSON(data []byte) (string, error) {
	//Parse if simple JSON format
	var config ServerInfo
	err := json.Unmarshal(data, &config)
	if err != nil {
		return "", err
	}
	return makeShadowsocksURLfromJSON(&config)
}

func parseSIP008(data []byte) ([]string, error) {
	//Parse if SIP008 JSON format
	var config SIP008Config
	err := json.Unmarshal(data, &config)
	if err != nil {
		return []string{}, err
	}
	if config.Version == 1 {
		var result []string
		for _, server := range config.Servers {
			configURL, err := makeShadowsocksURLfromJSON(&server)
			if err != nil {
				return []string{}, err
			}
			result = append(result, configURL)
		}
		return result, nil
	}
	return []string{}, fmt.Errorf("unknown SIP008 version: %d", config.Version)
}

func parseCSVformat(data []byte) ([]string, error) {
	// fmt.Println("Printing response string:")
	str := string(data)
	configs := strings.Split(str, "\n")
	fmt.Println("Printing response string:")
	fmt.Println(configs)
	// check of each line contains a valid URL
	var validConfigs []string
	for _, config := range configs {
		// Ignore blank lines
		if config == "" {
			continue
		}
		u, err := url.Parse(config)
		// ignore invalid URLs
		if err != nil {
			continue
		}
		fmt.Println("scheme:", u.Scheme)
		// ignore non-URLs
		if u.Scheme == "" {
			continue
		}
		validConfigs = append(validConfigs, config)
	}
	return validConfigs, nil
}

// https://www.v2fly.org/en_US/v5/config/service/subscription.html#subscription-container
func parseBase64URLLine(data []byte) ([]string, error) {
	decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(string(data))
	if err != nil {
		return []string{}, err
	}
	return parseCSVformat(decoded)
}

func makeShadowsocksURLfromJSON(config *ServerInfo) (string, error) {
	if config.ServerPort == 0 {
		return "", fmt.Errorf("missing server port")
	}
	if config.Method == "" {
		return "", fmt.Errorf("missing method")
	}
	if config.Password == "" {
		return "", fmt.Errorf("missing password")
	}
	if config.Server == "" {
		return "", fmt.Errorf("missing server")
	}
	configURL := "ss://" + config.Method + ":" + config.Password + "@" + config.Server + ":" + fmt.Sprint(config.ServerPort)
	if config.Prefix != "" {
		configURL += "/?prefix=" + url.QueryEscape(config.Prefix)
	}
	if config.Plugin != "" {
		configURL += "&plugin=" + url.QueryEscape(config.Plugin)
	}
	return configURL, nil
}
