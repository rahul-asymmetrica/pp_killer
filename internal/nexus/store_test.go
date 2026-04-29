package nexus

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveDraftDoesNotDeductInventoryOrCreateKOT(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	before := ingredientQty(t, store, "ing_milk")
	invoice := saveDraftOrder(t, store, []SaleLineInput{{ItemID: "item_latte", Quantity: 2}})

	after := ingredientQty(t, store, "ing_milk")
	if after != before {
		t.Fatalf("draft should not deduct inventory, before=%f after=%f", before, after)
	}
	if invoice.Status != "draft" {
		t.Fatalf("expected draft status, got %s", invoice.Status)
	}
	var ticketCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM kitchen_tickets WHERE sale_id = ?", invoice.ID).Scan(&ticketCount); err != nil {
		t.Fatalf("query ticket count: %v", err)
	}
	if ticketCount != 0 {
		t.Fatalf("expected no KOT for draft, got %d", ticketCount)
	}
}

func TestSendKOTCreatesRouteTicketsAndDeductsOnce(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	invoice := saveDraftOrder(t, store, []SaleLineInput{
		{ItemID: "item_latte", Quantity: 1},
		{ItemID: "item_sandwich", Quantity: 1},
	})
	before := ingredientQty(t, store, "ing_milk")

	if _, err := store.SendKOT(invoice.ID); err != nil {
		t.Fatalf("send kot: %v", err)
	}
	afterFirst := ingredientQty(t, store, "ing_milk")
	if afterFirst >= before {
		t.Fatalf("expected KOT to deduct recipe stock, before=%f after=%f", before, afterFirst)
	}

	if _, err := store.SendKOT(invoice.ID); err != nil {
		t.Fatalf("send kot again: %v", err)
	}
	afterSecond := ingredientQty(t, store, "ing_milk")
	if afterSecond != afterFirst {
		t.Fatalf("KOT should be idempotent, afterFirst=%f afterSecond=%f", afterFirst, afterSecond)
	}

	var ticketCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM kitchen_tickets WHERE sale_id = ?", invoice.ID).Scan(&ticketCount); err != nil {
		t.Fatalf("query ticket count: %v", err)
	}
	if ticketCount != 2 {
		t.Fatalf("expected one KOT per route, got %d", ticketCount)
	}
}

func TestCloseInvoiceFinalizesPaymentWithoutDoubleDeducting(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	invoice := saveDraftOrder(t, store, []SaleLineInput{{ItemID: "item_latte", Quantity: 1}})
	if _, err := store.SendKOT(invoice.ID); err != nil {
		t.Fatalf("send kot: %v", err)
	}
	afterKOT := ingredientQty(t, store, "ing_milk")

	if _, err := store.CloseInvoice(CloseInvoiceInput{InvoiceID: invoice.ID, PaymentMethod: "upi"}); err != nil {
		t.Fatalf("close invoice: %v", err)
	}
	afterClose := ingredientQty(t, store, "ing_milk")
	if afterClose != afterKOT {
		t.Fatalf("close should not double-deduct inventory, afterKOT=%f afterClose=%f", afterKOT, afterClose)
	}

	detail, err := store.GetInvoiceDetail(invoice.ID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.Summary.Status != "closed" {
		t.Fatalf("expected closed invoice, got %s", detail.Summary.Status)
	}
	if len(detail.Payments) != 1 {
		t.Fatalf("expected one payment, got %d", len(detail.Payments))
	}
	if len(detail.Notifications) != 1 || detail.Notifications[0].Status != "ready" {
		t.Fatalf("expected ready WhatsApp notification, got %#v", detail.Notifications)
	}
}

func TestVoidRequiresManagerPINAndWritesAuditEvent(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	invoice := saveDraftOrder(t, store, []SaleLineInput{{ItemID: "item_latte", Quantity: 1}})
	if _, err := store.VoidInvoice(VoidInvoiceInput{InvoiceID: invoice.ID, PIN: "0000", Reason: "Wrong table"}); err == nil {
		t.Fatalf("expected invalid PIN to fail")
	}
	if _, err := store.VoidInvoice(VoidInvoiceInput{InvoiceID: invoice.ID, PIN: "1234", Reason: "Wrong table"}); err != nil {
		t.Fatalf("void invoice: %v", err)
	}

	detail, err := store.GetInvoiceDetail(invoice.ID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.Summary.Status != "voided" {
		t.Fatalf("expected voided invoice, got %s", detail.Summary.Status)
	}
	if len(detail.Events) < 2 || detail.Events[len(detail.Events)-1].Type != "voided" {
		t.Fatalf("expected void audit event, got %#v", detail.Events)
	}
}

func TestRefundRequiresPINAndWritesReversalRecords(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	invoice := closeDraftOrder(t, store, []SaleLineInput{{ItemID: "item_latte", Quantity: 2}})
	if _, err := store.RefundInvoice(RefundInvoiceInput{InvoiceID: invoice.ID, PIN: "0000", Amount: 50, Reason: "Service issue"}); err == nil {
		t.Fatalf("expected invalid PIN to fail")
	}
	if _, err := store.RefundInvoice(RefundInvoiceInput{InvoiceID: invoice.ID, PIN: "1234", Amount: 50, Reason: "Service issue"}); err != nil {
		t.Fatalf("refund invoice: %v", err)
	}

	detail, err := store.GetInvoiceDetail(invoice.ID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.Summary.Status != "partially_refunded" {
		t.Fatalf("expected partial refund, got %s", detail.Summary.Status)
	}
	if len(detail.Refunds) != 1 {
		t.Fatalf("expected one refund record, got %d", len(detail.Refunds))
	}
	if len(detail.Payments) != 2 || detail.Payments[1].Amount != -50 {
		t.Fatalf("expected payment reversal, got %#v", detail.Payments)
	}
}

func TestItemSplitPreservesTotalsAndLinksInvoices(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	source := saveDraftOrder(t, store, []SaleLineInput{
		{ItemID: "item_latte", Quantity: 2},
		{ItemID: "item_sandwich", Quantity: 1},
	})
	detail, err := store.GetInvoiceDetail(source.ID)
	if err != nil {
		t.Fatalf("get source detail: %v", err)
	}
	originalTotal := detail.Summary.Total
	var latteLineID string
	for _, line := range detail.Lines {
		if line.ItemID == "item_latte" {
			latteLineID = line.ID
		}
	}
	if latteLineID == "" {
		t.Fatalf("latte line not found")
	}

	if _, err := store.SplitInvoice(SplitInvoiceInput{
		InvoiceID: source.ID,
		Mode:      "items",
		Lines:     []SplitLineInput{{LineID: latteLineID, Quantity: 1}},
	}); err != nil {
		t.Fatalf("split invoice: %v", err)
	}

	invoices, err := store.GetInvoices(InvoiceFilter{})
	if err != nil {
		t.Fatalf("get invoices: %v", err)
	}
	var updatedSource, child InvoiceSummary
	for _, invoice := range invoices {
		if invoice.ID == source.ID {
			updatedSource = invoice
		}
		if invoice.SourceInvoiceID == source.ID {
			child = invoice
		}
	}
	if child.ID == "" {
		t.Fatalf("expected child invoice")
	}
	if math.Abs((updatedSource.Total+child.Total)-originalTotal) > 0.01 {
		t.Fatalf("split totals drifted: source=%f child=%f original=%f", updatedSource.Total, child.Total, originalTotal)
	}
	if child.SplitGroupID == "" || child.SplitGroupID != updatedSource.SplitGroupID {
		t.Fatalf("expected linked split group, source=%s child=%s", updatedSource.SplitGroupID, child.SplitGroupID)
	}
}

func TestAmountSplitPreservesTotalPayable(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	source := saveDraftOrder(t, store, []SaleLineInput{
		{ItemID: "item_latte", Quantity: 1},
		{ItemID: "item_sandwich", Quantity: 1},
	})
	total := source.Total
	first := math.Round((total/2)*100) / 100
	second := math.Round((total-first)*100) / 100

	if _, err := store.SplitInvoice(SplitInvoiceInput{
		InvoiceID: source.ID,
		Mode:      "amount",
		Amounts:   []float64{first, second},
	}); err != nil {
		t.Fatalf("amount split: %v", err)
	}

	invoices, err := store.GetInvoices(InvoiceFilter{})
	if err != nil {
		t.Fatalf("get invoices: %v", err)
	}
	childTotal := 0.0
	childCount := 0
	firstChildID := ""
	for _, invoice := range invoices {
		if invoice.SourceInvoiceID == source.ID {
			childTotal += invoice.Total
			childCount++
			if firstChildID == "" {
				firstChildID = invoice.ID
			}
		}
	}
	if childCount != 2 {
		t.Fatalf("expected two child invoices, got %d", childCount)
	}
	if math.Abs(childTotal-total) > 0.01 {
		t.Fatalf("amount split totals drifted: children=%f original=%f", childTotal, total)
	}
	if _, err := store.CloseInvoice(CloseInvoiceInput{InvoiceID: firstChildID, PaymentMethod: "upi"}); err != nil {
		t.Fatalf("close amount split child: %v", err)
	}
}

func TestRecordSaleDeductsRecipeInventory(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	before := ingredientQty(t, store, "ing_milk")
	if _, err := store.RecordSale(SaleInput{
		CustomerName:  "Test Guest",
		CustomerPhone: "+91 70000 00000",
		Channel:       "counter",
		Lines: []SaleLineInput{
			{ItemID: "item_latte", Quantity: 2},
		},
	}); err != nil {
		t.Fatalf("record sale: %v", err)
	}

	after := ingredientQty(t, store, "ing_milk")
	if after >= before {
		t.Fatalf("expected milk stock to decrease, before=%f after=%f", before, after)
	}

	var pending int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM sync_log WHERE entity = 'sale' AND status = 'pending'").Scan(&pending); err != nil {
		t.Fatalf("query sync log: %v", err)
	}
	if pending != 1 {
		t.Fatalf("expected one pending sale sync item, got %d", pending)
	}

	var invoiceNumber string
	var taxTotal, discountTotal float64
	if err := store.db.QueryRow("SELECT invoice_number, tax_total, discount_total FROM sales ORDER BY created_at DESC LIMIT 1").Scan(&invoiceNumber, &taxTotal, &discountTotal); err != nil {
		t.Fatalf("query sale totals: %v", err)
	}
	if invoiceNumber != "NX-00001" {
		t.Fatalf("expected first invoice number NX-00001, got %s", invoiceNumber)
	}
	if taxTotal <= 0 {
		t.Fatalf("expected tax to be calculated, got %f", taxTotal)
	}
	if discountTotal != 0 {
		t.Fatalf("expected no discount by default, got %f", discountTotal)
	}
}

func TestRecordSaleCreatesKitchenTicketsByRoute(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.RecordSale(SaleInput{
		CustomerName:    "Table Guest",
		OrderType:       "dine_in",
		TableName:       "T-07",
		DiscountType:    "percent",
		DiscountValue:   10,
		PaymentMethod:   "upi",
		PaymentTendered: 0,
		Lines: []SaleLineInput{
			{ItemID: "item_latte", Quantity: 1},
			{ItemID: "item_sandwich", Quantity: 1, Notes: "No tomato"},
		},
	}); err != nil {
		t.Fatalf("record sale: %v", err)
	}

	var ticketCount int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM kitchen_tickets").Scan(&ticketCount); err != nil {
		t.Fatalf("query ticket count: %v", err)
	}
	if ticketCount != 2 {
		t.Fatalf("expected one KOT per route, got %d", ticketCount)
	}

	var discount float64
	if err := store.db.QueryRow("SELECT discount_total FROM sales ORDER BY created_at DESC LIMIT 1").Scan(&discount); err != nil {
		t.Fatalf("query discount: %v", err)
	}
	if discount <= 0 {
		t.Fatalf("expected discount to be calculated")
	}
}

func TestReceiveDeliveryKeepsRejectedStockOutOfInventory(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	before := ingredientQty(t, store, "ing_tomato")
	if _, err := store.ReceiveDelivery(DeliveryInput{
		VendorName:    "Fresh Farm Co",
		InvoiceNumber: "INV-9",
		Lines: []DeliveryLineInput{
			{
				IngredientID:    "ing_tomato",
				OrderedQty:      5000,
				AcceptedQty:     3500,
				RejectedQty:     1500,
				UnitCost:        0.08,
				RejectionReason: "Soft batch",
			},
		},
	}); err != nil {
		t.Fatalf("receive delivery: %v", err)
	}

	after := ingredientQty(t, store, "ing_tomato")
	if after != before+3500 {
		t.Fatalf("expected only accepted tomatoes to enter inventory, before=%f after=%f", before, after)
	}
}

func TestReconcileStockAdjustsWasteFactor(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.RecordSale(SaleInput{
		CustomerName: "Audit Guest",
		Channel:      "counter",
		Lines: []SaleLineInput{
			{ItemID: "item_latte", Quantity: 5},
		},
	}); err != nil {
		t.Fatalf("record sale: %v", err)
	}

	beforeWaste := ingredientWaste(t, store, "ing_milk")
	current := ingredientQty(t, store, "ing_milk")
	if _, err := store.ReconcileStock(ReconcileInput{
		IngredientID: "ing_milk",
		PhysicalQty:  current - 300,
		Note:         "Spot check variance",
	}); err != nil {
		t.Fatalf("reconcile stock: %v", err)
	}

	afterWaste := ingredientWaste(t, store, "ing_milk")
	if afterWaste <= beforeWaste {
		t.Fatalf("expected waste factor to increase, before=%f after=%f", beforeWaste, afterWaste)
	}
}

func TestUpdateRecipeAndIngredientSettings(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.UpdateRecipe(RecipeUpdateInput{
		ItemID: "item_latte",
		Components: []RecipeLineUpdateInput{
			{IngredientID: "ing_beans", Quantity: 20},
			{IngredientID: "ing_milk", Quantity: 220},
		},
	}); err != nil {
		t.Fatalf("update recipe: %v", err)
	}

	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM recipes WHERE item_id = ?", "item_latte").Scan(&count); err != nil {
		t.Fatalf("query recipe count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected updated recipe to have 2 components, got %d", count)
	}

	if _, err := store.UpdateIngredientSettings(IngredientUpdateInput{
		IngredientID:     "ing_milk",
		ReorderPoint:     12000,
		PurchaseUnit:     "crate",
		PurchaseToUsage:  12000,
		LastPurchaseCost: 720,
	}); err != nil {
		t.Fatalf("update ingredient settings: %v", err)
	}

	var reorderPoint, conversion float64
	var purchaseUnit string
	if err := store.db.QueryRow("SELECT reorder_point, purchase_unit, purchase_to_usage FROM ingredients WHERE id = ?", "ing_milk").Scan(&reorderPoint, &purchaseUnit, &conversion); err != nil {
		t.Fatalf("query ingredient settings: %v", err)
	}
	if reorderPoint != 12000 || purchaseUnit != "crate" || conversion != 12000 {
		t.Fatalf("ingredient settings were not saved")
	}
}

func TestOrderSessionTableFlowCreatesInvoiceAndReleasesTable(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	beforeMilk := ingredientQty(t, store, "ing_milk")
	session, err := store.OpenOrderSession(OpenOrderSessionInput{TableID: "tbl_01", WaiterID: "staff_waiter", GuestCount: 2})
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	session, err = store.AddOrderSessionLine(AddOrderSessionLineInput{
		SessionID:   session.Session.ID,
		ItemID:      "item_latte",
		Quantity:    1,
		ModifierIDs: []string{"mod_extra_shot"},
	})
	if err != nil {
		t.Fatalf("add session line: %v", err)
	}
	if len(session.Lines) != 1 || session.Lines[0].ModifierTotal != 45 {
		t.Fatalf("expected extra shot modifier on session line, got %#v", session.Lines)
	}

	session, err = store.SendOrderSessionKOT(session.Session.ID)
	if err != nil {
		t.Fatalf("send session kot: %v", err)
	}
	afterKOT := ingredientQty(t, store, "ing_milk")
	if afterKOT >= beforeMilk {
		t.Fatalf("expected session KOT to deduct recipe stock, before=%f after=%f", beforeMilk, afterKOT)
	}
	if session.Session.InvoiceID == "" {
		t.Fatalf("expected session invoice to be linked")
	}

	if _, err := store.CloseOrderSession(CloseOrderSessionInput{
		SessionID:       session.Session.ID,
		CustomerName:    "Pilot Guest",
		CustomerPhone:   "+917700000000",
		PaymentMethod:   "cash",
		PaymentTendered: session.Session.Total,
	}); err != nil {
		t.Fatalf("close session: %v", err)
	}
	afterClose := ingredientQty(t, store, "ing_milk")
	if afterClose != afterKOT {
		t.Fatalf("close session should not double deduct, afterKOT=%f afterClose=%f", afterKOT, afterClose)
	}
	var tableStatus, activeSession string
	if err := store.db.QueryRow("SELECT status, active_session_id FROM dining_tables WHERE id = 'tbl_01'").Scan(&tableStatus, &activeSession); err != nil {
		t.Fatalf("query table: %v", err)
	}
	if tableStatus != "available" || activeSession != "" {
		t.Fatalf("expected table released, got status=%s active=%s", tableStatus, activeSession)
	}
	detail, err := store.GetInvoiceDetail(session.Session.InvoiceID)
	if err != nil {
		t.Fatalf("get session invoice: %v", err)
	}
	if detail.Summary.Status != "closed" || len(detail.Notifications) != 1 || detail.Notifications[0].Status != "ready" {
		t.Fatalf("expected closed invoice with ready receipt queue, got status=%s notifications=%#v", detail.Summary.Status, detail.Notifications)
	}
}

func TestStaffDirectoryAndTableAssignmentHardening(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	workspace, err := store.SaveStaff(StaffInput{
		Name:   "Maya Floor Lead",
		Role:   "runner",
		PIN:    "2468",
		Status: "active",
	})
	if err != nil {
		t.Fatalf("save staff: %v", err)
	}
	var maya StaffMember
	for _, member := range workspace.Staff {
		if member.Name == "Maya Floor Lead" {
			maya = member
		}
	}
	if maya.ID == "" {
		t.Fatalf("expected saved staff in workspace")
	}
	if maya.PINHash != "stored" {
		t.Fatalf("expected redacted PIN marker, got %q", maya.PINHash)
	}
	if _, err := store.SaveStaff(StaffInput{Name: "Bad Role", Role: "pirate", Status: "active"}); err == nil {
		t.Fatalf("expected unsupported role to fail")
	}
	if _, err := store.OpenOrderSession(OpenOrderSessionInput{TableID: "tbl_01", GuestCount: 2}); err == nil {
		t.Fatalf("expected missing staff assignment to fail")
	}
	session, err := store.OpenOrderSession(OpenOrderSessionInput{TableID: "tbl_01", WaiterID: maya.ID, GuestCount: 3})
	if err != nil {
		t.Fatalf("open table with saved staff: %v", err)
	}
	if session.Session.WaiterName != "Maya Floor Lead" || session.Session.GuestCount != 3 {
		t.Fatalf("expected table assignment to saved staff, got %#v", session.Session)
	}
	if _, err := store.SaveStaff(StaffInput{ID: maya.ID, Name: maya.Name, Role: maya.Role, Status: "inactive"}); err != nil {
		t.Fatalf("archive staff: %v", err)
	}
	if _, err := store.OpenOrderSession(OpenOrderSessionInput{TableID: "tbl_02", WaiterID: maya.ID, GuestCount: 2}); err == nil {
		t.Fatalf("expected inactive staff assignment to fail")
	}
}

func TestKitchenStatusPropagatesToTableSession(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	session, err := store.OpenOrderSession(OpenOrderSessionInput{TableID: "tbl_01", WaiterID: "staff_waiter", GuestCount: 2})
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	session, err = store.AddOrderSessionLine(AddOrderSessionLineInput{
		SessionID: session.Session.ID,
		ItemID:    "item_latte",
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("add session line: %v", err)
	}
	session, err = store.SendOrderSessionKOT(session.Session.ID)
	if err != nil {
		t.Fatalf("send session kot: %v", err)
	}
	if len(session.Lines) != 1 || session.Lines[0].KOTStatus != "queued" {
		t.Fatalf("expected table line to show queued KOT status, got %#v", session.Lines)
	}

	var ticketID string
	if err := store.db.QueryRow("SELECT id FROM kitchen_tickets WHERE sale_id = ? AND route_id = 'route_bar'", session.Session.InvoiceID).Scan(&ticketID); err != nil {
		t.Fatalf("query barista ticket: %v", err)
	}
	if _, err := store.UpdateKOTStatus(ticketID, "ready"); err != nil {
		t.Fatalf("update kot status: %v", err)
	}
	detail, err := store.GetOrderSessionDetail(session.Session.ID)
	if err != nil {
		t.Fatalf("get session detail: %v", err)
	}
	if detail.Lines[0].KOTStatus != "ready" {
		t.Fatalf("expected kitchen status to propagate to table line, got %s", detail.Lines[0].KOTStatus)
	}
	if detail.Lines[0].Status != "open" {
		t.Fatalf("billing line status should remain open, got %s", detail.Lines[0].Status)
	}
	sessions, err := store.orderSessions()
	if err != nil {
		t.Fatalf("query session summaries: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ReadyLineCount != 1 || sessions[0].LineCount != 1 {
		t.Fatalf("expected ready count in live table summary, got %#v", sessions)
	}
}

func TestPaymentRequestAndPrintJobLifecycle(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	invoice := saveDraftOrder(t, store, []SaleLineInput{{ItemID: "item_latte", Quantity: 1}})
	request, err := store.CreatePaymentRequest(invoice.ID)
	if err != nil {
		t.Fatalf("create payment request: %v", err)
	}
	if request.Status != "ready" || request.CheckoutURL == "" {
		t.Fatalf("expected ready Razorpay-shaped request, got %#v", request)
	}
	if _, err := store.MarkPaymentRequestPaid(request.ID); err != nil {
		t.Fatalf("mark payment paid: %v", err)
	}
	var status string
	if err := store.db.QueryRow("SELECT status FROM payment_requests WHERE id = ?", request.ID).Scan(&status); err != nil {
		t.Fatalf("query payment request: %v", err)
	}
	if status != "paid" {
		t.Fatalf("expected paid request, got %s", status)
	}
	job, err := store.QueuePrintJob("invoice", invoice.ID)
	if err != nil {
		t.Fatalf("queue print job: %v", err)
	}
	if job.Status != "ready" || job.Payload == "" {
		t.Fatalf("expected ready print job payload, got %#v", job)
	}
	if job.Target == "" {
		t.Fatalf("expected print target to be captured")
	}
	if _, err := store.MarkPrintJobFailed(job.ID, "paper out"); err != nil {
		t.Fatalf("mark print failed: %v", err)
	}
	var jobStatus, lastError string
	var attempts int
	if err := store.db.QueryRow("SELECT status, attempts, last_error FROM print_jobs WHERE id = ?", job.ID).Scan(&jobStatus, &attempts, &lastError); err != nil {
		t.Fatalf("query failed job: %v", err)
	}
	if jobStatus != "failed" || attempts != 1 || lastError != "paper out" {
		t.Fatalf("expected failed print job with attempt, got status=%s attempts=%d error=%s", jobStatus, attempts, lastError)
	}
	if _, err := store.RetryPrintJob(job.ID); err != nil {
		t.Fatalf("retry print: %v", err)
	}
	if err := store.db.QueryRow("SELECT status, last_error FROM print_jobs WHERE id = ?", job.ID).Scan(&jobStatus, &lastError); err != nil {
		t.Fatalf("query retried job: %v", err)
	}
	if jobStatus != "ready" || lastError != "" {
		t.Fatalf("expected ready retried job, got status=%s error=%s", jobStatus, lastError)
	}
	if _, err := store.MarkPrintJobPrinted(job.ID); err != nil {
		t.Fatalf("mark print printed: %v", err)
	}
	if err := store.db.QueryRow("SELECT status, attempts FROM print_jobs WHERE id = ?", job.ID).Scan(&jobStatus, &attempts); err != nil {
		t.Fatalf("query printed job: %v", err)
	}
	if jobStatus != "printed" || attempts != 2 {
		t.Fatalf("expected printed job with second attempt, got status=%s attempts=%d", jobStatus, attempts)
	}

	secondInvoice := saveDraftOrder(t, store, []SaleLineInput{{ItemID: "item_cappuccino", Quantity: 1}})
	secondRequest, err := store.CreatePaymentRequest(secondInvoice.ID)
	if err != nil {
		t.Fatalf("create second payment request: %v", err)
	}
	if _, err := store.MarkPaymentRequestFailed(secondRequest.ID, "customer cancelled"); err != nil {
		t.Fatalf("mark payment failed: %v", err)
	}
	if err := store.db.QueryRow("SELECT status FROM payment_requests WHERE id = ?", secondRequest.ID).Scan(&status); err != nil {
		t.Fatalf("query failed request: %v", err)
	}
	if status != "failed" {
		t.Fatalf("expected failed request, got %s", status)
	}
	if _, err := store.CancelPaymentRequest(secondRequest.ID); err != nil {
		t.Fatalf("cancel payment request: %v", err)
	}
	if err := store.db.QueryRow("SELECT status FROM payment_requests WHERE id = ?", secondRequest.ID).Scan(&status); err != nil {
		t.Fatalf("query cancelled request: %v", err)
	}
	if status != "cancelled" {
		t.Fatalf("expected cancelled request, got %s", status)
	}
}

func TestPurchaseOrderReceivingCreatesDebitNoteAndMovement(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	beforeTomato := ingredientQty(t, store, "ing_tomato")
	if _, err := store.CreatePurchaseOrder(PurchaseOrderInput{
		VendorID: "vendor_fresh_farm",
		Lines: []PurchaseOrderLineInput{{
			IngredientID: "ing_tomato",
			OrderedQty:   5000,
			UnitCost:     0.08,
		}},
	}); err != nil {
		t.Fatalf("create po: %v", err)
	}
	workspace, err := store.GetPilotWorkspace()
	if err != nil {
		t.Fatalf("get workspace: %v", err)
	}
	po := workspace.PurchaseOrders[0]
	if _, err := store.ReceivePurchaseOrder(ReceivePurchaseOrderInput{
		PurchaseOrderID: po.ID,
		Lines: []ReceivePurchaseOrderLineInput{{
			LineID:          po.Lines[0].ID,
			AcceptedQty:     4200,
			RejectedQty:     800,
			RejectionReason: "Soft batch",
		}},
	}); err != nil {
		t.Fatalf("receive po: %v", err)
	}
	afterTomato := ingredientQty(t, store, "ing_tomato")
	if afterTomato != beforeTomato+4200 {
		t.Fatalf("expected only accepted stock to enter inventory, before=%f after=%f", beforeTomato, afterTomato)
	}
	var debitNotes, messages int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM vendor_debit_notes WHERE purchase_order_id = ?", po.ID).Scan(&debitNotes); err != nil {
		t.Fatalf("query debit notes: %v", err)
	}
	if err := store.db.QueryRow("SELECT COUNT(*) FROM notification_queue WHERE invoice_id = ? AND template = 'vendor_debit_note'", po.ID).Scan(&messages); err != nil {
		t.Fatalf("query vendor messages: %v", err)
	}
	if debitNotes != 1 || messages != 1 {
		t.Fatalf("expected debit note and WhatsApp queue record, debitNotes=%d messages=%d", debitNotes, messages)
	}
}

func TestCustomIngredientAndMultiLinePurchaseOrder(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.SaveIngredient(IngredientInput{
		Name:             "Fresh Basil",
		Unit:             "g",
		PurchaseUnit:     "bunch",
		PurchaseToUsage:  250,
		OnHandQty:        500,
		ReorderPoint:     200,
		LastPurchaseCost: 120,
	}); err != nil {
		t.Fatalf("save ingredient: %v", err)
	}
	if _, err := store.CreatePurchaseOrder(PurchaseOrderInput{
		VendorID: "vendor_fresh_farm",
		Lines: []PurchaseOrderLineInput{
			{IngredientID: "ing_fresh_basil", OrderedQty: 750, UnitCost: 0.5},
			{IngredientID: "ing_tomato", OrderedQty: 1500, UnitCost: 0.09},
		},
	}); err != nil {
		t.Fatalf("create custom po: %v", err)
	}
	workspace, err := store.GetPilotWorkspace()
	if err != nil {
		t.Fatalf("get workspace: %v", err)
	}
	if len(workspace.PurchaseOrders) == 0 || len(workspace.PurchaseOrders[0].Lines) != 2 {
		t.Fatalf("expected custom two-line purchase order, got %#v", workspace.PurchaseOrders)
	}
	var basilUnit string
	if err := store.db.QueryRow("SELECT purchase_unit FROM ingredients WHERE id = 'ing_fresh_basil'").Scan(&basilUnit); err != nil {
		t.Fatalf("query basil: %v", err)
	}
	if basilUnit != "bunch" {
		t.Fatalf("expected custom ingredient units, got %s", basilUnit)
	}
}

func TestDayCloseAndAccountingShell(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	invoice := closeDraftOrder(t, store, []SaleLineInput{{ItemID: "item_latte", Quantity: 1}})
	summary, err := store.CloseDay(DayCloseInput{BusinessDate: todayDate(), CashCounted: invoice.Total, Notes: "Test close"})
	if err != nil {
		t.Fatalf("close day: %v", err)
	}
	if summary.SalesTotal <= 0 || math.Abs(summary.CashVariance) > 0.01 {
		t.Fatalf("unexpected day close summary: %#v", summary)
	}
	snapshot, err := store.RebuildAccountingShell()
	if err != nil {
		t.Fatalf("rebuild accounting: %v", err)
	}
	if snapshot.VoucherCount == 0 || snapshot.SalesTotal <= 0 || len(snapshot.TrialBalance) == 0 {
		t.Fatalf("expected accounting snapshot, got %#v", snapshot)
	}
}

func TestMenuImportAndPriceUpdate(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	result, err := store.ImportMenuCSV(MenuImportInput{Text: "Name,Category,Price,Route,Tax,Cost\nIced Americano,Coffee,150,Barista,5%,42\nPaneer Wrap,Food,240,Kitchen,5%,92"})
	if err != nil {
		t.Fatalf("import menu csv: %v", err)
	}
	if result.Imported != 2 {
		t.Fatalf("expected two imported rows, got %#v", result)
	}
	items, err := store.menuItems()
	if err != nil {
		t.Fatalf("menu items: %v", err)
	}
	var americano MenuItem
	for _, item := range items {
		if item.Name == "Iced Americano" {
			americano = item
		}
	}
	if americano.ID == "" || americano.RouteID != "route_bar" {
		t.Fatalf("expected imported coffee routed to barista, got %#v", americano)
	}
	if _, err := store.SaveMenuItem(MenuItemInput{
		ID:       americano.ID,
		Name:     americano.Name,
		Category: americano.Category,
		Price:    140,
		Cost:     americano.Cost,
		Status:   americano.Status,
		RouteID:  americano.RouteID,
		TaxRate:  americano.TaxRate,
	}); err != nil {
		t.Fatalf("save menu price: %v", err)
	}
	var price float64
	if err := store.db.QueryRow("SELECT price FROM menu_items WHERE id = ?", americano.ID).Scan(&price); err != nil {
		t.Fatalf("query updated price: %v", err)
	}
	if price != 140 {
		t.Fatalf("expected updated price 140, got %f", price)
	}
}

func TestMenuGroupsModifiersAndAvailability(t *testing.T) {
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.SaveMenuCategory(MenuCategoryInput{Name: "Desserts", SortOrder: 10, Status: "active"}); err != nil {
		t.Fatalf("save category: %v", err)
	}
	if _, err := store.SaveMenuItem(MenuItemInput{
		Name:     "Basque Cheesecake",
		Category: "Desserts",
		Price:    280,
		Cost:     96,
		Status:   "active",
		RouteID:  "route_kitchen",
		TaxRate:  0.05,
	}); err != nil {
		t.Fatalf("save dessert item: %v", err)
	}

	workspace, err := store.SaveMenuModifier(MenuModifierInput{
		Name:       "Berry compote",
		PriceDelta: 40,
		RouteID:    "route_kitchen",
		Status:     "active",
	})
	if err != nil {
		t.Fatalf("save modifier: %v", err)
	}
	var itemID, modifierID string
	items, err := store.menuItems()
	if err != nil {
		t.Fatalf("menu items: %v", err)
	}
	for _, item := range items {
		if item.Name == "Basque Cheesecake" {
			itemID = item.ID
		}
	}
	for _, modifier := range workspace.Modifiers {
		if modifier.Name == "Berry compote" {
			modifierID = modifier.ID
		}
	}
	if itemID == "" || modifierID == "" {
		t.Fatalf("expected menu item and modifier ids, item=%s modifier=%s", itemID, modifierID)
	}
	if _, err := store.LinkModifierToItem(itemID, modifierID); err != nil {
		t.Fatalf("link modifier: %v", err)
	}
	if _, err := store.SaveMenuItem(MenuItemInput{
		ID:       itemID,
		Name:     "Basque Cheesecake",
		Category: "Desserts",
		Price:    280,
		Cost:     96,
		Status:   "out_of_stock",
		RouteID:  "route_kitchen",
		TaxRate:  0.05,
	}); err != nil {
		t.Fatalf("mark item out of stock: %v", err)
	}

	workspace, err = store.GetPilotWorkspace()
	if err != nil {
		t.Fatalf("pilot workspace: %v", err)
	}
	if len(workspace.Categories) == 0 || len(workspace.Modifiers) == 0 || len(workspace.ItemModifiers) == 0 {
		t.Fatalf("expected category/modifier/link to be visible in workspace, got %#v", workspace)
	}
	var status string
	if err := store.db.QueryRow("SELECT status FROM menu_items WHERE id = ?", itemID).Scan(&status); err != nil {
		t.Fatalf("query item status: %v", err)
	}
	if status != "out_of_stock" {
		t.Fatalf("expected out_of_stock status, got %s", status)
	}
}

func TestPrinterConnectionAndExports(t *testing.T) {
	exportPath := t.TempDir()
	t.Setenv("NEXUS_EXPORT_DIR", exportPath)
	store := openTestStore(t)
	defer store.Close()

	if _, err := store.SavePrinterConnection(PrinterConnectionInput{Target: "tcp://10.0.0.42:9100", DisplayName: "Receipt Printer"}); err != nil {
		t.Fatalf("save printer: %v", err)
	}
	var target, health string
	if err := store.db.QueryRow("SELECT base_url, health_status FROM integration_settings WHERE provider = 'escpos_printer'").Scan(&target, &health); err != nil {
		t.Fatalf("query printer: %v", err)
	}
	if target != "tcp://10.0.0.42:9100" || health != "ready_for_test" {
		t.Fatalf("unexpected printer settings target=%s health=%s", target, health)
	}

	invoice := closeDraftOrder(t, store, []SaleLineInput{{ItemID: "item_latte", Quantity: 1}})
	pdf, err := store.ExportInvoicePDF(invoice.ID)
	if err != nil {
		t.Fatalf("export invoice pdf: %v", err)
	}
	data, err := os.ReadFile(pdf.Path)
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF-") || pdf.Bytes <= 0 {
		t.Fatalf("expected pdf export, got %#v", pdf)
	}

	invoicesCSV, err := store.ExportInvoicesCSV(InvoiceFilter{})
	if err != nil {
		t.Fatalf("export invoices csv: %v", err)
	}
	csvData, err := os.ReadFile(invoicesCSV.Path)
	if err != nil {
		t.Fatalf("read invoices csv: %v", err)
	}
	if !strings.Contains(string(csvData), invoice.InvoiceNumber) {
		t.Fatalf("expected invoice number in csv")
	}

	accountingCSV, err := store.ExportAccountingCSV()
	if err != nil {
		t.Fatalf("export accounting csv: %v", err)
	}
	accountingData, err := os.ReadFile(accountingCSV.Path)
	if err != nil {
		t.Fatalf("read accounting csv: %v", err)
	}
	if !strings.Contains(string(accountingData), "Food And Beverage Sales") {
		t.Fatalf("expected accounting ledger in csv")
	}

	dayClosePDF, err := store.ExportDayClosePDF(todayDate())
	if err != nil {
		t.Fatalf("export day close pdf: %v", err)
	}
	dayCloseData, err := os.ReadFile(dayClosePDF.Path)
	if err != nil {
		t.Fatalf("read day close pdf: %v", err)
	}
	if !strings.HasPrefix(string(dayCloseData), "%PDF-") || dayClosePDF.Bytes <= 0 {
		t.Fatalf("expected day close pdf export, got %#v", dayClosePDF)
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "nexus.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return store
}

func ingredientQty(t *testing.T, store *Store, id string) float64 {
	t.Helper()
	var qty float64
	if err := store.db.QueryRow("SELECT on_hand_qty FROM ingredients WHERE id = ?", id).Scan(&qty); err != nil {
		t.Fatalf("query ingredient qty: %v", err)
	}
	return qty
}

func ingredientWaste(t *testing.T, store *Store, id string) float64 {
	t.Helper()
	var waste float64
	if err := store.db.QueryRow("SELECT waste_factor FROM ingredients WHERE id = ?", id).Scan(&waste); err != nil {
		t.Fatalf("query ingredient waste: %v", err)
	}
	return waste
}

func saveDraftOrder(t *testing.T, store *Store, lines []SaleLineInput) InvoiceSummary {
	t.Helper()
	if _, err := store.SaveOrder(SaleInput{
		CustomerName:  "Test Guest",
		CustomerPhone: "+917000000000",
		Channel:       "counter",
		OrderType:     "dine_in",
		TableName:     "T-01",
		PaymentMethod: "cash",
		Lines:         lines,
	}); err != nil {
		t.Fatalf("save draft: %v", err)
	}
	invoices, err := store.GetInvoices(InvoiceFilter{})
	if err != nil {
		t.Fatalf("get invoices: %v", err)
	}
	if len(invoices) == 0 {
		t.Fatalf("expected invoice")
	}
	return invoices[0]
}

func closeDraftOrder(t *testing.T, store *Store, lines []SaleLineInput) InvoiceSummary {
	t.Helper()
	invoice := saveDraftOrder(t, store, lines)
	if _, err := store.CloseInvoice(CloseInvoiceInput{InvoiceID: invoice.ID, PaymentMethod: "cash"}); err != nil {
		t.Fatalf("close draft: %v", err)
	}
	detail, err := store.GetInvoiceDetail(invoice.ID)
	if err != nil {
		t.Fatalf("get closed detail: %v", err)
	}
	return detail.Summary
}
