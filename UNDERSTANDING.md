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
| Channel intake & Realtime | **viasocket** (webhook intake/orchestration) + **Socket.IO** (real-time event streaming between gateway, processing & dashboard) |
| Compute / Backend | **Cloud Run / Next.js API Routes** (ingestion, processing, WebSocket server, dashboard) |
| AI/NLU (multilingual extraction, embeddings, summaries) | **Gemini** (via Vertex AI / Gemini API), structured output mode |
| Data store | **Firebase Firestore** (Mandeep's collections approach: `raw_events`, `citizen_signals`, `ai_extractions`, `clusters`, `hotspots`, `recommendations`, `audit_logs`) |
| Maps/visualization | **Google Maps Platform** (JS API for the hotspot map) |
| Frontend | Next.js + Tailwind + shadcn/ui + Recharts on Cloud Run / Firebase Hosting |

### Pitch Doc vs. Hackathon Build Scope
- **Pitch Doc (`mandeep.md`)**: Preserves the complete long-term architecture diagram with full multi-modal pipelines (Voice STT, Email OCR, WhatsApp Image OCR, Citizen Portal, Pub/Sub Event Bus).
- **Hackathon Build Scope**: Focuses on text/voice-note WhatsApp intake via viasocket + Socket.IO, Gemini structured extraction, Firestore collections, and Next.js policymaker dashboard. Cut for demo: live telephony, manual email OCR, Firebase Auth, live geocoding API calls.

## Data model essentials (Firestore Collections)

Following Mandeep's data architecture, Firestore stores state across separate collections to ensure traceability and auditability:
- `wards`: Indore ward reference dataset (~15–20 wards with centroids, demographic/infra index, historical investment).
- `raw_events`: Raw webhook event payloads from viasocket / Socket.IO.
- `citizen_signals`: Normalized canonical citizen signal (ID, provider, rawText, language, location, timestamp, metadata).
- `ai_extractions`: Gemini structured extractions (issue, category, department, severity/urgency, location, summary, embeddings).
- `clusters` / `hotspots`: Clustered issues with 4D scores (Need, Confidence, Equity, Actionability) and corroborating signal references.
- `recommendations`: Grounded AI policy recommendations citing cluster evidence.
- `audit_logs`: Policymaker human decisions (Accept / Reject / Investigate actions, timestamps, and notes).

- **Seed submissions**: ~30–50 synthetic citizen submissions across those wards, deliberately constructed so:
  - 3–4 wards get 5+ submissions each, in different languages/channels/wording, describing the *same* underlying issue (proves clustering).
  - 2–3 wards get almost no submissions but score poorly on the infra/demographic index (proves blind-spot detection).
  - The rest are background noise.
- During the live demo, 1–2 *real* messages are submitted live via WhatsApp/Socket.IO so judges see the actual pipeline fire in real-time alongside seeded history.

## Clustering approach

Gemini embeddings on the extracted issue text $\rightarrow$ in-memory cosine similarity threshold grouping (constrained to the same ward, since similar complaints in different wards are distinct issues) $\rightarrow$ write computed clusters and hotspots directly to Firestore. Simple, fast, and eliminates heavy external vector DB dependencies during hackathon execution.

## Scoring — four dimensions (defined here since the original deck named them but never specified the math)

- **Need** = normalized cluster size × average urgency (urgency comes from Gemini's extraction per submission)
- **Confidence** = corroboration strength — number of distinct channels + evidence types backing the cluster, normalized
- **Equity** = inverse of the ward's socioeconomic/infra index — a poor, underserved ward scores higher here even with fewer raw complaints
- **Actionability** = heuristic completeness check — is location resolved? is affected service mapped to a single clear department? (simple boolean/percentage, not ML)

These four are shown **as separate bars on the dashboard**, never collapsed into one hidden "priority score" — that's the deck's explainability claim and it must hold in the actual UI.

Blind spots use a plain rule: ward has a poor infra/demographic index but submission count is below a threshold → flag as blind-spot candidate.

## Build order / workstreams (parallelizable across the 4-person team)

1. **Ingestion & Realtime** — viasocket webhook + Socket.IO real-time event pipeline $\rightarrow$ Canonical Citizen Signal $\rightarrow$ write to Firestore `raw_events` & `citizen_signals`
2. **Extraction** — Gemini structured output (issue, ward, service, urgency, intent) + embeddings per submission $\rightarrow$ write to `ai_extractions`
3. **Data layer + seed** — Firestore collections setup, ward CSV, seed script (~30-50 complaints)
4. **Clustering + scoring** — In-memory cosine threshold clustering per ward + 4-dimension scoring engine + blind spot detection $\rightarrow$ write to `clusters` & `hotspots`
5. **Dashboard** — Next.js + Tailwind + shadcn/ui + Recharts: Google Maps + priority queue + 4 score bars + human decision action buttons + `audit_logs`
6. **Recommendation text** — Gemini grounded summary citing evidence $\rightarrow$ write to `recommendations`

Status as of this writing: **Architecture & Stack Aligned** — ready for implementation.

Related files: `rules.md` (collaboration & engineering rules), `PROGRESS.md` (task board & test separation), `AGENT.md` (agent work journal), `mandeep.md` (pitch architecture document).
