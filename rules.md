# Rules for AI models and Developers working in this repo

This repo is being built by a 4-person team under hackathon time pressure, with multiple AI assistants (different tools, IDEs, and sessions). These rules enforce consistency, clean handoffs, test separation, and industry-standard practices across all sessions.

---

## 1. Mandatory Workflow for Every Agent Session

Every AI agent/model MUST follow this lifecycle on every invocation:

### Step 1: Context Ingestion (Before touching any code)
1. Read `UNDERSTANDING.md` for architecture, problem scope, and locked constraints.
2. Read `rules.md` (this file) for operational rules and collaboration standards.
3. Check `PROGRESS.md` to see current workstream statuses, active ownership, and blocked items.
4. Read the latest entries in `AGENT.md` to understand what previous agents/teammates completed, what decisions were made, and what state the code is in.

### Step 2: Workstream Scoping & Task Claiming
- Claim your target task in `PROGRESS.md` by marking it `[/] In Progress` with your agent/session info.
- Do not work on tasks claimed by another active agent/teammate without explicit handoff.
- Strictly adhere to the locked tech stack and cut list in `UNDERSTANDING.md`.

### Step 3: Implementation & Test Separation
- Apply industry-standard practices: modularity, strict schema validation, clear boundary error handling, and test separation.
- Write/update tests alongside code changes. Keep tests isolated (unit tests and fixture-based tests that do not require live external APIs).

### Step 4: Documentation & Handoff (Before finishing)
1. **Update `PROGRESS.md`**: Update task completion status, add sub-tasks, and update verification/test notes.
2. **Append to `AGENT.md`**: Record an entry in the agent journal with timestamp, actions taken, decisions made, obstacles encountered, and handoff instructions.
3. **Sync `UNDERSTANDING.md`**: If any data model, API schema, formula, or architecture evolved, update `UNDERSTANDING.md` in the exact same change.

---

## 2. Agent Journal Protocol (`AGENT.md`)

`AGENT.md` is the persistent chronological journal of all AI agent activities across all tools and sessions.

- **Append-only**: Never delete or overwrite previous agent entries.
- **Entry Structure**: Every session must append an entry using the following template:

```markdown
### [YYYY-MM-DD HH:MM UTC/IST] - <Agent/Model Name or ID>
- **Workstream / Goal**: [e.g., Workstream 2 - Gemini Extraction Pipeline]
- **Tasks Claimed/Completed**: [e.g., Implemented Pydantic signal schema and structured output prompt]
- **Files Modified/Created**:
  - `[NEW/MOD] path/to/file` — brief description
- **Architectural & Design Decisions**: [Any non-obvious choices, trade-offs made, or schema decisions]
- **Testing & Verification Conducted**: [Tests run, results, fixtures verified]
- **Blockers / Open Questions**: [Any issues needing human or next-agent attention]
- **Handoff / Next Recommended Steps**: [Clear guidance for the next agent picking up this work]
```

---

## 3. Progress Tracking & Test Separation Protocol (`PROGRESS.md`)

`PROGRESS.md` coordinates multi-agent task distribution and ensures a clean separation between development and testing.

- **Workstream Organization**: Tasks are organized into the 6 build workstreams + cross-cutting tasks.
- **Status Indicators**:
  - `[ ]` Not Started
  - `[/]` In Progress (Must include assignee/agent identifier)
  - `[X]` Completed (Must include verification proof or test reference)
  - `[-]` Blocked / Paused (Must include reason in notes)
- **Test Separation Principle**:
  - Every workstream must have a dedicated test/mock section in `PROGRESS.md`.
  - Development must define mock contracts and test fixtures early so that downstream components (e.g., Dashboard or Clustering) can be built and tested independently of live upstream dependencies (e.g., WhatsApp webhooks or live Cloud SQL).
  - Include automated verification scripts/commands in `PROGRESS.md` for fast validation.

---

## 4. Locked Decisions — Do Not Relitigate Without Explicit Instruction

- **Core Tech Stack**: **Golang (Go)** for Backend Engine (Ingestion, Gemini AI pipelines, In-memory clustering, Urgency Engine, and Firestore) + **Next.js 15 (TypeScript)** for Frontend Dashboard.
- **Google Tech Stack**: Google technologies wherever a genuine choice exists (Gemini 2.5 Flash via official Go SDK, Cloud Run, Firebase Firestore, Leaflet + OpenStreetMap for the map — no Maps API key was provisioned, so any doc claiming Google Maps is wrong).
- **viasocket + WebSockets/Socket.IO**: viasocket handles webhook ingestion from WhatsApp/SMS channels; Go WebSocket/Socket.IO server streams events in real-time to the Next.js dashboard.
- **Data Store**: **Firebase Firestore** using Mandeep's collection architecture (`raw_events`, `citizen_signals`, `ai_extractions`, `clusters`, `hotspots`, `recommendations`, `audit_logs`).
- **Human-in-the-loop (Recommend, Never Auto-Decide)**: No auto-approval, auto-allocation of funds, or auto-closure of reports. Recommendations only.
- **Transparent 4-Dimensional Scoring**: Need, confidence, equity, and actionability must always remain separate visible scores/bars, never collapsed into a single opaque number.

---

## 5. Hackathon Scope Discipline & Cut List

- **Respect the cut list**: No Pub/Sub, no live telephony, no live geocoding API calls (use hardcoded ward reference tables), no Firebase Auth/RBAC, no Vertex AI Model Eval.
- **No over-engineering**: Three hardcoded Indore wards beat a generic multi-city system. Solve the concrete demo problem first.
- **Boundary validation only**: Validate external untrusted boundaries (webhooks, Gemini JSON responses). Do not write defensive boilerplate for internal in-memory function calls.

---

## 6. Industry-Standard Engineering Practices

### A. Contract-First & Schema Validation
- Define schemas (e.g., Pydantic models, TypeScript interfaces, JSON Schemas) before implementing business logic.
- Validate Gemini structured outputs rigorously against the defined schema and handle JSON parse errors gracefully with fallback defaults.

### B. Test Isolation & Mockability
- **Fixtures over Live APIs**: Store synthetic raw webhook payloads and Gemini output fixtures under `tests/fixtures/` or `data/seeds/`.
- **Standalone Execution**: Modules must be runnable standalone with mock inputs (e.g., `python -m src.extraction.extract --test-file ...` or Jest/Pytest suites).
- Test scoring formulas and clustering algorithms against synthetic datasets with deterministic edge cases.

### C. Secrets & Environment Configuration
- Never commit credentials, API keys, or database passwords.
- Maintain `.env.example` with all required environment variables and descriptive placeholder comments.
- Access configs via structured config loaders/environment variables.

### D. Clean Code & Commit Hygiene
- Keep changes modular, single-responsibility, and scoped to one workstream per commit/PR.
- Maintain readable, idiomatic code with clean type hints (TypeScript / Python type annotations).
- Preserve existing comments and docstrings. Update documentation simultaneously when making code changes.

