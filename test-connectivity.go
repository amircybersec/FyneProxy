// Copyright 2023 Jigsaw Operations LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Jigsaw-Code/outline-sdk/dns"
	"github.com/Jigsaw-Code/outline-sdk/x/config"
	"github.com/Jigsaw-Code/outline-sdk/x/connectivity"
	"github.com/Jigsaw-Code/outline-sdk/x/report"
)

var debugLog log.Logger = *log.New(io.Discard, "", 0)

// var errorLog log.Logger = *log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

type connectivityReport struct {
	// Inputs
	Resolver  string `json:"resolver"`
	Proto     string `json:"proto"`
	Transport string `json:"transport"`

	// Observations
	Time       time.Time  `json:"time"`
	DurationMs int64      `json:"duration_ms"`
	Error      *errorJSON `json:"error"`
	Collected  bool       `json:"collected"`
}

type errorJSON struct {
	Op         string `json:"op,omitempty"`
	PosixError string `json:"posix_error,omitempty"`
	Msg        string `json:"msg,omitempty"`
}

func makeErrorRecord(result *connectivity.ConnectivityError) *errorJSON {
	if result == nil {
		return nil
	}
	var record = new(errorJSON)
	record.Op = result.Op
	record.PosixError = result.PosixError
	record.Msg = unwrapAll(result.Err).Error()
	return record
}

func unwrapAll(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

func (r connectivityReport) IsSuccess() bool {
	if r.Error == nil {
		return true
	} else {
		return false
	}
}

func TestConfigs(setting *AppSettings) {
	var wg sync.WaitGroup // Step 1: Create a WaitGroup instance

	for i := range setting.Configs {
		wg.Add(1)        // Increment the WaitGroup counter
		go func(i int) { // Step 2: Launch a goroutine
			defer wg.Done() // Step 3: Decrement the counter when the goroutine completes
			TestSingleConfig(setting, i)
		}(i)
	}

	wg.Wait() // Step 4: Wait for all goroutines to complete
}

func TestSingleConfig(setting *AppSettings, i int) {
	var wg sync.WaitGroup
	var healthlyMutex sync.Mutex
	var healthly []bool
	protocols := []string{"tcp", "udp"}
	// check if i is within the range of the slice
	if i < 0 || i >= len(setting.Configs) {
		log.Fatalf("Index %v is out of range", i)
	}
	cnf := &setting.Configs[i]
	c, err := config.SanitizeConfig(cnf.Transport)
	if err != nil {
		log.Fatalf("Failed to sanitize config: %v", err)
	}
	// Clear previous test reports
	// maybe make it atomic to prevent losing previous reports if test fails for any reason
	// In other words, only clear the reports if the test is fully perfomed....
	cnf.TestReports = []*connectivityReport{}
	resolverHost := strings.TrimSpace(setting.ResolverHost)
	resolverAddress := net.JoinHostPort(resolverHost, "53")
	for _, proto := range protocols {
		wg.Add(1)
		go func(proto string, resolverAddress string) {
			defer wg.Done()
			var r connectivityReport
			var resolver dns.Resolver
			r.Transport = c
			r.Resolver = resolverAddress
			startTime := time.Now()
			switch proto {
			case "tcp":
				streamDialer, err := config.NewStreamDialer(cnf.Transport)
				r.Proto = "tcp"
				log.Printf("testing for protocol: %v", r.Proto)
				if err != nil {
					log.Printf("Failed to create StreamDialer: %v", err)
					r.Time = startTime.UTC().Truncate(time.Second)
					r.DurationMs = time.Duration(0).Milliseconds()
					r.Error = &errorJSON{Msg: err.Error()}
					cnf.TestReports = append(cnf.TestReports, &r)
					return
				}
				resolver = dns.NewTCPResolver(streamDialer, resolverAddress)
			case "udp":
				packetDialer, err := config.NewPacketDialer(cnf.Transport)
				r.Proto = "udp"
				log.Printf("testing for protocol: %v", r.Proto)
				if err != nil {
					log.Printf("Failed to create StreamDialer: %v", err)
					r.Time = startTime.UTC().Truncate(time.Second)
					r.DurationMs = time.Duration(0).Milliseconds()
					r.Error = &errorJSON{Msg: err.Error()}
					cnf.TestReports = append(cnf.TestReports, &r)
					return
				}
				resolver = dns.NewUDPResolver(packetDialer, resolverAddress)
			default:
				log.Fatalf(`Invalid proto %v. Must be "tcp" or "udp"`, proto)
			}
			result, err := connectivity.TestConnectivityWithResolver(context.Background(), resolver, setting.Domain)
			r.Time = startTime.UTC().Truncate(time.Second)
			r.DurationMs = time.Duration(0).Milliseconds()
			if err != nil {
				log.Fatalf("Connectivity test failed to run: %v", err)
				r.Error = &errorJSON{Msg: err.Error()}
				cnf.TestReports = append(cnf.TestReports, &r)
				return
			}
			r.Error = makeErrorRecord(result)
			//log.Printf("Connectivity test result: %v", r)
			// collectReport(r, "")
			cnf.TestReports = append(cnf.TestReports, &r)

			healthlyMutex.Lock()
			healthly = append(healthly, r.IsSuccess())
			healthlyMutex.Unlock()

		}(proto, resolverAddress)
	}
	wg.Wait()
	cnf.Health = CheckHealth(healthly)
}

func submitReports(setting *AppSettings) {
	log.Println("Submitting reports...")
	reporterURL := setting.ReporterURL
	log.Printf("Reporter URL: %v", reporterURL)

	var wg sync.WaitGroup // Create a WaitGroup instance

	for i := range setting.Configs {
		c := setting.Configs[i]
		log.Printf("Config: %v", c)
		for j := range c.TestReports {
			wg.Add(1)                        // Increment the WaitGroup counter
			go func(r *connectivityReport) { // Launch a goroutine
				defer wg.Done() // Decrement the counter when the goroutine completes
				err := collectReport(r, reporterURL)
				if err != nil {
					debugLog.Printf("Failed to collect report: %v\n", err)
					r.Collected = false
					return
				}
				log.Println("Report collected successfully")
				r.Collected = true
				log.Printf("Collecting report: %v", r)
			}(c.TestReports[j])
		}
	}

	wg.Wait() // Wait for all goroutines to complete
}

func collectReport(r report.Report, reporterURL string) error {
	var reportCollector report.Collector
	if strings.TrimSpace(reporterURL) != "" {
		log.Println("URL is not empty, using remote collector...")
		collectorURL, err := url.Parse(reporterURL)
		if err != nil {
			debugLog.Printf("Failed to parse collector URL: %v", err)
		}
		remoteCollector := &report.RemoteCollector{
			CollectorURL: collectorURL,
			HttpClient:   &http.Client{Timeout: 10 * time.Second},
		}
		retryCollector := &report.RetryCollector{
			Collector:    remoteCollector,
			MaxRetry:     3,
			InitialDelay: 1 * time.Second,
		}
		reportCollector = &report.SamplingCollector{
			Collector:       retryCollector,
			SuccessFraction: 1,
			FailureFraction: 1,
		}
	} else {
		log.Println("No collector URL provided, writing to stdout")
		reportCollector = &report.WriteCollector{Writer: os.Stdout}
	}
	log.Println("Collecting report...")
	err := reportCollector.Collect(context.Background(), r)
	if err != nil {
		return err
	}
	return nil
}

// CheckHealth takes a slice of booleans and returns:
// 1 if all elements are true (all tests have passed),
// 3 if all elements are false (all tests have failed),
// 2 if there is a mix of true and false (some have failed).
func CheckHealth(slice []bool) int {
	if len(slice) == 0 {
		return 0
	}
	allTrue := true
	allFalse := true
	for _, value := range slice {
		if value {
			allFalse = false
		} else {
			allTrue = false
		}
		// Early exit if we already know there's a mix
		if !allTrue && !allFalse {
			return 2
		}
	}

	if allTrue {
		return 1
	}
	return 3
}
