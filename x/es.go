package main

import (
	xFlag "flag"
	"fmt"
)

func es(args []string) {
	flag := xFlag.NewFlagSet("es", xFlag.ExitOnError)

	index := flag.String("index", "proxify", "es index")
	esUrl := flag.String("esURL", "http://localhost:9200", "es url")
	q := flag.String("q", "*", "es search query")
	s := flag.String("source", "host", "what field you want to get")
	num := flag.Int("num", 10000, "how many document you want to get")
	of := flag.String("of", "", "output file path")

	flag.Parse(args)

	// source
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": q,
			},
		},
		"collapse": map[string]interface{}{
			"field": *s + ".keyword",
		},
		"size": num,
	}

	result := newClient(*esUrl, *index).getDocument(query, *s)

	for _, v := range result {
		fmt.Println(v)
	}
	if *of != "" {
		writeFile(*of, result)
	}
}
