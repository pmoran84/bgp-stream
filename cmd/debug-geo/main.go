package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/sudorandom/bgp-stream/pkg/geoservice"
	"github.com/sudorandom/bgp-stream/pkg/utils"
)

func main() {
	ipStr := flag.String("ip", "", "IP address to resolve")
	flag.Parse()

	// Initialize GeoService
	geo := geoservice.NewGeoService(3840, 2160, 760.0)

	// Open Databases in Read-Only mode
	if err := geo.OpenHintDBs("data", true); err != nil {
		log.Printf("Warning: Failed to open hint databases: %v", err)
	}
	defer func() { _ = geo.Close() }()

	// Load city data
	dm := geoservice.NewDataManager(geo)
	dm.LoadWorldCities()
	if err := dm.LoadRemoteCityData(); err != nil {
		log.Printf("Warning: failed to load remote city data: %v", err)
	}

	// Initialize GeoIP
	geoReader, err := geoservice.GetGeoIPReader()
	if err == nil {
		geo.SetGeoIPReader(geoReader)
	} else {
		log.Printf("Warning: Failed to open GeoIP database: %v", err)
	}

	// Load prefix data
	cachePath := "data/prefix-dump-cache.json"
	if data, err := os.ReadFile(cachePath); err == nil {
		var prefixData geoservice.PrefixData
		if err := json.Unmarshal(data, &prefixData); err == nil {
			geo.SetPrefixData(prefixData)
		}
	}
	hubsCachePath := "data/hubs-dump-cache.json"
	if data, err := os.ReadFile(hubsCachePath); err == nil {
		var hubsData geoservice.PrefixData
		if err := json.Unmarshal(data, &hubsData); err == nil {
			geo.SetHubsData(hubsData)
		}
	}

	resolve := func(s string) {
		parsedIP := net.ParseIP(s).To4()
		if parsedIP == nil {
			fmt.Printf("Invalid IPv4: %s\n", s)
			return
		}
		ipUint := utils.IPToUint32(parsedIP)
		lat, lng, cc, resType := geo.GetIPCoords(ipUint)

		fmt.Printf("IP: %s\n", s)
		fmt.Printf("  Coords:     %f, %f\n", lat, lng)
		fmt.Printf("  Country:    %s\n", cc)
		fmt.Printf("  Resolution: %s\n", resType)
		fmt.Println("--------------------------------")
	}

	if *ipStr != "" {
		resolve(*ipStr)
		return
	}

	fmt.Println("Enter IPs to resolve (one per line, Ctrl+C to exit):")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			resolve(line)
		}
	}
}
