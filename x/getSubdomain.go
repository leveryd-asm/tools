package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// get subdomain from http request host, the subdomain ends with the domain
// filter out null result
func getSubdomain(es, index, domain, outputFilePath string) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"host": domain,
			},
		},
		"collapse": map[string]interface{}{
			"field": "host.keyword",
		},
		"size": 10000,
	}
	domains := newClient(es, index).getDocument(query, "host")

	filterDomains := make([]string, 0)
	for _, i := range domains {
		if strings.HasSuffix(i, "."+domain) {
			filterDomains = append(filterDomains, i)
		}
	}

	if outputFilePath == "" {
		for _, domain := range filterDomains {
			fmt.Println(domain)
		}
	} else {
		writeFile(outputFilePath, filterDomains)
	}
}

func writeFile(path string, domains []string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Error creating file: %s", err)
	}
	defer file.Close()

	for _, domain := range domains {
		_, err := file.WriteString(domain + "\n")
		if err != nil {
			log.Fatalf("Error writing string: %s", err)
		}
	}
}
