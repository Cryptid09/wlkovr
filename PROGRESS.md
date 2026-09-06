# Project Progress & Task Board (`PROGRESS.md`)

This board tracks task distribution, implementation status, and test separation across the 4-person team and AI agents.

## Legend
- `[ ]` **Not Started** — Available to be picked up
- `[/]` **In Progress** — Claimed by an agent/teammate (specify name/ID)
- `[X]` **Completed** — Built and verified with tests
- `[-]` **Blocked** — Blocked by dependency or decision

---

## Workstream Breakdown & Task Distribution

### Workstream 1: Ingestion & Realtime Intake (viasocket + Socket.IO)
- [ ] **WS1.1**: Define Canonical Citizen Signal TypeScript/Pydantic interface (`id`, `provider`, `rawText`, `language`, `location`, `timestamp`, `metadata`).
- [ ] **WS1.2**: Implement webhook endpoint receiving viasocket WhatsApp/SMS payloads $\rightarrow$ transform to Canonical Signal.
- [ ] **WS1.3**: Implement Socket.IO server & event emitter (`EVENT_CITIZEN_SIGNAL_RECEIVED`, `EVENT_CLUSTER_UPDATED`) for real-time dashboard sync.
- [ ] **WS1.4**: Persist raw payload to Firestore `raw_events` and canonical signal to `citizen_signals`.
- **Test Separation**:
  - `tests/fixtures/viasocket_sample_payload.json` (Mock webhook payload)
  - `tests/fixtures/canonical_signal_fixture.json`
  - Unit test: verify webhook conversion & Socket.IO event emission offline.

---

### Workstream 2: Extraction Pipeline (Gemini NLU)
- [ ] **WS2.1**: Define structured output schema (`issue`, `ward`, `service_category`, `urgency_1_to_5`, `intent`, `summary`).
- [ ] **WS2.2**: Implement Gemini structured output extraction prompt with Hindi/Hinglish/English few-shot examples.
- [ ] **WS2.3**: Generate text embeddings for the extracted issue description.
- [ ] **WS2.4**: Persist extraction & embeddings to Firestore `ai_extractions`.
- **Test Separation**:
  - `tests/fixtures/raw_complaints_multilingual.json` (10 synthetic mixed-language complaints)
  - `tests/test_extraction.py` with mock Gemini responses for offline test runs.

---

### Workstream 3: Data Layer & Seeds (Firebase Firestore)
- [ ] **WS3.1**: Firestore client setup & collection schema definitions (`wards`, `raw_events`, `citizen_signals`, `ai_extractions`, `clusters`, `hotspots`, `recommendations`, `audit_logs`).
- [ ] **WS3.2**: Indore Ward Reference dataset (15–20 wards with centroids, demographic/infra index, historical investment).
- [ ] **WS3.3**: Synthetic seed script (~30–50 complaints with engineered cluster hotspots and blind-spot wards).
- **Test Separation**:
  - `tests/fixtures/indore_wards.json`
  - Validation test: Seed Firestore / local Firestore emulator and verify document reads and indexes.

---

### Workstream 4: In-Memory Clustering & 4-Dimensional Scoring
- [ ] **WS4.1**: In-memory cosine similarity threshold grouping (cluster complaints by issue within the same ward).
- [ ] **WS4.2**: Implement 4 scoring dimension engines:
  - `Need` = normalized cluster size × average urgency
  - `Confidence` = channel diversity + corroborating signal count
  - `Equity` = inverse of ward infra/demographic index
  - `Actionability` = location resolved + clear department mapping heuristic
- [ ] **WS4.3**: Civic blind-spot candidate detection rule (poor infra index + low submission volume).
- [ ] **WS4.4**: Persist computed clusters and hotspots to Firestore `clusters` & `hotspots`.
- **Test Separation**:
  - `tests/test_scoring.py`: Deterministic test suite verifying math against fixture matrices (pure logic, no network).

---

### Workstream 5: Dashboard Frontend (Next.js + Socket.IO + Google Maps)
- [ ] **WS5.1**: Next.js project skeleton with Tailwind + shadcn/ui + Recharts.
- [ ] **WS5.2**: Hotspot & Blind-Spot Map view with Google Maps JS API (ward markers/polygons colored by priority).
- [ ] **WS5.3**: Priority Cluster Detail view displaying the **4 separate score bars** (Need, Confidence, Equity, Actionability) using Recharts/shadcn progress bars.
- [ ] **WS5.4**: Socket.IO client integration for real-time live complaint / hotspot animation as webhooks arrive.
- [ ] **WS5.5**: Policymaker Decision Panel (Accept / Reject / Investigate actions) writing to Firestore `audit_logs`.
- **Test Separation**:
  - `tests/fixtures/mock_dashboard_data.json` (Allows full frontend and UI development without backend dependency).
  - Mock API route in Next.js returning static fixture data.

---

### Workstream 6: Grounded Recommendation Generator
- [ ] **WS6.1**: Implement Gemini summary call per cluster generating human-readable recommendations citing specific evidence/submissions.
- [ ] **WS6.2**: Persist to Firestore `recommendations` and render in dashboard cluster drawer.
- **Test Separation**:
  - `tests/test_recommendation.py` with fixture cluster data.

---

### Workstream 7: Live Demo & End-to-End Verification
- [ ] **WS7.1**: End-to-end integration test (Simulate live WhatsApp message $\rightarrow$ viasocket $\rightarrow$ Socket.IO $\rightarrow$ Extraction $\rightarrow$ In-memory clustering $\rightarrow$ Firestore update $\rightarrow$ Live Dashboard animation).
- [ ] **WS7.2**: Demo script run-through checklist for GDG presentation.

---

## Test & Fixture Separation Matrix

| Component | Upstream Dependency | Mock / Fixture Strategy | Independent Verification Method |
|---|---|---|---|
| **Ingestion** | viasocket Webhook | `tests/fixtures/viasocket_sample_payload.json` | `curl` payload to local endpoint + Socket.IO event listener |
| **Extraction** | Gemini API | Multilingual raw text fixtures + Mock Gemini JSON | Pytest/Jest suite against mock responses |
| **Clustering/Scoring** | Database | In-memory cluster fixture arrays | Pure unit tests for scoring formulas |
| **Dashboard** | Backend API & Firestore | `mock_dashboard_data.json` fixture route | Next.js dev server with mock data toggle |
| **Live Demo** | WhatsApp | Pre-seeded Firestore + 1 live WhatsApp trigger | Verification check script |

