package main

import (
	"bufio"
	xFlag "flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

type output struct {
	outputType     string
	outputFilePath string
	mutex          sync.Mutex
}

func (o *output) write(url, screenshot string) {
	if o.outputFilePath == "" {
		return
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	var file *os.File
	var err error
	if o.outputType == "html" || o.outputType == "csv" {
		file, err = os.OpenFile(o.outputFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else {
		filename := strings.ReplaceAll(url, "/", "-")
		filename += ".jpeg"
		file, err = os.OpenFile(o.outputFilePath+"/"+filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}

	if o.outputType == "html" {
		html := generateHTML(url, screenshot)
		// write to file
		if _, err := file.WriteString(html); err != nil {
			panic(err)
		}
	} else if o.outputType == "csv" {
		// write to file
		file.Write([]byte(url + " " + screenshot))
	} else if o.outputType == "dir" {
		_, err = file.Write([]byte(screenshot))
		if err != nil {
			panic(err)
		}
	}
}

var screenshotServiceApi string
var outputIns output

func screenshot(args []string) {
	flag := xFlag.NewFlagSet("screenshot", xFlag.ExitOnError)
	targetUrl := flag.String("u", "", "url to shot")
	targetUrlFilePath := flag.String("if", "", "urls file path to shot")

	api := flag.String("sssUrl", "", "screen shot service url")
	concurrency := flag.Int("t", 10, "concurrency num when target is url file")

	// output config
	outputType := flag.String("ot", "html", "output type, support html/csv/dir")
	outputFilepath := flag.String("of", "/tmp/screenshot.html", "output file path")

	flag.Parse(args)
	if *api == "" {
		println("please input screenshot service url")
		flag.Usage()
		return
	} else if *outputFilepath == "" {
		println("please input output file path")
		flag.Usage()
		return
	}

	screenshotServiceApi = *api
	outputIns = output{
		outputType:     *outputType,
		outputFilePath: *outputFilepath,
	}

	if *targetUrl != "" {
		shot(*targetUrl)
	} else if *targetUrlFilePath != "" {
		shotFromFile(*targetUrlFilePath, *concurrency)
	} else {
		flag.Usage()
	}
}

func generateHTML(url, screenshot string) string {
	return "<img src=\"data:image/jpeg;base64," + screenshot + "\" />"
}

// screenshot
func shot(targetURL string) string {
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	// https://www.browserless.io/docs/screenshot
	apiAddress := screenshotServiceApi + "/screenshot"

	var data string
	if outputIns.outputType == "dir" {
		data = "{\"url\":\"" + targetURL + "\",\"options\":{\"fullPage\":true,\"type\":\"jpeg\",\"quality\":20, \"encoding\": \"binary\"}}"
	} else {
		data = "{\"url\":\"" + targetURL + "\",\"options\":{\"fullPage\":true,\"type\":\"jpeg\",\"quality\":20,\"omitBackground\":true,\"encoding\":\"base64\"}}"
	}
	res, err := http.Post(apiAddress, "application/json", strings.NewReader(data))
	if err != nil || res == nil {
		println("shot failed, err: ", err)
		return ""
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}
	outputIns.write(targetURL, string(body))
	return ""
}

func shotFromFile(filePath string, concurrency int) {
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
	count := 0
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			for {
				targetURL := <-c
				if targetURL == "" {
					break
				}
				shot(targetURL)
				count++
				fmt.Printf("shot: %d\r", count) // print progress, it is simple, but it works
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
