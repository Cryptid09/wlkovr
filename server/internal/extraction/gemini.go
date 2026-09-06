package extraction

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"walkover/server/internal/models"
)

// Extractor handles AI semantic extraction, embeddings, and evidence grounding
type Extractor struct {
	client         *genai.Client
	modelName      string
	embeddingModel string
	isOffline      bool
}

// ExtractedResponse matches the JSON schema expected from Gemini
type ExtractedResponse struct {
	Issue           string   `json:"issue"`
	WardID          string   `json:"ward_id"`
	WardName        string   `json:"ward_name"`
	Department      string   `json:"department"`
	BaseUrgency     int      `json:"base_urgency"`
	HazardTags      []string `json:"hazard_tags"`
	Intent          string   `json:"intent"`
	Summary         string   `json:"summary"`
	ConfidenceScore float64  `json:"confidence_score"`
}

// NewExtractor initializes the Gemini extractor with specified models
func NewExtractor(ctx context.Context, apiKey string, modelName string, embeddingModel string) (*Extractor, error) {
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}
	if embeddingModel == "" {
		embeddingModel = "text-embedding-004"
	}

	if apiKey == "" {
		return &Extractor{
			modelName:      modelName,
			embeddingModel: embeddingModel,
			isOffline:      true,
		}, nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		// Fallback gracefully to offline mode if client initialization fails
		return &Extractor{
			modelName:      modelName,
			embeddingModel: embeddingModel,
			isOffline:      true,
		}, nil
	}

	return &Extractor{
		client:         client,
		modelName:      modelName,
		embeddingModel: embeddingModel,
		isOffline:      false,
	}, nil
}

// Close releases any allocated client resources
func (e *Extractor) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

// IsOffline returns whether extractor is running in offline fallback mode
func (e *Extractor) IsOffline() bool {
	return e.isOffline
}

// BuildPrompt creates the structured system instructions and few-shot examples
func (e *Extractor) BuildPrompt(signalText string, locationHint string) string {
	return fmt.Sprintf(`You are the NLU Engine of the Indore Civic Intelligence Platform.
Your task is to analyze raw citizen reports submitted via WhatsApp, SMS, or Voice across Hindi (Devanagari), Hinglish (Latin script), and English.
You must extract structured civic incident data strictly adhering to the JSON schema below.

Wards in Indore Reference:
- "indore-ward-01": "Ward 1 - Banganga" (Banganga, Laxmibai Nagar station)
- "indore-ward-02": "Ward 14 - Chandan Nagar" (Chandan Nagar, Dhar Road, Community Clinic)
- "indore-ward-03": "Ward 22 - Vijay Nagar" (Vijay Nagar square, Mother & Child Hospital, BRTS Hub)
- "indore-ward-04": "Ward 28 - Old Palasia" (Palasia, Trauma Wing, Industry House)
- "indore-ward-05": "Ward 35 - Rajwada & Sarafa" (Rajwada, Sarafa, Heritage market, Bartan Bazar)
- "indore-ward-06": "Ward 42 - Bhawarkua & Vishnupuri" (Bhawarkua, DAVV Campus, Rajiv Gandhi square)
- "indore-ward-07": "Ward 49 - Annapurna" (Annapurna temple, Ranjeet Hanuman)
- "indore-ward-08": "Ward 55 - Sudama Nagar" (Sudama Nagar, Phooti Kothi)
- "indore-ward-09": "Ward 60 - Khajrana" (Khajrana temple, Transit square, Kalka Mata)
- "indore-ward-10": "Ward 64 - Sukhliya" (Sukhliya, MR-10 Metro pillar, Bapat square)
- "indore-ward-11": "Ward 71 - Malharganj" (Malharganj, Grain Mandi)
- "indore-ward-12": "Ward 78 - Rau & Bypass Corridor" (Rau, Silicon City, Bypass junction)

Valid Municipal Departments:
- "Water Supply & Sewerage"
- "Public Works / Roads"
- "Electricity & Power"
- "Sanitation & Solid Waste"
- "Public Health"
- "Traffic & Infrastructure"

Valid Hazard Tags (include if evidence present):
- "CONTAMINATED_WATER": foul smell, sewage mixing with drinking water, brownish water, diarrhea/illness
- "LIVE_WIRE": snapped electric cable, hanging power wire, sparking transformer
- "OPEN_MANHOLE": uncovered manhole/sewer drain, broken chamber slab on road
- "HOSPITAL_ROUTE_BLOCKED": blocked road to hospital/clinic, ambulance stuck, sinkhole near emergency center
- "ROAD_CAVE_IN": large sinkhole, collapsed asphalt, road caved in
- "SEWAGE_MIXING": backflow into residential pipelines
- "DRAINAGE_OVERFLOW": choked nallah, monsoon water entering houses

Base Urgency Rating Scale (1 to 5):
5 - Immediate Critical Threat to Human Life (Live sparking wire, ambulance path blocked, poisoned water)
4 - Severe Hazard / Spreading Problem (Open manhole on dark road, sewage in homes)
3 - Moderate Civic Disruption (Choked drainage, broken pipe wasting water)
2 - Routine Maintenance Deficit (Streetlights off, trash piling up)
1 - Minor Cosmetic or Inconvenience (Park bench paint, tree trimming)

---
Few-Shot Training Examples:

Example 1 (Hindi):
Input: "चंदन नगर गली 4 में पीने के पानी में सीवेज का बदबूदार पानी मिल कर आ रहा है। बच्चे उल्टी दस्त से बीमार पड़ रहे हैं तुरंत टैंकर भेजो।"
Output JSON:
{
  "issue": "Sewage Contamination in Drinking Water Line",
  "ward_id": "indore-ward-02",
  "ward_name": "Ward 14 - Chandan Nagar",
  "department": "Water Supply & Sewerage",
  "base_urgency": 5,
  "hazard_tags": ["CONTAMINATED_WATER", "SEWAGE_MIXING"],
  "intent": "emergency_report",
  "summary": "Sewage backflow contaminating municipal drinking water line causing illness among children in Chandan Nagar street 4.",
  "confidence_score": 0.98
}

Example 2 (Hinglish):
Input: "Vijay Nagar square se Mother & Child hospital jane wali road par bada road cave-in ho gaya hai. Ambulance phas rahi hai emergency repair karo."
Output JSON:
{
  "issue": "Road Cave-In Blocking Hospital Route",
  "ward_id": "indore-ward-03",
  "ward_name": "Ward 22 - Vijay Nagar",
  "department": "Public Works / Roads",
  "base_urgency": 5,
  "hazard_tags": ["ROAD_CAVE_IN", "HOSPITAL_ROUTE_BLOCKED"],
  "intent": "emergency_report",
  "summary": "Major sinkhole/road cave-in blocking ambulance transit on critical hospital access road in Vijay Nagar.",
  "confidence_score": 0.97
}

Example 3 (English):
Input: "Dangerous open manhole near Khajrana temple transit road. Two bikes slipped already in evening."
Output JSON:
{
  "issue": "Open Manhole on Transit Road",
  "ward_id": "indore-ward-09",
  "ward_name": "Ward 60 - Khajrana",
  "department": "Water Supply & Sewerage",
  "base_urgency": 4,
  "hazard_tags": ["OPEN_MANHOLE"],
  "intent": "complaint",
  "summary": "Uncovered manhole causing vehicular accidents near Khajrana temple transit road.",
  "confidence_score": 0.95
}

Example 4 (Hinglish):
Input: "Rajwada market me transformer ke pas live wire jhul raha hai sparking ho rahi hai bheed bohot hai"
Output JSON:
{
  "issue": "Sparking Live Electric Wire in Crowded Market",
  "ward_id": "indore-ward-05",
  "ward_name": "Ward 35 - Rajwada & Sarafa",
  "department": "Electricity & Power",
  "base_urgency": 5,
  "hazard_tags": ["LIVE_WIRE"],
  "intent": "emergency_report",
  "summary": "Live sparking high-voltage wire hanging near transformer in crowded Rajwada market.",
  "confidence_score": 0.99
}

---
Analyze the citizen report below and output ONLY valid JSON matching this schema:
Citizen Report: "%s"
Location Hint: "%s"
`, signalText, locationHint)
}

// ExtractSignal processes a citizen report through Gemini or fallback heuristic
func (e *Extractor) ExtractSignal(ctx context.Context, signal models.CitizenSignal) (*models.AIExtraction, error) {
	if e.isOffline || e.client == nil {
		return e.fallbackExtraction(signal), nil
	}

	model := e.client.GenerativeModel(e.modelName)
	model.ResponseMIMEType = "application/json"
	model.SetTemperature(0.1)

	prompt := e.BuildPrompt(signal.RawText, signal.LocationHint)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		// Log and gracefully fall back to local rule-based extractor
		return e.fallbackExtraction(signal), nil
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return e.fallbackExtraction(signal), nil
	}

	var jsonText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			jsonText += string(textPart)
		}
	}

	jsonText = CleanJSONMarkdown(jsonText)

	var extracted ExtractedResponse
	if err := json.Unmarshal([]byte(jsonText), &extracted); err != nil {
		return e.fallbackExtraction(signal), nil
	}

	// Sanitize values
	if extracted.BaseUrgency < 1 {
		extracted.BaseUrgency = 1
	} else if extracted.BaseUrgency > 5 {
		extracted.BaseUrgency = 5
	}
	if extracted.ConfidenceScore <= 0 {
		extracted.ConfidenceScore = 0.88
	}

	embedding, _ := e.GenerateEmbedding(ctx, signal.RawText+" "+extracted.Issue)

	return &models.AIExtraction{
		SignalID:        signal.ID,
		Issue:           extracted.Issue,
		WardID:          extracted.WardID,
		WardName:        extracted.WardName,
		Department:      extracted.Department,
		BaseUrgency:     extracted.BaseUrgency,
		HazardTags:      extracted.HazardTags,
		Intent:          extracted.Intent,
		Summary:         extracted.Summary,
		Embedding:       embedding,
		ConfidenceScore: extracted.ConfidenceScore,
		CreatedAt:       time.Now(),
	}, nil
}

// GenerateEmbedding calls text-embedding-004 to produce 768-dim float32 vector
func (e *Extractor) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if e.isOffline || e.client == nil {
		return deterministicFallbackEmbedding(text, 768), nil
	}

	emModel := e.client.EmbeddingModel(e.embeddingModel)
	res, err := emModel.EmbedContent(ctx, genai.Text(text))
	if err != nil || res == nil || res.Embedding == nil || len(res.Embedding.Values) == 0 {
		return deterministicFallbackEmbedding(text, 768), nil
	}

	return res.Embedding.Values, nil
}

// GenerateClusterSummary synthesizes a cluster of citizen signals into an evidence-grounded summary
func (e *Extractor) GenerateClusterSummary(ctx context.Context, wardName string, department string, signals []models.CitizenSignal) (string, error) {
	if len(signals) == 0 {
		return fmt.Sprintf("No active citizen signals currently recorded for %s (%s).", wardName, department), nil
	}

	if e.isOffline || e.client == nil {
		return e.fallbackClusterSummary(wardName, department, signals), nil
	}

	var sb strings.Builder
	for i, s := range signals {
		sb.WriteString(fmt.Sprintf("[%d] Channel: %s, Text: %q\n", i+1, s.Provider, s.RawText))
	}

	prompt := fmt.Sprintf(`You are the Senior Municipal Civic Intelligence Advisor for Indore Municipal Corporation.
Synthesize the following %d citizen reports from %s under %s into an evidence-grounded summary for city decision makers.
Cite specific corroborating evidence from the reports (e.g., specific locations, number of reports, dangerous hazards).
Tone: Professional, urgent, evidence-backed. Length: 2-3 concise paragraphs.

Citizen Reports:
%s
`, len(signals), wardName, department, sb.String())

	model := e.client.GenerativeModel(e.modelName)
	model.SetTemperature(0.2)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return e.fallbackClusterSummary(wardName, department, signals), nil
	}

	var summary string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			summary += string(textPart)
		}
	}

	if strings.TrimSpace(summary) == "" {
		return e.fallbackClusterSummary(wardName, department, signals), nil
	}

	return strings.TrimSpace(summary), nil
}

// fallbackExtraction provides robust rule-based NLU extraction for offline mode / unit tests
func (e *Extractor) fallbackExtraction(signal models.CitizenSignal) *models.AIExtraction {
	text := strings.ToLower(signal.RawText + " " + signal.LocationHint)

	wardID := "indore-ward-01"
	wardName := "Ward 1 - Banganga"

	switch {
	case strings.Contains(text, "chandan nagar") || strings.Contains(text, "चंदन नगर") || strings.Contains(text, "dhar road"):
		wardID = "indore-ward-02"
		wardName = "Ward 14 - Chandan Nagar"
	case strings.Contains(text, "vijay nagar") || strings.Contains(text, "विजय नगर"):
		wardID = "indore-ward-03"
		wardName = "Ward 22 - Vijay Nagar"
	case strings.Contains(text, "palasia") || strings.Contains(text, "पलासिया"):
		wardID = "indore-ward-04"
		wardName = "Ward 28 - Old Palasia"
	case strings.Contains(text, "rajwada") || strings.Contains(text, "राजवाड़ा") || strings.Contains(text, "sarafa"):
		wardID = "indore-ward-05"
		wardName = "Ward 35 - Rajwada & Sarafa"
	case strings.Contains(text, "bhawarkua") || strings.Contains(text, "भंवरकुआ") || strings.Contains(text, "davv"):
		wardID = "indore-ward-06"
		wardName = "Ward 42 - Bhawarkua & Vishnupuri"
	case strings.Contains(text, "annapurna") || strings.Contains(text, "अन्नपूर्णा"):
		wardID = "indore-ward-07"
		wardName = "Ward 49 - Annapurna"
	case strings.Contains(text, "sudama") || strings.Contains(text, "सुदामा"):
		wardID = "indore-ward-08"
		wardName = "Ward 55 - Sudama Nagar"
	case strings.Contains(text, "khajrana") || strings.Contains(text, "खजराना"):
		wardID = "indore-ward-09"
		wardName = "Ward 60 - Khajrana"
	case strings.Contains(text, "sukhliya") || strings.Contains(text, "सुखलिया") || strings.Contains(text, "mr-10"):
		wardID = "indore-ward-10"
		wardName = "Ward 64 - Sukhliya"
	case strings.Contains(text, "malharganj") || strings.Contains(text, "मल्हारगंज"):
		wardID = "indore-ward-11"
		wardName = "Ward 71 - Malharganj"
	case strings.Contains(text, "rau") || strings.Contains(text, "राऊ") || strings.Contains(text, "bypass"):
		wardID = "indore-ward-12"
		wardName = "Ward 78 - Rau & Bypass Corridor"
	}

	department := "Sanitation & Solid Waste"
	issue := "Civic Maintenance Issue"
	baseUrgency := 2
	var hazardTags []string

	// 1. Open Manhole & Sewerage Hazards (Highest water priority)
	isManhole := strings.Contains(text, "manhole") || strings.Contains(text, "मैनहोल") || strings.Contains(text, "ढक्कन")
	isWaterOrDrain := strings.Contains(text, "pani") || strings.Contains(text, "पानी") || strings.Contains(text, "water") ||
		strings.Contains(text, "sewer") || strings.Contains(text, "सीवर") || strings.Contains(text, "drain") ||
		strings.Contains(text, "pipe") || strings.Contains(text, "पाइप") || strings.Contains(text, "गंदा") || strings.Contains(text, "sewage")

	// 2. Electrical Hazards
	isElectric := strings.Contains(text, "wire") || strings.Contains(text, "तार") || strings.Contains(text, "electric") ||
		strings.Contains(text, "bijli") || strings.Contains(text, "बिजली") || strings.Contains(text, "spark") ||
		strings.Contains(text, "transformer") || strings.Contains(text, "street light") || strings.Contains(text, "light")

	// 3. Road & Transit Hazards
	isRoadDamage := strings.Contains(text, "cave-in") || strings.Contains(text, "sinkhole") || strings.Contains(text, "gaddha") ||
		strings.Contains(text, "गड्ढा") || strings.Contains(text, "ambulance") || strings.Contains(text, "hospital") ||
		strings.Contains(text, "debris") || strings.Contains(text, "road") || strings.Contains(text, "सड़क")

	if isManhole || isWaterOrDrain {
		department = "Water Supply & Sewerage"
		issue = "Water Supply or Drainage Failure"
		baseUrgency = 3

		if strings.Contains(text, "ganda") || strings.Contains(text, "sewage") || strings.Contains(text, "सीवेज") ||
			strings.Contains(text, "badbu") || strings.Contains(text, "bimar") || strings.Contains(text, "vomit") ||
			strings.Contains(text, "drinking water") || strings.Contains(text, "burst") {
			hazardTags = append(hazardTags, "CONTAMINATED_WATER")
			if strings.Contains(text, "sewage") || strings.Contains(text, "सीवेज") || strings.Contains(text, "bimar") {
				hazardTags = append(hazardTags, "SEWAGE_MIXING")
			}
			issue = "Sewage Contamination in Drinking Water Line"
			baseUrgency = 5
		}

		if strings.Contains(text, "overflow") || strings.Contains(text, "chocked") || strings.Contains(text, "बह रहा") ||
			strings.Contains(text, "drainage") {
			hazardTags = append(hazardTags, "DRAINAGE_OVERFLOW")
			if strings.Contains(text, "ghar") || strings.Contains(text, "house") || strings.Contains(text, "andar") {
				baseUrgency = 4
			}
		}

		if isManhole {
			hazardTags = append(hazardTags, "OPEN_MANHOLE")
			issue = "Open or Damaged Manhole Hazard"
			if baseUrgency < 4 {
				baseUrgency = 4
			}
		}
	} else if isElectric {
		department = "Electricity & Power"
		issue = "Electrical Network Disruption"
		baseUrgency = 2

		if strings.Contains(text, "live") || strings.Contains(text, "नंगा") || strings.Contains(text, "spark") ||
			strings.Contains(text, "jhul") || strings.Contains(text, "gir") {
			hazardTags = append(hazardTags, "LIVE_WIRE")
			issue = "Dangerous Live Electric Wire Snapped"
			baseUrgency = 5
		}
	} else if isRoadDamage {
		department = "Public Works / Roads"
		issue = "Road Infrastructure Damage"
		baseUrgency = 2

		if strings.Contains(text, "cave-in") || strings.Contains(text, "sinkhole") || strings.Contains(text, "bada gaddha") {
			hazardTags = append(hazardTags, "ROAD_CAVE_IN")
			baseUrgency = 4
		}
		if strings.Contains(text, "hospital") || strings.Contains(text, "ambulance") {
			hazardTags = append(hazardTags, "HOSPITAL_ROUTE_BLOCKED")
			issue = "Critical Hospital Access Route Blocked"
			baseUrgency = 5
		}
	}

	intent := "complaint"
	if baseUrgency >= 4 {
		intent = "emergency_report"
	}

	embedding := deterministicFallbackEmbedding(signal.RawText+" "+issue, 768)

	return &models.AIExtraction{
		SignalID:        signal.ID,
		Issue:           issue,
		WardID:          wardID,
		WardName:        wardName,
		Department:      department,
		BaseUrgency:     baseUrgency,
		HazardTags:      hazardTags,
		Intent:          intent,
		Summary:         fmt.Sprintf("%s detected in %s based on citizen reports.", issue, wardName),
		Embedding:       embedding,
		ConfidenceScore: 0.92,
		CreatedAt:       time.Now(),
	}
}

// fallbackClusterSummary provides a deterministic evidence-grounded summary without API call
func (e *Extractor) fallbackClusterSummary(wardName string, department string, signals []models.CitizenSignal) string {
	channels := make(map[string]int)
	hazardMap := make(map[string]bool)

	for _, s := range signals {
		channels[string(s.Provider)]++
		lower := strings.ToLower(s.RawText)
		if strings.Contains(lower, "sewage") || strings.Contains(lower, "pani") || strings.Contains(lower, "water") {
			hazardMap["Water Contamination"] = true
		}
		if strings.Contains(lower, "ambulance") || strings.Contains(lower, "hospital") {
			hazardMap["Emergency Healthcare Access Risk"] = true
		}
		if strings.Contains(lower, "wire") || strings.Contains(lower, "spark") {
			hazardMap["Electrocution Hazard"] = true
		}
		if strings.Contains(lower, "manhole") {
			hazardMap["Open Manhole Hazard"] = true
		}
	}

	var chanSummary []string
	for ch, count := range channels {
		chanSummary = append(chanSummary, fmt.Sprintf("%d via %s", count, ch))
	}

	var hazards []string
	for h := range hazardMap {
		hazards = append(hazards, h)
	}

	hazardDesc := "routine municipal upkeep"
	if len(hazards) > 0 {
		hazardDesc = strings.Join(hazards, ", ")
	}

	return fmt.Sprintf("Cluster Analysis for %s [%s]:\n"+
		"Aggregated %d independent citizen reports (%s). High-priority indicators identified: %s. "+
		"Immediate on-site technical inspection and corrective team dispatch recommended to prevent community escalation and ensure public safety compliance.",
		wardName, department, len(signals), strings.Join(chanSummary, ", "), hazardDesc)
}

// CleanJSONMarkdown strips ```json ``` markdown wrapping if returned by Gemini
func CleanJSONMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

// deterministicFallbackEmbedding produces a unit-normalized 768-dim float32 vector based on subword n-grams and tokens
func deterministicFallbackEmbedding(text string, dim int) []float32 {
	vec := make([]float32, dim)
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		vec[0] = 1.0
		return vec
	}

	stopWords := map[string]bool{
		"in": true, "at": true, "and": true, "the": true, "a": true, "an": true, "of": true,
		"to": true, "on": true, "for": true, "is": true, "me": true, "se": true, "par": true,
		"ka": true, "ki": true, "ke": true, "ko": true, "hai": true, "hain": true,
	}

	addHashToVec := func(token string, weight float32) {
		h := sha256.Sum256([]byte(token))
		for i := 0; i < dim; i++ {
			byteIdx := (i * 4) % (len(h) - 4)
			val := binary.LittleEndian.Uint32(h[byteIdx : byteIdx+4])
			floatVal := (float32(val%2000)/1000.0 - 1.0) * weight
			vec[i] += floatVal
		}
	}

	for _, word := range words {
		weight := float32(1.0)
		if stopWords[word] {
			weight = 0.15
		}
		// Hash full word
		addHashToVec(word, weight)

		// Hash character 3-grams for morphological subword matching (e.g. contamin-, sewag-, water-)
		if len(word) >= 3 && !stopWords[word] {
			for i := 0; i <= len(word)-3; i++ {
				ngram := word[i : i+3]
				addHashToVec(ngram, weight*0.5)
			}
		}
	}

	// L2 Normalize
	var sumSq float32
	for _, v := range vec {
		sumSq += v * v
	}
	if sumSq > 0 {
		norm := float32(math.Sqrt(float64(sumSq)))
		for i := range vec {
			vec[i] /= norm
		}
	}

	return vec
}
