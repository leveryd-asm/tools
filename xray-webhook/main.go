package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// 企业微信机器人的webhook地址
// https://developer.work.weixin.qq.com/document/path/91770
var webhookURL = os.Getenv("WEBHOOK_URL")

func sendToQYWX(msg string) {
	// text message can not contain \n
	msg = strings.ReplaceAll(msg, "\n", "<br>")
	msg = strings.ReplaceAll(msg, "\r", "<br>")

	// escape quotes
	msg = strconv.Quote(msg)

	// text msg length can not be more than 2048 byte
	// https://work.weixin.qq.com/api/doc/90000/90136/91770
	if len(msg) > 2048 {
		msg = msg[:2000] + "..."
	}

	content := fmt.Sprintf(`{"msgtype": "text", "text": {"content": %s}}`, msg)
	transport := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: &transport}
	_, err := client.Post(webhookURL, "application/json", strings.NewReader(content))

	if err != nil {
		log.Fatalln(err)
	}

	// todo: check response status code and body
}

func sendToApi(msg string) {
	api := os.Getenv("api")
	if api == "" {
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
	api = api + "/api/alarm/bbscan/add"

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

func xrayWebhook() func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(wrt http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Fatalln(err)
		}
		// xray webhook data format: https://docs.xray.cool/#/webhook/vuln
		value := gjson.Get(string(body), "data.detail")
		msg := value.String()
		sendToQYWX(msg)

		if os.Getenv("api") != "" {
			url := gjson.Get(string(body), "data.target.url")
			plugin := gjson.Get(string(body), "data.plugin")
			params := gjson.Get(string(body), "data.target.params")
			msg := fmt.Sprintf("url: %s\nplugin: %s\nparams: %s", url, plugin, params)
			sendToApi(msg)
		}
	})
}

func main() {
	if webhookURL == "" {
		log.Fatalln("env WEBHOOK_URL is empty")
	}
	// Add handle func for webhook.
	http.HandleFunc("/webhook", xrayWebhook())

	// Run the web server.
	fmt.Println("start webhook api !!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*
test command:

curl 127.0.0.1:8080/webhook -d '{
  "type": "web_vuln",
  "data": {
    "create_time": 1604736253090,
    "detail": {
      "addr": "http://127.0.0.1:9000/xss/example1.php?name=hacker",
      "extra": {
        "param": {
          "key": "name",
          "position": "query",
          "value": "pkbnekwkjhwzabxnfjwh"
        }
      },
      "payload": "<sCrIpT>alert(1)</ScRiPt>",
      "snapshot": [
        [
          "GET /xxx",
          "HTTP/1.1 200 OK"
        ]
      ]
    },
    "plugin": "xss/reflected/default",
    "target": {
      "params": [],
      "url": "http://127.0.0.1:9000/xss/example1.php"
    }
  }
}'

*/
