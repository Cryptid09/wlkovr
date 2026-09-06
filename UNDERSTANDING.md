# Project Understanding — Public Development Intelligence Platform

Team **Zen** (Nidhi Agrawal — lead, Akshat Mishra, Mandeep Yadav, Kirtan Prajapat).
Event: GDG Indore — Build with AI, "Code for Communities" 2nd Edition.
Problem statement: **AI for Digital Public Infrastructure & Governance**.

This file exists so any AI model or teammate picking up this repo has the full context without re-reading the whole conversation history. Source pitch deck: `Zen_ Build with ai Hackathon.pdf` (finalized by the team; this file supersedes it wherever the two disagree, since it reflects the hackathon-realistic cut).

## The problem

Citizens report development issues (roads, water, sanitation, etc.) through many disconnected channels — websites, WhatsApp, SMS, voice calls, call centres, public forums. Governments end up with fragmented, duplicated, per-department complaint records. Similar issues reported in different languages/channels/wording are never linked, so policymakers can't see which needs are actually widespread or urgent. Communities with low digital participation may be invisible entirely, even if their needs are serious.

## The solution, in one paragraph

An AI-assisted platform that sits *behind* existing citizen channels (no new app for citizens). It ingests submissions from multiple channels, uses Gemini to extract structured signals (issue, location, service, urgency, intent) from multilingual/mixed-language text and voice, clusters similar submissions across languages and channels, enriches these signals with public demographic/infrastructure/investment data, and surfaces both **demand hotspots** (many corroborating signals) and **civic data blind spots** (likely unmet need where participation is low). Priorities are shown across four transparent dimensions — **need, confidence, equity, actionability** — never collapsed into one opaque score. The platform recommends; it never auto-approves projects or allocates funds. A human policymaker reviews evidence and records the decision, creating an audit trail.

## Non-negotiable constraints

- **Google technologies wherever there is a genuine choice, without compromising functionality.** Don't pick a non-Google tool over a Google one unless the Google option would actually be worse for the feature.
- **viasocket is a hackathon sponsor and must be used for automation/channel orchestration.** It replaces the custom "API Gateway / Channel Adapters" layer from the original deck — viasocket receives the WhatsApp/SMS webhooks and forwards a normalized payload to one Cloud Run ingestion endpoint. Do not hand-roll WhatsApp/SMS adapter code that viasocket already covers.
- **Human-in-the-loop is core to the pitch, not a detail.** The platform must never auto-approve, auto-fund, or auto-close anything. Every "decision" is a human action recorded against the system's recommendation.
- **Build time was <12 hours from a standing start, 4-person team**, so scope has been cut hard from the original deck (see below). Treat the original PDF's full architecture as the long-term vision, not the hackathon target.

## Tech stack (locked in)

| Layer | Choice |
|---|---|
| Channel intake & Realtime | **viasocket** (webhook intake/orchestration) + **Socket.IO / WebSockets** (real-time event streaming between Go gateway & Next.js dashboard) |
| Compute / Backend Engine | **Golang (Go)** (Gin/Chi HTTP framework, Gorilla/go-socket.io WebSockets, official Google Cloud Go SDKs for Gemini & Firestore on Cloud Run) |
| AI/NLU (multilingual extraction, embeddings, summaries) | **Gemini 2.5 Flash** (via `github.com/google/generative-ai-go`), structured output mode & **`gemini-embedding-001`** (3072-dim). *`text-embedding-004` 404s on the v1beta endpoint this SDK uses — do not switch back.* |
| Data store | **Firebase Firestore** (via official `cloud.google.com/go/firestore` Go SDK, Mandeep's 7 collections approach) |
| Maps/visualization | **Leaflet + react-leaflet** with OpenStreetMap tiles. *Changed from Google Maps Platform during the build — no Maps API key was ever provisioned, and Leaflet needs none. The pitch must not claim Google Maps.* |
| Frontend Dashboard | **Next.js 15 (TypeScript)** + Tailwind CSS + shadcn/ui + Recharts |

### Pitch Doc vs. Hackathon Build Scope
- **Pitch Doc (`mandeep.md`)**: Preserves the complete long-term architecture diagram with full multi-modal pipelines (Voice STT, Email OCR, WhatsApp Image OCR, Citizen Portal, Pub/Sub Event Bus).
- **Hackathon Build Scope**: Focuses on text/voice-note WhatsApp intake via viasocket + Go Socket.IO/WebSocket server, Gemini structured extraction in Go, Firestore collections, and Next.js policymaker dashboard. Cut for demo: live telephony, manual email OCR, Firebase Auth, live geocoding API calls.

## Data model essentials (Firestore Collections)

Following Mandeep's data architecture, Firestore stores state across separate collections to ensure traceability and auditability:
- `wards`: Indore ward reference dataset (~15–20 wards with centroids, demographic/infra index, historical investment).
- `raw_events`: Raw webhook event payloads from viasocket / Socket.IO.
- `citizen_signals`: Normalized canonical citizen signal (ID, provider, rawText, language, location, timestamp, metadata).
- `ai_extractions`: Gemini structured extractions (issue, category, department, severity/urgency, location, summary, embeddings).
- `clusters` / `hotspots`: Clustered issues with 4D scores (Need, Confidence, Equity, Actionability) and corroborating signal references.
- `recommendations`: Grounded AI policy recommendations citing cluster evidence.
- `audit_logs`: Policymaker human decisions (Accept / Reject / Investigate actions, timestamps, and notes).

**Access layer** (`server/internal/db`): all collections are reached through one `db.Repository` interface — never by constructing a Firestore client directly. It has two interchangeable backends: **Firestore** when `FIRESTORE_EMULATOR_HOST` or `GOOGLE_APPLICATION_CREDENTIALS` is configured, and a **local JSON store** (`server/data/local_store/*.json`) otherwise, so every workstream can develop and test without Google Cloud access. Both store identical document shapes and list in ascending document-ID order. Two document types live in the `db` package rather than `models`: `RawEvent` (the untouched viasocket payload) and `Recommendation` (a grounded brief plus the signal IDs it cites).

- **Seed submissions**: ~30–50 synthetic citizen submissions across those wards, deliberately constructed so:
  - 3–4 wards get 5+ submissions each, in different languages/channels/wording, describing the *same* underlying issue (proves clustering).
  - 2–3 wards get almost no submissions but score poorly on the infra/demographic index (proves blind-spot detection).
  - The rest are background noise.
- During the live demo, 1–2 *real* messages are submitted live via WhatsApp/Socket.IO so judges see the actual pipeline fire in real-time alongside seeded history.

## Clustering approach

Gemini embeddings on the extracted issue text $\rightarrow$ in-memory cosine similarity threshold grouping (constrained to the same ward, since similar complaints in different wards are distinct issues) $\rightarrow$ write computed clusters and hotspots directly to Firestore. Simple, fast, and eliminates heavy external vector DB dependencies during hackathon execution.

**Matching rule** (`server/internal/api/pipeline.go`): a live signal joins a cluster when it is in the same ward *and* either its **mean** cosine similarity to the cluster's embedded signals clears **0.75**, or its department matches exactly (the fallback for extractions with no embedding).

Department values must come from the fixed six-department taxonomy defined in the Gemini prompt (`server/internal/extraction/gemini.go`), which the seed also uses. An earlier mismatch — the seed inventing names like "Sanitation & Drainage" that the prompt never lists — silently broke department matching, so **anything writing a department must use that vocabulary**.

Both numbers are measured, not guessed. Against the seeded corpus embedded with `gemini-embedding-001`:

| | mean cosine |
|---|---|
| Within cluster (same issue) | 0.804 – 0.899 |
| Across clusters (different issues) | 0.607 – 0.691 |
| Background noise vs cluster | 0.615 – 0.691 |

Lowest within-cluster mean 0.804 vs highest unrelated mean 0.691 → threshold at the 0.75 midpoint, ~0.11 margin either side. **Mean, not max**: the maxima overlap (background noise reaches 0.804 against a cluster while a genuine member pair can sit at 0.721), so one coincidentally similar sentence must not pull an unrelated complaint in. Re-measure if the embedding model changes.

A signal matching nothing stays in the live feed unclustered — a single report never creates a hotspot.

## Urgency Decision Engine

The Urgency Decision Engine dynamically computes an actionable urgency score and assigns a response SLA tier based on multi-factor civic signals:

### 1. Multi-Factor Scoring Model
$$\text{Urgency Score (0–100)} = \min\left(100, \, (\bar{U}_{\text{base}} \times 20) \times H_{\text{hazard}} \times (1 + \alpha \cdot V_{\text{velocity}}) \times S_{\text{sensitivity}}\right)$$

- **$\bar{U}_{\text{base}}$ (Base Urgency)**: Average urgency rating (1–5) extracted by Gemini from citizen reports.
- **$H_{\text{hazard}}$ (Hazard Multiplier, 1.0–2.0x)**: High-risk civic categories (e.g., contaminated drinking water, exposed live electrical wire, open manhole, hospital emergency route blocked $\rightarrow 1.8\times–2.0\times$).
- **$V_{\text{velocity}}$ (Temporal Spike Rate)**: Rate of incoming complaints within a sliding window (e.g., $>5$ complaints in $<2$ hours indicates an active emergency burst).
- **$S_{\text{sensitivity}}$ (Infrastructure Sensitivity, 1.0–1.3x)**: Proximity to critical facilities (hospitals, schools, major transit intersections, flood-prone zones).

### 2. Tiered Urgency Classification & SLAs
| Tier | Score Range | Label | Recommended SLA | Dashboard Action / Indicator |
|---|---|---|---|---|
| **Tier 1** | 80–100 | **Critical Emergency** | **< 4 Hours** | Red pulsing banner, top of priority queue, emergency dispatch recommendation |
| **Tier 2** | 60–79 | **High Urgency** | **< 24 Hours** | Orange badge, high-priority queue placement |
| **Tier 3** | 35–59 | **Medium Priority** | **< 72 Hours** | Yellow badge, standard department routing |
| **Tier 4** | 0–34 | **Routine Maintenance** | **< 7 Days** | Green badge, scheduled municipal maintenance |

---

## Scoring — four dimensions (defined here since the original deck named them but never specified the math)

- **Need** = normalized cluster size × Urgency Decision Engine score (0–1)
- **Confidence** = corroboration strength — number of distinct channels + evidence types backing the cluster, normalized
- **Equity** = inverse of the ward's socioeconomic/infra index — a poor, underserved ward scores higher here even with fewer raw complaints
- **Actionability** = heuristic completeness check — is location resolved? is affected service mapped to a single clear department? (simple boolean/percentage, not ML)

These four are shown **as separate bars on the dashboard**, never collapsed into one hidden "priority score" — that's the deck's explainability claim and it must hold in the actual UI.

Blind spots use a plain rule: ward has a poor infra/demographic index but submission count is below a threshold → flag as blind-spot candidate.

## Backend: running it

The Go backend is the whole pipeline — ingestion, Gemini extraction, clustering, scoring, persistence and the live WebSocket feed. It runs standalone.

```bash
cd server
go run ./cmd/seed --reset --embed   # load the demo corpus into Firestore, then embed it
go run ./cmd/api                    # serve on :8080
```

`.env` at the repo root drives both. Two variables decide where data goes:

| Variable | Effect |
|---|---|
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to the Firebase service account JSON. **Set → Firestore. Unset → local JSON files** under `server/data/local_store/`. |
| `GEMINI_API_KEY` | Set → live Gemini extraction and embeddings. Unset → deterministic offline fallbacks, and no embeddings are stored. |

The startup log states which of each is in effect — always check these two lines before demoing:

```
[STORE]  Hydrated from firestore — 12 wards, 3 clusters, 42 signals, 42 embeddings
[GEMINI] Extractor attached — live structured extraction and embeddings enabled
```

`Backend: local-json` or `OFFLINE mode` means credentials are missing. The server runs fine either way, which is the point — but the demo needs both live.

**The credentials file is machine-local.** `.env` holds an absolute path to a service account JSON that is git-ignored and not in the repo. Anyone running the backend needs their own copy of that file and must repoint `GOOGLE_APPLICATION_CREDENTIALS` at it.

Reseeding takes roughly two minutes to write ~103 documents to Firestore plus three minutes to embed 42 extractions. **Seed before presenting, never during.**

## Backend: API and WebSocket contract

Base URL `http://localhost:8080/api/v1`. Every REST response is wrapped:

```jsonc
{ "success": true, "data": <payload>, "message": "...", "error": "..." }
```

| Method | Path | Returns |
|---|---|---|
| GET | `/health` | server status |
| GET | `/wards` | `Ward[]` — 12 Indore wards with centroid lat/lng, infra index, `is_blind_spot` |
| GET | `/clusters` | `Cluster[]` — hotspots with the four scores, urgency tier, recommendation |
| GET | `/clusters/:id` | one `Cluster` (404 when unknown) |
| GET | `/signals` | `CitizenSignal[]` — live feed, newest first, capped at 50 |
| GET | `/audit-logs` | `AuditLog[]` — policymaker decision history |
| POST | `/decisions` | records a human decision: `{cluster_id, action, officer, notes}` |
| POST | `/webhooks/viasocket` | ingestion endpoint: `{provider, sender, body}` |
| POST | `/demo/simulate` | fires a synthetic citizen message through the identical path |

Go structs in `server/internal/models/models.go` are the source of truth for every shape; JSON field names come from their `json:` tags.

### WebSocket `ws://localhost:8080/ws`

Every frame is `{ "type": string, "timestamp": string, "data": object }`. Four types are emitted:

| Type | When | `data` |
|---|---|---|
| `SIGNAL_RECEIVED` | immediately on ingestion, before Gemini runs | `CitizenSignal` |
| `SIGNAL_EXTRACTED` | once Gemini has understood it | `AIExtraction` |
| `CLUSTER_UPDATED` | when a signal joins a cluster and it is rescored | `Cluster` |
| `DECISION_RECORDED` | when a policymaker acts | `{cluster, audit}` |

The ordering matters for the demo. `SIGNAL_RECEIVED` fires in milliseconds so a WhatsApp message appears instantly; Gemini takes several seconds, so `SIGNAL_EXTRACTED` and `CLUSTER_UPDATED` arrive after. A signal that matches no cluster produces the first two events and no third — a single report never creates a hotspot.


## Build order / workstreams (parallelizable across the 4-person team)

1. **Ingestion & Realtime** — viasocket webhook + Socket.IO real-time event pipeline $\rightarrow$ Canonical Citizen Signal $\rightarrow$ write to Firestore `raw_events` & `citizen_signals`
2. **Extraction** — Gemini structured output (issue, ward, service, urgency, intent) + embeddings per submission $\rightarrow$ write to `ai_extractions`
3. **Data layer + seed** — Firestore collections setup, ward CSV, seed script (~30-50 complaints)
4. **Clustering & Urgency Decision Engine** — In-memory cosine threshold clustering + multi-factor Urgency Decision Engine + 4-dimension scoring engine + blind spot detection $\rightarrow$ write to `clusters` & `hotspots`
5. **Dashboard** — Next.js + Tailwind + shadcn/ui + Recharts: Google Maps + priority queue + Urgency Tier badges + 4 score bars + human decision action buttons + `audit_logs`
6. **Recommendation text** — Gemini grounded summary citing evidence $\rightarrow$ write to `recommendations`

Status: **full system wired and verified end to end** — Next.js dashboard ↔ Go backend ↔ live Firestore ↔ live Gemini. A multilingual WhatsApp/SMS message posted to the viasocket webhook is extracted, embedded, matched to an existing cluster, rescored and persisted, with every step broadcast over the WebSocket. Seeded corpus: 42 signals, 3 hotspot clusters, 2 blind-spot wards, 12 wards.

Remaining: the Google Maps API key (`NEXT_PUBLIC_GOOGLE_MAPS_API_KEY` is empty), the frontend on its own branch, and pointing a real viasocket flow at `/api/v1/webhooks/viasocket` (see `docs/viasocket/VIASOCKET_SETUP_GUIDE.md`).

Related files: `rules.md` (collaboration & engineering rules), `PROGRESS.md` (task board & test separation), `AGENT.md` (agent work journal), `mandeep.md` (pitch architecture document).
