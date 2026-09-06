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
- [ ] **WS1.4**: Persist raw payload to Firestore `raw_events` and canonical signal to `citizen_signals` via official Go Firestore SDK.
- **Test Separation**:
  - `server/data/indore_wards.json` (Mock dataset)
  - Go unit test: `go test ./...` verifying webhook conversion & WebSocket broadcast offline.

---

### Workstream 2: Extraction Pipeline (Go + Gemini SDK)
- [X] **WS2.1**: Define Go structured extraction schema structs (`Issue`, `Ward`, `ServiceCategory`, `Urgency1To5`, `HazardTags`, `Intent`, `Summary`).
- [X] **WS2.2**: Implement Gemini structured output extraction prompt with Hindi/Hinglish/English few-shot examples using `github.com/google/generative-ai-go` (`server/internal/extraction/gemini.go`).
- [X] **WS2.3**: Generate text embeddings (`text-embedding-004` 768-dim float32) for issue clustering with normalized deterministic fallback.
- [ ] **WS2.4**: Persist extraction & embeddings to Firestore `ai_extractions` (Track 3 dependency).
- **Test Separation**:
  - `server/internal/extraction/testdata/raw_complaints_multilingual.json` (10 synthetic mixed-language complaints)
  - Go unit test: `go test -v ./internal/extraction/...` verified 100% PASS with few-shot Hindi/Hinglish validation, normalized embeddings, and grounded summary generation.

---

### Workstream 3: Data Layer & Seeds (Go + Firebase Firestore)
- [ ] **WS3.1**: Official Go Firestore client initialization (`cloud.google.com/go/firestore`) & collection schema bindings.
- [X] **WS3.2**: Indore Ward Reference dataset (12 wards with centroids, demographic/infra index, historical investment in `server/data/indore_wards.json`).
- [ ] **WS3.3**: Synthetic Go seed script (`cmd/seed/main.go`) to populate ~30–50 complaints with engineered cluster hotspots and blind-spot wards.
- **Test Separation**:
  - `server/data/indore_wards.json`
  - Validation test: Run seed script against local Firestore emulator / test project.

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
- [X] **WS4.4**: Civic blind-spot candidate detection rule (poor infra index + low submission volume).
- [ ] **WS4.5**: Persist computed clusters, urgency tiers, and hotspots to Firestore `clusters` & `hotspots`.
- **Test Separation**:
  - `server/internal/urgency/engine_test.go`: Verified with `go test -v ./internal/urgency/...`.
  - `server/internal/clustering/engine_test.go`: Verified with `go test -v ./internal/clustering/...`.

---

### Workstream 5: Dashboard Frontend (Next.js 15 + WebSockets + Google Maps)
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
- [ ] **WS6.2**: Persist to Firestore `recommendations` and render in Next.js dashboard cluster drawer.
- **Test Separation**:
  - `go test ./internal/clustering/...` verifies recommendation generation format.

---

### Workstream 7: Live Demo & End-to-End Verification
- [X] **WS7.1**: Real-time simulation endpoint (`POST /api/v1/demo/simulate`) with interactive UI trigger button.
- [X] **WS7.2**: Demo script run-through checklist for GDG presentation (`presentation/GDG_3_MIN_DEMO_SCRIPT.md`, `scripts/test_viasocket_webhook.sh`, `docs/viasocket/VIASOCKET_SETUP_GUIDE.md`).

---

## Test & Fixture Separation Matrix

| Component | Upstream Dependency | Mock / Fixture Strategy | Independent Verification Method |
|---|---|---|---|
| **Go Ingestion** | viasocket Webhook | `tests/fixtures/viasocket_sample_payload.json` | `curl` payload to local Go endpoint + WebSocket listener |
| **Go Extraction** | Gemini API | Multilingual raw text fixtures + Mock Gemini JSON | `go test ./internal/extraction/...` |
| **Go Clustering/Urgency** | None (In-memory) | In-memory cluster fixture arrays | `go test ./internal/clustering/...` & `go test ./internal/urgency/...` |
| **Next.js Dashboard** | Go Backend / Firestore | `mock_dashboard_data.json` fixture route | `npm run dev` with mock data toggle |
| **Live Demo** | WhatsApp | Pre-seeded Firestore + 1 live WhatsApp trigger | Verification check script |

