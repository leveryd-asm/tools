package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var nucleiResultFile, server string

func sendToApi(msg string) {
	if server == "" && os.Getenv("server") != "" {
		server = os.Getenv("server")
	}
	if server == "" {
		panic("server address is empty")
		return
	}
	alarmInfo := msg
	m := md5.New()
	m.Write([]byte(alarmInfo))
	alarmMd5 := fmt.Sprintf("%x", m.Sum(nil))

	timeNow := time.Now().Format("2006-01-02 15:04:05")

	post := map[string]interface{}{
		"alarminfo":  alarmInfo,
		"alarm_md5":  alarmMd5,
		"enable":     1,
		"status":     "待处理",
		"uploaddate": timeNow,
	}
	api := server + "/api/alarm/bbscan/add"

	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(post)
	if err != nil {
		return
	}

	_, err = http.Post(api, "application/json", requestBody)
	if err != nil {
		return
	}
}

func sendAlert(alert string) {
	x := strings.Split(alert, " ")
	if len(x) == 0 {
		return
	}

	url := x[len(x)-1]
	if strings.HasPrefix(url, "http") {
		msg := fmt.Sprintf("<a href='%s'>%s</a>", url, alert)
		sendToApi(msg)
	} else {
		sendToApi(alert)
	}
}

func main() {
	flag.StringVar(&nucleiResultFile, "f", "", "nuclei result file path")
	flag.StringVar(&server, "api", "", "api address, such as 'http://x.x.x.x:8080'")
	flag.Parse()

	if nucleiResultFile == "" {
		flag.Usage()
		return
	}

	file, err := os.Open(nucleiResultFile)
	if err != nil {
		panic(err)
		return
	}
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		println(fileScanner.Text())
		sendAlert(fileScanner.Text())
	}
}
