---
change_id: CHG-0036
artifact: delta
status: implemented
updated: 2026-10-14
---

# Delta

## What Is Changing

### Mission Board

- `GET /api/v1/missions/offers/{offer_id}` returns a single published offer or 404.

## Requirements

### REQ-MB-016 — Read a single offer (added)

A consumer can retrieve one offer by id; unknown ids return OFFER_NOT_AVAILABLE.

## BDD Scenarios

### SCN-MB-OFFER-GET — Fetch existing offer

Given offer `offer-1` exists for the active season, when a member calls GET /offers/offer-1, then it returns 200 with the offer.

### SCN-MB-OFFER-404 — Unknown offer

Given `offer-x` does not exist, when it is requested, then it returns 404 OFFER_NOT_AVAILABLE.

## Error Vocabulary

- Reuses `OFFER_NOT_AVAILABLE` — 404.
