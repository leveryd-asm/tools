package main

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
)

var manageKeyword = []string{"admin", "manage", "manager", "login", "backend", "monitor"}

//
//func makeKeywordQuery() string {
//	x := make([]string, 0)
//	for _, keyword := range manageKeyword {
//		x = append(x, "host:"+keyword)
//	}
//	return strings.Join(x, " OR ")
//}
//

func fingerprintManageSystem(es, index, q string) {
	//identifyByHost(es, index, q)
}

func isHostContainMSKeyword(host string) bool {
	// the easiest way to get domain prefix, maybe not right, but it doesn't matter now
	x := strings.Split(host, ".")
	if len(x) < 2 {
		// println("weired", host)
		return false
	}
	domainPrefix := strings.Join(x[0:len(x)-2], ".")
	//println(domainPrefix)

	for _, keyword := range manageKeyword {
		if strings.Contains(domainPrefix, keyword) {
			return true
		}
	}
	return false
}

func isVue(resBody string) bool {
	result := false

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return false
	}
	selection := doc.Contents()
	if selection.Find("script").Length() > 0 {
		selection.Find("script").Each(func(i int, s *goquery.Selection) {
			src, exist := s.Attr("src")
			if exist && (strings.Contains(src, "/manifest.") || strings.Contains(src, "/vendor.") || strings.Contains(src, "/app.")) {
				result = true
			}
		})
	}
	return result
}

/*
1. 域名中包含关键词，直接当作后台管理系统
2. 使用vue、react、angular等框架的网站
3. 存在<table>标签，可能需要渲染数据
*/
