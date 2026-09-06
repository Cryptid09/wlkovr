# Rules for AI models working in this repo

This repo is being built by a 4-person team under hackathon time pressure, likely with more than one AI assistant in the loop (different tools, different sessions). These rules exist to keep every model's output consistent with the others'. Read `UNDERSTANDING.md` first — it has the full problem/solution context. This file is about *how to behave* while working on it.

## Before doing anything
- Read `UNDERSTANDING.md`. If it's missing or looks stale relative to the actual code, say so rather than silently working around it.
- If you change scope, architecture, data model, or scoring logic, **update `UNDERSTANDING.md` in the same change**. It's the shared source of truth other models/teammates rely on — stale docs are worse than no docs.

## Locked decisions — do not relitigate without being asked
- Google technologies wherever there's a genuine choice, without sacrificing functionality. Don't swap in a non-Google service to save a few minutes of setup.
- **viasocket** is the channel/webhook orchestration layer (sponsor requirement). Don't build custom WhatsApp/SMS adapters that duplicate what viasocket already does.
- The platform **recommends, never auto-decides**. No feature should auto-approve a project, auto-allocate funds, or auto-close a citizen report without a human action. If a task seems to ask for that, flag it instead of building it.
- The four priority dimensions (need, confidence, equity, actionability) are shown **separately**, never collapsed into one opaque score.

## Hackathon scope discipline
- The cut list in `UNDERSTANDING.md` (no Pub/Sub, no live telephony, no live geocoding calls, no Firebase Auth, no Vertex AI Model Eval) is intentional, not an oversight. Don't add these back in "for correctness" or "for production-readiness" — this is a demo, not a production deploy. If you think one is actually needed, ask first.
- Don't add abstractions, config options, or generalized data models beyond what's needed for the Indore ward demo dataset. Three hardcoded wards beat a generic multi-city configuration system right now.
- No premature error handling for scenarios that can't occur in the demo. Validate real boundaries (webhook payloads, user input) — don't defensively guard against internal calls between your own services.

## Data and secrets
- API keys (Gemini/Vertex, Google Maps, viasocket webhook secret, DB credentials) go in environment variables, never hardcoded or committed.
- Seed/synthetic data is fine and expected (see `UNDERSTANDING.md`'s seed strategy) — just don't present synthetic submissions as real citizen data in demo copy or UI labels without being clear internally about which is which.

## Collaboration hygiene
- Stick to your workstream (see the build-order list in `UNDERSTANDING.md`) unless coordinating a handoff — four people/models editing the same files concurrently causes more merge pain than the time it saves.
- Keep commits/changes scoped to one workstream at a time so teammates (human or AI) can tell what changed and why.
- If you hit a decision that isn't covered by `UNDERSTANDING.md` or this file, don't silently pick an approach that contradicts the rest of the system — surface the choice instead of guessing.
