package main

import (
	"bufio"
	"crypto/md5"
	xFlag "flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func console(args []string) {
	flag := xFlag.NewFlagSet("console", xFlag.ExitOnError)

	consoleUrl := flag.String("url", "http://console.com:32115", "console service url")
	module := flag.String("module", "ms", "which module you want to use")

	inputFilePath := flag.String("if", "", "input file path")

	flag.Parse(args)

	// source
	f, err := os.Open(*inputFilePath)
	if err != nil {
		return
	}
	defer f.Close()
	fileScanner := bufio.NewScanner(f)

	// https://stackoverflow.com/questions/21124327/how-to-read-a-text-file-line-by-line-in-go-when-some-lines-are-long-enough-to-ca
	// fileScanner.Scan()
	// panic(fileScanner.Err())
	buf := make([]byte, 0, 64*1024)
	fileScanner.Buffer(buf, 1024*10*1024)

	for fileScanner.Scan() {
		if *module == "ms" {
			insertMSOne(*consoleUrl, fileScanner.Text())
		}
	}
}

func insertMSOne(consoleUrl string, text string) {
	x := strings.Split(text, " ")

	url := x[0]
	picture := x[1]
	timeNow := time.Now().Format("2006-01-02 15:04:05")
	status := "待处理"

	m := md5.New()
	m.Write([]byte(url))
	alarmMd5 := fmt.Sprintf("%x", m.Sum(nil))

	url = fmt.Sprintf("<a href=%s>%s</a>", url, url)

	postData := fmt.Sprintf(`{"alarm_md5":"%s","url":"%s","screenshot":"%s","uploaddate":"%s","status":"%s"}`,
		alarmMd5, url, picture, timeNow, status)
	res, err := http.Post(consoleUrl+"/api/alarm/ms/add", "application/json", strings.NewReader(postData))
	if err != nil {
		panic(err)
		return
	}
	defer res.Body.Close()
	//if res.StatusCode != 200 {
	//	body, _ := ioutil.ReadAll(res.Body)
	//	println("insert failed, err: ", err)
	//	println(string(body))
	//}
}
