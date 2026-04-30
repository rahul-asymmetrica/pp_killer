package nexus

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Store) AuthenticateStaff(input StaffLoginInput) (StaffSession, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return StaffSession{}, err
	}
	defer rollback(tx)

	session, err := authenticateStaffTx(tx, input)
	if err != nil {
		return StaffSession{}, err
	}
	if err := insertAuditLogTx(tx, "staff_login", "staff", session.StaffID, fmt.Sprintf("%s opened %s workspace", session.Name, firstNonEmpty(input.Workspace, "default")), session.IssuedAt); err != nil {
		return StaffSession{}, err
	}
	if err := tx.Commit(); err != nil {
		return StaffSession{}, err
	}
	return session, nil
}

func (s *Store) AuthorizeStaffAction(input StaffActionApprovalInput) (ApprovalToken, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return ApprovalToken{}, err
	}
	defer rollback(tx)

	approval, err := approveStaffActionTx(tx, input, nowUTC())
	if err != nil {
		return ApprovalToken{}, err
	}
	if err := tx.Commit(); err != nil {
		return ApprovalToken{}, err
	}
	return approval, nil
}

func authenticateStaffTx(tx *sql.Tx, input StaffLoginInput) (StaffSession, error) {
	staffID := strings.TrimSpace(input.StaffID)
	if staffID == "" {
		return StaffSession{}, errors.New("staff member is required")
	}
	pin := strings.TrimSpace(input.PIN)
	if pin == "" {
		return StaffSession{}, errors.New("staff PIN is required")
	}

	var name, role, pinHash, status string
	if err := tx.QueryRow("SELECT name, role, pin_hash, status FROM staff WHERE id = ?", staffID).Scan(&name, &role, &pinHash, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StaffSession{}, errors.New("staff member not found")
		}
		return StaffSession{}, err
	}
	if status != "active" {
		return StaffSession{}, errors.New("staff member is inactive")
	}
	if pinHash == "" {
		return StaffSession{}, errors.New("staff PIN is not configured")
	}
	if hashPIN(pin, staffID) != pinHash {
		return StaffSession{}, errors.New("invalid staff PIN")
	}

	role = strings.ToLower(strings.TrimSpace(role))
	workspace := strings.TrimSpace(input.Workspace)
	access := workspaceAccessForRole(role)
	if workspace != "" && !containsString(access, workspace) {
		return StaffSession{}, fmt.Errorf("%s cannot access %s workspace", humanStatusText(role), humanStatusText(workspace))
	}

	issued := time.Now().UTC()
	return StaffSession{
		StaffID:         staffID,
		Name:            name,
		Role:            role,
		Permissions:     permissionsForRole(role),
		WorkspaceAccess: access,
		IssuedAt:        issued.Format(time.RFC3339),
		ExpiresAt:       issued.Add(12 * time.Hour).Format(time.RFC3339),
	}, nil
}

func approveStaffActionTx(tx *sql.Tx, input StaffActionApprovalInput, createdAt string) (ApprovalToken, error) {
	action := strings.ToLower(strings.TrimSpace(input.Action))
	if action == "" {
		return ApprovalToken{}, errors.New("approval action is required")
	}
	session, err := authenticateStaffTx(tx, StaffLoginInput{StaffID: input.StaffID, PIN: input.PIN})
	if err != nil {
		return ApprovalToken{}, err
	}
	if !roleAllowsAction(session.Role, action) {
		return ApprovalToken{}, fmt.Errorf("%s cannot approve %s", humanStatusText(session.Role), humanStatusText(action))
	}

	token := mustID()
	expiresAt := time.Now().UTC().Add(15 * time.Minute).Format(time.RFC3339)
	if _, err := tx.Exec(`
		INSERT INTO manager_approvals (id, action, invoice_id, approved_by, approved_at, token, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, mustID(), action, strings.TrimSpace(input.TargetID), session.Name, createdAt, token, expiresAt); err != nil {
		return ApprovalToken{}, err
	}
	if err := insertAuditLogTx(tx, "staff_approval_"+action, "approval", strings.TrimSpace(input.TargetID), fmt.Sprintf("%s approved %s", session.Name, humanStatusText(action)), createdAt); err != nil {
		return ApprovalToken{}, err
	}
	return ApprovalToken{Token: token, ApprovedBy: session.Name, ApprovedAt: createdAt, ExpiresAt: expiresAt}, nil
}

func permissionsForRole(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner":
		return []string{"workspace:front", "workspace:kitchenOps", "workspace:backoffice", "workspace:admin", "pos:sell", "kds:update", "invoice:void", "invoice:refund", "session:void_line", "day_close:close", "settings:manage", "reports:view", "exports:run", "backup:create", "staff:manage"}
	case "manager":
		return []string{"workspace:front", "workspace:kitchenOps", "workspace:backoffice", "workspace:admin", "pos:sell", "kds:update", "invoice:void", "invoice:refund", "session:void_line", "day_close:close", "settings:manage", "reports:view", "exports:run", "backup:create", "staff:manage"}
	case "cashier":
		return []string{"workspace:front", "pos:sell", "reports:receipts", "day_close:close"}
	case "waiter", "runner":
		return []string{"workspace:front", "pos:tables", "pos:kot"}
	case "kitchen", "barista":
		return []string{"workspace:kitchenOps", "kds:update", "kds:print"}
	default:
		return []string{}
	}
}

func workspaceAccessForRole(role string) []string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "manager":
		return []string{"front", "kitchenOps", "backoffice", "admin"}
	case "cashier", "waiter", "runner":
		return []string{"front"}
	case "kitchen", "barista":
		return []string{"kitchenOps"}
	default:
		return []string{}
	}
}

func roleAllowsAction(role, action string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	action = strings.ToLower(strings.TrimSpace(action))
	if role == "owner" || role == "manager" {
		return true
	}
	switch action {
	case "close_day":
		return role == "cashier"
	case "kds_update", "kds_print":
		return role == "kitchen" || role == "barista"
	case "pos_sell", "table_service":
		return role == "cashier" || role == "waiter" || role == "runner"
	default:
		return false
	}
}

func ensureBusinessDateOpenTx(tx *sql.Tx, businessDate string) error {
	businessDate = strings.TrimSpace(businessDate)
	if businessDate == "" {
		businessDate = todayDate()
	}
	var status string
	err := tx.QueryRow("SELECT status FROM day_closes WHERE business_date = ?", businessDate).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if strings.ToLower(status) == "closed" {
		return fmt.Errorf("business date %s is closed", businessDate)
	}
	return nil
}

func ensureTimestampBusinessDateOpenTx(tx *sql.Tx, timestamp string) error {
	return ensureBusinessDateOpenTx(tx, businessDateFromTimestamp(timestamp))
}

func ensureSessionBusinessDateOpenTx(tx *sql.Tx, sessionID string) error {
	var openedAt string
	if err := tx.QueryRow("SELECT opened_at FROM order_sessions WHERE id = ?", sessionID).Scan(&openedAt); err != nil {
		return err
	}
	return ensureTimestampBusinessDateOpenTx(tx, openedAt)
}

func businessDateFromTimestamp(timestamp string) string {
	value := strings.TrimSpace(timestamp)
	if value == "" {
		return todayDate()
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		parsedDate, dateErr := time.Parse("2006-01-02", value)
		if dateErr != nil {
			return todayDate()
		}
		return parsedDate.Format("2006-01-02")
	}
	return parsed.UTC().Format("2006-01-02")
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
