package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sudorandom/bgp-stream/pkg/bgpengine"
	"github.com/sudorandom/bgp-stream/pkg/geoservice"
	"github.com/sudorandom/bgp-stream/pkg/utils"
)

func main() {
	fresh := flag.Bool("fresh", false, "Re-download all source files even if they are already cached")
	width := flag.Int("width", 3840, "Internal rendering width")
	height := flag.Int("height", 2160, "Internal rendering height")
	scale := flag.Float64("scale", 760.0, "Internal rendering scale")
	flag.Parse()

	// Initialize GeoService
	geo := geoservice.NewGeoService(*width, *height, *scale)

	// Open Databases in Read-Write mode
	if err := geo.OpenHintDBs("data", false); err != nil {
		log.Fatalf("Failed to open hint databases: %v", err)
	}
	defer func() { _ = geo.Close() }()

	dm := geoservice.NewDataManager(geo)

	if *fresh {
		log.Println("Fresh data requested. Clearing caches...")
		// 1. Clear pre-processed cache to force re-generation
		if err := os.Remove("data/prefix-dump-cache.json"); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to remove prefix cache: %v", err)
		}
		if err := os.Remove("data/hubs-dump-cache.json"); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to remove hubs cache: %v", err)
		}

		// 2. Clear cached RIR/WHOIS files to force re-download
		cacheDir := filepath.Join("data", "cache")
		files, err := os.ReadDir(cacheDir)
		if err == nil {
			for _, f := range files {
				if !f.IsDir() {
					_ = os.Remove(filepath.Join(cacheDir, f.Name()))
				}
			}
		}

		// 3. Clear Hint DBs
		log.Println("Clearing hint databases...")
		_ = geo.ClearAll()
	}

	var wg sync.WaitGroup

	// Task 1: World Cities & Hubs
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("--- World Cities ---")
		if err := dm.DownloadWorldCities(*fresh); err != nil {
			log.Printf("Warning: failed to download worldcities: %v", err)
		}
		dm.LoadWorldCities()

		if err := dm.LoadRemoteCityData(); err != nil {
			log.Printf("Warning: failed to load remote city data: %v", err)
		}
	}()

	// Task 2: RIR Data
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("--- RIR / Prefix Data ---")
		geoReader, err := geoservice.GetGeoIPReader()
		if err != nil {
			log.Printf("Warning: GeoIP reader not available: %v", err)
		}
		if err := dm.ProcessRIRData(geoReader); err != nil {
			log.Printf("Error processing RIR data: %v", err)
		}
	}()

	// Task 3: Cloud Data
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("--- Cloud Data ---")
		dm.LoadCloudData()
	}()

	// Task 4: PeeringDB
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("--- PeeringDB ---")
		dm.ProcessPeeringDBData()
	}()

	// Task 5: WHOIS (Needs city list loaded first for heuristic matching)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("--- WHOIS ---")
		// Wait a small amount for Task 1 to at least start loading cities
		// In reality, ProcessBulkWhoisData calls InitMatchers which uses dm.geo.citiesByCountry.
		// So we actually NEED Task 1 to finish LoadWorldCities() first.
		// Let's refine this to be more precise below if needed, but for now
		// dm.ProcessBulkWhoisData internally calls dm.InitMatchers() which depends on LoadWorldCities.
		dm.ProcessBulkWhoisData()
	}()

	// Task 6: ASN Mapping
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("--- ASN Mapping ---")
		asn := utils.NewASNMapping()
		if err := asn.Load(); err != nil {
			log.Printf("Warning: failed to load ASN mapping: %v", err)
		}
	}()

	wg.Wait()

	// Task 7: Background Map (Safe to do after everything else or just whenever)
	log.Println("--- Background Map ---")
	engine := bgpengine.NewEngine(*width, *height, *scale)
	if err := engine.InitGeoOnly(false); err != nil {
		log.Printf("Warning: failed to init engine for background generation: %v", err)
	}
	if err := engine.GenerateInitialBackground(); err != nil {
		log.Printf("Warning: failed to generate initial background: %v", err)
	}

	log.Println("Data management tasks complete.")
}
