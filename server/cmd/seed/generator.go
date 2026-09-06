package main

import (
	"fmt"
	"time"

	"walkover/server/internal/clustering"
	"walkover/server/internal/db"
	"walkover/server/internal/models"
	"walkover/server/internal/urgency"
)

// This file builds the synthetic demo corpus for the platform.
//
// Department and hazard-tag values must come from the canonical vocabulary
// defined in the Gemini prompt (internal/extraction/gemini.go). A live
// extraction and a seeded cluster that label the same issue differently will
// not match on department, which is the fallback path when embeddings are
// unavailable.
//
// The corpus is engineered, not random. Three wards receive many corroborating
// reports in mixed languages and channels so that cross-channel clustering has
// something real to group, and two structurally underserved wards receive no
// reports at all so that civic blind-spot detection has something real to find.
// Everything else is background noise from unrelated issues.
//
// Extractions are written with an empty Embedding. Vectors are added either by
// the live pipeline as signals arrive, or by `cmd/seed --embed`, which
// backfills the seeded corpus so live signals can be matched against
// historical evidence by similarity.

// complaintSpec is one synthetic citizen report before it is expanded into a
// raw event, a canonical signal, and an AI extraction.
type complaintSpec struct {
	text         string
	language     string
	provider     models.ProviderType
	locationHint string
	summary      string // the English normalisation Gemini would produce
	intent       string
	minutesAgo   int
}

// issueGroup is a set of reports about the same underlying problem in one ward.
// Groups with a clusterID become demand hotspots; groups without one are
// background noise that must not cluster.
type issueGroup struct {
	clusterID   string
	wardID      string
	title       string
	description string
	department  string
	issue       string
	hazardTags  []string
	baseUrgency int
	status      string
	complaints  []complaintSpec
}

// Dataset is the full seeded state, ready to be written to any Repository.
type Dataset struct {
	Wards           []models.Ward
	RawEvents       []db.RawEvent
	Signals         []models.CitizenSignal
	Extractions     []models.AIExtraction
	Clusters        []models.Cluster
	Hotspots        []models.Cluster
	Recommendations []db.Recommendation
	AuditLogs       []models.AuditLog
}

// hotspotGroups are the engineered demand hotspots. Signal counts (8 / 12 / 6)
// and cluster IDs match the fixtures the dashboard already renders.
func hotspotGroups() []issueGroup {
	return []issueGroup{
		{
			clusterID:   "cluster-indore-001",
			wardID:      "indore-ward-02",
			title:       "Sewage Contamination in Main Drinking Line",
			description: "Multiple citizen reports describing brownish, foul-smelling tap water and sewage backflow near the street 4 community clinic.",
			department:  "Water Supply & Sewerage",
			issue:       "Contaminated drinking water supply",
			hazardTags:  []string{"CONTAMINATED_WATER", "SEWAGE_MIXING"},
			baseUrgency: 5,
			status:      "PENDING",
			complaints: []complaintSpec{
				{
					text:         "हमारे यहाँ नल से गंदा पानी आ रहा है बहुत बदबू है चंदन नगर गली 4",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Chandan Nagar Street 4",
					summary:      "Dirty, foul-smelling water from household tap in Chandan Nagar street 4.",
					intent:       "COMPLAINT",
					minutesAgo:   25,
				},
				{
					text:         "Urgent: Sewage mixing with drinking water line near clinic in Chandan nagar.",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Chandan Nagar Near Community Clinic",
					summary:      "Sewage suspected to be mixing into the drinking water line near the community clinic.",
					intent:       "COMPLAINT",
					minutesAgo:   15,
				},
				{
					text:         "Chandan Nagar gali 6 me nal ka pani peela aa raha hai, bachche bimar ho rahe hain",
					language:     "Hinglish",
					provider:     models.ProviderSMS,
					locationHint: "Chandan Nagar Street 6",
					summary:      "Yellow tap water in Chandan Nagar street 6; children in the household falling ill.",
					intent:       "COMPLAINT",
					minutesAgo:   48,
				},
				{
					text:         "चंदन नगर में पीने के पानी में सीवेज मिल रहा है, तुरंत जांच कराइए",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Chandan Nagar",
					summary:      "Request for immediate inspection of sewage contamination in the drinking water supply.",
					intent:       "REQUEST",
					minutesAgo:   72,
				},
				{
					text:         "Water from tap smells like drainage since 3 days, Chandan Nagar street 4 near community clinic",
					language:     "English",
					provider:     models.ProviderSMS,
					locationHint: "Chandan Nagar Street 4",
					summary:      "Tap water has smelled of drainage for three days near the community clinic.",
					intent:       "COMPLAINT",
					minutesAgo:   95,
				},
				{
					text:         "Bhai sahab pura mohalla bimar hai, Chandan Nagar ka pani ekdum ganda aa raha hai",
					language:     "Hinglish",
					provider:     models.ProviderWhatsApp,
					locationHint: "Chandan Nagar",
					summary:      "Widespread illness reported in the locality attributed to contaminated water supply.",
					intent:       "COMPLAINT",
					minutesAgo:   115,
				},
				{
					text:         "गली नंबर 7 चंदन नगर, नल के पानी में कीड़े दिख रहे हैं",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Chandan Nagar Street 7",
					summary:      "Visible worms/insects in tap water in Chandan Nagar street 7.",
					intent:       "COMPLAINT",
					minutesAgo:   190,
				},
				{
					text:         "Complaint: contaminated tap water causing stomach infection in Chandan Nagar ward 14",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Chandan Nagar Ward 14",
					summary:      "Stomach infections in the household linked to contaminated tap water.",
					intent:       "COMPLAINT",
					minutesAgo:   240,
				},
			},
		},
		{
			clusterID:   "cluster-indore-002",
			wardID:      "indore-ward-03",
			title:       "Major Road Caved-in on Hospital Approach Road",
			description: "Deep crater and road surface collapse near the BRTS intersection obstructing emergency ambulance ingress.",
			department:  "Public Works / Roads",
			issue:       "Arterial road cave-in blocking emergency access",
			hazardTags:  []string{"ROAD_CAVE_IN", "HOSPITAL_ROUTE_BLOCKED"},
			baseUrgency: 4,
			status:      "INVESTIGATING",
			complaints: []complaintSpec{
				{
					text:         "Vijay Nagar square ke paas ambulance route par road dhas gayi hai",
					language:     "Hinglish",
					provider:     models.ProviderSMS,
					locationHint: "Vijay Nagar Square",
					summary:      "Road has caved in on the ambulance route near Vijay Nagar square.",
					intent:       "COMPLAINT",
					minutesAgo:   40,
				},
				{
					text:         "Massive road cave-in near BRTS junction Vijay Nagar, ambulances taking long detour",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Vijay Nagar BRTS Junction",
					summary:      "Large road collapse at the BRTS junction forcing ambulances onto a long detour.",
					intent:       "COMPLAINT",
					minutesAgo:   55,
				},
				{
					text:         "विजय नगर चौराहे के पास सड़क धंस गई है, बड़ा गड्ढा बन गया है",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Vijay Nagar Square",
					summary:      "Road subsidence near Vijay Nagar square has formed a large crater.",
					intent:       "COMPLAINT",
					minutesAgo:   80,
				},
				{
					text:         "Road collapse on hospital approach road, Vijay Nagar. Two-wheeler fell yesterday night.",
					language:     "English",
					provider:     models.ProviderWeb,
					locationHint: "Vijay Nagar Hospital Approach Road",
					summary:      "Road collapse on the hospital approach road; a two-wheeler fell in overnight.",
					intent:       "COMPLAINT",
					minutesAgo:   110,
				},
				{
					text:         "Mother & Child hospital jane wali road puri tut gayi hai, emergency me problem ho rahi",
					language:     "Hinglish",
					provider:     models.ProviderWhatsApp,
					locationHint: "Vijay Nagar Hospital Road",
					summary:      "Road to the Mother & Child hospital is broken, delaying emergency access.",
					intent:       "COMPLAINT",
					minutesAgo:   160,
				},
				{
					text:         "सड़क में गहरा गड्ढा है विजय नगर बीआरटीएस के पास, रात में खतरनाक है",
					language:     "Hindi",
					provider:     models.ProviderSMS,
					locationHint: "Vijay Nagar BRTS",
					summary:      "Deep pit in the road near Vijay Nagar BRTS, dangerous after dark.",
					intent:       "OBSERVATION",
					minutesAgo:   220,
				},
				{
					text:         "Ambulance stuck for 15 minutes due to caved road near Vijay Nagar BRTS hub",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Vijay Nagar BRTS Hub",
					summary:      "Ambulance delayed 15 minutes by the caved-in road near the BRTS hub.",
					intent:       "OBSERVATION",
					minutesAgo:   300,
				},
				{
					text:         "Road cave in Vijay Nagar - 4 din se koi nahi aaya dekhne",
					language:     "Hinglish",
					provider:     models.ProviderWeb,
					locationHint: "Vijay Nagar",
					summary:      "Follow-up: no municipal inspection four days after the road cave-in was reported.",
					intent:       "FOLLOW_UP",
					minutesAgo:   420,
				},
				{
					text:         "विजय नगर में सड़क धंसने से ट्रैफिक जाम रहता है पूरे दिन",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Vijay Nagar",
					summary:      "Road subsidence in Vijay Nagar causing all-day traffic congestion.",
					intent:       "COMPLAINT",
					minutesAgo:   540,
				},
				{
					text:         "Deep crater on main road Vijay Nagar blocking emergency vehicle access to hospital",
					language:     "English",
					provider:     models.ProviderSMS,
					locationHint: "Vijay Nagar Main Road",
					summary:      "Crater on the main road blocking emergency vehicle access to the hospital.",
					intent:       "COMPLAINT",
					minutesAgo:   610,
				},
				{
					text:         "Bhai road ka bada hissa baith gaya hai Vijay Nagar square, koi barricade bhi nahi",
					language:     "Hinglish",
					provider:     models.ProviderWhatsApp,
					locationHint: "Vijay Nagar Square",
					summary:      "Large section of road has sunk at Vijay Nagar square with no barricading in place.",
					intent:       "COMPLAINT",
					minutesAgo:   700,
				},
				{
					text:         "Urgent repair needed: collapsed road surface obstructing ambulance route, Ward 22",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Ward 22 Vijay Nagar",
					summary:      "Request for urgent repair of the collapsed road obstructing the ambulance route.",
					intent:       "REQUEST",
					minutesAgo:   780,
				},
			},
		},
		{
			clusterID:   "cluster-indore-003",
			wardID:      "indore-ward-09",
			title:       "Uncovered Deep Drainage Manhole on School Path",
			description: "Drain cover broken during monsoon runoff; high hazard for pedestrians and school children.",
			department:  "Water Supply & Sewerage",
			issue:       "Uncovered drainage manhole in pedestrian path",
			hazardTags:  []string{"OPEN_MANHOLE"},
			baseUrgency: 5,
			status:      "PENDING",
			complaints: []complaintSpec{
				{
					text:         "खजराना में स्कूल के रास्ते पर मैनहोल खुला पड़ा है, बच्चे गिर सकते हैं",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Khajrana School Path",
					summary:      "Open manhole on the school route in Khajrana poses a fall risk to children.",
					intent:       "COMPLAINT",
					minutesAgo:   35,
				},
				{
					text:         "Khajrana main road pe manhole ka dhakkan tuta hua hai, raat me dikhta nahi",
					language:     "Hinglish",
					provider:     models.ProviderWhatsApp,
					locationHint: "Khajrana Main Road",
					summary:      "Broken manhole cover on Khajrana main road, invisible at night.",
					intent:       "COMPLAINT",
					minutesAgo:   70,
				},
				{
					text:         "Open manhole without cover near Khajrana school path, very dangerous for children",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Khajrana School Path",
					summary:      "Uncovered manhole near the school path described as dangerous for children.",
					intent:       "COMPLAINT",
					minutesAgo:   105,
				},
				{
					text:         "खजराना बस्ती में नाली का ढक्कन टूटा है, बारिश में पूरा भर जाता है",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Khajrana Basti",
					summary:      "Broken drain cover in Khajrana basti; the drain overflows during rain.",
					intent:       "COMPLAINT",
					minutesAgo:   260,
				},
				{
					text:         "Manhole khula pada hai Khajrana shrine ke paas, kal ek bacha girte girte bacha",
					language:     "Hinglish",
					provider:     models.ProviderWhatsApp,
					locationHint: "Khajrana Shrine Transit",
					summary:      "Open manhole near the Khajrana shrine; a child nearly fell in yesterday.",
					intent:       "COMPLAINT",
					minutesAgo:   480,
				},
				{
					text:         "Uncovered deep drainage manhole on school route, Khajrana ward 60. Needs immediate cover.",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Khajrana Ward 60",
					summary:      "Request to immediately cover the deep drainage manhole on the school route.",
					intent:       "REQUEST",
					minutesAgo:   900,
				},
			},
		},
	}
}

// backgroundGroups are unrelated single-issue reports spread across healthier
// wards. They give the corpus realistic noise and must not form hotspots.
func backgroundGroups() []issueGroup {
	return []issueGroup{
		{
			wardID:      "indore-ward-04",
			department:  "Electricity & Power",
			issue:       "Street light not working",
			hazardTags:  []string{},
			baseUrgency: 2,
			complaints: []complaintSpec{
				{
					text:         "Street light near Old Palasia crossing has not been working for a week",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Old Palasia Crossing",
					summary:      "Street light out for a week at the Old Palasia crossing.",
					intent:       "COMPLAINT",
					minutesAgo:   320,
				},
				{
					text:         "ओल्ड पलासिया में स्ट्रीट लाइट बंद है, शाम को अंधेरा रहता है",
					language:     "Hindi",
					provider:     models.ProviderSMS,
					locationHint: "Old Palasia",
					summary:      "Street lighting outage in Old Palasia leaving the area dark in the evening.",
					intent:       "COMPLAINT",
					minutesAgo:   1450,
				},
			},
		},
		{
			wardID:      "indore-ward-05",
			department:  "Sanitation & Solid Waste",
			issue:       "Garbage collection missed",
			hazardTags:  []string{},
			baseUrgency: 2,
			complaints: []complaintSpec{
				{
					text:         "Sarafa area me 3 din se kachra gaadi nahi aayi hai",
					language:     "Hinglish",
					provider:     models.ProviderWhatsApp,
					locationHint: "Sarafa Market",
					summary:      "Garbage collection vehicle has not visited the Sarafa area for three days.",
					intent:       "COMPLAINT",
					minutesAgo:   410,
				},
				{
					text:         "राजवाड़ा के पास कचरा जमा हो रहा है, बदबू आ रही है",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Rajwada",
					summary:      "Garbage accumulating near Rajwada causing a foul smell.",
					intent:       "COMPLAINT",
					minutesAgo:   1600,
				},
				{
					text:         "Stray dogs near Rajwada heritage market chasing customers in the evening",
					language:     "English",
					provider:     models.ProviderWeb,
					locationHint: "Rajwada Heritage Market",
					summary:      "Stray dogs near the heritage market chasing visitors in the evening.",
					intent:       "OBSERVATION",
					minutesAgo:   2100,
				},
			},
		},
		{
			wardID:      "indore-ward-06",
			department:  "Public Works / Roads",
			issue:       "Potholes on internal road",
			hazardTags:  []string{},
			baseUrgency: 2,
			complaints: []complaintSpec{
				{
					text:         "Bhawarkua main road pe chote chote gaddhe ho gaye hain",
					language:     "Hinglish",
					provider:     models.ProviderSMS,
					locationHint: "Bhawarkua Main Road",
					summary:      "Several small potholes have formed on the Bhawarkua main road.",
					intent:       "COMPLAINT",
					minutesAgo:   880,
				},
				{
					text:         "Potholes near DAVV campus gate need patching before monsoon",
					language:     "English",
					provider:     models.ProviderWeb,
					locationHint: "DAVV Campus Gate",
					summary:      "Request to patch potholes near the university campus gate before monsoon.",
					intent:       "REQUEST",
					minutesAgo:   2600,
				},
			},
		},
		{
			wardID:      "indore-ward-07",
			department:  "Public Health",
			issue:       "Public park maintenance",
			hazardTags:  []string{},
			baseUrgency: 1,
			complaints: []complaintSpec{
				{
					text:         "अन्नपूर्णा क्षेत्र के पार्क में झूले टूटे हुए हैं",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Annapurna Park",
					summary:      "Broken swings in the Annapurna area public park.",
					intent:       "COMPLAINT",
					minutesAgo:   1200,
				},
				{
					text:         "Park lights and benches near Annapurna temple need repair, many senior citizens visit",
					language:     "English",
					provider:     models.ProviderWeb,
					locationHint: "Annapurna Temple Zone",
					summary:      "Proposal to repair park lighting and benches used by senior citizens.",
					intent:       "PROPOSAL",
					minutesAgo:   3000,
				},
			},
		},
		{
			wardID:      "indore-ward-08",
			department:  "Water Supply & Sewerage",
			issue:       "Drain overflow during rain",
			hazardTags:  []string{},
			baseUrgency: 3,
			complaints: []complaintSpec{
				{
					text:         "Sudama Nagar me barish ke baad naali overflow ho jati hai",
					language:     "Hinglish",
					provider:     models.ProviderWhatsApp,
					locationHint: "Sudama Nagar",
					summary:      "Drain overflows after rainfall in Sudama Nagar.",
					intent:       "COMPLAINT",
					minutesAgo:   1010,
				},
				{
					text:         "सुदामा नगर की नाली जाम है, पानी सड़क पर आ रहा है",
					language:     "Hindi",
					provider:     models.ProviderSMS,
					locationHint: "Sudama Nagar",
					summary:      "Blocked drain in Sudama Nagar pushing water onto the road.",
					intent:       "COMPLAINT",
					minutesAgo:   2300,
				},
				{
					text:         "Loud construction noise after 11pm near Sudama Nagar community centre",
					language:     "English",
					provider:     models.ProviderWeb,
					locationHint: "Sudama Nagar Community Centre",
					summary:      "Late-night construction noise reported near the community centre.",
					intent:       "COMPLAINT",
					minutesAgo:   3400,
				},
			},
		},
		{
			wardID:      "indore-ward-10",
			department:  "Electricity & Power",
			issue:       "Street light not working",
			hazardTags:  []string{},
			baseUrgency: 2,
			complaints: []complaintSpec{
				{
					text:         "MR-10 ke paas street light band hai kai dino se",
					language:     "Hinglish",
					provider:     models.ProviderSMS,
					locationHint: "MR-10 Metro Pillar Area",
					summary:      "Street lights near MR-10 have been off for several days.",
					intent:       "COMPLAINT",
					minutesAgo:   1750,
				},
				{
					text:         "Sukhliya sector road lights flicker and go off at night",
					language:     "English",
					provider:     models.ProviderWhatsApp,
					locationHint: "Sukhliya Sector Road",
					summary:      "Flickering street lights that switch off at night in Sukhliya.",
					intent:       "OBSERVATION",
					minutesAgo:   2900,
				},
			},
		},
		{
			wardID:      "indore-ward-11",
			department:  "Traffic & Infrastructure",
			issue:       "Footpath encroachment",
			hazardTags:  []string{},
			baseUrgency: 2,
			complaints: []complaintSpec{
				{
					text:         "मलहारगंज मंडी के बाहर फुटपाथ पर कब्जा है, चलने की जगह नहीं",
					language:     "Hindi",
					provider:     models.ProviderWhatsApp,
					locationHint: "Malharganj Mandi",
					summary:      "Footpath outside the Malharganj mandi is encroached, leaving no walking space.",
					intent:       "COMPLAINT",
					minutesAgo:   1900,
				},
				{
					text:         "Malharganj grain market entry blocked by illegal parking every morning",
					language:     "English",
					provider:     models.ProviderWeb,
					locationHint: "Malharganj Grain Mandi",
					summary:      "Illegal parking blocks the grain market entry each morning.",
					intent:       "COMPLAINT",
					minutesAgo:   3600,
				},
			},
		},
	}
}

// BuildDataset expands the engineered groups into the full seeded state.
// now anchors every relative timestamp, so passing a fixed time yields a
// deterministic dataset.
func BuildDataset(wards []models.Ward, now time.Time) (*Dataset, error) {
	wardsByID := make(map[string]models.Ward, len(wards))
	for _, ward := range wards {
		wardsByID[ward.ID] = ward
	}

	urgencyEngine := urgency.NewEngine()
	clusterEngine := clustering.NewEngine(urgencyEngine)

	dataset := &Dataset{}
	signalSeq := 0

	groups := append(hotspotGroups(), backgroundGroups()...)
	for _, group := range groups {
		ward, ok := wardsByID[group.wardID]
		if !ok {
			return nil, fmt.Errorf("seed: group %q references unknown ward %q", group.issue, group.wardID)
		}

		var (
			signalIDs   []string
			channelSet  = map[string]bool{}
			channels    []string
			recentIn2h  int
			oldestAt    = now
			signalCount = len(group.complaints)
		)

		for _, complaint := range group.complaints {
			signalSeq++
			signalID := fmt.Sprintf("seed-sig-%04d", signalSeq)
			signalIDs = append(signalIDs, signalID)

			receivedAt := now.Add(-time.Duration(complaint.minutesAgo) * time.Minute)
			if receivedAt.Before(oldestAt) {
				oldestAt = receivedAt
			}
			if complaint.minutesAgo <= 120 {
				recentIn2h++
			}

			channel := string(complaint.provider)
			if !channelSet[channel] {
				channelSet[channel] = true
				channels = append(channels, channel)
			}

			signal := models.CitizenSignal{
				ID:           signalID,
				Provider:     complaint.provider,
				RawText:      complaint.text,
				Language:     complaint.language,
				LocationHint: complaint.locationHint,
				SenderPhone:  syntheticPhone(signalSeq),
				Timestamp:    receivedAt,
				Metadata: map[string]interface{}{
					"source":  "seed",
					"ward_id": ward.ID,
				},
			}
			dataset.Signals = append(dataset.Signals, signal)

			dataset.RawEvents = append(dataset.RawEvents, db.RawEvent{
				ID:       fmt.Sprintf("seed-evt-%04d", signalSeq),
				Provider: channel,
				SignalID: signalID,
				Payload: models.ViasocketPayload{
					EventID:   fmt.Sprintf("seed-evt-%04d", signalSeq),
					Provider:  channel,
					Sender:    signal.SenderPhone,
					Body:      complaint.text,
					Timestamp: receivedAt.Format(time.RFC3339),
					Metadata: map[string]interface{}{
						"source": "seed",
					},
				},
				ReceivedAt: receivedAt,
			})

			// Embedding is intentionally nil — Workstream 2's Gemini pipeline
			// owns text-embedding-004 and backfills this field.
			dataset.Extractions = append(dataset.Extractions, models.AIExtraction{
				SignalID:        signalID,
				Issue:           group.issue,
				WardID:          ward.ID,
				WardName:        ward.Name,
				Department:      group.department,
				BaseUrgency:     group.baseUrgency,
				HazardTags:      group.hazardTags,
				Intent:          complaint.intent,
				Summary:         complaint.summary,
				ConfidenceScore: syntheticConfidence(signalSeq),
				CreatedAt:       receivedAt.Add(2 * time.Second),
			})
		}

		// Background groups are deliberately left unclustered: isolated reports
		// about unrelated issues are exactly what should not become a hotspot.
		if group.clusterID == "" {
			continue
		}

		urgencyResult := urgencyEngine.CalculateUrgency(
			group.baseUrgency,
			group.hazardTags,
			signalCount,
			recentIn2h,
			ward.CriticalFacilities,
			oldestAt,
		)
		scores := clusterEngine.ComputeFourDimensionalScores(
			signalCount,
			urgencyResult.Score,
			len(channels),
			ward.InfraIndex,
			true, // location resolved to a ward
			true, // department mapped
		)
		recommendationText := clusterEngine.GenerateGroundedRecommendation(
			group.department,
			ward.Name,
			group.title,
			signalCount,
			urgencyResult,
			channels,
		)

		cluster := models.Cluster{
			ID:             group.clusterID,
			WardID:         ward.ID,
			WardName:       ward.Name,
			Department:     group.department,
			Title:          group.title,
			Description:    group.description,
			SignalCount:    signalCount,
			SignalIDs:      signalIDs,
			Channels:       channels,
			Urgency:        urgencyResult,
			Scores:         scores,
			Recommendation: recommendationText,
			IsBlindSpot:    false,
			Status:         group.status,
			CentroidLat:    ward.Lat,
			CentroidLng:    ward.Lng,
			CreatedAt:      oldestAt,
			UpdatedAt:      now,
		}

		dataset.Clusters = append(dataset.Clusters, cluster)
		dataset.Hotspots = append(dataset.Hotspots, cluster)
		dataset.Recommendations = append(dataset.Recommendations, db.Recommendation{
			ID:                "seed-rec-" + group.clusterID,
			ClusterID:         group.clusterID,
			Department:        group.department,
			Text:              recommendationText,
			EvidenceSignalIDs: signalIDs,
			CreatedAt:         now,
		})

		// A cluster already under investigation must have a human decision
		// behind it — the platform never moves a cluster itself.
		if group.status == "INVESTIGATING" {
			dataset.AuditLogs = append(dataset.AuditLogs, models.AuditLog{
				ID:        "seed-audit-" + group.clusterID,
				ClusterID: group.clusterID,
				Action:    "INVESTIGATE",
				Officer:   "Zonal Officer (" + ward.Zone + ")",
				Notes:     "Field inspection team assigned after reviewing corroborating citizen reports.",
				Timestamp: now.Add(-1 * time.Hour),
			})
		}
	}

	dataset.Wards = clusterEngine.DetectBlindSpots(wards, dataset.Clusters)
	return dataset, nil
}

// syntheticPhone returns an obviously fake sender number so no real subscriber
// number ever enters the seeded corpus.
func syntheticPhone(seq int) string {
	return fmt.Sprintf("+91-90000-%05d", seq)
}

// syntheticConfidence spreads extraction confidence across a plausible band
// deterministically, standing in for Gemini's own confidence until the live
// pipeline overwrites it.
func syntheticConfidence(seq int) float64 {
	return 0.82 + float64(seq%13)/100.0
}
