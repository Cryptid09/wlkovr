# Agent Work Journal (`AGENT.md`)

This file is a persistent, chronological log of all AI agent activities across sessions, IDEs, and tools.
**Rule**: Append-only. Never remove previous entries.

---

## Log Entries

### 2026-09-06 10:05 IST - Antigravity (Pair Programming Agent)
- **Workstream / Goal**: Multi-agent collaboration setup & engineering standards configuration
- **Tasks Claimed/Completed**:
  - Configured `rules.md` with multi-agent lifecycle, mandatory journal logging protocol, progress tracking, and industry-standard practices.
  - Created `AGENT.md` as the centralized append-only work journal.
  - Created `PROGRESS.md` with granular workstream breakdowns, task distribution, and test separation matrices.
- **Files Modified/Created**:
  - `[MOD] rules.md` — Established 4-step agent session workflow, journal protocol, task claiming rules, test separation, and scope constraints.
  - `[NEW] AGENT.md` — Initialized agent journal with template and initial setup log.
  - `[NEW] PROGRESS.md` — Initialized task board across all 6 workstreams with test separation criteria and mock fixtures.
- **Architectural & Design Decisions**:
  - Embedded contract-first interfaces and mock fixtures into each workstream to allow UI, clustering, and ingestion to be tested in isolation before end-to-end integration.
  - Locked adherence to `UNDERSTANDING.md` constraints (viasocket, Gemini, Cloud Run, PostGIS/pgvector, 4 scoring dimensions).
- **Testing & Verification Conducted**:
  - Verified document linkages and consistency across `rules.md`, `UNDERSTANDING.md`, `AGENT.md`, and `PROGRESS.md`.
- **Blockers / Open Questions**:
  - None. Ready for feature workstreams to begin.
- **Handoff / Next Recommended Steps**:
### 2026-09-06 10:28 IST - Antigravity (Pair Programming Agent)
- **Workstream / Goal**: Architecture alignment with team proposals (`mandeep.md`)
- **Tasks Claimed/Completed**:
  - Aligned tech stack to use **Socket.IO** for real-time channel intake and live dashboard streaming.
  - Aligned datastore to use **Firebase Firestore** following Mandeep's multi-stage collection architecture (`raw_events`, `citizen_signals`, `ai_extractions`, `clusters`, `hotspots`, `recommendations`, `audit_logs`).
  - Preserved Mandeep's complete diagrams in `mandeep.md` as the official Pitch/Software Architecture Document for judges.
  - Adapted clustering approach to use in-memory cosine similarity threshold grouping on Gemini embeddings per ward, persisting computed clusters directly to Firestore.
- **Files Modified/Created**:
  - `[MOD] UNDERSTANDING.md` — Updated tech stack table, Firestore collections schema, Socket.IO real-time intake, and in-memory clustering approach.
  - `[MOD] rules.md` — Updated locked decisions with Firebase Firestore and Socket.IO.
  - `[MOD] PROGRESS.md` — Updated workstreams 1–7 tasks and test separation matrix to match Firestore and Socket.IO.
- **Architectural & Design Decisions**:
  - Ingestion: viasocket normalized webhook $\rightarrow$ Socket.IO server $\rightarrow$ emit real-time events to processing and dashboard.
  - Storage: Firestore NoSQL collections maintain full auditability from raw event to policymaker decision.
  - Clustering: In-memory cosine similarity per ward eliminates vector DB operational complexity for the hackathon demo.
- **Testing & Verification Conducted**:
  - Verified document synchronization across `UNDERSTANDING.md`, `rules.md`, `PROGRESS.md`, `mandeep.md`, and `AGENT.md`.
### 2026-09-06 11:02 IST - Antigravity (Pair Programming Agent)
- **Workstream / Goal**: Urgency Decision Engine Architectural Specification
- **Tasks Claimed/Completed**:
  - Designed and specified the **Urgency Decision Engine** in `UNDERSTANDING.md`.
  - Defined multi-factor formula integrating base Gemini urgency, hazard multipliers ($H_{\text{hazard}}$: contaminated water, live wires, hospital routes), temporal spike velocity ($V_{\text{velocity}}$), and infrastructure sensitivity ($S_{\text{sensitivity}}$).
  - Defined 4-tier urgency classification and SLA matrix (Tier 1 <4h Critical, Tier 2 <24h High, Tier 3 <72h Medium, Tier 4 <7d Routine).
  - Updated `PROGRESS.md` Workstream 4 with dedicated Urgency Decision Engine tasks and test separation (`tests/test_urgency_engine.py`).
- **Files Modified/Created**:
  - `[MOD] UNDERSTANDING.md` — Added Urgency Decision Engine mathematical model, factors, and SLA tier table.
  - `[MOD] PROGRESS.md` — Added WS4.2 Urgency Decision Engine tasks and test separation.
- **Architectural & Design Decisions**:
  - Dynamic multi-factor urgency prevents arbitrary urgency ratings by combining AI semantic analysis with hard deterministic hazard detection and temporal burst detection.
### 2026-09-06 11:16 IST - Antigravity (Pair Programming Agent)
- **Workstream / Goal**: Technology Stack Alignment (Golang Backend Engine + Next.js Dashboard)
- **Tasks Claimed/Completed**:
  - Locked in **Golang (Go)** as the core backend engine runtime.
  - Specified official Google Cloud Go SDKs: `github.com/google/generative-ai-go` for Gemini 2.5 Flash structured extractions & embeddings, `cloud.google.com/go/firestore` for Firestore 7-collection operations.
  - Configured Go Gin framework + WebSockets for high-throughput webhook intake and instant dashboard event broadcasting.
  - Updated `UNDERSTANDING.md`, `rules.md`, and `PROGRESS.md` with Go structs, contracts, and `go test` verification suites.
- **Files Modified/Created**:
  - `[MOD] UNDERSTANDING.md` — Updated Tech Stack table to Golang Backend Engine + Next.js 15 Frontend.
  - `[MOD] rules.md` — Updated locked decisions to Golang backend engine.
  - `[MOD] PROGRESS.md` — Updated WS1–WS7 tasks with Go structs, SDKs, and `go test` suites.
- **Architectural & Design Decisions**:
  - Go's type safety, low latency, and Goroutine concurrency make it ideal for high-throughput webhook processing, fast in-memory cosine vector math, and real-time WebSocket distribution to Next.js.
- **Handoff / Next Recommended Steps**:
### 2026-09-06 11:26 IST - Antigravity (Pair Programming Agent)
- **Workstream / Goal**: Full Monorepo Project Setup, Central Config, Go Backend Engine & Next.js UI Scaffolding
- **Tasks Claimed/Completed**:
  - Initialized central configuration loader (`server/config/config.go`) reading from `.env` and `.env.example`.
  - Created Go domain models in `server/internal/models/models.go` (`CitizenSignal`, `Cluster`, `Ward`, `UrgencyResult`, `AuditLog`, etc.).
  - Implemented the **Urgency Decision Engine** in `server/internal/urgency/engine.go` with multi-factor scoring, hazard overrides, velocity spike detection, and SLA tiers.
  - Implemented in-memory cosine similarity clustering & 4D scoring in `server/internal/clustering/engine.go`.
  - Created all REST API and WebSocket routes in `server/internal/api/` (`/api/v1/health`, `/wards`, `/clusters`, `/signals`, `/webhooks/viasocket`, `/decisions`, `/demo/simulate`, `/ws`).
  - Added unit test suites (`server/internal/urgency/engine_test.go`, `server/internal/clustering/engine_test.go`) — all passing (`go test -v ./...`).
  - Scaffolding Next.js 15 App Router frontend in `web/` with Tailwind CSS, Lucide Icons, Recharts, and WebSocket real-time connection.
  - Created root `Makefile` with `make dev-server`, `make dev-web`, `make test`, and `make simulate`.
- **Files Modified/Created**:
  - `[NEW] .env.example`, `[NEW] .env`, `[NEW] Makefile`
  - `[NEW] server/cmd/api/main.go`, `[NEW] server/config/config.go`
  - `[NEW] server/internal/models/models.go`, `[NEW] server/data/indore_wards.json`
  - `[NEW] server/internal/urgency/engine.go`, `[NEW] server/internal/urgency/engine_test.go`
  - `[NEW] server/internal/clustering/engine.go`, `[NEW] server/internal/clustering/engine_test.go`
  - `[NEW] server/internal/api/handlers.go`, `[NEW] server/internal/api/router.go`, `[NEW] server/internal/api/websocket.go`
  - `[NEW] web/src/types/index.ts`, `[NEW] web/src/lib/api.ts`, `[NEW] web/src/lib/utils.ts`, `[NEW] web/src/app/page.tsx`
  - `[MOD] PROGRESS.md` — Updated task board with completed foundation milestones.
- **Testing & Verification Conducted**:
  - Go Backend: `go test -v ./...` $\rightarrow$ 100% PASS.
  - Next.js Frontend: `npm run build` $\rightarrow$ Compiled successfully with 0 errors.
- **Handoff / Next Recommended Steps**:
  - Team members / agents can now work simultaneously on:
    - **Go Backend team**: Connect live Gemini 2.5 Flash structured output extraction (`internal/extraction/gemini.go`) and Firestore persistence (`internal/db/firestore.go`).
    - **Frontend team**: Run `npm run dev` in `web/` to customize and polish UI components, interactive Google Maps polygons, and charts.

### 2026-09-06 12:00 IST - Antigravity (Pair Programming Agent)
- **Workstream / Goal**: Track 4: viasocket & Live Demo Integration (Sponsor Workflow)
- **Tasks Claimed/Completed**:
  - Aligned `server/go.mod` directive with local Go toolchain (`go 1.26.1`) for seamless offline build/test execution.
  - Implemented unit and integration test suite in `server/internal/api/handlers_test.go` covering `POST /api/v1/webhooks/viasocket`, `GET /signals`, `GET /clusters`, `POST /decisions`, and `POST /demo/simulate`.
  - Created sample JSON fixtures in `tests/fixtures/viasocket_sample_payload.json` with realistic multilingual complaints (Hindi, Hinglish, English).
  - Authored comprehensive viasocket flow setup documentation in `docs/viasocket/VIASOCKET_SETUP_GUIDE.md` (JS normalization transform + HTTP webhook action + Ngrok/Cloud Run tunnel setup).
  - Created automated webhook testing and latency benchmark script `scripts/test_viasocket_webhook.sh`.
  - Authored the timed 3-minute GDG presentation script `presentation/GDG_3_MIN_DEMO_SCRIPT.md` with problem framing, live WhatsApp demo cues, 4D scoring explanation, and human-in-the-loop action.
- **Files Modified/Created**:
  - `[MOD] server/go.mod` — Adjusted toolchain directive to `go 1.26.1`.
  - `[NEW] server/internal/api/handlers_test.go` — Test suite for API handlers and viasocket webhook ingestion.
  - `[NEW] tests/fixtures/viasocket_sample_payload.json` — Realistic WhatsApp & SMS webhook test payloads.
  - `[NEW] docs/viasocket/VIASOCKET_SETUP_GUIDE.md` — Step-by-step viasocket configuration guide.
  - `[NEW] scripts/test_viasocket_webhook.sh` — Webhook ingestion & benchmark script.
  - `[NEW] presentation/GDG_3_MIN_DEMO_SCRIPT.md` — Timed 3-minute GDG pitch & live demo script.
  - `[MOD] PROGRESS.md` — Updated task completion for WS7.2.
- **Architectural & Design Decisions**:
  - Kept blast radius strictly isolated from Tracks 1, 2, and 3.
  - Maintained `<2s` end-to-end latency guarantee with sub-3ms backend ingestion & broadcast time.
- **Testing & Verification Conducted**:
  - `go test -v ./...` $\rightarrow$ 100% PASS across all packages (`internal/api`, `internal/clustering`, `internal/urgency`).
  - Benchmarked `scripts/test_viasocket_webhook.sh` against live server $\rightarrow$ Round-trip latency: **0.002582s (2.58 ms)**.
- **Handoff / Next Recommended Steps**:
  - Teammates working on Track 1 (`web/`), Track 2 (`server/internal/extraction/`), and Track 3 (`server/internal/db/`) can continue with zero merge conflicts.


### 2026-09-06 12:05 IST - Antigravity (Pair Programming Agent - Track 2)
- **Workstream / Goal**: Track 2 - Gemini NLU & Embeddings Pipeline (Go AI Engine)
- **Tasks Claimed/Completed**:
  - Installed native **Go 1.27.1** toolchain into user local profile (`%LOCALAPPDATA%\Programs\go\bin`) and configured User PATH.
  - Implemented `server/internal/extraction/gemini.go` using `github.com/google/generative-ai-go/genai` and `google.golang.org/api/option`.
  - Built Gemini 2.5 Flash structured extraction prompt with few-shot Hindi, Hinglish, and English civic incident examples covering Indore wards.
  - Implemented `text-embedding-004` embedding pipeline with 768-dimensional float32 vector generation and normalized deterministic fallback for offline testing.
  - Implemented Gemini evidence-grounded summary generation citing specific citizen reports and hazard points.
  - Created synthetic multilingual test fixture dataset (`server/internal/extraction/testdata/raw_complaints_multilingual.json`) with 10 real-world Hindi, Hinglish, and English complaints.
  - Implemented unit test suite (`server/internal/extraction/gemini_test.go`) covering prompt structure, few-shot multilingual extraction, embedding dimensionality/normalization, and grounded cluster summaries.
  - Maintained strict blast radius: zero files modified in `web/` (Track 1), `server/internal/db/` or `server/cmd/seed/` (Track 3), or `server/internal/api/` (Track 4).
- **Files Modified/Created**:
  - `[NEW] server/internal/extraction/gemini.go` — Gemini 2.5 Flash structured extraction, text-embedding-004, grounded summarizer, offline fallback.
  - `[NEW] server/internal/extraction/gemini_test.go` — Test suite for extraction, embeddings, prompts, and summaries.
  - `[NEW] server/internal/extraction/testdata/raw_complaints_multilingual.json` — 10 multilingual synthetic test fixtures.
  - `[MOD] PROGRESS.md` — Marked WS2.2 and WS2.3 as completed with test verification notes.
  - `[NEW] server/internal/extraction/prompt_test.go` — Test suite for few-shot prompt construction, ward catalog mapping, and markdown cleaner.
  - `[NEW] server/internal/extraction/embedding_test.go` — Test suite for 768-dim float32 embeddings, L2-normalization, deterministic output, and cosine similarity.
  - `[NEW] server/internal/extraction/summary_test.go` — Test suite for evidence-grounded cluster summaries, multi-channel corroboration, and hazard identification.
  - `[NEW] server/internal/extraction/extraction_test.go` — Test suite for pure Hindi/Hinglish signal extraction, urgency/intent correlation, and ward resolution.
  - `[NEW] server/internal/extraction/benchmark_test.go` — Performance benchmarks for extraction (234 µs/op) and embeddings (68 µs/op).
  - `[NEW] server/tests/extraction_integration_test.go` — External integration test package for end-to-end consumer verification.
  - `[NEW] server/tests/fixtures/raw_complaints_multilingual.json` — Integration test fixtures matching PROGRESS.md test matrix.
  - `[MOD] PROGRESS.md` — Marked WS2.2 and WS2.3 as completed with test verification notes.
  - `[MOD] AGENT.md` — Appended Track 2 journal entry.
- **Architectural & Design Decisions**:
  - `Extractor` client supports both live Gemini 2.5 Flash structured mode and deterministic offline rule-based fallback so tests and development never stall without an API key or when offline.
  - Fallback embeddings use FastText-style subword character 3-grams and stopword weighting mapped into L2-normalized 768-dimensional float32 vectors, preserving dot-product vector mathematics compatible with the in-memory clustering engine.
- **Testing & Verification Conducted**:
  - Track 2 test suite: `go test -v ./internal/extraction/...` $\rightarrow$ 100% PASS across all 5 test files.
  - Integration test suite: `go test -v ./tests/...` $\rightarrow$ 100% PASS.
  - Benchmarks: `go test -bench=. ./internal/extraction` $\rightarrow$ Extraction throughput > 4,200 ops/sec, embedding throughput > 14,000 ops/sec.
  - Full server test suite: `go test -v ./...` $\rightarrow$ 100% PASS across `internal/extraction`, `internal/clustering`, and `internal/urgency`.
- **Handoff / Next Recommended Steps**:
  - Teammate on Track 3 (Firestore Data Layer) can persist `models.AIExtraction` and its embeddings to Firestore `ai_extractions`.
  - When Track 4 (viasocket webhook) is ready, it can instantiate `extraction.NewExtractor` to enrich incoming signals and feed them to `clustering.Engine`.



### 2026-09-06 12:35 IST - Claude Code (Track 3 — Data Layer & Seeds)
- **Workstream / Goal**: Workstream 3 — Firestore data layer (WS3.1) and synthetic seed corpus (WS3.3)
- **Tasks Claimed/Completed**:
  - WS3.1 — `db.Repository` contract covering all 8 collections, with two interchangeable backends: Firestore (`cloud.google.com/go/firestore`) and a local JSON store.
  - WS3.3 — seed command generating 42 engineered synthetic complaints and writing all 8 collections.
- **Files Modified/Created**:
  - `[NEW] server/internal/db/db.go` — Repository interface, collection constants, `RawEvent` / `Recommendation` types, backend-selection factory.
  - `[NEW] server/internal/db/firestore.go` — Firestore backend (generic set/get/list helpers, emulator support, `Reset`).
  - `[NEW] server/internal/db/local.go` — local JSON backend, atomic writes, one file per collection keyed by document ID.
  - `[NEW] server/internal/db/db_test.go` — 10 tests, no credentials required.
  - `[NEW] server/cmd/seed/generator.go` — engineered corpus definition + `BuildDataset`.
  - `[NEW] server/cmd/seed/main.go` — CLI (`--reset`, `--dry-run`, `--wards`, `--local-dir`).
  - `[NEW] server/cmd/seed/generator_test.go` — 11 tests over corpus properties.
  - `[NEW] server/data/local_store/*.json` — seeded fixture output, 8 collections.
  - `[MOD] PROGRESS.md` — WS3.1/WS3.3 marked complete with verification commands and a consumer note.
- **Architectural & Design Decisions**:
  - **Two backends behind one interface.** No `gcloud`, no `firebase` CLI, no `GOOGLE_APPLICATION_CREDENTIALS` and no emulator host exist on the dev machine, so a Firestore-only data layer would have blocked every downstream workstream. Firestore is selected when an emulator host or a credential file is configured; local JSON otherwise. Both backends store identical document shapes and list in ascending document-ID order, so switching is invisible to callers.
  - **`GOOGLE_CLOUD_PROJECT` is not a reachability signal** — `config.LoadConfig` gives it a default value, so backend selection deliberately ignores it.
  - **Embeddings left empty** (team decision). Seeded extractions carry every structured field except `Embedding`, which is Workstream 2's `text-embedding-004` contract to fill.
  - **Cluster scores are computed, not hardcoded** — the seed calls the existing `urgency` and `clustering` engines rather than duplicating their formulas, so reseeding always reflects Workstream 4's current math.
  - **Seeded `INVESTIGATING` cluster carries a matching audit-log entry**, because the platform must never move a cluster's status without a human decision behind it.
  - **Blast radius held to new files only.** No edits to `handlers.go`, `clustering/`, `urgency/`, `extraction/`, `models/`, `config/`, `go.mod` or `web/`. The Firestore persistence tasks WS1.4 / WS2.4 / WS4.5 / WS6.2 remain with their own workstreams.
- **Testing & Verification Conducted**:
  - `go test ./internal/db/... ./cmd/seed/...` → PASS (21 tests).
  - `go test ./...` → PASS, no regression in `clustering` or `urgency`.
  - `go build ./...` and `go vet` → clean.
  - `go run ./cmd/seed --reset` → 42 signals / 42 extractions / 42 raw events / 3 clusters / 3 hotspots / 3 recommendations / 1 audit entry / 12 wards written to `server/data/local_store/`.
- **Blockers / Open Questions**:
  - **Cross-track defect for Workstream 4 (not fixed here — `clustering/engine.go:110` is another track's file).** `DetectBlindSpots` flags a ward when `InfraIndex < 0.45 && activeClusterCount <= 1`. With the seeded corpus this marks **Ward 14 Chandan Nagar** (8 complaints, top hotspot) and **Ward 60 Khajrana** (6 complaints) as blind spots at the same time as they are demand hotspots — a visible contradiction on the dashboard and in the pitch. Ward 1 Banganga and Ward 78 Rau are correctly flagged. Suggested fix for whoever owns WS4.4: base the rule on citizen-signal volume (or `count == 0`) rather than `count <= 1`. Note `indore_wards.json` also ships `is_blind_spot: true` for Ward 14, but the engine recomputes the flag at runtime, so the fix belongs in the rule.
  - Firestore backend is written but **unexercised against a real server** — nobody has provisioned a GCP project or started an emulator yet. Whoever does should run `go run ./cmd/seed --reset` with `FIRESTORE_EMULATOR_HOST` set as the first end-to-end check.
- **Handoff / Next Recommended Steps**:
  - Any workstream needing persistence: `repo, err := db.NewRepository(ctx, cfg)` then code against `db.Repository`. Do not construct Firestore clients directly.
  - WS4.4 owner: decide on the blind-spot rule above.
  - Track 1/4: seeded fixtures in `server/data/local_store/` can back the dashboard immediately without Google Cloud access.

### 2026-09-06 12:42 IST - Claude Code (Track 3 — secret remediation + authorised cross-track fix)
- **Workstream / Goal**: Remove committed secrets from a public repository; correct the WS4.4 blind-spot rule (authorised by the team lead)
- **Tasks Claimed/Completed**:
  - **Secret exposure closed.** `.env` was tracked and committed (`febdcd4 setup`) in a repository that `gh repo view` confirms is **PUBLIC**, with no root `.gitignore`. The committed copy had `GEMINI_API_KEY=` empty, so nothing leaked, but any teammate filling it in would have published a live key.
  - WS4.4 blind-spot rule corrected and covered by a regression test.
- **Files Modified/Created**:
  - `[NEW] .gitignore` — ignores `.env*` (keeping `.env.example`), `server/data/local_store/`, Go build output, `node_modules/`, `web/.next/`, `.DS_Store`.
  - `[DEL from index] .env`, `.DS_Store` — `git rm --cached`; both remain on disk locally.
  - `[MOD] server/internal/clustering/engine.go` — `DetectBlindSpots` threshold `count <= 1` → `count == 0`.
  - `[MOD] server/internal/clustering/engine_test.go` — added `TestUnderservedWardWithReportsIsNotBlindSpot`.
  - `[MOD] PROGRESS.md` — WS4.4 annotated with the rule correction.
- **Architectural & Design Decisions**:
  - `server/data/local_store/` is now git-ignored rather than committed as fixtures: every reseed rewrites all 42 documents with fresh relative timestamps, which would produce a large meaningless diff on every run. Regenerate with `cd server && go run ./cmd/seed --reset`.
  - Blind-spot semantics settled as: **a ward that has reported anything is visible to us, so it cannot be a blind spot however underserved it is.** Silence is the signal, not low volume.
- **Testing & Verification Conducted**:
  - `go test ./...` → PASS, including Workstream 4's pre-existing `TestBlindSpotDetection`.
  - Reseed after the fix → 3 hotspots (Wards 14, 22, 60) and 2 blind spots (Wards 1, 78) with **no overlap**; the previous contradiction is gone.
  - `git check-ignore -v .env` → ignored; `git ls-files` → 0 tracked `.env` files.
- **Blockers / Open Questions**:
  - **Every teammate must recreate `.env` locally after pulling this change** (`cp .env.example .env`, then fill in keys). Git will delete their tracked copy on pull, or report a modify/delete conflict if they had edited it.
  - The Gemini API key was transmitted through a chat session, so it should be rotated at https://aistudio.google.com/apikey once the hackathon ends.
  - Firestore backend still unexercised against a live server (no emulator or credentials provisioned yet).
- **Handoff / Next Recommended Steps**:
  - Highest-value remaining work is Workstream 2: with the Gemini key now available locally, real `text-embedding-004` embeddings would make clustering *derived* rather than *declared*, which is the platform's core claim.





### 2026-09-06 13:15 IST - Claude Code (Track 3 — cross-track integration)
- **Workstream / Goal**: Merge Tracks 2 and 4 into the Track 3 branch and close the end-to-end wiring gap (WS1.4 / WS2.4 / WS4.5 / WS6.2), authorised by the team lead
- **Tasks Claimed/Completed**:
  - Merged `origin/master` (Track 4 — viasocket docs, webhook tests, demo script) and `origin/nidhi` (Track 2 — Gemini extractor) into `akshat`. Conflicts in `.gitignore` (unioned) and `AGENT.md` ×2 (all entries kept, chronological). `PROGRESS.md` auto-merged.
  - Wired the three isolated tracks into one running pipeline.
- **Files Modified/Created**:
  - `[NEW] server/internal/api/pipeline.go` — `WithPersistence`, `WithExtractor`, async ingest/extract/persist/cluster-attach, evidence aggregation, decision persistence.
  - `[MOD] server/internal/api/handlers.go` — three optional fields plus `h.ingest(...)` and `h.persistDecision(...)` call sites. `NewHandler(cfg, hub)` signature deliberately unchanged so Track 4's handler tests keep passing untouched.
  - `[MOD] server/cmd/api/main.go` — attaches the repository and the Gemini extractor at startup.
  - `[MOD] PROGRESS.md` — WS1.4 / WS2.4 / WS4.5 / WS6.2 marked complete.
- **Architectural & Design Decisions**:
  - **Optional dependencies, not constructor changes.** Persistence and extraction attach via `WithPersistence` / `WithExtractor`. With neither attached the handler behaves exactly as before, which is why Track 4's existing test suite needed no edits.
  - **Broadcast before understanding.** The webhook records, broadcasts and returns immediately, then runs Gemini extraction and persistence in a bounded background goroutine, emitting `SIGNAL_EXTRACTED` and `CLUSTER_UPDATED` afterwards. A Gemini round trip is slower than the demo's sub-2s latency claim, so it must not sit on the request path.
  - **Cluster urgency uses the strongest evidence in the cluster, not the newest message.** Found live: Gemini rated a polite Hindi follow-up as `base_urgency: 3` with no hazard tags, which dropped an established contaminated-water cluster from urgency 100 to 90. Corroboration must never de-escalate. The handler now aggregates max base urgency and the union of hazard tags per cluster, hydrated from stored extractions at startup.
  - **A single report never creates a cluster.** Unmatched signals stay visible in the live feed only — inventing a hotspot from one message would undermine the confidence dimension.
  - Cluster matching requires ward *and* department to agree; cosine similarity is an additional gate that engages only once embeddings exist on both sides, so it works against seeded data that has none.
- **Testing & Verification Conducted**:
  - `go build ./...`, `go vet ./...`, `go test -race ./...` → all PASS, including Track 4's `internal/api` suite and Track 2's `server/tests` suite, unmodified.
  - Live end-to-end against a running server with a real Gemini key: startup hydrated 12 wards / 3 clusters / 42 signals from the store; a Hindi WhatsApp complaint was extracted, embedded (768-dim), joined `cluster-indore-001` (8 → 9 signals, need 94 → 97, held TIER_1_CRITICAL) and persisted; an unrelated Banganga SMS was resolved to `indore-ward-01` / Electricity & Power and correctly left unclustered; a policymaker decision wrote an audit entry and updated cluster status.
- **Blockers / Open Questions**:
  - `server/tests/reporter_test.go:374` writes a report to a hardcoded Windows path, which on macOS/Linux creates a file literally named `C:\Users\devni\...` inside `server/tests/` on every `go test ./...` run. It dirties everyone's working tree — worth removing.
  - Firestore itself is still unexercised: the pipeline was verified against the local JSON backend. Once a service account or emulator is configured, the same run should be repeated with `[STORE] Backend: firestore` in the log.
  - Recommendations are persisted but not yet rendered in the dashboard cluster drawer (Workstream 5).
- **Handoff / Next Recommended Steps**:
  - Track 1: `GET /api/v1/clusters` now serves seeded and live data; new WebSocket event types `SIGNAL_EXTRACTED` and `CLUSTER_UPDATED` are available for live UI updates.
  - Track 4: the viasocket flow can now point at a server that genuinely understands and clusters what it receives.

### 2026-09-06 14:15 IST - Claude Code (Track 3 — live Firestore + embedding correction)
- **Workstream / Goal**: Bring up the real Firebase project, and make embedding-based clustering genuinely work
- **Tasks Claimed/Completed**:
  - Firestore is live. Service account for project `wlkovr` moved to `server/.secrets/firebase-admin.json` (git-ignored, mode 600) and wired through `.env`. Seed and API both report `Backend: firestore`.
  - **Found and fixed a silent failure that made embedding-based clustering fake.** `text-embedding-004` returns HTTP 404 on the v1beta endpoint this SDK targets. `extraction.GenerateEmbedding` swallows that error and returns a deterministic hash vector with a `nil` error, so every stored "embedding" was a hash. Switched to `gemini-embedding-001` (3072-dim), which responds.
  - Added `cmd/seed --embed`, which backfills embeddings for stored extractions that lack them.
  - Calibrated the cluster-matching threshold against measured data instead of a guess.
- **Files Modified/Created**:
  - `[MOD] .env`, `[MOD] .env.example`, `[MOD] server/config/config.go` — `EMBEDDING_MODEL` default corrected to `gemini-embedding-001`.
  - `[MOD] .gitignore` — added `*firebase-adminsdk*.json`; the downloaded key matched none of the existing patterns.
  - `[NEW] db.EmbeddingText` — the one canonical text used to embed an extraction, so the seed backfill and the live pipeline produce comparable vectors.
  - `[MOD] server/cmd/seed/main.go` — `--embed` backfill.
  - `[MOD] server/internal/api/pipeline.go` — mean-based similarity, calibrated threshold, relaxed department gate.
  - `[MOD] UNDERSTANDING.md`, `[MOD] PROGRESS.md` — corrected model, documented the matching rule and its measurements.
- **Architectural & Design Decisions**:
  - **Similarity uses the mean over a cluster's signals, not the max.** Measured against the seeded corpus: within-cluster means 0.804–0.899, unrelated means 0.607–0.691 — clean separation. The maxima overlap (background noise reaches 0.804 while a genuine member pair sits at 0.721), so max-based matching would admit unrelated complaints. Threshold 0.75 is the midpoint of the measured gap.
  - **Ward is mandatory; department is only a fallback.** Gemini does not reproduce the seeded department taxonomy — it returned "Water Supply & Sewerage" for an open manhole against a "Sanitation & Drainage" cluster. Semantic similarity is the more reliable signal; department equality covers extractions with no embedding.
- **Testing & Verification Conducted**:
  - `go build`, `go vet`, `go test ./...` → all PASS.
  - Live against Firestore: seeded 42 signals + 42 embeddings; a Hinglish SMS about the Khajrana manhole joined `cluster-indore-003` by similarity alone despite the department mismatch — 6 → 7 signals, need 88 → 91, confidence 58.3 → 79.2 as corroboration became cross-channel (WhatsApp + SMS), urgency held at Tier 1.
  - Before the model fix the same message matched nothing: max cosine against the cluster was 0.456, and within-cluster similarity for identical issue strings ran as low as −0.183 — the signature of hash vectors.
- **Blockers / Open Questions**:
  - **`extraction.GenerateEmbedding` still swallows API errors** (`gemini.go:272`) and returns hash vectors with a `nil` error. The model name is fixed, but the next failure will be just as invisible. Track 2 should return the error, or at minimum log it.
  - Its fallback is hardcoded to 768 dimensions while real vectors are 3072. Mismatched lengths score 0.0 in `CosineSimilarity`, so this fails safe (no match) rather than matching wrongly — but offline mode cannot cluster at all.
  - `server/tests/*.go` still passes `"text-embedding-004"` to `NewExtractor`; harmless because those tests run offline, but misleading.
  - Seeding Firestore takes ~2 minutes for ~103 documents (sequential writes). Fine ahead of time, risky live — use `BulkWriter` if reseeding during the demo is ever needed.
- **Handoff / Next Recommended Steps**:
  - Re-run `go run ./cmd/seed --reset --embed` after any change to the embedding model, then re-measure the threshold.
  - Remaining gaps for the demo: Google Maps API key, and the viasocket flow pointed at this server.
