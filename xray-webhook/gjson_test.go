package main

import (
	"github.com/tidwall/gjson"
	"testing"
)

func TestGjson(t *testing.T) {
	body := `{
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
}`
	detail := gjson.Get(body, "data.detail")
	println(detail.String())
}
