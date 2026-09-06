package api

import (
	"context"
	"sort"
	"strings"
	"time"

	"walkover/server/internal/models"
)

type LiveRepairReport struct {
	LiveSignals      int `json:"live_signals"`
	IncorrectMatches int `json:"incorrect_matches"`
	ClustersRebuilt  int `json:"clusters_rebuilt"`
	ClustersRemoved  int `json:"clusters_removed"`
}

// PreviewLiveAssignmentRepair reports incompatible live assignments without
// changing memory or persistence.
func (h *Handler) PreviewLiveAssignmentRepair() LiveRepairReport {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	report := LiveRepairReport{}
	for _, signal := range h.signals {
		if strings.HasPrefix(signal.ID, "seed-") {
			continue
		}
		extraction, ok := h.extractions[signal.ID]
		if !ok {
			continue
		}
		report.LiveSignals++
		cluster := h.clusterForSignalLocked(signal.ID)
		if cluster.ID == "" || !categoriesCompatible(clusterCategory(cluster), extractionCategory(&extraction)) {
			report.IncorrectMatches++
		}
	}
	return report
}

// RepairLiveAssignments removes only non-seed signals from cluster membership,
// rebuilds seeded cluster scores, and reattaches live evidence with the current
// compatibility rules. Citizen signals and extractions are never deleted.
func (h *Handler) RepairLiveAssignments(ctx context.Context) (LiveRepairReport, error) {
	report := h.PreviewLiveAssignmentRepair()
	h.mutex.Lock()
	live := make([]models.CitizenSignal, 0)
	liveIDs := map[string]bool{}
	signalByID := map[string]models.CitizenSignal{}
	for _, signal := range h.signals {
		signalByID[signal.ID] = signal
		if !strings.HasPrefix(signal.ID, "seed-") {
			if _, ok := h.extractions[signal.ID]; ok {
				live = append(live, signal)
				liveIDs[signal.ID] = true
			}
		}
	}

	removedIDs := make([]string, 0)
	rebuilt := make([]models.Cluster, 0, len(h.clusters))
	h.evidence = make(map[string]*clusterEvidence)
	for _, cluster := range h.clusters {
		kept := cluster.SignalIDs[:0]
		for _, id := range cluster.SignalIDs {
			if !liveIDs[id] {
				kept = append(kept, id)
			}
		}
		cluster.SignalIDs = append([]string(nil), kept...)
		cluster.SignalCount = len(cluster.SignalIDs)
		if cluster.SignalCount == 0 {
			removedIDs = append(removedIDs, cluster.ID)
			continue
		}
		cluster.Channels = nil
		for _, id := range cluster.SignalIDs {
			if signal, ok := signalByID[id]; ok {
				cluster.Channels = appendUnique(cluster.Channels, string(signal.Provider))
			}
			if extraction, ok := h.extractions[id]; ok {
				h.mergeEvidenceLocked(cluster.ID, &extraction)
			}
		}
		evidence := h.evidence[cluster.ID]
		ward := h.wardByIDLocked(cluster.WardID)
		if evidence != nil {
			cluster.Urgency = h.urgencyEngine.CalculateUrgency(evidence.maxBaseUrgency, evidence.hazardTags, cluster.SignalCount, h.recentSignalCountLocked(cluster.SignalIDs, 2*time.Hour), ward.CriticalFacilities, cluster.CreatedAt)
			cluster.Scores = h.clusterEngine.ComputeFourDimensionalScores(cluster.SignalCount, cluster.Urgency.Score, len(cluster.Channels), ward.InfraIndex, true, cluster.Department != "")
			cluster.Recommendation = h.clusterEngine.GenerateGroundedRecommendation(cluster.Department, cluster.WardName, cluster.Title, cluster.SignalCount, cluster.Urgency, cluster.Channels)
		}
		cluster.UpdatedAt = time.Now()
		rebuilt = append(rebuilt, cluster)
	}
	h.clusters = rebuilt
	h.mutex.Unlock()

	if h.repo != nil {
		for _, id := range removedIDs {
			if err := h.repo.DeleteCluster(ctx, id); err != nil {
				return report, err
			}
		}
		for _, cluster := range rebuilt {
			if err := h.repo.UpsertCluster(ctx, cluster); err != nil {
				return report, err
			}
			if err := h.repo.UpsertHotspot(ctx, cluster); err != nil {
				return report, err
			}
		}
	}

	// Oldest first gives a stable cluster identity and lets newer reports join it.
	sort.Slice(live, func(i, j int) bool { return live[i].Timestamp.Before(live[j].Timestamp) })
	for _, signal := range live {
		extraction := h.extractions[signal.ID]
		h.attachToCluster(ctx, signal, &extraction)
	}
	report.ClustersRebuilt = len(rebuilt)
	report.ClustersRemoved = len(removedIDs)
	return report, nil
}
