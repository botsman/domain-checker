package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/likexian/whois"
)

type Config struct {
	Keywords     [][]string
	Combinations int
	TLDs         []string
	Separator    string
}

func main() {
	// Define CLI flags
	keywords := flag.String("keywords", "", "Comma-separated keywords (e.g., 'one,two,three')")
	keywordLists := flag.String("lists", "", "Semicolon-separated lists of keywords (e.g., 'one,two;three,four')")
	combinations := flag.Int("combinations", 2, "Number of keywords to combine (ignored if lists provided)")
	tlds := flag.String("tlds", "com", "Comma-separated TLDs to check (e.g., 'com,net,org')")
	useDash := flag.Bool("dash", false, "Use dash separator (e.g., 'one-two' instead of 'onetwo')")
	workers := flag.Int("workers", 10, "Number of concurrent workers")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Domain Checker - Check domain availability\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Check 2-word combinations from a single list\n")
		fmt.Fprintf(os.Stderr, "  %s -keywords=super,fast,cloud\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Check combinations between two lists\n")
		fmt.Fprintf(os.Stderr, "  %s -lists=\"super,fast;cloud,service\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use dash separator and check multiple TLDs\n")
		fmt.Fprintf(os.Stderr, "  %s -keywords=my,app -dash -tlds=com,net,org\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Check 3-word combinations\n")
		fmt.Fprintf(os.Stderr, "  %s -keywords=get,my,app,now -combinations=3\n\n", os.Args[0])
	}

	flag.Parse()

	// Validate input
	if *keywords == "" && *keywordLists == "" {
		fmt.Fprintf(os.Stderr, "Error: Either -keywords or -lists must be provided\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *keywords != "" && *keywordLists != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both -keywords and -lists at the same time\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse configuration
	config := Config{
		Combinations: *combinations,
		TLDs:         parseTLDs(*tlds),
		Separator:    "",
	}

	if *useDash {
		config.Separator = "-"
	}

	// Parse keywords
	if *keywordLists != "" {
		config.Keywords = parseKeywordLists(*keywordLists)
	} else {
		config.Keywords = [][]string{parseKeywords(*keywords)}
	}

	// Generate domain combinations
	domains := generateDomains(config)

	if len(domains) == 0 {
		fmt.Println("No domains to check")
		os.Exit(0)
	}

	fmt.Printf("Checking %d domains...\n\n", len(domains))

	// Check domains concurrently
	results := checkDomainsConcurrently(domains, *workers)

	// Print results
	printResults(results)
}

func parseKeywords(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return []string{}
	}
	keywords := strings.Split(input, ",")
	for i := range keywords {
		keywords[i] = strings.TrimSpace(keywords[i])
	}
	return keywords
}

func parseKeywordLists(input string) [][]string {
	lists := strings.Split(input, ";")
	result := make([][]string, 0, len(lists))
	for _, list := range lists {
		keywords := parseKeywords(list)
		if len(keywords) > 0 {
			result = append(result, keywords)
		}
	}
	return result
}

func parseTLDs(input string) []string {
	tlds := parseKeywords(input)
	for i := range tlds {
		// Remove leading dot if present
		tlds[i] = strings.TrimPrefix(tlds[i], ".")
	}
	return tlds
}

func generateDomains(config Config) []string {
	var domains []string

	if len(config.Keywords) == 1 {
		// Single list mode - generate combinations
		combinations := generateCombinations(config.Keywords[0], config.Combinations)
		for _, combo := range combinations {
			domainName := strings.Join(combo, config.Separator)
			for _, tld := range config.TLDs {
				domains = append(domains, fmt.Sprintf("%s.%s", domainName, tld))
			}
		}
	} else {
		// Multiple lists mode - cross product
		crossProducts := crossProduct(config.Keywords)
		for _, product := range crossProducts {
			domainName := strings.Join(product, config.Separator)
			for _, tld := range config.TLDs {
				domains = append(domains, fmt.Sprintf("%s.%s", domainName, tld))
			}
		}
	}

	return domains
}

func generateCombinations(keywords []string, n int) [][]string {
	if n <= 0 || n > len(keywords) {
		return [][]string{}
	}

	var result [][]string
	var current []string

	var backtrack func(start int)
	backtrack = func(start int) {
		if len(current) == n {
			combo := make([]string, len(current))
			copy(combo, current)
			result = append(result, combo)
			return
		}

		for i := start; i < len(keywords); i++ {
			current = append(current, keywords[i])
			backtrack(i + 1)
			current = current[:len(current)-1]
		}
	}

	backtrack(0)
	return result
}

func crossProduct(lists [][]string) [][]string {
	if len(lists) == 0 {
		return [][]string{}
	}

	if len(lists) == 1 {
		result := make([][]string, len(lists[0]))
		for i, item := range lists[0] {
			result[i] = []string{item}
		}
		return result
	}

	// Recursive cross product
	subProduct := crossProduct(lists[1:])
	var result [][]string

	for _, item := range lists[0] {
		for _, sub := range subProduct {
			product := make([]string, 0, len(sub)+1)
			product = append(product, item)
			product = append(product, sub...)
			result = append(result, product)
		}
	}

	return result
}

type DomainResult struct {
	Domain    string
	Available bool
	Error     error
}

func checkDomainsConcurrently(domains []string, workers int) []DomainResult {
	jobs := make(chan string, len(domains))
	results := make(chan DomainResult, len(domains))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range jobs {
				available, err := checkDomain(domain)
				results <- DomainResult{
					Domain:    domain,
					Available: available,
					Error:     err,
				}
			}
		}()
	}

	// Send jobs
	for _, domain := range domains {
		jobs <- domain
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allResults []DomainResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

func checkDomain(domain string) (bool, error) {
	result, err := whois.Whois(domain)
	if err != nil {
		return false, err
	}

	// Simple heuristic: if the result contains these keywords, domain is likely taken
	result = strings.ToLower(result)
	
	// Check for common "domain available" indicators
	if strings.Contains(result, "no match") ||
		strings.Contains(result, "not found") ||
		strings.Contains(result, "no entries found") ||
		strings.Contains(result, "no data found") ||
		strings.Contains(result, "available for registration") ||
		strings.Contains(result, "status: free") {
		return true, nil
	}

	// Check for common "domain taken" indicators
	if strings.Contains(result, "domain name:") ||
		strings.Contains(result, "registrar:") ||
		strings.Contains(result, "creation date:") ||
		strings.Contains(result, "expiration date:") ||
		strings.Contains(result, "updated date:") {
		return false, nil
	}

	// If we can't determine, assume it's taken (safer assumption)
	return false, nil
}

func printResults(results []DomainResult) {
	available := []string{}
	taken := []string{}
	errors := []DomainResult{}

	for _, result := range results {
		if result.Error != nil {
			errors = append(errors, result)
		} else if result.Available {
			available = append(available, result.Domain)
		} else {
			taken = append(taken, result.Domain)
		}
	}

	// Print available domains
	if len(available) > 0 {
		fmt.Printf("✓ AVAILABLE (%d):\n", len(available))
		for _, domain := range available {
			fmt.Printf("  %s\n", domain)
		}
		fmt.Println()
	}

	// Print taken domains
	if len(taken) > 0 {
		fmt.Printf("✗ TAKEN (%d):\n", len(taken))
		for _, domain := range taken {
			fmt.Printf("  %s\n", domain)
		}
		fmt.Println()
	}

	// Print errors
	if len(errors) > 0 {
		fmt.Printf("⚠ ERRORS (%d):\n", len(errors))
		for _, result := range errors {
			fmt.Printf("  %s: %v\n", result.Domain, result.Error)
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("Summary: %d available, %d taken, %d errors (total: %d)\n",
		len(available), len(taken), len(errors), len(results))
}
