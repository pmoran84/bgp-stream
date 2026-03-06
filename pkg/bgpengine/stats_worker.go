package bgpengine

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"

	"github.com/biter777/countries"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type hub struct {
	cc   string
	rate float64
}

func getMaskLen(prefix string) int {
	idx := strings.IndexByte(prefix, '/')
	if idx == -1 {
		return 0
	}
	mask, _ := strconv.Atoi(prefix[idx+1:])
	return mask
}

type statsEvent struct {
	ev         *bgpEvent
	name       string
	c          color.RGBA
	uiInterval float64
	trigger    bool
}

func (e *Engine) runStatsWorker() {
	countryActivity := make(map[string]int)
	currentAnomalies := make(map[ClassificationType]map[string]int)
	prefixToASN := make(map[string]uint32)
	prefixToClassification := make(map[string]ClassificationType)
	visualImpact := make(map[string]*VisualImpact)

	impactMap := make(map[string]*VisualImpact)
	countMap := make(map[string]*PrefixCount)
	asnsPerClass := make(map[string]map[uint32]struct{})
	asnGroups := make(map[asnGroupKey]*asnGroup)
	asnSortedGroups := make([]*asnGroup, 0, 100)
	hubCurrent := make([]hub, 0, 300)

	for {
		select {
		case msg, ok := <-e.statsCh:
			if !ok {
				return
			}
			if msg.trigger {
				uiInterval := msg.uiInterval
				if uiInterval <= 0 {
					uiInterval = 20.0
				}

				hubCurrent = hubCurrent[:0]
				for cc, val := range countryActivity {
					hubCurrent = append(hubCurrent, hub{cc, float64(val) / uiInterval})
				}
				sort.Slice(hubCurrent, func(i, j int) bool { return hubCurrent[i].rate > hubCurrent[j].rate })

				maxItems := 5
				if len(hubCurrent) < maxItems {
					maxItems = len(hubCurrent)
				}

				activeHubs := make([]*VisualHub, 0, maxItems)
				hubYBase := float64(e.Height) * 0.48
				fontSize := 18.0
				if e.Width > 2000 {
					fontSize = 36.0
					hubYBase = float64(e.Height) * 0.42
				}
				spacing := fontSize * 1.0

				for i := 0; i < maxItems; i++ {
					targetY := hubYBase + float64(i)*spacing

					cc := hubCurrent[i].cc
					countryName := countries.ByName(cc).String()
					if countryName == strUnknown {
						countryName = cc
					}
					if idx := strings.Index(countryName, " ("); idx != -1 {
						countryName = countryName[:idx]
					}
					if strings.Contains(countryName, "Hong Kong") {
						countryName = "Hong Kong"
					}
					if strings.Contains(countryName, "Macao") {
						countryName = "Macao"
					}
					if strings.Contains(countryName, "Taiwan") {
						countryName = "Taiwan"
					}
					const maxLen = 18
					if len(countryName) > maxLen {
						countryName = countryName[:maxLen-3] + "..."
					}

					vh := &VisualHub{
						CC:          cc,
						CountryStr:  countryName,
						DisplayY:    targetY,
						TargetY:     targetY,
						Alpha:       0,
						TargetAlpha: 1.0,
						Rate:        hubCurrent[i].rate,
						RateStr:     fmt.Sprintf("%.0f", hubCurrent[i].rate),
						Active:      true,
					}
					vh.RateWidth, _ = text.Measure(vh.RateStr, e.subMonoFace, 0)
					activeHubs = append(activeHubs, vh)
				}

				clear(impactMap)
				for et, prefixes := range currentAnomalies {
					_, name := e.getClassificationVisuals(et)
					prio := e.GetPriority(name)

					for p, count := range prefixes {
						visI, ok := impactMap[p]
						if !ok {
							visI, ok = visualImpact[p]
							if !ok {
								visI = &VisualImpact{Prefix: p, MaskLen: getMaskLen(p)}
								visualImpact[p] = visI
							}
							visI.ClassificationName = ""
							visI.Count = 0
							impactMap[p] = visI
						}

						if name != "" && (visI.ClassificationName == "" || prio > e.GetPriority(visI.ClassificationName)) {
							visI.ClassificationName = name
							visI.ClassificationColor, _ = e.getClassificationVisuals(et)
						}
						visI.Count += float64(count) / uiInterval
					}
				}

				allImpact := make([]*VisualImpact, 0, len(impactMap))
				for _, visI := range impactMap {
					allImpact = append(allImpact, visI)
				}

				clear(countMap)
				for _, m := range asnsPerClass {
					clear(m)
				}
				clear(asnsPerClass)

				allClasses := []ClassificationType{
					ClassificationRouteLeak, ClassificationOutage, ClassificationLinkFlap,
					ClassificationNextHopOscillation, ClassificationAggFlap, ClassificationPolicyChurn,
					ClassificationDDoSMitigation, ClassificationPathLengthOscillation, ClassificationPathHunting,
					ClassificationDiscovery,
				}
				for _, ct := range allClasses {
					name := ct.String()
					prio := e.GetPriority(name)
					countMap[name] = &PrefixCount{
						Name:     name,
						Count:    0,
						Rate:     0,
						Color:    e.getClassificationUIColor(name),
						Priority: prio,
						Type:     ct,
					}
				}

				for _, visI := range allImpact {
					if visI.ClassificationName == "" {
						continue
					}
					asn := prefixToASN[visI.Prefix]
					if pc, ok := countMap[visI.ClassificationName]; ok {
						pc.Count++
						pc.Rate += visI.Count
					}
					m, ok := asnsPerClass[visI.ClassificationName]
					if !ok {
						m = make(map[uint32]struct{})
						asnsPerClass[visI.ClassificationName] = m
					}
					m[asn] = struct{}{}
				}

				prefixCounts := make([]PrefixCount, 0, len(countMap))
				for name, pc := range countMap {
					pc.ASNCount = len(asnsPerClass[name])
					pc.ASNStr = strconv.Itoa(pc.ASNCount)
					pc.CountStr = strconv.Itoa(pc.Count)
					pc.RateStr = fmt.Sprintf("%.0f", pc.Rate)

					pc.RateWidth, _ = text.Measure(pc.RateStr, e.subMonoFace, 0)
					pc.ASNWidth, _ = text.Measure(pc.ASNStr, e.subMonoFace, 0)
					pc.CountWidth, _ = text.Measure(pc.CountStr, e.subMonoFace, 0)
					prefixCounts = append(prefixCounts, *pc)
				}

				sort.Slice(prefixCounts, func(i, j int) bool {
					if prefixCounts[i].Priority != prefixCounts[j].Priority {
						return prefixCounts[i].Priority > prefixCounts[j].Priority
					}
					if prefixCounts[i].Count != prefixCounts[j].Count {
						return prefixCounts[i].Count > prefixCounts[j].Count
					}
					return prefixCounts[i].Name < prefixCounts[j].Name
				})

				clear(asnGroups)
				for _, visI := range allImpact {
					prio := e.GetPriority(visI.ClassificationName)
					if prio < 1 {
						continue
					}
					asn := prefixToASN[visI.Prefix]
					if asn == 0 && visI.LeakerASN != 0 {
						asn = visI.LeakerASN
					}
					if asn == 0 {
						continue
					}
					key := asnGroupKey{ASN: asn, Anom: visI.ClassificationName}
					g, ok := asnGroups[key]
					if !ok {
						networkName := ""
						if e.asnMapping != nil {
							networkName = e.asnMapping.GetName(asn)
						}
						asnStr := fmt.Sprintf("AS%d", asn)
						if networkName != "" {
							asnStr = fmt.Sprintf("AS%d - %s", asn, networkName)
						}
						g = &asnGroup{
							asnStr:    asnStr,
							anom:      visI.ClassificationName,
							color:     e.getClassificationUIColor(visI.ClassificationName),
							priority:  prio,
							maxCount:  visI.Count,
							prefixes:  make([]string, 0, 4),
							locations: make(map[string]struct{}),
						}
						asnGroups[key] = g
					}

					if visI.Count > g.maxCount {
						g.maxCount = visI.Count
					}
					g.totalCount += visI.Count

					if visI.LeakType != LeakUnknown {
						g.leakType = visI.LeakType
						g.leakerASN = visI.LeakerASN
						g.victimASN = visI.VictimASN
					}

					for cc := range visI.CCs {
						g.locations[cc] = struct{}{}
					}
					g.prefixes = append(g.prefixes, visI.Prefix)
				}

				asnSortedGroups = asnSortedGroups[:0]
				for _, g := range asnGroups {
					asnSortedGroups = append(asnSortedGroups, g)
				}
				sort.Slice(asnSortedGroups, func(i, j int) bool {
					if asnSortedGroups[i].priority != asnSortedGroups[j].priority {
						return asnSortedGroups[i].priority > asnSortedGroups[j].priority
					}
					return asnSortedGroups[i].totalCount > asnSortedGroups[j].totalCount
				})

				activeASNImpacts := make([]*ASNImpact, 0, 5)
				for i := 0; i < len(asnSortedGroups) && i < 5; i++ {
					g := asnSortedGroups[i]
					displayPrefixes := g.prefixes
					moreCount := 0
					if len(displayPrefixes) > 1 {
						moreCount = len(displayPrefixes) - 1
						displayPrefixes = displayPrefixes[:1]
					}
					moreStr := ""
					if moreCount > 0 {
						moreStr = fmt.Sprintf("(%d more)", moreCount)
					}
					anomWidth, _ := text.Measure(g.anom, e.subMonoFace, 0)

					locs := make([]string, 0, len(g.locations))
					for cc := range g.locations {
						locs = append(locs, cc)
					}
					sort.Strings(locs)
					locStr := strings.Join(locs, ", ")

					activeASNImpacts = append(activeASNImpacts, &ASNImpact{
						ASNStr:    g.asnStr,
						Prefixes:  displayPrefixes,
						MoreStr:   moreStr,
						Anom:      g.anom,
						AnomWidth: anomWidth,
						Color:     g.color,
						Count:     len(g.prefixes),
						Rate:      g.totalCount,
						LeakType:  g.leakType,
						LeakerASN: g.leakerASN,
						VictimASN: g.victimASN,
						Locations: locStr,
					})
				}

				e.metricsMu.Lock()

				for _, vh := range e.VisualHubs {
					vh.Active = false
					vh.TargetAlpha = 0.0
				}
				e.ActiveHubs = e.ActiveHubs[:0]
				for _, newHub := range activeHubs {
					existing, ok := e.VisualHubs[newHub.CC]
					if !ok {
						existing = newHub
						e.VisualHubs[newHub.CC] = existing
					} else {
						existing.Active = true
						existing.TargetY = newHub.TargetY
						if existing.Alpha < 0.01 {
							existing.DisplayY = newHub.TargetY
						}
						existing.TargetAlpha = 1.0
						existing.Rate = newHub.Rate
						existing.RateStr = newHub.RateStr
						existing.RateWidth = newHub.RateWidth
					}
					e.ActiveHubs = append(e.ActiveHubs, existing)
				}
				for cc, vh := range e.VisualHubs {
					if !vh.Active {
						delete(e.VisualHubs, cc)
					}
				}

				e.prefixCounts = prefixCounts
				e.ActiveASNImpacts = activeASNImpacts
				e.metricsMu.Unlock()

				clear(countryActivity)
				for _, prefixes := range currentAnomalies {
					clear(prefixes)
				}
				clear(currentAnomalies)
				clear(visualImpact)

				continue
			}

			ev := msg.ev
			if ev.prefix != "" {
				if ev.asn != 0 {
					prefixToASN[ev.prefix] = ev.asn
				}
				prefixToClassification[ev.prefix] = ev.classificationType

				if ev.eventType == EventNew || ev.eventType == EventUpdate || ev.eventType == EventGossip {
					if prefixes, ok := currentAnomalies[ClassificationOutage]; ok {
						delete(prefixes, ev.prefix)
					}
				}
				if actualType, ok := prefixToClassification[ev.prefix]; ok {
					if currentAnomalies[actualType] == nil {
						currentAnomalies[actualType] = make(map[string]int)
					}
					currentAnomalies[actualType][ev.prefix]++
				}

				visI, ok := visualImpact[ev.prefix]
				if !ok {
					visI = &VisualImpact{Prefix: ev.prefix, CCs: make(map[string]struct{})}
					visualImpact[ev.prefix] = visI
				}
				if visI.CCs == nil {
					visI.CCs = make(map[string]struct{})
				}
				if ev.cc != "" {
					visI.CCs[ev.cc] = struct{}{}
				}
				if msg.name != "" {
					if e.GetPriority(msg.name) >= e.GetPriority(visI.ClassificationName) {
						visI.ClassificationName = msg.name
						visI.ClassificationColor = msg.c
						if ev.leakDetail != nil {
							visI.LeakType = ev.leakDetail.Type
							visI.LeakerASN = ev.leakDetail.LeakerASN
							visI.VictimASN = ev.leakDetail.VictimASN
						}
					}
				}
			}
			if ev.cc != "" {
				countryActivity[ev.cc]++
			}
		}
	}
}
