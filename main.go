package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func main() {
	filePath := flag.String("f", "", "path to file with domains (one per line)")
	concurrency := flag.Int("c", 5, "max concurrent checks")
	availableOnly := flag.Bool("available-only", false, "only print available domains")
	tlds := flag.String("tlds", "", "comma-separated TLDs to expand base names (e.g. com,net,io)")
	jsonOutput := flag.Bool("json", false, "output results in JSON format (one JSON object per line)")
	noColor := flag.Bool("no-color", false, "disable colored output")
	flag.Parse()

	if *noColor {
		color.NoColor = true
	}

	var domains []string

	// Collect from file
	if *filePath != "" {
		fileDomains, err := readDomainsFromFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(2)
		}
		domains = append(domains, fileDomains...)
	}

	// Collect from positional args
	args := flag.Args()
	if *tlds != "" {
		tldList := strings.Split(*tlds, ",")
		for _, base := range args {
			for _, tld := range tldList {
				domains = append(domains, base+"."+strings.TrimSpace(tld))
			}
		}
	} else {
		domains = append(domains, args...)
	}

	// Collect from stdin if piped
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		stdinDomains, err := readDomainsFromReader(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(2)
		}
		domains = append(domains, stdinDomains...)
	}

	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "No domains provided. Usage: domain-chooser [flags] domain1 domain2 ...")
		flag.PrintDefaults()
		os.Exit(2)
	}

	// Validate domains
	for _, d := range domains {
		if !strings.Contains(d, ".") {
			fmt.Fprintf(os.Stderr, "Invalid domain %q: must contain a dot (use -tlds to expand base names)\n", d)
			os.Exit(2)
		}
	}

	results := CheckDomains(domains, *concurrency)

	var availCount int
	if *jsonOutput {
		availCount = printJSON(results, *availableOnly)
	} else {
		availCount = printText(results, *availableOnly)
	}
	if availCount == 0 {
		os.Exit(1)
	}
}

type jsonResult struct {
	Domain    string `json:"domain"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
	CheckedAt string `json:"checked_at"`
}

func printJSON(results []Result, availableOnly bool) int {
	encoder := json.NewEncoder(os.Stdout)
	var availCount int
	for _, r := range results {
		if availableOnly && !r.Available {
			continue
		}
		jr := jsonResult{
			Domain:    r.Domain,
			Available: r.Available,
			CheckedAt: r.CheckedAt.Format(time.RFC3339),
		}
		if r.Err != nil {
			jr.Error = r.Err.Error()
		}
		if err := encoder.Encode(jr); err != nil {
			fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
			os.Exit(2)
		}
		if r.Available {
			availCount++
		}
	}
	return availCount
}

func printText(results []Result, availableOnly bool) int {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	var availCount, takenCount, errCount int
	for _, r := range results {
		switch {
		case r.Err != nil:
			errCount++
			if !availableOnly {
				fmt.Printf("  %s  %s: %v\n", yellow("ERROR"), r.Domain, r.Err)
			}
		case r.Available:
			availCount++
			fmt.Printf("  %s  %s\n", green("AVAIL"), r.Domain)
		default:
			takenCount++
			if !availableOnly {
				fmt.Printf("  %s  %s\n", red("TAKEN"), r.Domain)
			}
		}
	}

	total := availCount + takenCount + errCount
	fmt.Printf("\nSummary: %d available, %d taken, %d error (of %d checked)\n",
		availCount, takenCount, errCount, total)

	return availCount
}

func readDomainsFromFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readDomainsFromReader(f)
}

func readDomainsFromReader(r *os.File) ([]string, error) {
	var domains []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			domains = append(domains, line)
		}
	}
	return domains, scanner.Err()
}
