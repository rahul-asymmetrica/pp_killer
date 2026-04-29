# NEXUS Build Context

NEXUS is an offline-first restaurant operating system aimed at the parts of PetPooja that feel expensive, cluttered, and hard to train. The first build prioritizes the local truth layer: fast counter billing, recipe-linked stock movement, delivery exceptions, and signal-style insights.

## Product Principles

- The desktop app must stay useful without internet.
- Every operational write goes to local SQLite first.
- Stock math belongs in Go services, not in the React UI.
- Inventory should handle real restaurant messiness: waste, bad batches, partial receiving, and physical audit drift.
- The dashboard should show actions and exceptions, not spreadsheet-style walls of numbers.

## Current Implemented Slice

- Wails + Go + React TypeScript application scaffold.
- Local SQLite store under the user's application config directory, for example `~/Library/Application Support/NEXUS/nexus.db` on macOS. Packaged apps cannot rely on their launch working directory being writable.
- Seeded demo cafe, menu items, ingredients, recipes, customers, and marketing drafts.
- Recipe BOM deduction when a bill is recorded.
- Full checkout fields for order type, table, discounts, taxes, payment method, invoice number, and payment tendered/change.
- KOT generation by kitchen route, with one ticket per route for mixed coffee/kitchen orders.
- Dynamic waste factor update during physical stock reconciliation.
- Ingredient purchase units versus usage units, including reorder point and last purchase cost.
- Editable recipe BOMs from the UI through Go-backed validation.
- Menu setup now includes menu groups, menu item creation, item master edits, modifiers/add-ons, modifier-to-item linking, GST, food cost, route, and availability status.
- Hidden menu items are kept out of selling screens, while out-of-stock items remain visible but disabled for operator clarity.
- Delivery receiving with accepted and rejected quantities split.
- Custom ingredient creation, vendor creation, and multi-line purchase order drafting from the UI.
- Staff directory is exposed in Settings with name, role, optional PIN, and active/inactive status.
- Floor/table opening uses the selected active staff member and guest count instead of a hardcoded waiter.
- Counter billing is tap-first for service mode, table selection, recent customer, discount presets, payment method, and exact tender.
- Stock movement is intentionally tied to KOT/quick-close, not draft save; the UI now states this rule near the bill controls.
- Home now includes a pilot-readiness guide that points the operator to the next missing setup step.
- Day close can be exported as a PDF report, alongside invoice PDFs and accounting/invoice CSVs.
- Devices now has a real operator queue for printer jobs: receipt/KOT jobs capture payload, target, status, attempts, errors, retry state, and printed state.
- Printer setup supports pilot-facing target modes for network TCP, macOS/system print, USB adapter, and Bluetooth adapter paths.
- Razorpay payment requests now have a lifecycle beyond "created": ready, paid/reconciled, failed, and cancelled.
- Settings now exposes launch checklist readiness plus a security audit trail for approvals, settings, integrations, staff, payments, and printer actions.
- Pending `sync_log` records for sale, delivery, and reconciliation events.
- React operations dashboard with POS tiles, stock state, customer signals, delivery exceptions, and drafts.
- Go tests for sale deduction, rejected delivery stock handling, waste recalibration, payment lifecycle, and print job failure/retry/printed lifecycle.

## Engine Notes

### POS And Recipe Consumption

When a menu item is billed, Go loads its recipe components and deducts:

`component_qty * sold_qty * (1 + ingredient_waste_factor)`

The bill is closed even if inventory would run low. In a restaurant, billing cannot stop during a rush; low stock becomes a signal.

### Receiving And Bad Batch Handling

Deliveries are recorded as batches. Each line splits ordered, accepted, and rejected quantities. Only accepted quantity is added to live stock. Rejected quantity is logged as an exception movement, ready for a later vendor debit note or WhatsApp message.

### Reconciliation And Waste Drift

Physical audits do not simply overwrite stock. The reconciliation flow compares theoretical and physical stock, writes the adjustment, and nudges the ingredient waste factor based on consumption since the last audit. This is the start of the self-correcting inventory engine.

### Sync Bridge

The schema includes `sync_log` as the future bridge from SQLite to Postgres. Every core local operation writes a pending sync item with entity, operation, payload, and timestamp. The next backend milestone is a Go worker that pushes these rows to a cloud API and marks them synced after acknowledgement.

## Near-Term Build Roadmap

1. Add CSV preview/validation, mandatory modifier groups, item images, kitchen names, prep time, and price-change audit.
2. Add staff permissions by role, shift clock-in/out, and waiter sales/tip reports.
3. Implement actual adapter execution for network ESC/POS, macOS system print, and serial/Bluetooth raw paths. The queue and UI state machine now exist.
4. Add Meta WhatsApp Cloud API send worker and template registry, with webhook relay deferred to cloud.
5. Add vendor ledger, debit note generation, and PO send/receive status lifecycle.
6. Add Postgres sync service with idempotent upserts and conflict policy per entity.
7. Add menu intelligence: item contribution margin, velocity, deadstock recipes, seasonal views.
8. Add marketing pipeline stubs: brand kit, photo intake, Canva export state, Meta/WhatsApp dispatch queues.
