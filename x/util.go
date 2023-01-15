package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"strings"
)

type client struct {
	esClient *elasticsearch.Client
	index    string
}

func newClient(es, index string) *client {
	cfg := elasticsearch.Config{
		Addresses: []string{es},
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	return &client{esClient: esClient, index: index}
}

func (c *client) getDocument(query map[string]interface{}, source string) []string {
	var buf bytes.Buffer
	var r map[string]interface{}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	// Perform the search request.
	es := c.esClient
	res, err := es.Search(
		es.Search.WithSource(source),
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(c.index),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil || res == nil {
		fmt.Printf("Error getting response: %s\n", err)
		return nil
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	//Print the response status, number of results, and request duration.
	//log.Printf(
	//	"[%s] %d hits; took: %dms",
	//	res.Status(),
	//	int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
	//	int(r["took"].(float64)),
	//)

	// get field
	count := 0
	result := make([]string, 0)
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		field := hit.(map[string]interface{})["_source"].(map[string]interface{})[source]
		switch field.(type) {
		case string:
			//println(document.(string))
			result = append(result, field.(string))
		}
		count++
	}
	fmt.Printf("total: %d\n", count)
	return result
}

func (c *client) getDocumentMultiField(query map[string]interface{}, source string) map[string][]string {
	var buf bytes.Buffer
	var r map[string]interface{}

	//query := map[string]interface{}{
	//	//"query": map[string]interface{}{
	//	//	"match": map[string]interface{}{
	//	//		"request": "163",
	//	//	},
	//	//},
	//	"collapse": map[string]interface{}{
	//		"field": "host.keyword",
	//	},
	//	"size": 10000,
	//}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
	// Perform the search request.
	es := c.esClient
	res, err := es.Search(
		es.Search.WithSource(source),
		//es.Search.WithQuery(makeKeywordQuery()),
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(c.index),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil || res == nil {
		fmt.Printf("Error getting response: %s\n", err)
		return nil
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)
	// Print the ID and document source for each hit.
	count := 0
	result := make(map[string][]string, 0)

	keyList := strings.Split(source, ",")
	for _, key := range keyList {
		result[key] = make([]string, 0)
	}

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		document := hit.(map[string]interface{})["_source"].(map[string]interface{})
		for _, key := range keyList {
			value := document[key]
			switch value.(type) {
			case string:
				//println(document.(string))
				result[key] = append(result[key], value.(string))
			}
		}

		//log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		count++
	}
	fmt.Printf("total: %d\n", count)
	return result
}
