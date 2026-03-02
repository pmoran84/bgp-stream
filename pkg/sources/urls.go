package sources

import (
	"fmt"
	"time"
)

const (
	WorldCitiesURL       = "https://raw.githubusercontent.com/dr5hn/countries-states-cities-database/master/csv/cities.csv"
	CityDominanceMetaURL = "https://map.kmcd.dev/data/city-dominance/meta.json"
	CityDominanceDataURL = "https://map.kmcd.dev/data/city-dominance/%d.json"

	AWSRangesURL      = "https://ip-ranges.amazonaws.com/ip-ranges.json"
	GoogleGeofeedURL  = "https://www.gstatic.com/ipranges/cloud_geofeed"
	AzureRangesURL    = "https://download.microsoft.com/download/7/1/d/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_20260223.json"
	OracleRangesURL   = "https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json"
	DigitalOceanURL   = "https://digitalocean.com/geo/google.csv"
	PeeringDBIXPURL   = "https://peeringdb.com/api/ix"
	PeeringDBIXLANURL = "https://peeringdb.com/api/ixlan?fields=ix_id,ixpfx_set&depth=2"

	APNICDelegatedURL   = "https://ftp.apnic.net/stats/apnic/delegated-apnic-latest"
	RIPEDelegatedURL    = "https://ftp.ripe.net/pub/stats/ripencc/delegated-ripencc-latest"
	RIPEInetnumURL      = "https://ftp.ripe.net/ripe/dbase/split/ripe.db.inetnum.gz"
	AFRINICDelegatedURL = "https://ftp.afrinic.net/pub/stats/afrinic/delegated-afrinic-latest"
	LACNICDelegatedURL  = "https://ftp.lacnic.net/pub/stats/lacnic/delegated-lacnic-latest"
	ARINDelegatedURL    = "https://ftp.arin.net/pub/stats/arin/delegated-arin-extended-latest"
)

func GetPeeringDBBackupURL() string {
	t := time.Now().AddDate(0, 0, -1) // One day behind
	return fmt.Sprintf("https://publicdata.caida.org/datasets/peeringdb/%04d/%02d/peeringdb_2_dump_%04d_%02d_%02d.json",
		t.Year(), t.Month(), t.Year(), t.Month(), t.Day())
}
