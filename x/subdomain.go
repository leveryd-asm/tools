package main

import (
	"bufio"
	xFlag "flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func getSubdomainProxy(args []string) {
	flag := xFlag.NewFlagSet("subdomain", xFlag.ExitOnError)
	domain := flag.String("domain", "", "whose subdomain you want to get or save")
	source := flag.String("source", "es", "es or console")
	action := flag.String("action", "get", "get or save")
	debug := flag.Bool("debug", false, "debug mode")

	// es source
	index := flag.String("index", "proxify", "es index")
	esUrl := flag.String("esURL", "http://localhost:9200", "es url")

	// console source
	consoleUrl := flag.String("consoleUrl", "http://console.com:32115", "console url")
	q := flag.String("q", "limit=1000", "additional query") // if u fetch 10000, http client will time out

	// mysql source
	mysqlDataSource := flag.String("datasource", "root:console_db_root_password@tcp(mysql-service)/cute", "mysql datasource")
	mysqlSQL := flag.String("sql", "limit 10000", "mysql condition")

	// save action
	of := flag.String("of", "", "output file path")
	// get action
	subdomainsFilePath := flag.String("if", "", "input file path")

	flag.Parse(args)
	args = flag.Args()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *action == "get" && *source == "es" {
		getSubdomainFromES(*esUrl, *index, *domain, *of)
	} else if *action == "get" && *source == "console" {
		getSubdomainFromConsole(*consoleUrl, *domain, *q, *of)
	} else if *action == "save" && *source == "console" {
		saveSubdomainFromConsole(*consoleUrl, *domain, *subdomainsFilePath)
	} else if *action == "get" && *source == "mysql" {
		getSubdomainFromMysql(*mysqlDataSource, *domain, *mysqlSQL, *of, *debug)
	} else {
		flag.Usage()
	}

}

func getSubdomainFromMysql(DataSource string, domain, q, of string, debug bool) {
	log.Debug("get subdomain from mysql, ", DataSource)
	db, err := gorm.Open(mysql.Open(DataSource), &gorm.Config{})
	if err != nil {
		log.Warn("open mysql failed")
		return
	}
	if debug {
		db = db.Debug()
	}

	result := make([]string, 0)

	where := "parentdomain = ? "
	if q != "" {
		where += q
	}
	//db.Table("subdomain").Where(where, domain).Find(&result)
	db.Raw("select subdomain from subdomain where "+where, domain).Scan(&result)

	if of != "" {
		log.Debug("write subdomain to file: ", of)
		writeFile(of, result)
	} else {
		for _, i := range result {
			fmt.Println(i)
		}
	}
}

func getSubdomainFromConsole(consoleUrl, domain, q, outputFilePath string) {
	xUrl := consoleUrl + "/api/info/subdomain/query?parentdomain=" + domain
	if q != "" {
		xUrl += "&" + q
	}
	log.Debug("get subdomain from console, ", xUrl)

	res, err := http.Get(xUrl)
	if err != nil || res == nil {
		log.Warn("get subdomain from console failed")
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warn("read response body failed")
		return
	}

	rows := gjson.Get(string(body), "rows")

	domains := make([]string, 0)
	for _, row := range rows.Array() {
		println(row.Get("subdomain").String())
		domains = append(domains, row.Get("subdomain").String())
	}

	if outputFilePath != "" {
		log.Debug("write subdomain to file: ", outputFilePath)
		writeFile(outputFilePath, domains)
	}
}

func saveSubdomainFromConsole(consoleUrl, domain, subdomainsFilePath string) {
	f, err := os.Open(subdomainsFilePath)
	if err != nil {
		return
	}
	defer f.Close()

	xUrl := consoleUrl + "/api/info/subdomain/add"
	log.Debug("save subdomain to console, ", xUrl)

	scanner := bufio.NewScanner(f)

	count := 0
	for scanner.Scan() {
		subdomain := scanner.Text()
		timeNow := time.Now().Format("2006-01-02 15:04:05")

		postdata := `{"parentdomain":"%s","subdomain":"%s", "uploaddate":"%s", "enable": true}`
		postdata = fmt.Sprintf(postdata, domain, subdomain, timeNow)

		res, err := http.Post(xUrl, "application/json", strings.NewReader(postdata))
		if err != nil || res == nil {
			log.Warn("save subdomain to console failed")
			return
		}
		res.Body.Close()
		count++
		fmt.Printf("process:%d\r", count)
	}
	println("\nsave subdomain to console success, count: ", count)
}

// get subdomain from http request host, the subdomain ends with the domain
// filter out null result
func getSubdomainFromES(es, index, domain, outputFilePath string) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"host": domain,
			},
		},
		"collapse": map[string]interface{}{
			"field": "host.keyword",
		},
		"size": 10000,
	}
	domains := newClient(es, index).getDocument(query, "host")

	filterDomains := make([]string, 0)
	for _, i := range domains {
		if strings.HasSuffix(i, "."+domain) {
			filterDomains = append(filterDomains, i)
		}
	}

	if outputFilePath == "" {
		for _, domain := range filterDomains {
			fmt.Println(domain)
		}
	} else {
		writeFile(outputFilePath, filterDomains)
	}
}

func writeFile(path string, domains []string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Error creating file: %s", err)
	}
	defer file.Close()

	for _, domain := range domains {
		_, err := file.WriteString(domain + "\n")
		if err != nil {
			log.Fatalf("Error writing string: %s", err)
		}
	}
}
