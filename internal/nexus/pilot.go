package nexus

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

func (s *Store) ensurePilotReferenceData(tx *sql.Tx) error {
	now := nowUTC()
	if _, err := tx.Exec(`
		INSERT INTO restaurant_settings (
			id, restaurant_name, gstin, legal_name, address, state, tax_mode, default_tax_rate,
			service_charge_rate, invoice_prefix, business_hours, backup_path, receipt_printer_id, updated_at
		)
		VALUES ('primary', 'NEXUS Demo Cafe', '29ABCDE1234F1Z5', 'NEXUS Demo Hospitality LLP',
			'12 Brigade Road, Bengaluru', 'Karnataka', 'gst', 0.05, 0.05, 'NX',
			'10:00-23:00', '', 'printer_receipt', ?)
		ON CONFLICT(id) DO NOTHING
	`, now); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO app_settings (key, value) VALUES
			('next_po_seq', '1'),
			('manager_pin_salt', 'nexus-demo'),
			('manager_pin_hash', ?)
		ON CONFLICT(key) DO NOTHING
	`, hashPIN("1234", "nexus-demo")); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO roles (id, name, can_manage_settings, can_void, can_refund, can_close_day, created_at)
		VALUES
			('role_manager', 'manager', 1, 1, 1, 1, ?),
			('role_cashier', 'cashier', 0, 0, 0, 1, ?),
			('role_waiter', 'waiter', 0, 0, 0, 0, ?)
		ON CONFLICT(id) DO NOTHING
	`, now, now, now); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO integration_settings (id, provider, mode, display_name, base_url, credential_status, health_status, updated_at)
		VALUES
			('int_meta', 'meta_whatsapp', 'test', 'Meta Cloud API', 'https://graph.facebook.com', 'missing', 'needs_credentials', ?),
			('int_razorpay', 'razorpay', 'test', 'Razorpay', 'https://api.razorpay.com/v1', 'missing', 'needs_credentials', ?),
			('int_printer', 'escpos_printer', 'local', 'ESC/POS Printer', 'tcp://192.168.1.50:9100', 'not_required', 'ready_for_test', ?)
		ON CONFLICT(provider) DO NOTHING
	`, now, now, now); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO staff (id, name, role, pin_hash, status, created_at)
		VALUES
			('staff_manager', 'Ananya Manager', 'manager', ?, 'active', ?),
			('staff_waiter', 'Ravi Waiter', 'waiter', '', 'active', ?),
			('staff_cashier', 'Neha Cashier', 'cashier', '', 'active', ?)
		ON CONFLICT(id) DO NOTHING
	`, hashPIN("1234", "staff_manager"), now, now, now); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO menu_categories (id, name, sort_order, status)
		VALUES
			('cat_coffee', 'Coffee', 10, 'active'),
			('cat_food', 'Food', 20, 'active'),
			('cat_dessert', 'Dessert', 30, 'active')
		ON CONFLICT(id) DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO menu_modifiers (id, name, price_delta, route_id, status)
		VALUES
			('mod_extra_shot', 'Extra Shot', 45, 'route_bar', 'active'),
			('mod_oat_milk', 'Oat Milk', 60, 'route_bar', 'active'),
			('mod_extra_cheese', 'Extra Cheese', 55, 'route_kitchen', 'active'),
			('mod_no_onion', 'No Onion', 0, 'route_kitchen', 'active')
		ON CONFLICT(id) DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO menu_item_modifiers (item_id, modifier_id)
		VALUES
			('item_latte', 'mod_extra_shot'),
			('item_latte', 'mod_oat_milk'),
			('item_cappuccino', 'mod_extra_shot'),
			('item_sandwich', 'mod_extra_cheese'),
			('item_sandwich', 'mod_no_onion'),
			('item_wrap', 'mod_no_onion')
		ON CONFLICT(item_id, modifier_id) DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO floor_sections (id, name, sort_order)
		VALUES ('sec_main', 'Main Dining', 10), ('sec_patio', 'Patio', 20)
		ON CONFLICT(id) DO NOTHING
	`); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO dining_tables (id, section_id, name, seats, status)
		VALUES
			('tbl_01', 'sec_main', 'T-01', 2, 'available'),
			('tbl_02', 'sec_main', 'T-02', 4, 'available'),
			('tbl_03', 'sec_main', 'T-03', 4, 'available'),
			('tbl_p1', 'sec_patio', 'P-01', 2, 'available')
		ON CONFLICT(id) DO NOTHING
	`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO vendors (id, name, phone, gstin, payment_terms, quality_score, status)
		VALUES
			('vendor_fresh_farm', 'Fresh Farm Co', '+919876500001', '29FRESH1234F1Z5', '7 days', 92, 'active'),
			('vendor_dairy', 'Morning Dairy', '+919876500002', '29DAIRY1234F1Z5', 'Cash on delivery', 96, 'active')
		ON CONFLICT(id) DO NOTHING
	`); err != nil {
		return err
	}

	return s.seedAccountingDefaultsTx(tx)
}

func (s *Store) GetPilotWorkspace() (PilotWorkspace, error) {
	var workspace PilotWorkspace
	var err error
	if workspace.Settings, err = s.restaurantSettings(); err != nil {
		return workspace, err
	}
	if workspace.Integrations, err = s.integrationSettings(); err != nil {
		return workspace, err
	}
	if workspace.Staff, err = s.staffMembers(); err != nil {
		return workspace, err
	}
	if workspace.Categories, err = s.menuCategories(); err != nil {
		return workspace, err
	}
	if workspace.Modifiers, err = s.menuModifiers(); err != nil {
		return workspace, err
	}
	if workspace.ItemModifiers, err = s.menuItemModifiers(); err != nil {
		return workspace, err
	}
	if workspace.FloorSections, err = s.floorSections(); err != nil {
		return workspace, err
	}
	if workspace.Tables, err = s.diningTables(); err != nil {
		return workspace, err
	}
	if workspace.OrderSessions, err = s.orderSessions(); err != nil {
		return workspace, err
	}
	if workspace.PaymentRequests, err = s.paymentRequests(); err != nil {
		return workspace, err
	}
	if workspace.PrintJobs, err = s.printJobs(); err != nil {
		return workspace, err
	}
	if workspace.Vendors, err = s.vendors(); err != nil {
		return workspace, err
	}
	if workspace.PurchaseOrders, err = s.purchaseOrders(); err != nil {
		return workspace, err
	}
	if workspace.DebitNotes, err = s.vendorDebitNotes(); err != nil {
		return workspace, err
	}
	if workspace.DayClose, err = s.dayClosePreview(todayDate()); err != nil {
		return workspace, err
	}
	if workspace.Accounting, err = s.GetAccountingSnapshot(); err != nil {
		return workspace, err
	}
	if workspace.AuditLog, err = s.GetAuditLog(40); err != nil {
		return workspace, err
	}
	return workspace, nil
}

func (s *Store) SaveRestaurantSettings(input RestaurantSettings) (PilotWorkspace, error) {
	now := nowUTC()
	if strings.TrimSpace(input.RestaurantName) == "" {
		return PilotWorkspace{}, errors.New("restaurant name is required")
	}
	if input.DefaultTaxRate < 0 || input.ServiceChargeRate < 0 {
		return PilotWorkspace{}, errors.New("tax and service charge rates cannot be negative")
	}
	if input.InvoicePrefix == "" {
		input.InvoicePrefix = "NX"
	}
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	if _, err := tx.Exec(`
		INSERT INTO restaurant_settings (
			id, restaurant_name, gstin, legal_name, address, state, tax_mode, default_tax_rate,
			service_charge_rate, invoice_prefix, business_hours, backup_path, receipt_printer_id, updated_at
		)
		VALUES ('primary', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			restaurant_name = excluded.restaurant_name,
			gstin = excluded.gstin,
			legal_name = excluded.legal_name,
			address = excluded.address,
			state = excluded.state,
			tax_mode = excluded.tax_mode,
			default_tax_rate = excluded.default_tax_rate,
			service_charge_rate = excluded.service_charge_rate,
			invoice_prefix = excluded.invoice_prefix,
			business_hours = excluded.business_hours,
			backup_path = excluded.backup_path,
			receipt_printer_id = excluded.receipt_printer_id,
			updated_at = excluded.updated_at
	`, input.RestaurantName, strings.TrimSpace(input.GSTIN), strings.TrimSpace(input.LegalName), strings.TrimSpace(input.Address), strings.TrimSpace(input.State),
		firstNonEmpty(strings.TrimSpace(input.TaxMode), "gst"), input.DefaultTaxRate, input.ServiceChargeRate, strings.TrimSpace(input.InvoicePrefix),
		strings.TrimSpace(input.BusinessHours), strings.TrimSpace(input.BackupPath), strings.TrimSpace(input.ReceiptPrinterID), now); err != nil {
		return PilotWorkspace{}, err
	}
	if _, err := tx.Exec("UPDATE app_settings SET value = ? WHERE key = 'invoice_prefix'", input.InvoicePrefix); err != nil {
		return PilotWorkspace{}, err
	}
	if err := insertAuditLogTx(tx, "settings_updated", "restaurant_settings", "primary", "Restaurant settings saved", now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := insertSyncLog(tx, "restaurant_settings", "primary", "upsert", input, now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) SaveIntegrationSetting(input IntegrationSettingInput) (PilotWorkspace, error) {
	provider := strings.TrimSpace(input.Provider)
	if provider == "" {
		return PilotWorkspace{}, errors.New("provider is required")
	}
	now := nowUTC()
	envelope := ""
	status := "missing"
	if strings.TrimSpace(input.Secret) != "" {
		envelope = sealLocalSecret(provider, input.Secret)
		status = "stored"
	}
	if input.HealthStatus == "" {
		input.HealthStatus = "not_checked"
	}
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	if _, err := tx.Exec(`
		INSERT INTO integration_settings (id, provider, mode, display_name, base_url, secret_envelope, credential_status, health_status, last_checked_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(provider) DO UPDATE SET
			mode = excluded.mode,
			display_name = excluded.display_name,
			base_url = excluded.base_url,
			secret_envelope = CASE WHEN excluded.secret_envelope = '' THEN integration_settings.secret_envelope ELSE excluded.secret_envelope END,
			credential_status = CASE WHEN excluded.secret_envelope = '' THEN integration_settings.credential_status ELSE excluded.credential_status END,
			health_status = excluded.health_status,
			last_checked_at = excluded.last_checked_at,
			updated_at = excluded.updated_at
	`, "int_"+provider, provider, firstNonEmpty(input.Mode, "test"), input.DisplayName, input.BaseURL, envelope, status, input.HealthStatus, now, now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := insertAuditLogTx(tx, "integration_saved", "integration_settings", provider, fmt.Sprintf("%s saved in %s mode", firstNonEmpty(input.DisplayName, provider), firstNonEmpty(input.Mode, "test")), now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) SaveStaff(input StaffInput) (PilotWorkspace, error) {
	if strings.TrimSpace(input.Name) == "" {
		return PilotWorkspace{}, errors.New("staff name is required")
	}
	role := strings.ToLower(strings.TrimSpace(firstNonEmpty(input.Role, "waiter")))
	if !isAllowedStaffRole(role) {
		return PilotWorkspace{}, errors.New("staff role is not supported")
	}
	status := strings.ToLower(strings.TrimSpace(firstNonEmpty(input.Status, "active")))
	if status != "active" && status != "inactive" {
		return PilotWorkspace{}, errors.New("staff status must be active or inactive")
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = mustID()
	}
	hash := ""
	if strings.TrimSpace(input.PIN) != "" {
		hash = hashPIN(input.PIN, id)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	if _, err := tx.Exec(`
		INSERT INTO staff (id, name, role, pin_hash, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			role = excluded.role,
			pin_hash = CASE WHEN excluded.pin_hash = '' THEN staff.pin_hash ELSE excluded.pin_hash END,
			status = excluded.status
	`, id, strings.TrimSpace(input.Name), role, hash, status, nowUTC()); err != nil {
		return PilotWorkspace{}, err
	}
	if err := insertAuditLogTx(tx, "staff_saved", "staff", id, fmt.Sprintf("%s saved as %s", strings.TrimSpace(input.Name), role), nowUTC()); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) SaveMenuCategory(input MenuCategoryInput) (PilotWorkspace, error) {
	if strings.TrimSpace(input.Name) == "" {
		return PilotWorkspace{}, errors.New("category name is required")
	}
	id := firstNonEmpty(input.ID, mustID())
	_, err := s.db.Exec(`
		INSERT INTO menu_categories (id, name, sort_order, status)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, sort_order = excluded.sort_order, status = excluded.status
	`, id, strings.TrimSpace(input.Name), input.SortOrder, firstNonEmpty(input.Status, "active"))
	if err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) SaveMenuModifier(input MenuModifierInput) (PilotWorkspace, error) {
	if strings.TrimSpace(input.Name) == "" {
		return PilotWorkspace{}, errors.New("modifier name is required")
	}
	id := firstNonEmpty(input.ID, mustID())
	_, err := s.db.Exec(`
		INSERT INTO menu_modifiers (id, name, price_delta, route_id, status)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, price_delta = excluded.price_delta, route_id = excluded.route_id, status = excluded.status
	`, id, strings.TrimSpace(input.Name), input.PriceDelta, strings.TrimSpace(input.RouteID), firstNonEmpty(input.Status, "active"))
	if err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) LinkModifierToItem(itemID, modifierID string) (PilotWorkspace, error) {
	if itemID == "" || modifierID == "" {
		return PilotWorkspace{}, errors.New("item and modifier are required")
	}
	if _, err := s.db.Exec("INSERT OR IGNORE INTO menu_item_modifiers (item_id, modifier_id) VALUES (?, ?)", itemID, modifierID); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) SaveFloorSection(input FloorSectionInput) (PilotWorkspace, error) {
	if strings.TrimSpace(input.Name) == "" {
		return PilotWorkspace{}, errors.New("section name is required")
	}
	id := firstNonEmpty(input.ID, mustID())
	_, err := s.db.Exec(`
		INSERT INTO floor_sections (id, name, sort_order)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, sort_order = excluded.sort_order
	`, id, strings.TrimSpace(input.Name), input.SortOrder)
	if err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) SaveDiningTable(input DiningTableInput) (PilotWorkspace, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.SectionID) == "" {
		return PilotWorkspace{}, errors.New("table name and section are required")
	}
	id := firstNonEmpty(input.ID, mustID())
	if input.Seats <= 0 {
		input.Seats = 2
	}
	_, err := s.db.Exec(`
		INSERT INTO dining_tables (id, section_id, name, seats, status)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET section_id = excluded.section_id, name = excluded.name, seats = excluded.seats, status = excluded.status
	`, id, input.SectionID, input.Name, input.Seats, firstNonEmpty(input.Status, "available"))
	if err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) OpenOrderSession(input OpenOrderSessionInput) (OrderSessionDetail, error) {
	if strings.TrimSpace(input.TableID) == "" {
		return OrderSessionDetail{}, errors.New("table is required")
	}
	input.WaiterID = strings.TrimSpace(input.WaiterID)
	if input.WaiterID == "" {
		return OrderSessionDetail{}, errors.New("staff assignment is required")
	}
	if input.GuestCount <= 0 {
		input.GuestCount = 1
	}
	tx, err := s.db.Begin()
	if err != nil {
		return OrderSessionDetail{}, err
	}
	defer rollback(tx)
	var active string
	if err := tx.QueryRow("SELECT active_session_id FROM dining_tables WHERE id = ?", input.TableID).Scan(&active); err != nil {
		return OrderSessionDetail{}, err
	}
	if active != "" {
		return OrderSessionDetail{}, errors.New("table already has an open session")
	}
	var activeStaff int
	if err := tx.QueryRow("SELECT COUNT(*) FROM staff WHERE id = ? AND status = 'active'", input.WaiterID).Scan(&activeStaff); err != nil {
		return OrderSessionDetail{}, err
	}
	if activeStaff == 0 {
		return OrderSessionDetail{}, errors.New("assigned staff member is not active")
	}
	id := mustID()
	now := nowUTC()
	if _, err := tx.Exec(`
		INSERT INTO order_sessions (id, table_id, waiter_id, guest_count, service_mode, status, opened_at)
		VALUES (?, ?, ?, ?, ?, 'open', ?)
	`, id, input.TableID, input.WaiterID, input.GuestCount, firstNonEmpty(input.ServiceMode, "dine_in"), now); err != nil {
		return OrderSessionDetail{}, err
	}
	if _, err := tx.Exec("UPDATE dining_tables SET status = 'occupied', active_session_id = ? WHERE id = ?", id, input.TableID); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := insertSessionEventTx(tx, id, "opened", "Table opened", "operator", now); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := tx.Commit(); err != nil {
		return OrderSessionDetail{}, err
	}
	return s.GetOrderSessionDetail(id)
}

func (s *Store) AddOrderSessionLine(input AddOrderSessionLineInput) (OrderSessionDetail, error) {
	if input.SessionID == "" || input.ItemID == "" {
		return OrderSessionDetail{}, errors.New("session and item are required")
	}
	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	tx, err := s.db.Begin()
	if err != nil {
		return OrderSessionDetail{}, err
	}
	defer rollback(tx)
	var status string
	if err := tx.QueryRow("SELECT status FROM order_sessions WHERE id = ?", input.SessionID).Scan(&status); err != nil {
		return OrderSessionDetail{}, err
	}
	if status != "open" && status != "held" {
		return OrderSessionDetail{}, fmt.Errorf("cannot add lines to %s session", status)
	}
	item, err := menuItemByID(tx, input.ItemID)
	if err != nil {
		return OrderSessionDetail{}, err
	}
	modifiers, err := modifiersByIDTx(tx, input.ModifierIDs)
	if err != nil {
		return OrderSessionDetail{}, err
	}
	modifierTotal := 0.0
	for _, modifier := range modifiers {
		modifierTotal += modifier.PriceDelta
	}
	lineTotal := round2((item.Price + modifierTotal) * float64(input.Quantity))
	lineID := mustID()
	if _, err := tx.Exec(`
		INSERT INTO order_session_lines (id, session_id, item_id, item_name, quantity, unit_price, modifier_total, line_total, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, lineID, input.SessionID, item.ID, item.Name, input.Quantity, item.Price, modifierTotal, lineTotal, strings.TrimSpace(input.Notes)); err != nil {
		return OrderSessionDetail{}, err
	}
	for _, modifier := range modifiers {
		if _, err := tx.Exec(`
			INSERT INTO order_session_line_modifiers (line_id, modifier_id, modifier_name, price_delta)
			VALUES (?, ?, ?, ?)
		`, lineID, modifier.ID, modifier.Name, modifier.PriceDelta); err != nil {
			return OrderSessionDetail{}, err
		}
	}
	if err := recalcSessionTotalsTx(tx, input.SessionID); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := insertSessionEventTx(tx, input.SessionID, "line_added", fmt.Sprintf("%dx %s", input.Quantity, item.Name), "operator", nowUTC()); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := tx.Commit(); err != nil {
		return OrderSessionDetail{}, err
	}
	return s.GetOrderSessionDetail(input.SessionID)
}

func (s *Store) TransferOrderSession(input TableMoveInput) (OrderSessionDetail, error) {
	if input.SessionID == "" || input.TargetTableID == "" {
		return OrderSessionDetail{}, errors.New("session and target table are required")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return OrderSessionDetail{}, err
	}
	defer rollback(tx)
	var oldTableID, active string
	if err := tx.QueryRow("SELECT table_id FROM order_sessions WHERE id = ? AND status IN ('open', 'held')", input.SessionID).Scan(&oldTableID); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := tx.QueryRow("SELECT active_session_id FROM dining_tables WHERE id = ?", input.TargetTableID).Scan(&active); err != nil {
		return OrderSessionDetail{}, err
	}
	if active != "" {
		return OrderSessionDetail{}, errors.New("target table is occupied")
	}
	if _, err := tx.Exec("UPDATE order_sessions SET table_id = ? WHERE id = ?", input.TargetTableID, input.SessionID); err != nil {
		return OrderSessionDetail{}, err
	}
	if _, err := tx.Exec("UPDATE dining_tables SET status = 'available', active_session_id = '' WHERE id = ?", oldTableID); err != nil {
		return OrderSessionDetail{}, err
	}
	if _, err := tx.Exec("UPDATE dining_tables SET status = 'occupied', active_session_id = ? WHERE id = ?", input.SessionID, input.TargetTableID); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := insertSessionEventTx(tx, input.SessionID, "transferred", "Table transferred", "operator", nowUTC()); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := tx.Commit(); err != nil {
		return OrderSessionDetail{}, err
	}
	return s.GetOrderSessionDetail(input.SessionID)
}

func (s *Store) SendOrderSessionKOT(sessionID string) (OrderSessionDetail, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return OrderSessionDetail{}, err
	}
	defer rollback(tx)
	invoiceID, err := ensureSessionInvoiceTx(tx, sessionID, "", "", "cash")
	if err != nil {
		return OrderSessionDetail{}, err
	}
	now := nowUTC()
	if err := sendKOTTx(tx, invoiceID, now); err != nil {
		return OrderSessionDetail{}, err
	}
	if _, err := tx.Exec("UPDATE order_session_lines SET kot_status = 'queued' WHERE session_id = ? AND status = 'open' AND kot_status = 'not_sent'", sessionID); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := insertSessionEventTx(tx, sessionID, "kot_sent", "KOT sent by route", "operator", now); err != nil {
		return OrderSessionDetail{}, err
	}
	if err := tx.Commit(); err != nil {
		return OrderSessionDetail{}, err
	}
	return s.GetOrderSessionDetail(sessionID)
}

func (s *Store) CloseOrderSession(input CloseOrderSessionInput) (Dashboard, error) {
	if input.SessionID == "" {
		return Dashboard{}, errors.New("session is required")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)
	invoiceID, err := ensureSessionInvoiceTx(tx, input.SessionID, input.CustomerName, input.CustomerPhone, input.PaymentMethod)
	if err != nil {
		return Dashboard{}, err
	}
	now := nowUTC()
	if err := closeInvoiceTx(tx, CloseInvoiceInput{InvoiceID: invoiceID, PaymentMethod: input.PaymentMethod, PaymentTendered: input.PaymentTendered}, now); err != nil {
		return Dashboard{}, err
	}
	var tableID string
	if err := tx.QueryRow("SELECT table_id FROM order_sessions WHERE id = ?", input.SessionID).Scan(&tableID); err != nil {
		return Dashboard{}, err
	}
	if _, err := tx.Exec("UPDATE order_sessions SET status = 'closed', closed_at = ? WHERE id = ?", now, input.SessionID); err != nil {
		return Dashboard{}, err
	}
	if _, err := tx.Exec("UPDATE dining_tables SET status = 'available', active_session_id = '' WHERE id = ?", tableID); err != nil {
		return Dashboard{}, err
	}
	if err := insertSessionEventTx(tx, input.SessionID, "closed", "Payment captured and table released", "operator", now); err != nil {
		return Dashboard{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) UpdateKOTStatus(ticketID, status string) (Dashboard, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "queued", "preparing", "ready", "served", "cancelled":
	default:
		return Dashboard{}, errors.New("invalid KOT status")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	var saleID, ticketNumber, routeName string
	if err := tx.QueryRow("SELECT sale_id, ticket_number, route_name FROM kitchen_tickets WHERE id = ?", ticketID).Scan(&saleID, &ticketNumber, &routeName); err != nil {
		return Dashboard{}, err
	}
	if _, err := tx.Exec("UPDATE kitchen_tickets SET status = ? WHERE id = ?", status, ticketID); err != nil {
		return Dashboard{}, err
	}
	if _, err := tx.Exec("UPDATE kitchen_ticket_lines SET status = ? WHERE ticket_id = ?", status, ticketID); err != nil {
		return Dashboard{}, err
	}

	now := nowUTC()
	var sessionID string
	err = tx.QueryRow("SELECT id FROM order_sessions WHERE invoice_id = ?", saleID).Scan(&sessionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Dashboard{}, err
	}
	if sessionID != "" {
		if _, err := tx.Exec(`
			UPDATE order_session_lines
			SET kot_status = ?
			WHERE session_id = ?
				AND item_id IN (SELECT item_id FROM kitchen_ticket_lines WHERE ticket_id = ?)
		`, status, sessionID, ticketID); err != nil {
			return Dashboard{}, err
		}
		if err := insertSessionEventTx(tx, sessionID, "kot_"+status, fmt.Sprintf("%s marked %s", ticketNumber, humanStatusText(status)), "kitchen", now); err != nil {
			return Dashboard{}, err
		}
	}
	if err := insertInvoiceEventTx(tx, saleID, "kot_status", fmt.Sprintf("KOT %s", humanStatusText(status)), fmt.Sprintf("%s / %s", ticketNumber, routeName), "kitchen", now); err != nil {
		return Dashboard{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) GetOrderSessionDetail(sessionID string) (OrderSessionDetail, error) {
	sessions, err := s.orderSessionsWhere("os.id = ?", sessionID)
	if err != nil {
		return OrderSessionDetail{}, err
	}
	if len(sessions) == 0 {
		return OrderSessionDetail{}, errors.New("session not found")
	}
	lines, err := s.orderSessionLines(sessionID)
	if err != nil {
		return OrderSessionDetail{}, err
	}
	events, err := s.orderSessionEvents(sessionID)
	if err != nil {
		return OrderSessionDetail{}, err
	}
	return OrderSessionDetail{Session: sessions[0], Lines: lines, Events: events}, nil
}

func (s *Store) CreatePaymentRequest(invoiceID string) (PaymentRequest, error) {
	summary, err := s.GetInvoiceDetail(invoiceID)
	if err != nil {
		return PaymentRequest{}, err
	}
	now := nowUTC()
	req := PaymentRequest{
		ID:          mustID(),
		InvoiceID:   invoiceID,
		Provider:    "razorpay",
		Amount:      summary.Summary.Total,
		Currency:    "INR",
		Status:      "ready",
		Reference:   summary.Summary.InvoiceNumber,
		CheckoutURL: fmt.Sprintf("https://rzp.io/i/%s", strings.ReplaceAll(strings.ToLower(summary.Summary.InvoiceNumber), "-", "")),
		QRPayload:   fmt.Sprintf("razorpay://payment-link/%s", summary.Summary.InvoiceNumber),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err = s.db.Exec(`
		INSERT INTO payment_requests (id, invoice_id, provider, amount, currency, status, reference, checkout_url, qr_payload, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, req.ID, req.InvoiceID, req.Provider, req.Amount, req.Currency, req.Status, req.Reference, req.CheckoutURL, req.QRPayload, req.CreatedAt, req.UpdatedAt)
	return req, err
}

func (s *Store) MarkPaymentRequestPaid(requestID string) (Dashboard, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return Dashboard{}, errors.New("payment request id is required")
	}
	var invoiceID string
	var amount float64
	var status string
	if err := s.db.QueryRow("SELECT invoice_id, amount, status FROM payment_requests WHERE id = ?", requestID).Scan(&invoiceID, &amount, &status); err != nil {
		return Dashboard{}, err
	}
	if status == "paid" {
		return Dashboard{}, errors.New("payment request is already paid")
	}
	if status == "cancelled" {
		return Dashboard{}, errors.New("cancelled payment request cannot be marked paid")
	}
	dashboard, err := s.CloseInvoice(CloseInvoiceInput{InvoiceID: invoiceID, PaymentMethod: "razorpay", PaymentTendered: amount})
	if err != nil {
		return Dashboard{}, err
	}
	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)
	if _, err := tx.Exec("UPDATE payment_requests SET status = 'paid', updated_at = ? WHERE id = ?", now, requestID); err != nil {
		return Dashboard{}, err
	}
	if err := insertAuditLogTx(tx, "payment_paid", "payment_request", requestID, fmt.Sprintf("Razorpay payment reconciled for invoice %s", invoiceID), now); err != nil {
		return Dashboard{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return dashboard, nil
}

func (s *Store) MarkPaymentRequestFailed(requestID, reason string) (PilotWorkspace, error) {
	return s.updatePaymentRequestStatus(requestID, "failed", firstNonEmpty(reason, "Payment failed or expired"))
}

func (s *Store) CancelPaymentRequest(requestID string) (PilotWorkspace, error) {
	return s.updatePaymentRequestStatus(requestID, "cancelled", "Payment request cancelled by operator")
}

func (s *Store) updatePaymentRequestStatus(requestID, status, detail string) (PilotWorkspace, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return PilotWorkspace{}, errors.New("payment request id is required")
	}
	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	var current string
	if err := tx.QueryRow("SELECT status FROM payment_requests WHERE id = ?", requestID).Scan(&current); err != nil {
		return PilotWorkspace{}, err
	}
	if current == "paid" {
		return PilotWorkspace{}, errors.New("paid payment request cannot be changed")
	}
	if _, err := tx.Exec("UPDATE payment_requests SET status = ?, updated_at = ? WHERE id = ?", status, now, requestID); err != nil {
		return PilotWorkspace{}, err
	}
	if err := insertAuditLogTx(tx, "payment_"+status, "payment_request", requestID, detail, now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) QueuePrintJob(kind, referenceID string) (PrintJob, error) {
	payload := ""
	target := s.printerTarget()
	printerID := "printer_receipt"
	switch kind {
	case "invoice":
		detail, err := s.GetInvoiceDetail(referenceID)
		if err != nil {
			return PrintJob{}, err
		}
		payload = receiptText(detail)
	case "kot":
		tickets, err := s.invoiceKitchenTickets(referenceID)
		if err != nil {
			return PrintJob{}, err
		}
		payload = kotText(tickets)
		printerID = "printer_kitchen"
	default:
		return PrintJob{}, errors.New("unknown print job kind")
	}
	now := nowUTC()
	job := PrintJob{ID: mustID(), Kind: kind, ReferenceID: referenceID, PrinterID: printerID, Target: target, Payload: payload, Status: "ready", CreatedAt: now}
	_, err := s.db.Exec(`
		INSERT INTO print_jobs (id, kind, reference_id, printer_id, target, payload, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.Kind, job.ReferenceID, job.PrinterID, job.Target, job.Payload, job.Status, job.CreatedAt)
	return job, err
}

func (s *Store) MarkPrintJobPrinted(jobID string) (PilotWorkspace, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return PilotWorkspace{}, errors.New("print job id is required")
	}
	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	result, err := tx.Exec(`
		UPDATE print_jobs
		SET status = 'printed', attempts = attempts + 1, last_error = '', printed_at = ?
		WHERE id = ?
	`, now, jobID)
	if err != nil {
		return PilotWorkspace{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return PilotWorkspace{}, errors.New("print job not found")
	}
	if err := insertAuditLogTx(tx, "print_completed", "print_job", jobID, "Print job marked completed by operator", now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) MarkPrintJobFailed(jobID, message string) (PilotWorkspace, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return PilotWorkspace{}, errors.New("print job id is required")
	}
	detail := firstNonEmpty(message, "Printer did not acknowledge the job")
	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	result, err := tx.Exec(`
		UPDATE print_jobs
		SET status = 'failed', attempts = attempts + 1, last_error = ?
		WHERE id = ?
	`, detail, jobID)
	if err != nil {
		return PilotWorkspace{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return PilotWorkspace{}, errors.New("print job not found")
	}
	if err := insertAuditLogTx(tx, "print_failed", "print_job", jobID, detail, now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) RetryPrintJob(jobID string) (PilotWorkspace, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return PilotWorkspace{}, errors.New("print job id is required")
	}
	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	result, err := tx.Exec("UPDATE print_jobs SET status = 'ready', last_error = '' WHERE id = ?", jobID)
	if err != nil {
		return PilotWorkspace{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return PilotWorkspace{}, errors.New("print job not found")
	}
	if err := insertAuditLogTx(tx, "print_retry", "print_job", jobID, "Print job returned to ready queue", now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) SaveVendor(input VendorInput) (PilotWorkspace, error) {
	if strings.TrimSpace(input.Name) == "" {
		return PilotWorkspace{}, errors.New("vendor name is required")
	}
	id := firstNonEmpty(input.ID, mustID())
	_, err := s.db.Exec(`
		INSERT INTO vendors (id, name, phone, gstin, payment_terms, status)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, phone = excluded.phone, gstin = excluded.gstin, payment_terms = excluded.payment_terms, status = excluded.status
	`, id, input.Name, input.Phone, input.GSTIN, input.PaymentTerms, firstNonEmpty(input.Status, "active"))
	if err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) CreatePurchaseOrder(input PurchaseOrderInput) (PilotWorkspace, error) {
	if input.VendorID == "" || len(input.Lines) == 0 {
		return PilotWorkspace{}, errors.New("vendor and lines are required")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	now := nowUTC()
	poID := mustID()
	poNumber, err := nextDocumentNumber(tx, "invoice_prefix", "next_po_seq")
	if err != nil {
		return PilotWorkspace{}, err
	}
	subtotal := 0.0
	for _, line := range input.Lines {
		if line.OrderedQty <= 0 || line.UnitCost < 0 {
			return PilotWorkspace{}, errors.New("purchase order quantities and costs must be valid")
		}
		var ingredientCount int
		if err := tx.QueryRow("SELECT COUNT(*) FROM ingredients WHERE id = ?", line.IngredientID).Scan(&ingredientCount); err != nil {
			return PilotWorkspace{}, err
		}
		if ingredientCount == 0 {
			return PilotWorkspace{}, errors.New("purchase order ingredient not found")
		}
		subtotal += line.OrderedQty * line.UnitCost
	}
	if _, err := tx.Exec(`
		INSERT INTO purchase_orders (id, vendor_id, po_number, status, expected_date, subtotal, created_at)
		VALUES (?, ?, ?, 'draft', ?, ?, ?)
	`, poID, input.VendorID, poNumber, input.ExpectedDate, round2(subtotal), now); err != nil {
		return PilotWorkspace{}, err
	}
	for _, line := range input.Lines {
		if _, err := tx.Exec(`
			INSERT INTO purchase_order_lines (id, purchase_order_id, ingredient_id, ordered_qty, unit_cost)
			VALUES (?, ?, ?, ?, ?)
		`, mustID(), poID, line.IngredientID, line.OrderedQty, line.UnitCost); err != nil {
			return PilotWorkspace{}, err
		}
	}
	if err := insertSyncLog(tx, "purchase_order", poID, "create", input, now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) ReceivePurchaseOrder(input ReceivePurchaseOrderInput) (PilotWorkspace, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return PilotWorkspace{}, err
	}
	defer rollback(tx)
	now := nowUTC()
	var vendorID string
	if err := tx.QueryRow("SELECT vendor_id FROM purchase_orders WHERE id = ?", input.PurchaseOrderID).Scan(&vendorID); err != nil {
		return PilotWorkspace{}, err
	}
	rejectedTotal := 0.0
	reasons := []string{}
	for _, line := range input.Lines {
		var ingredientID string
		var unitCost float64
		if err := tx.QueryRow("SELECT ingredient_id, unit_cost FROM purchase_order_lines WHERE id = ? AND purchase_order_id = ?", line.LineID, input.PurchaseOrderID).Scan(&ingredientID, &unitCost); err != nil {
			return PilotWorkspace{}, err
		}
		if line.AcceptedQty < 0 || line.RejectedQty < 0 {
			return PilotWorkspace{}, errors.New("received quantities cannot be negative")
		}
		if _, err := tx.Exec(`
			UPDATE purchase_order_lines
			SET accepted_qty = ?, rejected_qty = ?, rejection_reason = ?
			WHERE id = ?
		`, line.AcceptedQty, line.RejectedQty, line.RejectionReason, line.LineID); err != nil {
			return PilotWorkspace{}, err
		}
		if line.AcceptedQty > 0 {
			if _, err := tx.Exec("UPDATE ingredients SET on_hand_qty = on_hand_qty + ?, last_purchase_cost = ? WHERE id = ?", line.AcceptedQty, unitCost, ingredientID); err != nil {
				return PilotWorkspace{}, err
			}
			if _, err := tx.Exec(`
				INSERT INTO inventory_movements (id, ingredient_id, type, quantity, reason, reference_id, created_at)
				VALUES (?, ?, 'purchase_acceptance', ?, 'PO accepted stock', ?, ?)
			`, mustID(), ingredientID, line.AcceptedQty, input.PurchaseOrderID, now); err != nil {
				return PilotWorkspace{}, err
			}
		}
		if line.RejectedQty > 0 {
			rejectedTotal += line.RejectedQty * unitCost
			reasons = append(reasons, firstNonEmpty(line.RejectionReason, "Rejected during receiving"))
		}
	}
	if _, err := tx.Exec(`
		UPDATE purchase_orders
		SET status = 'received', rejected_total = ?, closed_at = ?
		WHERE id = ?
	`, round2(rejectedTotal), now, input.PurchaseOrderID); err != nil {
		return PilotWorkspace{}, err
	}
	if rejectedTotal > 0 {
		noteID := mustID()
		reason := strings.Join(reasons, "; ")
		if _, err := tx.Exec(`
			INSERT INTO vendor_debit_notes (id, vendor_id, purchase_order_id, amount, reason, status, created_at)
			VALUES (?, ?, ?, ?, ?, 'ready', ?)
		`, noteID, vendorID, input.PurchaseOrderID, round2(rejectedTotal), reason, now); err != nil {
			return PilotWorkspace{}, err
		}
		if err := queueVendorNotificationTx(tx, vendorID, input.PurchaseOrderID, rejectedTotal, reason, now); err != nil {
			return PilotWorkspace{}, err
		}
	}
	if err := insertSyncLog(tx, "purchase_order", input.PurchaseOrderID, "receive", input, now); err != nil {
		return PilotWorkspace{}, err
	}
	if err := tx.Commit(); err != nil {
		return PilotWorkspace{}, err
	}
	return s.GetPilotWorkspace()
}

func (s *Store) CloseDay(input DayCloseInput) (DayCloseSummary, error) {
	if input.BusinessDate == "" {
		input.BusinessDate = todayDate()
	}
	summary, err := s.dayClosePreview(input.BusinessDate)
	if err != nil {
		return DayCloseSummary{}, err
	}
	summary.ID = mustID()
	summary.CashCounted = round2(input.CashCounted)
	summary.CashVariance = round2(summary.CashCounted - summary.CashExpected)
	summary.ClosedAt = nowUTC()
	summary.Status = "closed"
	summary.Notes = strings.TrimSpace(input.Notes)
	_, err = s.db.Exec(`
		INSERT INTO day_closes (
			id, business_date, closed_at, status, sales_total, cash_expected, cash_counted, cash_variance,
			upi_total, card_total, razorpay_total, refund_total, void_count, discount_total, notes
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(business_date) DO UPDATE SET
			closed_at = excluded.closed_at,
			status = excluded.status,
			sales_total = excluded.sales_total,
			cash_expected = excluded.cash_expected,
			cash_counted = excluded.cash_counted,
			cash_variance = excluded.cash_variance,
			upi_total = excluded.upi_total,
			card_total = excluded.card_total,
			razorpay_total = excluded.razorpay_total,
			refund_total = excluded.refund_total,
			void_count = excluded.void_count,
			discount_total = excluded.discount_total,
			notes = excluded.notes
	`, summary.ID, summary.BusinessDate, summary.ClosedAt, summary.Status, summary.SalesTotal, summary.CashExpected, summary.CashCounted, summary.CashVariance,
		summary.UPITotal, summary.CardTotal, summary.RazorpayTotal, summary.RefundTotal, summary.VoidCount, summary.DiscountTotal, summary.Notes)
	if err != nil {
		return DayCloseSummary{}, err
	}
	return summary, nil
}

func (s *Store) RebuildAccountingShell() (AccountingSnapshot, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return AccountingSnapshot{}, err
	}
	defer rollback(tx)
	if _, err := tx.Exec("DELETE FROM accounting_voucher_lines"); err != nil {
		return AccountingSnapshot{}, err
	}
	if _, err := tx.Exec("DELETE FROM accounting_vouchers"); err != nil {
		return AccountingSnapshot{}, err
	}
	rows, err := tx.Query("SELECT id FROM sales WHERE status IN ('closed', 'partially_refunded', 'refunded') ORDER BY created_at")
	if err != nil {
		return AccountingSnapshot{}, err
	}
	var invoiceIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return AccountingSnapshot{}, err
		}
		invoiceIDs = append(invoiceIDs, id)
	}
	if err := rows.Close(); err != nil {
		return AccountingSnapshot{}, err
	}
	for _, id := range invoiceIDs {
		if err := postInvoiceAccountingTx(tx, id); err != nil {
			return AccountingSnapshot{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return AccountingSnapshot{}, err
	}
	return s.GetAccountingSnapshot()
}

func (s *Store) GetAccountingSnapshot() (AccountingSnapshot, error) {
	if err := s.ensureAccountingVouchersCurrent(); err != nil {
		return AccountingSnapshot{}, err
	}
	rows, err := s.db.Query(`
		SELECT l.id, l.name, g.name, l.normal_balance,
			COALESCE(SUM(CASE WHEN vl.line_type = 'Debit' THEN vl.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN vl.line_type = 'Credit' THEN vl.amount ELSE 0 END), 0)
		FROM account_ledgers l
		JOIN account_groups g ON g.id = l.group_id
		LEFT JOIN accounting_voucher_lines vl ON vl.ledger_id = l.id
		GROUP BY l.id, l.name, g.name, l.normal_balance, g.sort_order, l.sort_order
		ORDER BY g.sort_order, l.sort_order
	`)
	if err != nil {
		return AccountingSnapshot{}, err
	}
	defer rows.Close()
	snapshot := AccountingSnapshot{TrialBalance: make([]AccountingBalanceRow, 0)}
	for rows.Next() {
		var row AccountingBalanceRow
		if err := rows.Scan(&row.LedgerID, &row.LedgerName, &row.GroupName, &row.NormalBalance, &row.DebitTotal, &row.CreditTotal); err != nil {
			return AccountingSnapshot{}, err
		}
		if row.NormalBalance == "Debit" {
			row.Balance = round2(row.DebitTotal - row.CreditTotal)
		} else {
			row.Balance = round2(row.CreditTotal - row.DebitTotal)
		}
		if row.Balance < 0 {
			row.Balance = math.Abs(row.Balance)
			if row.NormalBalance == "Debit" {
				row.BalanceSide = "Credit"
			} else {
				row.BalanceSide = "Debit"
			}
		} else {
			row.BalanceSide = row.NormalBalance
		}
		switch row.LedgerID {
		case "ledger_sales":
			snapshot.SalesTotal = row.CreditTotal
		case "ledger_output_gst":
			snapshot.TaxPayable = row.CreditTotal - row.DebitTotal
		case "ledger_cash", "ledger_bank":
			snapshot.CashAndBank += row.DebitTotal - row.CreditTotal
		case "ledger_razorpay":
			snapshot.RazorpayClearing = row.DebitTotal - row.CreditTotal
		case "ledger_vendor_payable":
			snapshot.VendorPayables = row.CreditTotal - row.DebitTotal
		case "ledger_refunds":
			snapshot.RefundTotal = row.DebitTotal
		}
		snapshot.TrialBalance = append(snapshot.TrialBalance, row)
	}
	if err := rows.Err(); err != nil {
		return AccountingSnapshot{}, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM accounting_vouchers").Scan(&snapshot.VoucherCount); err != nil {
		return AccountingSnapshot{}, err
	}
	snapshot.CashAndBank = round2(snapshot.CashAndBank)
	return snapshot, nil
}

func (s *Store) ensureAccountingVouchersCurrent() error {
	var voucherCount, closedCount int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM accounting_vouchers WHERE source_kind = 'invoice'").Scan(&voucherCount); err != nil {
		return err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM sales WHERE status IN ('closed', 'partially_refunded', 'refunded')").Scan(&closedCount); err != nil {
		return err
	}
	if voucherCount < closedCount {
		_, err := s.RebuildAccountingShell()
		return err
	}
	return nil
}

func (s *Store) restaurantSettings() (RestaurantSettings, error) {
	var settings RestaurantSettings
	err := s.db.QueryRow(`
		SELECT id, restaurant_name, gstin, legal_name, address, state, tax_mode, default_tax_rate,
			service_charge_rate, invoice_prefix, business_hours, backup_path, receipt_printer_id, updated_at
		FROM restaurant_settings
		WHERE id = 'primary'
	`).Scan(&settings.ID, &settings.RestaurantName, &settings.GSTIN, &settings.LegalName, &settings.Address, &settings.State, &settings.TaxMode,
		&settings.DefaultTaxRate, &settings.ServiceChargeRate, &settings.InvoicePrefix, &settings.BusinessHours, &settings.BackupPath, &settings.ReceiptPrinterID, &settings.UpdatedAt)
	return settings, err
}

func (s *Store) integrationSettings() ([]IntegrationSetting, error) {
	rows, err := s.db.Query(`
		SELECT id, provider, mode, display_name, base_url, credential_status, health_status, last_checked_at, last_error, updated_at
		FROM integration_settings
		ORDER BY provider
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]IntegrationSetting, 0)
	for rows.Next() {
		var item IntegrationSetting
		if err := rows.Scan(&item.ID, &item.Provider, &item.Mode, &item.DisplayName, &item.BaseURL, &item.CredentialStatus, &item.HealthStatus, &item.LastCheckedAt, &item.LastError, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) staffMembers() ([]StaffMember, error) {
	rows, err := s.db.Query("SELECT id, name, role, pin_hash, status, created_at FROM staff ORDER BY role, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]StaffMember, 0)
	for rows.Next() {
		var item StaffMember
		if err := rows.Scan(&item.ID, &item.Name, &item.Role, &item.PINHash, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.PINHash = redactSecret(item.PINHash)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) menuCategories() ([]MenuCategory, error) {
	rows, err := s.db.Query("SELECT id, name, sort_order, status FROM menu_categories ORDER BY sort_order, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]MenuCategory, 0)
	for rows.Next() {
		var item MenuCategory
		if err := rows.Scan(&item.ID, &item.Name, &item.SortOrder, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) menuModifiers() ([]MenuModifier, error) {
	rows, err := s.db.Query("SELECT id, name, price_delta, route_id, status FROM menu_modifiers ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]MenuModifier, 0)
	for rows.Next() {
		var item MenuModifier
		if err := rows.Scan(&item.ID, &item.Name, &item.PriceDelta, &item.RouteID, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) menuItemModifiers() ([]MenuItemModifier, error) {
	rows, err := s.db.Query("SELECT item_id, modifier_id FROM menu_item_modifiers ORDER BY item_id, modifier_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]MenuItemModifier, 0)
	for rows.Next() {
		var item MenuItemModifier
		if err := rows.Scan(&item.ItemID, &item.ModifierID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) floorSections() ([]FloorSection, error) {
	rows, err := s.db.Query("SELECT id, name, sort_order FROM floor_sections ORDER BY sort_order, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]FloorSection, 0)
	for rows.Next() {
		var item FloorSection
		if err := rows.Scan(&item.ID, &item.Name, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) diningTables() ([]DiningTable, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.section_id, s.name, t.name, t.seats, t.status, t.active_session_id
		FROM dining_tables t
		JOIN floor_sections s ON s.id = t.section_id
		ORDER BY s.sort_order, t.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]DiningTable, 0)
	for rows.Next() {
		var item DiningTable
		if err := rows.Scan(&item.ID, &item.SectionID, &item.SectionName, &item.Name, &item.Seats, &item.Status, &item.ActiveSessionID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) orderSessions() ([]OrderSession, error) {
	return s.orderSessionsWhere("os.status IN ('open', 'held')", nil)
}

func (s *Store) orderSessionsWhere(where string, arg any) ([]OrderSession, error) {
	query := `
		SELECT os.id, os.table_id, t.name, fs.name, os.waiter_id, COALESCE(st.name, ''), os.guest_count, os.service_mode,
			os.status, os.invoice_id, os.subtotal, os.tax_total, os.service_charge, os.total, os.opened_at, os.closed_at,
			(SELECT COUNT(*) FROM order_session_lines osl WHERE osl.session_id = os.id AND osl.status = 'open'),
			(SELECT COUNT(*) FROM order_session_lines osl WHERE osl.session_id = os.id AND osl.status = 'open' AND osl.kot_status IN ('ready', 'served')),
			(SELECT COUNT(*) FROM order_session_lines osl WHERE osl.session_id = os.id AND osl.status = 'open' AND osl.kot_status = 'preparing'),
			(SELECT COUNT(*) FROM order_session_lines osl WHERE osl.session_id = os.id AND osl.status = 'open' AND osl.kot_status = 'queued'),
			(SELECT COUNT(*) FROM order_session_lines osl WHERE osl.session_id = os.id AND osl.status = 'open' AND osl.kot_status = 'not_sent')
		FROM order_sessions os
		JOIN dining_tables t ON t.id = os.table_id
		JOIN floor_sections fs ON fs.id = t.section_id
		LEFT JOIN staff st ON st.id = os.waiter_id
		WHERE ` + where + `
		ORDER BY os.opened_at DESC
	`
	var rows *sql.Rows
	var err error
	if arg == nil {
		rows, err = s.db.Query(query)
	} else {
		rows, err = s.db.Query(query, arg)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OrderSession, 0)
	for rows.Next() {
		var item OrderSession
		if err := rows.Scan(&item.ID, &item.TableID, &item.TableName, &item.SectionName, &item.WaiterID, &item.WaiterName, &item.GuestCount,
			&item.ServiceMode, &item.Status, &item.InvoiceID, &item.Subtotal, &item.TaxTotal, &item.ServiceCharge, &item.Total, &item.OpenedAt, &item.ClosedAt,
			&item.LineCount, &item.ReadyLineCount, &item.PreparingLineCount, &item.QueuedLineCount, &item.NotSentLineCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) orderSessionLines(sessionID string) ([]OrderSessionLine, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, item_id, item_name, quantity, unit_price, modifier_total, line_total, notes, status, kot_status
		FROM order_session_lines
		WHERE session_id = ?
		ORDER BY rowid
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OrderSessionLine, 0)
	for rows.Next() {
		var item OrderSessionLine
		if err := rows.Scan(&item.ID, &item.SessionID, &item.ItemID, &item.ItemName, &item.Quantity, &item.UnitPrice, &item.ModifierTotal, &item.LineTotal, &item.Notes, &item.Status, &item.KOTStatus); err != nil {
			return nil, err
		}
		mods, names, err := sessionLineModifiers(s.db, item.ID)
		if err != nil {
			return nil, err
		}
		item.ModifierIDs = mods
		item.ModifierNames = names
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) orderSessionEvents(sessionID string) ([]OrderSessionEvent, error) {
	rows, err := s.db.Query("SELECT id, session_id, type, detail, actor, created_at FROM order_session_events WHERE session_id = ? ORDER BY created_at", sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OrderSessionEvent, 0)
	for rows.Next() {
		var item OrderSessionEvent
		if err := rows.Scan(&item.ID, &item.SessionID, &item.Type, &item.Detail, &item.Actor, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) paymentRequests() ([]PaymentRequest, error) {
	rows, err := s.db.Query("SELECT id, invoice_id, provider, amount, currency, status, reference, checkout_url, qr_payload, created_at, updated_at FROM payment_requests ORDER BY created_at DESC LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PaymentRequest, 0)
	for rows.Next() {
		var item PaymentRequest
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.Provider, &item.Amount, &item.Currency, &item.Status, &item.Reference, &item.CheckoutURL, &item.QRPayload, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) printJobs() ([]PrintJob, error) {
	rows, err := s.db.Query("SELECT id, kind, reference_id, printer_id, target, payload, status, attempts, last_error, created_at, printed_at FROM print_jobs ORDER BY created_at DESC LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PrintJob, 0)
	for rows.Next() {
		var item PrintJob
		if err := rows.Scan(&item.ID, &item.Kind, &item.ReferenceID, &item.PrinterID, &item.Target, &item.Payload, &item.Status, &item.Attempts, &item.LastError, &item.CreatedAt, &item.PrintedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) printerTarget() string {
	var target string
	if err := s.db.QueryRow("SELECT base_url FROM integration_settings WHERE provider = 'escpos_printer'").Scan(&target); err != nil {
		return "system://default"
	}
	return firstNonEmpty(target, "system://default")
}

func (s *Store) GetAuditLog(limit int) ([]AuditLogEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 40
	}
	rows, err := s.db.Query(`
		SELECT id, event_type, target_type, target_id, detail, actor, created_at
		FROM audit_log
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AuditLogEntry, 0)
	for rows.Next() {
		var item AuditLogEntry
		if err := rows.Scan(&item.ID, &item.EventType, &item.TargetType, &item.TargetID, &item.Detail, &item.Actor, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) vendors() ([]Vendor, error) {
	rows, err := s.db.Query("SELECT id, name, phone, gstin, payment_terms, quality_score, status FROM vendors ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Vendor, 0)
	for rows.Next() {
		var item Vendor
		if err := rows.Scan(&item.ID, &item.Name, &item.Phone, &item.GSTIN, &item.PaymentTerms, &item.QualityScore, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) purchaseOrders() ([]PurchaseOrder, error) {
	rows, err := s.db.Query(`
		SELECT po.id, po.vendor_id, v.name, po.po_number, po.status, po.expected_date, po.subtotal, po.rejected_total, po.created_at
		FROM purchase_orders po
		JOIN vendors v ON v.id = po.vendor_id
		ORDER BY po.created_at DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PurchaseOrder, 0)
	for rows.Next() {
		var item PurchaseOrder
		if err := rows.Scan(&item.ID, &item.VendorID, &item.VendorName, &item.PONumber, &item.Status, &item.ExpectedDate, &item.Subtotal, &item.RejectedTotal, &item.CreatedAt); err != nil {
			return nil, err
		}
		lines, err := s.purchaseOrderLines(item.ID)
		if err != nil {
			return nil, err
		}
		item.Lines = lines
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) purchaseOrderLines(poID string) ([]PurchaseOrderLine, error) {
	rows, err := s.db.Query(`
		SELECT pol.id, pol.purchase_order_id, pol.ingredient_id, i.name, pol.ordered_qty, pol.accepted_qty, pol.rejected_qty, pol.unit_cost, pol.rejection_reason
		FROM purchase_order_lines pol
		JOIN ingredients i ON i.id = pol.ingredient_id
		WHERE pol.purchase_order_id = ?
		ORDER BY i.name
	`, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PurchaseOrderLine, 0)
	for rows.Next() {
		var item PurchaseOrderLine
		if err := rows.Scan(&item.ID, &item.PurchaseOrderID, &item.IngredientID, &item.IngredientName, &item.OrderedQty, &item.AcceptedQty, &item.RejectedQty, &item.UnitCost, &item.RejectionReason); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) vendorDebitNotes() ([]VendorDebitNote, error) {
	rows, err := s.db.Query(`
		SELECT dn.id, dn.vendor_id, v.name, dn.purchase_order_id, dn.amount, dn.reason, dn.status, dn.created_at
		FROM vendor_debit_notes dn
		JOIN vendors v ON v.id = dn.vendor_id
		ORDER BY dn.created_at DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]VendorDebitNote, 0)
	for rows.Next() {
		var item VendorDebitNote
		if err := rows.Scan(&item.ID, &item.VendorID, &item.VendorName, &item.PurchaseOrderID, &item.Amount, &item.Reason, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) dayClosePreview(date string) (DayCloseSummary, error) {
	start, end := dayBounds(date)
	summary := DayCloseSummary{BusinessDate: date, Status: "preview"}
	if err := s.db.QueryRow(`
		SELECT COALESCE(SUM(total), 0), COALESCE(SUM(discount_total), 0)
		FROM sales
		WHERE created_at >= ? AND created_at < ? AND status IN ('closed', 'partially_refunded', 'refunded')
	`, start, end).Scan(&summary.SalesTotal, &summary.DiscountTotal); err != nil {
		return summary, err
	}
	if err := s.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM refunds WHERE created_at >= ? AND created_at < ?", start, end).Scan(&summary.RefundTotal); err != nil {
		return summary, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM sales WHERE created_at >= ? AND created_at < ? AND status = 'voided'", start, end).Scan(&summary.VoidCount); err != nil {
		return summary, err
	}
	rows, err := s.db.Query("SELECT method, COALESCE(SUM(amount), 0) FROM payments WHERE created_at >= ? AND created_at < ? GROUP BY method", start, end)
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	for rows.Next() {
		var method string
		var amount float64
		if err := rows.Scan(&method, &amount); err != nil {
			return summary, err
		}
		switch strings.ToLower(method) {
		case "cash":
			summary.CashExpected += amount
		case "upi":
			summary.UPITotal += amount
		case "card":
			summary.CardTotal += amount
		case "razorpay":
			summary.RazorpayTotal += amount
		}
	}
	return summary, rows.Err()
}

func (s *Store) seedAccountingDefaultsTx(tx *sql.Tx) error {
	if _, err := tx.Exec(`
		INSERT INTO account_groups (id, name, system_key, category, normal_balance, sort_order, is_system)
		VALUES
			('grp_assets', 'Assets', 'assets', 'assets', 'Debit', 10, 1),
			('grp_liabilities', 'Liabilities', 'liabilities', 'liabilities', 'Credit', 20, 1),
			('grp_income', 'Income', 'income', 'income', 'Credit', 30, 1),
			('grp_expenses', 'Expenses', 'expenses', 'expenses', 'Debit', 40, 1),
			('grp_equity', 'Equity', 'equity', 'equity', 'Credit', 50, 1)
		ON CONFLICT(id) DO NOTHING
	`); err != nil {
		return err
	}
	_, err := tx.Exec(`
		INSERT INTO account_ledgers (id, group_id, name, system_key, normal_balance, sort_order, is_system)
		VALUES
			('ledger_cash', 'grp_assets', 'Cash On Hand', 'cash_on_hand', 'Debit', 10, 1),
			('ledger_bank', 'grp_assets', 'Bank / UPI Settlement', 'bank_main', 'Debit', 20, 1),
			('ledger_razorpay', 'grp_assets', 'Razorpay Clearing', 'razorpay_clearing', 'Debit', 30, 1),
			('ledger_inventory', 'grp_assets', 'Inventory Asset', 'inventory_asset', 'Debit', 40, 1),
			('ledger_output_gst', 'grp_liabilities', 'Output GST Payable', 'output_gst_payable', 'Credit', 10, 1),
			('ledger_vendor_payable', 'grp_liabilities', 'Vendor Payables', 'vendor_payables', 'Credit', 20, 1),
			('ledger_sales', 'grp_income', 'Food And Beverage Sales', 'food_beverage_sales', 'Credit', 10, 1),
			('ledger_service_charge', 'grp_income', 'Service Charge Income', 'service_charge_income', 'Credit', 20, 1),
			('ledger_discounts', 'grp_expenses', 'Discounts Given', 'discounts_given', 'Debit', 10, 1),
			('ledger_refunds', 'grp_expenses', 'Refunds Given', 'refunds_given', 'Debit', 20, 1),
			('ledger_cogs', 'grp_expenses', 'Cost Of Goods Sold', 'cost_of_goods_sold', 'Debit', 30, 1)
		ON CONFLICT(id) DO NOTHING
	`)
	return err
}

func postInvoiceAccountingTx(tx *sql.Tx, invoiceID string) error {
	summary, err := invoiceSummaryByIDTx(tx, invoiceID)
	if err != nil {
		return err
	}
	postings := []accountingPostingLine{
		{ledgerKey: paymentLedger(summary.PaymentMethod), lineType: "Debit", amount: summary.Total, narration: summary.InvoiceNumber, sortOrder: 10},
		{ledgerKey: "ledger_sales", lineType: "Credit", amount: summary.Subtotal - summary.DiscountTotal, narration: "Restaurant sales", sortOrder: 20},
		{ledgerKey: "ledger_output_gst", lineType: "Credit", amount: summary.TaxTotal, narration: "GST payable", sortOrder: 30},
	}
	if summary.DiscountTotal > 0 {
		postings = append(postings, accountingPostingLine{ledgerKey: "ledger_discounts", lineType: "Debit", amount: summary.DiscountTotal, narration: "Discounts", sortOrder: 40})
		postings = append(postings, accountingPostingLine{ledgerKey: "ledger_sales", lineType: "Credit", amount: summary.DiscountTotal, narration: "Gross discount offset", sortOrder: 41})
	}
	if err := upsertAccountingVoucherTx(tx, summary.ClosedAt, "Sales Invoice", fmt.Sprintf("Invoice %s", summary.InvoiceNumber), "invoice", invoiceID, summary.InvoiceNumber, postings); err != nil {
		return err
	}
	rows, err := tx.Query("SELECT id, amount, reason FROM refunds WHERE invoice_id = ?", invoiceID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var refundID, reason string
		var amount float64
		if err := rows.Scan(&refundID, &amount, &reason); err != nil {
			return err
		}
		refundPostings := []accountingPostingLine{
			{ledgerKey: "ledger_refunds", lineType: "Debit", amount: amount, narration: reason, sortOrder: 10},
			{ledgerKey: paymentLedger(summary.PaymentMethod), lineType: "Credit", amount: amount, narration: summary.InvoiceNumber, sortOrder: 20},
		}
		if err := upsertAccountingVoucherTx(tx, summary.ClosedAt, "Refund", fmt.Sprintf("Refund %s", summary.InvoiceNumber), "refund", refundID, summary.InvoiceNumber, refundPostings); err != nil {
			return err
		}
	}
	return rows.Err()
}

type accountingPostingLine struct {
	ledgerKey string
	lineType  string
	amount    float64
	narration string
	sortOrder int
}

func upsertAccountingVoucherTx(tx *sql.Tx, voucherDate, voucherType, narration, sourceKind, sourceID, reference string, postings []accountingPostingLine) error {
	postings = compactAccountingPostings(postings)
	if len(postings) == 0 {
		return nil
	}
	var debit, credit float64
	for _, posting := range postings {
		if posting.lineType == "Debit" {
			debit += posting.amount
		} else {
			credit += posting.amount
		}
	}
	if round2(debit) != round2(credit) {
		return fmt.Errorf("unbalanced accounting voucher: debit %.2f credit %.2f", debit, credit)
	}
	voucherID := ""
	err := tx.QueryRow("SELECT id FROM accounting_vouchers WHERE source_kind = ? AND source_id = ?", sourceKind, sourceID).Scan(&voucherID)
	if err == sql.ErrNoRows {
		voucherID = mustID()
		if _, err := tx.Exec(`
			INSERT INTO accounting_vouchers (id, voucher_date, voucher_type, narration, source_kind, source_id, reference_number, total_amount, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, voucherID, firstNonEmpty(voucherDate, nowUTC()), voucherType, narration, sourceKind, sourceID, reference, round2(debit), nowUTC()); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if _, err := tx.Exec("DELETE FROM accounting_voucher_lines WHERE voucher_id = ?", voucherID); err != nil {
		return err
	}
	for _, posting := range postings {
		if _, err := tx.Exec(`
			INSERT INTO accounting_voucher_lines (id, voucher_id, ledger_id, line_type, amount, narration, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, mustID(), voucherID, posting.ledgerKey, posting.lineType, round2(posting.amount), posting.narration, posting.sortOrder); err != nil {
			return err
		}
	}
	return nil
}

func compactAccountingPostings(postings []accountingPostingLine) []accountingPostingLine {
	result := make([]accountingPostingLine, 0, len(postings))
	for _, posting := range postings {
		posting.amount = round2(posting.amount)
		if posting.amount > 0 {
			result = append(result, posting)
		}
	}
	return result
}

func paymentLedger(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "cash":
		return "ledger_cash"
	case "razorpay":
		return "ledger_razorpay"
	default:
		return "ledger_bank"
	}
}

func ensureSessionInvoiceTx(tx *sql.Tx, sessionID, customerName, customerPhone, paymentMethod string) (string, error) {
	var invoiceID, tableName, serviceMode string
	var subtotal, taxTotal, serviceCharge, total float64
	err := tx.QueryRow(`
		SELECT os.invoice_id, t.name, os.service_mode, os.subtotal, os.tax_total, os.service_charge, os.total
		FROM order_sessions os
		JOIN dining_tables t ON t.id = os.table_id
		WHERE os.id = ?
	`, sessionID).Scan(&invoiceID, &tableName, &serviceMode, &subtotal, &taxTotal, &serviceCharge, &total)
	if err != nil {
		return "", err
	}
	if invoiceID != "" {
		if strings.TrimSpace(customerName) != "" || strings.TrimSpace(customerPhone) != "" || strings.TrimSpace(paymentMethod) != "" {
			if _, err := tx.Exec(`
				UPDATE sales
				SET
					customer_name = CASE WHEN ? != '' THEN ? ELSE customer_name END,
					customer_phone = CASE WHEN ? != '' THEN ? ELSE customer_phone END,
					payment_method = CASE WHEN ? != '' THEN ? ELSE payment_method END
				WHERE id = ?
			`, strings.TrimSpace(customerName), strings.TrimSpace(customerName), normalizePhone(customerPhone), normalizePhone(customerPhone),
				strings.TrimSpace(paymentMethod), strings.TrimSpace(paymentMethod), invoiceID); err != nil {
				return "", err
			}
		}
		return invoiceID, nil
	}
	rows, err := tx.Query(`
		SELECT item_id, item_name, quantity, unit_price + modifier_total, line_total, notes
		FROM order_session_lines
		WHERE session_id = ? AND status = 'open'
		ORDER BY rowid
	`, sessionID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	type line struct {
		itemID, itemName, notes string
		qty                     int
		unitPrice, lineTotal    float64
	}
	lines := make([]line, 0)
	for rows.Next() {
		var item line
		if err := rows.Scan(&item.itemID, &item.itemName, &item.qty, &item.unitPrice, &item.lineTotal, &item.notes); err != nil {
			return "", err
		}
		lines = append(lines, item)
	}
	if len(lines) == 0 {
		return "", errors.New("session has no open lines")
	}
	now := nowUTC()
	invoiceID = mustID()
	invoiceNumber, err := nextDocumentNumber(tx, "invoice_prefix", "next_invoice_seq")
	if err != nil {
		return "", err
	}
	if _, err := tx.Exec(`
		INSERT INTO sales (
			id, invoice_number, customer_name, customer_phone, table_name, order_type,
			subtotal, discount_total, tax_total, total, channel, payment_method, status, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?, 'floor', ?, 'draft', ?)
	`, invoiceID, invoiceNumber, strings.TrimSpace(customerName), normalizePhone(customerPhone), tableName, serviceMode, round2(subtotal+serviceCharge), taxTotal, total, firstNonEmpty(paymentMethod, "cash"), now); err != nil {
		return "", err
	}
	for _, line := range lines {
		if _, err := tx.Exec(`
			INSERT INTO sale_lines (id, sale_id, item_id, item_name, quantity, unit_price, line_total, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, mustID(), invoiceID, line.itemID, line.itemName, line.qty, line.unitPrice, line.lineTotal, line.notes); err != nil {
			return "", err
		}
	}
	if _, err := tx.Exec("UPDATE order_sessions SET invoice_id = ? WHERE id = ?", invoiceID, sessionID); err != nil {
		return "", err
	}
	if err := insertInvoiceEventTx(tx, invoiceID, "session_invoice", "Created from table session", tableName, "operator", now); err != nil {
		return "", err
	}
	return invoiceID, nil
}

func recalcSessionTotalsTx(tx *sql.Tx, sessionID string) error {
	var subtotal float64
	if err := tx.QueryRow("SELECT COALESCE(SUM(line_total), 0) FROM order_session_lines WHERE session_id = ? AND status = 'open'", sessionID).Scan(&subtotal); err != nil {
		return err
	}
	var taxRate, serviceRate float64
	if err := tx.QueryRow("SELECT default_tax_rate, service_charge_rate FROM restaurant_settings WHERE id = 'primary'").Scan(&taxRate, &serviceRate); err != nil {
		return err
	}
	serviceCharge := round2(subtotal * serviceRate)
	tax := round2((subtotal + serviceCharge) * taxRate)
	total := round2(subtotal + serviceCharge + tax)
	_, err := tx.Exec("UPDATE order_sessions SET subtotal = ?, service_charge = ?, tax_total = ?, total = ? WHERE id = ?", round2(subtotal), serviceCharge, tax, total, sessionID)
	return err
}

func modifiersByIDTx(tx *sql.Tx, ids []string) ([]MenuModifier, error) {
	modifiers := make([]MenuModifier, 0, len(ids))
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			continue
		}
		var mod MenuModifier
		err := tx.QueryRow("SELECT id, name, price_delta, route_id, status FROM menu_modifiers WHERE id = ? AND status = 'active'", id).
			Scan(&mod.ID, &mod.Name, &mod.PriceDelta, &mod.RouteID, &mod.Status)
		if err != nil {
			return nil, err
		}
		modifiers = append(modifiers, mod)
	}
	return modifiers, nil
}

func sessionLineModifiers(db *sql.DB, lineID string) ([]string, []string, error) {
	rows, err := db.Query("SELECT modifier_id, modifier_name FROM order_session_line_modifiers WHERE line_id = ? ORDER BY modifier_name", lineID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	ids := []string{}
	names := []string{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		names = append(names, name)
	}
	return ids, names, rows.Err()
}

func insertSessionEventTx(tx *sql.Tx, sessionID, eventType, detail, actor, createdAt string) error {
	_, err := tx.Exec(`
		INSERT INTO order_session_events (id, session_id, type, detail, actor, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, mustID(), sessionID, eventType, detail, actor, createdAt)
	return err
}

func insertAuditLogTx(tx *sql.Tx, eventType, targetType, targetID, detail, createdAt string) error {
	if _, err := tx.Exec(`
		INSERT INTO audit_log (id, event_type, target_type, target_id, detail, actor, created_at)
		VALUES (?, ?, ?, ?, ?, 'operator', ?)
	`, mustID(), eventType, targetType, targetID, detail, createdAt); err != nil {
		return err
	}
	payload := map[string]any{"eventType": eventType, "targetType": targetType, "targetID": targetID, "detail": detail, "createdAt": createdAt}
	return insertSyncLog(tx, "audit_log", targetID, eventType, payload, createdAt)
}

func queueVendorNotificationTx(tx *sql.Tx, vendorID, poID string, amount float64, reason string, createdAt string) error {
	var phone, name string
	if err := tx.QueryRow("SELECT phone, name FROM vendors WHERE id = ?", vendorID).Scan(&phone, &name); err != nil {
		return err
	}
	status := "ready"
	if normalizePhone(phone) == "" {
		status = "skipped"
	}
	payload, _ := json.Marshal(map[string]any{"vendor": name, "purchaseOrderID": poID, "amount": amount, "reason": reason})
	_, err := tx.Exec(`
		INSERT INTO notification_queue (id, invoice_id, channel, recipient, template, payload, status, created_at)
		VALUES (?, ?, 'whatsapp', ?, 'vendor_debit_note', ?, ?, ?)
	`, mustID(), poID, normalizePhone(phone), string(payload), status, createdAt)
	return err
}

func (s *Store) invoiceKitchenTickets(invoiceID string) ([]KitchenTicket, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer rollback(tx)
	tickets, err := invoiceKitchenTicketsTx(tx, invoiceID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return tickets, nil
}

func receiptText(detail InvoiceDetail) string {
	lines := []string{"NEXUS", detail.Summary.InvoiceNumber, "----------------"}
	for _, line := range detail.Lines {
		lines = append(lines, fmt.Sprintf("%dx %s %.2f", line.Quantity, line.ItemName, line.LineTotal))
	}
	lines = append(lines, "----------------", fmt.Sprintf("TOTAL %.2f", detail.Summary.Total))
	return strings.Join(lines, "\n")
}

func kotText(tickets []KitchenTicket) string {
	lines := []string{"NEXUS KOT"}
	for _, ticket := range tickets {
		lines = append(lines, ticket.TicketNumber+" "+ticket.RouteName)
		for _, line := range ticket.Lines {
			lines = append(lines, fmt.Sprintf("%dx %s %s", line.Quantity, line.ItemName, line.Notes))
		}
	}
	return strings.Join(lines, "\n")
}

func sealLocalSecret(provider, value string) string {
	sum := sha256.Sum256([]byte(provider + ":" + value))
	return "local:v1:" + base64.StdEncoding.EncodeToString(sum[:])
}

func hashPIN(pin, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + pin))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func isAllowedStaffRole(role string) bool {
	switch role {
	case "owner", "manager", "cashier", "waiter", "runner", "kitchen", "barista":
		return true
	default:
		return false
	}
}

func redactSecret(value string) string {
	if value == "" {
		return ""
	}
	return "stored"
}

func humanStatusText(status string) string {
	return strings.ReplaceAll(strings.TrimSpace(status), "_", " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func todayDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

func dayBounds(date string) (string, string) {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		parsed = time.Now().UTC()
	}
	start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	return start.Format(time.RFC3339), start.Add(24 * time.Hour).Format(time.RFC3339)
}
