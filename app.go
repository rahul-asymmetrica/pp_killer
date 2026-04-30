package main

import (
	"context"

	"NEXUS/internal/nexus"
)

// App struct
type App struct {
	ctx     context.Context
	store   *nexus.Store
	initErr error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	path, err := nexus.DefaultDBPath()
	if err != nil {
		a.initErr = err
		return
	}

	store, err := nexus.Open(path)
	if err != nil {
		a.initErr = err
		return
	}
	a.store = store
}

func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) GetDashboard() (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.Dashboard()
}

func (a *App) GetAdminAnalytics(rangeKey string) (nexus.AdminAnalytics, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.AdminAnalytics{}, err
	}
	return a.store.GetAdminAnalytics(rangeKey)
}

func (a *App) AuthenticateStaff(input nexus.StaffLoginInput) (nexus.StaffSession, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.StaffSession{}, err
	}
	return a.store.AuthenticateStaff(input)
}

func (a *App) AuthorizeStaffAction(input nexus.StaffActionApprovalInput) (nexus.ApprovalToken, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.ApprovalToken{}, err
	}
	return a.store.AuthorizeStaffAction(input)
}

func (a *App) RecordSale(input nexus.SaleInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.RecordSale(input)
}

func (a *App) SaveOrder(input nexus.SaleInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.SaveOrder(input)
}

func (a *App) SendKOT(invoiceID string) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.SendKOT(invoiceID)
}

func (a *App) CloseInvoice(input nexus.CloseInvoiceInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.CloseInvoice(input)
}

func (a *App) GetInvoices(filter nexus.InvoiceFilter) ([]nexus.InvoiceSummary, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.store.GetInvoices(filter)
}

func (a *App) GetInvoiceDetail(invoiceID string) (nexus.InvoiceDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.InvoiceDetail{}, err
	}
	return a.store.GetInvoiceDetail(invoiceID)
}

func (a *App) VoidInvoice(input nexus.VoidInvoiceInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.VoidInvoice(input)
}

func (a *App) RefundInvoice(input nexus.RefundInvoiceInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.RefundInvoice(input)
}

func (a *App) SplitInvoice(input nexus.SplitInvoiceInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.SplitInvoice(input)
}

func (a *App) ValidateManagerPIN(pin string) (nexus.ApprovalToken, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.ApprovalToken{}, err
	}
	return a.store.ValidateManagerPIN(pin)
}

func (a *App) GetNotifications() ([]nexus.NotificationRecord, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.store.GetNotifications()
}

func (a *App) MarkNotificationRead(id string) error {
	if err := a.ensureReady(); err != nil {
		return err
	}
	return a.store.MarkNotificationRead(id)
}

func (a *App) ReceiveDelivery(input nexus.DeliveryInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.ReceiveDelivery(input)
}

func (a *App) ReconcileStock(input nexus.ReconcileInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.ReconcileStock(input)
}

func (a *App) UpdateRecipe(input nexus.RecipeUpdateInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.UpdateRecipe(input)
}

func (a *App) UpdateIngredientSettings(input nexus.IngredientUpdateInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.UpdateIngredientSettings(input)
}

func (a *App) SaveIngredient(input nexus.IngredientInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.SaveIngredient(input)
}

func (a *App) GetPilotWorkspace() (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.GetPilotWorkspace()
}

func (a *App) SaveRestaurantSettings(input nexus.RestaurantSettings) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveRestaurantSettings(input)
}

func (a *App) SaveIntegrationSetting(input nexus.IntegrationSettingInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveIntegrationSetting(input)
}

func (a *App) SaveStaff(input nexus.StaffInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveStaff(input)
}

func (a *App) SaveMenuCategory(input nexus.MenuCategoryInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveMenuCategory(input)
}

func (a *App) SaveMenuModifier(input nexus.MenuModifierInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveMenuModifier(input)
}

func (a *App) SaveMenuItem(input nexus.MenuItemInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.SaveMenuItem(input)
}

func (a *App) ImportMenuCSV(input nexus.MenuImportInput) (nexus.MenuImportResult, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.MenuImportResult{}, err
	}
	return a.store.ImportMenuCSV(input)
}

func (a *App) LinkModifierToItem(itemID string, modifierID string) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.LinkModifierToItem(itemID, modifierID)
}

func (a *App) SaveFloorSection(input nexus.FloorSectionInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveFloorSection(input)
}

func (a *App) SaveDiningTable(input nexus.DiningTableInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveDiningTable(input)
}

func (a *App) OpenOrderSession(input nexus.OpenOrderSessionInput) (nexus.OrderSessionDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.OrderSessionDetail{}, err
	}
	return a.store.OpenOrderSession(input)
}

func (a *App) AssignOrderSessionStaff(input nexus.AssignOrderSessionStaffInput) (nexus.OrderSessionDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.OrderSessionDetail{}, err
	}
	return a.store.AssignOrderSessionStaff(input)
}

func (a *App) AddOrderSessionLine(input nexus.AddOrderSessionLineInput) (nexus.OrderSessionDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.OrderSessionDetail{}, err
	}
	return a.store.AddOrderSessionLine(input)
}

func (a *App) VoidOrderSessionLine(input nexus.SessionLineVoidInput) (nexus.OrderSessionDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.OrderSessionDetail{}, err
	}
	return a.store.VoidOrderSessionLine(input)
}

func (a *App) TransferOrderSession(input nexus.TableMoveInput) (nexus.OrderSessionDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.OrderSessionDetail{}, err
	}
	return a.store.TransferOrderSession(input)
}

func (a *App) SendOrderSessionKOT(sessionID string) (nexus.OrderSessionDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.OrderSessionDetail{}, err
	}
	return a.store.SendOrderSessionKOT(sessionID)
}

func (a *App) CloseOrderSession(input nexus.CloseOrderSessionInput) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.CloseOrderSession(input)
}

func (a *App) UpdateKOTStatus(ticketID string, status string) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.UpdateKOTStatus(ticketID, status)
}

func (a *App) GetOrderSessionDetail(sessionID string) (nexus.OrderSessionDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.OrderSessionDetail{}, err
	}
	return a.store.GetOrderSessionDetail(sessionID)
}

func (a *App) CreatePaymentRequest(invoiceID string) (nexus.PaymentRequest, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PaymentRequest{}, err
	}
	return a.store.CreatePaymentRequest(invoiceID)
}

func (a *App) MarkPaymentRequestPaid(requestID string) (nexus.Dashboard, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.Dashboard{}, err
	}
	return a.store.MarkPaymentRequestPaid(requestID)
}

func (a *App) MarkPaymentRequestFailed(requestID string, reason string) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.MarkPaymentRequestFailed(requestID, reason)
}

func (a *App) CancelPaymentRequest(requestID string) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.CancelPaymentRequest(requestID)
}

func (a *App) QueuePrintJob(kind string, referenceID string) (nexus.PrintJob, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PrintJob{}, err
	}
	return a.store.QueuePrintJob(kind, referenceID)
}

func (a *App) MarkPrintJobPrinted(jobID string) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.MarkPrintJobPrinted(jobID)
}

func (a *App) MarkPrintJobFailed(jobID string, message string) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.MarkPrintJobFailed(jobID, message)
}

func (a *App) RetryPrintJob(jobID string) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.RetryPrintJob(jobID)
}

func (a *App) GetAuditLog(limit int) ([]nexus.AuditLogEntry, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.store.GetAuditLog(limit)
}

func (a *App) SavePrinterConnection(input nexus.PrinterConnectionInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SavePrinterConnection(input)
}

func (a *App) ExportInvoicePDF(invoiceID string) (nexus.ExportResult, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.ExportResult{}, err
	}
	return a.store.ExportInvoicePDF(invoiceID)
}

func (a *App) ExportAccountingCSV() (nexus.ExportResult, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.ExportResult{}, err
	}
	return a.store.ExportAccountingCSV()
}

func (a *App) ExportInvoicesCSV(filter nexus.InvoiceFilter) (nexus.ExportResult, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.ExportResult{}, err
	}
	return a.store.ExportInvoicesCSV(filter)
}

func (a *App) ExportDayClosePDF(businessDate string) (nexus.ExportResult, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.ExportResult{}, err
	}
	return a.store.ExportDayClosePDF(businessDate)
}

func (a *App) GetSyncStatus() (nexus.SyncStatus, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.SyncStatus{}, err
	}
	return a.store.GetSyncStatus()
}

func (a *App) ExportBackup(input nexus.BackupInput) (nexus.ExportResult, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.ExportResult{}, err
	}
	return a.store.ExportBackup(input)
}

func (a *App) SeedDemoOperations(months int) (nexus.DemoSeedResult, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.DemoSeedResult{}, err
	}
	return a.store.SeedDemoOperations(months)
}

func (a *App) SaveVendor(input nexus.VendorInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.SaveVendor(input)
}

func (a *App) CreatePurchaseOrder(input nexus.PurchaseOrderInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.CreatePurchaseOrder(input)
}

func (a *App) ReceivePurchaseOrder(input nexus.ReceivePurchaseOrderInput) (nexus.PilotWorkspace, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.PilotWorkspace{}, err
	}
	return a.store.ReceivePurchaseOrder(input)
}

func (a *App) CloseDay(input nexus.DayCloseInput) (nexus.DayCloseSummary, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.DayCloseSummary{}, err
	}
	return a.store.CloseDay(input)
}

func (a *App) RebuildAccountingShell() (nexus.AccountingSnapshot, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.AccountingSnapshot{}, err
	}
	return a.store.RebuildAccountingShell()
}

func (a *App) GetAccountingSnapshot() (nexus.AccountingSnapshot, error) {
	if err := a.ensureReady(); err != nil {
		return nexus.AccountingSnapshot{}, err
	}
	return a.store.GetAccountingSnapshot()
}

func (a *App) ensureReady() error {
	if a.initErr != nil {
		return a.initErr
	}
	if a.store == nil {
		path, err := nexus.DefaultDBPath()
		if err != nil {
			return err
		}
		store, err := nexus.Open(path)
		if err != nil {
			return err
		}
		a.store = store
	}
	return nil
}
