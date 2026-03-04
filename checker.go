package main

import (
	"errors"
	"sync"

	"github.com/openrdap/rdap"
)

type Result struct {
	Domain    string
	Available bool
	Err       error
}

func CheckDomain(domain string) Result {
	client := &rdap.Client{}
	_, err := client.QueryDomain(domain)
	if err != nil {
		var clientErr *rdap.ClientError
		if errors.As(err, &clientErr) && clientErr.Type == rdap.ObjectDoesNotExist {
			return Result{Domain: domain, Available: true}
		}
		return Result{Domain: domain, Err: err}
	}
	return Result{Domain: domain, Available: false}
}

func CheckDomains(domains []string, concurrency int) []Result {
	results := make([]Result, len(domains))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, domain := range domains {
		wg.Add(1)
		go func(i int, domain string) {
			defer wg.Done()
			sem <- struct{}{}
			results[i] = CheckDomain(domain)
			<-sem
		}(i, domain)
	}

	wg.Wait()
	return results
}
