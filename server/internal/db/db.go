// Package db is the persistence layer for the platform's Firestore collections.
//
// It exposes one Repository contract with two interchangeable backends:
//
//   - Firestore, used whenever a Firestore emulator host or a Google service
//     account credential file is configured.
//   - Local JSON files under server/data/local_store, used otherwise.
//
// The local backend exists so that every other workstream (ingestion,
// extraction, clustering, dashboard) can run the seed script and develop
// against realistic data without a Google Cloud project. Both backends store
// identical document shapes and list documents in ascending document-ID order,
// so switching between them changes nothing for callers.
package db

import (
	"context"
	"errors"
	"os"
	"time"

	"walkover/server/config"
	"walkover/server/internal/models"
)

// Collection names — the canonical Firestore layout described in UNDERSTANDING.md.
const (
	CollectionWards           = "wards"
	CollectionRawEvents       = "raw_events"
	CollectionCitizenSignals  = "citizen_signals"
	CollectionAIExtractions   = "ai_extractions"
	CollectionClusters        = "clusters"
	CollectionHotspots        = "hotspots"
	CollectionRecommendations = "recommendations"
	CollectionAuditLogs       = "audit_logs"
)

// AllCollections lists every collection the platform owns, in write order.
var AllCollections = []string{
	CollectionWards,
	CollectionRawEvents,
	CollectionCitizenSignals,
	CollectionAIExtractions,
	CollectionClusters,
	CollectionHotspots,
	CollectionRecommendations,
	CollectionAuditLogs,
}

// Backend identifiers returned by Repository.Backend.
const (
	BackendFirestore         = "firestore"
	BackendFirestoreEmulator = "firestore-emulator"
	BackendLocalJSON         = "local-json"
)

// ErrNotFound is returned when a document does not exist in a collection.
var ErrNotFound = errors.New("db: document not found")

// RawEvent is the untouched provider payload exactly as viasocket delivered it,
// kept separately from the canonical signal so the audit trail can always show
// what the citizen originally sent.
type RawEvent struct {
	ID         string                  `json:"id" firestore:"id"`
	Provider   string                  `json:"provider" firestore:"provider"`
	SignalID   string                  `json:"signal_id" firestore:"signal_id"`
	Payload    models.ViasocketPayload `json:"payload" firestore:"payload"`
	ReceivedAt time.Time               `json:"received_at" firestore:"received_at"`
}

// Recommendation is a grounded policy brief generated for one cluster, with the
// signal IDs it was grounded on so a policymaker can trace every claim back to
// the citizen reports behind it.
type Recommendation struct {
	ID                string    `json:"id" firestore:"id"`
	ClusterID         string    `json:"cluster_id" firestore:"cluster_id"`
	Department        string    `json:"department" firestore:"department"`
	Text              string    `json:"text" firestore:"text"`
	EvidenceSignalIDs []string  `json:"evidence_signal_ids" firestore:"evidence_signal_ids"`
	CreatedAt         time.Time `json:"created_at" firestore:"created_at"`
}

// Repository is the persistence contract every workstream codes against.
// A limit of 0 or less means "no limit".
type Repository interface {
	// Backend reports which implementation is in use (see Backend* constants).
	Backend() string

	UpsertWards(ctx context.Context, wards []models.Ward) error
	ListWards(ctx context.Context) ([]models.Ward, error)

	SaveRawEvent(ctx context.Context, event RawEvent) error
	ListRawEvents(ctx context.Context, limit int) ([]RawEvent, error)

	SaveSignal(ctx context.Context, signal models.CitizenSignal) error
	ListSignals(ctx context.Context, limit int) ([]models.CitizenSignal, error)

	// SaveExtraction stores a Gemini extraction keyed by its SignalID.
	SaveExtraction(ctx context.Context, extraction models.AIExtraction) error
	ListExtractions(ctx context.Context, limit int) ([]models.AIExtraction, error)

	UpsertCluster(ctx context.Context, cluster models.Cluster) error
	GetCluster(ctx context.Context, id string) (models.Cluster, error)
	ListClusters(ctx context.Context) ([]models.Cluster, error)

	// UpsertHotspot mirrors a cluster into the hotspots collection. Hotspots are
	// clusters the intelligence engine surfaced as concentrated demand, kept in
	// their own collection so the map view can read them without scanning every
	// cluster.
	UpsertHotspot(ctx context.Context, cluster models.Cluster) error
	ListHotspots(ctx context.Context) ([]models.Cluster, error)

	SaveRecommendation(ctx context.Context, rec Recommendation) error
	ListRecommendations(ctx context.Context, limit int) ([]Recommendation, error)

	AppendAuditLog(ctx context.Context, entry models.AuditLog) error
	ListAuditLogs(ctx context.Context, limit int) ([]models.AuditLog, error)

	// Reset deletes every document in every collection. It exists so the seed
	// script can be re-run idempotently — never call it from request handlers.
	Reset(ctx context.Context) error

	Close() error
}

// NewRepository picks a backend from configuration: Firestore when an emulator
// host or a service-account credential file is configured, local JSON otherwise.
//
// Note that cfg.GoogleCloudProject always carries a default value, so its
// presence alone is not evidence that Firestore is actually reachable.
func NewRepository(ctx context.Context, cfg *config.Config) (Repository, error) {
	if cfg.FirestoreEmulatorHost != "" || os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		return NewFirestoreRepository(ctx, cfg.GoogleCloudProject, cfg.FirestoreEmulatorHost)
	}
	return NewLocalRepository("")
}

// EmbeddingText is the canonical text used to embed a stored extraction.
// Every producer of embeddings must use it, or vectors written by the seed
// backfill and vectors written by the live pipeline would not be comparable.
func EmbeddingText(extraction models.AIExtraction) string {
	return extraction.Issue + " " + extraction.Summary
}

// applyLimit truncates a result slice, treating limit <= 0 as unlimited.
func applyLimit[T any](docs []T, limit int) []T {
	if limit > 0 && len(docs) > limit {
		return docs[:limit]
	}
	return docs
}
