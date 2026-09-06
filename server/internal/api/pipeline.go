package api

import (
	"context"
	"log"
	"time"

	"walkover/server/internal/clustering"
	"walkover/server/internal/db"
	"walkover/server/internal/extraction"
	"walkover/server/internal/models"
)

// This file wires the three isolated workstreams together:
//
//	viasocket webhook -> canonical signal -> Gemini extraction (Workstream 2)
//	                  -> Firestore persistence (Workstream 3)
//	                  -> cluster attachment and rescoring (Workstream 4)
//	                  -> live WebSocket broadcast (Workstream 1)
//
// Both dependencies are optional. With neither attached the handler keeps its
// original in-memory demo behaviour, which is what the existing handler tests
// exercise.

const (
	// clusterMatchThreshold is the mean cosine similarity above which a new
	// signal is treated as describing the same issue as an existing cluster.
	//
	// Calibrated against the seeded corpus embedded with gemini-embedding-001:
	// the lowest within-cluster mean is 0.804 and the highest mean for an
	// unrelated complaint is 0.691, so the midpoint sits at 0.75 with roughly
	// 0.11 of margin on either side. Re-measure with cmd/seed --embed if the
	// embedding model ever changes.
	clusterMatchThreshold = 0.75

	// pipelineTimeout bounds the asynchronous extraction and persistence work
	// so a slow Gemini call can never leak a goroutine.
	pipelineTimeout = 30 * time.Second
)

// WithPersistence attaches a repository and hydrates the in-memory state from
// it. Seeded wards, clusters and signals replace the built-in demo fixtures,
// but only when the store actually holds data — an empty store leaves the
// fixtures in place so the dashboard is never blank.
func (h *Handler) WithPersistence(ctx context.Context, repo db.Repository) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.repo = repo

	wards, err := repo.ListWards(ctx)
	if err != nil {
		return err
	}
	clusters, err := repo.ListClusters(ctx)
	if err != nil {
		return err
	}
	signals, err := repo.ListSignals(ctx, 50)
	if err != nil {
		return err
	}
	auditLogs, err := repo.ListAuditLogs(ctx, 50)
	if err != nil {
		return err
	}

	if len(wards) > 0 {
		h.wards = wards
	}
	if len(clusters) > 0 {
		h.clusters = clusters
	}
	if len(signals) > 0 {
		h.signals = newestFirst(signals)
	}
	if len(auditLogs) > 0 {
		h.auditLogs = auditLogs
	}

	// Embeddings let a live signal be matched against seeded evidence. They are
	// absent until the extraction pipeline backfills them, which is expected.
	extractions, err := repo.ListExtractions(ctx, 0)
	if err != nil {
		return err
	}

	clusterOfSignal := make(map[string]string)
	for _, cluster := range h.clusters {
		for _, signalID := range cluster.SignalIDs {
			clusterOfSignal[signalID] = cluster.ID
		}
	}

	for _, item := range extractions {
		if len(item.Embedding) > 0 {
			h.embeddings[item.SignalID] = item.Embedding
		}
		if clusterID, ok := clusterOfSignal[item.SignalID]; ok {
			h.mergeEvidenceLocked(clusterID, &item)
		}
	}

	log.Printf("[STORE] Hydrated from %s — %d wards, %d clusters, %d signals, %d embeddings",
		repo.Backend(), len(h.wards), len(h.clusters), len(h.signals), len(h.embeddings))
	return nil
}

// WithExtractor attaches the Gemini extraction pipeline. Without it, incoming
// signals are still ingested, persisted and broadcast — they simply are not
// understood or attached to a cluster.
func (h *Handler) WithExtractor(extractor *extraction.Extractor) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.extractor = extractor
	if extractor.IsOffline() {
		log.Printf("[GEMINI] Extractor attached in OFFLINE mode — no API key, deterministic fallbacks in use")
		return
	}
	log.Printf("[GEMINI] Extractor attached — live structured extraction and embeddings enabled")
}

// ingest records a newly received signal in memory and on the live feed, then
// hands the slow work to a background goroutine.
//
// The broadcast happens before extraction deliberately: the demo's headline
// claim is that a WhatsApp message reaches the dashboard in under two seconds,
// and a Gemini round trip takes longer than that. The dashboard receives a
// second event once the signal has been understood.
func (h *Handler) ingest(signal models.CitizenSignal, payload *models.ViasocketPayload) {
	h.mutex.Lock()
	h.signals = append([]models.CitizenSignal{signal}, h.signals...)
	if len(h.signals) > 50 {
		h.signals = h.signals[:50]
	}
	h.mutex.Unlock()

	h.hub.Broadcast("SIGNAL_RECEIVED", signal)

	if h.repo == nil && h.extractor == nil {
		return
	}
	go h.process(signal, payload)
}

// process runs extraction, persistence and cluster attachment off the request
// path.
func (h *Handler) process(signal models.CitizenSignal, payload *models.ViasocketPayload) {
	ctx, cancel := context.WithTimeout(context.Background(), pipelineTimeout)
	defer cancel()

	if h.repo != nil {
		if payload != nil {
			event := db.RawEvent{
				ID:         "evt-" + signal.ID,
				Provider:   string(signal.Provider),
				SignalID:   signal.ID,
				Payload:    *payload,
				ReceivedAt: signal.Timestamp,
			}
			if err := h.repo.SaveRawEvent(ctx, event); err != nil {
				log.Printf("[STORE] raw event %s: %v", event.ID, err)
			}
		}
		if err := h.repo.SaveSignal(ctx, signal); err != nil {
			log.Printf("[STORE] signal %s: %v", signal.ID, err)
		}
	}

	if h.extractor == nil {
		return
	}

	result, err := h.extractor.ExtractSignal(ctx, signal)
	if err != nil {
		log.Printf("[GEMINI] extraction failed for %s: %v", signal.ID, err)
		return
	}

	// An embedding failure is not fatal: the signal is still understood, it
	// just falls back to ward and department matching for clustering.
	embedding, err := h.extractor.GenerateEmbedding(ctx, db.EmbeddingText(*result))
	if err != nil {
		log.Printf("[GEMINI] embedding failed for %s: %v", signal.ID, err)
	} else {
		result.Embedding = embedding
	}

	if h.repo != nil {
		if err := h.repo.SaveExtraction(ctx, *result); err != nil {
			log.Printf("[STORE] extraction %s: %v", signal.ID, err)
		}
	}

	h.hub.Broadcast("SIGNAL_EXTRACTED", result)
	h.attachToCluster(ctx, signal, result)
}

// attachToCluster folds a newly understood signal into an existing cluster and
// rescores it, so the dashboard shows corroboration arriving in real time.
// A signal that matches nothing stays visible in the live feed — the platform
// never invents a cluster from a single report.
func (h *Handler) attachToCluster(ctx context.Context, signal models.CitizenSignal, result *models.AIExtraction) {
	h.mutex.Lock()

	if len(result.Embedding) > 0 {
		h.embeddings[signal.ID] = result.Embedding
	}

	index := h.matchClusterLocked(result)
	if index < 0 {
		h.mutex.Unlock()
		log.Printf("[CLUSTER] signal %s (%s / %s) matched no existing cluster", signal.ID, result.WardID, result.Department)
		return
	}

	cluster := h.clusters[index]
	cluster.SignalIDs = append(cluster.SignalIDs, signal.ID)
	cluster.SignalCount = len(cluster.SignalIDs)
	cluster.Channels = appendUnique(cluster.Channels, string(signal.Provider))
	cluster.UpdatedAt = time.Now()

	ward := h.wardByIDLocked(cluster.WardID)
	recent := h.recentSignalCountLocked(cluster.SignalIDs, 2*time.Hour)

	// Urgency is driven by the strongest evidence anyone has reported for this
	// cluster, not by the newest message. A citizen following up with a polite
	// "please fix this" must never de-escalate a contaminated water supply.
	h.mergeEvidenceLocked(cluster.ID, result)
	evidence := h.evidence[cluster.ID]

	cluster.Urgency = h.urgencyEngine.CalculateUrgency(
		evidence.maxBaseUrgency,
		evidence.hazardTags,
		cluster.SignalCount,
		recent,
		ward.CriticalFacilities,
		cluster.CreatedAt,
	)
	cluster.Scores = h.clusterEngine.ComputeFourDimensionalScores(
		cluster.SignalCount,
		cluster.Urgency.Score,
		len(cluster.Channels),
		ward.InfraIndex,
		result.WardID != "",
		result.Department != "",
	)
	cluster.Recommendation = h.clusterEngine.GenerateGroundedRecommendation(
		cluster.Department,
		cluster.WardName,
		cluster.Title,
		cluster.SignalCount,
		cluster.Urgency,
		cluster.Channels,
	)

	h.clusters[index] = cluster
	h.wards = h.clusterEngine.DetectBlindSpots(h.wards, h.clusters)
	wards := h.wards
	h.mutex.Unlock()

	if h.repo != nil {
		if err := h.repo.UpsertCluster(ctx, cluster); err != nil {
			log.Printf("[STORE] cluster %s: %v", cluster.ID, err)
		}
		if err := h.repo.UpsertHotspot(ctx, cluster); err != nil {
			log.Printf("[STORE] hotspot %s: %v", cluster.ID, err)
		}
		if err := h.repo.UpsertWards(ctx, wards); err != nil {
			log.Printf("[STORE] wards: %v", err)
		}
		recommendation := db.Recommendation{
			ID:                "rec-" + cluster.ID,
			ClusterID:         cluster.ID,
			Department:        cluster.Department,
			Text:              cluster.Recommendation,
			EvidenceSignalIDs: cluster.SignalIDs,
			CreatedAt:         cluster.UpdatedAt,
		}
		if err := h.repo.SaveRecommendation(ctx, recommendation); err != nil {
			log.Printf("[STORE] recommendation %s: %v", recommendation.ID, err)
		}
	}

	log.Printf("[CLUSTER] signal %s joined %s — now %d signals, urgency %.1f (%s)",
		signal.ID, cluster.ID, cluster.SignalCount, cluster.Urgency.Score, cluster.Urgency.Tier)

	h.hub.Broadcast("CLUSTER_UPDATED", cluster)
}

// matchClusterLocked returns the index of the cluster a signal belongs to, or
// -1 when it belongs to none. The caller holds the mutex.
//
// A signal must be in the same ward — two identical complaints in different
// wards are different development needs. Within a ward, a match needs either
// semantic similarity above the threshold or an exact department match.
//
// Department alone is not enough on its own terms: Gemini classifies freely
// and does not always reproduce the seeded taxonomy (an open manhole came back
// as "Water Supply & Sewerage" against a "Sanitation & Drainage" cluster).
// Semantic similarity is the more reliable signal wherever embeddings exist,
// and department equality is the fallback for extractions that have none.
func (h *Handler) matchClusterLocked(result *models.AIExtraction) int {
	best := -1
	bestSimilarity := clusterMatchThreshold
	departmentFallback := -1

	for i, cluster := range h.clusters {
		if cluster.WardID != result.WardID {
			continue
		}

		similarity, comparable := h.clusterSimilarityLocked(cluster, result.Embedding)
		if comparable && similarity >= bestSimilarity {
			best = i
			bestSimilarity = similarity
			continue
		}

		if departmentFallback < 0 && cluster.Department == result.Department {
			departmentFallback = i
		}
	}

	if best >= 0 {
		return best
	}
	return departmentFallback
}

// clusterSimilarityLocked returns the mean cosine similarity between a
// candidate embedding and the cluster's known signal embeddings, reporting
// whether any comparison was possible at all.
//
// Mean rather than max, measured against the seeded corpus with
// gemini-embedding-001: mean separates cleanly (same issue 0.80–0.90,
// unrelated 0.61–0.69) whereas the maxima overlap — background noise reaches
// 0.80 against a cluster while a genuine member pair can sit at 0.72. One
// coincidentally close sentence should not pull an unrelated complaint in.
func (h *Handler) clusterSimilarityLocked(cluster models.Cluster, candidate []float32) (float64, bool) {
	if len(candidate) == 0 {
		return 0, false
	}

	total := 0.0
	compared := 0
	for _, signalID := range cluster.SignalIDs {
		known, ok := h.embeddings[signalID]
		if !ok || len(known) == 0 {
			continue
		}
		total += clustering.CosineSimilarity(candidate, known)
		compared++
	}

	if compared == 0 {
		return 0, false
	}
	return total / float64(compared), true
}

func (h *Handler) wardByIDLocked(wardID string) models.Ward {
	for _, ward := range h.wards {
		if ward.ID == wardID {
			return ward
		}
	}
	return models.Ward{}
}

// recentSignalCountLocked counts how many of a cluster's signals arrived inside
// the window, feeding the urgency engine's velocity factor.
func (h *Handler) recentSignalCountLocked(signalIDs []string, window time.Duration) int {
	cutoff := time.Now().Add(-window)
	ids := make(map[string]bool, len(signalIDs))
	for _, id := range signalIDs {
		ids[id] = true
	}

	count := 0
	for _, signal := range h.signals {
		if ids[signal.ID] && signal.Timestamp.After(cutoff) {
			count++
		}
	}
	return count
}

// persistDecision writes a policymaker's decision and the cluster it changed.
// Decisions are always human-initiated; this only records what was decided.
func (h *Handler) persistDecision(cluster models.Cluster, entry models.AuditLog) {
	if h.repo == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), pipelineTimeout)
		defer cancel()

		if err := h.repo.AppendAuditLog(ctx, entry); err != nil {
			log.Printf("[STORE] audit log %s: %v", entry.ID, err)
		}
		if err := h.repo.UpsertCluster(ctx, cluster); err != nil {
			log.Printf("[STORE] cluster %s: %v", cluster.ID, err)
		}
	}()
}

// clusterEvidence is the running high-water mark of what a cluster's citizen
// reports have collectively established.
type clusterEvidence struct {
	maxBaseUrgency int
	hazardTags     []string
}

// mergeEvidenceLocked folds one extraction into a cluster's aggregate evidence.
// The caller holds the mutex.
func (h *Handler) mergeEvidenceLocked(clusterID string, result *models.AIExtraction) {
	current, ok := h.evidence[clusterID]
	if !ok {
		current = &clusterEvidence{}
		h.evidence[clusterID] = current
	}

	if result.BaseUrgency > current.maxBaseUrgency {
		current.maxBaseUrgency = result.BaseUrgency
	}
	for _, tag := range result.HazardTags {
		current.hazardTags = appendUnique(current.hazardTags, tag)
	}
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

// newestFirst orders signals for the live feed, which shows the most recent
// report at the top. The repository returns documents in document-ID order.
func newestFirst(signals []models.CitizenSignal) []models.CitizenSignal {
	ordered := make([]models.CitizenSignal, len(signals))
	copy(ordered, signals)

	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if ordered[j].Timestamp.After(ordered[i].Timestamp) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	return ordered
}
