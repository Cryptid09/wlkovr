# Project Progress & Task Board (`PROGRESS.md`)

This board tracks task distribution, implementation status, and test separation across the 4-person team and AI agents.

## Legend
- `[ ]` **Not Started** — Available to be picked up
- `[/]` **In Progress** — Claimed by an agent/teammate (specify name/ID)
- `[X]` **Completed** — Built and verified with tests
- `[-]` **Blocked** — Blocked by dependency or decision

---

## Workstream Breakdown & Task Distribution

### Workstream 1: Ingestion & Realtime Intake (Go + viasocket + WebSockets)
- [X] **WS1.1**: Define Canonical Citizen Signal Go struct / JSON contract (`ID`, `Provider`, `RawText`, `Language`, `Location`, `Timestamp`, `Metadata`).
- [X] **WS1.2**: Implement Go Gin webhook endpoint receiving viasocket WhatsApp/SMS payloads $\rightarrow$ transform to Canonical Signal.
- [X] **WS1.3**: Implement Go WebSocket event broadcaster (`SIGNAL_RECEIVED`, `DECISION_RECORDED`) for real-time dashboard sync.
- [X] **WS1.4**: Persist raw payload to `raw_events` and canonical signal to `citizen_signals` via the `db.Repository` layer. Wired in `server/internal/api/pipeline.go` — the webhook broadcasts first, then persists off the request path so live-demo latency is unaffected.
- **Test Separation**:
  - `server/data/indore_wards.json` (Mock dataset)
  - Go unit test: `go test ./...` verifying webhook conversion & WebSocket broadcast offline.

---

### Workstream 2: Extraction Pipeline (Go + Gemini SDK)
- [X] **WS2.1**: Define Go structured extraction schema structs (`Issue`, `Ward`, `ServiceCategory`, `Urgency1To5`, `HazardTags`, `Intent`, `Summary`).
- [X] **WS2.2**: Implement Gemini structured output extraction prompt with Hindi/Hinglish/English few-shot examples using `github.com/google/generative-ai-go` (`server/internal/extraction/gemini.go`).
- [X] **WS2.3**: Generate text embeddings for issue clustering with normalized deterministic fallback. **Model corrected to `gemini-embedding-001` (3072-dim)** — `text-embedding-004` returns HTTP 404 on the v1beta endpoint this SDK targets. The fallback width now tracks the configured model via `embeddingDimensions`, and `GenerateEmbedding`/`ExtractSignal` return an error when they degrade instead of passing fallbacks off as model output. Regression test: `internal/extraction/dimensions_test.go`. Backfill with `go run ./cmd/seed --reset --embed`.
- [X] **WS2.4**: Persist extraction & embeddings to `ai_extractions`. Incoming signals run through `extraction.Extractor` (Gemini 3.6 Flash + `gemini-embedding-001`, 3072-dim) and are stored with their embedding.
- **Test Separation**:
  - `server/internal/extraction/testdata/raw_complaints_multilingual.json` (10 synthetic mixed-language complaints)
  - Go unit test: `go test -v ./internal/extraction/...` verified 100% PASS with few-shot Hindi/Hinglish validation, normalized embeddings, and grounded summary generation.

---

### Workstream 3: Data Layer & Seeds (Go + Firebase Firestore)
- [X] **WS3.1**: Official Go Firestore client initialization (`cloud.google.com/go/firestore`) & collection schema bindings. — `server/internal/db/` exposes one `Repository` contract over all 8 collections with two backends: Firestore (`firestore.go`) and a local JSON store (`local.go`). Verified: `go test ./internal/db/...`
- [X] **WS3.2**: Indore Ward Reference dataset (12 wards with centroids, demographic/infra index, historical investment in `server/data/indore_wards.json`).
- [X] **WS3.3**: Synthetic Go seed script (`cmd/seed/`) — 42 signals: 8 (Ward 14 water) + 12 (Ward 22 road) + 6 (Ward 60 manhole) engineered hotspots, 0 in Ward 1 / Ward 78 (blind spots), 16 background noise. Verified: `go run ./cmd/seed --reset`
- **Test Separation**:
  - `server/data/indore_wards.json` (ward reference input)
  - `server/data/local_store/*.json` (seeded fixture output — 8 collections, readable by any workstream without Google Cloud access)
  - `go test ./internal/db/...` — repository round-trips against a temp-dir local store, no credentials needed.
  - `go test ./cmd/seed/...` — corpus size, hotspot volumes, blind-spot silence, multilingual coverage, equity inversion, determinism.
  - `go run ./cmd/seed --dry-run` — build and summarise without writing.

**Backend selection**: Firestore is used when `FIRESTORE_EMULATOR_HOST` or `GOOGLE_APPLICATION_CREDENTIALS` is set; the local JSON store is used otherwise. Note `GOOGLE_CLOUD_PROJECT` always carries a default value, so it is not evidence Firestore is reachable.

**Consumer note (other workstreams)**: call `db.NewRepository(ctx, cfg)` and code against the `db.Repository` interface. Never construct a Firestore client directly.

**End-to-end wiring (WS1.4 / WS2.4 / WS4.5 / WS6.2)** is complete in `server/internal/api/pipeline.go`. `NewHandler(cfg, hub)` is unchanged and still serves in-memory demo state on its own; `WithPersistence(ctx, repo)` and `WithExtractor(ex)` attach the optional dependencies, and `cmd/api/main.go` attaches both at startup. Both degrade gracefully: no credentials means a local JSON store, no Gemini key means deterministic offline extraction.

Verified live end-to-end: a Hindi WhatsApp complaint posted to `/api/v1/webhooks/viasocket` was extracted by Gemini, embedded (768-dim), matched to `cluster-indore-001`, raised it from 8 to 9 signals and need 94 to 97 while holding Tier 1, and was persisted — while an unrelated SMS from a blind-spot ward was correctly left unclustered.

---

### Workstream 4: In-Memory Clustering & Urgency Decision Engine (Go)
- [X] **WS4.1**: Go in-memory cosine similarity threshold grouping (cluster complaints by issue within the same ward using vector dot-product).
- [X] **WS4.2**: **Urgency Decision Engine**:
  - Hazard multiplier rules ($H_{\text{hazard}}$: contaminated water, live wires, open manholes, hospital access routes).
  - Temporal spike rate velocity ($V_{\text{velocity}}$: sliding window complaint surge).
  - Infrastructure sensitivity factor ($S_{\text{sensitivity}}$).
  - Tiered SLA classification (Tier 1 <4h, Tier 2 <24h, Tier 3 <72h, Tier 4 <7d).
- [X] **WS4.3**: Implement 4 scoring dimension engines:
  - `Need` = normalized cluster size × Urgency Decision Engine score
  - `Confidence` = channel diversity + corroborating signal count
  - `Equity` = inverse of ward infra/demographic index
  - `Actionability` = location resolved + clear department mapping heuristic
- [X] **WS4.4**: Civic blind-spot candidate detection rule (poor infra index + **zero** active clusters). *Rule corrected 2026-09-06 by Claude Code (Track 3, authorised cross-track fix): the previous `count <= 1` threshold flagged Ward 14 and Ward 60 as blind spots while they were simultaneously the top demand hotspots. Regression test: `TestUnderservedWardWithReportsIsNotBlindSpot`.*
- [X] **WS4.5**: Persist computed clusters, urgency tiers, and hotspots to `clusters` & `hotspots`. A newly understood signal is matched to an existing cluster (ward + department, plus cosine similarity once embeddings exist on both sides), the cluster is rescored, and blind spots are recomputed. Cluster urgency uses the **strongest** evidence across the cluster, never the newest message, so a mild follow-up cannot de-escalate a critical issue.
- **Test Separation**:
  - `server/internal/urgency/engine_test.go`: Verified with `go test -v ./internal/urgency/...`.
  - `server/internal/clustering/engine_test.go`: Verified with `go test -v ./internal/clustering/...`.

---

### Workstream 5: Dashboard Frontend (Next.js 15 + WebSockets + Leaflet/OpenStreetMap)
- [X] **WS5.1**: Next.js 15 project initialized in `web/` with Tailwind CSS + Lucide Icons + Recharts.
- [X] **WS5.2**: Hotspot & Ward list view with Urgency Tier badges, Metric cards, and live stream feed.
- [X] **WS5.3**: Priority Cluster Detail view displaying the **4 separate score bars** (Need, Confidence, Equity, Actionability).
- [X] **WS5.4**: Real-time WebSocket client integration connecting to Go backend (`/ws`) for live signal and decision updates.
- [X] **WS5.5**: Policymaker Decision Panel (Accept / Reject / Investigate actions) with audit trail feedback.
- **Test Separation**:
  - `web/src/lib/api.ts` with typed fallback support.
  - Verified with `npm run build`.

---

### Workstream 6: Grounded Recommendation Generator (Go + Gemini)
- [X] **WS6.1**: Implement grounded recommendation engine generating human-readable recommendations citing specific evidence/submissions.
- [X] **WS6.2**: Persist regenerated grounded recommendations to `recommendations` on every cluster update. *Rendering in the Next.js cluster drawer remains with Workstream 5.*
- **Test Separation**:
  - `go test ./internal/clustering/...` verifies recommendation generation format.

---

### Workstream 8: Policymaker Assistant (Go + Gemini)
- [X] **WS8.1**: `POST /api/v1/assistant` — free-text questions answered from the platform's own evidence (selected cluster, ward, citizen reports in their original language, blind-spot state).
- [X] **WS8.2**: Prompt constrained so the assistant explains and recommends investigation but never approves, funds or closes anything; it says so plainly when the evidence cannot answer.
- [X] **WS8.3**: Deterministic offline answer path, so the panel still explains stored evidence when Gemini is unreachable. The response carries `source: "gemini" | "offline"` and the UI labels an offline answer.
- [X] **WS8.4**: Dashboard help panel wired to the endpoint with a pending state, replacing the previous hardcoded keyword matcher.
- **Test Separation**:
  - `go test ./internal/api/... -run Assistant` — 5 tests: question required, offline answer references the selected cluster, never claims authority, evidence includes platform state, evidence quotes citizen reports.

---

### Workstream 7: Live Demo & End-to-End Verification
- [X] **WS7.1**: Real-time simulation endpoint (`POST /api/v1/demo/simulate`) with interactive UI trigger button.
- [X] **WS7.2**: Demo script run-through checklist for GDG presentation (`presentation/GDG_3_MIN_DEMO_SCRIPT.md`, `scripts/test_viasocket_webhook.sh`, `docs/viasocket/VIASOCKET_SETUP_GUIDE.md`).

---

## Backend status (2026-09-06)

The Go backend is complete and verified end to end against live Firestore (project `wlkovr`) and live Gemini. Run it with:

```bash
cd server
go run ./cmd/seed --reset --embed   # ~5 min: writes the corpus, then embeds it
go run ./cmd/api                    # :8080
```

Confirm these two startup lines before demoing — anything else means credentials are missing and the server has quietly fallen back:

```
[STORE]  Hydrated from firestore — 12 wards, 3 clusters, 42 signals, 42 embeddings
[GEMINI] Extractor attached — live structured extraction and embeddings enabled
```

Verified behaviour: three multilingual follow-ups posted to `/api/v1/webhooks/viasocket` each joined the correct cluster and rescored it — Khajrana manhole 6→7 (need 88→91, confidence 58.3→79.2 as corroboration became cross-channel), Chandan Nagar water 8→9 (need 94→97), Vijay Nagar road 12→13. All held TIER_1_CRITICAL. A complaint from a blind-spot ward correctly matched nothing.

REST endpoints and the four WebSocket event types are documented in `UNDERSTANDING.md` under "Backend: API and WebSocket contract". `server/internal/models/models.go` is the source of truth for every JSON shape.

**Not done**: `NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` is empty, and no real viasocket flow points at the webhook yet (`docs/viasocket/VIASOCKET_SETUP_GUIDE.md` has the steps).

---

## Test & Fixture Separation Matrix

| Component | Upstream Dependency | Mock / Fixture Strategy | Independent Verification Method |
|---|---|---|---|
| **Go Ingestion** | viasocket Webhook | `tests/fixtures/viasocket_sample_payload.json` | `curl` payload to local Go endpoint + WebSocket listener |
| **Go Extraction** | Gemini API | Multilingual raw text fixtures + Mock Gemini JSON | `go test ./internal/extraction/...` |
| **Go Clustering/Urgency** | None (In-memory) | In-memory cluster fixture arrays | `go test ./internal/clustering/...` & `go test ./internal/urgency/...` |
| **Next.js Dashboard** | Go Backend / Firestore | `mock_dashboard_data.json` fixture route | `npm run dev` with mock data toggle |
| **Live Demo** | WhatsApp | Pre-seeded Firestore + 1 live WhatsApp trigger | Verification check script |

