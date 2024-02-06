package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test connectivity with both TCP and UDP protocols
func TestConnectivityWithTCPAndUDPProtocols(t *testing.T) {
    // Set up test data
    setting := &AppSettings{
        Configs: []Config{
            {
                Transport: "ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTpLeTUyN2duU3FEVFB3R0JpQ1RxUnlT@104.238.183.16:65496",
                TestReports: []*connectivityReport{},
            },
            {
                Transport: "ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTpLeTUyN2duU3FEVFB3R0JpQ1RxUnlT@104.238.183.15:65496",
                TestReports: []*connectivityReport{},
            },
        },
        DnsList: "8.8.8.8",
        Domain: "example.com",
    }
    i := 0

    // Run the function under test
    TestSingleConfig(setting, i)

    // Assert the results
    assert.Equal(t, 2, len(setting.Configs[i].TestReports))
    assert.Equal(t, "tcp", setting.Configs[i].TestReports[0].Proto)
    assert.Equal(t, "udp", setting.Configs[i].TestReports[1].Proto)
    assert.Equal(t, true, setting.Configs[i].TestReports[0].IsSuccess())
    assert.Equal(t, true, setting.Configs[i].TestReports[1].IsSuccess())
}

func TestCollectReport(t *testing.T) {
    r := connectivityReport{Resolver: "8.8.8.8", Proto: "tcp", Transport: "testss://", Error: nil, Collected: false}
    u := "https://script.google.com/macros/s/AKfycbzoMBmftQaR9Aw4jzTB-w4TwkDjLHtSfBCFhh4_2NhTEZAUdj85Qt8uYCKCNOEAwCg4/exec"
    collectReport(&r, u)
}

tls:10