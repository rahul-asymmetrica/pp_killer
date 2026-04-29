package nexus

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func (s *Store) SaveMenuItem(input MenuItemInput) (Dashboard, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Dashboard{}, errors.New("item name is required")
	}
	if input.Price < 0 || input.Cost < 0 {
		return Dashboard{}, errors.New("price and cost cannot be negative")
	}
	if input.TaxRate < 0 {
		input.TaxRate = 0
	}
	category := firstNonEmpty(input.Category, "Food")
	routeID := firstNonEmpty(input.RouteID, defaultRouteForCategory(category))
	status := firstNonEmpty(input.Status, "active")
	id := strings.TrimSpace(input.ID)
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)
	if id == "" {
		_ = tx.QueryRow("SELECT id FROM menu_items WHERE lower(name) = lower(?)", input.Name).Scan(&id)
	}
	if id == "" {
		id = uniqueMenuItemIDTx(tx, input.Name)
	}
	if _, err := tx.Exec(`
		INSERT INTO menu_items (id, name, category, price, cost, status, route_id, tax_rate)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			category = excluded.category,
			price = excluded.price,
			cost = excluded.cost,
			status = excluded.status,
			route_id = excluded.route_id,
			tax_rate = excluded.tax_rate
	`, id, strings.TrimSpace(input.Name), category, round2(input.Price), round2(input.Cost), status, routeID, input.TaxRate); err != nil {
		return Dashboard{}, err
	}
	if err := ensureMenuCategoryTx(tx, category); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "menu_item", id, "upsert", input, nowUTC()); err != nil {
		return Dashboard{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) SaveIngredient(input IngredientInput) (Dashboard, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Dashboard{}, errors.New("ingredient name is required")
	}
	if strings.TrimSpace(input.Unit) == "" {
		return Dashboard{}, errors.New("usage unit is required")
	}
	if strings.TrimSpace(input.PurchaseUnit) == "" {
		input.PurchaseUnit = strings.TrimSpace(input.Unit)
	}
	if input.PurchaseToUsage <= 0 {
		input.PurchaseToUsage = 1
	}
	if input.OnHandQty < 0 || input.ReorderPoint < 0 || input.LastPurchaseCost < 0 {
		return Dashboard{}, errors.New("ingredient quantities and costs cannot be negative")
	}
	if input.WasteFactor < 0 {
		input.WasteFactor = 0
	}
	id := strings.TrimSpace(input.ID)
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)
	if id == "" {
		_ = tx.QueryRow("SELECT id FROM ingredients WHERE lower(name) = lower(?)", input.Name).Scan(&id)
	}
	if id == "" {
		id = uniqueIngredientIDTx(tx, input.Name)
	}
	if _, err := tx.Exec(`
		INSERT INTO ingredients (
			id, name, unit, purchase_unit, purchase_to_usage, on_hand_qty,
			reorder_point, waste_factor, last_purchase_cost
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			unit = excluded.unit,
			purchase_unit = excluded.purchase_unit,
			purchase_to_usage = excluded.purchase_to_usage,
			on_hand_qty = excluded.on_hand_qty,
			reorder_point = excluded.reorder_point,
			waste_factor = excluded.waste_factor,
			last_purchase_cost = excluded.last_purchase_cost
	`, id, strings.TrimSpace(input.Name), strings.TrimSpace(input.Unit), strings.TrimSpace(input.PurchaseUnit),
		input.PurchaseToUsage, input.OnHandQty, input.ReorderPoint, input.WasteFactor, input.LastPurchaseCost); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "ingredient", id, "upsert", input, nowUTC()); err != nil {
		return Dashboard{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) ImportMenuCSV(input MenuImportInput) (MenuImportResult, error) {
	reader := csv.NewReader(strings.NewReader(input.Text))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return MenuImportResult{}, err
	}
	result := MenuImportResult{Skipped: []string{}}
	tx, err := s.db.Begin()
	if err != nil {
		return result, err
	}
	defer rollback(tx)
	for index, row := range rows {
		if len(row) == 0 || strings.TrimSpace(strings.Join(row, "")) == "" {
			continue
		}
		if index == 0 && looksLikeMenuHeader(row) {
			continue
		}
		if len(row) < 3 {
			result.Skipped = append(result.Skipped, fmt.Sprintf("row %d: need name, category, price", index+1))
			continue
		}
		name := strings.TrimSpace(row[0])
		category := firstNonEmpty(cell(row, 1), "Food")
		price, err := parseMoney(cell(row, 2))
		if name == "" || err != nil {
			result.Skipped = append(result.Skipped, fmt.Sprintf("row %d: invalid item or price", index+1))
			continue
		}
		routeID := routeIDFromLabel(cell(row, 3), category)
		taxRate := 0.05
		if strings.TrimSpace(cell(row, 4)) != "" {
			if parsed, err := parseRate(cell(row, 4)); err == nil {
				taxRate = parsed
			}
		}
		cost := 0.0
		if strings.TrimSpace(cell(row, 5)) != "" {
			cost, _ = parseMoney(cell(row, 5))
		}
		status := firstNonEmpty(cell(row, 6), "active")
		id := ""
		_ = tx.QueryRow("SELECT id FROM menu_items WHERE lower(name) = lower(?)", name).Scan(&id)
		operation := "update"
		if id == "" {
			id = uniqueMenuItemIDTx(tx, name)
			operation = "create"
			result.Imported++
		} else {
			result.Updated++
		}
		if _, err := tx.Exec(`
			INSERT INTO menu_items (id, name, category, price, cost, status, route_id, tax_rate)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				category = excluded.category,
				price = excluded.price,
				cost = excluded.cost,
				status = excluded.status,
				route_id = excluded.route_id,
				tax_rate = excluded.tax_rate
		`, id, name, category, round2(price), round2(cost), status, routeID, taxRate); err != nil {
			return result, err
		}
		if err := ensureMenuCategoryTx(tx, category); err != nil {
			return result, err
		}
		if err := insertSyncLog(tx, "menu_item", id, operation, map[string]any{"name": name, "price": price}, nowUTC()); err != nil {
			return result, err
		}
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Store) SavePrinterConnection(input PrinterConnectionInput) (PilotWorkspace, error) {
	target := strings.TrimSpace(input.Target)
	if target == "" {
		return PilotWorkspace{}, errors.New("printer target is required")
	}
	mode := firstNonEmpty(input.Mode, "local")
	health := "ready_for_test"
	if !printerTargetLooksValid(target) {
		health = "needs_check"
	}
	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	if _, err := tx.Exec(`
		INSERT INTO integration_settings (id, provider, mode, display_name, base_url, credential_status, health_status, last_checked_at, updated_at)
		VALUES ('int_printer', 'escpos_printer', ?, ?, ?, 'not_required', ?, ?, ?)
		ON CONFLICT(provider) DO UPDATE SET
			mode = excluded.mode,
			display_name = excluded.display_name,
			base_url = excluded.base_url,
			credential_status = excluded.credential_status,
			health_status = excluded.health_status,
			last_checked_at = excluded.last_checked_at,
			updated_at = excluded.updated_at
	`, mode, firstNonEmpty(input.DisplayName, "ESC/POS Printer"), target, health, now, now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := insertAuditLogTx(tx, "printer_saved", "integration_settings", "escpos_printer", fmt.Sprintf("Printer target set to %s", target), now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) ExportInvoicePDF(invoiceID string) (ExportResult, error) {
	detail, err := s.GetInvoiceDetail(invoiceID)
	if err != nil {
		return ExportResult{}, err
	}
	dir, err := exportDir()
	if err != nil {
		return ExportResult{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ExportResult{}, err
	}
	fileName := safeFileName(firstNonEmpty(detail.Summary.InvoiceNumber, "invoice")) + ".pdf"
	path := filepath.Join(dir, fileName)
	lines := invoicePDFLines(detail)
	if err := writeSimplePDF(path, firstNonEmpty(detail.Summary.InvoiceNumber, "Invoice"), lines); err != nil {
		return ExportResult{}, err
	}
	return exportResult("invoice_pdf", path, "application/pdf")
}

func (s *Store) ExportAccountingCSV() (ExportResult, error) {
	snapshot, err := s.GetAccountingSnapshot()
	if err != nil {
		return ExportResult{}, err
	}
	dir, err := exportDir()
	if err != nil {
		return ExportResult{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ExportResult{}, err
	}
	path := filepath.Join(dir, "nexus_accounting_"+time.Now().UTC().Format("20060102_150405")+".csv")
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{"NEXUS Accounting Summary"})
	_ = writer.Write([]string{"Sales", fmt.Sprintf("%.2f", snapshot.SalesTotal)})
	_ = writer.Write([]string{"GST Payable", fmt.Sprintf("%.2f", snapshot.TaxPayable)})
	_ = writer.Write([]string{"Cash And Bank", fmt.Sprintf("%.2f", snapshot.CashAndBank)})
	_ = writer.Write([]string{"Razorpay Clearing", fmt.Sprintf("%.2f", snapshot.RazorpayClearing)})
	_ = writer.Write([]string{"Vendor Payables", fmt.Sprintf("%.2f", snapshot.VendorPayables)})
	_ = writer.Write([]string{})
	_ = writer.Write([]string{"Ledger", "Group", "Debit", "Credit", "Balance", "Side"})
	for _, row := range snapshot.TrialBalance {
		_ = writer.Write([]string{row.LedgerName, row.GroupName, fmt.Sprintf("%.2f", row.DebitTotal), fmt.Sprintf("%.2f", row.CreditTotal), fmt.Sprintf("%.2f", row.Balance), row.BalanceSide})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return ExportResult{}, err
	}
	if err := os.WriteFile(path, buffer.Bytes(), 0o644); err != nil {
		return ExportResult{}, err
	}
	return exportResult("accounting_csv", path, "text/csv")
}

func (s *Store) ExportInvoicesCSV(filter InvoiceFilter) (ExportResult, error) {
	invoices, err := s.GetInvoices(filter)
	if err != nil {
		return ExportResult{}, err
	}
	dir, err := exportDir()
	if err != nil {
		return ExportResult{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ExportResult{}, err
	}
	path := filepath.Join(dir, "nexus_invoices_"+time.Now().UTC().Format("20060102_150405")+".csv")
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	_ = writer.Write([]string{"Invoice", "Status", "Customer", "Phone", "Table", "Payment", "Subtotal", "Discount", "GST", "Total", "Created", "Closed"})
	for _, invoice := range invoices {
		_ = writer.Write([]string{
			invoice.InvoiceNumber,
			invoice.Status,
			invoice.CustomerName,
			invoice.CustomerPhone,
			invoice.TableName,
			invoice.PaymentMethod,
			fmt.Sprintf("%.2f", invoice.Subtotal),
			fmt.Sprintf("%.2f", invoice.DiscountTotal),
			fmt.Sprintf("%.2f", invoice.TaxTotal),
			fmt.Sprintf("%.2f", invoice.Total),
			invoice.CreatedAt,
			invoice.ClosedAt,
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return ExportResult{}, err
	}
	if err := os.WriteFile(path, buffer.Bytes(), 0o644); err != nil {
		return ExportResult{}, err
	}
	return exportResult("invoices_csv", path, "text/csv")
}

func (s *Store) ExportDayClosePDF(businessDate string) (ExportResult, error) {
	if strings.TrimSpace(businessDate) == "" {
		businessDate = todayDate()
	}
	summary, err := s.dayClosePreview(businessDate)
	if err != nil {
		return ExportResult{}, err
	}
	dir, err := exportDir()
	if err != nil {
		return ExportResult{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ExportResult{}, err
	}
	fileName := safeFileName("nexus_day_close_"+summary.BusinessDate) + ".pdf"
	path := filepath.Join(dir, fileName)
	if err := writeSimplePDF(path, "NEXUS Day Close", dayClosePDFLines(summary)); err != nil {
		return ExportResult{}, err
	}
	return exportResult("day_close_pdf", path, "application/pdf")
}

func ensureMenuCategoryTx(tx *sql.Tx, category string) error {
	category = strings.TrimSpace(category)
	if category == "" {
		return nil
	}
	_, err := tx.Exec(`
		INSERT INTO menu_categories (id, name, sort_order, status)
		VALUES (?, ?, 100, 'active')
		ON CONFLICT(id) DO NOTHING
	`, "cat_"+slug(category), category)
	return err
}

func uniqueMenuItemIDTx(tx *sql.Tx, name string) string {
	base := "item_" + slug(name)
	if base == "item_" {
		return mustID()
	}
	id := base
	for suffix := 2; suffix < 1000; suffix++ {
		var existing string
		err := tx.QueryRow("SELECT id FROM menu_items WHERE id = ?", id).Scan(&existing)
		if err == sql.ErrNoRows {
			return id
		}
		if err != nil {
			return mustID()
		}
		id = fmt.Sprintf("%s_%d", base, suffix)
	}
	return mustID()
}

func uniqueIngredientIDTx(tx *sql.Tx, name string) string {
	base := "ing_" + slug(name)
	if base == "ing_" {
		return mustID()
	}
	id := base
	for suffix := 2; suffix < 1000; suffix++ {
		var existing string
		err := tx.QueryRow("SELECT id FROM ingredients WHERE id = ?", id).Scan(&existing)
		if err == sql.ErrNoRows {
			return id
		}
		if err != nil {
			return mustID()
		}
		id = fmt.Sprintf("%s_%d", base, suffix)
	}
	return mustID()
}

func looksLikeMenuHeader(row []string) bool {
	if len(row) == 0 {
		return false
	}
	head := strings.ToLower(strings.TrimSpace(row[0]))
	return head == "name" || head == "item" || head == "item name"
}

func cell(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func parseMoney(value string) (float64, error) {
	var builder strings.Builder
	for _, r := range strings.TrimSpace(value) {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			builder.WriteRune(r)
		}
	}
	return strconv.ParseFloat(builder.String(), 64)
}

func parseRate(value string) (float64, error) {
	cleaned := strings.TrimSpace(strings.TrimSuffix(value, "%"))
	parsed, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, err
	}
	if strings.Contains(value, "%") || parsed > 1 {
		parsed = parsed / 100
	}
	return parsed, nil
}

func defaultRouteForCategory(category string) string {
	category = strings.ToLower(category)
	if strings.Contains(category, "coffee") || strings.Contains(category, "bar") || strings.Contains(category, "beverage") || strings.Contains(category, "drink") {
		return "route_bar"
	}
	return "route_kitchen"
}

func routeIDFromLabel(label, category string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	switch {
	case label == "route_bar" || strings.Contains(label, "bar") || strings.Contains(label, "coffee"):
		return "route_bar"
	case label == "route_kitchen" || strings.Contains(label, "kitchen") || strings.Contains(label, "food"):
		return "route_kitchen"
	default:
		return defaultRouteForCategory(category)
	}
}

func printerTargetLooksValid(target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	return strings.HasPrefix(target, "tcp://") ||
		strings.HasPrefix(target, "usb://") ||
		strings.HasPrefix(target, "bluetooth://") ||
		strings.HasPrefix(target, "system://")
}

func invoicePDFLines(detail InvoiceDetail) []string {
	lines := []string{
		"NEXUS Demo Cafe",
		"Invoice: " + detail.Summary.InvoiceNumber,
		"Customer: " + firstNonEmpty(detail.Summary.CustomerName, "Walk-in Guest"),
		"Table: " + firstNonEmpty(detail.Summary.TableName, detail.Summary.OrderType),
		"Payment: " + strings.ToUpper(detail.Summary.PaymentMethod),
		"",
		"Items",
	}
	for _, line := range detail.Lines {
		lines = append(lines, fmt.Sprintf("%d x %s  INR %.2f", line.Quantity, line.ItemName, line.LineTotal))
	}
	lines = append(lines,
		"",
		fmt.Sprintf("Subtotal: INR %.2f", detail.Summary.Subtotal),
		fmt.Sprintf("Discount: INR %.2f", detail.Summary.DiscountTotal),
		fmt.Sprintf("GST: INR %.2f", detail.Summary.TaxTotal),
		fmt.Sprintf("Total: INR %.2f", detail.Summary.Total),
	)
	return lines
}

func dayClosePDFLines(summary DayCloseSummary) []string {
	return []string{
		"NEXUS Demo Cafe",
		"Business Date: " + summary.BusinessDate,
		"Status: " + strings.ToUpper(summary.Status),
		"",
		fmt.Sprintf("Sales: INR %.2f", summary.SalesTotal),
		fmt.Sprintf("Cash Expected: INR %.2f", summary.CashExpected),
		fmt.Sprintf("Cash Counted: INR %.2f", summary.CashCounted),
		fmt.Sprintf("Cash Variance: INR %.2f", summary.CashVariance),
		fmt.Sprintf("UPI: INR %.2f", summary.UPITotal),
		fmt.Sprintf("Card: INR %.2f", summary.CardTotal),
		fmt.Sprintf("Razorpay: INR %.2f", summary.RazorpayTotal),
		fmt.Sprintf("Refunds: INR %.2f", summary.RefundTotal),
		fmt.Sprintf("Discounts: INR %.2f", summary.DiscountTotal),
		fmt.Sprintf("Voids: %d", summary.VoidCount),
		"",
		"Notes: " + firstNonEmpty(summary.Notes, "No notes"),
	}
}

func exportDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("NEXUS_EXPORT_DIR")); override != "" {
		return override, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads", "NEXUS_Exports"), nil
}

func exportResult(kind, path, mimeType string) (ExportResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{Kind: kind, Path: path, FileName: filepath.Base(path), MIMEType: mimeType, Bytes: info.Size()}, nil
}

func safeFileName(value string) string {
	name := slug(value)
	if name == "" {
		return "nexus_export"
	}
	return name
}

func writeSimplePDF(path, title string, lines []string) error {
	content := buildPDFContent(title, lines)
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
	}
	var buffer bytes.Buffer
	buffer.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		offsets[index+1] = buffer.Len()
		buffer.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", index+1, object))
	}
	xref := buffer.Len()
	buffer.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	buffer.WriteString("0000000000 65535 f \n")
	for index := 1; index <= len(objects); index++ {
		buffer.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[index]))
	}
	buffer.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref))
	return os.WriteFile(path, buffer.Bytes(), 0o644)
}

func buildPDFContent(title string, lines []string) string {
	var builder strings.Builder
	builder.WriteString("BT\n")
	builder.WriteString("/F1 18 Tf\n50 800 Td\n")
	builder.WriteString("(" + pdfEscape(title) + ") Tj\n")
	builder.WriteString("/F1 11 Tf\n0 -28 Td\n")
	for _, line := range lines {
		builder.WriteString("(" + pdfEscape(line) + ") Tj\n0 -16 Td\n")
	}
	builder.WriteString("ET")
	return builder.String()
}

func pdfEscape(value string) string {
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		if r > unicode.MaxASCII {
			return '?'
		}
		return r
	}, value)
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	return value
}

func slug(value string) string {
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteRune('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}
