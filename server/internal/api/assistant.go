package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"walkover/server/internal/models"
)

// The policymaker assistant answers questions about what is on screen. It is
// deliberately constrained: it reads the evidence the platform already holds
// and explains it. It does not decide anything, and it is told in its own
// instructions never to recommend approval, funding or closure — those remain
// human actions recorded through /decisions.

const assistantTimeout = 25 * time.Second

// maxEvidenceSignals bounds how many citizen reports are quoted into the
// prompt. Enough to show multilingual corroboration, small enough to keep the
// request fast.
const maxEvidenceSignals = 6

// AssistantRequest is a question asked from the dashboard's help panel.
type AssistantRequest struct {
	Question  string `json:"question" binding:"required"`
	ClusterID string `json:"cluster_id,omitempty"`
	WardID    string `json:"ward_id,omitempty"`
}

// AssistantResponse carries the answer plus what it was grounded on, so a
// policymaker can check the claim against the same evidence.
type AssistantResponse struct {
	Answer     string   `json:"answer"`
	GroundedOn []string `json:"grounded_on"`
	Source     string   `json:"source"` // "gemini" or "offline"
}

// AskAssistant answers a free-text question about the current evidence.
func (h *Handler) AskAssistant(c *gin.Context) {
	var req AssistantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "A question is required: " + err.Error(),
		})
		return
	}

	question := strings.TrimSpace(req.Question)
	if question == "" {
		c.JSON(http.StatusBadRequest, models.ApiResponse{
			Success: false,
			Error:   "A question is required",
		})
		return
	}

	evidence, grounded := h.buildEvidence(req.ClusterID, req.WardID)

	answer, source := "", "offline"
	if h.extractor != nil && !h.extractor.IsOffline() {
		ctx, cancel := context.WithTimeout(c.Request.Context(), assistantTimeout)
		defer cancel()

		generated, err := h.extractor.Ask(ctx, assistantPrompt(question, evidence))
		if err != nil {
			log.Printf("[ASSISTANT] Gemini unavailable, answering from local evidence: %v", err)
		} else {
			answer, source = generated, "gemini"
		}
	}

	// Never leave the panel empty: a deterministic explanation of the same
	// evidence is more useful than an error, and it keeps the demo alive if
	// the model is unreachable.
	if answer == "" {
		answer = offlineAnswer(question, h.lookupCluster(req.ClusterID), h.lookupWard(req.WardID))
	}

	c.JSON(http.StatusOK, models.ApiResponse{
		Success: true,
		Data: AssistantResponse{
			Answer:     answer,
			GroundedOn: grounded,
			Source:     source,
		},
	})
}

// assistantPrompt wraps the question in instructions that keep the assistant
// inside its remit.
func assistantPrompt(question, evidence string) string {
	return fmt.Sprintf(`You are the analyst assistant inside a public development intelligence platform used by Indian municipal policymakers in Indore.

Your job is to help an official READ AND UNDERSTAND the evidence below. You are not a decision maker.

Rules you must follow:
1. Answer only from the EVIDENCE section. Never invent numbers, wards, departments or citizen reports.
2. If the evidence does not answer the question, say so plainly and state what would need to be checked.
3. Never recommend approving, funding, sanctioning or closing anything. You may recommend investigation, field verification, outreach or which department should look at it. Final decisions belong to the official.
4. Treat low complaint volume in a poorly served ward as a possible blind spot — a reason for outreach, never evidence that there is no need.
5. Answer in plain language in at most 120 words. No markdown headings, no bullet symbols.

EVIDENCE:
%s

OFFICIAL'S QUESTION: %s

ANSWER:`, evidence, question)
}

// buildEvidence assembles the grounded context and a human-readable list of
// what that context was drawn from.
func (h *Handler) buildEvidence(clusterID, wardID string) (string, []string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var evidence strings.Builder
	var grounded []string

	blindSpots := make([]string, 0)
	for _, ward := range h.wards {
		if ward.IsBlindSpot {
			blindSpots = append(blindSpots, ward.Name)
		}
	}

	fmt.Fprintf(&evidence, "PLATFORM OVERVIEW\n- %d wards monitored, %d active issue clusters, %d recent citizen signals.\n",
		len(h.wards), len(h.clusters), len(h.signals))
	if len(blindSpots) > 0 {
		fmt.Fprintf(&evidence, "- Civic data blind spots (poor infrastructure, no reports): %s\n", strings.Join(blindSpots, "; "))
	}
	grounded = append(grounded, fmt.Sprintf("%d clusters, %d wards", len(h.clusters), len(h.wards)))

	if cluster := h.findClusterLocked(clusterID); cluster != nil {
		fmt.Fprintf(&evidence, "\nSELECTED ISSUE CLUSTER\n- Title: %s\n- Ward: %s\n- Responsible department: %s\n- Corroborating citizen signals: %d, arriving via %s\n- Urgency: %.0f/100 (%s, response target %d hours)\n- Why it scored that way: %s\n- Need %.1f | Confidence %.1f | Equity %.1f | Actionability %.1f (shown separately, never combined)\n- Current official status: %s\n- Existing recommendation: %s\n",
			cluster.Title, cluster.WardName, cluster.Department, cluster.SignalCount,
			strings.Join(cluster.Channels, ", "), cluster.Urgency.Score, cluster.Urgency.Tier, cluster.Urgency.SLAHours,
			strings.Join(cluster.Urgency.Factors, "; "),
			cluster.Scores.Need, cluster.Scores.Confidence, cluster.Scores.Equity, cluster.Scores.Actionability,
			cluster.Status, cluster.Recommendation)
		grounded = append(grounded, cluster.Title)

		if quoted := h.quoteSignalsLocked(cluster.SignalIDs); quoted != "" {
			fmt.Fprintf(&evidence, "- Citizen reports behind this cluster:\n%s", quoted)
			grounded = append(grounded, "citizen reports")
		}
	}

	if ward := h.findWardLocked(wardID); ward != nil {
		fmt.Fprintf(&evidence, "\nSELECTED WARD\n- %s (%s)\n- Infrastructure index %.2f (0 worst, 1 best), population %d, historical spend Rs %.1f crore\n- Critical facilities: %s\n- Active clusters: %d. Flagged as a civic data blind spot: %t\n",
			ward.Name, ward.Zone, ward.InfraIndex, ward.Population, ward.HistoricalSpend,
			strings.Join(ward.CriticalFacilities, ", "), ward.ActiveClusterCount, ward.IsBlindSpot)
		grounded = append(grounded, ward.Name)
	}

	if clusterID == "" && wardID == "" && len(h.clusters) > 0 {
		evidence.WriteString("\nALL ACTIVE CLUSTERS\n")
		for _, cluster := range h.clusters {
			fmt.Fprintf(&evidence, "- %s in %s: %d signals, urgency %.0f (%s), need %.1f equity %.1f, status %s\n",
				cluster.Title, cluster.WardName, cluster.SignalCount, cluster.Urgency.Score,
				cluster.Urgency.Tier, cluster.Scores.Need, cluster.Scores.Equity, cluster.Status)
		}
	}

	return evidence.String(), grounded
}

// quoteSignalsLocked returns the citizen reports behind a cluster in their
// original language, which is what lets the assistant cite real evidence.
func (h *Handler) quoteSignalsLocked(signalIDs []string) string {
	wanted := make(map[string]bool, len(signalIDs))
	for _, id := range signalIDs {
		wanted[id] = true
	}

	var quoted strings.Builder
	count := 0
	for _, signal := range h.signals {
		if !wanted[signal.ID] {
			continue
		}
		fmt.Fprintf(&quoted, "    * [%s, %s] %s\n", signal.Provider, signal.Language, signal.RawText)
		if count++; count >= maxEvidenceSignals {
			break
		}
	}
	return quoted.String()
}

func (h *Handler) findClusterLocked(id string) *models.Cluster {
	if id == "" {
		return nil
	}
	for i := range h.clusters {
		if h.clusters[i].ID == id {
			return &h.clusters[i]
		}
	}
	return nil
}

func (h *Handler) findWardLocked(id string) *models.Ward {
	if id == "" {
		return nil
	}
	for i := range h.wards {
		if h.wards[i].ID == id {
			return &h.wards[i]
		}
	}
	return nil
}

func (h *Handler) lookupCluster(id string) *models.Cluster {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.findClusterLocked(id)
}

func (h *Handler) lookupWard(id string) *models.Ward {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.findWardLocked(id)
}

// offlineAnswer explains the same evidence deterministically when Gemini is
// unreachable. It states its own limits rather than pretending to reason.
func offlineAnswer(question string, cluster *models.Cluster, ward *models.Ward) string {
	lowered := strings.ToLower(question)

	switch {
	case strings.Contains(lowered, "score") || strings.Contains(lowered, "need") || strings.Contains(lowered, "equity"):
		base := "The four dimensions are kept separate on purpose. Need reflects how serious and widespread the issue is, Confidence how strong the corroborating evidence is, Equity how underserved the ward is, and Actionability whether the issue is clear enough to act on. They are never combined into one score."
		if cluster != nil {
			return fmt.Sprintf("%s For %s: need %.1f, confidence %.1f, equity %.1f, actionability %.1f.",
				base, cluster.Title, cluster.Scores.Need, cluster.Scores.Confidence, cluster.Scores.Equity, cluster.Scores.Actionability)
		}
		return base

	case strings.Contains(lowered, "blind") || strings.Contains(lowered, "silent"):
		if ward != nil && ward.IsBlindSpot {
			return fmt.Sprintf("%s is flagged as a civic data blind spot: its infrastructure index is %.2f yet almost no citizen reports have arrived. That silence is a reason for outreach or field verification, not evidence that there is no need.", ward.Name, ward.InfraIndex)
		}
		return "A blind spot is a ward with poor infrastructure indicators and almost no citizen reports. Low participation can hide need, so these wards warrant outreach rather than being read as satisfied."

	case cluster != nil:
		return fmt.Sprintf("%s is visible because %d citizen signals were grouped into one issue in %s, arriving via %s. Urgency is %.0f/100 (%s). Key drivers: %s. Verify on site and confirm the responsible department before recording a decision — this is a recommendation for investigation, not an approval.",
			cluster.Title, cluster.SignalCount, cluster.WardName, strings.Join(cluster.Channels, ", "),
			cluster.Urgency.Score, cluster.Urgency.Tier, strings.Join(cluster.Urgency.Factors, "; "))

	case ward != nil:
		return fmt.Sprintf("%s has an infrastructure index of %.2f, a population of %d and historical spend of Rs %.1f crore, with %d active clusters. Compare participation against those indicators before concluding anything about need.",
			ward.Name, ward.InfraIndex, ward.Population, ward.HistoricalSpend, ward.ActiveClusterCount)

	default:
		return "Select a cluster or a ward and ask again, and I can explain the evidence behind it. The AI service is currently unreachable, so this is a direct reading of the stored data rather than a generated answer."
	}
}
