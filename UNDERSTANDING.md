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
| Channel intake (WhatsApp/SMS webhooks) | **viasocket** → forwards to Cloud Run |
| Compute | **Cloud Run** (ingestion, processing, dashboard) |
| AI/NLU (multilingual extraction, embeddings, summaries) | **Gemini** (via Vertex AI / Gemini API), structured output mode |
| Data store | **Cloud SQL for PostgreSQL** with **pgvector** (embeddings) + **PostGIS** (geo) in one DB — no separate vector DB service |
| Maps/visualization | **Google Maps Platform** (JS API for the hotspot map) |
| Frontend | Next.js dashboard on Cloud Run/Firebase Hosting |

### Cut for the hackathon build (do not build these unless time allows / explicitly asked)
- **Cloud Pub/Sub** — not needed at demo data volumes; direct Cloud Run→Cloud Run calls instead.
- **Live voice calls / telephony / call centres** — the hard part is the telephony leg, not the AI. Cut entirely for v1.
- **Voice notes via Speech-to-Text** — good v1.5 addition if time remains (viasocket already delivers WhatsApp audio files; one Cloud Speech-to-Text call transcribes them), but not required for the core demo.
- **Live Google Maps Geocoding API calls** — ward centroids are hardcoded in the seed dataset instead; Gemini extracts the ward *name*, matched by lookup, not geocoded live.
- **Firebase Auth / RBAC** — demo dashboard is unlocked; auth is a roadmap line only.
- **Vertex AI Model Evaluation** — MLOps nicety, not a demo feature.
- Everything under "Scalability across India and BRICS" and "Future Development" in the original deck is pitch narrative only — no engineering time against it.

## Data model essentials

- **Ward reference dataset**: ~15–20 Indore wards, hardcoded, each with: ward name/ID, centroid lat/long, a demographic/infra index, past public investment. Real open data if findable quickly; plausible synthetic numbers otherwise — this is fine to say out loud, it's a demo dataset.
- **Seed submissions**: ~30–50 synthetic citizen submissions across those wards, deliberately constructed so:
  - 3–4 wards get 5+ submissions each, in different languages/channels/wording, describing the *same* underlying issue (this is what proves clustering works).
  - 2–3 wards get almost no submissions but score poorly on the infra/demographic index (this is what proves blind-spot detection works).
  - The rest are background noise.
- During the live demo, 1–2 *real* messages are submitted live via WhatsApp so judges see the actual pipeline fire, landing alongside the seeded history. Be upfront that historical volume is seeded and the live path is real — don't blur this distinction if asked.

## Clustering approach

Gemini embeddings on the extracted issue text → pgvector cosine similarity, threshold-based grouping (constrained to same ward, since two similar complaints in different wards are different issues). No need for Vertex AI Vector Search or any heavyweight clustering algorithm at this data scale — a simple threshold/union-find grouping is sufficient and realistic.

## Scoring — four dimensions (defined here since the original deck named them but never specified the math)

- **Need** = normalized cluster size × average urgency (urgency comes from Gemini's extraction per submission)
- **Confidence** = corroboration strength — number of distinct channels + evidence types backing the cluster, normalized
- **Equity** = inverse of the ward's socioeconomic/infra index — a poor, underserved ward scores higher here even with fewer raw complaints
- **Actionability** = heuristic completeness check — is location resolved? is affected service mapped to a single clear department? (simple boolean/percentage, not ML)

These four are shown **as separate bars on the dashboard**, never collapsed into one hidden "priority score" — that's the deck's explainability claim and it must hold in the actual UI.

Blind spots use a plain rule: ward has a poor infra/demographic index but submission count is below a threshold → flag as blind-spot candidate. No anomaly-detection model needed.

## Build order / workstreams (parallelizable across the 4-person team)

1. **Ingestion** — viasocket → WhatsApp webhook → Cloud Run endpoint → common signal schema → write raw submission to Postgres
2. **Extraction** — Gemini structured output (issue, ward, service, urgency, intent) per submission
3. **Data layer + seed** — Postgres (pgvector + PostGIS), ward CSV, seed script
4. **Clustering + scoring** — embedding threshold clustering, then the four formulas above (plain app logic)
5. **Dashboard** — Next.js: map + priority list + four score bars + accept/reject/investigate buttons + decision log
6. **Recommendation text** — one Gemini call per cluster, generating a grounded summary citing the cluster's actual evidence; bolt on last

Status as of this writing: **nothing built yet** — this document was written at the planning stage.

Related file: `rules.md` — collaboration rules for any AI model working in this repo. Read it before making changes.
