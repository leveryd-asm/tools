package main

import (
	xFlag "flag"
	"log"
)

var server string

func saveNucleiAlarm(args []string) {
	flag := xFlag.NewFlagSet("nuclei", xFlag.ExitOnError)

	flag.StringVar(&server, "api", "", "api address, such as 'http://x.x.x.x:8080'")
	nucleiResultFile := flag.String("f", "", "nuclei result file path")

	flag.Parse(args)
	args = flag.Args()

	if *nucleiResultFile == "" {
		flag.Usage()
		return
	}

	saveNucleiResult(*nucleiResultFile)
}

func scanManageSystemProxy(args []string) {
	flag := xFlag.NewFlagSet("scan", xFlag.ExitOnError)
	index := flag.String("index", "proxify", "es index")
	esUrl := flag.String("esURL", "http://localhost:9200", "es url")
	q := flag.String("q", "*", "what you want to search")

	flag.Parse(args)
	args = flag.Args()

	fingerprintManageSystem(*esUrl, *index, *q)
}

func getSubdomainProxy(args []string) {
	flag := xFlag.NewFlagSet("subdomain", xFlag.ExitOnError)
	index := flag.String("index", "proxify", "es index")
	esUrl := flag.String("esURL", "http://localhost:9200", "es url")
	domain := flag.String("domain", "", "whose subdomain you want to get")
	of := flag.String("of", "", "output file path")

	flag.Parse(args)
	args = flag.Args()

	getSubdomain(*esUrl, *index, *domain, *of)
}

func main() {
	xFlag.Parse()

	args := xFlag.Args()
	if len(args) == 0 {
		log.Fatal("Please specify a subcommand.")
	}
	cmd, args := args[0], args[1:]
	switch cmd {
	case "nuclei":
		saveNucleiAlarm(args)
	case "ms", "manage-system":
		scanManageSystemProxy(args)
	case "subdomain": // from es http log extract subdomain
		getSubdomainProxy(args)
	case "ims", "identifyMS":
		identifyMS(args)
	case "es":
		es(args)
	default:
		log.Fatalf("Unrecognized command %q. "+
			"Command must be one of: branch, checkout", cmd)
	}
}
