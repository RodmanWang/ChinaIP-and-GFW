package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/oschwald/geoip2-golang"
)

const defaultDataFile = "Country.mmdb"

var defaultIPs = []string{
	"103.200.30.143", "103.228.130.61", "216.58.200.238",
	"103.200.30.245", "118.184.26.113", "103.200.31.172",
	"69.171.235.101", "123.126.55.41", "117.23.61.238",
}

func main() {
	dataFile := defaultDataFile
	if len(os.Args) > 1 {
		dataFile = os.Args[1]
	}
	ips := defaultIPs
	if len(os.Args) > 2 {
		ips = os.Args[2:]
	}

	db, err := geoip2.Open(dataFile)
	if err != nil {
		log.Fatalf("open database %q: %v", dataFile, err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	failed := false
	for _, ipText := range ips {
		ip := net.ParseIP(ipText)
		if ip == nil {
			log.Printf("skip invalid IP %q", ipText)
			failed = true
			continue
		}
		record, err := db.Country(ip)
		if err != nil {
			log.Printf("lookup %s: %v", ipText, err)
			failed = true
			continue
		}
		fmt.Printf("IP:%s-Locale:%s\n", ipText, record.Country.IsoCode)
	}
	if failed {
		os.Exit(1)
	}
}
