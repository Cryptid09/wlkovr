package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"walkover/server/internal/models"
)

// firestoreRepo is the Google Firestore implementation of Repository.
type firestoreRepo struct {
	client     *firestore.Client
	projectID  string
	isEmulator bool
}

// NewFirestoreRepository connects to Firestore for the given project. When
// emulatorHost is non-empty the client is pointed at a local emulator instead
// of Google Cloud, which needs no credentials.
func NewFirestoreRepository(ctx context.Context, projectID, emulatorHost string) (Repository, error) {
	if projectID == "" {
		return nil, errors.New("db: GOOGLE_CLOUD_PROJECT is required for the Firestore backend")
	}

	// The Firestore SDK reads the emulator host from the environment, so make
	// sure a host supplied purely through config reaches it.
	if emulatorHost != "" {
		if err := os.Setenv("FIRESTORE_EMULATOR_HOST", emulatorHost); err != nil {
			return nil, fmt.Errorf("db: setting emulator host: %w", err)
		}
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("db: connecting to Firestore project %s: %w", projectID, err)
	}

	return &firestoreRepo{
		client:     client,
		projectID:  projectID,
		isEmulator: emulatorHost != "",
	}, nil
}

func (r *firestoreRepo) Backend() string {
	if r.isEmulator {
		return BackendFirestoreEmulator
	}
	return BackendFirestore
}

func (r *firestoreRepo) Close() error { return r.client.Close() }

func setDoc[T any](ctx context.Context, r *firestoreRepo, collection, id string, doc T) error {
	if id == "" {
		return fmt.Errorf("db: %s document requires a non-empty id", collection)
	}
	if _, err := r.client.Collection(collection).Doc(id).Set(ctx, doc); err != nil {
		return fmt.Errorf("db: writing %s/%s: %w", collection, id, err)
	}
	return nil
}

func getDoc[T any](ctx context.Context, r *firestoreRepo, collection, id string) (T, error) {
	var doc T

	snapshot, err := r.client.Collection(collection).Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return doc, fmt.Errorf("%w: %s/%s", ErrNotFound, collection, id)
	}
	if err != nil {
		return doc, fmt.Errorf("db: reading %s/%s: %w", collection, id, err)
	}
	if err := snapshot.DataTo(&doc); err != nil {
		return doc, fmt.Errorf("db: decoding %s/%s: %w", collection, id, err)
	}
	return doc, nil
}

// listDocs returns documents in ascending document-ID order, matching the local
// JSON backend's ordering.
func listDocs[T any](ctx context.Context, r *firestoreRepo, collection string, limit int) ([]T, error) {
	query := r.client.Collection(collection).OrderBy(firestore.DocumentID, firestore.Asc)
	if limit > 0 {
		query = query.Limit(limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	results := make([]T, 0)
	for {
		snapshot, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("db: listing %s: %w", collection, err)
		}

		var doc T
		if err := snapshot.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("db: decoding %s/%s: %w", collection, snapshot.Ref.ID, err)
		}
		results = append(results, doc)
	}
	return results, nil
}

func (r *firestoreRepo) UpsertWards(ctx context.Context, wards []models.Ward) error {
	for _, ward := range wards {
		if err := setDoc(ctx, r, CollectionWards, ward.ID, ward); err != nil {
			return err
		}
	}
	return nil
}

func (r *firestoreRepo) ListWards(ctx context.Context) ([]models.Ward, error) {
	return listDocs[models.Ward](ctx, r, CollectionWards, 0)
}

func (r *firestoreRepo) SaveRawEvent(ctx context.Context, event RawEvent) error {
	return setDoc(ctx, r, CollectionRawEvents, event.ID, event)
}

func (r *firestoreRepo) ListRawEvents(ctx context.Context, limit int) ([]RawEvent, error) {
	return listDocs[RawEvent](ctx, r, CollectionRawEvents, limit)
}

func (r *firestoreRepo) SaveSignal(ctx context.Context, signal models.CitizenSignal) error {
	return setDoc(ctx, r, CollectionCitizenSignals, signal.ID, signal)
}

func (r *firestoreRepo) ListSignals(ctx context.Context, limit int) ([]models.CitizenSignal, error) {
	return listDocs[models.CitizenSignal](ctx, r, CollectionCitizenSignals, limit)
}

func (r *firestoreRepo) SaveExtraction(ctx context.Context, extraction models.AIExtraction) error {
	return setDoc(ctx, r, CollectionAIExtractions, extraction.SignalID, extraction)
}

func (r *firestoreRepo) ListExtractions(ctx context.Context, limit int) ([]models.AIExtraction, error) {
	return listDocs[models.AIExtraction](ctx, r, CollectionAIExtractions, limit)
}

func (r *firestoreRepo) UpsertCluster(ctx context.Context, cluster models.Cluster) error {
	return setDoc(ctx, r, CollectionClusters, cluster.ID, cluster)
}

func (r *firestoreRepo) GetCluster(ctx context.Context, id string) (models.Cluster, error) {
	return getDoc[models.Cluster](ctx, r, CollectionClusters, id)
}

func (r *firestoreRepo) ListClusters(ctx context.Context) ([]models.Cluster, error) {
	return listDocs[models.Cluster](ctx, r, CollectionClusters, 0)
}

func (r *firestoreRepo) UpsertHotspot(ctx context.Context, cluster models.Cluster) error {
	return setDoc(ctx, r, CollectionHotspots, cluster.ID, cluster)
}

func (r *firestoreRepo) ListHotspots(ctx context.Context) ([]models.Cluster, error) {
	return listDocs[models.Cluster](ctx, r, CollectionHotspots, 0)
}

func (r *firestoreRepo) SaveRecommendation(ctx context.Context, rec Recommendation) error {
	return setDoc(ctx, r, CollectionRecommendations, rec.ID, rec)
}

func (r *firestoreRepo) ListRecommendations(ctx context.Context, limit int) ([]Recommendation, error) {
	return listDocs[Recommendation](ctx, r, CollectionRecommendations, limit)
}

func (r *firestoreRepo) AppendAuditLog(ctx context.Context, entry models.AuditLog) error {
	return setDoc(ctx, r, CollectionAuditLogs, entry.ID, entry)
}

func (r *firestoreRepo) ListAuditLogs(ctx context.Context, limit int) ([]models.AuditLog, error) {
	return listDocs[models.AuditLog](ctx, r, CollectionAuditLogs, limit)
}

func (r *firestoreRepo) Reset(ctx context.Context) error {
	for _, collection := range AllCollections {
		iter := r.client.Collection(collection).Documents(ctx)
		for {
			snapshot, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				iter.Stop()
				return fmt.Errorf("db: clearing %s: %w", collection, err)
			}
			if _, err := snapshot.Ref.Delete(ctx); err != nil {
				iter.Stop()
				return fmt.Errorf("db: deleting %s/%s: %w", collection, snapshot.Ref.ID, err)
			}
		}
		iter.Stop()
	}
	return nil
}
