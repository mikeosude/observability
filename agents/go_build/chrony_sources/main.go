package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	version   = "1.0.0"
	signature = "Designed by Ifesinachi Osude"
)

type ChronySource struct {
	Mode       string
	Name       string
	Stratum    int
	Poll       int
	Reach      int
	LastRx     int
	Offset     float64
	Jitter     float64
	Selected   bool
	InUse      bool
	Reachable  bool
}

func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func unitToSeconds(value string) float64 {
	// Parse values like "-120us", "1.2ms", "0.0003s"
	re := regexp.MustCompile(`^([+-]?[\d.]+)([a-z]*)$`)
	matches := re.FindStringSubmatch(value)
	
	if len(matches) < 2 {
		return 0.0
	}

	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.0
	}

	unit := ""
	if len(matches) >= 3 {
		unit = matches[2]
	}

	switch unit {
	case "ns":
		return num / 1000000000.0
	case "us":
		return num / 1000000.0
	case "ms":
		return num / 1000.0
	default: // "s" or empty
		return num
	}
}

func parseChronySource(line string) (*ChronySource, error) {
	// Parse chrony sources output line
	// Format: MS Name/IP Stratum Poll Reach LastRx [Last sample]
	// Example: ^* 10.0.0.1  2  10  377   32  -120us[ +30us] +/- 500us
	
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return nil, fmt.Errorf("insufficient fields")
	}

	src := &ChronySource{
		Mode:      fields[0],
		Name:      fields[1],
		Selected:  strings.Contains(fields[0], "*"),
		InUse:     strings.Contains(fields[0], "+"),
	}

	// Parse numeric fields
	stratum, _ := strconv.Atoi(fields[2])
	src.Stratum = stratum

	poll, _ := strconv.Atoi(fields[3])
	src.Poll = poll

	// Reach might be octal
	reach := fields[4]
	if reachInt, err := strconv.ParseInt(reach, 8, 64); err == nil {
		src.Reach = int(reachInt)
	} else if reachInt, err := strconv.Atoi(reach); err == nil {
		src.Reach = reachInt
	}

	src.Reachable = src.Reach > 0

	lastRx, _ := strconv.Atoi(fields[5])
	src.LastRx = lastRx

	// Parse offset (field 6, might have brackets)
	offsetRaw := fields[6]
	offsetRaw = strings.Trim(offsetRaw, "[]")
	src.Offset = unitToSeconds(offsetRaw)

	// Parse jitter (last field, might have +/-)
	jitterRaw := fields[len(fields)-1]
	jitterRaw = strings.TrimPrefix(jitterRaw, "+/-")
	jitterRaw = strings.TrimSpace(jitterRaw)
	src.Jitter = unitToSeconds(jitterRaw)

	return src, nil
}

func getChronySources() ([]*ChronySource, error) {
	// Check if chronyc is available
	if _, err := exec.LookPath("chronyc"); err != nil {
		return nil, fmt.Errorf("chronyc not found")
	}

	// Run chronyc sources -v
	cmd := exec.Command("chronyc", "-n", "sources", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var sources []*ChronySource
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Look for lines starting with source indicators
		if len(line) > 0 && (strings.HasPrefix(line, "^") || strings.HasPrefix(line, "=")) {
			src, err := parseChronySource(line)
			if err == nil {
				sources = append(sources, src)
			}
		}
	}

	return sources, nil
}

func main() {
	sources, err := getChronySources()
	
	fmt.Println("# TYPE chrony_sources_up gauge")
	if err != nil {
		fmt.Println("chrony_sources_up 0")
		fmt.Printf("# VERSION %s\n", version)
		fmt.Printf("# %s\n", signature)
		return
	}
	fmt.Println("chrony_sources_up 1")

	// Output metrics
	fmt.Println("# HELP chrony_source_selected 1 if this source is selected (*), else 0")
	fmt.Println("# TYPE chrony_source_selected gauge")
	fmt.Println("# HELP chrony_source_in_use 1 if this source is being combined/used (+), else 0")
	fmt.Println("# TYPE chrony_source_in_use gauge")
	fmt.Println("# HELP chrony_source_reachable 1 if reachable (reach>0), else 0")
	fmt.Println("# TYPE chrony_source_reachable gauge")
	fmt.Println("# HELP chrony_source_reach Reach register value")
	fmt.Println("# TYPE chrony_source_reach gauge")
	fmt.Println("# HELP chrony_source_last_rx_seconds Seconds since last sample")
	fmt.Println("# TYPE chrony_source_last_rx_seconds gauge")
	fmt.Println("# HELP chrony_source_offset_seconds Reported offset in seconds")
	fmt.Println("# TYPE chrony_source_offset_seconds gauge")
	fmt.Println("# HELP chrony_source_jitter_seconds Reported jitter in seconds")
	fmt.Println("# TYPE chrony_source_jitter_seconds gauge")

	for _, src := range sources {
		nameEsc := escapeLabel(src.Name)
		modeEsc := escapeLabel(src.Mode)
		stratumStr := strconv.Itoa(src.Stratum)
		pollStr := strconv.Itoa(src.Poll)

		labels := fmt.Sprintf("source=\"%s\",mode=\"%s\",stratum=\"%s\",poll=\"%s\"",
			nameEsc, modeEsc, stratumStr, pollStr)

		sel := 0
		if src.Selected {
			sel = 1
		}
		use := 0
		if src.InUse {
			use = 1
		}
		reachable := 0
		if src.Reachable {
			reachable = 1
		}

		fmt.Printf("chrony_source_selected{%s} %d\n", labels, sel)
		fmt.Printf("chrony_source_in_use{%s} %d\n", labels, use)
		fmt.Printf("chrony_source_reachable{%s} %d\n", labels, reachable)
		fmt.Printf("chrony_source_reach{%s} %d\n", labels, src.Reach)
		fmt.Printf("chrony_source_last_rx_seconds{%s} %d\n", labels, src.LastRx)
		fmt.Printf("chrony_source_offset_seconds{%s} %.9f\n", labels, src.Offset)
		fmt.Printf("chrony_source_jitter_seconds{%s} %.9f\n", labels, src.Jitter)
	}

	fmt.Printf("# VERSION %s\n", version)
	fmt.Printf("# %s\n", signature)
}
