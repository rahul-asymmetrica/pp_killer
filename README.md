# NEXUS

Offline-first restaurant OS built with Wails, Go, React, and SQLite. The current slice focuses on the operational core: cart checkout, invoice numbers, discounts, taxes, payments, KOT routing, recipe-linked inventory movement, editable recipe BOMs, delivery rejection handling, waste recalibration, customer signals, and a sync queue ready for a Postgres bridge.

## Run

```bash
wails dev
```

## Verify

```bash
go test ./...
npm run build --prefix frontend
```

## Current Shape

- Backend domain code lives in `internal/nexus`.
- Wails bindings are exposed from `app.go`.
- Local data is created under the user's application config directory, for example `~/Library/Application Support/NEXUS/nexus.db` on macOS.
- Product context and next milestones are in `docs/NEXUS_BUILD_CONTEXT.md`.
