package nexus

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS restaurants (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	phone TEXT NOT NULL DEFAULT '',
	website TEXT NOT NULL DEFAULT '',
	brand_voice TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ingredients (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	unit TEXT NOT NULL,
	purchase_unit TEXT NOT NULL DEFAULT '',
	purchase_to_usage REAL NOT NULL DEFAULT 1,
	on_hand_qty REAL NOT NULL DEFAULT 0,
	reorder_point REAL NOT NULL DEFAULT 0,
	waste_factor REAL NOT NULL DEFAULT 0,
	last_purchase_cost REAL NOT NULL DEFAULT 0,
	last_audit_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS menu_items (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	category TEXT NOT NULL,
	price REAL NOT NULL,
	cost REAL NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	route_id TEXT NOT NULL DEFAULT 'route_kitchen',
	tax_rate REAL NOT NULL DEFAULT 0.05
);

CREATE TABLE IF NOT EXISTS recipes (
	id TEXT PRIMARY KEY,
	item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
	ingredient_id TEXT NOT NULL REFERENCES ingredients(id),
	quantity REAL NOT NULL,
	waste_factor_override REAL,
	UNIQUE(item_id, ingredient_id)
);

CREATE TABLE IF NOT EXISTS customers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	phone TEXT NOT NULL UNIQUE,
	total_spend REAL NOT NULL DEFAULT 0,
	visit_count INTEGER NOT NULL DEFAULT 0,
	last_visit_at TEXT NOT NULL,
	favorite_item TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS sales (
	id TEXT PRIMARY KEY,
	invoice_number TEXT NOT NULL DEFAULT '',
	customer_name TEXT NOT NULL DEFAULT '',
	customer_phone TEXT NOT NULL DEFAULT '',
	table_name TEXT NOT NULL DEFAULT '',
	order_type TEXT NOT NULL DEFAULT 'dine_in',
	subtotal REAL NOT NULL DEFAULT 0,
	discount_total REAL NOT NULL DEFAULT 0,
	tax_total REAL NOT NULL DEFAULT 0,
	total REAL NOT NULL,
	channel TEXT NOT NULL,
	payment_method TEXT NOT NULL DEFAULT 'cash',
	payment_tendered REAL NOT NULL DEFAULT 0,
	payment_change REAL NOT NULL DEFAULT 0,
	status TEXT NOT NULL,
	source_invoice_id TEXT NOT NULL DEFAULT '',
	split_group_id TEXT NOT NULL DEFAULT '',
	void_reason TEXT NOT NULL DEFAULT '',
	refund_reason TEXT NOT NULL DEFAULT '',
	approved_by TEXT NOT NULL DEFAULT '',
	approved_at TEXT NOT NULL DEFAULT '',
	kot_sent_at TEXT NOT NULL DEFAULT '',
	closed_at TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sale_lines (
	id TEXT PRIMARY KEY,
	sale_id TEXT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
	item_id TEXT NOT NULL REFERENCES menu_items(id),
	item_name TEXT NOT NULL,
	quantity INTEGER NOT NULL,
	unit_price REAL NOT NULL,
	line_total REAL NOT NULL,
	notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS app_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	can_manage_settings INTEGER NOT NULL DEFAULT 0,
	can_void INTEGER NOT NULL DEFAULT 0,
	can_refund INTEGER NOT NULL DEFAULT 0,
	can_close_day INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS restaurant_settings (
	id TEXT PRIMARY KEY CHECK (id = 'primary'),
	restaurant_name TEXT NOT NULL DEFAULT '',
	gstin TEXT NOT NULL DEFAULT '',
	legal_name TEXT NOT NULL DEFAULT '',
	address TEXT NOT NULL DEFAULT '',
	state TEXT NOT NULL DEFAULT '',
	tax_mode TEXT NOT NULL DEFAULT 'gst',
	default_tax_rate REAL NOT NULL DEFAULT 0.05,
	service_charge_rate REAL NOT NULL DEFAULT 0,
	invoice_prefix TEXT NOT NULL DEFAULT 'NX',
	business_hours TEXT NOT NULL DEFAULT '',
	backup_path TEXT NOT NULL DEFAULT '',
	receipt_printer_id TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS integration_settings (
	id TEXT PRIMARY KEY,
	provider TEXT NOT NULL UNIQUE,
	mode TEXT NOT NULL DEFAULT 'test',
	display_name TEXT NOT NULL DEFAULT '',
	base_url TEXT NOT NULL DEFAULT '',
	secret_envelope TEXT NOT NULL DEFAULT '',
	credential_status TEXT NOT NULL DEFAULT 'missing',
	health_status TEXT NOT NULL DEFAULT 'not_configured',
	last_checked_at TEXT NOT NULL DEFAULT '',
	last_error TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS staff (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	role TEXT NOT NULL,
	pin_hash TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'active',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_log (
	id TEXT PRIMARY KEY,
	event_type TEXT NOT NULL,
	target_type TEXT NOT NULL,
	target_id TEXT NOT NULL DEFAULT '',
	detail TEXT NOT NULL DEFAULT '',
	actor TEXT NOT NULL DEFAULT 'operator',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS menu_categories (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS menu_modifiers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	price_delta REAL NOT NULL DEFAULT 0,
	route_id TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS menu_item_modifiers (
	item_id TEXT NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
	modifier_id TEXT NOT NULL REFERENCES menu_modifiers(id) ON DELETE CASCADE,
	PRIMARY KEY (item_id, modifier_id)
);

CREATE TABLE IF NOT EXISTS floor_sections (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS dining_tables (
	id TEXT PRIMARY KEY,
	section_id TEXT NOT NULL REFERENCES floor_sections(id),
	name TEXT NOT NULL,
	seats INTEGER NOT NULL DEFAULT 2,
	status TEXT NOT NULL DEFAULT 'available',
	active_session_id TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS order_sessions (
	id TEXT PRIMARY KEY,
	table_id TEXT NOT NULL REFERENCES dining_tables(id),
	waiter_id TEXT NOT NULL DEFAULT '',
	guest_count INTEGER NOT NULL DEFAULT 1,
	service_mode TEXT NOT NULL DEFAULT 'dine_in',
	status TEXT NOT NULL DEFAULT 'open',
	invoice_id TEXT NOT NULL DEFAULT '',
	subtotal REAL NOT NULL DEFAULT 0,
	tax_total REAL NOT NULL DEFAULT 0,
	service_charge REAL NOT NULL DEFAULT 0,
	total REAL NOT NULL DEFAULT 0,
	opened_at TEXT NOT NULL,
	closed_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS order_session_lines (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES order_sessions(id) ON DELETE CASCADE,
	item_id TEXT NOT NULL REFERENCES menu_items(id),
	item_name TEXT NOT NULL,
	quantity INTEGER NOT NULL,
	unit_price REAL NOT NULL,
	modifier_total REAL NOT NULL DEFAULT 0,
	line_total REAL NOT NULL,
	notes TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'open',
	kot_status TEXT NOT NULL DEFAULT 'not_sent'
);

CREATE TABLE IF NOT EXISTS order_session_line_modifiers (
	line_id TEXT NOT NULL REFERENCES order_session_lines(id) ON DELETE CASCADE,
	modifier_id TEXT NOT NULL REFERENCES menu_modifiers(id),
	modifier_name TEXT NOT NULL,
	price_delta REAL NOT NULL DEFAULT 0,
	PRIMARY KEY (line_id, modifier_id)
);

CREATE TABLE IF NOT EXISTS order_session_events (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL REFERENCES order_sessions(id) ON DELETE CASCADE,
	type TEXT NOT NULL,
	detail TEXT NOT NULL DEFAULT '',
	actor TEXT NOT NULL DEFAULT 'operator',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS kitchen_routes (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	printer_name TEXT NOT NULL DEFAULT '',
	color TEXT NOT NULL DEFAULT '#28b09b'
);

CREATE TABLE IF NOT EXISTS kitchen_tickets (
	id TEXT PRIMARY KEY,
	sale_id TEXT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
	invoice_number TEXT NOT NULL,
	ticket_number TEXT NOT NULL,
	route_id TEXT NOT NULL,
	route_name TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS kitchen_ticket_lines (
	id TEXT PRIMARY KEY,
	ticket_id TEXT NOT NULL REFERENCES kitchen_tickets(id) ON DELETE CASCADE,
	item_id TEXT NOT NULL REFERENCES menu_items(id),
	item_name TEXT NOT NULL,
	quantity INTEGER NOT NULL,
	notes TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS invoice_events (
	id TEXT PRIMARY KEY,
	invoice_id TEXT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	detail TEXT NOT NULL DEFAULT '',
	actor TEXT NOT NULL DEFAULT 'system',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
	id TEXT PRIMARY KEY,
	invoice_id TEXT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
	method TEXT NOT NULL,
	amount REAL NOT NULL,
	tendered REAL NOT NULL DEFAULT 0,
	change_due REAL NOT NULL DEFAULT 0,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS refunds (
	id TEXT PRIMARY KEY,
	invoice_id TEXT NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
	amount REAL NOT NULL,
	reason TEXT NOT NULL DEFAULT '',
	approved_by TEXT NOT NULL DEFAULT '',
	approved_at TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS manager_approvals (
	id TEXT PRIMARY KEY,
	action TEXT NOT NULL,
	invoice_id TEXT NOT NULL DEFAULT '',
	approved_by TEXT NOT NULL DEFAULT 'Manager',
	approved_at TEXT NOT NULL,
	token TEXT NOT NULL,
	expires_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS notification_queue (
	id TEXT PRIMARY KEY,
	invoice_id TEXT NOT NULL DEFAULT '',
	channel TEXT NOT NULL,
	recipient TEXT NOT NULL DEFAULT '',
	template TEXT NOT NULL,
	payload TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	error TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	sent_at TEXT NOT NULL DEFAULT '',
	read_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS payment_requests (
	id TEXT PRIMARY KEY,
	invoice_id TEXT NOT NULL DEFAULT '',
	provider TEXT NOT NULL DEFAULT 'razorpay',
	amount REAL NOT NULL DEFAULT 0,
	currency TEXT NOT NULL DEFAULT 'INR',
	status TEXT NOT NULL DEFAULT 'draft',
	reference TEXT NOT NULL DEFAULT '',
	checkout_url TEXT NOT NULL DEFAULT '',
	qr_payload TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS print_jobs (
	id TEXT PRIMARY KEY,
	kind TEXT NOT NULL,
	reference_id TEXT NOT NULL DEFAULT '',
	printer_id TEXT NOT NULL DEFAULT '',
	target TEXT NOT NULL DEFAULT '',
	payload TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'ready',
	attempts INTEGER NOT NULL DEFAULT 0,
	last_error TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	printed_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS inventory_movements (
	id TEXT PRIMARY KEY,
	ingredient_id TEXT NOT NULL REFERENCES ingredients(id),
	type TEXT NOT NULL,
	quantity REAL NOT NULL,
	reason TEXT NOT NULL,
	reference_id TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS delivery_batches (
	id TEXT PRIMARY KEY,
	vendor_name TEXT NOT NULL,
	invoice_number TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS delivery_lines (
	id TEXT PRIMARY KEY,
	batch_id TEXT NOT NULL REFERENCES delivery_batches(id) ON DELETE CASCADE,
	ingredient_id TEXT NOT NULL REFERENCES ingredients(id),
	ordered_qty REAL NOT NULL,
	accepted_qty REAL NOT NULL,
	rejected_qty REAL NOT NULL,
	unit_cost REAL NOT NULL,
	rejection_reason TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS vendors (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	phone TEXT NOT NULL DEFAULT '',
	gstin TEXT NOT NULL DEFAULT '',
	payment_terms TEXT NOT NULL DEFAULT '',
	quality_score REAL NOT NULL DEFAULT 100,
	status TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS purchase_orders (
	id TEXT PRIMARY KEY,
	vendor_id TEXT NOT NULL REFERENCES vendors(id),
	po_number TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'draft',
	expected_date TEXT NOT NULL DEFAULT '',
	subtotal REAL NOT NULL DEFAULT 0,
	rejected_total REAL NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	closed_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS purchase_order_lines (
	id TEXT PRIMARY KEY,
	purchase_order_id TEXT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
	ingredient_id TEXT NOT NULL REFERENCES ingredients(id),
	ordered_qty REAL NOT NULL DEFAULT 0,
	accepted_qty REAL NOT NULL DEFAULT 0,
	rejected_qty REAL NOT NULL DEFAULT 0,
	unit_cost REAL NOT NULL DEFAULT 0,
	rejection_reason TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS vendor_debit_notes (
	id TEXT PRIMARY KEY,
	vendor_id TEXT NOT NULL REFERENCES vendors(id),
	purchase_order_id TEXT NOT NULL DEFAULT '',
	amount REAL NOT NULL DEFAULT 0,
	reason TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'draft',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS day_closes (
	id TEXT PRIMARY KEY,
	business_date TEXT NOT NULL UNIQUE,
	closed_at TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'closed',
	sales_total REAL NOT NULL DEFAULT 0,
	cash_expected REAL NOT NULL DEFAULT 0,
	cash_counted REAL NOT NULL DEFAULT 0,
	cash_variance REAL NOT NULL DEFAULT 0,
	upi_total REAL NOT NULL DEFAULT 0,
	card_total REAL NOT NULL DEFAULT 0,
	razorpay_total REAL NOT NULL DEFAULT 0,
	refund_total REAL NOT NULL DEFAULT 0,
	void_count INTEGER NOT NULL DEFAULT 0,
	discount_total REAL NOT NULL DEFAULT 0,
	notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS account_groups (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	system_key TEXT NOT NULL UNIQUE,
	category TEXT NOT NULL,
	normal_balance TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	is_system INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS account_ledgers (
	id TEXT PRIMARY KEY,
	group_id TEXT NOT NULL REFERENCES account_groups(id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	system_key TEXT NOT NULL UNIQUE,
	normal_balance TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	is_system INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS accounting_vouchers (
	id TEXT PRIMARY KEY,
	voucher_date TEXT NOT NULL,
	voucher_type TEXT NOT NULL,
	narration TEXT NOT NULL DEFAULT '',
	source_kind TEXT NOT NULL,
	source_id TEXT NOT NULL,
	reference_number TEXT NOT NULL DEFAULT '',
	total_amount REAL NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	UNIQUE(source_kind, source_id)
);

CREATE TABLE IF NOT EXISTS accounting_voucher_lines (
	id TEXT PRIMARY KEY,
	voucher_id TEXT NOT NULL REFERENCES accounting_vouchers(id) ON DELETE CASCADE,
	ledger_id TEXT NOT NULL REFERENCES account_ledgers(id),
	line_type TEXT NOT NULL,
	amount REAL NOT NULL DEFAULT 0,
	narration TEXT NOT NULL DEFAULT '',
	sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sync_log (
	id TEXT PRIMARY KEY,
	entity TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	operation TEXT NOT NULL,
	payload TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL,
	synced_at TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS marketing_drafts (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	channel TEXT NOT NULL,
	caption TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL
);
`

type Store struct {
	db   *sql.DB
	path string
}

func DefaultDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "NEXUS", "nexus.db"), nil
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db, path: path}
	if err := store.configure(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.seed(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) configure() error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmas {
		if _, err := s.db.Exec(pragma); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	columns := []struct {
		table      string
		name       string
		definition string
	}{
		{"ingredients", "purchase_unit", "TEXT NOT NULL DEFAULT ''"},
		{"ingredients", "purchase_to_usage", "REAL NOT NULL DEFAULT 1"},
		{"ingredients", "last_purchase_cost", "REAL NOT NULL DEFAULT 0"},
		{"menu_items", "route_id", "TEXT NOT NULL DEFAULT 'route_kitchen'"},
		{"menu_items", "tax_rate", "REAL NOT NULL DEFAULT 0.05"},
		{"sales", "invoice_number", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "table_name", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "order_type", "TEXT NOT NULL DEFAULT 'dine_in'"},
		{"sales", "subtotal", "REAL NOT NULL DEFAULT 0"},
		{"sales", "discount_total", "REAL NOT NULL DEFAULT 0"},
		{"sales", "tax_total", "REAL NOT NULL DEFAULT 0"},
		{"sales", "payment_method", "TEXT NOT NULL DEFAULT 'cash'"},
		{"sales", "payment_tendered", "REAL NOT NULL DEFAULT 0"},
		{"sales", "payment_change", "REAL NOT NULL DEFAULT 0"},
		{"sales", "source_invoice_id", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "split_group_id", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "void_reason", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "refund_reason", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "approved_by", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "approved_at", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "kot_sent_at", "TEXT NOT NULL DEFAULT ''"},
		{"sales", "closed_at", "TEXT NOT NULL DEFAULT ''"},
		{"sale_lines", "notes", "TEXT NOT NULL DEFAULT ''"},
		{"order_session_lines", "kot_status", "TEXT NOT NULL DEFAULT 'not_sent'"},
		{"notification_queue", "read_at", "TEXT NOT NULL DEFAULT ''"},
	}
	for _, column := range columns {
		if err := s.addColumnIfMissing(column.table, column.name, column.definition); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) addColumnIfMissing(table, name, definition string) error {
	rows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var columnName, columnType string
		var notNull int
		var defaultValue any
		var primaryKey int
		if err := rows.Scan(&cid, &columnName, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if columnName == name {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = s.db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, name, definition))
	return err
}

func (s *Store) seed() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM restaurants").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return s.ensureReferenceData()
	}

	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer rollback(tx)

	if _, err := tx.Exec(
		"INSERT INTO restaurants (id, name, phone, website, brand_voice, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		"rest_demo", "NEXUS Demo Cafe", "+91 98765 43210", "demo.nexus.local", "Warm, clean, confident, local-first", now,
	); err != nil {
		return err
	}

	routes := []KitchenRoute{
		{ID: "route_bar", Name: "Barista", PrinterName: "Coffee Counter", Color: "#2d7dd2"},
		{ID: "route_kitchen", Name: "Kitchen", PrinterName: "Hot Kitchen", Color: "#e84f3d"},
	}
	for _, route := range routes {
		if _, err := tx.Exec(
			"INSERT INTO kitchen_routes (id, name, printer_name, color) VALUES (?, ?, ?, ?)",
			route.ID, route.Name, route.PrinterName, route.Color,
		); err != nil {
			return err
		}
	}
	settings := map[string]string{
		"invoice_prefix":   "NX",
		"next_invoice_seq": "1",
		"kot_prefix":       "KOT",
		"next_kot_seq":     "1",
		"manager_pin":      "1234",
	}
	for key, value := range settings {
		if _, err := tx.Exec("INSERT INTO app_settings (key, value) VALUES (?, ?)", key, value); err != nil {
			return err
		}
	}

	ingredients := []Ingredient{
		{ID: "ing_beans", Name: "Arabica Beans", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 9000, ReorderPoint: 2500, WasteFactor: 0.05, LastPurchaseCost: 820, LastAuditAt: now},
		{ID: "ing_milk", Name: "Full Cream Milk", Unit: "ml", PurchaseUnit: "ltr", PurchaseToUsage: 1000, OnHandQty: 36000, ReorderPoint: 9000, WasteFactor: 0.07, LastPurchaseCost: 62, LastAuditAt: now},
		{ID: "ing_cup", Name: "12oz Paper Cup", Unit: "pcs", PurchaseUnit: "sleeve", PurchaseToUsage: 50, OnHandQty: 420, ReorderPoint: 80, WasteFactor: 0.02, LastPurchaseCost: 190, LastAuditAt: now},
		{ID: "ing_tomato", Name: "Tomato", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 14500, ReorderPoint: 4000, WasteFactor: 0.12, LastPurchaseCost: 42, LastAuditAt: now},
		{ID: "ing_chicken", Name: "Chicken Breast", Unit: "g", PurchaseUnit: "kg", PurchaseToUsage: 1000, OnHandQty: 18000, ReorderPoint: 5000, WasteFactor: 0.08, LastPurchaseCost: 280, LastAuditAt: now},
		{ID: "ing_bun", Name: "Brioche Bun", Unit: "pcs", PurchaseUnit: "tray", PurchaseToUsage: 24, OnHandQty: 160, ReorderPoint: 40, WasteFactor: 0.03, LastPurchaseCost: 360, LastAuditAt: now},
	}
	for _, ingredient := range ingredients {
		if _, err := tx.Exec(
			"INSERT INTO ingredients (id, name, unit, purchase_unit, purchase_to_usage, on_hand_qty, reorder_point, waste_factor, last_purchase_cost, last_audit_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			ingredient.ID, ingredient.Name, ingredient.Unit, ingredient.PurchaseUnit, ingredient.PurchaseToUsage, ingredient.OnHandQty, ingredient.ReorderPoint, ingredient.WasteFactor, ingredient.LastPurchaseCost, ingredient.LastAuditAt,
		); err != nil {
			return err
		}
	}

	items := []MenuItem{
		{ID: "item_latte", Name: "Classic Latte", Category: "Coffee", Price: 180, Cost: 58, Status: "active", RouteID: "route_bar", TaxRate: 0.05},
		{ID: "item_cappuccino", Name: "Cappuccino", Category: "Coffee", Price: 170, Cost: 54, Status: "active", RouteID: "route_bar", TaxRate: 0.05},
		{ID: "item_sandwich", Name: "Chicken Club Sandwich", Category: "Food", Price: 260, Cost: 112, Status: "active", RouteID: "route_kitchen", TaxRate: 0.05},
		{ID: "item_wrap", Name: "Tomato Basil Wrap", Category: "Food", Price: 220, Cost: 86, Status: "active", RouteID: "route_kitchen", TaxRate: 0.05},
	}
	for _, item := range items {
		if _, err := tx.Exec(
			"INSERT INTO menu_items (id, name, category, price, cost, status, route_id, tax_rate) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			item.ID, item.Name, item.Category, item.Price, item.Cost, item.Status, item.RouteID, item.TaxRate,
		); err != nil {
			return err
		}
	}

	recipes := []struct {
		itemID       string
		ingredientID string
		qty          float64
	}{
		{"item_latte", "ing_beans", 18},
		{"item_latte", "ing_milk", 240},
		{"item_latte", "ing_cup", 1},
		{"item_cappuccino", "ing_beans", 18},
		{"item_cappuccino", "ing_milk", 180},
		{"item_cappuccino", "ing_cup", 1},
		{"item_sandwich", "ing_chicken", 120},
		{"item_sandwich", "ing_tomato", 55},
		{"item_sandwich", "ing_bun", 1},
		{"item_wrap", "ing_tomato", 90},
		{"item_wrap", "ing_bun", 1},
	}
	for _, recipe := range recipes {
		if _, err := tx.Exec(
			"INSERT INTO recipes (id, item_id, ingredient_id, quantity) VALUES (?, ?, ?, ?)",
			mustID(), recipe.itemID, recipe.ingredientID, recipe.qty,
		); err != nil {
			return err
		}
	}

	customers := []Customer{
		{ID: "cust_rahul", Name: "Rahul Mehra", Phone: "+919999111222", TotalSpend: 3420, VisitCount: 14, LastVisitAt: now, FavoriteItem: "Classic Latte"},
		{ID: "cust_isha", Name: "Isha Kapoor", Phone: "+918888222333", TotalSpend: 1280, VisitCount: 5, LastVisitAt: now, FavoriteItem: "Tomato Basil Wrap"},
	}
	for _, customer := range customers {
		if _, err := tx.Exec(
			"INSERT INTO customers (id, name, phone, total_spend, visit_count, last_visit_at, favorite_item) VALUES (?, ?, ?, ?, ?, ?, ?)",
			customer.ID, customer.Name, customer.Phone, customer.TotalSpend, customer.VisitCount, customer.LastVisitAt, customer.FavoriteItem,
		); err != nil {
			return err
		}
	}

	drafts := []MarketingDraft{
		{
			ID:        "draft_deadstock",
			Title:     "Move tomato stock",
			Channel:   "Instagram + WhatsApp",
			Caption:   "Fresh tomato basil wraps are on the counter today. Light, fast, and made for the lunch rush.",
			Status:    "draft",
			CreatedAt: now,
		},
		{
			ID:        "draft_whales",
			Title:     "Reward regulars",
			Channel:   "WhatsApp",
			Caption:   "Your regular table misses you. Drop by this week and your coffee upgrade is on us.",
			Status:    "draft",
			CreatedAt: now,
		},
	}
	for _, draft := range drafts {
		if _, err := tx.Exec(
			"INSERT INTO marketing_drafts (id, title, channel, caption, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			draft.ID, draft.Title, draft.Channel, draft.Caption, draft.Status, draft.CreatedAt,
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return s.ensureReferenceData()
}

func (s *Store) ensureReferenceData() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer rollback(tx)

	routes := []KitchenRoute{
		{ID: "route_bar", Name: "Barista", PrinterName: "Coffee Counter", Color: "#2d7dd2"},
		{ID: "route_kitchen", Name: "Kitchen", PrinterName: "Hot Kitchen", Color: "#e84f3d"},
	}
	for _, route := range routes {
		if _, err := tx.Exec(`
			INSERT INTO kitchen_routes (id, name, printer_name, color)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				printer_name = excluded.printer_name,
				color = excluded.color
		`, route.ID, route.Name, route.PrinterName, route.Color); err != nil {
			return err
		}
	}

	settings := map[string]string{
		"invoice_prefix":   "NX",
		"next_invoice_seq": "1",
		"kot_prefix":       "KOT",
		"next_kot_seq":     "1",
		"manager_pin":      "1234",
	}
	for key, value := range settings {
		if _, err := tx.Exec("INSERT OR IGNORE INTO app_settings (key, value) VALUES (?, ?)", key, value); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`
		UPDATE menu_items
		SET route_id = CASE WHEN category = 'Coffee' THEN 'route_bar' ELSE 'route_kitchen' END
		WHERE route_id = '' OR route_id IS NULL
	`); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE menu_items SET tax_rate = 0.05 WHERE tax_rate = 0"); err != nil {
		return err
	}

	unitDefaults := map[string]struct {
		purchaseUnit    string
		purchaseToUsage float64
		cost            float64
	}{
		"ing_beans":   {"kg", 1000, 820},
		"ing_milk":    {"ltr", 1000, 62},
		"ing_cup":     {"sleeve", 50, 190},
		"ing_tomato":  {"kg", 1000, 42},
		"ing_chicken": {"kg", 1000, 280},
		"ing_bun":     {"tray", 24, 360},
	}
	for id, defaults := range unitDefaults {
		if _, err := tx.Exec(`
			UPDATE ingredients
			SET
				purchase_unit = CASE WHEN purchase_unit = '' THEN ? ELSE purchase_unit END,
				purchase_to_usage = CASE WHEN purchase_to_usage <= 0 THEN ? ELSE purchase_to_usage END,
				last_purchase_cost = CASE WHEN last_purchase_cost <= 0 THEN ? ELSE last_purchase_cost END
			WHERE id = ?
		`, defaults.purchaseUnit, defaults.purchaseToUsage, defaults.cost, id); err != nil {
			return err
		}
	}

	if err := s.ensurePilotReferenceData(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) Dashboard() (Dashboard, error) {
	var dashboard Dashboard

	restaurant, err := s.restaurant()
	if err != nil {
		return dashboard, err
	}
	ingredients, err := s.ingredients()
	if err != nil {
		return dashboard, err
	}
	items, err := s.menuItems()
	if err != nil {
		return dashboard, err
	}
	routes, err := s.kitchenRoutes()
	if err != nil {
		return dashboard, err
	}
	recipes, err := s.recipes()
	if err != nil {
		return dashboard, err
	}
	customers, err := s.customers()
	if err != nil {
		return dashboard, err
	}
	recentSales, err := s.recentSales()
	if err != nil {
		return dashboard, err
	}
	tickets, err := s.kitchenTickets()
	if err != nil {
		return dashboard, err
	}
	deliveries, err := s.deliveries()
	if err != nil {
		return dashboard, err
	}
	drafts, err := s.marketingDrafts()
	if err != nil {
		return dashboard, err
	}
	metrics, err := s.metrics(ingredients)
	if err != nil {
		return dashboard, err
	}

	dashboard = Dashboard{
		Restaurant:      restaurant,
		Metrics:         metrics,
		MenuItems:       items,
		Ingredients:     ingredients,
		KitchenRoutes:   routes,
		Recipes:         recipes,
		Customers:       customers,
		RecentSales:     recentSales,
		KitchenTickets:  tickets,
		Deliveries:      deliveries,
		Signals:         s.signals(ingredients, customers, deliveries, items),
		MarketingDrafts: drafts,
	}

	return dashboard, nil
}

func (s *Store) recordSaleLegacyUnused(input SaleInput) (Dashboard, error) {
	if len(input.Lines) == 0 {
		return Dashboard{}, errors.New("sale requires at least one line")
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

	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	saleID := mustID()
	createdAt := nowUTC()
	invoiceNumber, err := nextDocumentNumber(tx, "invoice_prefix", "next_invoice_seq")
	if err != nil {
		return Dashboard{}, err
	}

	type saleLine struct {
		item     MenuItem
		quantity int
		notes    string
		subtotal float64
	}

	lines := make([]saleLine, 0, len(input.Lines))
	subtotal := 0.0
	favoriteItem := ""

	for _, line := range input.Lines {
		if line.Quantity <= 0 {
			return Dashboard{}, errors.New("sale quantity must be positive")
		}

		item, err := menuItemByID(tx, line.ItemID)
		if err != nil {
			return Dashboard{}, err
		}
		if item.Status != "active" {
			return Dashboard{}, fmt.Errorf("%s is not active", item.Name)
		}

		lineTotal := item.Price * float64(line.Quantity)
		subtotal += lineTotal
		if favoriteItem == "" {
			favoriteItem = item.Name
		}
		lines = append(lines, saleLine{
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
		taxable := line.subtotal - (discountTotal * share)
		taxRate := line.item.TaxRate
		if input.TaxRate > 0 {
			taxRate = input.TaxRate
		}
		taxTotal += taxable * taxRate
	}
	total := round2(subtotal - discountTotal + taxTotal)
	discountTotal = round2(discountTotal)
	taxTotal = round2(taxTotal)
	tendered := input.PaymentTendered
	if tendered <= 0 || strings.EqualFold(input.PaymentMethod, "upi") || strings.EqualFold(input.PaymentMethod, "card") {
		tendered = total
	}
	changeDue := math.Max(0, tendered-total)

	if _, err := tx.Exec(
		`INSERT INTO sales (
			id, invoice_number, customer_name, customer_phone, table_name, order_type,
			subtotal, discount_total, tax_total, total, channel,
			payment_method, payment_tendered, payment_change, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		saleID, invoiceNumber, strings.TrimSpace(input.CustomerName), normalizePhone(input.CustomerPhone), strings.TrimSpace(input.TableName), input.OrderType,
		subtotal, discountTotal, taxTotal, total, input.Channel, input.PaymentMethod, tendered, changeDue, "open", createdAt,
	); err != nil {
		return Dashboard{}, err
	}

	ticketIDs := map[string]string{}
	for _, line := range lines {
		if _, err := tx.Exec(
			"INSERT INTO sale_lines (id, sale_id, item_id, item_name, quantity, unit_price, line_total, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			mustID(), saleID, line.item.ID, line.item.Name, line.quantity, line.item.Price, line.subtotal, line.notes,
		); err != nil {
			return Dashboard{}, err
		}

		components, err := recipeComponents(tx, line.item.ID)
		if err != nil {
			return Dashboard{}, err
		}
		for _, component := range components {
			deduction := component.quantity * float64(line.quantity) * (1 + component.wasteFactor)
			if _, err := tx.Exec(
				"UPDATE ingredients SET on_hand_qty = on_hand_qty - ? WHERE id = ?",
				deduction, component.ingredientID,
			); err != nil {
				return Dashboard{}, err
			}
			if _, err := tx.Exec(
				"INSERT INTO inventory_movements (id, ingredient_id, type, quantity, reason, reference_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mustID(), component.ingredientID, "sale_consumption", -deduction, fmt.Sprintf("%s x%d", line.item.Name, line.quantity), saleID, createdAt,
			); err != nil {
				return Dashboard{}, err
			}
		}

		routeID := line.item.RouteID
		if routeID == "" {
			routeID = "route_kitchen"
		}
		ticketID := ticketIDs[routeID]
		if ticketID == "" {
			routeName, err := routeNameByID(tx, routeID)
			if err != nil {
				return Dashboard{}, err
			}
			ticketNumber, err := nextDocumentNumber(tx, "kot_prefix", "next_kot_seq")
			if err != nil {
				return Dashboard{}, err
			}
			ticketID = mustID()
			ticketIDs[routeID] = ticketID
			if _, err := tx.Exec(
				"INSERT INTO kitchen_tickets (id, sale_id, invoice_number, ticket_number, route_id, route_name, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				ticketID, saleID, invoiceNumber, ticketNumber, routeID, routeName, "open", createdAt,
			); err != nil {
				return Dashboard{}, err
			}
		}
		if _, err := tx.Exec(
			"INSERT INTO kitchen_ticket_lines (id, ticket_id, item_id, item_name, quantity, notes, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
			mustID(), ticketID, line.item.ID, line.item.Name, line.quantity, line.notes, "queued",
		); err != nil {
			return Dashboard{}, err
		}
	}

	if _, err := tx.Exec(
		"UPDATE sales SET status = ? WHERE id = ?",
		"closed", saleID,
	); err != nil {
		return Dashboard{}, err
	}

	if phone := normalizePhone(input.CustomerPhone); phone != "" {
		name := strings.TrimSpace(input.CustomerName)
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
		`, mustID(), name, phone, total, createdAt, favoriteItem); err != nil {
			return Dashboard{}, err
		}
	}

	payload := map[string]any{
		"invoiceNumber": invoiceNumber,
		"sale":          input,
		"subtotal":      subtotal,
		"discountTotal": discountTotal,
		"taxTotal":      taxTotal,
		"total":         total,
	}
	if err := insertSyncLog(tx, "sale", saleID, "create", payload, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) ReceiveDelivery(input DeliveryInput) (Dashboard, error) {
	if strings.TrimSpace(input.VendorName) == "" {
		return Dashboard{}, errors.New("vendor name is required")
	}
	if len(input.Lines) == 0 {
		return Dashboard{}, errors.New("delivery requires at least one line")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	batchID := mustID()
	createdAt := nowUTC()
	status := "accepted"
	for _, line := range input.Lines {
		if line.AcceptedQty < 0 || line.RejectedQty < 0 || line.OrderedQty < 0 {
			return Dashboard{}, errors.New("delivery quantities cannot be negative")
		}
		if line.RejectedQty > 0 {
			status = "partial_rejected"
		}
		if line.AcceptedQty == 0 && line.RejectedQty == 0 {
			status = "staged"
		}
	}

	if _, err := tx.Exec(
		"INSERT INTO delivery_batches (id, vendor_name, invoice_number, status, created_at) VALUES (?, ?, ?, ?, ?)",
		batchID, strings.TrimSpace(input.VendorName), strings.TrimSpace(input.InvoiceNumber), status, createdAt,
	); err != nil {
		return Dashboard{}, err
	}

	for _, line := range input.Lines {
		lineStatus := "accepted"
		if line.RejectedQty > 0 && line.AcceptedQty > 0 {
			lineStatus = "partial_rejected"
		} else if line.RejectedQty > 0 {
			lineStatus = "rejected"
		} else if line.AcceptedQty == 0 {
			lineStatus = "staged"
		}

		if _, err := tx.Exec(
			"INSERT INTO delivery_lines (id, batch_id, ingredient_id, ordered_qty, accepted_qty, rejected_qty, unit_cost, rejection_reason, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			mustID(), batchID, line.IngredientID, line.OrderedQty, line.AcceptedQty, line.RejectedQty, line.UnitCost, strings.TrimSpace(line.RejectionReason), lineStatus,
		); err != nil {
			return Dashboard{}, err
		}

		if line.AcceptedQty > 0 {
			if _, err := tx.Exec(
				"UPDATE ingredients SET on_hand_qty = on_hand_qty + ? WHERE id = ?",
				line.AcceptedQty, line.IngredientID,
			); err != nil {
				return Dashboard{}, err
			}
			if line.UnitCost > 0 {
				if _, err := tx.Exec(
					"UPDATE ingredients SET last_purchase_cost = ? WHERE id = ?",
					line.UnitCost, line.IngredientID,
				); err != nil {
					return Dashboard{}, err
				}
			}
			if _, err := tx.Exec(
				"INSERT INTO inventory_movements (id, ingredient_id, type, quantity, reason, reference_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mustID(), line.IngredientID, "delivery_acceptance", line.AcceptedQty, "Verified delivery stock", batchID, createdAt,
			); err != nil {
				return Dashboard{}, err
			}
		}

		if line.RejectedQty > 0 {
			reason := strings.TrimSpace(line.RejectionReason)
			if reason == "" {
				reason = "Rejected during receiving"
			}
			if _, err := tx.Exec(
				"INSERT INTO inventory_movements (id, ingredient_id, type, quantity, reason, reference_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				mustID(), line.IngredientID, "delivery_rejection", 0, reason, batchID, createdAt,
			); err != nil {
				return Dashboard{}, err
			}
		}
	}

	if err := insertSyncLog(tx, "delivery_batch", batchID, "create", input, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) ReconcileStock(input ReconcileInput) (Dashboard, error) {
	if input.PhysicalQty < 0 {
		return Dashboard{}, errors.New("physical quantity cannot be negative")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	var current, wasteFactor float64
	var lastAuditAt string
	if err := tx.QueryRow(
		"SELECT on_hand_qty, waste_factor, last_audit_at FROM ingredients WHERE id = ?",
		input.IngredientID,
	).Scan(&current, &wasteFactor, &lastAuditAt); err != nil {
		return Dashboard{}, err
	}

	var consumed sql.NullFloat64
	if err := tx.QueryRow(`
		SELECT ABS(COALESCE(SUM(quantity), 0))
		FROM inventory_movements
		WHERE ingredient_id = ?
			AND type = 'sale_consumption'
			AND created_at >= ?
	`, input.IngredientID, lastAuditAt).Scan(&consumed); err != nil {
		return Dashboard{}, err
	}

	variance := current - input.PhysicalQty
	adjustedWaste := wasteFactor
	if variance > 0 && consumed.Valid && consumed.Float64 > 0 {
		adjustedWaste = clamp(wasteFactor+(variance/consumed.Float64)*0.35, 0, 0.35)
	} else if variance < 0 && consumed.Valid && consumed.Float64 > 0 {
		adjustedWaste = clamp(wasteFactor+(variance/consumed.Float64)*0.2, 0, 0.35)
	}

	referenceID := mustID()
	createdAt := nowUTC()
	note := strings.TrimSpace(input.Note)
	if note == "" {
		note = "Physical audit reconciliation"
	}

	if _, err := tx.Exec(
		"UPDATE ingredients SET on_hand_qty = ?, waste_factor = ?, last_audit_at = ? WHERE id = ?",
		input.PhysicalQty, round4(adjustedWaste), createdAt, input.IngredientID,
	); err != nil {
		return Dashboard{}, err
	}
	if _, err := tx.Exec(
		"INSERT INTO inventory_movements (id, ingredient_id, type, quantity, reason, reference_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		referenceID, input.IngredientID, "audit_adjustment", -variance, note, referenceID, createdAt,
	); err != nil {
		return Dashboard{}, err
	}
	if err := insertSyncLog(tx, "inventory_reconciliation", referenceID, "create", input, createdAt); err != nil {
		return Dashboard{}, err
	}

	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) restaurant() (Restaurant, error) {
	var r Restaurant
	err := s.db.QueryRow("SELECT id, name, phone, website, brand_voice, created_at FROM restaurants LIMIT 1").
		Scan(&r.ID, &r.Name, &r.Phone, &r.Website, &r.BrandVoice, &r.CreatedAt)
	return r, err
}

func (s *Store) ingredients() ([]Ingredient, error) {
	rows, err := s.db.Query(`
		SELECT id, name, unit, purchase_unit, purchase_to_usage, on_hand_qty, reorder_point, waste_factor, last_purchase_cost, last_audit_at
		FROM ingredients
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ingredients := make([]Ingredient, 0)
	for rows.Next() {
		var ingredient Ingredient
		if err := rows.Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.Unit,
			&ingredient.PurchaseUnit,
			&ingredient.PurchaseToUsage,
			&ingredient.OnHandQty,
			&ingredient.ReorderPoint,
			&ingredient.WasteFactor,
			&ingredient.LastPurchaseCost,
			&ingredient.LastAuditAt,
		); err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}
	return ingredients, rows.Err()
}

func (s *Store) menuItems() ([]MenuItem, error) {
	rows, err := s.db.Query(`
		SELECT m.id, m.name, m.category, m.price, m.cost, m.status, m.route_id, COALESCE(r.name, 'Kitchen'), m.tax_rate
		FROM menu_items m
		LEFT JOIN kitchen_routes r ON r.id = m.route_id
		ORDER BY m.category, m.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MenuItem, 0)
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.Price, &item.Cost, &item.Status, &item.RouteID, &item.RouteName, &item.TaxRate); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) kitchenRoutes() ([]KitchenRoute, error) {
	rows, err := s.db.Query("SELECT id, name, printer_name, color FROM kitchen_routes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	routes := make([]KitchenRoute, 0)
	for rows.Next() {
		var route KitchenRoute
		if err := rows.Scan(&route.ID, &route.Name, &route.PrinterName, &route.Color); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}

func (s *Store) recipes() ([]RecipeCard, error) {
	items, err := s.menuItems()
	if err != nil {
		return nil, err
	}

	cards := make([]RecipeCard, 0, len(items))
	for _, item := range items {
		rows, err := s.db.Query(`
			SELECT r.item_id, r.ingredient_id, i.name, i.unit, i.purchase_unit, i.purchase_to_usage, r.quantity, r.waste_factor_override
			FROM recipes r
			JOIN ingredients i ON i.id = r.ingredient_id
			WHERE r.item_id = ?
			ORDER BY i.name
		`, item.ID)
		if err != nil {
			return nil, err
		}

		card := RecipeCard{ItemID: item.ID, ItemName: item.Name, RouteName: item.RouteName, Components: make([]RecipeComponent, 0)}
		for rows.Next() {
			var component RecipeComponent
			var override sql.NullFloat64
			if err := rows.Scan(
				&component.ItemID,
				&component.IngredientID,
				&component.IngredientName,
				&component.Unit,
				&component.PurchaseUnit,
				&component.PurchaseToUsage,
				&component.Quantity,
				&override,
			); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if override.Valid {
				value := override.Float64
				component.WasteFactorOverride = &value
			}
			card.Components = append(card.Components, component)
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (s *Store) customers() ([]Customer, error) {
	rows, err := s.db.Query("SELECT id, name, phone, total_spend, visit_count, last_visit_at, favorite_item FROM customers ORDER BY total_spend DESC LIMIT 8")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := make([]Customer, 0)
	for rows.Next() {
		var customer Customer
		if err := rows.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.TotalSpend, &customer.VisitCount, &customer.LastVisitAt, &customer.FavoriteItem); err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	return customers, rows.Err()
}

func (s *Store) recentSales() ([]SaleSummary, error) {
	invoices, err := s.GetInvoices(InvoiceFilter{})
	if err != nil {
		return nil, err
	}
	if len(invoices) > 8 {
		invoices = invoices[:8]
	}
	return invoices, nil
}

func (s *Store) kitchenTickets() ([]KitchenTicket, error) {
	rows, err := s.db.Query(`
		SELECT kt.id, kt.sale_id, kt.invoice_number, kt.ticket_number, kt.route_id, kt.route_name, kt.status, kt.created_at
		FROM kitchen_tickets kt
		ORDER BY kt.created_at DESC
		LIMIT 8
	`)
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
		lines, err := s.kitchenTicketLines(ticket.ID)
		if err != nil {
			return nil, err
		}
		ticket.Lines = lines
		tickets = append(tickets, ticket)
	}
	return tickets, rows.Err()
}

func (s *Store) kitchenTicketLines(ticketID string) ([]KitchenTicketLine, error) {
	rows, err := s.db.Query(`
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

func (s *Store) deliveries() ([]DeliveryBatch, error) {
	rows, err := s.db.Query("SELECT id, vendor_name, invoice_number, status, created_at FROM delivery_batches ORDER BY created_at DESC LIMIT 6")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := make([]DeliveryBatch, 0)
	for rows.Next() {
		var batch DeliveryBatch
		if err := rows.Scan(&batch.ID, &batch.VendorName, &batch.InvoiceNumber, &batch.Status, &batch.CreatedAt); err != nil {
			return nil, err
		}
		lines, err := s.deliveryLines(batch.ID)
		if err != nil {
			return nil, err
		}
		batch.Lines = lines
		batches = append(batches, batch)
	}
	return batches, rows.Err()
}

func (s *Store) deliveryLines(batchID string) ([]DeliveryLine, error) {
	rows, err := s.db.Query(`
		SELECT dl.id, dl.batch_id, dl.ingredient_id, i.name, i.unit, dl.ordered_qty, dl.accepted_qty, dl.rejected_qty, dl.unit_cost, dl.rejection_reason, dl.status
		FROM delivery_lines dl
		JOIN ingredients i ON i.id = dl.ingredient_id
		WHERE dl.batch_id = ?
		ORDER BY i.name
	`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]DeliveryLine, 0)
	for rows.Next() {
		var line DeliveryLine
		if err := rows.Scan(&line.ID, &line.BatchID, &line.IngredientID, &line.IngredientName, &line.Unit, &line.OrderedQty, &line.AcceptedQty, &line.RejectedQty, &line.UnitCost, &line.RejectionReason, &line.Status); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func (s *Store) marketingDrafts() ([]MarketingDraft, error) {
	rows, err := s.db.Query("SELECT id, title, channel, caption, status, created_at FROM marketing_drafts ORDER BY created_at DESC LIMIT 4")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	drafts := make([]MarketingDraft, 0)
	for rows.Next() {
		var draft MarketingDraft
		if err := rows.Scan(&draft.ID, &draft.Title, &draft.Channel, &draft.Caption, &draft.Status, &draft.CreatedAt); err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}
	return drafts, rows.Err()
}

func (s *Store) metrics(ingredients []Ingredient) (Metrics, error) {
	start := time.Now().UTC().Truncate(24 * time.Hour).Format(time.RFC3339)
	var metrics Metrics
	if err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM sales
		WHERE created_at >= ?
			AND status NOT IN ('voided', 'split')
	`, start).Scan(&metrics.OrdersToday); err != nil {
		return metrics, err
	}
	if err := s.db.QueryRow(`
		SELECT COALESCE(SUM(total), 0), COALESCE(AVG(total), 0)
		FROM sales
		WHERE created_at >= ?
			AND status IN ('closed', 'partially_refunded', 'refunded')
	`, start).Scan(&metrics.SalesToday, &metrics.AverageOrder); err != nil {
		return metrics, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM sync_log WHERE status = 'pending'").Scan(&metrics.PendingSyncItems); err != nil {
		return metrics, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM customers").Scan(&metrics.CustomerCount); err != nil {
		return metrics, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM kitchen_tickets WHERE status IN ('open', 'queued', 'preparing', 'ready')").Scan(&metrics.OpenKOTs); err != nil {
		return metrics, err
	}
	for _, ingredient := range ingredients {
		if ingredient.OnHandQty <= ingredient.ReorderPoint {
			metrics.StockAlerts++
		}
	}
	return metrics, nil
}

type componentRow struct {
	ingredientID string
	quantity     float64
	wasteFactor  float64
}

func menuItemByID(tx *sql.Tx, itemID string) (MenuItem, error) {
	var item MenuItem
	err := tx.QueryRow(`
		SELECT m.id, m.name, m.category, m.price, m.cost, m.status, m.route_id, COALESCE(r.name, 'Kitchen'), m.tax_rate
		FROM menu_items m
		LEFT JOIN kitchen_routes r ON r.id = m.route_id
		WHERE m.id = ?
	`, itemID).
		Scan(&item.ID, &item.Name, &item.Category, &item.Price, &item.Cost, &item.Status, &item.RouteID, &item.RouteName, &item.TaxRate)
	if err == sql.ErrNoRows {
		return item, fmt.Errorf("menu item not found: %s", itemID)
	}
	return item, err
}

func recipeComponents(tx *sql.Tx, itemID string) ([]componentRow, error) {
	rows, err := tx.Query(`
		SELECT r.ingredient_id, r.quantity, COALESCE(r.waste_factor_override, i.waste_factor)
		FROM recipes r
		JOIN ingredients i ON i.id = r.ingredient_id
		WHERE r.item_id = ?
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var components []componentRow
	for rows.Next() {
		var component componentRow
		if err := rows.Scan(&component.ingredientID, &component.quantity, &component.wasteFactor); err != nil {
			return nil, err
		}
		components = append(components, component)
	}
	return components, rows.Err()
}

func routeNameByID(tx *sql.Tx, routeID string) (string, error) {
	var name string
	if err := tx.QueryRow("SELECT name FROM kitchen_routes WHERE id = ?", routeID).Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			return "Kitchen", nil
		}
		return "", err
	}
	return name, nil
}

func nextDocumentNumber(tx *sql.Tx, prefixKey, seqKey string) (string, error) {
	var prefix string
	if err := tx.QueryRow("SELECT value FROM app_settings WHERE key = ?", prefixKey).Scan(&prefix); err != nil {
		return "", err
	}
	var seq int
	if err := tx.QueryRow("SELECT value FROM app_settings WHERE key = ?", seqKey).Scan(&seq); err != nil {
		return "", err
	}
	if _, err := tx.Exec("UPDATE app_settings SET value = ? WHERE key = ?", fmt.Sprintf("%d", seq+1), seqKey); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%05d", prefix, seq), nil
}

func calculateDiscount(subtotal float64, discountType string, discountValue float64) float64 {
	if subtotal <= 0 || discountValue <= 0 {
		return 0
	}
	switch strings.ToLower(strings.TrimSpace(discountType)) {
	case "percent", "percentage":
		return math.Min(subtotal, subtotal*(discountValue/100))
	case "fixed", "amount":
		return math.Min(subtotal, discountValue)
	default:
		return 0
	}
}

func insertSyncLog(tx *sql.Tx, entity string, entityID string, operation string, payload any, createdAt string) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"INSERT INTO sync_log (id, entity, entity_id, operation, payload, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		mustID(), entity, entityID, operation, string(encoded), "pending", createdAt,
	)
	return err
}

func rollback(tx *sql.Tx) {
	if tx != nil {
		_ = tx.Rollback()
	}
}

func mustID() string {
	id, err := uuid.NewV7()
	if err == nil {
		return id.String()
	}
	return uuid.NewString()
}

func normalizePhone(phone string) string {
	return strings.Join(strings.Fields(phone), "")
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func clamp(value, min, max float64) float64 {
	return math.Min(max, math.Max(min, value))
}

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
