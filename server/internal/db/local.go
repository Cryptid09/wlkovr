package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"walkover/server/internal/models"
)

// DefaultLocalStoreDirName is the directory (under server/data) holding the
// local JSON collections.
const DefaultLocalStoreDirName = "local_store"

// localRepo persists each collection as one JSON file mapping document ID to
// document, which keeps the seeded data readable and diffable as a fixture for
// the other workstreams.
type localRepo struct {
	dir string
	mu  sync.RWMutex
}

// NewLocalRepository opens (creating if needed) a JSON-file collection store.
// An empty dir resolves to server/data/local_store relative to the caller's
// working directory, matching how the API server locates indore_wards.json.
func NewLocalRepository(dir string) (Repository, error) {
	if dir == "" {
		dir = resolveLocalStoreDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("db: creating local store %s: %w", dir, err)
	}
	return &localRepo{dir: dir}, nil
}

// resolveLocalStoreDir finds the server/data directory whether the process was
// started from the repo root or from server/.
func resolveLocalStoreDir() string {
	for _, base := range []string{"data", "../data", "server/data"} {
		if info, err := os.Stat(base); err == nil && info.IsDir() {
			return filepath.Join(base, DefaultLocalStoreDirName)
		}
	}
	return filepath.Join("data", DefaultLocalStoreDirName)
}

// Dir reports where the local store writes its collection files.
func (r *localRepo) Dir() string { return r.dir }

func (r *localRepo) Backend() string { return BackendLocalJSON }

func (r *localRepo) Close() error { return nil }

func (r *localRepo) path(collection string) string {
	return filepath.Join(r.dir, collection+".json")
}

// read loads a collection file, returning an empty map when it does not exist.
func (r *localRepo) read(collection string) (map[string]json.RawMessage, error) {
	raw, err := os.ReadFile(r.path(collection))
	if os.IsNotExist(err) {
		return map[string]json.RawMessage{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("db: reading %s: %w", collection, err)
	}

	docs := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &docs); err != nil {
		return nil, fmt.Errorf("db: parsing %s: %w", collection, err)
	}
	return docs, nil
}

// write replaces a collection file atomically so a crash mid-seed cannot leave
// a half-written JSON document behind.
func (r *localRepo) write(collection string, docs map[string]json.RawMessage) error {
	encoded, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("db: encoding %s: %w", collection, err)
	}

	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("db: creating local store %s: %w", r.dir, err)
	}

	final := r.path(collection)
	tmp, err := os.CreateTemp(r.dir, collection+".*.tmp")
	if err != nil {
		return fmt.Errorf("db: staging %s: %w", collection, err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(encoded); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("db: writing %s: %w", collection, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("db: closing %s: %w", collection, err)
	}
	// CreateTemp makes the file 0600; these are shared fixtures other
	// workstreams read, so widen to the same mode a normal write would give.
	if err := os.Chmod(tmpName, 0o644); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("db: setting mode on %s: %w", collection, err)
	}
	if err := os.Rename(tmpName, final); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("db: committing %s: %w", collection, err)
	}
	return nil
}

func localSet[T any](r *localRepo, collection, id string, doc T) error {
	if id == "" {
		return fmt.Errorf("db: %s document requires a non-empty id", collection)
	}

	encoded, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("db: encoding %s/%s: %w", collection, id, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.read(collection)
	if err != nil {
		return err
	}
	docs[id] = encoded
	return r.write(collection, docs)
}

func localGet[T any](r *localRepo, collection, id string) (T, error) {
	var doc T

	r.mu.RLock()
	defer r.mu.RUnlock()

	docs, err := r.read(collection)
	if err != nil {
		return doc, err
	}

	encoded, ok := docs[id]
	if !ok {
		return doc, fmt.Errorf("%w: %s/%s", ErrNotFound, collection, id)
	}
	if err := json.Unmarshal(encoded, &doc); err != nil {
		return doc, fmt.Errorf("db: decoding %s/%s: %w", collection, id, err)
	}
	return doc, nil
}

func localDelete(r *localRepo, collection, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	docs, err := r.read(collection)
	if err != nil {
		return err
	}
	delete(docs, id)
	return r.write(collection, docs)
}

// localList returns documents in ascending document-ID order, matching the
// Firestore backend's ordering.
func localList[T any](r *localRepo, collection string, limit int) ([]T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs, err := r.read(collection)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(docs))
	for id := range docs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	results := make([]T, 0, len(ids))
	for _, id := range ids {
		var doc T
		if err := json.Unmarshal(docs[id], &doc); err != nil {
			return nil, fmt.Errorf("db: decoding %s/%s: %w", collection, id, err)
		}
		results = append(results, doc)
	}
	return applyLimit(results, limit), nil
}

func (r *localRepo) UpsertWards(_ context.Context, wards []models.Ward) error {
	for _, ward := range wards {
		if err := localSet(r, CollectionWards, ward.ID, ward); err != nil {
			return err
		}
	}
	return nil
}

func (r *localRepo) ListWards(_ context.Context) ([]models.Ward, error) {
	return localList[models.Ward](r, CollectionWards, 0)
}

func (r *localRepo) SaveRawEvent(_ context.Context, event RawEvent) error {
	return localSet(r, CollectionRawEvents, event.ID, event)
}

func (r *localRepo) ListRawEvents(_ context.Context, limit int) ([]RawEvent, error) {
	return localList[RawEvent](r, CollectionRawEvents, limit)
}

func (r *localRepo) SaveSignal(_ context.Context, signal models.CitizenSignal) error {
	return localSet(r, CollectionCitizenSignals, signal.ID, signal)
}

func (r *localRepo) ListSignals(_ context.Context, limit int) ([]models.CitizenSignal, error) {
	return localList[models.CitizenSignal](r, CollectionCitizenSignals, limit)
}

func (r *localRepo) SaveExtraction(_ context.Context, extraction models.AIExtraction) error {
	return localSet(r, CollectionAIExtractions, extraction.SignalID, extraction)
}

func (r *localRepo) ListExtractions(_ context.Context, limit int) ([]models.AIExtraction, error) {
	return localList[models.AIExtraction](r, CollectionAIExtractions, limit)
}

func (r *localRepo) UpsertCluster(_ context.Context, cluster models.Cluster) error {
	return localSet(r, CollectionClusters, cluster.ID, cluster)
}

func (r *localRepo) DeleteCluster(_ context.Context, id string) error {
	if err := localDelete(r, CollectionClusters, id); err != nil {
		return err
	}
	return localDelete(r, CollectionHotspots, id)
}

func (r *localRepo) GetCluster(_ context.Context, id string) (models.Cluster, error) {
	return localGet[models.Cluster](r, CollectionClusters, id)
}

func (r *localRepo) ListClusters(_ context.Context) ([]models.Cluster, error) {
	return localList[models.Cluster](r, CollectionClusters, 0)
}

func (r *localRepo) UpsertHotspot(_ context.Context, cluster models.Cluster) error {
	return localSet(r, CollectionHotspots, cluster.ID, cluster)
}

func (r *localRepo) ListHotspots(_ context.Context) ([]models.Cluster, error) {
	return localList[models.Cluster](r, CollectionHotspots, 0)
}

func (r *localRepo) SaveRecommendation(_ context.Context, rec Recommendation) error {
	return localSet(r, CollectionRecommendations, rec.ID, rec)
}

func (r *localRepo) ListRecommendations(_ context.Context, limit int) ([]Recommendation, error) {
	return localList[Recommendation](r, CollectionRecommendations, limit)
}

func (r *localRepo) AppendAuditLog(_ context.Context, entry models.AuditLog) error {
	return localSet(r, CollectionAuditLogs, entry.ID, entry)
}

func (r *localRepo) ListAuditLogs(_ context.Context, limit int) ([]models.AuditLog, error) {
	return localList[models.AuditLog](r, CollectionAuditLogs, limit)
}

func (r *localRepo) Reset(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, collection := range AllCollections {
		if err := os.Remove(r.path(collection)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("db: clearing %s: %w", collection, err)
		}
	}
	return nil
}
