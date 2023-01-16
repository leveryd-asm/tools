package main

import (
	"bufio"
	xFlag "flag"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type detectWayConfig struct {
	detectWay1 bool
	detectWay2 bool
	detectWay3 bool
}

type screenshotConfig struct {
	screenshotServiceApi string
}

var shotConfigIns screenshotConfig
var detectWayConfigIns detectWayConfig

// this file is used to identify the manage system
var resultUrlList = make([]string, 0)
var resultUrlScreenshotMap = make(map[string]string, 0)

func identifyMS(args []string) {
	flag := xFlag.NewFlagSet("identifyMS", xFlag.ExitOnError)

	target := flag.String("u", "", "target url")
	targetFilePath := flag.String("if", "", "target url file path")
	concurrency := flag.Int("t", 10, "concurrency num when target is url file")
	outputFilePath := flag.String("of", "", "result output file path")
	debug := flag.Bool("debug", false, "debug mode")

	detectWay1 := flag.Bool("whk", true, "detect way host keyword")
	detectWay2 := flag.Bool("wrk", false, "detect way response keyword")
	detectWay3 := flag.Bool("wre", true, "detect way response element")

	flag.Parse(args)

	detectWayConfigIns = detectWayConfig{
		detectWay1: *detectWay1,
		detectWay2: *detectWay2,
		detectWay3: *detectWay3,
	}

	if target == nil && targetFilePath == nil {
		flag.Usage()
		return
	}

	if debug != nil && *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *target != "" {
		identifyManageSystem(*target)
	} else {
		identifyManageSystemFromFile(*targetFilePath, *concurrency)
	}

	if *outputFilePath != "" {
		writeFile(*outputFilePath, resultUrlList)
	}
}

func identifyManageSystem(url string) {
	log.Debug("identify: ", url)
	if detectWayConfigIns.detectWay1 {
		host := getHost(url)
		if isHostContainMSKeyword(host) {
			println(url)
			resultUrlList = append(resultUrlList, url)
		}
	}

	if detectWayConfigIns.detectWay2 || detectWayConfigIns.detectWay3 {
		if !strings.HasPrefix(url, "http") {
			url = "https://" + url
		}
		res, err := http.Get(url)
		if err != nil || res == nil {
			log.Warn("get url error: ", url)
			return
		}

		defer res.Body.Close()

		body := make([]byte, 1024000) // 100kB is enough
		res.Body.Read(body)

		ct := res.Header.Get("Content-Type")
		if strings.Contains(ct, "text/html") {
			if detectWayConfigIns.detectWay2 && isBodyContainMSKeyword(string(body)) {
				afterFindMS(url)
			} else if detectWayConfigIns.detectWay3 && isVue(string(body)) {
				afterFindMS(url)
			}
		}
	}
}

func afterFindMS(msURL string) {
	println(msURL)
	resultUrlList = append(resultUrlList, msURL)
	if shotConfigIns.screenshotServiceApi != "" {
		resultUrlScreenshotMap[msURL] = shot(msURL)
	}
}

func identifyManageSystemFromFile(filePath string, concurrency int) {
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
				identifyManageSystem(targetURL)
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
