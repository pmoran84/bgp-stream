package sources

import (
	"io"

	"github.com/sudorandom/bgp-stream/pkg/utils"
)

type RIRSource struct {
	Name string
	URL  string
}

func GetRIRSources() []RIRSource {
	return []RIRSource{
		{"APNIC", "https://ftp.apnic.net/stats/apnic/delegated-apnic-extended-latest"},
		{"RIPE", "https://ftp.ripe.net/pub/stats/ripencc/delegated-ripencc-extended-latest"},
		{"AFRINIC", "https://ftp.afrinic.net/pub/stats/afrinic/delegated-afrinic-extended-latest"},
		{"LACNIC", "https://ftp.lacnic.net/pub/stats/lacnic/delegated-lacnic-extended-latest"},
		{"ARIN", "https://ftp.arin.net/pub/stats/arin/delegated-arin-extended-latest"},
	}
}

func GetBulkWhoisSources() []RIRSource {
	return []RIRSource{
		{"RIPE", RIPEInetnumURL},
		{"APNIC", "https://ftp.apnic.net/apnic/whois/apnic.db.inetnum.gz"},
		{"AFRINIC", "https://ftp.afrinic.net/pub/dbase/afrinic.db.gz"},
		{"LACNIC", "https://ftp.lacnic.net/lacnic/dbase/lacnic.db.gz"},
		// ARIN bulk WHOIS requires an AUP and is not publicly linkable in the same way.
	}
}

type GeoHint struct {
	Start, End uint32
	CC         string
	City       string
	Lat, Lng   float32
}

func GetRIPEWhoisReader() (io.ReadCloser, error) {
	return utils.GetCachedReader(RIPEInetnumURL, true, "[RIPE-WHOIS]")
}
