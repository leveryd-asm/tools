package main

import (
	"bufio"
	xFlag "flag"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type detectWay struct {
	detectWay1 bool
	detectWay2 bool
	detectWay3 bool
}

// this file is used to identify the manage system
var resultUrlList = make([]string, 0)

func identifyMS(args []string) {
	flag := xFlag.NewFlagSet("identifyMS", xFlag.ExitOnError)

	target := flag.String("u", "", "target url")
	targetFilePath := flag.String("if", "", "target url file path")
	concurrency := flag.Int("t", 10, "concurrency num when target is url file")
	outputFilePath := flag.String("of", "", "result output file path")

	detectWay1 := flag.Bool("whk", true, "detect way host keyword")
	detectWay2 := flag.Bool("wrk", true, "detect way response keyword")
	detectWay3 := flag.Bool("wre", true, "detect way response element")

	flag.Parse(args)

	detectWay := detectWay{
		detectWay1: *detectWay1,
		detectWay2: *detectWay2,
		detectWay3: *detectWay3,
	}

	if target == nil && targetFilePath == nil {
		flag.Usage()
		return
	}

	if *target != "" {
		identifyManageSystem(*target, detectWay)
	} else {
		identifyManageSystemFromFile(*targetFilePath, *concurrency, detectWay)
	}

	if *outputFilePath != "" {
		writeFile(*outputFilePath, resultUrlList)
	}
}

func identifyManageSystem(url string, detectWay detectWay) {
	//println("DEBUG", url)
	if detectWay.detectWay1 {
		host := getHost(url)
		if isHostContainMSKeyword(host) {
			println(url)
			resultUrlList = append(resultUrlList, url)
		}
	}

	if detectWay.detectWay2 || detectWay.detectWay3 {
		if !strings.HasPrefix(url, "http") {
			url = "https://" + url
		}
		res, err := http.Get(url)
		if err != nil || res == nil {
			return
		}

		defer res.Body.Close()

		body := make([]byte, 1024000) // 100kB is enough
		res.Body.Read(body)

		ct := res.Header.Get("Content-Type")
		if strings.Contains(ct, "text/html") {
			if detectWay.detectWay2 && isBodyContainMSKeyword(string(body)) {
				println(url)
				resultUrlList = append(resultUrlList, url)
			}
			if detectWay.detectWay3 && isVue(string(body)) {
				println(url)
				resultUrlList = append(resultUrlList, url)
			}
		}
	}
}

func identifyManageSystemFromFile(filePath string, concurrency int, detectWay detectWay) {
	f, err := os.Open(filePath)
	if err != nil {
		println("file does not exist")
		return
	}
	defer f.Close()

	fileScanner := bufio.NewScanner(f)

	c := make(chan string)
	var wg sync.WaitGroup

	// consumer
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			for {
				targetURL := <-c
				if targetURL == "" {
					break
				}
				identifyManageSystem(targetURL, detectWay)
			}
			wg.Done()
		}()
	}

	// producer
	for fileScanner.Scan() {
		c <- fileScanner.Text()
	}
	close(c)

	// wait
	wg.Wait()
}

func getHost(targetURL string) string {
	parse, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}
	return parse.Host
}

func isBodyContainMSKeyword(body string) bool {
	if strings.Contains(body, "<table") {
		return true
	}
	return false
}
