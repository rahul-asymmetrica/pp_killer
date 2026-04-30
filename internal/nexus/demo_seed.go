package nexus

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"time"
)

const demoSeedRand = 20260430

type demoMenuItem struct {
	id       string
	name     string
	category string
	price    float64
	cost     float64
	routeID  string
	taxRate  float64
	weight   int
}

type demoSaleLine struct {
	item      demoMenuItem
	quantity  int
	unitPrice float64
	lineTotal float64
	notes     string
}

type demoCustomerProfile struct {
	id    string
	name  string
	phone string
}

type demoCustomerStats struct {
	id        string
	name      string
	phone     string
	total     float64
	visits    int
	lastVisit string
	favorites map[string]int
}

type demoDayStats struct {
	salesTotal    float64
	orderCount    int
	refundTotal   float64
	voidCount     int
	discountTotal float64
	cashTotal     float64
	upiTotal      float64
	cardTotal     float64
	razorpayTotal float64
}

type demoSeedState struct {
	result        DemoSeedResult
	invoiceIDs    []string
	invoiceSeq    int
	kotSeq        int
	purchaseSeq   int
	customerStats map[string]*demoCustomerStats
}

func (s *Store) SeedDemoOperations(months int) (DemoSeedResult, error) {
	if months <= 0 {
		months = 6
	}
	if months > 24 {
		months = 24
	}

	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	end := today.AddDate(0, 0, -1)
	start := today.AddDate(0, -months, 0)
	if start.After(end) {
		start = end
	}

	state := &demoSeedState{
		result: DemoSeedResult{
			Months:    months,
			StartedAt: start.Format("2006-01-02"),
			EndedAt:   end.Format("2006-01-02"),
		},
		invoiceSeq:    1,
		kotSeq:        1,
		purchaseSeq:   1,
		customerStats: map[string]*demoCustomerStats{},
	}

	tx, err := s.db.Begin()
	if err != nil {
		return DemoSeedResult{}, err
	}
	defer rollback(tx)

	if err := clearDemoSeedTx(tx); err != nil {
		return DemoSeedResult{}, err
	}
	now := nowUTC()
	if err := seedDemoReferenceTx(tx, now); err != nil {
		return DemoSeedResult{}, err
	}
	menu := demoMenuCatalog()
	customers := demoCustomerProfiles(280)
	rng := rand.New(rand.NewSource(demoSeedRand))

	for date, dayIndex := start, 0; !date.After(end); date, dayIndex = date.AddDate(0, 0, 1), dayIndex+1 {
		stats := &demoDayStats{}
		orderCount := demoOrderCount(date, dayIndex, rng)
		for orderIndex := 0; orderIndex < orderCount; orderIndex++ {
			if err := seedDemoSaleTx(tx, date, orderIndex, menu, customers, rng, state, stats); err != nil {
				return DemoSeedResult{}, err
			}
		}
		if date.Weekday() == time.Monday || date.Weekday() == time.Thursday {
			if err := seedDemoPurchaseOrderTx(tx, date, rng, state); err != nil {
				return DemoSeedResult{}, err
			}
		}
		if err := seedDemoDayCloseTx(tx, date, stats, rng); err != nil {
			return DemoSeedResult{}, err
		}
		state.result.DayCloses++
		state.result.BusinessDays++
	}

	if err := seedDemoCustomersTx(tx, state.customerStats); err != nil {
		return DemoSeedResult{}, err
	}
	if err := seedDemoLiveSessionsTx(tx, today, menu, rng, state); err != nil {
		return DemoSeedResult{}, err
	}
	if err := seedDemoOperationalNoiseTx(tx, today, rng, state); err != nil {
		return DemoSeedResult{}, err
	}
	if err := seedDemoSequencesTx(tx, state); err != nil {
		return DemoSeedResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return DemoSeedResult{}, err
	}

	if _, err := s.RebuildAccountingShell(); err != nil {
		return DemoSeedResult{}, err
	}
	if err := s.refreshAnalyticsSnapshots(); err != nil {
		return DemoSeedResult{}, err
	}

	state.result.Tables = 50
	state.result.Waiters = 20
	state.result.Customers = len(state.customerStats)
	state.result.SalesTotal = round2(state.result.SalesTotal)
	return state.result, nil
}

func clearDemoSeedTx(tx *sql.Tx) error {
	statements := []string{
		"DELETE FROM order_session_line_modifiers",
		"DELETE FROM order_session_events",
		"DELETE FROM order_session_lines",
		"DELETE FROM order_sessions",
		"DELETE FROM notification_queue",
		"DELETE FROM payment_requests",
		"DELETE FROM print_jobs",
		"DELETE FROM manager_approvals",
		"DELETE FROM invoice_events",
		"DELETE FROM refunds",
		"DELETE FROM payments",
		"DELETE FROM kitchen_ticket_lines",
		"DELETE FROM kitchen_tickets",
		"DELETE FROM sale_lines",
		"DELETE FROM sales",
		"DELETE FROM accounting_voucher_lines",
		"DELETE FROM accounting_vouchers",
		"DELETE FROM day_closes",
		"DELETE FROM analytics_daily_snapshots",
		"DELETE FROM analytics_item_snapshots",
		"DELETE FROM analytics_hourly_snapshots",
		"DELETE FROM purchase_order_lines",
		"DELETE FROM vendor_debit_notes",
		"DELETE FROM purchase_orders",
		"DELETE FROM delivery_lines",
		"DELETE FROM delivery_batches",
		"DELETE FROM inventory_movements",
		"DELETE FROM sync_log",
		"DELETE FROM customers",
		"DELETE FROM dining_tables",
		"DELETE FROM floor_sections",
		"DELETE FROM staff WHERE role IN ('waiter', 'runner')",
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoReferenceTx(tx *sql.Tx, now string) error {
	if _, err := tx.Exec(`
		UPDATE restaurants
		SET name = 'NEXUS Demo Bistro', phone = '+91 80412 88990', website = 'demo.nexus.local',
			brand_voice = 'Polished, operationally sharp, premium casual dining'
		WHERE id = 'rest_demo'
	`); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO restaurant_settings (
			id, restaurant_name, gstin, legal_name, address, state, tax_mode, default_tax_rate,
			service_charge_rate, invoice_prefix, business_hours, backup_path, receipt_printer_id, updated_at
		)
		VALUES ('primary', 'NEXUS Demo Bistro', '29ABCDE1234F1Z5', 'NEXUS Demo Hospitality LLP',
			'12 Brigade Road, Bengaluru', 'Karnataka', 'gst', 0.05, 0.05, 'NX',
			'08:00-23:30', '', 'printer_receipt', ?)
		ON CONFLICT(id) DO UPDATE SET
			restaurant_name = excluded.restaurant_name,
			business_hours = excluded.business_hours,
			updated_at = excluded.updated_at
	`, now); err != nil {
		return err
	}
	if err := seedDemoStaffTx(tx, now); err != nil {
		return err
	}
	if err := seedDemoFloorTx(tx); err != nil {
		return err
	}
	if err := seedDemoInventoryTx(tx, now); err != nil {
		return err
	}
	if err := seedDemoMenuTx(tx); err != nil {
		return err
	}
	return seedDemoVendorsTx(tx)
}

func seedDemoStaffTx(tx *sql.Tx, now string) error {
	names := []string{
		"Aarav Floor", "Vivaan Floor", "Aditya Floor", "Vihaan Floor", "Arjun Floor",
		"Sai Floor", "Reyansh Floor", "Ayaan Floor", "Krishna Floor", "Ishaan Floor",
		"Anaya Floor", "Diya Floor", "Myra Floor", "Ira Floor", "Anika Floor",
		"Kiara Floor", "Sara Floor", "Tara Floor", "Meera Floor", "Rhea Floor",
	}
	for index, name := range names {
		id := demoWaiterID(index)
		if _, err := tx.Exec(`
			INSERT INTO staff (id, name, role, pin_hash, status, created_at)
			VALUES (?, ?, 'waiter', ?, 'active', ?)
		`, id, name, hashPIN("1234", id), now); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoFloorTx(tx *sql.Tx) error {
	sections := []struct {
		id     string
		name   string
		sort   int
		count  int
		prefix string
	}{
		{"sec_main", "Main Dining", 10, 20, "M"},
		{"sec_patio", "Patio", 20, 10, "P"},
		{"sec_family", "Family Lounge", 30, 8, "F"},
		{"sec_private", "Private Dining", 40, 6, "D"},
		{"sec_rooftop", "Rooftop", 50, 6, "R"},
	}
	tableNumber := 1
	for _, section := range sections {
		if _, err := tx.Exec("INSERT INTO floor_sections (id, name, sort_order) VALUES (?, ?, ?)", section.id, section.name, section.sort); err != nil {
			return err
		}
		for index := 1; index <= section.count; index++ {
			seats := 2
			switch {
			case section.prefix == "D":
				seats = 8
			case index%5 == 0:
				seats = 6
			case index%3 == 0:
				seats = 4
			}
			id := fmt.Sprintf("tbl_%02d", tableNumber)
			if section.prefix == "P" && index == 1 {
				id = "tbl_p1"
			}
			name := fmt.Sprintf("%s-%02d", section.prefix, index)
			if _, err := tx.Exec(`
				INSERT INTO dining_tables (id, section_id, name, seats, status, active_session_id)
				VALUES (?, ?, ?, ?, 'available', '')
			`, id, section.id, name, seats); err != nil {
				return err
			}
			tableNumber++
		}
	}
	return nil
}

func seedDemoInventoryTx(tx *sql.Tx, now string) error {
	ingredients := []Ingredient{
		{ID: "ing_beans", Name: "Arabica Beans", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 18500, ReorderPoint: 8000, WasteFactor: 0.04, LastPurchaseCost: 840, LastAuditAt: now},
		{ID: "ing_milk", Name: "Full Cream Milk", Unit: "ml", PurchaseUnit: "ltr", PurchaseToUsage: 1000, OnHandQty: 62000, ReorderPoint: 24000, WasteFactor: 0.06, LastPurchaseCost: 64, LastAuditAt: now},
		{ID: "ing_cup", Name: "12oz Paper Cup", Unit: "pcs", PurchaseUnit: "sleeve", PurchaseToUsage: 50, OnHandQty: 730, ReorderPoint: 240, WasteFactor: 0.02, LastPurchaseCost: 205, LastAuditAt: now},
		{ID: "ing_tomato", Name: "Tomato", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 9800, ReorderPoint: 9000, WasteFactor: 0.11, LastPurchaseCost: 44, LastAuditAt: now},
		{ID: "ing_chicken", Name: "Chicken Breast", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 21000, ReorderPoint: 13000, WasteFactor: 0.08, LastPurchaseCost: 292, LastAuditAt: now},
		{ID: "ing_bun", Name: "Brioche Bun", Unit: "pcs", PurchaseUnit: "tray", PurchaseToUsage: 24, OnHandQty: 96, ReorderPoint: 120, WasteFactor: 0.03, LastPurchaseCost: 380, LastAuditAt: now},
		{ID: "ing_paneer", Name: "Paneer", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 15500, ReorderPoint: 6000, WasteFactor: 0.04, LastPurchaseCost: 330, LastAuditAt: now},
		{ID: "ing_basmati", Name: "Basmati Rice", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 44000, ReorderPoint: 20000, WasteFactor: 0.01, LastPurchaseCost: 118, LastAuditAt: now},
		{ID: "ing_pasta", Name: "Penne Pasta", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 17000, ReorderPoint: 7500, WasteFactor: 0.02, LastPurchaseCost: 145, LastAuditAt: now},
		{ID: "ing_chocolate", Name: "Dark Chocolate", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 5200, ReorderPoint: 6500, WasteFactor: 0.03, LastPurchaseCost: 640, LastAuditAt: now},
		{ID: "ing_flour", Name: "Flour", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 26000, ReorderPoint: 15000, WasteFactor: 0.02, LastPurchaseCost: 58, LastAuditAt: now},
		{ID: "ing_potato", Name: "Potato", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 18500, ReorderPoint: 10000, WasteFactor: 0.08, LastPurchaseCost: 34, LastAuditAt: now},
		{ID: "ing_salmon", Name: "Salmon Fillet", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 3600, ReorderPoint: 5000, WasteFactor: 0.05, LastPurchaseCost: 1150, LastAuditAt: now},
		{ID: "ing_greens", Name: "Salad Greens", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 4300, ReorderPoint: 4000, WasteFactor: 0.16, LastPurchaseCost: 180, LastAuditAt: now},
	}
	for _, ingredient := range ingredients {
		if _, err := tx.Exec(`
			INSERT INTO ingredients (
				id, name, unit, purchase_unit, purchase_to_usage, on_hand_qty,
				reorder_point, waste_factor, last_purchase_cost, last_audit_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				unit = excluded.unit,
				purchase_unit = excluded.purchase_unit,
				purchase_to_usage = excluded.purchase_to_usage,
				on_hand_qty = excluded.on_hand_qty,
				reorder_point = excluded.reorder_point,
				waste_factor = excluded.waste_factor,
				last_purchase_cost = excluded.last_purchase_cost,
				last_audit_at = excluded.last_audit_at
		`, ingredient.ID, ingredient.Name, ingredient.Unit, ingredient.PurchaseUnit, ingredient.PurchaseToUsage, ingredient.OnHandQty,
			ingredient.ReorderPoint, ingredient.WasteFactor, ingredient.LastPurchaseCost, ingredient.LastAuditAt); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoMenuTx(tx *sql.Tx) error {
	categories := []string{"Coffee", "Breakfast", "Small Plates", "Mains", "Dessert", "Beverage"}
	for index, category := range categories {
		if _, err := tx.Exec(`
			INSERT INTO menu_categories (id, name, sort_order, status)
			VALUES (?, ?, ?, 'active')
			ON CONFLICT(id) DO UPDATE SET name = excluded.name, sort_order = excluded.sort_order, status = excluded.status
		`, "cat_"+slug(category), category, (index+1)*10); err != nil {
			return err
		}
	}
	for _, item := range demoMenuCatalog() {
		if _, err := tx.Exec(`
			INSERT INTO menu_items (id, name, category, price, cost, status, route_id, tax_rate)
			VALUES (?, ?, ?, ?, ?, 'active', ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				category = excluded.category,
				price = excluded.price,
				cost = excluded.cost,
				status = excluded.status,
				route_id = excluded.route_id,
				tax_rate = excluded.tax_rate
		`, item.id, item.name, item.category, item.price, item.cost, item.routeID, item.taxRate); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoVendorsTx(tx *sql.Tx) error {
	vendors := []Vendor{
		{ID: "vendor_fresh_farm", Name: "Fresh Farm Co", Phone: "+919876500001", GSTIN: "29FRESH1234F1Z5", PaymentTerms: "7 days", QualityScore: 92, Status: "active"},
		{ID: "vendor_dairy", Name: "Morning Dairy", Phone: "+919876500002", GSTIN: "29DAIRY1234F1Z5", PaymentTerms: "Cash on delivery", QualityScore: 96, Status: "active"},
		{ID: "vendor_coffee", Name: "Blue Ridge Roasters", Phone: "+919876500003", GSTIN: "29COFFEE1234F1Z5", PaymentTerms: "Net 15", QualityScore: 95, Status: "active"},
		{ID: "vendor_bakery", Name: "Bangalore Bakehouse", Phone: "+919876500004", GSTIN: "29BAKE1234F1Z5", PaymentTerms: "Net 7", QualityScore: 89, Status: "active"},
		{ID: "vendor_seafood", Name: "Coastline Seafood", Phone: "+919876500005", GSTIN: "29SEA1234F1Z5", PaymentTerms: "Net 10", QualityScore: 86, Status: "active"},
	}
	for _, vendor := range vendors {
		if _, err := tx.Exec(`
			INSERT INTO vendors (id, name, phone, gstin, payment_terms, quality_score, status)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				phone = excluded.phone,
				gstin = excluded.gstin,
				payment_terms = excluded.payment_terms,
				quality_score = excluded.quality_score,
				status = excluded.status
		`, vendor.ID, vendor.Name, vendor.Phone, vendor.GSTIN, vendor.PaymentTerms, vendor.QualityScore, vendor.Status); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoSaleTx(tx *sql.Tx, date time.Time, orderIndex int, menu []demoMenuItem, customers []demoCustomerProfile, rng *rand.Rand, state *demoSeedState, stats *demoDayStats) error {
	hour := demoOrderHour(rng)
	created := date.Add(time.Duration(hour)*time.Hour + time.Duration(rng.Intn(60))*time.Minute + time.Duration(rng.Intn(50))*time.Second)
	closed := created.Add(time.Duration(18+rng.Intn(58)) * time.Minute)
	orderType := demoOrderType(hour, rng)
	tableName := ""
	if orderType == "dine_in" {
		tableName = demoTableName(1 + rng.Intn(50))
	}
	channel := "counter"
	if orderType == "delivery" {
		channel = "delivery"
	} else if orderType == "takeaway" {
		channel = "takeaway"
	}

	lines := demoOrderLines(menu, hour, rng)
	subtotal := 0.0
	for _, line := range lines {
		subtotal += line.lineTotal
	}
	discountTotal := demoDiscount(subtotal, date, rng)
	taxTotal := round2((subtotal - discountTotal) * 0.05)
	total := round2(subtotal - discountTotal + taxTotal)
	paymentMethod := demoPaymentMethod(orderType, rng)
	tendered, changeDue := demoTender(paymentMethod, total, rng)

	customerName := ""
	customerPhone := ""
	var customer *demoCustomerProfile
	if rng.Float64() < 0.64 {
		customer = demoPickCustomer(customers, rng)
		customerName = customer.name
		customerPhone = customer.phone
	}

	status := invoiceStatusClosed
	refundAmount := 0.0
	isRefunded := false
	isVoid := rng.Float64() < 0.008
	if !isVoid && total > 300 && rng.Float64() < 0.014 {
		isRefunded = true
		refundAmount = round2(total * (0.18 + rng.Float64()*0.38))
		if rng.Float64() < 0.16 {
			refundAmount = total
			status = invoiceStatusRefunded
		} else {
			status = invoiceStatusPartiallyRefunded
		}
	}
	if isVoid {
		status = invoiceStatusVoided
	}

	saleID := fmt.Sprintf("demo_sale_%s_%04d", date.Format("20060102"), orderIndex+1)
	invoiceNumber := fmt.Sprintf("NX-%s-%04d", date.Format("060102"), orderIndex+1)
	kotSentAt := closed.Add(-12 * time.Minute).Format(time.RFC3339)
	closedAt := closed.Format(time.RFC3339)
	voidReason := ""
	refundReason := ""
	approvedBy := ""
	approvedAt := ""
	if isVoid {
		kotSentAt = ""
		closedAt = ""
		voidReason = demoVoidReason(rng)
		approvedBy = "Ananya Manager"
		approvedAt = created.Add(4 * time.Minute).Format(time.RFC3339)
	}
	if isRefunded {
		refundReason = demoRefundReason(rng)
		approvedBy = "Ananya Manager"
		approvedAt = closed.Add(time.Duration(20+rng.Intn(160)) * time.Minute).Format(time.RFC3339)
	}

	if _, err := tx.Exec(`
		INSERT INTO sales (
			id, invoice_number, customer_name, customer_phone, table_name, order_type,
			subtotal, discount_total, tax_total, total, channel, payment_method, payment_tendered,
			payment_change, status, void_reason, refund_reason, approved_by, approved_at,
			kot_sent_at, closed_at, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, saleID, invoiceNumber, customerName, customerPhone, tableName, orderType, round2(subtotal), discountTotal, taxTotal, total,
		channel, paymentMethod, tendered, changeDue, status, voidReason, refundReason, approvedBy, approvedAt, kotSentAt, closedAt, created.Format(time.RFC3339)); err != nil {
		return err
	}

	for lineIndex, line := range lines {
		if _, err := tx.Exec(`
			INSERT INTO sale_lines (id, sale_id, item_id, item_name, quantity, unit_price, line_total, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, fmt.Sprintf("%s_line_%02d", saleID, lineIndex+1), saleID, line.item.id, line.item.name, line.quantity, line.unitPrice, line.lineTotal, line.notes); err != nil {
			return err
		}
	}

	if !isVoid {
		if _, err := tx.Exec(`
			INSERT INTO payments (id, invoice_id, method, amount, tendered, change_due, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, 'captured', ?)
		`, saleID+"_pay", saleID, paymentMethod, total, tendered, changeDue, closedAt); err != nil {
			return err
		}
		if err := seedDemoKOTsTx(tx, saleID, invoiceNumber, lines, kotSentAt, state); err != nil {
			return err
		}
		if isRefunded {
			refundID := saleID + "_refund"
			if _, err := tx.Exec(`
				INSERT INTO refunds (id, invoice_id, amount, reason, approved_by, approved_at, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, refundID, saleID, refundAmount, refundReason, approvedBy, approvedAt, approvedAt); err != nil {
				return err
			}
			if _, err := tx.Exec(`
				INSERT INTO payments (id, invoice_id, method, amount, tendered, change_due, status, created_at)
				VALUES (?, ?, 'refund', ?, ?, 0, 'refunded', ?)
			`, saleID+"_refund_pay", saleID, -refundAmount, -refundAmount, approvedAt); err != nil {
				return err
			}
			state.result.Refunds++
			stats.refundTotal += refundAmount
		}
		if customer != nil {
			demoAccumulateCustomer(state.customerStats, *customer, total-refundAmount, closedAt, lines[0].item.name)
		}
		stats.salesTotal += total
		stats.orderCount++
		stats.discountTotal += discountTotal
		switch paymentMethod {
		case "cash":
			stats.cashTotal += total
		case "upi":
			stats.upiTotal += total
		case "card":
			stats.cardTotal += total
		case "razorpay":
			stats.razorpayTotal += total
		}
		state.result.SalesTotal += total - refundAmount
		state.invoiceIDs = appendRecentID(state.invoiceIDs, saleID, 80)
	} else {
		state.result.Voids++
		stats.voidCount++
	}
	if err := seedDemoInvoiceEventsTx(tx, saleID, isVoid, isRefunded, created, closed, invoiceNumber); err != nil {
		return err
	}
	state.result.Invoices++
	state.invoiceSeq++
	return nil
}

func seedDemoKOTsTx(tx *sql.Tx, saleID, invoiceNumber string, lines []demoSaleLine, createdAt string, state *demoSeedState) error {
	byRoute := map[string][]demoSaleLine{}
	for _, line := range lines {
		byRoute[line.item.routeID] = append(byRoute[line.item.routeID], line)
	}
	for routeID, routeLines := range byRoute {
		routeName := "Kitchen"
		if routeID == "route_bar" {
			routeName = "Barista"
		}
		ticketID := fmt.Sprintf("demo_kot_%06d", state.kotSeq)
		ticketNumber := fmt.Sprintf("KOT-%06d", state.kotSeq)
		if _, err := tx.Exec(`
			INSERT INTO kitchen_tickets (id, sale_id, invoice_number, ticket_number, route_id, route_name, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, 'served', ?)
		`, ticketID, saleID, invoiceNumber, ticketNumber, routeID, routeName, createdAt); err != nil {
			return err
		}
		for index, line := range routeLines {
			if _, err := tx.Exec(`
				INSERT INTO kitchen_ticket_lines (id, ticket_id, item_id, item_name, quantity, notes, status)
				VALUES (?, ?, ?, ?, ?, ?, 'served')
			`, fmt.Sprintf("%s_line_%02d", ticketID, index+1), ticketID, line.item.id, line.item.name, line.quantity, line.notes); err != nil {
				return err
			}
		}
		state.kotSeq++
		state.result.KitchenTickets++
	}
	return nil
}

func seedDemoInvoiceEventsTx(tx *sql.Tx, saleID string, isVoid, isRefunded bool, created, closed time.Time, invoiceNumber string) error {
	events := []InvoiceEvent{
		{ID: saleID + "_event_created", InvoiceID: saleID, Type: "created", Title: "Bill created", Detail: invoiceNumber, Actor: "operator", CreatedAt: created.Format(time.RFC3339)},
	}
	if isVoid {
		events = append(events, InvoiceEvent{ID: saleID + "_event_void", InvoiceID: saleID, Type: "voided", Title: "Invoice voided", Detail: "Manager approved operational void", Actor: "Ananya Manager", CreatedAt: created.Add(4 * time.Minute).Format(time.RFC3339)})
	} else {
		events = append(events,
			InvoiceEvent{ID: saleID + "_event_kot", InvoiceID: saleID, Type: "kot_sent", Title: "KOT sent", Detail: "Kitchen route tickets created", Actor: "operator", CreatedAt: closed.Add(-12 * time.Minute).Format(time.RFC3339)},
			InvoiceEvent{ID: saleID + "_event_closed", InvoiceID: saleID, Type: "closed", Title: "Bill closed", Detail: "Payment captured", Actor: "operator", CreatedAt: closed.Format(time.RFC3339)},
		)
		if isRefunded {
			events = append(events, InvoiceEvent{ID: saleID + "_event_refund", InvoiceID: saleID, Type: "refunded", Title: "Refund recorded", Detail: "Customer refund approved", Actor: "Ananya Manager", CreatedAt: closed.Add(45 * time.Minute).Format(time.RFC3339)})
		}
	}
	for _, event := range events {
		if _, err := tx.Exec(`
			INSERT INTO invoice_events (id, invoice_id, type, title, detail, actor, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, event.ID, event.InvoiceID, event.Type, event.Title, event.Detail, event.Actor, event.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoDayCloseTx(tx *sql.Tx, date time.Time, stats *demoDayStats, rng *rand.Rand) error {
	variance := 0.0
	if rng.Float64() < 0.24 {
		variance = float64(rng.Intn(421) - 210)
	}
	cashCounted := round2(stats.cashTotal + variance)
	id := "demo_dayclose_" + date.Format("20060102")
	closedAt := date.Add(23*time.Hour + 42*time.Minute).Format(time.RFC3339)
	_, err := tx.Exec(`
		INSERT INTO day_closes (
			id, business_date, closed_at, status, sales_total, cash_expected, cash_counted,
			cash_variance, upi_total, card_total, razorpay_total, refund_total, void_count,
			discount_total, notes
		)
		VALUES (?, ?, ?, 'closed', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, date.Format("2006-01-02"), closedAt, round2(stats.salesTotal), round2(stats.cashTotal), cashCounted,
		round2(cashCounted-stats.cashTotal), round2(stats.upiTotal), round2(stats.cardTotal), round2(stats.razorpayTotal),
		round2(stats.refundTotal), stats.voidCount, round2(stats.discountTotal), "Demo day close with realistic tender mix")
	return err
}

func seedDemoCustomersTx(tx *sql.Tx, customers map[string]*demoCustomerStats) error {
	for _, customer := range customers {
		if customer.visits == 0 {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO customers (id, name, phone, total_spend, visit_count, last_visit_at, favorite_item)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, customer.id, customer.name, customer.phone, round2(customer.total), customer.visits, customer.lastVisit, demoFavoriteItem(customer.favorites)); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoPurchaseOrderTx(tx *sql.Tx, date time.Time, rng *rand.Rand, state *demoSeedState) error {
	vendors := []string{"vendor_fresh_farm", "vendor_dairy", "vendor_coffee", "vendor_bakery", "vendor_seafood"}
	ingredients := []struct {
		id       string
		unitCost float64
		minQty   float64
		maxQty   float64
	}{
		{"ing_beans", 840, 8, 22},
		{"ing_milk", 64, 45, 130},
		{"ing_tomato", 44, 12, 38},
		{"ing_chicken", 292, 12, 30},
		{"ing_paneer", 330, 8, 24},
		{"ing_basmati", 118, 20, 60},
		{"ing_salmon", 1150, 4, 12},
		{"ing_greens", 180, 5, 18},
		{"ing_chocolate", 640, 4, 12},
		{"ing_bun", 16, 60, 160},
	}
	orderCount := 1
	if rng.Float64() < 0.45 {
		orderCount = 2
	}
	for order := 0; order < orderCount; order++ {
		vendorID := vendors[(state.purchaseSeq+order)%len(vendors)]
		poID := fmt.Sprintf("demo_po_%06d", state.purchaseSeq)
		poNumber := fmt.Sprintf("PO-%06d", state.purchaseSeq)
		created := date.Add(time.Duration(7+order)*time.Hour + 20*time.Minute)
		subtotal := 0.0
		rejectedTotal := 0.0
		lineCount := 3 + rng.Intn(4)
		if _, err := tx.Exec(`
			INSERT INTO purchase_orders (id, vendor_id, po_number, status, expected_date, subtotal, rejected_total, created_at, closed_at)
			VALUES (?, ?, ?, 'received', ?, 0, 0, ?, ?)
		`, poID, vendorID, poNumber, date.AddDate(0, 0, 1).Format("2006-01-02"), created.Format(time.RFC3339), created.AddDate(0, 0, 1).Format(time.RFC3339)); err != nil {
			return err
		}
		batchID := poID + "_delivery"
		if _, err := tx.Exec(`
			INSERT INTO delivery_batches (id, vendor_name, invoice_number, status, created_at)
			VALUES (?, ?, ?, 'accepted', ?)
		`, batchID, demoVendorName(vendorID), "INV-"+poNumber, created.AddDate(0, 0, 1).Format(time.RFC3339)); err != nil {
			return err
		}
		for line := 0; line < lineCount; line++ {
			ingredient := ingredients[(state.purchaseSeq+line*3)%len(ingredients)]
			ordered := round2(ingredient.minQty + rng.Float64()*(ingredient.maxQty-ingredient.minQty))
			rejected := 0.0
			reason := ""
			if rng.Float64() < 0.18 {
				rejected = round2(ordered * (0.04 + rng.Float64()*0.12))
				reason = "QC rejection during receiving"
			}
			accepted := round2(ordered - rejected)
			subtotal += ordered * ingredient.unitCost
			rejectedTotal += rejected * ingredient.unitCost
			lineID := fmt.Sprintf("%s_line_%02d", poID, line+1)
			if _, err := tx.Exec(`
				INSERT INTO purchase_order_lines (
					id, purchase_order_id, ingredient_id, ordered_qty, accepted_qty, rejected_qty,
					unit_cost, rejection_reason
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, lineID, poID, ingredient.id, ordered, accepted, rejected, ingredient.unitCost, reason); err != nil {
				return err
			}
			lineStatus := "accepted"
			if rejected > 0 {
				lineStatus = "partial_rejected"
			}
			if _, err := tx.Exec(`
				INSERT INTO delivery_lines (
					id, batch_id, ingredient_id, ordered_qty, accepted_qty, rejected_qty,
					unit_cost, rejection_reason, status
				)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, lineID+"_delivery", batchID, ingredient.id, ordered, accepted, rejected, ingredient.unitCost, reason, lineStatus); err != nil {
				return err
			}
			if _, err := tx.Exec(`
				INSERT INTO inventory_movements (id, ingredient_id, type, quantity, reason, reference_id, created_at)
				VALUES (?, ?, 'purchase_acceptance', ?, 'Demo PO receiving', ?, ?)
			`, lineID+"_movement", ingredient.id, accepted, poID, created.AddDate(0, 0, 1).Format(time.RFC3339)); err != nil {
				return err
			}
		}
		if _, err := tx.Exec("UPDATE purchase_orders SET subtotal = ?, rejected_total = ? WHERE id = ?", round2(subtotal), round2(rejectedTotal), poID); err != nil {
			return err
		}
		if rejectedTotal > 0 {
			if _, err := tx.Exec(`
				INSERT INTO vendor_debit_notes (id, vendor_id, purchase_order_id, amount, reason, status, created_at)
				VALUES (?, ?, ?, ?, 'Quality rejection adjustment', 'ready', ?)
			`, poID+"_debit", vendorID, poID, round2(rejectedTotal), created.AddDate(0, 0, 1).Format(time.RFC3339)); err != nil {
				return err
			}
		}
		state.purchaseSeq++
		state.result.PurchaseOrders++
	}
	return nil
}

func seedDemoLiveSessionsTx(tx *sql.Tx, today time.Time, menu []demoMenuItem, rng *rand.Rand, state *demoSeedState) error {
	statuses := []string{"queued", "preparing", "ready", "queued", "preparing", "ready", "preparing", "queued"}
	for index := 0; index < 12; index++ {
		sessionID := fmt.Sprintf("demo_live_session_%02d", index+1)
		tableID := fmt.Sprintf("tbl_%02d", index+1)
		waiterID := demoWaiterID(index % 20)
		opened := today.Add(time.Duration(11+index%9)*time.Hour + time.Duration(5+index*3)*time.Minute)
		lines := demoOrderLines(menu, 12+index%9, rng)
		subtotal := 0.0
		for _, line := range lines {
			subtotal += line.lineTotal
		}
		service := round2(subtotal * 0.05)
		tax := round2((subtotal + service) * 0.05)
		total := round2(subtotal + service + tax)
		invoiceID := fmt.Sprintf("demo_live_invoice_%02d", index+1)
		invoiceNumber := fmt.Sprintf("NX-LIVE-%02d", index+1)
		status := invoiceStatusDraft
		kotSentAt := ""
		if index < len(statuses) {
			status = invoiceStatusKOTSent
			kotSentAt = opened.Add(8 * time.Minute).Format(time.RFC3339)
		}
		if _, err := tx.Exec(`
			INSERT INTO order_sessions (
				id, table_id, waiter_id, guest_count, service_mode, status, invoice_id,
				subtotal, tax_total, service_charge, total, opened_at
			)
			VALUES (?, ?, ?, ?, 'dine_in', 'open', ?, ?, ?, ?, ?, ?)
		`, sessionID, tableID, waiterID, 2+rng.Intn(5), invoiceID, round2(subtotal), tax, service, total, opened.Format(time.RFC3339)); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE dining_tables SET status = 'occupied', active_session_id = ? WHERE id = ?", sessionID, tableID); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO sales (
				id, invoice_number, customer_name, customer_phone, table_name, order_type,
				subtotal, discount_total, tax_total, total, channel, payment_method, status,
				kot_sent_at, created_at
			)
			VALUES (?, ?, 'Live Table', '', ?, 'dine_in', ?, 0, ?, ?, 'floor', 'cash', ?, ?, ?)
		`, invoiceID, invoiceNumber, demoTableName(index+1), round2(subtotal+service), tax, total, status, kotSentAt, opened.Format(time.RFC3339)); err != nil {
			return err
		}
		for lineIndex, line := range lines {
			kotStatus := "not_sent"
			if index < len(statuses) {
				kotStatus = statuses[index]
			}
			lineID := fmt.Sprintf("%s_line_%02d", sessionID, lineIndex+1)
			if _, err := tx.Exec(`
				INSERT INTO order_session_lines (
					id, session_id, item_id, item_name, quantity, unit_price,
					modifier_total, line_total, notes, status, kot_status
				)
				VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, 'open', ?)
			`, lineID, sessionID, line.item.id, line.item.name, line.quantity, line.unitPrice, line.lineTotal, line.notes, kotStatus); err != nil {
				return err
			}
			if _, err := tx.Exec(`
				INSERT INTO sale_lines (id, sale_id, item_id, item_name, quantity, unit_price, line_total, notes)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, lineID+"_sale", invoiceID, line.item.id, line.item.name, line.quantity, line.unitPrice, line.lineTotal, line.notes); err != nil {
				return err
			}
		}
		if index < len(statuses) {
			if err := seedDemoLiveKOTsTx(tx, invoiceID, invoiceNumber, lines, statuses[index], opened.Add(8*time.Minute), state); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`
			INSERT INTO order_session_events (id, session_id, type, detail, actor, created_at)
			VALUES (?, ?, 'opened', 'Demo live table opened', 'operator', ?)
		`, sessionID+"_event", sessionID, opened.Format(time.RFC3339)); err != nil {
			return err
		}
		state.result.Invoices++
		state.invoiceSeq++
	}
	return nil
}

func seedDemoLiveKOTsTx(tx *sql.Tx, saleID, invoiceNumber string, lines []demoSaleLine, status string, created time.Time, state *demoSeedState) error {
	byRoute := map[string][]demoSaleLine{}
	for _, line := range lines {
		byRoute[line.item.routeID] = append(byRoute[line.item.routeID], line)
	}
	for routeID, routeLines := range byRoute {
		routeName := "Kitchen"
		if routeID == "route_bar" {
			routeName = "Barista"
		}
		ticketID := fmt.Sprintf("demo_live_kot_%06d", state.kotSeq)
		ticketNumber := fmt.Sprintf("KOT-%06d", state.kotSeq)
		if _, err := tx.Exec(`
			INSERT INTO kitchen_tickets (id, sale_id, invoice_number, ticket_number, route_id, route_name, status, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, ticketID, saleID, invoiceNumber, ticketNumber, routeID, routeName, status, created.Format(time.RFC3339)); err != nil {
			return err
		}
		for index, line := range routeLines {
			if _, err := tx.Exec(`
				INSERT INTO kitchen_ticket_lines (id, ticket_id, item_id, item_name, quantity, notes, status)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, fmt.Sprintf("%s_line_%02d", ticketID, index+1), ticketID, line.item.id, line.item.name, line.quantity, line.notes, status); err != nil {
				return err
			}
		}
		state.kotSeq++
		state.result.KitchenTickets++
	}
	return nil
}

func seedDemoOperationalNoiseTx(tx *sql.Tx, today time.Time, rng *rand.Rand, state *demoSeedState) error {
	now := today.Add(13*time.Hour + 30*time.Minute).Format(time.RFC3339)
	for index := 0; index < 28; index++ {
		status := "sent"
		if index%7 == 0 {
			status = "ready"
		}
		if _, err := tx.Exec(`
			INSERT INTO notification_queue (id, invoice_id, channel, recipient, template, payload, status, created_at, sent_at)
			VALUES (?, ?, 'whatsapp', ?, 'receipt', '{}', ?, ?, ?)
		`, fmt.Sprintf("demo_notification_%02d", index+1), demoRecentInvoiceID(state, index), fmt.Sprintf("+9198%08d", 700000+index), status, now, ternaryString(status == "sent", now, "")); err != nil {
			return err
		}
	}
	for index := 0; index < 8; index++ {
		status := "ready"
		if index%4 == 0 {
			status = "failed"
		}
		if _, err := tx.Exec(`
			INSERT INTO print_jobs (id, kind, reference_id, printer_id, target, payload, status, attempts, last_error, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, fmt.Sprintf("demo_print_%02d", index+1), ternaryString(index%2 == 0, "kot", "invoice"), demoRecentInvoiceID(state, index),
			ternaryString(index%2 == 0, "printer_kitchen", "printer_receipt"), "tcp://192.168.1.50:9100", "NEXUS DEMO PRINT", status,
			ternaryInt(status == "failed", 2, 0), ternaryString(status == "failed", "Printer offline during rush", ""), now); err != nil {
			return err
		}
	}
	for index := 0; index < 10; index++ {
		status := "paid"
		if index%5 == 0 {
			status = "failed"
		}
		if _, err := tx.Exec(`
			INSERT INTO payment_requests (
				id, invoice_id, provider, amount, currency, status, reference, checkout_url,
				qr_payload, created_at, updated_at
			)
			VALUES (?, ?, 'razorpay', ?, 'INR', ?, ?, ?, ?, ?, ?)
		`, fmt.Sprintf("demo_payment_request_%02d", index+1), demoRecentInvoiceID(state, index), round2(900+rng.Float64()*4200), status,
			fmt.Sprintf("DEMO-RZP-%02d", index+1), "https://rzp.io/i/demo", "razorpay://payment-link/demo", now, now); err != nil {
			return err
		}
	}
	for index := 0; index < 180; index++ {
		status := "synced"
		syncedAt := today.Add(-time.Duration(index) * time.Hour).Format(time.RFC3339)
		lastError := ""
		if index < 36 {
			status = "pending"
			syncedAt = ""
		}
		if index >= 36 && index < 40 {
			status = "failed"
			syncedAt = ""
			lastError = "Demo cloud endpoint unavailable"
		}
		if _, err := tx.Exec(`
			INSERT INTO sync_log (
				id, entity, entity_id, operation, payload, status, created_at, synced_at, attempts, last_error
			)
			VALUES (?, ?, ?, 'upsert', '{}', ?, ?, ?, ?, ?)
		`, fmt.Sprintf("demo_sync_%03d", index+1), demoSyncEntity(index), demoRecentInvoiceID(state, index), status,
			today.Add(-time.Duration(index+1)*time.Hour).Format(time.RFC3339), syncedAt, ternaryInt(status == "failed", 3, 1), lastError); err != nil {
			return err
		}
	}
	return nil
}

func seedDemoSequencesTx(tx *sql.Tx, state *demoSeedState) error {
	settings := map[string]string{
		"next_invoice_seq": fmt.Sprintf("%d", state.invoiceSeq+100),
		"next_kot_seq":     fmt.Sprintf("%d", state.kotSeq+100),
		"next_po_seq":      fmt.Sprintf("%d", state.purchaseSeq+100),
	}
	for key, value := range settings {
		if _, err := tx.Exec(`
			INSERT INTO app_settings (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value
		`, key, value); err != nil {
			return err
		}
	}
	return nil
}

func demoMenuCatalog() []demoMenuItem {
	return []demoMenuItem{
		{"item_latte", "Classic Latte", "Coffee", 180, 58, "route_bar", 0.05, 95},
		{"item_cappuccino", "Cappuccino", "Coffee", 170, 54, "route_bar", 0.05, 80},
		{"item_americano", "Iced Americano", "Coffee", 155, 38, "route_bar", 0.05, 86},
		{"item_flat_white", "Flat White", "Coffee", 190, 64, "route_bar", 0.05, 52},
		{"item_mocha", "Dark Mocha", "Coffee", 220, 82, "route_bar", 0.05, 38},
		{"item_filter_kaapi", "South Filter Kaapi", "Coffee", 120, 28, "route_bar", 0.05, 74},
		{"item_avocado_toast", "Avocado Toast", "Breakfast", 340, 132, "route_kitchen", 0.05, 46},
		{"item_masala_omelette", "Masala Omelette", "Breakfast", 260, 88, "route_kitchen", 0.05, 62},
		{"item_banana_pancake", "Banana Pancake Stack", "Breakfast", 310, 118, "route_kitchen", 0.05, 34},
		{"item_idli_benedict", "Idli Benedict", "Breakfast", 290, 92, "route_kitchen", 0.05, 42},
		{"item_sandwich", "Chicken Club Sandwich", "Small Plates", 260, 112, "route_kitchen", 0.05, 88},
		{"item_wrap", "Tomato Basil Wrap", "Small Plates", 220, 86, "route_kitchen", 0.05, 70},
		{"item_truffle_fries", "Truffle Masala Fries", "Small Plates", 240, 72, "route_kitchen", 0.05, 78},
		{"item_paneer_tikka", "Paneer Tikka Bites", "Small Plates", 360, 142, "route_kitchen", 0.05, 57},
		{"item_chicken_wings", "Korean Chicken Wings", "Small Plates", 420, 178, "route_kitchen", 0.05, 66},
		{"item_salmon_bites", "Crispy Salmon Bites", "Small Plates", 620, 310, "route_kitchen", 0.05, 24},
		{"item_butter_chicken", "Butter Chicken Bowl", "Mains", 520, 222, "route_kitchen", 0.05, 90},
		{"item_paneer_bowl", "Paneer Makhani Bowl", "Mains", 460, 168, "route_kitchen", 0.05, 74},
		{"item_pesto_pasta", "Basil Pesto Pasta", "Mains", 480, 172, "route_kitchen", 0.05, 60},
		{"item_salmon_plate", "Grilled Salmon Plate", "Mains", 890, 468, "route_kitchen", 0.05, 22},
		{"item_mushroom_risotto", "Wild Mushroom Risotto", "Mains", 560, 212, "route_kitchen", 0.05, 39},
		{"item_biryani", "Hyderabadi Chicken Biryani", "Mains", 540, 224, "route_kitchen", 0.05, 84},
		{"item_caesar_salad", "Charred Caesar Salad", "Mains", 410, 138, "route_kitchen", 0.05, 44},
		{"item_cheesecake", "Basque Cheesecake", "Dessert", 310, 98, "route_kitchen", 0.05, 58},
		{"item_brownie", "Walnut Brownie", "Dessert", 240, 70, "route_kitchen", 0.05, 76},
		{"item_tiramisu", "Filter Coffee Tiramisu", "Dessert", 340, 112, "route_kitchen", 0.05, 45},
		{"item_gulab_jamun", "Gulab Jamun Brulee", "Dessert", 280, 76, "route_kitchen", 0.05, 36},
		{"item_lime_soda", "Fresh Lime Soda", "Beverage", 140, 32, "route_bar", 0.05, 82},
		{"item_kombucha", "Hibiscus Kombucha", "Beverage", 220, 84, "route_bar", 0.05, 34},
		{"item_iced_tea", "Peach Iced Tea", "Beverage", 190, 48, "route_bar", 0.05, 64},
		{"item_sparkling_water", "Sparkling Water", "Beverage", 160, 52, "route_bar", 0.05, 28},
	}
}

func demoCustomerProfiles(count int) []demoCustomerProfile {
	first := []string{"Aarav", "Vivaan", "Aditya", "Vihaan", "Arjun", "Sai", "Reyansh", "Ayaan", "Krishna", "Ishaan", "Ananya", "Diya", "Myra", "Ira", "Anika", "Kiara", "Sara", "Tara", "Meera", "Rhea", "Rahul", "Neha", "Priya", "Rohan", "Kavya", "Kabir", "Nikhil", "Aisha", "Dev", "Maya"}
	last := []string{"Mehra", "Rao", "Iyer", "Menon", "Kapoor", "Sinha", "Shah", "Reddy", "Nair", "Gupta", "Bose", "Khanna", "Shetty", "Pillai", "Agarwal", "Chopra", "Patel", "Varma", "Malhotra", "Joshi"}
	profiles := make([]demoCustomerProfile, 0, count)
	for index := 0; index < count; index++ {
		profiles = append(profiles, demoCustomerProfile{
			id:    fmt.Sprintf("demo_customer_%03d", index+1),
			name:  fmt.Sprintf("%s %s", first[index%len(first)], last[(index*7)%len(last)]),
			phone: fmt.Sprintf("+9198%08d", 500000+index),
		})
	}
	return profiles
}

func demoWaiterID(index int) string {
	if index <= 0 {
		return "staff_waiter"
	}
	return fmt.Sprintf("staff_waiter_%02d", index+1)
}

func demoOrderCount(date time.Time, dayIndex int, rng *rand.Rand) int {
	base := 52.0 + math.Sin(float64(dayIndex)/14.0)*7
	switch date.Weekday() {
	case time.Friday:
		base *= 1.24
	case time.Saturday:
		base *= 1.58
	case time.Sunday:
		base *= 1.18
	case time.Monday:
		base *= 0.82
	case time.Tuesday:
		base *= 0.9
	}
	if date.Month() == time.December || date.Month() == time.January {
		base *= 1.12
	}
	return maxInt(28, int(math.Round(base))+rng.Intn(17)-8)
}

func demoOrderHour(rng *rand.Rand) int {
	weighted := []int{8, 8, 9, 9, 10, 11, 12, 12, 12, 13, 13, 13, 14, 15, 16, 18, 18, 19, 19, 20, 20, 20, 21, 21, 22}
	return weighted[rng.Intn(len(weighted))]
}

func demoOrderType(hour int, rng *rand.Rand) string {
	value := rng.Float64()
	if hour < 11 {
		if value < 0.48 {
			return "dine_in"
		}
		if value < 0.86 {
			return "takeaway"
		}
		return "delivery"
	}
	if value < 0.68 {
		return "dine_in"
	}
	if value < 0.88 {
		return "takeaway"
	}
	return "delivery"
}

func demoOrderLines(menu []demoMenuItem, hour int, rng *rand.Rand) []demoSaleLine {
	count := 1
	value := rng.Float64()
	switch {
	case value < 0.34:
		count = 1
	case value < 0.74:
		count = 2
	case value < 0.92:
		count = 3
	case value < 0.985:
		count = 4
	default:
		count = 5
	}
	lines := make([]demoSaleLine, 0, count)
	for index := 0; index < count; index++ {
		item := demoPickMenuItem(menu, hour, rng)
		quantity := 1
		if (item.category == "Coffee" || item.category == "Beverage") && rng.Float64() < 0.18 {
			quantity = 2
		}
		if item.category == "Small Plates" && rng.Float64() < 0.1 {
			quantity = 2
		}
		notes := ""
		if rng.Float64() < 0.08 {
			notes = demoLineNote(item.category, rng)
		}
		lines = append(lines, demoSaleLine{
			item:      item,
			quantity:  quantity,
			unitPrice: item.price,
			lineTotal: round2(item.price * float64(quantity)),
			notes:     notes,
		})
	}
	return lines
}

func demoPickMenuItem(menu []demoMenuItem, hour int, rng *rand.Rand) demoMenuItem {
	total := 0
	weights := make([]int, len(menu))
	for index, item := range menu {
		weight := item.weight
		switch {
		case hour < 11 && item.category == "Breakfast":
			weight *= 3
		case hour < 11 && item.category == "Coffee":
			weight *= 2
		case hour >= 11 && hour < 16 && (item.category == "Mains" || item.category == "Small Plates"):
			weight *= 2
		case hour >= 18 && item.category == "Dessert":
			weight *= 2
		case hour >= 18 && item.category == "Mains":
			weight *= 2
		}
		if hour < 11 && item.category == "Mains" {
			weight = maxInt(1, weight/3)
		}
		weights[index] = weight
		total += weight
	}
	pick := rng.Intn(total)
	for index, weight := range weights {
		pick -= weight
		if pick < 0 {
			return menu[index]
		}
	}
	return menu[len(menu)-1]
}

func demoPaymentMethod(orderType string, rng *rand.Rand) string {
	value := rng.Float64()
	if orderType == "delivery" {
		if value < 0.58 {
			return "razorpay"
		}
		if value < 0.86 {
			return "upi"
		}
		return "card"
	}
	if value < 0.31 {
		return "cash"
	}
	if value < 0.69 {
		return "upi"
	}
	if value < 0.92 {
		return "card"
	}
	return "razorpay"
}

func demoTender(method string, total float64, rng *rand.Rand) (float64, float64) {
	if method != "cash" {
		return total, 0
	}
	tendered := total
	if rng.Float64() < 0.68 {
		tendered = math.Ceil(total/100) * 100
	}
	return round2(tendered), round2(math.Max(0, tendered-total))
}

func demoDiscount(subtotal float64, date time.Time, rng *rand.Rand) float64 {
	if subtotal <= 0 {
		return 0
	}
	chance := 0.11
	if date.Weekday() == time.Tuesday || date.Weekday() == time.Wednesday {
		chance = 0.17
	}
	if rng.Float64() > chance {
		return 0
	}
	rate := 0.05
	if rng.Float64() < 0.35 {
		rate = 0.1
	}
	return round2(subtotal * rate)
}

func demoPickCustomer(customers []demoCustomerProfile, rng *rand.Rand) *demoCustomerProfile {
	index := int(math.Pow(rng.Float64(), 1.55) * float64(len(customers)))
	if index >= len(customers) {
		index = len(customers) - 1
	}
	return &customers[index]
}

func demoAccumulateCustomer(stats map[string]*demoCustomerStats, profile demoCustomerProfile, spend float64, lastVisit, favorite string) {
	customer := stats[profile.phone]
	if customer == nil {
		customer = &demoCustomerStats{id: profile.id, name: profile.name, phone: profile.phone, favorites: map[string]int{}}
		stats[profile.phone] = customer
	}
	customer.total += math.Max(0, spend)
	customer.visits++
	customer.lastVisit = lastVisit
	customer.favorites[favorite]++
}

func demoFavoriteItem(favorites map[string]int) string {
	bestName := ""
	bestCount := 0
	for name, count := range favorites {
		if count > bestCount || (count == bestCount && name < bestName) {
			bestName = name
			bestCount = count
		}
	}
	return bestName
}

func demoLineNote(category string, rng *rand.Rand) string {
	switch category {
	case "Coffee":
		return []string{"extra hot", "less sugar", "oat milk", "double shot"}[rng.Intn(4)]
	case "Mains", "Small Plates":
		return []string{"no onion", "extra spicy", "sauce on side", "gluten conscious"}[rng.Intn(4)]
	default:
		return []string{"less ice", "share plates", "birthday table"}[rng.Intn(3)]
	}
}

func demoVoidReason(rng *rand.Rand) string {
	reasons := []string{"Wrong table punched", "Duplicate order", "Guest changed order", "Training correction"}
	return reasons[rng.Intn(len(reasons))]
}

func demoRefundReason(rng *rand.Rand) string {
	reasons := []string{"Service recovery", "Item returned", "Delivery delay", "Payment correction"}
	return reasons[rng.Intn(len(reasons))]
}

func demoTableName(number int) string {
	switch {
	case number <= 20:
		return fmt.Sprintf("M-%02d", number)
	case number <= 30:
		return fmt.Sprintf("P-%02d", number-20)
	case number <= 38:
		return fmt.Sprintf("F-%02d", number-30)
	case number <= 44:
		return fmt.Sprintf("D-%02d", number-38)
	default:
		return fmt.Sprintf("R-%02d", number-44)
	}
}

func demoVendorName(id string) string {
	names := map[string]string{
		"vendor_fresh_farm": "Fresh Farm Co",
		"vendor_dairy":      "Morning Dairy",
		"vendor_coffee":     "Blue Ridge Roasters",
		"vendor_bakery":     "Bangalore Bakehouse",
		"vendor_seafood":    "Coastline Seafood",
	}
	return firstNonEmpty(names[id], id)
}

func demoRecentInvoiceID(state *demoSeedState, offset int) string {
	if len(state.invoiceIDs) == 0 {
		return ""
	}
	return state.invoiceIDs[offset%len(state.invoiceIDs)]
}

func demoSyncEntity(index int) string {
	entities := []string{"sale", "customer", "inventory", "purchase_order", "day_close"}
	return entities[index%len(entities)]
}

func appendRecentID(ids []string, id string, limit int) []string {
	ids = append([]string{id}, ids...)
	if len(ids) > limit {
		return ids[:limit]
	}
	return ids
}

func ternaryString(condition bool, yes, no string) string {
	if condition {
		return yes
	}
	return no
}

func ternaryInt(condition bool, yes, no int) int {
	if condition {
		return yes
	}
	return no
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
