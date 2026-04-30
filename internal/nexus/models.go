package nexus

type Restaurant struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Website    string `json:"website"`
	BrandVoice string `json:"brandVoice"`
	CreatedAt  string `json:"createdAt"`
}

type Metrics struct {
	OrdersToday      int     `json:"ordersToday"`
	SalesToday       float64 `json:"salesToday"`
	StockAlerts      int     `json:"stockAlerts"`
	PendingSyncItems int     `json:"pendingSyncItems"`
	CustomerCount    int     `json:"customerCount"`
	AverageOrder     float64 `json:"averageOrder"`
	OpenKOTs         int     `json:"openKots"`
}

type Ingredient struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Unit             string  `json:"unit"`
	PurchaseUnit     string  `json:"purchaseUnit"`
	PurchaseToUsage  float64 `json:"purchaseToUsage"`
	OnHandQty        float64 `json:"onHandQty"`
	ReorderPoint     float64 `json:"reorderPoint"`
	WasteFactor      float64 `json:"wasteFactor"`
	LastPurchaseCost float64 `json:"lastPurchaseCost"`
	LastAuditAt      string  `json:"lastAuditAt"`
}

type IngredientInput struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Unit             string  `json:"unit"`
	PurchaseUnit     string  `json:"purchaseUnit"`
	PurchaseToUsage  float64 `json:"purchaseToUsage"`
	OnHandQty        float64 `json:"onHandQty"`
	ReorderPoint     float64 `json:"reorderPoint"`
	WasteFactor      float64 `json:"wasteFactor"`
	LastPurchaseCost float64 `json:"lastPurchaseCost"`
}

type MenuItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Price     float64 `json:"price"`
	Cost      float64 `json:"cost"`
	Status    string  `json:"status"`
	RouteID   string  `json:"routeId"`
	RouteName string  `json:"routeName"`
	TaxRate   float64 `json:"taxRate"`
}

type MenuItemInput struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Cost     float64 `json:"cost"`
	Status   string  `json:"status"`
	RouteID  string  `json:"routeId"`
	TaxRate  float64 `json:"taxRate"`
}

type MenuImportInput struct {
	Text string `json:"text"`
}

type MenuImportResult struct {
	Imported int      `json:"imported"`
	Updated  int      `json:"updated"`
	Skipped  []string `json:"skipped"`
}

type RecipeComponent struct {
	ItemID              string   `json:"itemId"`
	IngredientID        string   `json:"ingredientId"`
	IngredientName      string   `json:"ingredientName"`
	Unit                string   `json:"unit"`
	PurchaseUnit        string   `json:"purchaseUnit"`
	PurchaseToUsage     float64  `json:"purchaseToUsage"`
	Quantity            float64  `json:"quantity"`
	WasteFactorOverride *float64 `json:"wasteFactorOverride,omitempty"`
}

type RecipeCard struct {
	ItemID     string            `json:"itemId"`
	ItemName   string            `json:"itemName"`
	RouteName  string            `json:"routeName"`
	Components []RecipeComponent `json:"components"`
}

type Customer struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	TotalSpend   float64 `json:"totalSpend"`
	VisitCount   int     `json:"visitCount"`
	LastVisitAt  string  `json:"lastVisitAt"`
	FavoriteItem string  `json:"favoriteItem"`
}

type SaleSummary struct {
	ID              string  `json:"id"`
	InvoiceNumber   string  `json:"invoiceNumber"`
	CustomerName    string  `json:"customerName"`
	CustomerPhone   string  `json:"customerPhone"`
	Channel         string  `json:"channel"`
	OrderType       string  `json:"orderType"`
	TableName       string  `json:"tableName"`
	Subtotal        float64 `json:"subtotal"`
	DiscountTotal   float64 `json:"discountTotal"`
	TaxTotal        float64 `json:"taxTotal"`
	Total           float64 `json:"total"`
	PaymentMethod   string  `json:"paymentMethod"`
	Status          string  `json:"status"`
	SourceInvoiceID string  `json:"sourceInvoiceId"`
	SplitGroupID    string  `json:"splitGroupId"`
	VoidReason      string  `json:"voidReason"`
	RefundReason    string  `json:"refundReason"`
	ApprovedBy      string  `json:"approvedBy"`
	ApprovedAt      string  `json:"approvedAt"`
	CreatedAt       string  `json:"createdAt"`
	ClosedAt        string  `json:"closedAt"`
	KOTSentAt       string  `json:"kotSentAt"`
}

type InvoiceSummary = SaleSummary

type InvoiceLine struct {
	ID        string  `json:"id"`
	InvoiceID string  `json:"invoiceId"`
	ItemID    string  `json:"itemId"`
	ItemName  string  `json:"itemName"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	LineTotal float64 `json:"lineTotal"`
	Notes     string  `json:"notes"`
	RouteID   string  `json:"routeId"`
	RouteName string  `json:"routeName"`
}

type InvoiceEvent struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoiceId"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	Actor     string `json:"actor"`
	CreatedAt string `json:"createdAt"`
}

type PaymentRecord struct {
	ID        string  `json:"id"`
	InvoiceID string  `json:"invoiceId"`
	Method    string  `json:"method"`
	Amount    float64 `json:"amount"`
	Tendered  float64 `json:"tendered"`
	ChangeDue float64 `json:"changeDue"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
}

type RefundRecord struct {
	ID         string  `json:"id"`
	InvoiceID  string  `json:"invoiceId"`
	Amount     float64 `json:"amount"`
	Reason     string  `json:"reason"`
	ApprovedBy string  `json:"approvedBy"`
	ApprovedAt string  `json:"approvedAt"`
	CreatedAt  string  `json:"createdAt"`
}

type NotificationRecord struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoiceId"`
	Channel   string `json:"channel"`
	Recipient string `json:"recipient"`
	Template  string `json:"template"`
	Payload   string `json:"payload"`
	Status    string `json:"status"`
	Error     string `json:"error"`
	CreatedAt string `json:"createdAt"`
	SentAt    string `json:"sentAt"`
	ReadAt    string `json:"readAt"`
}

type InvoiceDetail struct {
	Summary        InvoiceSummary       `json:"summary"`
	Lines          []InvoiceLine        `json:"lines"`
	Payments       []PaymentRecord      `json:"payments"`
	Refunds        []RefundRecord       `json:"refunds"`
	Events         []InvoiceEvent       `json:"events"`
	KitchenTickets []KitchenTicket      `json:"kitchenTickets"`
	Notifications  []NotificationRecord `json:"notifications"`
}

type InvoiceFilter struct {
	Status        string `json:"status"`
	Search        string `json:"search"`
	PaymentMethod string `json:"paymentMethod"`
	DateFrom      string `json:"dateFrom"`
	DateTo        string `json:"dateTo"`
}

type CloseInvoiceInput struct {
	InvoiceID       string  `json:"invoiceId"`
	PaymentMethod   string  `json:"paymentMethod"`
	PaymentTendered float64 `json:"paymentTendered"`
}

type VoidInvoiceInput struct {
	InvoiceID string `json:"invoiceId"`
	PIN       string `json:"pin"`
	Reason    string `json:"reason"`
}

type RefundInvoiceInput struct {
	InvoiceID string  `json:"invoiceId"`
	PIN       string  `json:"pin"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}

type SplitInvoiceInput struct {
	InvoiceID string           `json:"invoiceId"`
	Mode      string           `json:"mode"`
	Lines     []SplitLineInput `json:"lines"`
	Amounts   []float64        `json:"amounts"`
}

type SplitLineInput struct {
	LineID   string `json:"lineId"`
	Quantity int    `json:"quantity"`
}

type ApprovalToken struct {
	Token      string `json:"token"`
	ApprovedBy string `json:"approvedBy"`
	ApprovedAt string `json:"approvedAt"`
	ExpiresAt  string `json:"expiresAt"`
}

type KitchenRoute struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PrinterName string `json:"printerName"`
	Color       string `json:"color"`
}

type KitchenTicket struct {
	ID            string              `json:"id"`
	SaleID        string              `json:"saleId"`
	InvoiceNumber string              `json:"invoiceNumber"`
	TicketNumber  string              `json:"ticketNumber"`
	RouteID       string              `json:"routeId"`
	RouteName     string              `json:"routeName"`
	Status        string              `json:"status"`
	CreatedAt     string              `json:"createdAt"`
	Lines         []KitchenTicketLine `json:"lines"`
}

type KitchenTicketLine struct {
	ID       string `json:"id"`
	TicketID string `json:"ticketId"`
	ItemID   string `json:"itemId"`
	ItemName string `json:"itemName"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes"`
	Status   string `json:"status"`
}

type DeliveryBatch struct {
	ID            string         `json:"id"`
	VendorName    string         `json:"vendorName"`
	InvoiceNumber string         `json:"invoiceNumber"`
	Status        string         `json:"status"`
	CreatedAt     string         `json:"createdAt"`
	Lines         []DeliveryLine `json:"lines"`
}

type DeliveryLine struct {
	ID              string  `json:"id"`
	BatchID         string  `json:"batchId"`
	IngredientID    string  `json:"ingredientId"`
	IngredientName  string  `json:"ingredientName"`
	Unit            string  `json:"unit"`
	OrderedQty      float64 `json:"orderedQty"`
	AcceptedQty     float64 `json:"acceptedQty"`
	RejectedQty     float64 `json:"rejectedQty"`
	UnitCost        float64 `json:"unitCost"`
	RejectionReason string  `json:"rejectionReason"`
	Status          string  `json:"status"`
}

type Signal struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Action   string `json:"action"`
	Priority int    `json:"priority"`
}

type MarketingDraft struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Channel   string `json:"channel"`
	Caption   string `json:"caption"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type Dashboard struct {
	Restaurant      Restaurant       `json:"restaurant"`
	Metrics         Metrics          `json:"metrics"`
	MenuItems       []MenuItem       `json:"menuItems"`
	Ingredients     []Ingredient     `json:"ingredients"`
	KitchenRoutes   []KitchenRoute   `json:"kitchenRoutes"`
	Recipes         []RecipeCard     `json:"recipes"`
	Customers       []Customer       `json:"customers"`
	RecentSales     []SaleSummary    `json:"recentSales"`
	KitchenTickets  []KitchenTicket  `json:"kitchenTickets"`
	Deliveries      []DeliveryBatch  `json:"deliveries"`
	Signals         []Signal         `json:"signals"`
	MarketingDrafts []MarketingDraft `json:"marketingDrafts"`
}

type SaleInput struct {
	CustomerName    string          `json:"customerName"`
	CustomerPhone   string          `json:"customerPhone"`
	Channel         string          `json:"channel"`
	OrderType       string          `json:"orderType"`
	TableName       string          `json:"tableName"`
	DiscountType    string          `json:"discountType"`
	DiscountValue   float64         `json:"discountValue"`
	TaxRate         float64         `json:"taxRate"`
	PaymentMethod   string          `json:"paymentMethod"`
	PaymentTendered float64         `json:"paymentTendered"`
	Lines           []SaleLineInput `json:"lines"`
}

type SaleLineInput struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes"`
}

type DeliveryInput struct {
	VendorName    string              `json:"vendorName"`
	InvoiceNumber string              `json:"invoiceNumber"`
	Lines         []DeliveryLineInput `json:"lines"`
}

type DeliveryLineInput struct {
	IngredientID    string  `json:"ingredientId"`
	OrderedQty      float64 `json:"orderedQty"`
	AcceptedQty     float64 `json:"acceptedQty"`
	RejectedQty     float64 `json:"rejectedQty"`
	UnitCost        float64 `json:"unitCost"`
	RejectionReason string  `json:"rejectionReason"`
}

type ReconcileInput struct {
	IngredientID string  `json:"ingredientId"`
	PhysicalQty  float64 `json:"physicalQty"`
	Note         string  `json:"note"`
}

type RecipeUpdateInput struct {
	ItemID     string                  `json:"itemId"`
	Components []RecipeLineUpdateInput `json:"components"`
}

type RecipeLineUpdateInput struct {
	IngredientID        string   `json:"ingredientId"`
	Quantity            float64  `json:"quantity"`
	WasteFactorOverride *float64 `json:"wasteFactorOverride,omitempty"`
}

type IngredientUpdateInput struct {
	IngredientID     string  `json:"ingredientId"`
	ReorderPoint     float64 `json:"reorderPoint"`
	PurchaseUnit     string  `json:"purchaseUnit"`
	PurchaseToUsage  float64 `json:"purchaseToUsage"`
	LastPurchaseCost float64 `json:"lastPurchaseCost"`
}

type RestaurantSettings struct {
	ID                string  `json:"id"`
	RestaurantName    string  `json:"restaurantName"`
	GSTIN             string  `json:"gstin"`
	LegalName         string  `json:"legalName"`
	Address           string  `json:"address"`
	State             string  `json:"state"`
	TaxMode           string  `json:"taxMode"`
	DefaultTaxRate    float64 `json:"defaultTaxRate"`
	ServiceChargeRate float64 `json:"serviceChargeRate"`
	InvoicePrefix     string  `json:"invoicePrefix"`
	BusinessHours     string  `json:"businessHours"`
	BackupPath        string  `json:"backupPath"`
	ReceiptPrinterID  string  `json:"receiptPrinterId"`
	UpdatedAt         string  `json:"updatedAt"`
}

type IntegrationSetting struct {
	ID               string `json:"id"`
	Provider         string `json:"provider"`
	Mode             string `json:"mode"`
	DisplayName      string `json:"displayName"`
	BaseURL          string `json:"baseUrl"`
	CredentialStatus string `json:"credentialStatus"`
	HealthStatus     string `json:"healthStatus"`
	LastCheckedAt    string `json:"lastCheckedAt"`
	LastError        string `json:"lastError"`
	UpdatedAt        string `json:"updatedAt"`
}

type IntegrationSettingInput struct {
	Provider     string `json:"provider"`
	Mode         string `json:"mode"`
	DisplayName  string `json:"displayName"`
	BaseURL      string `json:"baseUrl"`
	Secret       string `json:"secret"`
	HealthStatus string `json:"healthStatus"`
}

type PrinterConnectionInput struct {
	Target      string `json:"target"`
	DisplayName string `json:"displayName"`
	Mode        string `json:"mode"`
}

type StaffMember struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	PINHash   string `json:"pinHash"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type AuditLogEntry struct {
	ID         string `json:"id"`
	EventType  string `json:"eventType"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Detail     string `json:"detail"`
	Actor      string `json:"actor"`
	CreatedAt  string `json:"createdAt"`
}

type StaffInput struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	PIN    string `json:"pin"`
	Status string `json:"status"`
}

type StaffLoginInput struct {
	StaffID   string `json:"staffId"`
	PIN       string `json:"pin"`
	Workspace string `json:"workspace"`
}

type StaffSession struct {
	StaffID         string   `json:"staffId"`
	Name            string   `json:"name"`
	Role            string   `json:"role"`
	Permissions     []string `json:"permissions"`
	WorkspaceAccess []string `json:"workspaceAccess"`
	IssuedAt        string   `json:"issuedAt"`
	ExpiresAt       string   `json:"expiresAt"`
}

type StaffActionApprovalInput struct {
	StaffID  string `json:"staffId"`
	PIN      string `json:"pin"`
	Action   string `json:"action"`
	TargetID string `json:"targetId"`
}

type MenuCategory struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
	Status    string `json:"status"`
}

type MenuCategoryInput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
	Status    string `json:"status"`
}

type MenuModifier struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	PriceDelta float64 `json:"priceDelta"`
	RouteID    string  `json:"routeId"`
	Status     string  `json:"status"`
}

type MenuModifierInput struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	PriceDelta float64 `json:"priceDelta"`
	RouteID    string  `json:"routeId"`
	Status     string  `json:"status"`
}

type MenuItemModifier struct {
	ItemID     string `json:"itemId"`
	ModifierID string `json:"modifierId"`
}

type FloorSection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}

type FloorSectionInput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}

type DiningTable struct {
	ID              string `json:"id"`
	SectionID       string `json:"sectionId"`
	SectionName     string `json:"sectionName"`
	Name            string `json:"name"`
	Seats           int    `json:"seats"`
	Status          string `json:"status"`
	ActiveSessionID string `json:"activeSessionId"`
}

type DiningTableInput struct {
	ID        string `json:"id"`
	SectionID string `json:"sectionId"`
	Name      string `json:"name"`
	Seats     int    `json:"seats"`
	Status    string `json:"status"`
}

type OrderSession struct {
	ID                 string  `json:"id"`
	TableID            string  `json:"tableId"`
	TableName          string  `json:"tableName"`
	SectionName        string  `json:"sectionName"`
	WaiterID           string  `json:"waiterId"`
	WaiterName         string  `json:"waiterName"`
	GuestCount         int     `json:"guestCount"`
	ServiceMode        string  `json:"serviceMode"`
	Status             string  `json:"status"`
	InvoiceID          string  `json:"invoiceId"`
	Subtotal           float64 `json:"subtotal"`
	TaxTotal           float64 `json:"taxTotal"`
	ServiceCharge      float64 `json:"serviceCharge"`
	Total              float64 `json:"total"`
	OpenedAt           string  `json:"openedAt"`
	ClosedAt           string  `json:"closedAt"`
	LineCount          int     `json:"lineCount"`
	ReadyLineCount     int     `json:"readyLineCount"`
	PreparingLineCount int     `json:"preparingLineCount"`
	QueuedLineCount    int     `json:"queuedLineCount"`
	NotSentLineCount   int     `json:"notSentLineCount"`
}

type OrderSessionLine struct {
	ID            string   `json:"id"`
	SessionID     string   `json:"sessionId"`
	ItemID        string   `json:"itemId"`
	ItemName      string   `json:"itemName"`
	Quantity      int      `json:"quantity"`
	UnitPrice     float64  `json:"unitPrice"`
	ModifierTotal float64  `json:"modifierTotal"`
	LineTotal     float64  `json:"lineTotal"`
	Notes         string   `json:"notes"`
	Status        string   `json:"status"`
	KOTStatus     string   `json:"kotStatus"`
	ModifierIDs   []string `json:"modifierIds"`
	ModifierNames []string `json:"modifierNames"`
}

type OrderSessionEvent struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	Type      string `json:"type"`
	Detail    string `json:"detail"`
	Actor     string `json:"actor"`
	CreatedAt string `json:"createdAt"`
}

type OrderSessionDetail struct {
	Session OrderSession        `json:"session"`
	Lines   []OrderSessionLine  `json:"lines"`
	Events  []OrderSessionEvent `json:"events"`
}

type OpenOrderSessionInput struct {
	TableID     string `json:"tableId"`
	WaiterID    string `json:"waiterId"`
	GuestCount  int    `json:"guestCount"`
	ServiceMode string `json:"serviceMode"`
}

type AddOrderSessionLineInput struct {
	SessionID   string   `json:"sessionId"`
	ItemID      string   `json:"itemId"`
	Quantity    int      `json:"quantity"`
	Notes       string   `json:"notes"`
	ModifierIDs []string `json:"modifierIds"`
}

type CloseOrderSessionInput struct {
	SessionID       string  `json:"sessionId"`
	CustomerName    string  `json:"customerName"`
	CustomerPhone   string  `json:"customerPhone"`
	PaymentMethod   string  `json:"paymentMethod"`
	PaymentTendered float64 `json:"paymentTendered"`
}

type SessionLineVoidInput struct {
	SessionID string `json:"sessionId"`
	LineID    string `json:"lineId"`
	StaffID   string `json:"staffId"`
	PIN       string `json:"pin"`
	Reason    string `json:"reason"`
}

type TableMoveInput struct {
	SessionID     string `json:"sessionId"`
	TargetTableID string `json:"targetTableId"`
}

type MergeTableInput struct {
	SourceSessionID string `json:"sourceSessionId"`
	TargetSessionID string `json:"targetSessionId"`
}

type DayCloseSummary struct {
	ID            string  `json:"id"`
	BusinessDate  string  `json:"businessDate"`
	ClosedAt      string  `json:"closedAt"`
	Status        string  `json:"status"`
	SalesTotal    float64 `json:"salesTotal"`
	CashExpected  float64 `json:"cashExpected"`
	CashCounted   float64 `json:"cashCounted"`
	CashVariance  float64 `json:"cashVariance"`
	UPITotal      float64 `json:"upiTotal"`
	CardTotal     float64 `json:"cardTotal"`
	RazorpayTotal float64 `json:"razorpayTotal"`
	RefundTotal   float64 `json:"refundTotal"`
	VoidCount     int     `json:"voidCount"`
	DiscountTotal float64 `json:"discountTotal"`
	Notes         string  `json:"notes"`
}

type DayCloseInput struct {
	BusinessDate string  `json:"businessDate"`
	CashCounted  float64 `json:"cashCounted"`
	Notes        string  `json:"notes"`
	StaffID      string  `json:"staffId"`
	PIN          string  `json:"pin"`
}

type PaymentRequest struct {
	ID          string  `json:"id"`
	InvoiceID   string  `json:"invoiceId"`
	Provider    string  `json:"provider"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Reference   string  `json:"reference"`
	CheckoutURL string  `json:"checkoutUrl"`
	QRPayload   string  `json:"qrPayload"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type PrintJob struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	ReferenceID string `json:"referenceId"`
	PrinterID   string `json:"printerId"`
	Target      string `json:"target"`
	Payload     string `json:"payload"`
	Status      string `json:"status"`
	Attempts    int    `json:"attempts"`
	LastError   string `json:"lastError"`
	CreatedAt   string `json:"createdAt"`
	PrintedAt   string `json:"printedAt"`
}

type Vendor struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	GSTIN        string  `json:"gstin"`
	PaymentTerms string  `json:"paymentTerms"`
	QualityScore float64 `json:"qualityScore"`
	Status       string  `json:"status"`
}

type VendorInput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	GSTIN        string `json:"gstin"`
	PaymentTerms string `json:"paymentTerms"`
	Status       string `json:"status"`
}

type PurchaseOrder struct {
	ID            string              `json:"id"`
	VendorID      string              `json:"vendorId"`
	VendorName    string              `json:"vendorName"`
	PONumber      string              `json:"poNumber"`
	Status        string              `json:"status"`
	ExpectedDate  string              `json:"expectedDate"`
	Subtotal      float64             `json:"subtotal"`
	RejectedTotal float64             `json:"rejectedTotal"`
	CreatedAt     string              `json:"createdAt"`
	Lines         []PurchaseOrderLine `json:"lines"`
}

type PurchaseOrderLine struct {
	ID              string  `json:"id"`
	PurchaseOrderID string  `json:"purchaseOrderId"`
	IngredientID    string  `json:"ingredientId"`
	IngredientName  string  `json:"ingredientName"`
	OrderedQty      float64 `json:"orderedQty"`
	AcceptedQty     float64 `json:"acceptedQty"`
	RejectedQty     float64 `json:"rejectedQty"`
	UnitCost        float64 `json:"unitCost"`
	RejectionReason string  `json:"rejectionReason"`
}

type PurchaseOrderInput struct {
	VendorID     string                   `json:"vendorId"`
	ExpectedDate string                   `json:"expectedDate"`
	Lines        []PurchaseOrderLineInput `json:"lines"`
}

type PurchaseOrderLineInput struct {
	IngredientID string  `json:"ingredientId"`
	OrderedQty   float64 `json:"orderedQty"`
	UnitCost     float64 `json:"unitCost"`
}

type ReceivePurchaseOrderInput struct {
	PurchaseOrderID string                          `json:"purchaseOrderId"`
	Lines           []ReceivePurchaseOrderLineInput `json:"lines"`
}

type ReceivePurchaseOrderLineInput struct {
	LineID          string  `json:"lineId"`
	AcceptedQty     float64 `json:"acceptedQty"`
	RejectedQty     float64 `json:"rejectedQty"`
	RejectionReason string  `json:"rejectionReason"`
}

type VendorDebitNote struct {
	ID              string  `json:"id"`
	VendorID        string  `json:"vendorId"`
	VendorName      string  `json:"vendorName"`
	PurchaseOrderID string  `json:"purchaseOrderId"`
	Amount          float64 `json:"amount"`
	Reason          string  `json:"reason"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"createdAt"`
}

type AccountingBalanceRow struct {
	LedgerID      string  `json:"ledgerId"`
	LedgerName    string  `json:"ledgerName"`
	GroupName     string  `json:"groupName"`
	NormalBalance string  `json:"normalBalance"`
	DebitTotal    float64 `json:"debitTotal"`
	CreditTotal   float64 `json:"creditTotal"`
	Balance       float64 `json:"balance"`
	BalanceSide   string  `json:"balanceSide"`
}

type AccountingSnapshot struct {
	SalesTotal       float64                `json:"salesTotal"`
	TaxPayable       float64                `json:"taxPayable"`
	CashAndBank      float64                `json:"cashAndBank"`
	RazorpayClearing float64                `json:"razorpayClearing"`
	VendorPayables   float64                `json:"vendorPayables"`
	RefundTotal      float64                `json:"refundTotal"`
	VoucherCount     int                    `json:"voucherCount"`
	TrialBalance     []AccountingBalanceRow `json:"trialBalance"`
}

type AdminAnalytics struct {
	GeneratedAt        string                    `json:"generatedAt"`
	RangeKey           string                    `json:"rangeKey"`
	RangeLabel         string                    `json:"rangeLabel"`
	DemoFallback       bool                      `json:"demoFallback"`
	Executive          []AnalyticsMetric         `json:"executive"`
	SalesTrend         []AnalyticsPoint          `json:"salesTrend"`
	HourlyHeatmap      []AnalyticsPoint          `json:"hourlyHeatmap"`
	TenderMix          []AnalyticsPoint          `json:"tenderMix"`
	CategoryMix        []AnalyticsPoint          `json:"categoryMix"`
	ItemVelocity       []AnalyticsPoint          `json:"itemVelocity"`
	ContributionMargin []AnalyticsPoint          `json:"contributionMargin"`
	InventoryHealth    []InventoryHealthRow      `json:"inventoryHealth"`
	KitchenPerformance []KitchenPerformanceRow   `json:"kitchenPerformance"`
	PurchaseTrend      []AnalyticsPoint          `json:"purchaseTrend"`
	ItemMatrix         []ItemMatrixBucket        `json:"itemMatrix"`
	Settlement         SettlementHealth          `json:"settlement"`
	Exceptions         []AnalyticsException      `json:"exceptions"`
	Recommendations    []AnalyticsRecommendation `json:"recommendations"`
	SnapshotStatus     AnalyticsSnapshotStatus   `json:"snapshotStatus"`
}

type AnalyticsMetric struct {
	ID     string  `json:"id"`
	Label  string  `json:"label"`
	Value  float64 `json:"value"`
	Format string  `json:"format"`
	Detail string  `json:"detail"`
	Tone   string  `json:"tone"`
}

type AnalyticsPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Count int     `json:"count"`
	Tone  string  `json:"tone"`
}

type InventoryHealthRow struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	OnHandQty    float64 `json:"onHandQty"`
	ReorderPoint float64 `json:"reorderPoint"`
	Unit         string  `json:"unit"`
	RiskScore    float64 `json:"riskScore"`
	Status       string  `json:"status"`
}

type KitchenPerformanceRow struct {
	RouteName     string  `json:"routeName"`
	OpenTickets   int     `json:"openTickets"`
	ReadyTickets  int     `json:"readyTickets"`
	ServedTickets int     `json:"servedTickets"`
	AverageAgeMin float64 `json:"averageAgeMin"`
	OldestAgeMin  float64 `json:"oldestAgeMin"`
}

type ItemMatrixBucket struct {
	ID          string            `json:"id"`
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Items       []ItemMatrixEntry `json:"items"`
}

type ItemMatrixEntry struct {
	ItemID    string  `json:"itemId"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Quantity  int     `json:"quantity"`
	Sales     float64 `json:"sales"`
	Margin    float64 `json:"margin"`
	MarginPct float64 `json:"marginPct"`
}

type SettlementHealth struct {
	CashExpected     float64 `json:"cashExpected"`
	CashVariance     float64 `json:"cashVariance"`
	UPITotal         float64 `json:"upiTotal"`
	CardTotal        float64 `json:"cardTotal"`
	RazorpayTotal    float64 `json:"razorpayTotal"`
	RazorpayClearing float64 `json:"razorpayClearing"`
	TaxPayable       float64 `json:"taxPayable"`
	VendorPayables   float64 `json:"vendorPayables"`
	RefundTotal      float64 `json:"refundTotal"`
}

type AnalyticsException struct {
	ID       string  `json:"id"`
	Kind     string  `json:"kind"`
	Title    string  `json:"title"`
	Detail   string  `json:"detail"`
	Severity string  `json:"severity"`
	Value    float64 `json:"value"`
}

type AnalyticsRecommendation struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Priority int    `json:"priority"`
	Page     string `json:"page"`
}

type AnalyticsSnapshotStatus struct {
	DailyRows  int    `json:"dailyRows"`
	ItemRows   int    `json:"itemRows"`
	HourlyRows int    `json:"hourlyRows"`
	UpdatedAt  string `json:"updatedAt"`
}

type ExportResult struct {
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	FileName string `json:"fileName"`
	MIMEType string `json:"mimeType"`
	Bytes    int64  `json:"bytes"`
}

type BackupInput struct {
	Destination string `json:"destination"`
}

type SyncStatus struct {
	PendingCount    int    `json:"pendingCount"`
	SyncedCount     int    `json:"syncedCount"`
	FailedCount     int    `json:"failedCount"`
	OldestPendingAt string `json:"oldestPendingAt"`
	LastSyncedAt    string `json:"lastSyncedAt"`
	LastError       string `json:"lastError"`
	DatabasePath    string `json:"databasePath"`
	DatabaseBytes   int64  `json:"databaseBytes"`
	WALBytes        int64  `json:"walBytes"`
	UpdatedAt       string `json:"updatedAt"`
}

type PilotWorkspace struct {
	Settings        RestaurantSettings   `json:"settings"`
	Integrations    []IntegrationSetting `json:"integrations"`
	Staff           []StaffMember        `json:"staff"`
	Categories      []MenuCategory       `json:"categories"`
	Modifiers       []MenuModifier       `json:"modifiers"`
	ItemModifiers   []MenuItemModifier   `json:"itemModifiers"`
	FloorSections   []FloorSection       `json:"floorSections"`
	Tables          []DiningTable        `json:"tables"`
	OrderSessions   []OrderSession       `json:"orderSessions"`
	PaymentRequests []PaymentRequest     `json:"paymentRequests"`
	PrintJobs       []PrintJob           `json:"printJobs"`
	Vendors         []Vendor             `json:"vendors"`
	PurchaseOrders  []PurchaseOrder      `json:"purchaseOrders"`
	DebitNotes      []VendorDebitNote    `json:"debitNotes"`
	DayClose        DayCloseSummary      `json:"dayClose"`
	Accounting      AccountingSnapshot   `json:"accounting"`
	AuditLog        []AuditLogEntry      `json:"auditLog"`
}
