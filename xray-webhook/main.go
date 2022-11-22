package main

import (
	"crypto/tls"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// 企业微信机器人的webhook地址
// https://developer.work.weixin.qq.com/document/path/91770
var webhookURL = os.Getenv("WEBHOOK_URL")

func sendToQYWX(msg string) {
	content := fmt.Sprintf(`{"msgtype": "text", "text": {"content": "%s"}}`, msg)
	transport := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: &transport}
	_, err := client.Post(webhookURL, "application/json", strings.NewReader(content))
	if err != nil {
		log.Fatalln(err)
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
