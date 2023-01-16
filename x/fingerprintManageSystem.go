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
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return false
	}
	selection := doc.Contents()
	c := selection.Children()

	bodyTags := make([]string, 0)
	divElements := 0
	for i := 0; i < c.Length(); i++ {
		if goquery.NodeName(c.Eq(i)) == "body" {
			for j := 0; j < c.Eq(i).Children().Length(); j++ {
				selection := c.Eq(i).Children().Eq(j)
				tagName := goquery.NodeName(selection)
				bodyTags = append(bodyTags, tagName)
				if tagName == "div" {
					divElements = selection.Children().Length()
				} else if tagName == "script" {
					if selection.Text() != "" {
						return false
					}
				}
			}
		}
	}

	var hasScript, hasDiv bool

	if divElements > 3 {
		return false
	}
	for _, t := range bodyTags {
		if t == "div" {
			hasDiv = true
		} else if t == "script" {
			hasScript = true
		} else if t == "noscript" {
		} else {
			return false
		}
	}

	if hasScript && hasDiv {
		return true
	}
	return false
}

/*
1. 域名中包含关键词，直接当作后台管理系统
2. 使用vue、react、angular等框架的网站
3. 存在<table>标签，可能需要渲染数据
*/
