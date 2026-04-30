package nexus

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	invoiceStatusDraft             = "draft"
	invoiceStatusKOTSent           = "kot_sent"
	invoiceStatusClosed            = "closed"
	invoiceStatusVoided            = "voided"
	invoiceStatusRefunded          = "refunded"
	invoiceStatusPartiallyRefunded = "partially_refunded"
	invoiceStatusSplit             = "split"
)

type preparedSaleLine struct {
	item     MenuItem
	quantity int
	notes    string
	subtotal float64
}

type existingInvoiceLine struct {
	id        string
	itemID    string
	itemName  string
	quantity  int
	unitPrice float64
	lineTotal float64
	notes     string
	routeID   string
	routeName string
	taxRate   float64
}

func (s *Store) SaveOrder(input SaleInput) (Dashboard, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	invoiceID, invoiceNumber, _, err := createInvoiceTx(tx, input, createdAt)
	if err != nil {
		return Dashboard{}, err
	}
	if err := insertInvoiceEventTx(tx, invoiceID, "draft_saved", "Draft saved", fmt.Sprintf("%s is ready to reopen", invoiceNumber), "operator", createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "sale", invoiceID, "create", map[string]any{"invoiceNumber": invoiceNumber, "status": invoiceStatusDraft, "input": input}, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) RecordSale(input SaleInput) (Dashboard, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	invoiceID, invoiceNumber, _, err := createInvoiceTx(tx, input, createdAt)
	if err != nil {
		return Dashboard{}, err
	}
	if err := sendKOTTx(tx, invoiceID, createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := closeInvoiceTx(tx, CloseInvoiceInput{
		InvoiceID:       invoiceID,
		PaymentMethod:   input.PaymentMethod,
		PaymentTendered: input.PaymentTendered,
	}, createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "sale", invoiceID, "create", map[string]any{"invoiceNumber": invoiceNumber, "status": invoiceStatusClosed, "input": input}, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) SendKOT(invoiceID string) (Dashboard, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	if err := sendKOTTx(tx, strings.TrimSpace(invoiceID), createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "sale", invoiceID, "kot_sent", map[string]any{"invoiceID": invoiceID}, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) CloseInvoice(input CloseInvoiceInput) (Dashboard, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	if err := closeInvoiceTx(tx, input, createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "sale", input.InvoiceID, "close", input, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) VoidInvoice(input VoidInvoiceInput) (Dashboard, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	approval, err := approveManagerTx(tx, "void", input.InvoiceID, input.PIN, createdAt)
	if err != nil {
		return Dashboard{}, err
	}

	summary, err := invoiceSummaryByIDTx(tx, input.InvoiceID)
	if err != nil {
		return Dashboard{}, err
	}
	if err := ensureTimestampBusinessDateOpenTx(tx, summary.CreatedAt); err != nil {
		return Dashboard{}, err
	}
	var ticketCount int
	if err := tx.QueryRow("SELECT COUNT(*) FROM kitchen_tickets WHERE sale_id = ?", input.InvoiceID).Scan(&ticketCount); err != nil {
		return Dashboard{}, err
	}
	if summary.Status != invoiceStatusDraft || ticketCount > 0 || summary.KOTSentAt != "" {
		return Dashboard{}, errors.New("only unsent draft invoices can be voided without stock reversal")
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "Voided by manager"
	}
	if _, err := tx.Exec(`
		UPDATE sales
		SET status = ?, void_reason = ?, approved_by = ?, approved_at = ?
		WHERE id = ?
	`, invoiceStatusVoided, reason, approval.ApprovedBy, approval.ApprovedAt, input.InvoiceID); err != nil {
		return Dashboard{}, err
	}
	if err := insertInvoiceEventTx(tx, input.InvoiceID, "voided", "Invoice voided", reason, approval.ApprovedBy, createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "sale", input.InvoiceID, "void", input, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) RefundInvoice(input RefundInvoiceInput) (Dashboard, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	approval, err := approveManagerTx(tx, "refund", input.InvoiceID, input.PIN, createdAt)
	if err != nil {
		return Dashboard{}, err
	}
	summary, err := invoiceSummaryByIDTx(tx, input.InvoiceID)
	if err != nil {
		return Dashboard{}, err
	}
	if err := ensureTimestampBusinessDateOpenTx(tx, summary.CreatedAt); err != nil {
		return Dashboard{}, err
	}
	if summary.Status != invoiceStatusClosed && summary.Status != invoiceStatusPartiallyRefunded {
		return Dashboard{}, errors.New("only closed invoices can be refunded")
	}

	var refunded float64
	if err := tx.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM refunds WHERE invoice_id = ?", input.InvoiceID).Scan(&refunded); err != nil {
		return Dashboard{}, err
	}
	remaining := round2(summary.Total - refunded)
	amount := round2(input.Amount)
	if amount <= 0 {
		amount = remaining
	}
	if amount <= 0 || amount > remaining+0.01 {
		return Dashboard{}, fmt.Errorf("refund exceeds remaining invoice value %.2f", remaining)
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = "Customer refund"
	}
	refundID := mustID()
	if _, err := tx.Exec(`
		INSERT INTO refunds (id, invoice_id, amount, reason, approved_by, approved_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, refundID, input.InvoiceID, amount, reason, approval.ApprovedBy, approval.ApprovedAt, createdAt); err != nil {
		return Dashboard{}, err
	}
	if _, err := tx.Exec(`
		INSERT INTO payments (id, invoice_id, method, amount, tendered, change_due, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, mustID(), input.InvoiceID, "refund", -amount, -amount, 0, "refunded", createdAt); err != nil {
		return Dashboard{}, err
	}

	nextStatus := invoiceStatusPartiallyRefunded
	if math.Abs(remaining-amount) <= 0.01 {
		nextStatus = invoiceStatusRefunded
	}
	if _, err := tx.Exec(`
		UPDATE sales
		SET status = ?, refund_reason = ?, approved_by = ?, approved_at = ?
		WHERE id = ?
	`, nextStatus, reason, approval.ApprovedBy, approval.ApprovedAt, input.InvoiceID); err != nil {
		return Dashboard{}, err
	}
	if err := insertInvoiceEventTx(tx, input.InvoiceID, "refunded", "Refund recorded", fmt.Sprintf("%s refunded: %s", formatINR(amount), reason), approval.ApprovedBy, createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := queueInvoiceNotificationTx(tx, input.InvoiceID, "refund_update", "Refund recorded", createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "sale", input.InvoiceID, "refund", input, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) SplitInvoice(input SplitInvoiceInput) (Dashboard, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	mode := strings.ToLower(strings.TrimSpace(input.Mode))
	if mode == "" {
		mode = "items"
	}
	summary, err := invoiceSummaryByIDTx(tx, input.InvoiceID)
	if err != nil {
		return Dashboard{}, err
	}
	if err := ensureTimestampBusinessDateOpenTx(tx, summary.CreatedAt); err != nil {
		return Dashboard{}, err
	}
	switch mode {
	case "items", "item":
		if err := splitInvoiceByItemsTx(tx, input, createdAt); err != nil {
			return Dashboard{}, err
		}
	case "amount", "amounts":
		if err := splitInvoiceByAmountsTx(tx, input, createdAt); err != nil {
			return Dashboard{}, err
		}
	default:
		return Dashboard{}, fmt.Errorf("unsupported split mode: %s", input.Mode)
	}
	if err := insertSyncLog(tx, "sale", input.InvoiceID, "split", input, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) ValidateManagerPIN(pin string) (ApprovalToken, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return ApprovalToken{}, err
	}
	defer rollback(tx)

	createdAt := nowUTC()
	approval, err := approveManagerTx(tx, "validate", "", pin, createdAt)
	if err != nil {
		return ApprovalToken{}, err
	}
	if err := tx.Commit(); err != nil {
		return ApprovalToken{}, err
	}
	return approval, nil
}

func (s *Store) GetInvoices(filter InvoiceFilter) ([]InvoiceSummary, error) {
	clauses := []string{"1 = 1"}
	args := make([]any, 0)
	if status := strings.TrimSpace(filter.Status); status != "" && status != "all" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if method := strings.TrimSpace(filter.PaymentMethod); method != "" && method != "all" {
		clauses = append(clauses, "payment_method = ?")
		args = append(args, method)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		clauses = append(clauses, "(invoice_number LIKE ? OR customer_name LIKE ? OR customer_phone LIKE ? OR table_name LIKE ?)")
		args = append(args, like, like, like, like)
	}
	if filter.DateFrom != "" {
		clauses = append(clauses, "created_at >= ?")
		args = append(args, filter.DateFrom)
	}
	if filter.DateTo != "" {
		clauses = append(clauses, "created_at <= ?")
		args = append(args, filter.DateTo)
	}

	rows, err := s.db.Query(`
		SELECT id, invoice_number, customer_name, customer_phone, channel, order_type, table_name,
			subtotal, discount_total, tax_total, total, payment_method, status,
			source_invoice_id, split_group_id, void_reason, refund_reason, approved_by, approved_at,
			created_at, closed_at, kot_sent_at
		FROM sales
		WHERE `+strings.Join(clauses, " AND ")+`
		ORDER BY created_at DESC
		LIMIT 100
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]InvoiceSummary, 0)
	for rows.Next() {
		var invoice InvoiceSummary
		if err := scanInvoiceSummary(rows, &invoice); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, rows.Err()
}

func (s *Store) GetInvoiceDetail(invoiceID string) (InvoiceDetail, error) {
	var detail InvoiceDetail
	tx, err := s.db.Begin()
	if err != nil {
		return detail, err
	}
	defer rollback(tx)

	summary, err := invoiceSummaryByIDTx(tx, invoiceID)
	if err != nil {
		return detail, err
	}
	lines, err := invoiceLinesTx(tx, invoiceID)
	if err != nil {
		return detail, err
	}
	payments, err := invoicePaymentsTx(tx, invoiceID)
	if err != nil {
		return detail, err
	}
	refunds, err := invoiceRefundsTx(tx, invoiceID)
	if err != nil {
		return detail, err
	}
	events, err := invoiceEventsTx(tx, invoiceID)
	if err != nil {
		return detail, err
	}
	notifications, err := invoiceNotificationsTx(tx, invoiceID)
	if err != nil {
		return detail, err
	}
	tickets, err := invoiceKitchenTicketsTx(tx, invoiceID)
	if err != nil {
		return detail, err
	}

	if err := tx.Commit(); err != nil {
		return detail, err
	}
	detail = InvoiceDetail{
		Summary:        summary,
		Lines:          lines,
		Payments:       payments,
		Refunds:        refunds,
		Events:         events,
		KitchenTickets: tickets,
		Notifications:  notifications,
	}
	return detail, nil
}

func (s *Store) GetNotifications() ([]NotificationRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, invoice_id, channel, recipient, template, payload, status, error, created_at, sent_at, read_at
		FROM notification_queue
		ORDER BY created_at DESC
		LIMIT 80
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]NotificationRecord, 0)
	for rows.Next() {
		var notification NotificationRecord
		if err := rows.Scan(
			&notification.ID,
			&notification.InvoiceID,
			&notification.Channel,
			&notification.Recipient,
			&notification.Template,
			&notification.Payload,
			&notification.Status,
			&notification.Error,
			&notification.CreatedAt,
			&notification.SentAt,
			&notification.ReadAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

func (s *Store) MarkNotificationRead(id string) error {
	_, err := s.db.Exec("UPDATE notification_queue SET read_at = ? WHERE id = ?", nowUTC(), id)
	return err
}

func createInvoiceTx(tx *sql.Tx, input SaleInput, createdAt string) (string, string, float64, error) {
	if len(input.Lines) == 0 {
		return "", "", 0, errors.New("invoice requires at least one line")
	}
	if strings.TrimSpace(input.Channel) == "" {
		input.Channel = "counter"
	}
	if strings.TrimSpace(input.OrderType) == "" {
		input.OrderType = "dine_in"
	}
	if strings.TrimSpace(input.PaymentMethod) == "" {
		input.PaymentMethod = "cash"
	}
	if err := ensureTimestampBusinessDateOpenTx(tx, createdAt); err != nil {
		return "", "", 0, err
	}

	invoiceID := mustID()
	invoiceNumber, err := nextDocumentNumber(tx, "invoice_prefix", "next_invoice_seq")
	if err != nil {
		return "", "", 0, err
	}

	lines, subtotal, discountTotal, taxTotal, total, err := prepareSaleLinesTx(tx, input)
	if err != nil {
		return "", "", 0, err
	}

	if _, err := tx.Exec(`
		INSERT INTO sales (
			id, invoice_number, customer_name, customer_phone, table_name, order_type,
			subtotal, discount_total, tax_total, total, channel,
			payment_method, payment_tendered, payment_change, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, invoiceID, invoiceNumber, strings.TrimSpace(input.CustomerName), normalizePhone(input.CustomerPhone), strings.TrimSpace(input.TableName), input.OrderType,
		subtotal, discountTotal, taxTotal, total, input.Channel, input.PaymentMethod, 0, 0, invoiceStatusDraft, createdAt); err != nil {
		return "", "", 0, err
	}

	for _, line := range lines {
		if _, err := tx.Exec(`
			INSERT INTO sale_lines (id, sale_id, item_id, item_name, quantity, unit_price, line_total, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, mustID(), invoiceID, line.item.ID, line.item.Name, line.quantity, line.item.Price, line.subtotal, line.notes); err != nil {
			return "", "", 0, err
		}
	}

	return invoiceID, invoiceNumber, total, nil
}

func prepareSaleLinesTx(tx *sql.Tx, input SaleInput) ([]preparedSaleLine, float64, float64, float64, float64, error) {
	lines := make([]preparedSaleLine, 0, len(input.Lines))
	subtotal := 0.0
	for _, line := range input.Lines {
		if line.Quantity <= 0 {
			return nil, 0, 0, 0, 0, errors.New("invoice quantity must be positive")
		}
		item, err := menuItemByID(tx, line.ItemID)
		if err != nil {
			return nil, 0, 0, 0, 0, err
		}
		if item.Status != "active" {
			return nil, 0, 0, 0, 0, fmt.Errorf("%s is not active", item.Name)
		}
		lineTotal := item.Price * float64(line.Quantity)
		subtotal += lineTotal
		lines = append(lines, preparedSaleLine{
			item:     item,
			quantity: line.Quantity,
			notes:    strings.TrimSpace(line.Notes),
			subtotal: lineTotal,
		})
	}

	discountTotal := calculateDiscount(subtotal, input.DiscountType, input.DiscountValue)
	taxTotal := 0.0
	for _, line := range lines {
		share := 0.0
		if subtotal > 0 {
			share = line.subtotal / subtotal
		}
		taxRate := line.item.TaxRate
		if input.TaxRate > 0 {
			taxRate = input.TaxRate
		}
		taxTotal += (line.subtotal - (discountTotal * share)) * taxRate
	}
	discountTotal = round2(discountTotal)
	taxTotal = round2(taxTotal)
	total := round2(subtotal - discountTotal + taxTotal)
	return lines, subtotal, discountTotal, taxTotal, total, nil
}

func sendKOTTx(tx *sql.Tx, invoiceID string, createdAt string) error {
	if invoiceID == "" {
		return errors.New("invoice id is required")
	}
	summary, err := invoiceSummaryByIDTx(tx, invoiceID)
	if err != nil {
		return err
	}
	if summary.Status == invoiceStatusVoided || summary.Status == invoiceStatusRefunded || summary.Status == invoiceStatusSplit {
		return fmt.Errorf("cannot send KOT for %s invoice", summary.Status)
	}
	if err := ensureTimestampBusinessDateOpenTx(tx, summary.CreatedAt); err != nil {
		return err
	}

	var existingTickets int
	if err := tx.QueryRow("SELECT COUNT(*) FROM kitchen_tickets WHERE sale_id = ?", invoiceID).Scan(&existingTickets); err != nil {
		return err
	}
	if existingTickets > 0 || summary.KOTSentAt != "" {
		if summary.Status == invoiceStatusDraft {
			_, err := tx.Exec("UPDATE sales SET status = ?, kot_sent_at = CASE WHEN kot_sent_at = '' THEN ? ELSE kot_sent_at END WHERE id = ?", invoiceStatusKOTSent, createdAt, invoiceID)
			return err
		}
		return nil
	}

	lines, err := invoiceLinesForTotalsTx(tx, invoiceID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return errors.New("invoice has no lines to send to kitchen")
	}

	ticketIDs := map[string]string{}
	for _, line := range lines {
		components, err := recipeComponents(tx, line.itemID)
		if err != nil {
			return err
		}
		for _, component := range components {
			deduction := component.quantity * float64(line.quantity) * (1 + component.wasteFactor)
			if _, err := tx.Exec("UPDATE ingredients SET on_hand_qty = on_hand_qty - ? WHERE id = ?", deduction, component.ingredientID); err != nil {
				return err
			}
			if _, err := tx.Exec(`
				INSERT INTO inventory_movements (id, ingredient_id, type, quantity, reason, reference_id, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, mustID(), component.ingredientID, "sale_consumption", -deduction, fmt.Sprintf("%s x%d", line.itemName, line.quantity), invoiceID, createdAt); err != nil {
				return err
			}
		}

		routeID := line.routeID
		if routeID == "" {
			routeID = "route_kitchen"
		}
		ticketID := ticketIDs[routeID]
		if ticketID == "" {
			routeName := line.routeName
			if routeName == "" {
				routeName, err = routeNameByID(tx, routeID)
				if err != nil {
					return err
				}
			}
			ticketNumber, err := nextDocumentNumber(tx, "kot_prefix", "next_kot_seq")
			if err != nil {
				return err
			}
			ticketID = mustID()
			ticketIDs[routeID] = ticketID
			if _, err := tx.Exec(`
				INSERT INTO kitchen_tickets (id, sale_id, invoice_number, ticket_number, route_id, route_name, status, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, ticketID, invoiceID, summary.InvoiceNumber, ticketNumber, routeID, routeName, "queued", createdAt); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`
			INSERT INTO kitchen_ticket_lines (id, ticket_id, item_id, item_name, quantity, notes, status)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, mustID(), ticketID, line.itemID, line.itemName, line.quantity, line.notes, "queued"); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`
		UPDATE sales
		SET status = CASE WHEN status = ? THEN status ELSE ? END,
			kot_sent_at = ?
		WHERE id = ?
	`, invoiceStatusClosed, invoiceStatusKOTSent, createdAt, invoiceID); err != nil {
		return err
	}
	return insertInvoiceEventTx(tx, invoiceID, "kot_sent", "KOT sent", fmt.Sprintf("%d kitchen route ticket(s) created", len(ticketIDs)), "operator", createdAt)
}

func closeInvoiceTx(tx *sql.Tx, input CloseInvoiceInput, createdAt string) error {
	input.InvoiceID = strings.TrimSpace(input.InvoiceID)
	if input.InvoiceID == "" {
		return errors.New("invoice id is required")
	}
	summary, err := invoiceSummaryByIDTx(tx, input.InvoiceID)
	if err != nil {
		return err
	}
	if summary.Status == invoiceStatusClosed {
		return errors.New("invoice is already closed")
	}
	if summary.Status == invoiceStatusVoided || summary.Status == invoiceStatusRefunded || summary.Status == invoiceStatusSplit {
		return fmt.Errorf("cannot close %s invoice", summary.Status)
	}
	if err := ensureTimestampBusinessDateOpenTx(tx, summary.CreatedAt); err != nil {
		return err
	}
	if summary.Status == invoiceStatusDraft && summary.Total <= 0 {
		return errors.New("cannot close an empty invoice")
	}
	if summary.KOTSentAt == "" {
		var lineCount int
		if err := tx.QueryRow("SELECT COUNT(*) FROM sale_lines WHERE sale_id = ?", input.InvoiceID).Scan(&lineCount); err != nil {
			return err
		}
		if lineCount > 0 {
			if err := sendKOTTx(tx, input.InvoiceID, createdAt); err != nil {
				return err
			}
		}
	}

	method := strings.ToLower(strings.TrimSpace(input.PaymentMethod))
	if method == "" {
		method = summary.PaymentMethod
	}
	if method == "" {
		method = "cash"
	}
	tendered := input.PaymentTendered
	if tendered <= 0 || method == "upi" || method == "card" {
		tendered = summary.Total
	}
	changeDue := round2(math.Max(0, tendered-summary.Total))

	if _, err := tx.Exec(`
		INSERT INTO payments (id, invoice_id, method, amount, tendered, change_due, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, mustID(), input.InvoiceID, method, summary.Total, tendered, changeDue, "captured", createdAt); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE sales
		SET payment_method = ?, payment_tendered = ?, payment_change = ?, status = ?, closed_at = ?
		WHERE id = ?
	`, method, tendered, changeDue, invoiceStatusClosed, createdAt, input.InvoiceID); err != nil {
		return err
	}

	if summary.CustomerPhone != "" {
		favoriteItem := ""
		_ = tx.QueryRow("SELECT item_name FROM sale_lines WHERE sale_id = ? ORDER BY rowid LIMIT 1", input.InvoiceID).Scan(&favoriteItem)
		name := strings.TrimSpace(summary.CustomerName)
		if name == "" {
			name = "Walk-in Guest"
		}
		if _, err := tx.Exec(`
			INSERT INTO customers (id, name, phone, total_spend, visit_count, last_visit_at, favorite_item)
			VALUES (?, ?, ?, ?, 1, ?, ?)
			ON CONFLICT(phone) DO UPDATE SET
				name = excluded.name,
				total_spend = customers.total_spend + excluded.total_spend,
				visit_count = customers.visit_count + 1,
				last_visit_at = excluded.last_visit_at,
				favorite_item = excluded.favorite_item
		`, mustID(), name, summary.CustomerPhone, summary.Total, createdAt, favoriteItem); err != nil {
			return err
		}
	}

	if err := insertInvoiceEventTx(tx, input.InvoiceID, "closed", "Bill closed", fmt.Sprintf("Payment captured by %s", strings.ToUpper(method)), "operator", createdAt); err != nil {
		return err
	}
	return queueInvoiceNotificationTx(tx, input.InvoiceID, "receipt", "Bill closed", createdAt)
}

func splitInvoiceByItemsTx(tx *sql.Tx, input SplitInvoiceInput, createdAt string) error {
	if len(input.Lines) == 0 {
		return errors.New("item split requires selected lines")
	}
	source, err := invoiceSummaryByIDTx(tx, input.InvoiceID)
	if err != nil {
		return err
	}
	if source.Status == invoiceStatusVoided || source.Status == invoiceStatusRefunded || source.Status == invoiceStatusSplit {
		return fmt.Errorf("cannot split %s invoice", source.Status)
	}
	lines, err := invoiceLinesForTotalsTx(tx, input.InvoiceID)
	if err != nil {
		return err
	}
	selection := map[string]int{}
	for _, line := range input.Lines {
		if line.Quantity <= 0 {
			return errors.New("split line quantity must be positive")
		}
		selection[line.LineID] += line.Quantity
	}

	childLines := make([]existingInvoiceLine, 0)
	sourceLines := make([]existingInvoiceLine, 0, len(lines))
	for _, line := range lines {
		selectedQty := selection[line.id]
		if selectedQty == 0 {
			sourceLines = append(sourceLines, line)
			continue
		}
		if selectedQty > line.quantity {
			return fmt.Errorf("split quantity exceeds %s quantity", line.itemName)
		}
		child := line
		child.quantity = selectedQty
		child.lineTotal = round2(child.unitPrice * float64(selectedQty))
		childLines = append(childLines, child)

		remaining := line.quantity - selectedQty
		if remaining > 0 {
			line.quantity = remaining
			line.lineTotal = round2(line.unitPrice * float64(remaining))
			sourceLines = append(sourceLines, line)
		} else if _, err := tx.Exec("DELETE FROM sale_lines WHERE id = ?", line.id); err != nil {
			return err
		}
	}
	if len(childLines) == 0 {
		return errors.New("no matching invoice lines were selected")
	}
	if len(sourceLines) == 0 {
		return errors.New("keep at least one item on the source invoice")
	}

	splitGroupID := source.SplitGroupID
	if splitGroupID == "" {
		splitGroupID = mustID()
	}
	childID := mustID()
	childInvoiceNumber, err := nextDocumentNumber(tx, "invoice_prefix", "next_invoice_seq")
	if err != nil {
		return err
	}

	sourceDiscount, childDiscount := splitDiscount(source.DiscountTotal, source.Subtotal, sumLineSubtotal(sourceLines), sumLineSubtotal(childLines))
	sourceSubtotal, sourceTax, sourceTotal := totalsForExistingLines(sourceLines, sourceDiscount)
	childSubtotal, childTax, childTotal := totalsForExistingLines(childLines, childDiscount)

	if _, err := tx.Exec(`
		UPDATE sales
		SET subtotal = ?, discount_total = ?, tax_total = ?, total = ?, split_group_id = ?
		WHERE id = ?
	`, sourceSubtotal, sourceDiscount, sourceTax, sourceTotal, splitGroupID, source.ID); err != nil {
		return err
	}
	for _, line := range sourceLines {
		if _, err := tx.Exec("UPDATE sale_lines SET quantity = ?, line_total = ? WHERE id = ?", line.quantity, line.lineTotal, line.id); err != nil {
			return err
		}
	}

	childStatus := invoiceStatusDraft
	childKOTSentAt := ""
	if source.KOTSentAt != "" {
		childStatus = invoiceStatusKOTSent
		childKOTSentAt = source.KOTSentAt
	}
	if _, err := tx.Exec(`
		INSERT INTO sales (
			id, invoice_number, customer_name, customer_phone, table_name, order_type,
			subtotal, discount_total, tax_total, total, channel,
			payment_method, payment_tendered, payment_change, status,
			source_invoice_id, split_group_id, kot_sent_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, childID, childInvoiceNumber, source.CustomerName, source.CustomerPhone, source.TableName, source.OrderType,
		childSubtotal, childDiscount, childTax, childTotal, source.Channel, source.PaymentMethod, 0, 0, childStatus,
		source.ID, splitGroupID, childKOTSentAt, createdAt); err != nil {
		return err
	}
	for _, line := range childLines {
		if _, err := tx.Exec(`
			INSERT INTO sale_lines (id, sale_id, item_id, item_name, quantity, unit_price, line_total, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, mustID(), childID, line.itemID, line.itemName, line.quantity, line.unitPrice, line.lineTotal, line.notes); err != nil {
			return err
		}
	}

	if err := insertInvoiceEventTx(tx, source.ID, "split", "Item split created", fmt.Sprintf("%s moved to %s", formatINR(childTotal), childInvoiceNumber), "operator", createdAt); err != nil {
		return err
	}
	return insertInvoiceEventTx(tx, childID, "split_child", "Created from item split", fmt.Sprintf("Source invoice %s", source.InvoiceNumber), "operator", createdAt)
}

func splitInvoiceByAmountsTx(tx *sql.Tx, input SplitInvoiceInput, createdAt string) error {
	source, err := invoiceSummaryByIDTx(tx, input.InvoiceID)
	if err != nil {
		return err
	}
	if source.Status == invoiceStatusVoided || source.Status == invoiceStatusRefunded || source.Status == invoiceStatusSplit {
		return fmt.Errorf("cannot split %s invoice", source.Status)
	}
	amounts := input.Amounts
	if len(amounts) == 0 {
		half := round2(source.Total / 2)
		amounts = []float64{half, round2(source.Total - half)}
	}
	sum := 0.0
	for _, amount := range amounts {
		if amount <= 0 {
			return errors.New("split amounts must be positive")
		}
		sum += amount
	}
	if math.Abs(round2(sum)-source.Total) > 0.01 {
		return fmt.Errorf("split amounts must equal invoice total %.2f", source.Total)
	}

	splitGroupID := source.SplitGroupID
	if splitGroupID == "" {
		splitGroupID = mustID()
	}
	if _, err := tx.Exec("UPDATE sales SET status = ?, split_group_id = ? WHERE id = ?", invoiceStatusSplit, splitGroupID, source.ID); err != nil {
		return err
	}

	for index, amount := range amounts {
		childID := mustID()
		invoiceNumber, err := nextDocumentNumber(tx, "invoice_prefix", "next_invoice_seq")
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO sales (
				id, invoice_number, customer_name, customer_phone, table_name, order_type,
				subtotal, discount_total, tax_total, total, channel,
				payment_method, payment_tendered, payment_change, status,
				source_invoice_id, split_group_id, kot_sent_at, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, childID, invoiceNumber, source.CustomerName, source.CustomerPhone, source.TableName, source.OrderType,
			amount, 0, 0, round2(amount), source.Channel, source.PaymentMethod, 0, 0, invoiceStatusDraft,
			source.ID, splitGroupID, source.KOTSentAt, createdAt); err != nil {
			return err
		}
		if err := insertInvoiceEventTx(tx, childID, "amount_split_child", "Amount split created", fmt.Sprintf("Part %d from %s", index+1, source.InvoiceNumber), "operator", createdAt); err != nil {
			return err
		}
	}
	return insertInvoiceEventTx(tx, source.ID, "split", "Amount split created", fmt.Sprintf("%d payable invoices created", len(amounts)), "operator", createdAt)
}

func invoiceSummaryByIDTx(tx *sql.Tx, invoiceID string) (InvoiceSummary, error) {
	var invoice InvoiceSummary
	row := tx.QueryRow(`
		SELECT id, invoice_number, customer_name, customer_phone, channel, order_type, table_name,
			subtotal, discount_total, tax_total, total, payment_method, status,
			source_invoice_id, split_group_id, void_reason, refund_reason, approved_by, approved_at,
			created_at, closed_at, kot_sent_at
		FROM sales
		WHERE id = ?
	`, invoiceID)
	if err := scanInvoiceSummary(row, &invoice); err != nil {
		if err == sql.ErrNoRows {
			return invoice, fmt.Errorf("invoice not found: %s", invoiceID)
		}
		return invoice, err
	}
	return invoice, nil
}

type summaryScanner interface {
	Scan(dest ...any) error
}

func scanInvoiceSummary(scanner summaryScanner, invoice *InvoiceSummary) error {
	return scanner.Scan(
		&invoice.ID,
		&invoice.InvoiceNumber,
		&invoice.CustomerName,
		&invoice.CustomerPhone,
		&invoice.Channel,
		&invoice.OrderType,
		&invoice.TableName,
		&invoice.Subtotal,
		&invoice.DiscountTotal,
		&invoice.TaxTotal,
		&invoice.Total,
		&invoice.PaymentMethod,
		&invoice.Status,
		&invoice.SourceInvoiceID,
		&invoice.SplitGroupID,
		&invoice.VoidReason,
		&invoice.RefundReason,
		&invoice.ApprovedBy,
		&invoice.ApprovedAt,
		&invoice.CreatedAt,
		&invoice.ClosedAt,
		&invoice.KOTSentAt,
	)
}

func invoiceLinesTx(tx *sql.Tx, invoiceID string) ([]InvoiceLine, error) {
	lines, err := invoiceLinesForTotalsTx(tx, invoiceID)
	if err != nil {
		return nil, err
	}
	result := make([]InvoiceLine, 0, len(lines))
	for _, line := range lines {
		result = append(result, InvoiceLine{
			ID:        line.id,
			InvoiceID: invoiceID,
			ItemID:    line.itemID,
			ItemName:  line.itemName,
			Quantity:  line.quantity,
			UnitPrice: line.unitPrice,
			LineTotal: line.lineTotal,
			Notes:     line.notes,
			RouteID:   line.routeID,
			RouteName: line.routeName,
		})
	}
	return result, nil
}

func invoiceLinesForTotalsTx(tx *sql.Tx, invoiceID string) ([]existingInvoiceLine, error) {
	rows, err := tx.Query(`
		SELECT sl.id, sl.item_id, sl.item_name, sl.quantity, sl.unit_price, sl.line_total, sl.notes,
			COALESCE(m.route_id, 'route_kitchen'), COALESCE(r.name, 'Kitchen'), COALESCE(m.tax_rate, 0)
		FROM sale_lines sl
		LEFT JOIN menu_items m ON m.id = sl.item_id
		LEFT JOIN kitchen_routes r ON r.id = m.route_id
		WHERE sl.sale_id = ?
		ORDER BY sl.rowid
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]existingInvoiceLine, 0)
	for rows.Next() {
		var line existingInvoiceLine
		if err := rows.Scan(&line.id, &line.itemID, &line.itemName, &line.quantity, &line.unitPrice, &line.lineTotal, &line.notes, &line.routeID, &line.routeName, &line.taxRate); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func invoicePaymentsTx(tx *sql.Tx, invoiceID string) ([]PaymentRecord, error) {
	rows, err := tx.Query(`
		SELECT id, invoice_id, method, amount, tendered, change_due, status, created_at
		FROM payments
		WHERE invoice_id = ?
		ORDER BY created_at
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := make([]PaymentRecord, 0)
	for rows.Next() {
		var payment PaymentRecord
		if err := rows.Scan(&payment.ID, &payment.InvoiceID, &payment.Method, &payment.Amount, &payment.Tendered, &payment.ChangeDue, &payment.Status, &payment.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, rows.Err()
}

func invoiceRefundsTx(tx *sql.Tx, invoiceID string) ([]RefundRecord, error) {
	rows, err := tx.Query(`
		SELECT id, invoice_id, amount, reason, approved_by, approved_at, created_at
		FROM refunds
		WHERE invoice_id = ?
		ORDER BY created_at
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refunds := make([]RefundRecord, 0)
	for rows.Next() {
		var refund RefundRecord
		if err := rows.Scan(&refund.ID, &refund.InvoiceID, &refund.Amount, &refund.Reason, &refund.ApprovedBy, &refund.ApprovedAt, &refund.CreatedAt); err != nil {
			return nil, err
		}
		refunds = append(refunds, refund)
	}
	return refunds, rows.Err()
}

func invoiceEventsTx(tx *sql.Tx, invoiceID string) ([]InvoiceEvent, error) {
	rows, err := tx.Query(`
		SELECT id, invoice_id, type, title, detail, actor, created_at
		FROM invoice_events
		WHERE invoice_id = ?
		ORDER BY created_at
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]InvoiceEvent, 0)
	for rows.Next() {
		var event InvoiceEvent
		if err := rows.Scan(&event.ID, &event.InvoiceID, &event.Type, &event.Title, &event.Detail, &event.Actor, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func invoiceNotificationsTx(tx *sql.Tx, invoiceID string) ([]NotificationRecord, error) {
	rows, err := tx.Query(`
		SELECT id, invoice_id, channel, recipient, template, payload, status, error, created_at, sent_at, read_at
		FROM notification_queue
		WHERE invoice_id = ?
		ORDER BY created_at
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]NotificationRecord, 0)
	for rows.Next() {
		var notification NotificationRecord
		if err := rows.Scan(&notification.ID, &notification.InvoiceID, &notification.Channel, &notification.Recipient, &notification.Template, &notification.Payload, &notification.Status, &notification.Error, &notification.CreatedAt, &notification.SentAt, &notification.ReadAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

func invoiceKitchenTicketsTx(tx *sql.Tx, invoiceID string) ([]KitchenTicket, error) {
	rows, err := tx.Query(`
		SELECT id, sale_id, invoice_number, ticket_number, route_id, route_name, status, created_at
		FROM kitchen_tickets
		WHERE sale_id = ?
		ORDER BY created_at
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]KitchenTicket, 0)
	for rows.Next() {
		var ticket KitchenTicket
		if err := rows.Scan(&ticket.ID, &ticket.SaleID, &ticket.InvoiceNumber, &ticket.TicketNumber, &ticket.RouteID, &ticket.RouteName, &ticket.Status, &ticket.CreatedAt); err != nil {
			return nil, err
		}
		lines, err := kitchenTicketLinesTx(tx, ticket.ID)
		if err != nil {
			return nil, err
		}
		ticket.Lines = lines
		tickets = append(tickets, ticket)
	}
	return tickets, rows.Err()
}

func kitchenTicketLinesTx(tx *sql.Tx, ticketID string) ([]KitchenTicketLine, error) {
	rows, err := tx.Query(`
		SELECT id, ticket_id, item_id, item_name, quantity, notes, status
		FROM kitchen_ticket_lines
		WHERE ticket_id = ?
		ORDER BY item_name
	`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]KitchenTicketLine, 0)
	for rows.Next() {
		var line KitchenTicketLine
		if err := rows.Scan(&line.ID, &line.TicketID, &line.ItemID, &line.ItemName, &line.Quantity, &line.Notes, &line.Status); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func insertInvoiceEventTx(tx *sql.Tx, invoiceID string, eventType string, title string, detail string, actor string, createdAt string) error {
	_, err := tx.Exec(`
		INSERT INTO invoice_events (id, invoice_id, type, title, detail, actor, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, mustID(), invoiceID, eventType, title, detail, actor, createdAt)
	return err
}

func queueInvoiceNotificationTx(tx *sql.Tx, invoiceID string, template string, title string, createdAt string) error {
	summary, err := invoiceSummaryByIDTx(tx, invoiceID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{
		"invoiceNumber": summary.InvoiceNumber,
		"customerName":  summary.CustomerName,
		"total":         summary.Total,
		"title":         title,
	})
	if err != nil {
		return err
	}
	status := "ready"
	recipient := summary.CustomerPhone
	if recipient == "" {
		status = "skipped"
	}
	_, err = tx.Exec(`
		INSERT INTO notification_queue (id, invoice_id, channel, recipient, template, payload, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, mustID(), invoiceID, "whatsapp", recipient, template, string(payload), status, createdAt)
	return err
}

func approveManagerTx(tx *sql.Tx, action string, invoiceID string, pin string, createdAt string) (ApprovalToken, error) {
	trimmedPIN := strings.TrimSpace(pin)
	var storedHash, salt string
	_ = tx.QueryRow("SELECT value FROM app_settings WHERE key = 'manager_pin_hash'").Scan(&storedHash)
	_ = tx.QueryRow("SELECT value FROM app_settings WHERE key = 'manager_pin_salt'").Scan(&salt)
	if storedHash != "" && salt != "" {
		if hashPIN(trimmedPIN, salt) != storedHash {
			return ApprovalToken{}, errors.New("manager PIN is invalid")
		}
	} else {
		var expected string
		if err := tx.QueryRow("SELECT value FROM app_settings WHERE key = 'manager_pin'").Scan(&expected); err != nil {
			return ApprovalToken{}, err
		}
		if trimmedPIN != expected {
			return ApprovalToken{}, errors.New("manager PIN is invalid")
		}
	}
	expiresAt := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
	token := mustID()
	approval := ApprovalToken{
		Token:      token,
		ApprovedBy: "Manager",
		ApprovedAt: createdAt,
		ExpiresAt:  expiresAt,
	}
	_, err := tx.Exec(`
		INSERT INTO manager_approvals (id, action, invoice_id, approved_by, approved_at, token, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, mustID(), action, invoiceID, approval.ApprovedBy, approval.ApprovedAt, approval.Token, approval.ExpiresAt)
	if err != nil {
		return ApprovalToken{}, err
	}
	_, err = tx.Exec(`
		INSERT INTO audit_log (id, event_type, target_type, target_id, detail, actor, created_at)
		VALUES (?, ?, 'invoice', ?, ?, ?, ?)
	`, mustID(), "manager_approval", invoiceID, action, approval.ApprovedBy, createdAt)
	return approval, err
}

func sumLineSubtotal(lines []existingInvoiceLine) float64 {
	total := 0.0
	for _, line := range lines {
		total += line.lineTotal
	}
	return round2(total)
}

func splitDiscount(originalDiscount float64, originalSubtotal float64, sourceSubtotal float64, childSubtotal float64) (float64, float64) {
	if originalDiscount <= 0 || originalSubtotal <= 0 {
		return 0, 0
	}
	childDiscount := round2(originalDiscount * (childSubtotal / originalSubtotal))
	sourceDiscount := round2(originalDiscount - childDiscount)
	if sourceDiscount > sourceSubtotal {
		sourceDiscount = sourceSubtotal
	}
	if childDiscount > childSubtotal {
		childDiscount = childSubtotal
	}
	return sourceDiscount, childDiscount
}

func totalsForExistingLines(lines []existingInvoiceLine, discount float64) (float64, float64, float64) {
	subtotal := sumLineSubtotal(lines)
	discount = round2(math.Min(discount, subtotal))
	tax := 0.0
	for _, line := range lines {
		share := 0.0
		if subtotal > 0 {
			share = line.lineTotal / subtotal
		}
		tax += (line.lineTotal - (discount * share)) * line.taxRate
	}
	tax = round2(tax)
	total := round2(subtotal - discount + tax)
	return subtotal, tax, total
}

func formatINR(value float64) string {
	return fmt.Sprintf("₹%.2f", value)
}
