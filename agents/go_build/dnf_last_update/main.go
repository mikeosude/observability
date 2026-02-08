package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	version   = "1.0.0"
	signature = "Designed by Ifesinachi Osude"
)

// escapeLabel escapes Prometheus label values
func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func main() {
	// Get most recent package installation from rpm
	cmd := exec.Command("rpm", "-qa", "--last")
	output, err := cmd.Output()
	
	if err != nil || len(output) == 0 {
		// Error case
		fmt.Println("# TYPE dnf_last_update gauge")
		fmt.Println("dnf_last_update 0")
		fmt.Println("# TYPE dnf_last_update_error gauge")
		fmt.Println("dnf_last_update_error 1")
		fmt.Printf("# VERSION %s\n", version)
		fmt.Printf("# %s\n", signature)
		return
	}

	// Parse first line (most recent package)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	if !scanner.Scan() {
		fmt.Println("# TYPE dnf_last_update gauge")
		fmt.Println("dnf_last_update 0")
		fmt.Println("# TYPE dnf_last_update_error gauge")
		fmt.Println("dnf_last_update_error 1")
		fmt.Printf("# VERSION %s\n", version)
		fmt.Printf("# %s\n", signature)
		return
	}

	line := scanner.Text()
	
	// Extract date portion (everything after first whitespace)
	parts := strings.Fields(line)
	if len(parts) < 2 {
		fmt.Println("# TYPE dnf_last_update gauge")
		fmt.Println("dnf_last_update 0")
		fmt.Println("# TYPE dnf_last_update_error gauge")
		fmt.Println("dnf_last_update_error 1")
		fmt.Printf("# VERSION %s\n", version)
		fmt.Printf("# %s\n", signature)
		return
	}

	// Reconstruct date string (skip package name)
	dateStr := strings.Join(parts[1:], " ")

	// Parse date using various formats
	var epoch int64
	var parseErr error

	// Try common formats
	formats := []string{
		"Mon 02 Jan 2006 03:04:05 PM MST",
		"Mon 2 Jan 2006 03:04:05 PM MST",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			epoch = t.Unix()
			parseErr = nil
			break
		}
		parseErr = err
	}

	errVal := 0
	if parseErr != nil || epoch == 0 {
		epoch = 0
		errVal = 1
	}

	dateStrEsc := escapeLabel(dateStr)

	// Output Prometheus metrics
	fmt.Println("# TYPE dnf_last_update gauge")
	fmt.Printf("dnf_last_update %d\n", epoch)

	fmt.Println("# TYPE dnf_last_update_info gauge")
	fmt.Printf("dnf_last_update_info{time=\"%s\"} 1\n", dateStrEsc)

	fmt.Println("# TYPE dnf_last_update_error gauge")
	fmt.Printf("dnf_last_update_error %d\n", errVal)

	fmt.Printf("# VERSION %s\n", version)
	fmt.Printf("# %s\n", signature)
}
