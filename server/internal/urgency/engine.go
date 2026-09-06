package urgency

import (
	"math"
	"strings"
	"time"

	"walkover/server/internal/models"
)

// Engine calculates multi-factor urgency scores and SLA tiers
type Engine struct{}

// NewEngine creates a new Urgency Decision Engine
func NewEngine() *Engine {
	return &Engine{}
}

// CalculateUrgency computes dynamic urgency and explainability factors
func (e *Engine) CalculateUrgency(
	baseUrgency int, // 1 - 5 from Gemini
	hazardTags []string,
	signalCount int,
	recentSignalCountIn2Hours int,
	criticalFacilities []string,
	oldestSignalTime time.Time,
) models.UrgencyResult {
	var factors []string
	var hazardBoost float64 = 0.0

	// 1. Base Score calculation (scaled from 1-5 to 0-100)
	rawBase := float64(baseUrgency) * 20.0
	if rawBase <= 0 {
		rawBase = 40.0 // default medium
	}

	// 2. Deterministic Hazard Overrides & Boosts
	isHardCritical := false
	for _, tag := range hazardTags {
		normalized := strings.ToUpper(strings.TrimSpace(tag))
		switch {
		case strings.Contains(normalized, "LIVE_WIRE") || strings.Contains(normalized, "ELECTRIC"):
			hazardBoost += 35.0
			isHardCritical = true
			factors = append(factors, "Exposed live electrical wire hazard detected")
		case strings.Contains(normalized, "MANHOLE") || strings.Contains(normalized, "OPEN_DRAIN"):
			hazardBoost += 30.0
			isHardCritical = true
			factors = append(factors, "Uncovered open manhole in pedestrian path")
		case strings.Contains(normalized, "CONTAMINATED_WATER") || strings.Contains(normalized, "SEWAGE_MIXING"):
			hazardBoost += 30.0
			isHardCritical = true
			factors = append(factors, "Drinking water contamination / sewage mixing")
		case strings.Contains(normalized, "CAVE_IN") || strings.Contains(normalized, "STRUCTURAL_COLLAPSE"):
			hazardBoost += 25.0
			factors = append(factors, "Arterial road cave-in / structural collapse risk")
		case strings.Contains(normalized, "HOSPITAL") || strings.Contains(normalized, "EMERGENCY_ROUTE"):
			hazardBoost += 25.0
			factors = append(factors, "Emergency ambulance route / hospital access blocked")
		}
	}

	// 3. Temporal Velocity Rate (Spike in < 2 hours)
	var velocityBoost float64 = 0.0
	velocityRate := float64(recentSignalCountIn2Hours)
	if recentSignalCountIn2Hours >= 5 {
		velocityBoost = 20.0
		factors = append(factors, "High velocity spike (>5 citizen reports within 2 hours)")
	} else if recentSignalCountIn2Hours >= 3 {
		velocityBoost = 10.0
		factors = append(factors, "Emerging complaint surge (3-4 reports within 2 hours)")
	}

	// 4. Infrastructure / Spatial Sensitivity
	var sensitivityBoost float64 = 0.0
	for _, facility := range criticalFacilities {
		lowered := strings.ToLower(facility)
		if strings.Contains(lowered, "hospital") || strings.Contains(lowered, "clinic") {
			sensitivityBoost += 10.0
			factors = append(factors, "Proximity to health facility: "+facility)
			break
		} else if strings.Contains(lowered, "school") {
			sensitivityBoost += 5.0
			factors = append(factors, "Proximity to educational facility: "+facility)
			break
		}
	}

	// 5. Backlog Aging Escalation
	if !oldestSignalTime.IsZero() {
		daysPending := time.Since(oldestSignalTime).Hours() / 24.0
		if daysPending > 7.0 {
			agingBoost := math.Min(15.0, daysPending*1.5)
			rawBase += agingBoost
			factors = append(factors, "Aging backlog escalation (>7 days pending)")
		}
	}

	// Compute Total Combined Score
	totalScore := rawBase + hazardBoost + velocityBoost + sensitivityBoost

	if isHardCritical && totalScore < 85.0 {
		totalScore = 85.0 // Force Tier 1 for life-safety triggers
	}

	if totalScore > 100.0 {
		totalScore = 100.0
	}
	if totalScore < 0.0 {
		totalScore = 15.0
	}

	// Determine Urgency Tier and SLA
	var tier models.UrgencyTier
	var slaHours int

	switch {
	case totalScore >= 80.0:
		tier = models.Tier1Critical
		slaHours = 4
	case totalScore >= 60.0:
		tier = models.Tier2High
		slaHours = 24
	case totalScore >= 35.0:
		tier = models.Tier3Medium
		slaHours = 72
	default:
		tier = models.Tier4Routine
		slaHours = 168 // 7 days
	}

	if len(factors) == 0 {
		factors = append(factors, "Standard citizen report evaluated by Gemini NLU")
	}

	// Two hazard tags can map to the same explanation — CONTAMINATED_WATER and
	// SEWAGE_MIXING share one branch — which would repeat a line in the
	// policymaker's explainability panel and collide as a React key.
	factors = dedupe(factors)

	return models.UrgencyResult{
		Score:        math.Round(totalScore*10) / 10,
		Tier:         tier,
		SLAHours:     slaHours,
		Factors:      factors,
		HazardBoost:  hazardBoost,
		VelocityRate: velocityRate,
	}
}

// dedupe removes repeated factors while preserving the order they were derived
// in, so the explanation still reads as a sequence of findings.
func dedupe(values []string) []string {
	seen := make(map[string]bool, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}
