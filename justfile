# Run tests
test:
	go test ./...
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run ./...

run:
	go run ./cmd/bgp-viewer \
		-capture-interval 1h \
		-capture-dir ./archive

# Watch BGP updates for a specific prefix
debug-prefix prefix="146.66.28.0/22":
	go run ./cmd/debug-prefix -prefix {{prefix}}

# Fetch and process all required geolocation data
fetch-data:
	go run ./cmd/bgp-data-fetcher

# Force a re-download of all source data files
refresh-data:
	go run ./cmd/bgp-data-fetcher -fresh
