// 从文本中提取 二级域名、子域名、邮箱、ip、url 等信息 (暂时完成部分提取)
package main

import (
	"bufio"
	xFlag "flag"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"os"
	"regexp"
	"strings"
)

var (
	domainRegex   = regexp.MustCompile(`\b(?i)(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]\b`)
	subdomainList = make([]string, 0)
	tldDomainList = make([]string, 0)
	subdomainMap  = make(map[string]bool)
	tldDomainMap  = make(map[string]bool)
)

func txt(args []string) {
	flag := xFlag.NewFlagSet("txt", xFlag.ExitOnError)

	directoryPath := flag.String("d", ".", "directory path")
	filenameSuffix := flag.String("suffix", "txt", "file name suffix")
	outputDirectoryPath := flag.String("od", "./output", "output directory path")

	flag.Parse(args)

	if *directoryPath == "" {
		fmt.Println("directory path is empty")
		return
	}
	if *filenameSuffix == "" {
		fmt.Println("file name suffix is empty")
		return
	}
	if *outputDirectoryPath == "" {
		fmt.Println("output directory path is empty")
		return
	}
	err := os.MkdirAll(*outputDirectoryPath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}

	entries, err := os.ReadDir(*directoryPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	os.Chdir(*directoryPath)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), *filenameSuffix) {
			extractFile(e.Name())
		}
	}

	for k, _ := range tldDomainMap {
		tldDomainList = append(tldDomainList, k)
	}
	for k, _ := range subdomainMap {
		subdomainList = append(subdomainList, k)
	}

	if *outputDirectoryPath != "" {
		writeFile(*outputDirectoryPath+"/subdomain.txt", subdomainList)
		writeFile(*outputDirectoryPath+"/tldDomain.txt", tldDomainList)
	}
}

func extractFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		extractSubdomain(fileScanner.Text())
	}
}

func extractSubdomain(txt string) {
	txt = strings.ToLower(txt)
	txt = strings.ReplaceAll(txt, "%3a%2f%2f", "=")
	txt = strings.ReplaceAll(txt, "%253a%252f%252f", "=")
	txt = strings.ReplaceAll(txt, "%2f", "=")

	lines := strings.Split(txt, "<br>")
	for _, line := range lines {
		domains := domainRegex.FindAllString(line, -1)
		for _, domain := range domains {
			eTLD, icann := publicsuffix.PublicSuffix(domain)
			// Only ICANN managed domains can have a single label. Privately
			// managed domains must have multiple labels.
			manager := "Unmanaged"
			if icann {
				manager = "ICANN Managed"
			} else if strings.IndexByte(eTLD, '.') >= 0 {
				manager = "Privately Managed"
			}
			if manager == "ICANN Managed" {
				tldDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
				if err != nil {
					continue
				}

				subdomainMap[domain] = true
				tldDomainMap[tldDomain] = true
			}
		}
	}
}

//func main() {
//	extractSubdomain("*.wwww.baidu.com.cn")
//}
