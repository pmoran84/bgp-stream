package bgpengine

import (
	"image/color"
	"testing"
	"time"

	"github.com/sudorandom/bgp-stream/pkg/utils"
)

func TestCriticalStreamDeduplication(t *testing.T) {
	e := &Engine{
		criticalCooldown: make(map[string]time.Time),
		asnMapping:       utils.NewASNMapping(),
	}
	// Initializing fonts and other UI stuff is not needed for this logic test

	c := color.RGBA{255, 0, 0, 255}
	name := nameHardOutage

	// Event 1: Outage for ASN 1234, prefix 1.1.1.0/24
	ev1 := &bgpEvent{
		classificationType: ClassificationOutage,
		prefix:             "1.1.1.0/24",
		asn:                1234,
		cc:                 "US",
	}
	e.recordToCriticalStream(ev1, c, name)

	if len(e.CriticalStream) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(e.CriticalStream))
	}

	// Event 2: Same outage (same ASN), different prefix 1.1.2.0/24
	ev2 := &bgpEvent{
		classificationType: ClassificationOutage,
		prefix:             "1.1.2.0/24",
		asn:                1234,
		cc:                 "US",
	}
	e.recordToCriticalStream(ev2, c, name)

	if len(e.CriticalStream) != 1 {
		t.Errorf("Expected 1 event after deduplication, got %d", len(e.CriticalStream))
	}

	// Event 3: Outage with ASN 0 (unknown), prefix 2.2.1.0/24
	ev3 := &bgpEvent{
		classificationType: ClassificationOutage,
		prefix:             "2.2.1.0/24",
		asn:                0,
		historicalASN:      5678,
		cc:                 "FR",
	}
	e.recordToCriticalStream(ev3, c, name)
	if len(e.CriticalStream) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(e.CriticalStream))
	}

	// Event 4: Same outage (ASN 0), but another prefix 2.2.2.0/24, same historical ASN
	// This should now be deduplicated if we use historicalASN
	ev4 := &bgpEvent{
		classificationType: ClassificationOutage,
		prefix:             "2.2.2.0/24",
		asn:                0,
		historicalASN:      5678,
		cc:                 "FR",
	}
	e.recordToCriticalStream(ev4, c, name)

	if len(e.CriticalStream) != 2 {
		t.Errorf("Expected 2 events after deduplication of ASN 0 outage, got %d", len(e.CriticalStream))
	}

	// Event 5: DDoS Mitigation, Provider 13335, Victim 9999, Prefix 3.3.1.0/24
	nameDDoS := nameDDoSMitigation
	ev5 := &bgpEvent{
		classificationType: ClassificationDDoSMitigation,
		prefix:             "3.3.1.0/24",
		asn:                13335,
		historicalASN:      9999,
		cc:                 "NL",
		leakDetail: &LeakDetail{
			LeakerASN: 13335,
			VictimASN: 9999,
		},
	}
	e.recordToCriticalStream(ev5, c, nameDDoS)
	if len(e.CriticalStream) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(e.CriticalStream))
	}

	// Event 6: Same DDoS Mitigation, different prefix
	// This should now be deduplicated
	ev6 := &bgpEvent{
		classificationType: ClassificationDDoSMitigation,
		prefix:             "3.3.2.0/24",
		asn:                13335,
		historicalASN:      9999,
		cc:                 "NL",
		leakDetail: &LeakDetail{
			LeakerASN: 13335,
			VictimASN: 9999,
		},
	}
	e.recordToCriticalStream(ev6, c, nameDDoS)
	if len(e.CriticalStream) != 3 {
		t.Errorf("Expected 3 events after deduplication of DDoS mitigation, got %d", len(e.CriticalStream))
	}
}

type fakeWriteCloser struct{}

func (f *fakeWriteCloser) Write(p []byte) (n int, err error) { return len(p), nil }
func (f *fakeWriteCloser) Close() error                      { return nil }

func TestCriticalStreamExpiration(t *testing.T) {
	startTime := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	e := &Engine{
		criticalCooldown: make(map[string]time.Time),
		asnMapping:       utils.NewASNMapping(),
		virtualTime:      startTime,
		virtualStartTime: startTime, // Enable virtual time logic in e.Now()
	}

	// Mock Now() behavior without needing VideoWriter
	// Actually, looking at e.Now(), it needs VideoWriter != nil
	e.VideoWriter = &fakeWriteCloser{}

	c := color.RGBA{255, 0, 0, 255}
	name := nameHardOutage

	// T=0: Event 1 arrives
	ev1 := &bgpEvent{
		classificationType: ClassificationOutage,
		prefix:             "1.1.1.0/24",
		asn:                1234,
	}
	e.recordToCriticalStream(ev1, c, name)

	if len(e.CriticalStream) != 1 {
		t.Fatalf("Expected 1 event at T=0, got %d", len(e.CriticalStream))
	}

	// T=5m: Duplicate event arrives
	e.virtualTime = startTime.Add(5 * time.Minute)
	e.recordToCriticalStream(ev1, c, name)

	if len(e.CriticalStream) != 1 {
		t.Fatalf("Expected 1 event at T=5m, got %d", len(e.CriticalStream))
	}

	// T=11m: Event should expire even though there was a duplicate at T=5m
	e.virtualTime = startTime.Add(11 * time.Minute)
	e.updateMetrics() // This runs the cleanup logic

	if len(e.CriticalStream) != 0 {
		t.Errorf("Expected 0 events at T=11m (expired), got %d", len(e.CriticalStream))
	}
}
