package clustering

import (
	"math"

	"walkover/server/internal/models"
	"walkover/server/internal/urgency"
)

// CosineSimilarity computes cosine similarity between two float32 vectors
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct float64 = 0.0
	var normA float64 = 0.0
	var normB float64 = 0.0

	for i := 0; i < len(a); i++ {
		valA := float64(a[i])
		valB := float64(b[i])
		dotProduct += valA * valB
		normA += valA * valA
		normB += valB * valB
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Engine performs clustering, 4D scoring, and blind spot detection
type Engine struct {
	urgencyEngine *urgency.Engine
}

// NewEngine creates a new clustering engine
func NewEngine(urgencyEngine *urgency.Engine) *Engine {
	return &Engine{
		urgencyEngine: urgencyEngine,
	}
}

// ComputeFourDimensionalScores calculates Need, Confidence, Equity, and Actionability (0 - 100 each)
func (e *Engine) ComputeFourDimensionalScores(
	signalCount int,
	urgencyScore float64,
	channelCount int,
	wardInfraIndex float64, // 0.0 - 1.0 (Lower = worse infra)
	isLocationResolved bool,
	isDepartmentMapped bool,
) models.FourDimensionalScore {
	// 1. Need Score: Volume normalized * Urgency
	volumeMultiplier := math.Min(1.0, float64(signalCount)/10.0)
	need := (urgencyScore * 0.7) + (volumeMultiplier * 30.0)
	if need > 100.0 {
		need = 100.0
	}

	// 2. Confidence Score: Channel diversity + count of independent citizen signals
	channelFactor := math.Min(1.0, float64(channelCount)/3.0) * 40.0
	signalFactor := math.Min(1.0, float64(signalCount)/8.0) * 60.0
	confidence := channelFactor + signalFactor
	if confidence > 100.0 {
		confidence = 100.0
	}

	// 3. Equity Score: Inverse of ward infra index (underserved areas get higher score)
	equity := (1.0 - wardInfraIndex) * 100.0
	if equity > 100.0 {
		equity = 100.0
	}
	if equity < 10.0 {
		equity = 10.0
	}

	// 4. Actionability Score: Location and department clarity heuristics
	actionability := 20.0
	if isLocationResolved {
		actionability += 40.0
	}
	if isDepartmentMapped {
		actionability += 40.0
	}

	return models.FourDimensionalScore{
		Need:          math.Round(need*10) / 10,
		Confidence:    math.Round(confidence*10) / 10,
		Equity:        math.Round(equity*10) / 10,
		Actionability: math.Round(actionability*10) / 10,
	}
}

// DetectBlindSpots evaluates all wards to flag silent, underserved areas
func (e *Engine) DetectBlindSpots(wards []models.Ward, activeClusters []models.Cluster) []models.Ward {
	wardClusterCounts := make(map[string]int)
	for _, cluster := range activeClusters {
		wardClusterCounts[cluster.WardID]++
	}

	var evaluatedWards []models.Ward
	for _, ward := range wards {
		count := wardClusterCounts[ward.ID]
		ward.ActiveClusterCount = count

		// Rule: Poor infrastructure (< 0.45) and complete silence = Blind Spot Candidate.
		// A ward that has reported anything is visible to us, so it is not a blind
		// spot however underserved it is — otherwise an underserved ward could be
		// shown as a demand hotspot and a blind spot at the same time.
		if ward.InfraIndex < 0.45 && count == 0 {
			ward.IsBlindSpot = true
		} else {
			ward.IsBlindSpot = false
		}
		evaluatedWards = append(evaluatedWards, ward)
	}

	return evaluatedWards
}

// GenerateGroundedRecommendation produces an evidence-grounded action brief
func (e *Engine) GenerateGroundedRecommendation(
	department string,
	wardName string,
	issueTitle string,
	signalCount int,
	urgency models.UrgencyResult,
	channels []string,
) string {
	var channelSummary string
	if len(channels) > 1 {
		channelSummary = "multi-channel corroboration across " + joinStrings(channels, " & ")
	} else if len(channels) == 1 {
		channelSummary = "reported via " + channels[0]
	} else {
		channelSummary = "citizen reports"
	}

	rec := "Evidence-Grounded Recommendation for " + department + ": " +
		"A " + string(urgency.Tier) + " priority incident ('" + issueTitle + "') has been identified in " + wardName +
		" with " + formatInt(signalCount) + " corroborating citizen signals (" + channelSummary + "). "

	if len(urgency.Factors) > 0 {
		rec += "Key Risk Drivers: " + urgency.Factors[0] + ". "
	}

	switch urgency.Tier {
	case models.Tier1Critical:
		rec += "Immediate Action Required: Dispatch emergency field inspection and remediation unit within " + formatInt(urgency.SLAHours) + " hours."
	case models.Tier2High:
		rec += "Recommended Action: Route high-priority work order to " + department + " engineering team within " + formatInt(urgency.SLAHours) + " hours."
	case models.Tier3Medium:
		rec += "Action: Schedule standard maintenance crew and notify ward supervisor."
	default:
		rec += "Action: Include in weekly municipal maintenance backlog."
	}

	return rec
}

func joinStrings(arr []string, sep string) string {
	res := ""
	for i, s := range arr {
		if i > 0 {
			res += sep
		}
		res += s
	}
	return res
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	res := ""
	for n > 0 {
		res = string(rune('0'+(n%10))) + res
		n /= 10
	}
	return res
}
