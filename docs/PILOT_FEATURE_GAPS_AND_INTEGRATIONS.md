# NEXUS Pilot Feature Gaps And Integration Guide

Date: 2026-04-29

This note captures the next product gaps and the practical integration path for printers and Meta WhatsApp Cloud API. The target remains a single-terminal full-service restaurant pilot.

## What The Research Implies

Restaurant POS products that feel complete generally converge on the same operational spine:

- Fast order entry with searchable menu groups, modifiers, split checks, discounts, service charge, and table transfer.
- Kitchen routing by station, with KOT/KDS status flowing back to the table view.
- Menu availability controls so an item can be active, out of stock, or hidden without deleting it.
- Ingredient-level inventory tied to recipes, waste, receiving, purchase orders, and reorder suggestions.
- End-of-day reports with cash variance, tax, refunds, voids, waiter/category sales, and downloadable exports.
- Printer health and queued print jobs, because restaurants need failed prints to be visible and retryable.
- WhatsApp receipts and campaigns using approved templates and opt-in status, not informal device automation.

Useful source references:

- Toast order screens mention menus, groups, search, modifiers, hold/send/pay, split, discounts, transfers, voids, and reprint kitchen tickets.
- Toast KDS highlights prep station routing, ticket status, production counts, timers, and reporting.
- Square Restaurants highlights menu customization, live table status, hold/start firing, close-of-day reports, station printing, inventory, online ordering, and QR/tableside ordering.
- Epson documents ESC/POS as the command set for direct thermal printer control.
- Apple documents the normal macOS path for USB, Bluetooth, AirPrint, IPP, LPD, and HP Jetdirect/Socket printers.
- Meta's Postman collection for WhatsApp Cloud API documents WABA subscription, phone number ID lookup, phone registration, permissions, and the `/{Phone-Number-ID}/messages` send endpoint.

## Implemented In This Pass

- Menu groups can be created from Menu Setup and used while adding/editing sale items.
- New menu items can be created with category, route, price, cost, and GST.
- Modifiers/add-ons can be created with price delta, route, and status.
- Modifiers can be attached to specific sale items.
- Item Master now edits category, kitchen route, sale price, food cost, GST, and availability status.
- Hidden items are removed from Counter/Table selling screens.
- Out-of-stock items stay visible but disabled, so operators understand why an item cannot be sold.
- Stock and Purchases now support custom ingredients, vendors, and multi-line purchase orders.
- Settings now has a staff directory for people working in the restaurant.
- Table opening now uses a selected active staff member and guest count, so assignment is explicit.
- Home now shows pilot readiness and the next incomplete setup action.
- Day Close now exports a PDF report for owner/accountant review.
- Counter billing now avoids typing for the rush-hour path: service mode, table, recent customer, discounts, payment method, and exact tender are tap/select driven.
- Devices now exposes printer modes for network TCP, system print, USB adapter, and Bluetooth adapter paths.
- Print jobs are visible and actionable with payload preview, attempts, errors, Printed, Failed, and Retry states.
- Razorpay payment requests now support local ready, paid/reconciled, failed, and cancelled states.
- Settings now includes launch readiness and a recent security audit log.

## Next Missing Feature List

### Menu Setup

- Bulk CSV preview before import: validate rows, show errors, then commit.
- Duplicate detection by normalized name/category/route before saving.
- Item images and short kitchen names.
- Mandatory modifier groups, for example "Choose size" or "Choose bread".
- Modifier quantity rules: min/max, free quantity, paid extras.
- Item-level prep time, route override for modifiers, and print name.
- Menu version history and price-change audit.

### POS And Tables

- Search-first item entry.
- Counter bill table entry is now a dropdown; next add seat/person-level item routing and table-session bill pickup from the counter.
- Seat numbers and course/firing controls.
- Staff assignment already exists for opening a table; next add change-server and waiter sales reports.
- Item void/comp from table session with manager PIN.
- Table merge/split at the table view, not only invoice split.
- Reprint KOT and receipt history visible from both table and invoice.
- Open item/custom item for one-off sales.

### Kitchen

- Station filters: Barista, Kitchen, Dessert, Expeditor.
- Aging timers and color changes.
- All-day production counts.
- Re-fire item and cancel item flows.
- KOT print retry and failure reason beside each ticket. The generic print queue now supports this; next wire each ticket directly to route printer jobs.

### Inventory And Purchases

- Purchase order statuses: draft, sent, partial, received, closed, cancelled.
- Vendor item catalog and last price comparison.
- Goods received note and vendor debit note PDFs.
- Stock count batches, variance approval, and waste reason catalog.
- Reorder suggestions using par, velocity, lead time, and waste.

### Accounting And Reports

- Downloadable PDFs and CSVs for day close, sales register, tax register, stock movement, purchases, vendor ledger, and cash book.
- Day-close PDF exists; next add sales register, tax register, stock movement, vendor ledger, and cash book PDFs.
- Day-close lock so closed days cannot be mutated without approval.
- Source links from accounting vouchers back to invoice/payment/refund/purchase/adjustment.
- Tax summary by GST rate.
- Razorpay clearing report and settlement reconciliation.

## Printer Integration Plan

NEXUS should support three printer modes:

1. Network ESC/POS: `tcp://192.168.1.50:9100`
   - Best for pilots.
   - Send raw ESC/POS bytes directly over TCP port 9100.
   - Keep printers on the restaurant LAN, not exposed to the internet.

2. System printer: `system://Receipt Printer`
   - Best for USB and Bluetooth printers on macOS.
   - User pairs/adds the printer in macOS Printers & Scanners.
   - NEXUS sends a PDF/plain receipt through the OS print queue.

3. Serial/Bluetooth raw path: `serial:///dev/tty.SomePrinter`
   - Use only where the printer exposes a serial port.
   - More fragile than system print because every model behaves differently.

Implementation steps:

- Expand `PrinterConnectionInput` to include route, paper width, copies, and mode.
- Add print adapter interfaces in Go: `TCPAdapter`, `SystemAdapter`, `SerialAdapter`. The persisted queue and status transitions are already in place.
- Generate ESC/POS receipt/KOT bytes centrally, then hand those bytes to the selected adapter.
- Add `Test Print` and `Printer Health` actions.
- Persist every print attempt in `print_jobs` with retries, `last_error`, and `printed_at`.

## Meta WhatsApp Cloud API Plan

Required settings:

- Meta Business Portfolio ID.
- WhatsApp Business Account ID.
- Phone Number ID.
- System User Access Token with WhatsApp permissions.
- App Secret.
- Webhook verify token.
- Template names and language codes for receipt, refund, vendor debit note, customer offer, and birthday/return offers.

Desktop limitation:

- A desktop Wails app cannot reliably receive public HTTPS webhooks from Meta. For pilot, use send-only plus local delivery state. For production, add a small HTTPS relay service that receives Meta webhooks, verifies the signature, and forwards status updates into NEXUS sync/cloud.

Implementation steps:

- Store credentials encrypted in local settings and never expose secrets to React except during the active save/test call.
- Keep `notification_queue` as the source of outbound messages.
- Add queue states: `ready`, `sending`, `sent`, `failed`, `retrying`, `skipped`.
- Send via `POST https://graph.facebook.com/{version}/{phone_number_id}/messages`.
- Use template messages for receipts and outbound marketing unless the customer is inside the 24-hour customer-service window.
- Record Meta `wamid`, error code, error message, template name, recipient, and payload snapshot per queue row.
- Add opt-in and quiet-hour checks before marketing sends.
