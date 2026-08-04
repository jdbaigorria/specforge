# Resync Detection

If `sf-propose` is invoked on a feature that already has artefacts:

1. Check which artefacts exist (`requirements.md`, `design.md`, `tasks.md`).
2. Compare timestamps to find the upstream/downstream relationship.
3. If an upstream artefact is newer than a downstream one → it may be **stale**.
   Offer to regenerate the downstream artefact(s) so they reflect the change.
4. If the user modified an artefact directly → acknowledge the edit and cascade
   the implications downstream (e.g. a requirements edit may invalidate design).

The goal is to never silently build on top of an artefact that no longer matches
the artefact it was derived from. When in doubt, surface the mismatch and let the
user decide whether to regenerate or keep.
