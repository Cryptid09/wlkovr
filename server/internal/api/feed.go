package api

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"walkover/server/internal/models"
)

// GetFeed returns a privacy-safe, fully joined operational feed. Unlike
// /signals, callers do not need to race separate extraction and cluster APIs.
func (h *Handler) GetFeed(c *gin.Context) {
	limit := 50
	if requested, err := strconv.Atoi(c.Query("limit")); err == nil && requested > 0 {
		if requested > 100 {
			requested = 100
		}
		limit = requested
	}

	h.mutex.RLock()
	items := make([]models.SignalFeedItem, 0, len(h.signals))
	for _, signal := range h.signals {
		extraction, ok := h.extractions[signal.ID]
		if !ok {
			continue
		}
		cluster := h.clusterForSignalLocked(signal.ID)
		items = append(items, buildFeedItem(signal, extraction, cluster))
	}
	h.mutex.RUnlock()

	sort.SliceStable(items, func(i, j int) bool { return items[i].Timestamp.After(items[j].Timestamp) })
	if len(items) > limit {
		items = items[:limit]
	}
	c.JSON(http.StatusOK, models.ApiResponse{Success: true, Data: items})
}

func (h *Handler) clusterForSignalLocked(signalID string) models.Cluster {
	for _, cluster := range h.clusters {
		for _, id := range cluster.SignalIDs {
			if id == signalID {
				return cluster
			}
		}
	}
	return models.Cluster{}
}

func buildFeedItem(signal models.CitizenSignal, extraction models.AIExtraction, cluster models.Cluster) models.SignalFeedItem {
	category := extractionCategory(&extraction)
	locationSource := extraction.LocationSource
	locationConfidence := extraction.LocationConfidence
	locationRationale := extraction.LocationRationale
	if locationSource == "" {
		locationSource, locationConfidence, locationRationale = "model_inferred", 0.55, "Resolved by the original AI extraction; send a precise landmark for confirmation"
	}
	analysisSource := extraction.AnalysisSource
	if analysisSource == "" {
		analysisSource = "gemini"
	}
	return models.SignalFeedItem{
		ID: signal.ID, Provider: signal.Provider, RawText: signal.RawText,
		Language: signal.Language, TranslatedText: signal.TranslatedText, Timestamp: signal.Timestamp,
		Status: "VERIFIED", Issue: extraction.Issue, IssueCategory: category,
		Department: extraction.Department, WardID: extraction.WardID, WardName: extraction.WardName,
		LocationSource: locationSource, LocationConfidence: locationConfidence,
		LocationRationale: locationRationale, BaseUrgency: extraction.BaseUrgency,
		Severity: severityLabel(extraction.BaseUrgency), HazardTags: extraction.HazardTags,
		Summary: extraction.Summary, AIConfidence: extraction.ConfidenceScore,
		AnalysisSource: analysisSource, ClusterID: cluster.ID, ClusterTitle: cluster.Title,
		ClusterStatus: cluster.Status, ClusterUrgency: cluster.Urgency, ClusterSignalCount: cluster.SignalCount,
	}
}

func severityLabel(base int) string {
	switch {
	case base >= 5:
		return "CRITICAL"
	case base == 4:
		return "HIGH"
	case base == 3:
		return "MEDIUM"
	default:
		return "ROUTINE"
	}
}
