import {useEffect, useMemo, useState, type CSSProperties, type DragEvent, type ReactNode} from 'react';
import './App.css';
import {
    AddOrderSessionLine,
    AssignOrderSessionStaff,
    AuthenticateStaff,
    CancelPaymentRequest,
    CloseInvoice,
    CloseOrderSession,
    CloseDay,
    CreatePaymentRequest,
    CreatePurchaseOrder,
    ExportAccountingCSV,
    ExportBackup,
    ExportDayClosePDF,
    ExportInvoicePDF,
    ExportInvoicesCSV,
    GetAdminAnalytics,
    GetDashboard,
    GetPilotWorkspace,
    GetInvoiceDetail,
    GetInvoices,
    GetNotifications,
    GetSyncStatus,
    GetOrderSessionDetail,
    ImportMenuCSV,
    LinkModifierToItem,
    MarkNotificationRead,
    MarkPaymentRequestPaid,
    MarkPaymentRequestFailed,
    MarkPrintJobFailed,
    MarkPrintJobPrinted,
    OpenOrderSession,
    QueuePrintJob,
    ReceiveDelivery,
    ReceivePurchaseOrder,
    ReconcileStock,
    RecordSale,
    RebuildAccountingShell,
    RefundInvoice,
    SaveIngredient,
    SaveOrder,
    SaveIntegrationSetting,
    SaveMenuCategory,
    SaveMenuItem,
    SaveMenuModifier,
    SavePrinterConnection,
    SaveRestaurantSettings,
    SaveStaff,
    SaveVendor,
    SendKOT,
    SendOrderSessionKOT,
    SplitInvoice,
    RetryPrintJob,
    TransferOrderSession,
    UpdateIngredientSettings,
    UpdateKOTStatus,
    UpdateRecipe,
    ValidateManagerPIN,
    VoidInvoice,
    VoidOrderSessionLine,
} from '../wailsjs/go/main/App';
import {nexus} from '../wailsjs/go/models';
import {WindowFullscreen} from '../wailsjs/runtime/runtime';

type WorkspaceKey = 'front' | 'kitchenOps' | 'backoffice' | 'admin';
type PageKey = 'home' | 'counter' | 'floor' | 'orders' | 'invoices' | 'kitchen' | 'inventory' | 'vendors' | 'recipes' | 'customers' | 'marketing' | 'integrations' | 'dayclose' | 'accounting' | 'settings' | 'admin';
type ToastKind = 'success' | 'error' | 'info';

type Metrics = {
    ordersToday: number;
    salesToday: number;
    stockAlerts: number;
    pendingSyncItems: number;
    customerCount: number;
    averageOrder: number;
    openKots: number;
};

type Restaurant = {
    name: string;
    website: string;
    brandVoice: string;
};

type MenuItem = {
    id: string;
    name: string;
    category: string;
    price: number;
    cost: number;
    status: string;
    routeId: string;
    routeName: string;
    taxRate: number;
};

type KitchenRoute = {
    id: string;
    name: string;
    printerName: string;
    color: string;
};

type ExportResult = {
    kind: string;
    path: string;
    fileName: string;
    mimeType: string;
    bytes: number;
};

type MenuImportResult = {
    imported: number;
    updated: number;
    skipped: string[];
};

type Ingredient = {
    id: string;
    name: string;
    unit: string;
    purchaseUnit: string;
    purchaseToUsage: number;
    onHandQty: number;
    reorderPoint: number;
    wasteFactor: number;
    lastPurchaseCost: number;
};

type RecipeComponent = {
    itemId: string;
    ingredientId: string;
    ingredientName: string;
    unit: string;
    purchaseUnit: string;
    purchaseToUsage: number;
    quantity: number;
    wasteFactorOverride?: number;
};

type RecipeCard = {
    itemId: string;
    itemName: string;
    routeName: string;
    components: RecipeComponent[];
};

type Customer = {
    id: string;
    name: string;
    phone: string;
    totalSpend: number;
    visitCount: number;
    favoriteItem: string;
    lastVisitAt: string;
};

type InvoiceSummary = {
    id: string;
    invoiceNumber: string;
    customerName: string;
    customerPhone: string;
    channel: string;
    orderType: string;
    tableName: string;
    subtotal: number;
    discountTotal: number;
    taxTotal: number;
    total: number;
    paymentMethod: string;
    status: string;
    sourceInvoiceId: string;
    splitGroupId: string;
    voidReason: string;
    refundReason: string;
    approvedBy: string;
    approvedAt: string;
    createdAt: string;
    closedAt: string;
    kotSentAt: string;
};

type InvoiceLine = {
    id: string;
    invoiceId: string;
    itemId: string;
    itemName: string;
    quantity: number;
    unitPrice: number;
    lineTotal: number;
    notes: string;
    routeName: string;
};

type PaymentRecord = {
    id: string;
    method: string;
    amount: number;
    tendered: number;
    changeDue: number;
    status: string;
    createdAt: string;
};

type RefundRecord = {
    id: string;
    amount: number;
    reason: string;
    approvedBy: string;
    createdAt: string;
};

type InvoiceEvent = {
    id: string;
    type: string;
    title: string;
    detail: string;
    actor: string;
    createdAt: string;
};

type NotificationRecord = {
    id: string;
    invoiceId: string;
    channel: string;
    recipient: string;
    template: string;
    payload: string;
    status: string;
    error: string;
    createdAt: string;
    readAt: string;
};

type InvoiceDetail = {
    summary: InvoiceSummary;
    lines: InvoiceLine[];
    payments: PaymentRecord[];
    refunds: RefundRecord[];
    events: InvoiceEvent[];
    kitchenTickets: KitchenTicket[];
    notifications: NotificationRecord[];
};

type DeliveryLine = {
    id: string;
    ingredientName: string;
    unit: string;
    acceptedQty: number;
    rejectedQty: number;
};

type DeliveryBatch = {
    id: string;
    vendorName: string;
    invoiceNumber: string;
    status: string;
    lines: DeliveryLine[];
};

type Signal = {
    id: string;
    kind: string;
    title: string;
    detail: string;
    action: string;
    priority: number;
};

type MarketingDraft = {
    id: string;
    title: string;
    channel: string;
    caption: string;
    status: string;
};

type KitchenTicketLine = {
    id: string;
    itemName: string;
    quantity: number;
    notes: string;
    status: string;
};

type KitchenTicket = {
    id: string;
    saleId: string;
    invoiceNumber: string;
    ticketNumber: string;
    routeName: string;
    status: string;
    createdAt: string;
    lines: KitchenTicketLine[];
};

type Dashboard = {
    restaurant: Restaurant;
    metrics: Metrics;
    menuItems: MenuItem[];
    ingredients: Ingredient[];
    kitchenRoutes: KitchenRoute[];
    recipes: RecipeCard[];
    customers: Customer[];
    recentSales: InvoiceSummary[];
    kitchenTickets: KitchenTicket[];
    deliveries: DeliveryBatch[];
    signals: Signal[];
    marketingDrafts: MarketingDraft[];
};

type RestaurantSettings = {
    id: string;
    restaurantName: string;
    gstin: string;
    legalName: string;
    address: string;
    state: string;
    taxMode: string;
    defaultTaxRate: number;
    serviceChargeRate: number;
    invoicePrefix: string;
    businessHours: string;
    backupPath: string;
    receiptPrinterId: string;
    updatedAt: string;
};

type IntegrationSetting = {
    id: string;
    provider: string;
    mode: string;
    displayName: string;
    baseUrl: string;
    credentialStatus: string;
    healthStatus: string;
    lastCheckedAt: string;
    lastError: string;
    updatedAt: string;
};

type StaffMember = {
    id: string;
    name: string;
    role: string;
    pinHash: string;
    status: string;
};

type StaffSession = {
    staffId: string;
    name: string;
    role: string;
    permissions: string[];
    workspaceAccess: WorkspaceKey[];
    issuedAt: string;
    expiresAt: string;
};

type LoginDraft = {
    staffId: string;
    pin: string;
    workspace: WorkspaceKey;
};

type AuditLogEntry = {
    id: string;
    eventType: string;
    targetType: string;
    targetId: string;
    detail: string;
    actor: string;
    createdAt: string;
};

type StaffDraft = {
    name: string;
    role: string;
    pin: string;
    status: string;
};

type SyncStatus = {
    pendingCount: number;
    syncedCount: number;
    failedCount: number;
    oldestPendingAt: string;
    lastSyncedAt: string;
    lastError: string;
    databasePath: string;
    databaseBytes: number;
    walBytes: number;
    updatedAt: string;
};

type MenuCategory = {
    id: string;
    name: string;
    sortOrder: number;
    status: string;
};

type MenuModifier = {
    id: string;
    name: string;
    priceDelta: number;
    routeId: string;
    status: string;
};

type MenuItemModifier = {
    itemId: string;
    modifierId: string;
};

type FloorSection = {
    id: string;
    name: string;
    sortOrder: number;
};

type DiningTable = {
    id: string;
    sectionId: string;
    sectionName: string;
    name: string;
    seats: number;
    status: string;
    activeSessionId: string;
};

type OrderSession = {
    id: string;
    tableId: string;
    tableName: string;
    sectionName: string;
    waiterId: string;
    waiterName: string;
    guestCount: number;
    serviceMode: string;
    status: string;
    invoiceId: string;
    subtotal: number;
    taxTotal: number;
    serviceCharge: number;
    total: number;
    openedAt: string;
    closedAt: string;
    lineCount: number;
    readyLineCount: number;
    preparingLineCount: number;
    queuedLineCount: number;
    notSentLineCount: number;
};

type OrderSessionLine = {
    id: string;
    itemId: string;
    itemName: string;
    quantity: number;
    unitPrice: number;
    modifierTotal: number;
    lineTotal: number;
    notes: string;
    status: string;
    kotStatus: string;
    modifierIds: string[];
    modifierNames: string[];
};

type OrderSessionEvent = {
    id: string;
    type: string;
    detail: string;
    actor: string;
    createdAt: string;
};

type OrderSessionDetail = {
    session: OrderSession;
    lines: OrderSessionLine[];
    events: OrderSessionEvent[];
};

type PaymentRequest = {
    id: string;
    invoiceId: string;
    provider: string;
    amount: number;
    currency: string;
    status: string;
    reference: string;
    checkoutUrl: string;
    qrPayload: string;
    createdAt: string;
    updatedAt: string;
};

type PrintJob = {
    id: string;
    kind: string;
    referenceId: string;
    printerId: string;
    target: string;
    payload: string;
    status: string;
    attempts: number;
    lastError: string;
    createdAt: string;
    printedAt: string;
};

type Vendor = {
    id: string;
    name: string;
    phone: string;
    gstin: string;
    paymentTerms: string;
    qualityScore: number;
    status: string;
};

type PurchaseOrderLine = {
    id: string;
    purchaseOrderId: string;
    ingredientId: string;
    ingredientName: string;
    orderedQty: number;
    acceptedQty: number;
    rejectedQty: number;
    unitCost: number;
    rejectionReason: string;
};

type PurchaseOrder = {
    id: string;
    vendorId: string;
    vendorName: string;
    poNumber: string;
    status: string;
    expectedDate: string;
    subtotal: number;
    rejectedTotal: number;
    createdAt: string;
    lines: PurchaseOrderLine[];
};

type VendorDebitNote = {
    id: string;
    vendorName: string;
    purchaseOrderId: string;
    amount: number;
    reason: string;
    status: string;
    createdAt: string;
};

type DayCloseSummary = {
    businessDate: string;
    status: string;
    salesTotal: number;
    cashExpected: number;
    cashCounted: number;
    cashVariance: number;
    upiTotal: number;
    cardTotal: number;
    razorpayTotal: number;
    refundTotal: number;
    voidCount: number;
    discountTotal: number;
    notes: string;
};

type AccountingBalanceRow = {
    ledgerId: string;
    ledgerName: string;
    groupName: string;
    normalBalance: string;
    debitTotal: number;
    creditTotal: number;
    balance: number;
    balanceSide: string;
};

type AccountingSnapshot = {
    salesTotal: number;
    taxPayable: number;
    cashAndBank: number;
    razorpayClearing: number;
    vendorPayables: number;
    refundTotal: number;
    voucherCount: number;
    trialBalance: AccountingBalanceRow[];
};

type PilotWorkspace = {
    settings: RestaurantSettings;
    integrations: IntegrationSetting[];
    staff: StaffMember[];
    categories: MenuCategory[];
    modifiers: MenuModifier[];
    itemModifiers: MenuItemModifier[];
    floorSections: FloorSection[];
    tables: DiningTable[];
    orderSessions: OrderSession[];
    paymentRequests: PaymentRequest[];
    printJobs: PrintJob[];
    vendors: Vendor[];
    purchaseOrders: PurchaseOrder[];
    debitNotes: VendorDebitNote[];
    dayClose: DayCloseSummary;
    accounting: AccountingSnapshot;
    auditLog: AuditLogEntry[];
};

type AnalyticsMetric = {
    id: string;
    label: string;
    value: number;
    format: string;
    detail: string;
    tone: string;
};

type AnalyticsPoint = {
    label: string;
    value: number;
    count: number;
    tone: string;
};

type InventoryHealthRow = {
    id: string;
    name: string;
    onHandQty: number;
    reorderPoint: number;
    unit: string;
    riskScore: number;
    status: string;
};

type KitchenPerformanceRow = {
    routeName: string;
    openTickets: number;
    readyTickets: number;
    servedTickets: number;
    averageAgeMin: number;
    oldestAgeMin: number;
};

type ItemMatrixEntry = {
    itemId: string;
    name: string;
    category: string;
    quantity: number;
    sales: number;
    margin: number;
    marginPct: number;
};

type ItemMatrixBucket = {
    id: string;
    label: string;
    description: string;
    items: ItemMatrixEntry[];
};

type SettlementHealth = {
    cashExpected: number;
    cashVariance: number;
    upiTotal: number;
    cardTotal: number;
    razorpayTotal: number;
    razorpayClearing: number;
    taxPayable: number;
    vendorPayables: number;
    refundTotal: number;
};

type AnalyticsException = {
    id: string;
    kind: string;
    title: string;
    detail: string;
    severity: string;
    value: number;
};

type AnalyticsRecommendation = {
    id: string;
    title: string;
    detail: string;
    priority: number;
    page: PageKey;
};

type AnalyticsSnapshotStatus = {
    dailyRows: number;
    itemRows: number;
    hourlyRows: number;
    updatedAt: string;
};

type AdminAnalytics = {
    generatedAt: string;
    rangeKey: string;
    rangeLabel: string;
    demoFallback: boolean;
    executive: AnalyticsMetric[];
    salesTrend: AnalyticsPoint[];
    hourlyHeatmap: AnalyticsPoint[];
    tenderMix: AnalyticsPoint[];
    categoryMix: AnalyticsPoint[];
    itemVelocity: AnalyticsPoint[];
    contributionMargin: AnalyticsPoint[];
    inventoryHealth: InventoryHealthRow[];
    kitchenPerformance: KitchenPerformanceRow[];
    purchaseTrend: AnalyticsPoint[];
    itemMatrix: ItemMatrixBucket[];
    settlement: SettlementHealth;
    exceptions: AnalyticsException[];
    recommendations: AnalyticsRecommendation[];
    snapshotStatus: AnalyticsSnapshotStatus;
};

type CartLine = {
    itemId: string;
    quantity: number;
    notes: string;
};

type RecipeDraftLine = {
    ingredientId: string;
    quantity: number;
};

type MenuItemDraft = {
    name: string;
    category: string;
    price: number;
    cost: number;
    routeId: string;
    taxPercent: number;
};

type ModifierDraft = {
    name: string;
    priceDelta: number;
    routeId: string;
    status: string;
};

type ModifierLinkDraft = {
    itemId: string;
    modifierId: string;
};

type IngredientDraft = {
    reorderPoint: number;
    purchaseUnit: string;
    purchaseToUsage: number;
    lastPurchaseCost: number;
};

type IngredientCreateDraft = {
    name: string;
    unit: string;
    purchaseUnit: string;
    purchaseToUsage: number;
    onHandQty: number;
    reorderPoint: number;
    lastPurchaseCost: number;
};

type VendorDraft = {
    name: string;
    phone: string;
    gstin: string;
    paymentTerms: string;
};

type PurchaseOrderDraftLine = {
    ingredientId: string;
    orderedQty: number;
    unitCost: number;
};

type PurchaseOrderDraft = {
    vendorId: string;
    expectedDate: string;
    lines: PurchaseOrderDraftLine[];
};

type Toast = {
    id: string;
    kind: ToastKind;
    title: string;
    detail: string;
};

type ReadinessItem = {
    id: string;
    title: string;
    detail: string;
    done: boolean;
    page: PageKey;
};

const workspaces: Array<{key: WorkspaceKey; label: string; short: string; detail: string; defaultPage: PageKey}> = [
    {key: 'front', label: 'Front of House', short: 'FOH', detail: 'Tables, counter, bills', defaultPage: 'home'},
    {key: 'kitchenOps', label: 'Kitchen', short: 'KDS', detail: 'Station queue', defaultPage: 'kitchen'},
    {key: 'backoffice', label: 'Back Office', short: 'BO', detail: 'Menu, stock, devices', defaultPage: 'inventory'},
    {key: 'admin', label: 'Admin Command', short: 'ADM', detail: 'Business health', defaultPage: 'admin'},
];

const pages: Array<{key: PageKey; label: string; group: string; workspaces: WorkspaceKey[]}> = [
    {key: 'home', label: 'Today', group: 'Floor', workspaces: ['front']},
    {key: 'counter', label: 'Quick Bill', group: 'Floor', workspaces: ['front']},
    {key: 'floor', label: 'Tables', group: 'Floor', workspaces: ['front']},
    {key: 'orders', label: 'Open Bills', group: 'Floor', workspaces: ['front']},
    {key: 'invoices', label: 'Receipts', group: 'Money', workspaces: ['front', 'admin']},
    {key: 'kitchen', label: 'Kitchen Display', group: 'Production', workspaces: ['kitchenOps']},
    {key: 'admin', label: 'Command Center', group: 'Executive', workspaces: ['admin']},
    {key: 'dayclose', label: 'Close Day', group: 'Executive', workspaces: ['admin']},
    {key: 'accounting', label: 'Reports', group: 'Executive', workspaces: ['admin']},
    {key: 'inventory', label: 'Stock', group: 'Operations', workspaces: ['backoffice']},
    {key: 'vendors', label: 'Purchases', group: 'Operations', workspaces: ['backoffice']},
    {key: 'recipes', label: 'Menu Setup', group: 'Operations', workspaces: ['backoffice']},
    {key: 'customers', label: 'Customers', group: 'Growth', workspaces: ['backoffice']},
    {key: 'marketing', label: 'Marketing', group: 'Growth', workspaces: ['backoffice']},
    {key: 'integrations', label: 'Devices', group: 'Systems', workspaces: ['backoffice']},
    {key: 'settings', label: 'Settings', group: 'Systems', workspaces: ['backoffice', 'admin']},
];

const currency = new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
});

const compactNumber = new Intl.NumberFormat('en-IN', {
    maximumFractionDigits: 1,
});

function groupedPages(workspace: WorkspaceKey) {
    const groups: Array<[string, typeof pages]> = [];
    for (const page of pages.filter((item) => item.workspaces.includes(workspace))) {
        const last = groups[groups.length - 1];
        if (last && last[0] === page.group) {
            last[1].push(page);
        } else {
            groups.push([page.group, [page]]);
        }
    }
    return groups;
}

function App() {
    const [dashboard, setDashboard] = useState<Dashboard | null>(null);
    const [pilot, setPilot] = useState<PilotWorkspace | null>(null);
    const [analytics, setAnalytics] = useState<AdminAnalytics | null>(null);
    const [invoices, setInvoices] = useState<InvoiceSummary[]>([]);
    const [notifications, setNotifications] = useState<NotificationRecord[]>([]);
    const [selectedInvoice, setSelectedInvoice] = useState<InvoiceDetail | null>(null);
    const [selectedSession, setSelectedSession] = useState<OrderSessionDetail | null>(null);
    const [currentStaff, setCurrentStaff] = useState<StaffSession | null>(null);
    const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
    const [loginDraft, setLoginDraft] = useState<LoginDraft>({staffId: '', pin: '', workspace: 'front'});
    const [activeWorkspace, setActiveWorkspace] = useState<WorkspaceKey>('front');
    const [activePage, setActivePage] = useState<PageKey>('home');
    const [analyticsRange, setAnalyticsRange] = useState('30d');
    const [busy, setBusy] = useState('');
    const [error, setError] = useState('');
    const [toasts, setToasts] = useState<Toast[]>([]);

    const [cart, setCart] = useState<CartLine[]>([]);
    const [customerName, setCustomerName] = useState('Rahul Mehra');
    const [customerPhone, setCustomerPhone] = useState('+919999111222');
    const [orderType, setOrderType] = useState('dine_in');
    const [tableName, setTableName] = useState('T-04');
    const [discountType, setDiscountType] = useState('percent');
    const [discountValue, setDiscountValue] = useState(0);
    const [paymentMethod, setPaymentMethod] = useState('cash');
    const [paymentTendered, setPaymentTendered] = useState(0);
    const [menuSearch, setMenuSearch] = useState('');

    const [invoiceStatusFilter, setInvoiceStatusFilter] = useState('all');
    const [invoiceSearch, setInvoiceSearch] = useState('');
    const [invoicePaymentFilter, setInvoicePaymentFilter] = useState('all');
    const [managerPIN, setManagerPIN] = useState('');
    const [dayClosePIN, setDayClosePIN] = useState('');
    const [lineVoidPIN, setLineVoidPIN] = useState('');
    const [lineVoidReason, setLineVoidReason] = useState('');
    const [actionReason, setActionReason] = useState('');
    const [refundAmount, setRefundAmount] = useState(0);
    const [splitMode, setSplitMode] = useState<'items' | 'amount'>('items');
    const [splitLineQty, setSplitLineQty] = useState<Record<string, number>>({});
    const [splitAmountA, setSplitAmountA] = useState(0);
    const [splitAmountB, setSplitAmountB] = useState(0);
    const [dayCloseCash, setDayCloseCash] = useState(0);
    const [settingsDraft, setSettingsDraft] = useState<RestaurantSettings | null>(null);
    const [staffDraft, setStaffDraft] = useState<StaffDraft>({
        name: '',
        role: 'waiter',
        pin: '',
        status: 'active',
    });
    const [floorWaiterId, setFloorWaiterId] = useState('');
    const [floorGuestCount, setFloorGuestCount] = useState(2);
    const [menuImportText, setMenuImportText] = useState('Name,Category,Price,Route,Tax,Cost\nIced Americano,Coffee,150,Barista,5%,42\nPaneer Wrap,Food,240,Kitchen,5%,92');
    const [menuItemDraft, setMenuItemDraft] = useState<MenuItemDraft>({
        name: '',
        category: 'Coffee',
        price: 0,
        cost: 0,
        routeId: 'route_bar',
        taxPercent: 5,
    });
    const [categoryDraft, setCategoryDraft] = useState('');
    const [modifierDraft, setModifierDraft] = useState<ModifierDraft>({
        name: '',
        priceDelta: 0,
        routeId: 'route_kitchen',
        status: 'active',
    });
    const [modifierLinkDraft, setModifierLinkDraft] = useState<ModifierLinkDraft>({
        itemId: '',
        modifierId: '',
    });
    const [printerTarget, setPrinterTarget] = useState('tcp://192.168.1.50:9100');

    const [activeRecipeId, setActiveRecipeId] = useState('');
    const [recipeDraft, setRecipeDraft] = useState<RecipeDraftLine[]>([]);
    const [activeIngredientId, setActiveIngredientId] = useState('');
    const [ingredientDraft, setIngredientDraft] = useState<IngredientDraft>({
        reorderPoint: 0,
        purchaseUnit: '',
        purchaseToUsage: 1,
        lastPurchaseCost: 0,
    });
    const [ingredientCreateDraft, setIngredientCreateDraft] = useState<IngredientCreateDraft>({
        name: '',
        unit: 'g',
        purchaseUnit: 'kg',
        purchaseToUsage: 1000,
        onHandQty: 0,
        reorderPoint: 0,
        lastPurchaseCost: 0,
    });
    const [vendorDraft, setVendorDraft] = useState<VendorDraft>({
        name: '',
        phone: '',
        gstin: '',
        paymentTerms: 'Net 7',
    });
    const [poDraft, setPODraft] = useState<PurchaseOrderDraft>({
        vendorId: '',
        expectedDate: new Date().toISOString().slice(0, 10),
        lines: [{ingredientId: '', orderedQty: 0, unitCost: 0}],
    });

    useEffect(() => {
        void refreshAll();
    }, []);

    useEffect(() => {
        void refreshAnalytics();
    }, [analyticsRange]);

    useEffect(() => {
        void refreshInvoices();
    }, [invoiceStatusFilter, invoiceSearch, invoicePaymentFilter]);

    useEffect(() => {
        if (!dashboard) return;
        if (!activeRecipeId && dashboard.menuItems.length > 0) {
            setActiveRecipeId(dashboard.menuItems[0].id);
        }
        if (!activeIngredientId && dashboard.ingredients.length > 0) {
            setActiveIngredientId(dashboard.ingredients[0].id);
        }
        if (!menuItemDraft.routeId && dashboard.kitchenRoutes.length > 0) {
            setMenuItemDraft((current) => ({...current, routeId: dashboard.kitchenRoutes[0].id}));
        }
    }, [activeIngredientId, activeRecipeId, dashboard, menuItemDraft.routeId]);

    useEffect(() => {
        if (activePage !== 'invoices' || invoices.length === 0) return;
        const selectedID = selectedInvoice?.summary.id ?? '';
        if (selectedID && invoices.some((invoice) => invoice.id === selectedID)) return;
        void openInvoice(invoices[0].id, false);
    }, [activePage, invoices, selectedInvoice?.summary.id]);

    useEffect(() => {
        if (!dashboard || !activeRecipeId) return;
        const card = dashboard.recipes.find((recipe) => recipe.itemId === activeRecipeId);
        setRecipeDraft((card?.components ?? []).map((component) => ({
            ingredientId: component.ingredientId,
            quantity: component.quantity,
        })));
    }, [activeRecipeId, dashboard]);

    useEffect(() => {
        if (!dashboard || !activeIngredientId) return;
        const ingredient = dashboard.ingredients.find((item) => item.id === activeIngredientId);
        if (!ingredient) return;
        setIngredientDraft({
            reorderPoint: ingredient.reorderPoint,
            purchaseUnit: ingredient.purchaseUnit,
            purchaseToUsage: ingredient.purchaseToUsage,
            lastPurchaseCost: ingredient.lastPurchaseCost,
        });
    }, [activeIngredientId, dashboard]);

    useEffect(() => {
        if (!dashboard || !pilot) return;
        setPODraft((current) => {
            const vendorId = current.vendorId || pilot.vendors[0]?.id || '';
            const fallbackIngredient = dashboard.ingredients[0]?.id || '';
            const nextLines = current.lines.length > 0
                ? current.lines.map((line) => line.ingredientId ? line : {...line, ingredientId: fallbackIngredient})
                : [{ingredientId: fallbackIngredient, orderedQty: 0, unitCost: 0}];
            const linesChanged = nextLines.length !== current.lines.length || nextLines.some((line, index) => line !== current.lines[index]);
            if (vendorId === current.vendorId && !linesChanged) {
                return current;
            }
            return {...current, vendorId, lines: nextLines};
        });
    }, [dashboard, pilot]);

    useEffect(() => {
        if (!dashboard || !pilot) return;
        setModifierDraft((current) => {
            if (current.routeId || dashboard.kitchenRoutes.length === 0) return current;
            return {...current, routeId: dashboard.kitchenRoutes[0].id};
        });
        setModifierLinkDraft((current) => {
            const itemId = current.itemId || dashboard.menuItems[0]?.id || '';
            const modifierId = current.modifierId || pilot.modifiers[0]?.id || '';
            if (itemId === current.itemId && modifierId === current.modifierId) return current;
            return {...current, itemId, modifierId};
        });
    }, [dashboard, pilot]);

    useEffect(() => {
        if (!pilot) return;
        const activeStaff = pilot.staff.filter((member) => member.status === 'active');
        if (floorWaiterId && activeStaff.some((member) => member.id === floorWaiterId)) return;
        const preferred = activeStaff.find((member) => member.role === 'waiter') ?? activeStaff[0];
        setFloorWaiterId(preferred?.id ?? '');
    }, [floorWaiterId, pilot]);

    useEffect(() => {
        if (!pilot) return;
        setLoginDraft((current) => {
            const activeStaff = pilot.staff.filter((member) => member.status === 'active');
            const selected = activeStaff.find((member) => member.id === current.staffId) ?? activeStaff.find((member) => member.role === 'manager') ?? activeStaff[0];
            if (!selected) return current;
            const access = workspaceAccessForRole(selected.role);
            const workspace = access.includes(current.workspace) ? current.workspace : access[0] ?? 'front';
            if (current.staffId === selected.id && current.workspace === workspace) return current;
            return {...current, staffId: selected.id, workspace};
        });
    }, [pilot]);

    useEffect(() => {
        if (!currentStaff) return;
        if (currentStaff.workspaceAccess.includes(activeWorkspace)) return;
        const nextWorkspace = currentStaff.workspaceAccess[0] ?? 'front';
        switchWorkspace(nextWorkspace);
    }, [activeWorkspace, currentStaff]);

    useEffect(() => {
        if (!pilot || orderType !== 'dine_in') return;
        if (tableName && pilot.tables.some((table) => table.name === tableName)) return;
        const preferred = pilot.tables.find((table) => table.status === 'available') ?? pilot.tables[0];
        if (preferred) setTableName(preferred.name);
    }, [orderType, pilot, tableName]);

    useEffect(() => {
        if (!selectedInvoice) return;
        const total = selectedInvoice.summary.total;
        const first = Math.round((total / 2) * 100) / 100;
        setRefundAmount(total);
        setSplitAmountA(first);
        setSplitAmountB(Math.round((total - first) * 100) / 100);
        setSplitLineQty(Object.fromEntries(selectedInvoice.lines.map((line) => [line.id, 0])));
    }, [selectedInvoice?.summary.id]);

    const itemById = useMemo(() => new Map((dashboard?.menuItems ?? []).map((item) => [item.id, item])), [dashboard]);
    const modifiersByItem = useMemo(() => {
        const map = new Map<string, MenuModifier[]>();
        const modifiers = new Map((pilot?.modifiers ?? []).map((modifier) => [modifier.id, modifier]));
        for (const link of pilot?.itemModifiers ?? []) {
            const modifier = modifiers.get(link.modifierId);
            if (!modifier) continue;
            map.set(link.itemId, [...(map.get(link.itemId) ?? []), modifier]);
        }
        return map;
    }, [pilot]);
    const openOrders = useMemo(() => invoices.filter((invoice) => ['draft', 'open', 'kot_sent'].includes(invoice.status)), [invoices]);
    const closedInvoices = useMemo(() => invoices.filter((invoice) => ['closed', 'partially_refunded', 'refunded', 'voided', 'split'].includes(invoice.status)), [invoices]);
    const topSignals = useMemo(() => [...(dashboard?.signals ?? [])].sort((a, b) => a.priority - b.priority).slice(0, 4), [dashboard]);
    const currentPage = useMemo(() => pages.find((page) => page.key === activePage) ?? pages[0], [activePage]);
    const currentWorkspace = useMemo(() => workspaces.find((workspace) => workspace.key === activeWorkspace) ?? workspaces[0], [activeWorkspace]);
    const availableWorkspaces = useMemo(() => {
        const access = currentStaff?.workspaceAccess ?? [];
        return access.length > 0 ? workspaces.filter((workspace) => access.includes(workspace.key)) : workspaces;
    }, [currentStaff]);

    const cartTotals = useMemo(() => {
        const subtotal = cart.reduce((sum, line) => {
            const item = itemById.get(line.itemId);
            return sum + (item ? item.price * line.quantity : 0);
        }, 0);
        const discount = calculateDiscount(subtotal, discountType, discountValue);
        const tax = cart.reduce((sum, line) => {
            const item = itemById.get(line.itemId);
            if (!item || subtotal <= 0) return sum;
            const lineSubtotal = item.price * line.quantity;
            const share = lineSubtotal / subtotal;
            return sum + ((lineSubtotal - discount * share) * item.taxRate);
        }, 0);
        return {
            subtotal,
            discount,
            tax,
            total: Math.round((subtotal - discount + tax) * 100) / 100,
        };
    }, [cart, discountType, discountValue, itemById]);

    async function refreshAll() {
        setError('');
        try {
            const [nextDashboard, nextNotifications, nextPilot, nextAnalytics, nextSyncStatus] = await Promise.all([
                GetDashboard(),
                GetNotifications(),
                GetPilotWorkspace(),
                GetAdminAnalytics(analyticsRange),
                GetSyncStatus(),
            ]);
            setDashboard(normalizeDashboard(nextDashboard as Dashboard));
            setNotifications(nextNotifications as NotificationRecord[] ?? []);
            setAnalytics(normalizeAdminAnalytics(nextAnalytics as AdminAnalytics));
            setSyncStatus(normalizeSyncStatus(nextSyncStatus as SyncStatus));
            const normalizedPilot = normalizePilotWorkspace(nextPilot as PilotWorkspace);
            setPilot(normalizedPilot);
            setSettingsDraft(normalizedPilot.settings);
            const printer = normalizedPilot.integrations.find((integration) => integration.provider === 'escpos_printer');
            if (printer?.baseUrl) setPrinterTarget(printer.baseUrl);
            if (selectedSession?.session.id && normalizedPilot.orderSessions.some((session) => session.id === selectedSession.session.id)) {
                await openSession(selectedSession.session.id, false);
            } else if (selectedSession?.session.id) {
                setSelectedSession(null);
            }
            await refreshInvoices();
            if (selectedInvoice?.summary.id) {
                await openInvoice(selectedInvoice.summary.id, false);
            }
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            setError(message);
        }
    }

    async function refreshAnalytics() {
        try {
            const next = await GetAdminAnalytics(analyticsRange);
            setAnalytics(normalizeAdminAnalytics(next as AdminAnalytics));
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            setError(message);
        }
    }

    async function refreshInvoices() {
        try {
            const next = await GetInvoices(new nexus.InvoiceFilter({
                status: invoiceStatusFilter,
                search: invoiceSearch,
                paymentMethod: invoicePaymentFilter,
            }));
            setInvoices((next as InvoiceSummary[]) ?? []);
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            setError(message);
        }
    }

    async function runAction(name: string, successTitle: string, action: () => Promise<unknown>) {
        setBusy(name);
        setError('');
        try {
            await action();
            await refreshAll();
            pushToast('success', successTitle, 'Local ledger updated');
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            setError(message);
            pushToast('error', 'Action failed', message);
        } finally {
            setBusy('');
        }
    }

    function pushToast(kind: ToastKind, title: string, detail: string) {
        const id = `${Date.now()}-${Math.random()}`;
        setToasts((current) => [...current, {id, kind, title, detail}].slice(-4));
        window.setTimeout(() => {
            setToasts((current) => current.filter((toast) => toast.id !== id));
        }, 4200);
    }

    async function signIn() {
        setBusy('login');
        setError('');
        try {
            const session = await AuthenticateStaff(new nexus.StaffLoginInput({
                staffId: loginDraft.staffId,
                pin: loginDraft.pin,
                workspace: loginDraft.workspace,
            }));
            const normalized = normalizeStaffSession(session as StaffSession);
            setCurrentStaff(normalized);
            const workspace = normalized.workspaceAccess.includes(loginDraft.workspace) ? loginDraft.workspace : normalized.workspaceAccess[0] ?? 'front';
            switchWorkspace(workspace);
            setLoginDraft((current) => ({...current, pin: '', workspace}));
            pushToast('success', 'Signed in', `${normalized.name} / ${humanStatus(normalized.role)}`);
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            setError(message);
            pushToast('error', 'Sign in failed', message);
        } finally {
            setBusy('');
        }
    }

    function signOut() {
        setCurrentStaff(null);
        setManagerPIN('');
        setDayClosePIN('');
        setLineVoidPIN('');
        setLoginDraft((current) => ({...current, pin: '', workspace: 'front'}));
    }

    function switchWorkspace(workspaceKey: WorkspaceKey) {
        if (currentStaff && !currentStaff.workspaceAccess.includes(workspaceKey)) {
            pushToast('error', 'Workspace locked', `${currentStaff.name} cannot access ${workspaceLabel(workspaceKey)}`);
            return;
        }
        const workspace = workspaces.find((item) => item.key === workspaceKey) ?? workspaces[0];
        setActiveWorkspace(workspace.key);
        setActivePage(workspace.defaultPage);
    }

    function navigateToPage(pageKey: PageKey) {
        const page = pages.find((item) => item.key === pageKey);
        const workspace = page?.workspaces.includes(activeWorkspace) ? activeWorkspace : page?.workspaces[0];
        if (workspace) {
            if (currentStaff && !currentStaff.workspaceAccess.includes(workspace)) {
                pushToast('error', 'Workspace locked', `${currentStaff.name} cannot access ${workspaceLabel(workspace)}`);
                return;
            }
            setActiveWorkspace(workspace);
        }
        setActivePage(pageKey);
    }

    function orderInput() {
        return new nexus.SaleInput({
            customerName,
            customerPhone,
            channel: 'counter',
            orderType,
            tableName: orderType === 'dine_in' ? tableName : '',
            discountType,
            discountValue,
            paymentMethod,
            paymentTendered,
            lines: cart.map((line) => ({
                itemId: line.itemId,
                quantity: line.quantity,
                notes: line.notes,
            })),
        });
    }

    function addToCart(item: MenuItem) {
        setCart((current) => {
            const existing = current.find((line) => line.itemId === item.id);
            if (existing) {
                return current.map((line) => line.itemId === item.id ? {...line, quantity: line.quantity + 1} : line);
            }
            return [...current, {itemId: item.id, quantity: 1, notes: ''}];
        });
    }

    function updateCartLine(itemId: string, patch: Partial<CartLine>) {
        setCart((current) => current
            .map((line) => line.itemId === itemId ? {...line, ...patch} : line)
            .filter((line) => line.quantity > 0));
    }

    function clearCart() {
        setCart([]);
        setPaymentTendered(0);
    }

    function saveDraft() {
        if (cart.length === 0) {
            pushToast('error', 'Cart is empty', 'Add at least one item');
            return;
        }
        return runAction('save-draft', 'Draft saved', async () => {
            await SaveOrder(orderInput());
            clearCart();
        });
    }

    function quickCloseBill() {
        if (cart.length === 0) {
            pushToast('error', 'Cart is empty', 'Add at least one item');
            return;
        }
        return runAction('quick-close', 'Bill closed', async () => {
            await RecordSale(orderInput());
            clearCart();
        });
    }

    function sendKOT(invoiceID: string) {
        return runAction(`kot-${invoiceID}`, 'KOT sent', () => SendKOT(invoiceID));
    }

    function closeInvoice(invoiceID: string) {
        return runAction(`close-${invoiceID}`, 'Invoice closed', () => CloseInvoice(new nexus.CloseInvoiceInput({
            invoiceId: invoiceID,
            paymentMethod,
            paymentTendered,
        })));
    }

    async function openSession(sessionID: string, navigate = true) {
        if (!sessionID) return;
        try {
            const detail = await GetOrderSessionDetail(sessionID);
            setSelectedSession(normalizeOrderSessionDetail(detail as OrderSessionDetail));
            if (navigate) setActivePage('floor');
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            if (navigate) {
                setError(message);
                pushToast('error', 'Could not open table', message);
            }
            setSelectedSession(null);
        }
    }

    function openTable(tableID: string, staffID = floorWaiterId) {
        if (!staffID) {
            pushToast('error', 'Choose staff first', 'Select the person handling this table');
            return;
        }
        return runAction(`open-table-${tableID}`, 'Table opened', async () => {
            const detail = await OpenOrderSession(new nexus.OpenOrderSessionInput({
                tableId: tableID,
                waiterId: staffID,
                guestCount: Math.max(1, Math.round(floorGuestCount) || 1),
                serviceMode: 'dine_in',
            }));
            setSelectedSession(normalizeOrderSessionDetail(detail as OrderSessionDetail));
            setActivePage('floor');
        });
    }

    function assignSessionStaff(sessionID: string, staffID: string) {
        if (!staffID) {
            pushToast('error', 'Choose staff first', 'Select the person handling this table');
            return;
        }
        return runAction(`assign-table-${sessionID}-${staffID}`, 'Staff assigned', async () => {
            const detail = await AssignOrderSessionStaff(new nexus.AssignOrderSessionStaffInput({
                sessionId: sessionID,
                staffId: staffID,
            }));
            setSelectedSession(normalizeOrderSessionDetail(detail as OrderSessionDetail));
        });
    }

    function addSessionItem(itemID: string, modifierIDs: string[] = []) {
        if (!selectedSession) {
            pushToast('error', 'Open a table first', 'Select an occupied table or start a new session');
            return;
        }
        return runAction(`session-line-${itemID}`, 'Item added', async () => {
            const detail = await AddOrderSessionLine(new nexus.AddOrderSessionLineInput({
                sessionId: selectedSession.session.id,
                itemId: itemID,
                quantity: 1,
                notes: '',
                modifierIds: modifierIDs,
            }));
            setSelectedSession(normalizeOrderSessionDetail(detail as OrderSessionDetail));
        });
    }

    function voidSessionLine(lineID: string) {
        if (!selectedSession) return;
        return runAction(`session-line-void-${lineID}`, 'Line voided', async () => {
            const detail = await VoidOrderSessionLine(new nexus.SessionLineVoidInput({
                sessionId: selectedSession.session.id,
                lineId: lineID,
                staffId: currentStaff && ['owner', 'manager'].includes(currentStaff.role) ? currentStaff.staffId : '',
                pin: lineVoidPIN || managerPIN,
                reason: lineVoidReason || 'Guest changed order',
            }));
            setSelectedSession(normalizeOrderSessionDetail(detail as OrderSessionDetail));
            setLineVoidReason('');
        });
    }

    function transferSession(targetTableID: string) {
        if (!selectedSession) return;
        return runAction(`transfer-${targetTableID}`, 'Table transferred', async () => {
            const detail = await TransferOrderSession(new nexus.TableMoveInput({
                sessionId: selectedSession.session.id,
                targetTableId: targetTableID,
            }));
            setSelectedSession(normalizeOrderSessionDetail(detail as OrderSessionDetail));
        });
    }

    function sendSessionKOT() {
        if (!selectedSession) return;
        return runAction('session-kot', 'KOT sent', async () => {
            const detail = await SendOrderSessionKOT(selectedSession.session.id);
            setSelectedSession(normalizeOrderSessionDetail(detail as OrderSessionDetail));
        });
    }

    function closeSession() {
        if (!selectedSession) return;
        return runAction('session-close', 'Table closed', async () => {
            await CloseOrderSession(new nexus.CloseOrderSessionInput({
                sessionId: selectedSession.session.id,
                customerName,
                customerPhone,
                paymentMethod,
                paymentTendered,
            }));
            setSelectedSession(null);
        });
    }

    function updateTicket(ticketID: string, status: string) {
        return runAction(`ticket-${ticketID}-${status}`, 'KOT updated', () => UpdateKOTStatus(ticketID, status));
    }

    function createPayment(invoiceID: string) {
        return runAction(`payment-${invoiceID}`, 'Razorpay request ready', () => CreatePaymentRequest(invoiceID));
    }

    function markPaymentPaid(requestID: string) {
        return runAction(`payment-paid-${requestID}`, 'Payment reconciled', () => MarkPaymentRequestPaid(requestID));
    }

    function markPaymentFailed(requestID: string) {
        return runAction(`payment-failed-${requestID}`, 'Payment marked failed', () => MarkPaymentRequestFailed(requestID, 'Operator could not collect this Razorpay request'));
    }

    function cancelPayment(requestID: string) {
        return runAction(`payment-cancel-${requestID}`, 'Payment request cancelled', () => CancelPaymentRequest(requestID));
    }

    function queuePrint(kind: string, referenceID: string) {
        return runAction(`print-${kind}-${referenceID}`, 'Print job queued', () => QueuePrintJob(kind, referenceID));
    }

    function markPrintPrinted(jobID: string) {
        return runAction(`print-done-${jobID}`, 'Print completed', () => MarkPrintJobPrinted(jobID));
    }

    function markPrintFailed(jobID: string) {
        return runAction(`print-failed-${jobID}`, 'Print marked failed', () => MarkPrintJobFailed(jobID, 'Operator did not receive a printer acknowledgement'));
    }

    function retryPrint(jobID: string) {
        return runAction(`print-retry-${jobID}`, 'Print returned to queue', () => RetryPrintJob(jobID));
    }

    async function runFileAction(name: string, successTitle: string, action: () => Promise<unknown>) {
        setBusy(name);
        setError('');
        try {
            const result = await action() as ExportResult;
            await refreshAll();
            pushToast('success', successTitle, result.fileName || result.path);
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            setError(message);
            pushToast('error', 'Export failed', message);
        } finally {
            setBusy('');
        }
    }

    function exportSelectedInvoicePDF() {
        if (!selectedInvoice) return;
        return runFileAction('invoice-pdf', 'Invoice PDF saved', () => ExportInvoicePDF(selectedInvoice.summary.id));
    }

    function exportInvoicesCSV() {
        return runFileAction('invoices-csv', 'Invoice report saved', () => ExportInvoicesCSV(new nexus.InvoiceFilter({
            status: invoiceStatusFilter,
            search: invoiceSearch,
            paymentMethod: invoicePaymentFilter,
        })));
    }

    function exportAccountingCSV() {
        return runFileAction('accounting-csv', 'Accounting report saved', () => ExportAccountingCSV());
    }

    function exportDayClosePDF() {
        return runFileAction('dayclose-pdf', 'Day close PDF saved', () => ExportDayClosePDF(pilot?.dayClose.businessDate || new Date().toISOString().slice(0, 10)));
    }

    function exportBackup() {
        return runFileAction('backup', 'Backup saved', () => ExportBackup(new nexus.BackupInput({
            destination: settingsDraft?.backupPath || '',
        })));
    }

    function savePrinter() {
        return runAction('save-printer', 'Printer saved', () => SavePrinterConnection(new nexus.PrinterConnectionInput({
            target: printerTarget,
            displayName: 'Receipt Printer',
            mode: printerTarget.startsWith('system://') ? 'system' : 'local',
        })));
    }

    function importMenuRows() {
        return runAction('menu-import', 'Menu imported', async () => {
            const result = await ImportMenuCSV(new nexus.MenuImportInput({text: menuImportText})) as MenuImportResult;
            pushToast('info', 'Menu rows', `${result.imported} new / ${result.updated} updated`);
        });
    }

    function saveMenuItemPrice(item: MenuItem, price: number) {
        return saveMenuItemPatch(item, {price});
    }

    function saveMenuItemPatch(item: MenuItem, patch: Partial<MenuItem>) {
        const next = {...item, ...patch};
        return runAction(`menu-item-${item.id}`, 'Menu item saved', () => SaveMenuItem(new nexus.MenuItemInput({
            id: item.id,
            name: next.name,
            category: next.category,
            price: Number(next.price) || 0,
            cost: Number(next.cost) || 0,
            status: next.status || 'active',
            routeId: next.routeId,
            taxRate: Number(next.taxRate) || 0,
        })));
    }

    function saveMenuItemDraft() {
        const name = menuItemDraft.name.trim();
        if (!name) {
            pushToast('error', 'Item name missing', 'Add a name before saving');
            return;
        }
        const routeID = menuItemDraft.routeId || dashboard?.kitchenRoutes[0]?.id || 'route_kitchen';
        return runAction('menu-item-new', 'Menu item added', async () => {
            await SaveMenuItem(new nexus.MenuItemInput({
                id: '',
                name,
                category: menuItemDraft.category.trim() || 'Food',
                price: Number(menuItemDraft.price) || 0,
                cost: Number(menuItemDraft.cost) || 0,
                status: 'active',
                routeId: routeID,
                taxRate: (Number(menuItemDraft.taxPercent) || 0) / 100,
            }));
            setMenuItemDraft({
                name: '',
                category: menuItemDraft.category.trim() || 'Food',
                price: 0,
                cost: 0,
                routeId: routeID,
                taxPercent: menuItemDraft.taxPercent || 5,
            });
        });
    }

    function saveCategoryDraft() {
        const name = categoryDraft.trim();
        if (!name) {
            pushToast('error', 'Category missing', 'Add a group name before saving');
            return;
        }
        return runAction('menu-category', 'Menu group saved', async () => {
            await SaveMenuCategory(new nexus.MenuCategoryInput({
                id: '',
                name,
                sortOrder: (pilot?.categories.length ?? 0) + 1,
                status: 'active',
            }));
            setCategoryDraft('');
            setMenuItemDraft((current) => ({...current, category: name}));
        });
    }

    function saveModifierDraft() {
        const name = modifierDraft.name.trim();
        if (!name) {
            pushToast('error', 'Modifier missing', 'Add an add-on name before saving');
            return;
        }
        return runAction('menu-modifier', 'Modifier saved', async () => {
            await SaveMenuModifier(new nexus.MenuModifierInput({
                id: '',
                name,
                priceDelta: Number(modifierDraft.priceDelta) || 0,
                routeId: modifierDraft.routeId || dashboard?.kitchenRoutes[0]?.id || 'route_kitchen',
                status: modifierDraft.status || 'active',
            }));
            setModifierDraft((current) => ({...current, name: '', priceDelta: 0}));
        });
    }

    function linkModifierDraft() {
        if (!modifierLinkDraft.itemId || !modifierLinkDraft.modifierId) {
            pushToast('error', 'Selection missing', 'Choose both an item and a modifier');
            return;
        }
        return runAction('modifier-link', 'Modifier attached', () => LinkModifierToItem(modifierLinkDraft.itemId, modifierLinkDraft.modifierId));
    }

    async function openInvoice(invoiceID: string, navigate = true) {
        setBusy(`detail-${invoiceID}`);
        setError('');
        try {
            const detail = await GetInvoiceDetail(invoiceID);
            setSelectedInvoice(normalizeInvoiceDetail(detail as InvoiceDetail));
            if (navigate) setActivePage('invoices');
        } catch (err) {
            const message = err instanceof Error ? err.message : String(err);
            setError(message);
            pushToast('error', 'Could not open invoice', message);
        } finally {
            setBusy('');
        }
    }

    function voidSelectedInvoice() {
        if (!selectedInvoice) return;
        return runAction('void-invoice', 'Invoice voided', () => VoidInvoice(new nexus.VoidInvoiceInput({
            invoiceId: selectedInvoice.summary.id,
            pin: managerPIN,
            reason: actionReason || 'Operator void',
        })));
    }

    function refundSelectedInvoice() {
        if (!selectedInvoice) return;
        return runAction('refund-invoice', 'Refund recorded', () => RefundInvoice(new nexus.RefundInvoiceInput({
            invoiceId: selectedInvoice.summary.id,
            pin: managerPIN,
            amount: refundAmount,
            reason: actionReason || 'Customer refund',
        })));
    }

    function validatePIN() {
        return runAction('validate-pin', 'PIN approved', () => ValidateManagerPIN(managerPIN));
    }

    function splitSelectedInvoice() {
        if (!selectedInvoice) return;
        if (splitMode === 'items') {
            const lines = Object.entries(splitLineQty)
                .filter(([, quantity]) => quantity > 0)
                .map(([lineId, quantity]) => ({lineId, quantity}));
            return runAction('split-items', 'Item split created', () => SplitInvoice(new nexus.SplitInvoiceInput({
                invoiceId: selectedInvoice.summary.id,
                mode: 'items',
                lines,
            })));
        }
        return runAction('split-amounts', 'Amount split created', () => SplitInvoice(new nexus.SplitInvoiceInput({
            invoiceId: selectedInvoice.summary.id,
            mode: 'amount',
            amounts: [splitAmountA, splitAmountB],
        })));
    }

    function receiveBadBatch() {
        const tomato = dashboard?.ingredients.find((ingredient) => ingredient.id === 'ing_tomato');
        if (!tomato) return;
        return runAction('bad-batch', 'Vendor exception logged', () => ReceiveDelivery(new nexus.DeliveryInput({
            vendorName: 'Fresh Farm Co',
            invoiceNumber: `FF-${new Date().getHours()}${new Date().getMinutes()}`,
            lines: [{
                ingredientId: tomato.id,
                orderedQty: 5000,
                acceptedQty: 4200,
                rejectedQty: 800,
                unitCost: 0.08,
                rejectionReason: 'Soft batch found at receiving',
            }],
        })));
    }

    function addPOLine() {
        setPODraft((current) => ({
            ...current,
            lines: [...current.lines, {ingredientId: dashboard?.ingredients[0]?.id || '', orderedQty: 0, unitCost: 0}],
        }));
    }

    function updatePOLine(index: number, patch: Partial<PurchaseOrderDraftLine>) {
        setPODraft((current) => ({
            ...current,
            lines: current.lines.map((line, lineIndex) => lineIndex === index ? {...line, ...patch} : line),
        }));
    }

    function removePOLine(index: number) {
        setPODraft((current) => ({
            ...current,
            lines: current.lines.length > 1 ? current.lines.filter((_, lineIndex) => lineIndex !== index) : current.lines,
        }));
    }

    function saveVendorDraft() {
        const name = vendorDraft.name.trim();
        if (!name) {
            pushToast('error', 'Vendor name missing', 'Add a supplier name');
            return;
        }
        return runAction('save-vendor', 'Vendor saved', async () => {
            await SaveVendor(new nexus.VendorInput({
                id: '',
                name,
                phone: vendorDraft.phone,
                gstin: vendorDraft.gstin,
                paymentTerms: vendorDraft.paymentTerms || 'Net 7',
                status: 'active',
            }));
            setVendorDraft({name: '', phone: '', gstin: '', paymentTerms: 'Net 7'});
        });
    }

    function createPurchaseOrderDraft() {
        const lines = poDraft.lines
            .filter((line) => line.ingredientId && line.orderedQty > 0)
            .map((line) => ({
                ingredientId: line.ingredientId,
                orderedQty: line.orderedQty,
                unitCost: Math.max(0, line.unitCost),
            }));
        if (!poDraft.vendorId || lines.length === 0) {
            pushToast('error', 'PO is incomplete', 'Choose a vendor and add at least one quantity');
            return;
        }
        return runAction('create-po', 'Purchase order drafted', async () => {
            await CreatePurchaseOrder(new nexus.PurchaseOrderInput({
                vendorId: poDraft.vendorId,
                expectedDate: poDraft.expectedDate,
                lines,
            }));
            setPODraft({
                vendorId: poDraft.vendorId,
                expectedDate: new Date().toISOString().slice(0, 10),
                lines: [{ingredientId: dashboard?.ingredients[0]?.id || '', orderedQty: 0, unitCost: 0}],
            });
        });
    }

    function receiveFirstPurchaseOrder() {
        const po = pilot?.purchaseOrders.find((order) => order.status !== 'received' && order.lines.length > 0);
        if (!po) {
            pushToast('error', 'No PO to receive', 'Create a purchase order first');
            return;
        }
        return runAction('receive-po', 'Delivery received', () => ReceivePurchaseOrder(new nexus.ReceivePurchaseOrderInput({
            purchaseOrderId: po.id,
            lines: po.lines.map((line) => ({
                lineId: line.id,
                acceptedQty: Math.round(line.orderedQty * 0.84),
                rejectedQty: Math.round(line.orderedQty * 0.16),
                rejectionReason: 'Quality check rejection',
            })),
        })));
    }

    function closeDayNow() {
        return runAction('close-day', 'Day closed', () => CloseDay(new nexus.DayCloseInput({
            businessDate: pilot?.dayClose.businessDate || new Date().toISOString().slice(0, 10),
            cashCounted: dayCloseCash,
            notes: 'Operator close run',
            staffId: currentStaff?.staffId || '',
            pin: dayClosePIN || managerPIN,
        })));
    }

    function rebuildAccounting() {
        return runAction('rebuild-accounting', 'Accounting shell rebuilt', () => RebuildAccountingShell());
    }

    function saveSettings() {
        if (!settingsDraft) return;
        return runAction('save-settings', 'Setup saved', () => SaveRestaurantSettings(new nexus.RestaurantSettings(settingsDraft)));
    }

    function saveStaffDraft() {
        const name = staffDraft.name.trim();
        if (!name) {
            pushToast('error', 'Staff name missing', 'Add the person before saving');
            return;
        }
        if (!staffDraft.pin.trim()) {
            pushToast('error', 'Staff PIN missing', 'Add a login PIN for this staff member');
            return;
        }
        return runAction('save-staff', 'Staff saved', async () => {
            await SaveStaff(new nexus.StaffInput({
                id: '',
                name,
                role: staffDraft.role || 'waiter',
                pin: staffDraft.pin,
                status: staffDraft.status || 'active',
            }));
            setStaffDraft({name: '', role: staffDraft.role || 'waiter', pin: '', status: 'active'});
        });
    }

    function updateStaffStatus(member: StaffMember, status: string) {
        return runAction(`staff-${member.id}`, status === 'active' ? 'Staff restored' : 'Staff archived', () => SaveStaff(new nexus.StaffInput({
            id: member.id,
            name: member.name,
            role: member.role,
            pin: '',
            status,
        })));
    }

    function storeIntegration(provider: string) {
        const label = provider === 'razorpay' ? 'Razorpay' : provider === 'meta_whatsapp' ? 'Meta Cloud API' : 'ESC/POS Printer';
        const baseUrl = provider === 'razorpay' ? 'https://api.razorpay.com/v1' : provider === 'meta_whatsapp' ? 'https://graph.facebook.com' : 'tcp://192.168.1.50:9100';
        return runAction(`integration-${provider}`, `${label} saved`, () => SaveIntegrationSetting(new nexus.IntegrationSettingInput({
            provider,
            mode: provider === 'escpos_printer' ? 'local' : 'test',
            displayName: label,
            baseUrl,
            secret: provider === 'escpos_printer' ? '' : 'test_secret',
            healthStatus: 'ready_for_test',
        })));
    }

    function reconcileMilk() {
        const milk = dashboard?.ingredients.find((ingredient) => ingredient.id === 'ing_milk');
        if (!milk) return;
        return runAction('audit-milk', 'Stock audit saved', () => ReconcileStock(new nexus.ReconcileInput({
            ingredientId: milk.id,
            physicalQty: Math.max(0, milk.onHandQty - 450),
            note: 'End-of-shift spot check',
        })));
    }

    function addRecipeLine() {
        const ingredient = dashboard?.ingredients.find((item) => !recipeDraft.some((line) => line.ingredientId === item.id));
        if (!ingredient) return;
        setRecipeDraft((current) => [...current, {ingredientId: ingredient.id, quantity: 1}]);
    }

    function updateRecipeLine(index: number, patch: Partial<RecipeDraftLine>) {
        setRecipeDraft((current) => current.map((line, lineIndex) => lineIndex === index ? {...line, ...patch} : line));
    }

    function removeRecipeLine(index: number) {
        setRecipeDraft((current) => current.filter((_, lineIndex) => lineIndex !== index));
    }

    function saveRecipe() {
        if (!activeRecipeId) return;
        return runAction('save-recipe', 'Recipe saved', () => UpdateRecipe(new nexus.RecipeUpdateInput({
            itemId: activeRecipeId,
            components: recipeDraft
                .filter((line) => line.ingredientId && line.quantity > 0)
                .map((line) => ({
                    ingredientId: line.ingredientId,
                    quantity: line.quantity,
                })),
        })));
    }

    function saveIngredientSettings() {
        if (!activeIngredientId) return;
        return runAction('save-ingredient', 'Ingredient saved', () => UpdateIngredientSettings(new nexus.IngredientUpdateInput({
            ingredientId: activeIngredientId,
            reorderPoint: ingredientDraft.reorderPoint,
            purchaseUnit: ingredientDraft.purchaseUnit,
            purchaseToUsage: ingredientDraft.purchaseToUsage,
            lastPurchaseCost: ingredientDraft.lastPurchaseCost,
        })));
    }

    function saveNewIngredient() {
        const name = ingredientCreateDraft.name.trim();
        if (!name) {
            pushToast('error', 'Ingredient name missing', 'Add the stock item name');
            return;
        }
        return runAction('add-ingredient', 'Ingredient added', async () => {
            await SaveIngredient(new nexus.IngredientInput({
                id: '',
                name,
                unit: ingredientCreateDraft.unit || 'g',
                purchaseUnit: ingredientCreateDraft.purchaseUnit || ingredientCreateDraft.unit || 'g',
                purchaseToUsage: ingredientCreateDraft.purchaseToUsage || 1,
                onHandQty: ingredientCreateDraft.onHandQty || 0,
                reorderPoint: ingredientCreateDraft.reorderPoint || 0,
                wasteFactor: 0,
                lastPurchaseCost: ingredientCreateDraft.lastPurchaseCost || 0,
            }));
            setIngredientCreateDraft({
                name: '',
                unit: ingredientCreateDraft.unit || 'g',
                purchaseUnit: ingredientCreateDraft.purchaseUnit || 'kg',
                purchaseToUsage: ingredientCreateDraft.purchaseToUsage || 1000,
                onHandQty: 0,
                reorderPoint: 0,
                lastPurchaseCost: 0,
            });
        });
    }

    function markNotification(id: string) {
        return runAction(`notification-${id}`, 'Notification cleared', () => MarkNotificationRead(id));
    }

    if (!dashboard || !pilot) {
        return (
            <main className="boot-screen">
                <div className="boot-card">
                    <div className="brand-mark">NX</div>
                    <h1>NEXUS</h1>
                    <p>{error || 'Opening local operations ledger'}</p>
                    <button type="button" onClick={refreshAll}>Retry</button>
                </div>
            </main>
        );
    }

    if (!currentStaff) {
        return (
            <LoginScreen
                restaurant={dashboard.restaurant}
                staff={pilot.staff}
                draft={loginDraft}
                busy={busy}
                error={error}
                onDraft={setLoginDraft}
                onSignIn={signIn}
            />
        );
    }

    return (
        <main className={`app-shell workspace-${activeWorkspace}`}>
            <aside className="sidebar">
                <div className="brand-lockup">
                    <div className="brand-mark">NX</div>
                    <div>
                        <strong>NEXUS</strong>
                        <span>{dashboard.restaurant.website}</span>
                    </div>
                </div>

                <div className="workspace-switcher" aria-label="Role workspace">
                    {availableWorkspaces.map((workspace) => (
                        <button
                            type="button"
                            className={activeWorkspace === workspace.key ? 'active' : ''}
                            key={workspace.key}
                            onClick={() => switchWorkspace(workspace.key)}
                        >
                            <strong>{workspace.short}</strong>
                            <span>{workspace.label}</span>
                        </button>
                    ))}
                </div>

                <nav className="nav-list" aria-label="Workspace">
                    {groupedPages(activeWorkspace).map(([group, groupPages]) => (
                        <div className="nav-group" key={group}>
                            <span>{group}</span>
                            {groupPages.map((page) => (
                                <button
                                    className={activePage === page.key ? 'active' : ''}
                                    type="button"
                                    key={page.key}
                                    onClick={() => setActivePage(page.key)}
                                >
                                    {page.label}
                                </button>
                            ))}
                        </div>
                    ))}
                </nav>

                <div className="sync-card">
                    <span>Local queue</span>
                    <strong>{syncStatus?.pendingCount ?? dashboard.metrics.pendingSyncItems}</strong>
                    <small>{syncStatus?.failedCount ?? 0} failed / {pilot.printJobs.length} print jobs</small>
                </div>
            </aside>

            <section className="workspace">
                <header className="topbar">
                    <div>
                        <span className="eyebrow">{currentWorkspace.label} / {dashboard.restaurant.name}</span>
                        <h1>{currentPage.label}</h1>
                        <p>{currentWorkspace.detail}</p>
                    </div>
                    <div className="topbar-actions">
                        <div className="operator-chip">
                            <span>{humanStatus(currentStaff.role)}</span>
                            <strong>{currentStaff.name}</strong>
                        </div>
                        <button className="ghost-button" type="button" onClick={() => WindowFullscreen()}>Full Screen</button>
                        <button className="ghost-button" type="button" onClick={refreshAll} disabled={busy === 'refresh'}>Refresh</button>
                        <button className="ghost-button" type="button" onClick={signOut}>Sign Out</button>
                    </div>
                </header>

                {error && <div className="error-strip">{error}</div>}
                <ToastStack toasts={toasts}/>

                {activeWorkspace !== 'kitchenOps' && activePage !== 'admin' && <section className="metric-grid" aria-label="Today">
                    <Metric label="Sales" value={currency.format(dashboard.metrics.salesToday)} tone="ink"/>
                    <Metric label="Orders" value={String(dashboard.metrics.ordersToday)} tone="green"/>
                    <Metric label="Avg Bill" value={currency.format(dashboard.metrics.averageOrder)} tone="gold"/>
                    <Metric label="KOTs" value={String(dashboard.metrics.openKots)} tone="coral"/>
                </section>}

                {activePage === 'home' && (
                    <HomePage
                        dashboard={dashboard}
                        pilot={pilot}
                        invoices={invoices}
                        notifications={notifications}
                        onNavigate={navigateToPage}
                    />
                )}

                {activePage === 'counter' && (
                    <CounterPage
                        dashboard={dashboard}
                        pilot={pilot}
                        cart={cart}
                        itemById={itemById}
                        cartTotals={cartTotals}
                        customerName={customerName}
                        customerPhone={customerPhone}
                        orderType={orderType}
                        tableName={tableName}
                        discountType={discountType}
                        discountValue={discountValue}
                        paymentMethod={paymentMethod}
                        paymentTendered={paymentTendered}
                        menuSearch={menuSearch}
                        busy={busy}
                        openOrders={openOrders}
                        onAddToCart={addToCart}
                        onUpdateCartLine={updateCartLine}
                        onCustomerName={setCustomerName}
                        onCustomerPhone={setCustomerPhone}
                        onOrderType={setOrderType}
                        onTableName={setTableName}
                        onDiscountType={setDiscountType}
                        onDiscountValue={setDiscountValue}
                        onPaymentMethod={setPaymentMethod}
                        onPaymentTendered={setPaymentTendered}
                        onMenuSearch={setMenuSearch}
                        onSaveDraft={saveDraft}
                        onQuickClose={quickCloseBill}
                        onClearCart={clearCart}
                        onOpenInvoice={openInvoice}
                        onSendKOT={sendKOT}
                        onCloseInvoice={closeInvoice}
                    />
                )}

                {activePage === 'floor' && (
                    <FloorPage
                        dashboard={dashboard}
                        pilot={pilot}
                        selectedSession={selectedSession}
                        modifiersByItem={modifiersByItem}
                        busy={busy}
                        floorWaiterId={floorWaiterId}
                        floorGuestCount={floorGuestCount}
                        menuSearch={menuSearch}
                        lineVoidPIN={lineVoidPIN}
                        lineVoidReason={lineVoidReason}
                        onFloorWaiter={setFloorWaiterId}
                        onFloorGuestCount={setFloorGuestCount}
                        onMenuSearch={setMenuSearch}
                        onLineVoidPIN={setLineVoidPIN}
                        onLineVoidReason={setLineVoidReason}
                        onOpenTable={openTable}
                        onOpenSession={openSession}
                        onAssignSessionStaff={assignSessionStaff}
                        onAddSessionItem={addSessionItem}
                        onTransferSession={transferSession}
                        onSendKOT={sendSessionKOT}
                        onCloseSession={closeSession}
                        onVoidLine={voidSessionLine}
                    />
                )}

                {activePage === 'orders' && (
                    <OrdersPage
                        invoices={openOrders}
                        busy={busy}
                        onOpenInvoice={openInvoice}
                        onSendKOT={sendKOT}
                        onCloseInvoice={closeInvoice}
                    />
                )}

                {activePage === 'invoices' && (
                    <InvoicesPage
                        invoices={closedInvoices.length > 0 ? invoices : invoices}
                        selectedInvoice={selectedInvoice}
                        statusFilter={invoiceStatusFilter}
                        search={invoiceSearch}
                        paymentFilter={invoicePaymentFilter}
                        managerPIN={managerPIN}
                        actionReason={actionReason}
                        refundAmount={refundAmount}
                        splitMode={splitMode}
                        splitLineQty={splitLineQty}
                        splitAmountA={splitAmountA}
                        splitAmountB={splitAmountB}
                        busy={busy}
                        onStatusFilter={setInvoiceStatusFilter}
                        onSearch={setInvoiceSearch}
                        onPaymentFilter={setInvoicePaymentFilter}
                        onOpenInvoice={openInvoice}
                        onManagerPIN={setManagerPIN}
                        onActionReason={setActionReason}
                        onRefundAmount={setRefundAmount}
                        onValidatePIN={validatePIN}
                        onVoid={voidSelectedInvoice}
                        onRefund={refundSelectedInvoice}
                        onSplitMode={setSplitMode}
                        onSplitLineQty={(lineId, quantity) => setSplitLineQty((current) => ({...current, [lineId]: quantity}))}
                        onSplitAmountA={setSplitAmountA}
                        onSplitAmountB={setSplitAmountB}
                        onSplit={splitSelectedInvoice}
                        onExportPDF={exportSelectedInvoicePDF}
                        onPrintReceipt={() => selectedInvoice && queuePrint('invoice', selectedInvoice.summary.id)}
                        onExportInvoices={exportInvoicesCSV}
                    />
                )}

                {activePage === 'kitchen' && <KitchenPage tickets={dashboard.kitchenTickets} printJobs={pilot.printJobs} onUpdateTicket={updateTicket} onQueuePrint={queuePrint}/>}

                {activePage === 'admin' && (
                    <AdminCommandCenter
                        analytics={analytics}
                        dashboard={dashboard}
                        pilot={pilot}
                        range={analyticsRange}
                        onRange={setAnalyticsRange}
                        onRefresh={refreshAnalytics}
                        onNavigate={navigateToPage}
                    />
                )}

                {activePage === 'inventory' && (
                    <InventoryPage
                        dashboard={dashboard}
                        ingredientDraft={ingredientDraft}
                        ingredientCreateDraft={ingredientCreateDraft}
                        activeIngredientId={activeIngredientId}
                        busy={busy}
                        onActiveIngredient={setActiveIngredientId}
                        onIngredientDraft={setIngredientDraft}
                        onSaveIngredient={saveIngredientSettings}
                        onIngredientCreateDraft={setIngredientCreateDraft}
                        onSaveNewIngredient={saveNewIngredient}
                        onReconcileMilk={reconcileMilk}
                        onReceiveBadBatch={receiveBadBatch}
                    />
                )}

                {activePage === 'vendors' && (
                    <VendorsPage
                        pilot={pilot}
                        ingredients={dashboard.ingredients}
                        vendorDraft={vendorDraft}
                        poDraft={poDraft}
                        busy={busy}
                        onVendorDraft={setVendorDraft}
                        onSaveVendor={saveVendorDraft}
                        onPODraft={setPODraft}
                        onAddPOLine={addPOLine}
                        onUpdatePOLine={updatePOLine}
                        onRemovePOLine={removePOLine}
                        onCreatePO={createPurchaseOrderDraft}
                        onReceivePO={receiveFirstPurchaseOrder}
                    />
                )}

                {activePage === 'recipes' && (
                    <RecipesPage
                        dashboard={dashboard}
                        pilot={pilot}
                        activeRecipeId={activeRecipeId}
                        recipeDraft={recipeDraft}
                        busy={busy}
                        menuItemDraft={menuItemDraft}
                        categoryDraft={categoryDraft}
                        modifierDraft={modifierDraft}
                        modifierLinkDraft={modifierLinkDraft}
                        onActiveRecipe={setActiveRecipeId}
                        onAddLine={addRecipeLine}
                        onUpdateLine={updateRecipeLine}
                        onRemoveLine={removeRecipeLine}
                        onSaveRecipe={saveRecipe}
                        onMenuItemDraft={setMenuItemDraft}
                        onSaveMenuItem={saveMenuItemDraft}
                        onCategoryDraft={setCategoryDraft}
                        onSaveCategory={saveCategoryDraft}
                        onModifierDraft={setModifierDraft}
                        onSaveModifier={saveModifierDraft}
                        onModifierLinkDraft={setModifierLinkDraft}
                        onLinkModifier={linkModifierDraft}
                        menuImportText={menuImportText}
                        onMenuImportText={setMenuImportText}
                        onImportMenu={importMenuRows}
                        onSaveMenuItemPrice={saveMenuItemPrice}
                        onSaveMenuItemPatch={saveMenuItemPatch}
                    />
                )}

                {activePage === 'customers' && <CustomersPage customers={dashboard.customers} signals={topSignals}/>}

                {activePage === 'marketing' && (
                    <MarketingPage
                        drafts={dashboard.marketingDrafts}
                        notifications={notifications}
                        busy={busy}
                        onMarkRead={markNotification}
                    />
                )}

                {activePage === 'integrations' && (
                    <IntegrationsPage
                        pilot={pilot}
                        invoices={invoices}
                        busy={busy}
                        onCreatePayment={createPayment}
                        onMarkPaymentPaid={markPaymentPaid}
                        onMarkPaymentFailed={markPaymentFailed}
                        onCancelPayment={cancelPayment}
                        onQueuePrint={queuePrint}
                        onMarkPrintPrinted={markPrintPrinted}
                        onMarkPrintFailed={markPrintFailed}
                        onRetryPrint={retryPrint}
                        onStoreIntegration={storeIntegration}
                        printerTarget={printerTarget}
                        onPrinterTarget={setPrinterTarget}
                        onSavePrinter={savePrinter}
                    />
                )}

                {activePage === 'dayclose' && (
                    <DayClosePage
                        summary={pilot.dayClose}
                        cashCounted={dayCloseCash}
                        closePIN={dayClosePIN}
                        busy={busy}
                        onCashCounted={setDayCloseCash}
                        onClosePIN={setDayClosePIN}
                        onCloseDay={closeDayNow}
                        onExportPDF={exportDayClosePDF}
                    />
                )}

                {activePage === 'accounting' && (
                    <AccountingPage
                        snapshot={pilot.accounting}
                        busy={busy}
                        onRebuild={rebuildAccounting}
                        onExportCSV={exportAccountingCSV}
                    />
                )}

                {activePage === 'settings' && (
                    <SettingsPage
                        metrics={dashboard.metrics}
                        notifications={notifications}
                        pilot={pilot}
                        syncStatus={syncStatus}
                        settingsDraft={settingsDraft}
                        staffDraft={staffDraft}
                        busy={busy}
                        onSettingsDraft={setSettingsDraft}
                        onSaveSettings={saveSettings}
                        onStaffDraft={setStaffDraft}
                        onSaveStaff={saveStaffDraft}
                        onUpdateStaffStatus={updateStaffStatus}
                        onExportBackup={exportBackup}
                        readinessItems={buildReadinessItems(dashboard, pilot)}
                    />
                )}
            </section>
        </main>
    );
}

function LoginScreen({restaurant, staff, draft, busy, error, onDraft, onSignIn}: {
    restaurant: Restaurant;
    staff: StaffMember[];
    draft: LoginDraft;
    busy: string;
    error: string;
    onDraft: (value: LoginDraft | ((current: LoginDraft) => LoginDraft)) => void;
    onSignIn: () => void;
}) {
    const activeStaff = staff.filter((member) => member.status === 'active');
    const selected = activeStaff.find((member) => member.id === draft.staffId) ?? activeStaff[0];
    const allowedWorkspaces = workspaceAccessForRole(selected?.role ?? 'waiter');
    return (
        <main className="login-shell">
            <section className="login-panel">
                <div className="brand-lockup">
                    <div className="brand-mark">NX</div>
                    <div>
                        <strong>NEXUS</strong>
                        <span>{restaurant.name}</span>
                    </div>
                </div>
                <form
                    className="login-form"
                    onSubmit={(event) => {
                        event.preventDefault();
                        void onSignIn();
                    }}
                >
                    <Field label="Staff">
                        <select
                            value={selected?.id ?? ''}
                            onChange={(event) => {
                                const nextStaff = activeStaff.find((member) => member.id === event.target.value);
                                const nextAccess = workspaceAccessForRole(nextStaff?.role ?? 'waiter');
                                onDraft((current) => ({...current, staffId: event.target.value, workspace: nextAccess[0] ?? 'front'}));
                            }}
                        >
                            {activeStaff.map((member) => (
                                <option value={member.id} key={member.id}>{member.name} / {humanStatus(member.role)}</option>
                            ))}
                        </select>
                    </Field>
                    <Field label="PIN">
                        <input
                            type="password"
                            value={draft.pin}
                            onChange={(event) => onDraft((current) => ({...current, pin: event.target.value}))}
                            autoFocus
                        />
                    </Field>
                    <div className="workspace-login-grid">
                        {workspaces.filter((workspace) => allowedWorkspaces.includes(workspace.key)).map((workspace) => (
                            <button
                                type="button"
                                className={draft.workspace === workspace.key ? 'active' : ''}
                                key={workspace.key}
                                onClick={() => onDraft((current) => ({...current, workspace: workspace.key}))}
                            >
                                <strong>{workspace.short}</strong>
                                <span>{workspace.label}</span>
                            </button>
                        ))}
                    </div>
                    {error && <div className="error-strip">{error}</div>}
                    <button type="submit" disabled={busy === 'login' || !selected}>Sign In</button>
                </form>
            </section>
        </main>
    );
}

function HomePage({dashboard, pilot, invoices, notifications, onNavigate}: {
    dashboard: Dashboard;
    pilot: PilotWorkspace;
    invoices: InvoiceSummary[];
    notifications: NotificationRecord[];
    onNavigate: (page: PageKey) => void;
}) {
    const openBills = invoices.filter((invoice) => ['draft', 'open', 'kot_sent'].includes(invoice.status)).length;
    const occupiedTables = pilot.tables.filter((table) => table.status === 'occupied').length;
    const pendingMessages = notifications.filter((notification) => !notification.readAt).length;
    const readinessItems = buildReadinessItems(dashboard, pilot);
    const readinessDone = readinessItems.filter((item) => item.done).length;
    const readinessScore = Math.round((readinessDone / readinessItems.length) * 100);
    const nextReadiness = readinessItems.find((item) => !item.done);
    return (
        <section className="home-grid">
            <button className="flow-card primary" type="button" onClick={() => onNavigate('counter')}>
                <span>Walk-in and takeaway</span>
                <strong>Quick bill</strong>
                <em>{currency.format(dashboard.metrics.averageOrder)} avg</em>
            </button>
            <button className="flow-card" type="button" onClick={() => onNavigate('floor')}>
                <span>Dining room</span>
                <strong>Tables</strong>
                <em>{occupiedTables}/{pilot.tables.length} occupied</em>
            </button>
            <button className="flow-card" type="button" onClick={() => onNavigate('orders')}>
                <span>Unfinished bills</span>
                <strong>Open bills</strong>
                <em>{openBills} active</em>
            </button>
            <button className="flow-card" type="button" onClick={() => onNavigate('kitchen')}>
                <span>Preparation queue</span>
                <strong>Kitchen</strong>
                <em>{dashboard.metrics.openKots} KOTs</em>
            </button>
            <button className="flow-card" type="button" onClick={() => onNavigate('invoices')}>
                <span>Receipts and refunds</span>
                <strong>Bills</strong>
                <em>{invoices.length} records</em>
            </button>
            <button className="flow-card" type="button" onClick={() => onNavigate('inventory')}>
                <span>Ingredients</span>
                <strong>Stock</strong>
                <em>{dashboard.metrics.stockAlerts} alerts</em>
            </button>
            <button className="flow-card" type="button" onClick={() => onNavigate('integrations')}>
                <span>Printer, WhatsApp, Razorpay</span>
                <strong>Devices</strong>
                <em>{pilot.printJobs.length} print jobs</em>
            </button>
            <button className="flow-card" type="button" onClick={() => onNavigate('dayclose')}>
                <span>Cash and settlement</span>
                <strong>Close day</strong>
                <em>{currency.format(pilot.dayClose.cashExpected)} cash</em>
            </button>
            <Panel title="Pilot Readiness" action={`${readinessScore}% ready`}>
                <div className="readiness-summary">
                    <strong>{nextReadiness ? nextReadiness.title : 'Pilot path is ready'}</strong>
                    <span>{nextReadiness ? nextReadiness.detail : 'Core setup is in place for the single-terminal demo flow.'}</span>
                    {nextReadiness && (
                        <button type="button" onClick={() => onNavigate(nextReadiness.page)}>Open {pages.find((page) => page.key === nextReadiness.page)?.label ?? 'Setup'}</button>
                    )}
                </div>
                <div className="readiness-list">
                    {readinessItems.map((item) => (
                        <button
                            type="button"
                            className={`readiness-row ${item.done ? 'done' : 'todo'}`}
                            key={item.id}
                            onClick={() => onNavigate(item.page)}
                        >
                            <span>{item.done ? 'Done' : 'Next'}</span>
                            <strong>{item.title}</strong>
                        </button>
                    ))}
                </div>
            </Panel>
            <Panel title="Today Signals" action={`${dashboard.signals.length} live`}>
                <div className="signal-list">
                    {dashboard.signals.slice(0, 3).map((signal) => (
                        <article className={`signal-card priority-${signal.priority}`} key={signal.id}>
                            <span>{signal.kind}</span>
                            <strong>{signal.title}</strong>
                            <p>{signal.detail}</p>
                        </article>
                    ))}
                    {pendingMessages > 0 && (
                        <article className="signal-card priority-3">
                            <span>WhatsApp</span>
                            <strong>{pendingMessages} queued messages</strong>
                            <p>Receipts and customer messages are waiting in the local queue.</p>
                        </article>
                    )}
                </div>
            </Panel>
        </section>
    );
}

function CounterPage(props: {
    dashboard: Dashboard;
    pilot: PilotWorkspace;
    cart: CartLine[];
    itemById: Map<string, MenuItem>;
    cartTotals: {subtotal: number; discount: number; tax: number; total: number};
    customerName: string;
    customerPhone: string;
    orderType: string;
    tableName: string;
    discountType: string;
    discountValue: number;
    paymentMethod: string;
    paymentTendered: number;
    menuSearch: string;
    busy: string;
    openOrders: InvoiceSummary[];
    onAddToCart: (item: MenuItem) => void;
    onUpdateCartLine: (itemId: string, patch: Partial<CartLine>) => void;
    onCustomerName: (value: string) => void;
    onCustomerPhone: (value: string) => void;
    onOrderType: (value: string) => void;
    onTableName: (value: string) => void;
    onDiscountType: (value: string) => void;
    onDiscountValue: (value: number) => void;
    onPaymentMethod: (value: string) => void;
    onPaymentTendered: (value: number) => void;
    onMenuSearch: (value: string) => void;
    onSaveDraft: () => void;
    onQuickClose: () => void;
    onClearCart: () => void;
    onOpenInvoice: (id: string) => void;
    onSendKOT: (id: string) => void;
    onCloseInvoice: (id: string) => void;
}) {
    const query = props.menuSearch.trim().toLowerCase();
    const saleItems = props.dashboard.menuItems.filter((item) => item.status !== 'hidden' && (
        query === '' ||
        item.name.toLowerCase().includes(query) ||
        item.category.toLowerCase().includes(query) ||
        item.routeName.toLowerCase().includes(query)
    ));
    return (
        <section className="counter-grid">
            <Panel title="Counter" action={`${saleItems.length} sale items`}>
                <div className="toolbar-row">
                    <Field label="Search">
                        <input value={props.menuSearch} onChange={(event) => props.onMenuSearch(event.target.value)} placeholder="Find item"/>
                    </Field>
                </div>
                <div className="menu-grid">
                    {saleItems.map((item) => (
                        <button
                            className={`menu-tile ${item.status}`}
                            type="button"
                            key={item.id}
                            disabled={item.status !== 'active'}
                            onClick={() => props.onAddToCart(item)}
                        >
                            <span>{item.category} / {item.routeName}</span>
                            <strong>{item.name}</strong>
                            <em>{item.status === 'active' ? currency.format(item.price) : humanStatus(item.status)}</em>
                        </button>
                    ))}
                </div>
            </Panel>

            <Panel title="Bill" action={`${props.cart.length} lines`}>
                <CartEditor {...props}/>
            </Panel>

            <Panel title="Open Orders" action={`${props.openOrders.length} active`}>
                <InvoiceStack
                    invoices={props.openOrders.slice(0, 7)}
                    empty="No open orders"
                    onOpen={props.onOpenInvoice}
                    renderActions={(invoice) => (
                        <>
                            {invoice.status === 'draft' && <button type="button" onClick={() => props.onSendKOT(invoice.id)}>KOT</button>}
                            {invoice.status !== 'closed' && <button type="button" onClick={() => props.onCloseInvoice(invoice.id)}>Close</button>}
                        </>
                    )}
                />
            </Panel>
        </section>
    );
}

function FloorPage(props: {
    dashboard: Dashboard;
    pilot: PilotWorkspace;
    selectedSession: OrderSessionDetail | null;
    modifiersByItem: Map<string, MenuModifier[]>;
    busy: string;
    floorWaiterId: string;
    floorGuestCount: number;
    menuSearch: string;
    lineVoidPIN: string;
    lineVoidReason: string;
    onFloorWaiter: (value: string) => void;
    onFloorGuestCount: (value: number) => void;
    onMenuSearch: (value: string) => void;
    onLineVoidPIN: (value: string) => void;
    onLineVoidReason: (value: string) => void;
    onOpenTable: (tableID: string, staffID?: string) => void;
    onOpenSession: (sessionID: string) => void;
    onAssignSessionStaff: (sessionID: string, staffID: string) => void;
    onAddSessionItem: (itemID: string, modifierIDs?: string[]) => void;
    onTransferSession: (tableID: string) => void;
    onSendKOT: () => void;
    onCloseSession: () => void;
    onVoidLine: (lineID: string) => void;
}) {
    const selected = props.selectedSession;
    const [focusedTableId, setFocusedTableId] = useState('');
    const [sectionFilter, setSectionFilter] = useState('all');
    const [dragStaffId, setDragStaffId] = useState('');
    const sessionById = new Map(props.pilot.orderSessions.map((session) => [session.id, session]));
    const occupiedSessions = props.pilot.orderSessions.filter((session) => ['open', 'held'].includes(session.status));
    const occupiedCount = props.pilot.tables.filter((table) => table.status === 'occupied').length;
    const focusedTable = props.pilot.tables.find((table) => table.id === focusedTableId) ?? null;
    const modalSession = focusedTable?.activeSessionId && selected?.session.id === focusedTable.activeSessionId ? selected : null;
    const modalSummary = focusedTable?.activeSessionId ? sessionById.get(focusedTable.activeSessionId) : undefined;
    const availableTargets = props.pilot.tables.filter((table) => table.status === 'available' && table.id !== focusedTable?.id);
    const query = props.menuSearch.trim().toLowerCase();
    const saleItems = props.dashboard.menuItems.filter((item) => item.status !== 'hidden' && (
        query === '' ||
        item.name.toLowerCase().includes(query) ||
        item.category.toLowerCase().includes(query) ||
        item.routeName.toLowerCase().includes(query)
    ));
    const activeStaff = props.pilot.staff.filter((member) => member.status === 'active');
    const staffDirectory = activeStaff.filter((member) => ['owner', 'manager', 'cashier', 'waiter', 'runner'].includes(member.role));
    const assignableStaff = staffDirectory.length > 0 ? staffDirectory : activeStaff;
    const selectedStaff = assignableStaff.find((member) => member.id === props.floorWaiterId) ?? assignableStaff[0];
    const floorSections = props.pilot.floorSections.length > 0
        ? props.pilot.floorSections
        : Array.from(new Map(props.pilot.tables.map((table) => [table.sectionId, {id: table.sectionId, name: table.sectionName, sortOrder: 0}])).values());
    const visibleTables = sectionFilter === 'all'
        ? props.pilot.tables
        : props.pilot.tables.filter((table) => table.sectionId === sectionFilter);

    useEffect(() => {
        if (focusedTableId && !props.pilot.tables.some((table) => table.id === focusedTableId)) {
            setFocusedTableId('');
        }
    }, [focusedTableId, props.pilot.tables]);

    function focusTable(table: DiningTable) {
        setFocusedTableId(table.id);
        if (table.activeSessionId) {
            props.onOpenSession(table.activeSessionId);
        }
    }

    function assignStaffToTable(table: DiningTable, staffID: string) {
        if (!staffID) return;
        props.onFloorWaiter(staffID);
        setFocusedTableId(table.id);
        if (table.activeSessionId) {
            props.onAssignSessionStaff(table.activeSessionId, staffID);
        } else {
            props.onOpenTable(table.id, staffID);
        }
    }

    function handleTableDrop(event: DragEvent<HTMLButtonElement>, table: DiningTable) {
        event.preventDefault();
        const staffID = event.dataTransfer.getData('text/staff-id') || dragStaffId;
        setDragStaffId('');
        assignStaffToTable(table, staffID);
    }

    function pickStaff(staffID: string) {
        props.onFloorWaiter(staffID);
        if (focusedTable?.activeSessionId) {
            props.onAssignSessionStaff(focusedTable.activeSessionId, staffID);
        }
    }

    return (
        <section className="floor-grid floor-grid-touch">
            <Panel title="Floor" action={`${occupiedCount}/${props.pilot.tables.length} occupied`}>
                <div className="floor-console">
                    <div className="floor-console-head">
                        <div className="section-rail" aria-label="Floor sections">
                            <button type="button" className={sectionFilter === 'all' ? 'active' : ''} onClick={() => setSectionFilter('all')}>
                                All
                            </button>
                            {floorSections.map((section) => (
                                <button
                                    type="button"
                                    className={sectionFilter === section.id ? 'active' : ''}
                                    key={section.id}
                                    onClick={() => setSectionFilter(section.id)}
                                >
                                    {section.name}
                                </button>
                            ))}
                        </div>
                        <div className="table-assignment-bar">
                            <Field label="Server">
                                <select value={props.floorWaiterId} onChange={(event) => props.onFloorWaiter(event.target.value)}>
                                    {assignableStaff.length === 0 && <option value="">Add staff in Settings</option>}
                                    {assignableStaff.map((member) => (
                                        <option value={member.id} key={member.id}>
                                            {member.name} / {humanStatus(member.role)}
                                        </option>
                                    ))}
                                </select>
                            </Field>
                            <Field label="Guests">
                                <input
                                    type="number"
                                    min="1"
                                    value={props.floorGuestCount}
                                    onChange={(event) => props.onFloorGuestCount(Math.max(1, Number(event.target.value) || 1))}
                                />
                            </Field>
                        </div>
                    </div>

                    <div className="floor-touch-layout">
                        <div className="table-directory" aria-label="Tables directory">
                            {visibleTables.map((table) => {
                                const session = table.activeSessionId ? sessionById.get(table.activeSessionId) : undefined;
                                const active = Boolean(focusedTable?.id === table.id);
                                return (
                                    <button
                                        type="button"
                                        className={`table-chip ${table.status} ${active ? 'active' : ''} ${dragStaffId ? 'drop-ready' : ''}`}
                                        key={table.id}
                                        onClick={() => focusTable(table)}
                                        onDragOver={(event) => event.preventDefault()}
                                        onDrop={(event) => handleTableDrop(event, table)}
                                    >
                                        <span>{table.sectionName}</span>
                                        <strong>{table.name}</strong>
                                        <StatusBadge status={session ? sessionKitchenStatus(session) : table.status}/>
                                        <em>{session ? `${session.waiterName || 'Floor'} / ${currency.format(session.total)}` : `${table.seats} seats`}</em>
                                    </button>
                                );
                            })}
                        </div>

                        <aside className="staff-dock">
                            <div className="staff-dock-head">
                                <strong>Staff</strong>
                                <span>{assignableStaff.length}</span>
                            </div>
                            <div className="staff-chip-list">
                                {assignableStaff.map((member) => (
                                    <button
                                        type="button"
                                        className={`staff-chip ${props.floorWaiterId === member.id ? 'active' : ''}`}
                                        key={member.id}
                                        draggable
                                        onClick={() => props.onFloorWaiter(member.id)}
                                        onDragStart={(event) => {
                                            event.dataTransfer.setData('text/staff-id', member.id);
                                            event.dataTransfer.effectAllowed = 'copyMove';
                                            setDragStaffId(member.id);
                                        }}
                                        onDragEnd={() => setDragStaffId('')}
                                    >
                                        <strong>{member.name}</strong>
                                        <span>{humanStatus(member.role)}</span>
                                    </button>
                                ))}
                            </div>
                        </aside>
                    </div>
                </div>
            </Panel>

            {focusedTable && (
                <div className="table-modal-backdrop" role="dialog" aria-modal="true" aria-label={`${focusedTable.name} table`} onMouseDown={(event) => {
                    if (event.target === event.currentTarget) setFocusedTableId('');
                }}>
                    <div
                        className={`table-modal ${focusedTable.status}`}
                        onDragOver={(event) => event.preventDefault()}
                        onDrop={(event) => {
                            const staffID = event.dataTransfer.getData('text/staff-id') || dragStaffId;
                            setDragStaffId('');
                            assignStaffToTable(focusedTable, staffID);
                        }}
                    >
                        <div className="table-modal-head">
                            <div>
                                <span>{focusedTable.sectionName}</span>
                                <h2>{focusedTable.name}</h2>
                            </div>
                            <div className="table-modal-actions">
                                <StatusBadge status={modalSummary ? sessionKitchenStatus(modalSummary) : focusedTable.status}/>
                                <button className="ghost-button" type="button" onClick={() => setFocusedTableId('')}>Close</button>
                            </div>
                        </div>

                        <div className="table-modal-grid">
                            <div className="table-modal-main">
                                <div className="detail-head table-detail-head">
                                    <div>
                                        <span>Server</span>
                                        <strong>{modalSession?.session.waiterName || modalSummary?.waiterName || selectedStaff?.name || 'Unassigned'}</strong>
                                    </div>
                                    <div>
                                        <span>Guests</span>
                                        <strong>{modalSession?.session.guestCount ?? modalSummary?.guestCount ?? props.floorGuestCount}</strong>
                                    </div>
                                    <div>
                                        <span>Seats</span>
                                        <strong>{focusedTable.seats}</strong>
                                    </div>
                                    <div>
                                        <span>Total</span>
                                        <strong>{modalSession || modalSummary ? currency.format((modalSession?.session ?? modalSummary)?.total ?? 0) : '-'}</strong>
                                    </div>
                                </div>

                                {!focusedTable.activeSessionId && (
                                    <div className="empty-table-card">
                                        <DataRow label="Server" value={selectedStaff?.name ?? 'Unassigned'}/>
                                        <DataRow label="Guests" value={String(props.floorGuestCount)}/>
                                        <button type="button" disabled={!selectedStaff || props.busy === `open-table-${focusedTable.id}`} onClick={() => props.onOpenTable(focusedTable.id)}>
                                            Open Table
                                        </button>
                                    </div>
                                )}

                                {focusedTable.activeSessionId && !modalSession && (
                                    <p className="empty-copy">Loading table</p>
                                )}

                                {modalSession && (
                                    <>
                                        <div className="line-stack table-modal-lines">
                                            {modalSession.lines.length === 0 && <p className="empty-copy">No items</p>}
                                            {modalSession.lines.map((line) => (
                                                <article className="session-line" key={line.id}>
                                                    <div>
                                                        <strong>{line.quantity} x {line.itemName}</strong>
                                                        <span>{line.modifierNames.join(', ') || 'Base item'}</span>
                                                    </div>
                                                    <div className="session-line-side">
                                                        <StatusBadge status={line.kotStatus && line.kotStatus !== 'not_sent' ? line.kotStatus : line.status}/>
                                                        <em>{currency.format(line.lineTotal)}</em>
                                                        {line.status === 'open' && (
                                                            <button className="ghost-button compact" type="button" onClick={() => props.onVoidLine(line.id)}>
                                                                Void
                                                            </button>
                                                        )}
                                                    </div>
                                                </article>
                                            ))}
                                        </div>

                                        <div className="modal-menu-panel">
                                            <div className="toolbar-row">
                                                <Field label="Search">
                                                    <input value={props.menuSearch} onChange={(event) => props.onMenuSearch(event.target.value)} placeholder="Find item"/>
                                                </Field>
                                            </div>
                                            <div className="modal-menu-grid">
                                                {saleItems.map((item) => {
                                                    const modifiers = props.modifiersByItem.get(item.id) ?? [];
                                                    return (
                                                        <article className={`session-menu-tile ${item.status}`} key={item.id}>
                                                            <button type="button" disabled={item.status !== 'active'} onClick={() => props.onAddSessionItem(item.id)}>
                                                                <span>{item.routeName}</span>
                                                                <strong>{item.name}</strong>
                                                                <em>{item.status === 'active' ? currency.format(item.price) : humanStatus(item.status)}</em>
                                                            </button>
                                                            {item.status === 'active' && modifiers.length > 0 && (
                                                                <div className="modifier-row">
                                                                    {modifiers.slice(0, 2).map((modifier) => (
                                                                        <button className="ghost-button" type="button" key={modifier.id} onClick={() => props.onAddSessionItem(item.id, [modifier.id])}>
                                                                            {modifier.name}
                                                                        </button>
                                                                    ))}
                                                                </div>
                                                            )}
                                                        </article>
                                                    );
                                                })}
                                            </div>
                                        </div>
                                    </>
                                )}
                            </div>

                            <aside className="table-modal-side">
                                <section className="modal-side-section">
                                    <div className="staff-dock-head">
                                        <strong>Staff</strong>
                                        <span>{assignableStaff.length}</span>
                                    </div>
                                    <div className="modal-staff-list">
                                        {assignableStaff.map((member) => {
                                            const assigned = (modalSession?.session.waiterId || modalSummary?.waiterId || props.floorWaiterId) === member.id;
                                            return (
                                                <button
                                                    type="button"
                                                    className={`staff-chip ${assigned ? 'active' : ''}`}
                                                    key={member.id}
                                                    draggable
                                                    onClick={() => pickStaff(member.id)}
                                                    onDragStart={(event) => {
                                                        event.dataTransfer.setData('text/staff-id', member.id);
                                                        event.dataTransfer.effectAllowed = 'copyMove';
                                                        setDragStaffId(member.id);
                                                    }}
                                                    onDragEnd={() => setDragStaffId('')}
                                                >
                                                    <strong>{member.name}</strong>
                                                    <span>{humanStatus(member.role)}</span>
                                                </button>
                                            );
                                        })}
                                    </div>
                                </section>

                                {modalSession && (
                                    <>
                                        <section className="modal-side-section">
                                            <div className="totals-card">
                                                <MoneyRow label="Subtotal" value={modalSession.session.subtotal}/>
                                                <MoneyRow label="Service" value={modalSession.session.serviceCharge}/>
                                                <MoneyRow label="GST" value={modalSession.session.taxTotal}/>
                                                <MoneyRow label="Total" value={modalSession.session.total} strong/>
                                            </div>
                                            <div className="button-row">
                                                <button type="button" disabled={props.busy === 'session-kot'} onClick={props.onSendKOT}>Send KOT</button>
                                                <button type="button" disabled={props.busy === 'session-close'} onClick={props.onCloseSession}>Close Table</button>
                                            </div>
                                        </section>

                                        <section className="modal-side-section">
                                            <div className="approval-strip">
                                                <Field label="Void PIN">
                                                    <input
                                                        type="password"
                                                        value={props.lineVoidPIN}
                                                        onChange={(event) => props.onLineVoidPIN(event.target.value)}
                                                        placeholder="Manager PIN"
                                                    />
                                                </Field>
                                                <Field label="Reason">
                                                    <input
                                                        value={props.lineVoidReason}
                                                        onChange={(event) => props.onLineVoidReason(event.target.value)}
                                                        placeholder="Guest changed order"
                                                    />
                                                </Field>
                                            </div>
                                        </section>

                                        <section className="modal-side-section">
                                            <div className="transfer-row">
                                                {availableTargets.slice(0, 6).map((table) => (
                                                    <button className="ghost-button" type="button" key={table.id} onClick={() => {
                                                        setFocusedTableId(table.id);
                                                        props.onTransferSession(table.id);
                                                    }}>
                                                        Move {table.name}
                                                    </button>
                                                ))}
                                            </div>
                                        </section>
                                    </>
                                )}
                            </aside>
                        </div>
                    </div>
                </div>
            )}

        </section>
    );
}

function CartEditor(props: {
    dashboard: Dashboard;
    pilot: PilotWorkspace;
    cart: CartLine[];
    itemById: Map<string, MenuItem>;
    cartTotals: {subtotal: number; discount: number; tax: number; total: number};
    customerName: string;
    customerPhone: string;
    orderType: string;
    tableName: string;
    discountType: string;
    discountValue: number;
    paymentMethod: string;
    paymentTendered: number;
    busy: string;
    onUpdateCartLine: (itemId: string, patch: Partial<CartLine>) => void;
    onCustomerName: (value: string) => void;
    onCustomerPhone: (value: string) => void;
    onOrderType: (value: string) => void;
    onTableName: (value: string) => void;
    onDiscountType: (value: string) => void;
    onDiscountValue: (value: number) => void;
    onPaymentMethod: (value: string) => void;
    onPaymentTendered: (value: number) => void;
    onSaveDraft: () => void;
    onQuickClose: () => void;
    onClearCart: () => void;
}) {
    const serviceModes = [
        {id: 'dine_in', label: 'Dine in'},
        {id: 'takeaway', label: 'Takeaway'},
        {id: 'delivery', label: 'Delivery'},
    ];
    const paymentOptions = [
        {id: 'cash', label: 'Cash'},
        {id: 'upi', label: 'UPI'},
        {id: 'card', label: 'Card'},
        {id: 'razorpay', label: 'Razorpay'},
    ];
    const discountPresets = [
        {label: 'None', type: 'none', value: 0},
        {label: '5%', type: 'percent', value: 5},
        {label: '10%', type: 'percent', value: 10},
        {label: '₹50', type: 'fixed', value: 50},
    ];
    const recentCustomers = props.dashboard.customers.slice(0, 4);
    return (
        <>
            <div className="cart-list">
                {props.cart.length === 0 && <p className="empty-copy">No items added</p>}
                {props.cart.map((line) => {
                    const item = props.itemById.get(line.itemId);
                    if (!item) return null;
                    return (
                        <article className="cart-row" key={line.itemId}>
                            <div>
                                <strong>{item.name}</strong>
                                <span>{currency.format(item.price)} / {item.routeName}</span>
                            </div>
                            <input
                                aria-label={`${item.name} quantity`}
                                min="0"
                                type="number"
                                value={line.quantity}
                                onChange={(event) => props.onUpdateCartLine(line.itemId, {quantity: Number(event.target.value)})}
                            />
                            <input
                                aria-label={`${item.name} notes`}
                                placeholder="Notes"
                                value={line.notes}
                                onChange={(event) => props.onUpdateCartLine(line.itemId, {notes: event.target.value})}
                            />
                        </article>
                    );
                })}
            </div>

            <div className="tap-stack">
                <MiniSection title="Service">
                    <div className="tap-grid three">
                        {serviceModes.map((mode) => (
                            <button
                                type="button"
                                className={`tap-button ${props.orderType === mode.id ? 'active' : ''}`}
                                key={mode.id}
                                onClick={() => props.onOrderType(mode.id)}
                            >
                                {mode.label}
                            </button>
                        ))}
                    </div>
                    {props.orderType === 'dine_in' && (
                        <select value={props.tableName} onChange={(event) => props.onTableName(event.target.value)}>
                            {props.pilot.tables.map((table) => (
                                <option value={table.name} key={table.id}>
                                    {table.name} / {table.sectionName} / {humanStatus(table.status)}
                                </option>
                            ))}
                        </select>
                    )}
                </MiniSection>

                <MiniSection title="Customer">
                    <div className="tap-grid">
                        <button
                            type="button"
                            className={`tap-button ${props.customerName === 'Walk-in Guest' ? 'active' : ''}`}
                            onClick={() => {
                                props.onCustomerName('Walk-in Guest');
                                props.onCustomerPhone('');
                            }}
                        >
                            Walk-in
                        </button>
                        {recentCustomers.map((customer) => (
                            <button
                                type="button"
                                className={`tap-button ${props.customerPhone === customer.phone ? 'active' : ''}`}
                                key={customer.id}
                                onClick={() => {
                                    props.onCustomerName(customer.name);
                                    props.onCustomerPhone(customer.phone);
                                }}
                            >
                                {customer.name}
                            </button>
                        ))}
                    </div>
                    <div className="form-grid compact">
                        <input value={props.customerName} onChange={(event) => props.onCustomerName(event.target.value)} placeholder="Customer"/>
                        <input value={props.customerPhone} onChange={(event) => props.onCustomerPhone(event.target.value)} placeholder="Phone"/>
                    </div>
                </MiniSection>

                <MiniSection title="Discount">
                    <div className="tap-grid four">
                        {discountPresets.map((preset) => {
                            const active = props.discountType === preset.type && props.discountValue === preset.value;
                            return (
                                <button
                                    type="button"
                                    className={`tap-button ${active ? 'active' : ''}`}
                                    key={`${preset.type}-${preset.value}`}
                                    onClick={() => {
                                        props.onDiscountType(preset.type);
                                        props.onDiscountValue(preset.value);
                                    }}
                                >
                                    {preset.label}
                                </button>
                            );
                        })}
                    </div>
                </MiniSection>

                <MiniSection title="Payment">
                    <div className="tap-grid four">
                        {paymentOptions.map((method) => (
                            <button
                                type="button"
                                className={`tap-button ${props.paymentMethod === method.id ? 'active' : ''}`}
                                key={method.id}
                                onClick={() => props.onPaymentMethod(method.id)}
                            >
                                {method.label}
                            </button>
                        ))}
                    </div>
                    <div className="form-grid compact">
                        <button type="button" className="ghost-button" onClick={() => props.onPaymentTendered(props.cartTotals.total)}>Exact tender</button>
                        <input type="number" min="0" value={props.paymentTendered} onChange={(event) => props.onPaymentTendered(Number(event.target.value))}/>
                    </div>
                </MiniSection>
            </div>

            <div className="totals-card">
                <MoneyRow label="Subtotal" value={props.cartTotals.subtotal}/>
                <MoneyRow label="Discount" value={props.cartTotals.discount}/>
                <MoneyRow label="Tax" value={props.cartTotals.tax}/>
                <MoneyRow label="Total" value={props.cartTotals.total} strong/>
            </div>

            <p className="stock-hint">Stock is reduced once when KOT is sent or a quick bill is closed. Drafts do not move stock.</p>

            <div className="button-row">
                <button type="button" onClick={props.onSaveDraft} disabled={props.busy === 'save-draft'}>Save Draft</button>
                <button type="button" onClick={props.onQuickClose} disabled={props.busy === 'quick-close'}>Close Bill</button>
                <button type="button" className="ghost-button" onClick={props.onClearCart}>Clear</button>
            </div>
        </>
    );
}

function OrdersPage({invoices, busy, onOpenInvoice, onSendKOT, onCloseInvoice}: {
    invoices: InvoiceSummary[];
    busy: string;
    onOpenInvoice: (id: string) => void;
    onSendKOT: (id: string) => void;
    onCloseInvoice: (id: string) => void;
}) {
    return (
        <section className="page-grid two">
            <Panel title="Active Orders" action={`${invoices.length} open`}>
                <InvoiceStack
                    invoices={invoices}
                    empty="No active orders"
                    onOpen={onOpenInvoice}
                    renderActions={(invoice) => (
                        <>
                            {invoice.status === 'draft' && <button type="button" disabled={busy === `kot-${invoice.id}`} onClick={() => onSendKOT(invoice.id)}>Send KOT</button>}
                            <button type="button" disabled={busy === `close-${invoice.id}`} onClick={() => onCloseInvoice(invoice.id)}>Close</button>
                        </>
                    )}
                />
            </Panel>
            <Panel title="Order Controls" action="Lifecycle">
                <div className="control-list">
                    <LifecycleRow title="Save Draft" detail="No KOT, no stock movement" tone="neutral"/>
                    <LifecycleRow title="Send KOT" detail="Kitchen tickets and recipe deduction once" tone="live"/>
                    <LifecycleRow title="Close Bill" detail="Payment and receipt queue" tone="closed"/>
                    <LifecycleRow title="Void / Refund" detail="Manager PIN audit path" tone="risk"/>
                </div>
            </Panel>
        </section>
    );
}

function InvoicesPage(props: {
    invoices: InvoiceSummary[];
    selectedInvoice: InvoiceDetail | null;
    statusFilter: string;
    search: string;
    paymentFilter: string;
    managerPIN: string;
    actionReason: string;
    refundAmount: number;
    splitMode: 'items' | 'amount';
    splitLineQty: Record<string, number>;
    splitAmountA: number;
    splitAmountB: number;
    busy: string;
    onStatusFilter: (value: string) => void;
    onSearch: (value: string) => void;
    onPaymentFilter: (value: string) => void;
    onOpenInvoice: (id: string) => void;
    onManagerPIN: (value: string) => void;
    onActionReason: (value: string) => void;
    onRefundAmount: (value: number) => void;
    onValidatePIN: () => void;
    onVoid: () => void;
    onRefund: () => void;
    onSplitMode: (value: 'items' | 'amount') => void;
    onSplitLineQty: (lineId: string, quantity: number) => void;
    onSplitAmountA: (value: number) => void;
    onSplitAmountB: (value: number) => void;
    onSplit: () => void;
    onExportPDF: () => void;
    onPrintReceipt: () => void;
    onExportInvoices: () => void;
}) {
    return (
        <section className="invoice-layout">
            <Panel title="Invoices" action={`${props.invoices.length} records`}>
                <div className="filter-row">
                    <select value={props.statusFilter} onChange={(event) => props.onStatusFilter(event.target.value)}>
                        <option value="all">All status</option>
                        <option value="draft">Draft</option>
                        <option value="kot_sent">KOT sent</option>
                        <option value="closed">Closed</option>
                        <option value="voided">Voided</option>
                        <option value="refunded">Refunded</option>
                        <option value="partially_refunded">Partial refund</option>
                        <option value="split">Split</option>
                    </select>
                    <select value={props.paymentFilter} onChange={(event) => props.onPaymentFilter(event.target.value)}>
                        <option value="all">All payments</option>
                        <option value="cash">Cash</option>
                        <option value="upi">UPI</option>
                        <option value="card">Card</option>
                    </select>
                    <input value={props.search} onChange={(event) => props.onSearch(event.target.value)} placeholder="Search"/>
                </div>
                <div className="button-row">
                    <button type="button" className="ghost-button" disabled={props.busy === 'invoices-csv'} onClick={props.onExportInvoices}>Download CSV</button>
                </div>
                <InvoiceStack invoices={props.invoices} empty="No invoices found" onOpen={props.onOpenInvoice}/>
            </Panel>

            <InvoiceDetailPanel {...props}/>
        </section>
    );
}

function InvoiceDetailPanel(props: {
    selectedInvoice: InvoiceDetail | null;
    managerPIN: string;
    actionReason: string;
    refundAmount: number;
    splitMode: 'items' | 'amount';
    splitLineQty: Record<string, number>;
    splitAmountA: number;
    splitAmountB: number;
    busy: string;
    onManagerPIN: (value: string) => void;
    onActionReason: (value: string) => void;
    onRefundAmount: (value: number) => void;
    onValidatePIN: () => void;
    onVoid: () => void;
    onRefund: () => void;
    onSplitMode: (value: 'items' | 'amount') => void;
    onSplitLineQty: (lineId: string, quantity: number) => void;
    onSplitAmountA: (value: number) => void;
    onSplitAmountB: (value: number) => void;
    onSplit: () => void;
    onExportPDF: () => void;
    onPrintReceipt: () => void;
}) {
    const detail = props.selectedInvoice;
    if (!detail) {
        return (
            <Panel title="Invoice Detail" action="Select">
                <p className="empty-copy">No invoice selected</p>
            </Panel>
        );
    }
    const summary = detail.summary;
    const canVoid = summary.status === 'draft';
    const canRefund = summary.status === 'closed' || summary.status === 'partially_refunded';
    const canSplit = !['voided', 'refunded', 'split'].includes(summary.status);

    return (
        <Panel title={summary.invoiceNumber || 'Invoice'} action={<StatusBadge status={summary.status}/>}>
            <div className="detail-head">
                <div>
                    <span>{summary.customerName || 'Walk-in Guest'}</span>
                    <strong>{currency.format(summary.total)}</strong>
                </div>
                <div>
                    <span>{summary.tableName || summary.orderType}</span>
                    <strong>{summary.paymentMethod || 'unpaid'}</strong>
                </div>
            </div>

            <section className="receipt-card">
                <div className="receipt-top">
                    <strong>{summary.invoiceNumber}</strong>
                    <span>{formatDate(summary.createdAt)}</span>
                </div>
                {detail.lines.length === 0 && <p className="empty-copy">Amount split invoice</p>}
                {detail.lines.map((line) => (
                    <div className="receipt-line" key={line.id}>
                        <span>{line.quantity} x {line.itemName}</span>
                        <strong>{currency.format(line.lineTotal)}</strong>
                    </div>
                ))}
                <MoneyRow label="Subtotal" value={summary.subtotal}/>
                <MoneyRow label="Discount" value={summary.discountTotal}/>
                <MoneyRow label="Tax" value={summary.taxTotal}/>
                <MoneyRow label="Payable" value={summary.total} strong/>
            </section>

            <div className="button-row">
                <button type="button" disabled={props.busy === 'invoice-pdf'} onClick={props.onExportPDF}>Save PDF</button>
                <button type="button" className="ghost-button" onClick={props.onPrintReceipt}>Print Receipt</button>
            </div>

            <div className="detail-columns">
                <MiniSection title="Payments">
                    {detail.payments.length === 0 && <p className="empty-copy">No payment captured</p>}
                    {detail.payments.map((payment) => (
                        <DataRow key={payment.id} label={`${payment.method} / ${payment.status}`} value={currency.format(payment.amount)}/>
                    ))}
                </MiniSection>
                <MiniSection title="KOTs">
                    {detail.kitchenTickets.length === 0 && <p className="empty-copy">No KOT</p>}
                    {detail.kitchenTickets.map((ticket) => (
                        <DataRow key={ticket.id} label={ticket.ticketNumber} value={ticket.routeName}/>
                    ))}
                </MiniSection>
                <MiniSection title="Events">
                    {detail.events.map((event) => (
                        <DataRow key={event.id} label={event.title} value={formatTime(event.createdAt)}/>
                    ))}
                </MiniSection>
                <MiniSection title="Queue">
                    {detail.notifications.length === 0 && <p className="empty-copy">No message queued</p>}
                    {detail.notifications.map((notification) => (
                        <DataRow key={notification.id} label={notification.template} value={notification.status}/>
                    ))}
                </MiniSection>
            </div>

            <div className="approval-box">
                <input type="password" value={props.managerPIN} onChange={(event) => props.onManagerPIN(event.target.value)} placeholder="Manager PIN"/>
                <input value={props.actionReason} onChange={(event) => props.onActionReason(event.target.value)} placeholder="Reason"/>
                <button type="button" className="ghost-button" onClick={props.onValidatePIN}>Validate</button>
            </div>

            <div className="button-row">
                <button type="button" disabled={!canVoid || props.busy === 'void-invoice'} onClick={props.onVoid}>Void</button>
                <button type="button" disabled={!canRefund || props.busy === 'refund-invoice'} onClick={props.onRefund}>Refund</button>
                <input className="amount-input" type="number" min="0" value={props.refundAmount} onChange={(event) => props.onRefundAmount(Number(event.target.value))}/>
            </div>

            <div className="split-box">
                <div className="segmented">
                    <button type="button" className={props.splitMode === 'items' ? 'active' : ''} onClick={() => props.onSplitMode('items')}>Item</button>
                    <button type="button" className={props.splitMode === 'amount' ? 'active' : ''} onClick={() => props.onSplitMode('amount')}>Amount</button>
                </div>
                {props.splitMode === 'items' ? (
                    <div className="split-lines">
                        {detail.lines.map((line) => (
                            <label key={line.id}>
                                <span>{line.itemName}</span>
                                <input
                                    type="number"
                                    min="0"
                                    max={line.quantity}
                                    value={props.splitLineQty[line.id] ?? 0}
                                    onChange={(event) => props.onSplitLineQty(line.id, Number(event.target.value))}
                                />
                            </label>
                        ))}
                    </div>
                ) : (
                    <div className="amount-split">
                        <input type="number" min="0" value={props.splitAmountA} onChange={(event) => props.onSplitAmountA(Number(event.target.value))}/>
                        <input type="number" min="0" value={props.splitAmountB} onChange={(event) => props.onSplitAmountB(Number(event.target.value))}/>
                    </div>
                )}
                <button type="button" disabled={!canSplit || props.busy === 'split-items' || props.busy === 'split-amounts'} onClick={props.onSplit}>Split Bill</button>
            </div>
        </Panel>
    );
}

function KitchenPage({tickets, printJobs, onUpdateTicket, onQueuePrint}: {
    tickets: KitchenTicket[];
    printJobs: PrintJob[];
    onUpdateTicket: (ticketID: string, status: string) => void;
    onQueuePrint: (kind: string, referenceID: string) => void;
}) {
    const [station, setStation] = useState('all');
    const stations = ['all', ...Array.from(new Set(tickets.map((ticket) => ticket.routeName || 'Kitchen')))];
    const visibleTickets = tickets.filter((ticket) => station === 'all' || ticket.routeName === station);
    const activeTickets = visibleTickets.filter((ticket) => !['served', 'cancelled'].includes(ticket.status));
    const allDayCounts = new Map<string, number>();
    visibleTickets.forEach((ticket) => {
        ticket.lines.forEach((line) => {
            allDayCounts.set(line.itemName, (allDayCounts.get(line.itemName) ?? 0) + line.quantity);
        });
    });
    const productionCounts = Array.from(allDayCounts.entries())
        .sort((a, b) => b[1] - a[1])
        .slice(0, 8);
    return (
        <section className="kds-layout">
            <div className="kds-command">
                <div className="segmented station-tabs">
                    {stations.map((name) => (
                        <button type="button" className={station === name ? 'active' : ''} key={name} onClick={() => setStation(name)}>
                            {name === 'all' ? 'All Stations' : name}
                        </button>
                    ))}
                </div>
                <div className="kds-stat">
                    <span>Active</span>
                    <strong>{activeTickets.length}</strong>
                </div>
                <div className="kds-stat">
                    <span>Ready</span>
                    <strong>{visibleTickets.filter((ticket) => ticket.status === 'ready').length}</strong>
                </div>
                <div className="kds-stat">
                    <span>Failed prints</span>
                    <strong>{printJobs.filter((job) => job.status === 'failed').length}</strong>
                </div>
            </div>

            <section className="kds-board">
                {visibleTickets.length === 0 && <p className="empty-copy">No tickets for this station</p>}
                {visibleTickets.map((ticket) => {
                    const age = ticketAgeMinutes(ticket.createdAt);
                    const ageClass = age >= 20 ? 'age-hot' : age >= 10 ? 'age-warn' : 'age-fresh';
                    return (
                        <article className={`kds-ticket ${ticket.status} ${ageClass}`} key={ticket.id}>
                            <div className="kds-ticket-head">
                                <div>
                                    <strong>{ticket.ticketNumber}</strong>
                                    <span>{ticket.routeName} / {ticket.invoiceNumber}</span>
                                </div>
                                <div>
                                    <StatusBadge status={ticket.status}/>
                                    <em>{Math.round(age)}m</em>
                                </div>
                            </div>
                            <div className="kds-lines">
                                {ticket.lines.map((line) => (
                                    <div key={line.id}>
                                        <strong>{line.quantity} x {line.itemName}</strong>
                                        {line.notes && <span>{line.notes}</span>}
                                    </div>
                                ))}
                            </div>
                            <div className="kds-actions">
                                {ticket.status !== 'preparing' && ticket.status !== 'ready' && ticket.status !== 'served' && ticket.status !== 'cancelled' && (
                                    <button type="button" onClick={() => onUpdateTicket(ticket.id, 'preparing')}>Start</button>
                                )}
                                {ticket.status === 'preparing' && <button type="button" onClick={() => onUpdateTicket(ticket.id, 'ready')}>Ready</button>}
                                {ticket.status === 'ready' && <button type="button" onClick={() => onUpdateTicket(ticket.id, 'served')}>Served</button>}
                                {ticket.status === 'served' && <button type="button" onClick={() => onUpdateTicket(ticket.id, 'preparing')}>Recall</button>}
                                {ticket.status !== 'served' && ticket.status !== 'cancelled' && (
                                    <button className="ghost-button" type="button" onClick={() => onUpdateTicket(ticket.id, 'cancelled')}>Cancel</button>
                                )}
                                {ticket.status === 'cancelled' && <button type="button" onClick={() => onUpdateTicket(ticket.id, 'queued')}>Refire</button>}
                                <button className="ghost-button" type="button" onClick={() => onQueuePrint('kot', ticket.saleId)}>Print</button>
                            </div>
                        </article>
                    );
                })}
            </section>

            <aside className="kds-side">
                <Panel title="All-Day Counts" action={`${productionCounts.length} items`}>
                    <div className="data-list">
                        {productionCounts.length === 0 && <p className="empty-copy">Counts appear as KOTs arrive</p>}
                        {productionCounts.map(([name, count]) => (
                            <DataRow key={name} label={name} value={String(count)}/>
                        ))}
                    </div>
                </Panel>
                <Panel title="Print Jobs" action={`${printJobs.length} queued`}>
                <div className="print-list">
                    {printJobs.length === 0 && <p className="empty-copy">No print jobs</p>}
                    {printJobs.slice(0, 8).map((job) => (
                        <article className="print-job" key={job.id}>
                            <div>
                                <strong>{job.kind} / {job.printerId}</strong>
                                <span>{job.target} / {job.status}</span>
                            </div>
                            <pre>{job.payload.slice(0, 220)}</pre>
                        </article>
                    ))}
                </div>
                </Panel>
            </aside>
        </section>
    );
}

function InventoryPage(props: {
    dashboard: Dashboard;
    ingredientDraft: IngredientDraft;
    ingredientCreateDraft: IngredientCreateDraft;
    activeIngredientId: string;
    busy: string;
    onActiveIngredient: (value: string) => void;
    onIngredientDraft: (value: IngredientDraft) => void;
    onSaveIngredient: () => void;
    onIngredientCreateDraft: (value: IngredientCreateDraft) => void;
    onSaveNewIngredient: () => void;
    onReconcileMilk: () => void;
    onReceiveBadBatch: () => void;
}) {
    return (
        <section className="page-grid two">
            <Panel title="Add Ingredient" action="Stock item">
                <div className="editor-stack">
                    <div className="form-grid">
                        <Field label="Name">
                            <input value={props.ingredientCreateDraft.name} onChange={(event) => props.onIngredientCreateDraft({...props.ingredientCreateDraft, name: event.target.value})}/>
                        </Field>
                        <Field label="Usage unit">
                            <input value={props.ingredientCreateDraft.unit} onChange={(event) => props.onIngredientCreateDraft({...props.ingredientCreateDraft, unit: event.target.value})}/>
                        </Field>
                        <Field label="Buy unit">
                            <input value={props.ingredientCreateDraft.purchaseUnit} onChange={(event) => props.onIngredientCreateDraft({...props.ingredientCreateDraft, purchaseUnit: event.target.value})}/>
                        </Field>
                        <Field label="Usage per buy">
                            <input type="number" min="0.01" value={props.ingredientCreateDraft.purchaseToUsage} onChange={(event) => props.onIngredientCreateDraft({...props.ingredientCreateDraft, purchaseToUsage: Number(event.target.value)})}/>
                        </Field>
                        <Field label="Opening stock">
                            <input type="number" min="0" value={props.ingredientCreateDraft.onHandQty} onChange={(event) => props.onIngredientCreateDraft({...props.ingredientCreateDraft, onHandQty: Number(event.target.value)})}/>
                        </Field>
                        <Field label="Reorder">
                            <input type="number" min="0" value={props.ingredientCreateDraft.reorderPoint} onChange={(event) => props.onIngredientCreateDraft({...props.ingredientCreateDraft, reorderPoint: Number(event.target.value)})}/>
                        </Field>
                        <Field label="Last cost">
                            <input type="number" min="0" value={props.ingredientCreateDraft.lastPurchaseCost} onChange={(event) => props.onIngredientCreateDraft({...props.ingredientCreateDraft, lastPurchaseCost: Number(event.target.value)})}/>
                        </Field>
                    </div>
                    <button type="button" disabled={props.busy === 'add-ingredient'} onClick={props.onSaveNewIngredient}>Add Ingredient</button>
                </div>
            </Panel>

            <Panel title="Deep Stock" action={`${props.dashboard.metrics.stockAlerts} alerts`}>
                <div className="stock-list">
                    {props.dashboard.ingredients.map((ingredient) => {
                        const level = Math.min(100, Math.max(4, (ingredient.onHandQty / Math.max(ingredient.reorderPoint * 3, 1)) * 100));
                        return (
                            <article className="stock-row" key={ingredient.id}>
                                <div>
                                    <strong>{ingredient.name}</strong>
                                    <span>{compactNumber.format(ingredient.onHandQty)} {ingredient.unit}</span>
                                </div>
                                <div className="stock-meter" aria-hidden="true">
                                    <i style={{width: `${level}%`}}/>
                                </div>
                                <small>{Math.round(ingredient.wasteFactor * 100)}% waste</small>
                            </article>
                        );
                    })}
                </div>
                <div className="button-row">
                    <button type="button" onClick={props.onReconcileMilk} disabled={props.busy === 'audit-milk'}>Audit Milk</button>
                    <button type="button" onClick={props.onReceiveBadBatch} disabled={props.busy === 'bad-batch'}>Bad Batch</button>
                </div>
            </Panel>

            <Panel title="Ingredient Units" action="Purchasing">
                <div className="editor-stack">
                    <select value={props.activeIngredientId} onChange={(event) => props.onActiveIngredient(event.target.value)}>
                        {props.dashboard.ingredients.map((ingredient) => (
                            <option value={ingredient.id} key={ingredient.id}>{ingredient.name}</option>
                        ))}
                    </select>
                    <div className="form-grid">
                        <Field label="Reorder">
                            <input type="number" min="0" value={props.ingredientDraft.reorderPoint} onChange={(event) => props.onIngredientDraft({...props.ingredientDraft, reorderPoint: Number(event.target.value)})}/>
                        </Field>
                        <Field label="Buy unit">
                            <input value={props.ingredientDraft.purchaseUnit} onChange={(event) => props.onIngredientDraft({...props.ingredientDraft, purchaseUnit: event.target.value})}/>
                        </Field>
                        <Field label="Usage per buy">
                            <input type="number" min="0.01" value={props.ingredientDraft.purchaseToUsage} onChange={(event) => props.onIngredientDraft({...props.ingredientDraft, purchaseToUsage: Number(event.target.value)})}/>
                        </Field>
                        <Field label="Last cost">
                            <input type="number" min="0" value={props.ingredientDraft.lastPurchaseCost} onChange={(event) => props.onIngredientDraft({...props.ingredientDraft, lastPurchaseCost: Number(event.target.value)})}/>
                        </Field>
                    </div>
                    <button type="button" onClick={props.onSaveIngredient} disabled={props.busy === 'save-ingredient'}>Save Ingredient</button>
                </div>
            </Panel>
        </section>
    );
}

function VendorsPage({pilot, ingredients, vendorDraft, poDraft, busy, onVendorDraft, onSaveVendor, onPODraft, onAddPOLine, onUpdatePOLine, onRemovePOLine, onCreatePO, onReceivePO}: {
    pilot: PilotWorkspace;
    ingredients: Ingredient[];
    vendorDraft: VendorDraft;
    poDraft: PurchaseOrderDraft;
    busy: string;
    onVendorDraft: (value: VendorDraft) => void;
    onSaveVendor: () => void;
    onPODraft: (value: PurchaseOrderDraft) => void;
    onAddPOLine: () => void;
    onUpdatePOLine: (index: number, patch: Partial<PurchaseOrderDraftLine>) => void;
    onRemovePOLine: (index: number) => void;
    onCreatePO: () => void;
    onReceivePO: () => void;
}) {
    return (
        <section className="page-grid two">
            <Panel title="New Vendor" action="Supplier">
                <div className="editor-stack">
                    <div className="form-grid">
                        <input placeholder="Vendor name" value={vendorDraft.name} onChange={(event) => onVendorDraft({...vendorDraft, name: event.target.value})}/>
                        <input placeholder="Phone / WhatsApp" value={vendorDraft.phone} onChange={(event) => onVendorDraft({...vendorDraft, phone: event.target.value})}/>
                        <input placeholder="GSTIN" value={vendorDraft.gstin} onChange={(event) => onVendorDraft({...vendorDraft, gstin: event.target.value})}/>
                        <input placeholder="Payment terms" value={vendorDraft.paymentTerms} onChange={(event) => onVendorDraft({...vendorDraft, paymentTerms: event.target.value})}/>
                    </div>
                    <button type="button" disabled={busy === 'save-vendor'} onClick={onSaveVendor}>Save Vendor</button>
                </div>
            </Panel>

            <Panel title="Create Purchase Order" action={`${poDraft.lines.length} lines`}>
                <div className="editor-stack">
                    <div className="form-grid compact">
                        <select value={poDraft.vendorId} onChange={(event) => onPODraft({...poDraft, vendorId: event.target.value})}>
                            {pilot.vendors.map((vendor) => (
                                <option value={vendor.id} key={vendor.id}>{vendor.name}</option>
                            ))}
                        </select>
                        <input type="date" value={poDraft.expectedDate} onChange={(event) => onPODraft({...poDraft, expectedDate: event.target.value})}/>
                    </div>
                    <div className="po-line-editor">
                        {poDraft.lines.map((line, index) => (
                            <article className="po-draft-line" key={`${line.ingredientId}-${index}`}>
                                <select value={line.ingredientId} onChange={(event) => onUpdatePOLine(index, {ingredientId: event.target.value})}>
                                    {ingredients.map((ingredient) => (
                                        <option value={ingredient.id} key={ingredient.id}>{ingredient.name}</option>
                                    ))}
                                </select>
                                <input type="number" min="0" placeholder="Qty" value={line.orderedQty || ''} onChange={(event) => onUpdatePOLine(index, {orderedQty: Number(event.target.value)})}/>
                                <input type="number" min="0" placeholder="Unit cost" value={line.unitCost || ''} onChange={(event) => onUpdatePOLine(index, {unitCost: Number(event.target.value)})}/>
                                <button type="button" className="ghost-button" onClick={() => onRemovePOLine(index)}>Remove</button>
                            </article>
                        ))}
                    </div>
                    <div className="button-row">
                        <button type="button" className="ghost-button" onClick={onAddPOLine}>Add Line</button>
                        <button type="button" disabled={busy === 'create-po'} onClick={onCreatePO}>Draft PO</button>
                    </div>
                </div>
            </Panel>

            <Panel title="Vendors" action={`${pilot.vendors.length} active`}>
                <div className="data-list">
                    {pilot.vendors.map((vendor) => (
                        <article className="vendor-card" key={vendor.id}>
                            <div>
                                <strong>{vendor.name}</strong>
                                <span>{vendor.phone} / {vendor.paymentTerms}</span>
                            </div>
                            <em>{Math.round(vendor.qualityScore)}%</em>
                        </article>
                    ))}
                </div>
                <div className="button-row">
                    <button type="button" disabled={busy === 'receive-po'} onClick={onReceivePO}>Receive PO</button>
                </div>
            </Panel>

            <Panel title="Purchase Orders" action={`${pilot.purchaseOrders.length} recent`}>
                <div className="data-list">
                    {pilot.purchaseOrders.length === 0 && <p className="empty-copy">No purchase orders</p>}
                    {pilot.purchaseOrders.map((po) => (
                        <article className="po-card" key={po.id}>
                            <div className="po-head">
                                <div>
                                    <strong>{po.poNumber}</strong>
                                    <span>{po.vendorName} / {po.status}</span>
                                </div>
                                <em>{currency.format(po.subtotal)}</em>
                            </div>
                            {po.lines.map((line) => (
                                <DataRow
                                    key={line.id}
                                    label={`${line.ingredientName} ${compactNumber.format(line.acceptedQty || line.orderedQty)}`}
                                    value={line.rejectedQty ? `${compactNumber.format(line.rejectedQty)} rejected` : currency.format(line.orderedQty * line.unitCost)}
                                />
                            ))}
                        </article>
                    ))}
                </div>
            </Panel>

            <Panel title="Debit Notes" action={`${pilot.debitNotes.length} vendor`}>
                <div className="data-list">
                    {pilot.debitNotes.length === 0 && <p className="empty-copy">No debit notes</p>}
                    {pilot.debitNotes.map((note) => (
                        <article className="notification-row" key={note.id}>
                            <div>
                                <strong>{note.vendorName}</strong>
                                <span>{note.reason || note.status}</span>
                            </div>
                            <em>{currency.format(note.amount)}</em>
                        </article>
                    ))}
                </div>
            </Panel>
        </section>
    );
}

function RecipesPage(props: {
    dashboard: Dashboard;
    pilot: PilotWorkspace;
    activeRecipeId: string;
    recipeDraft: RecipeDraftLine[];
    busy: string;
    menuItemDraft: MenuItemDraft;
    categoryDraft: string;
    modifierDraft: ModifierDraft;
    modifierLinkDraft: ModifierLinkDraft;
    menuImportText: string;
    onActiveRecipe: (value: string) => void;
    onAddLine: () => void;
    onUpdateLine: (index: number, patch: Partial<RecipeDraftLine>) => void;
    onRemoveLine: (index: number) => void;
    onSaveRecipe: () => void;
    onMenuItemDraft: (value: MenuItemDraft | ((current: MenuItemDraft) => MenuItemDraft)) => void;
    onSaveMenuItem: () => void;
    onCategoryDraft: (value: string) => void;
    onSaveCategory: () => void;
    onModifierDraft: (value: ModifierDraft | ((current: ModifierDraft) => ModifierDraft)) => void;
    onSaveModifier: () => void;
    onModifierLinkDraft: (value: ModifierLinkDraft | ((current: ModifierLinkDraft) => ModifierLinkDraft)) => void;
    onLinkModifier: () => void;
    onMenuImportText: (value: string) => void;
    onImportMenu: () => void;
    onSaveMenuItemPrice: (item: MenuItem, price: number) => void;
    onSaveMenuItemPatch: (item: MenuItem, patch: Partial<MenuItem>) => void;
}) {
    const activeCategories = props.pilot.categories.filter((category) => category.status !== 'hidden');
    const itemModifierRows = props.pilot.itemModifiers.map((link) => ({
        link,
        item: props.dashboard.menuItems.find((item) => item.id === link.itemId),
        modifier: props.pilot.modifiers.find((modifier) => modifier.id === link.modifierId),
    })).filter((row) => row.item && row.modifier);

    return (
        <section className="page-grid two">
            <Panel title="Menu Groups" action={`${props.pilot.categories.length} groups`}>
                <div className="editor-stack">
                    <div className="form-grid compact">
                        <input
                            placeholder="Coffee, Food, Dessert..."
                            value={props.categoryDraft}
                            onChange={(event) => props.onCategoryDraft(event.target.value)}
                        />
                        <button type="button" disabled={props.busy === 'menu-category'} onClick={props.onSaveCategory}>Save Group</button>
                    </div>
                    <div className="chip-list">
                        {props.pilot.categories.length === 0 && <p className="empty-copy">Create menu groups before adding many items.</p>}
                        {props.pilot.categories.map((category) => (
                            <button
                                type="button"
                                className="choice-chip"
                                key={category.id}
                                onClick={() => props.onMenuItemDraft((current) => ({...current, category: category.name}))}
                            >
                                {category.name}
                            </button>
                        ))}
                    </div>
                </div>
            </Panel>

            <Panel title="Add Menu Item" action="Live">
                <div className="editor-stack">
                    <div className="form-grid">
                        <input
                            placeholder="Item name"
                            value={props.menuItemDraft.name}
                            onChange={(event) => props.onMenuItemDraft((current) => ({...current, name: event.target.value}))}
                        />
                        <input
                            list="menu-category-options"
                            placeholder="Category"
                            value={props.menuItemDraft.category}
                            onChange={(event) => props.onMenuItemDraft((current) => ({...current, category: event.target.value}))}
                        />
                        <datalist id="menu-category-options">
                            {activeCategories.map((category) => (
                                <option value={category.name} key={category.id}/>
                            ))}
                        </datalist>
                        <input
                            type="number"
                            min="0"
                            placeholder="Price"
                            value={props.menuItemDraft.price || ''}
                            onChange={(event) => props.onMenuItemDraft((current) => ({...current, price: Number(event.target.value)}))}
                        />
                        <input
                            type="number"
                            min="0"
                            placeholder="Cost"
                            value={props.menuItemDraft.cost || ''}
                            onChange={(event) => props.onMenuItemDraft((current) => ({...current, cost: Number(event.target.value)}))}
                        />
                        <select
                            value={props.menuItemDraft.routeId}
                            onChange={(event) => props.onMenuItemDraft((current) => ({...current, routeId: event.target.value}))}
                        >
                            {props.dashboard.kitchenRoutes.map((route) => (
                                <option value={route.id} key={route.id}>{route.name}</option>
                            ))}
                        </select>
                        <input
                            type="number"
                            min="0"
                            placeholder="GST %"
                            value={props.menuItemDraft.taxPercent || ''}
                            onChange={(event) => props.onMenuItemDraft((current) => ({...current, taxPercent: Number(event.target.value)}))}
                        />
                    </div>
                    <button type="button" disabled={props.busy === 'menu-item-new'} onClick={props.onSaveMenuItem}>Add Item</button>
                </div>
            </Panel>

            <Panel title="Modifier Library" action={`${props.pilot.modifiers.length} add-ons`}>
                <div className="editor-stack">
                    <div className="form-grid">
                        <input
                            placeholder="Extra cheese, oat milk, no onion..."
                            value={props.modifierDraft.name}
                            onChange={(event) => props.onModifierDraft((current) => ({...current, name: event.target.value}))}
                        />
                        <input
                            type="number"
                            placeholder="Price delta"
                            value={props.modifierDraft.priceDelta || ''}
                            onChange={(event) => props.onModifierDraft((current) => ({...current, priceDelta: Number(event.target.value)}))}
                        />
                        <select
                            value={props.modifierDraft.routeId}
                            onChange={(event) => props.onModifierDraft((current) => ({...current, routeId: event.target.value}))}
                        >
                            {props.dashboard.kitchenRoutes.map((route) => (
                                <option value={route.id} key={route.id}>{route.name}</option>
                            ))}
                        </select>
                        <select
                            value={props.modifierDraft.status}
                            onChange={(event) => props.onModifierDraft((current) => ({...current, status: event.target.value}))}
                        >
                            <option value="active">Active</option>
                            <option value="hidden">Hidden</option>
                        </select>
                    </div>
                    <button type="button" disabled={props.busy === 'menu-modifier'} onClick={props.onSaveModifier}>Save Modifier</button>
                    <div className="data-list">
                        {props.pilot.modifiers.map((modifier) => (
                            <DataRow
                                key={modifier.id}
                                label={`${modifier.name} / ${props.dashboard.kitchenRoutes.find((route) => route.id === modifier.routeId)?.name ?? 'No route'}`}
                                value={modifier.priceDelta ? currency.format(modifier.priceDelta) : 'Included'}
                            />
                        ))}
                    </div>
                </div>
            </Panel>

            <Panel title="Attach Modifiers" action={`${itemModifierRows.length} links`}>
                <div className="editor-stack">
                    <div className="form-grid compact">
                        <select
                            value={props.modifierLinkDraft.itemId}
                            onChange={(event) => props.onModifierLinkDraft((current) => ({...current, itemId: event.target.value}))}
                        >
                            {props.dashboard.menuItems.map((item) => (
                                <option value={item.id} key={item.id}>{item.name}</option>
                            ))}
                        </select>
                        <select
                            value={props.modifierLinkDraft.modifierId}
                            onChange={(event) => props.onModifierLinkDraft((current) => ({...current, modifierId: event.target.value}))}
                        >
                            {props.pilot.modifiers.map((modifier) => (
                                <option value={modifier.id} key={modifier.id}>{modifier.name}</option>
                            ))}
                        </select>
                    </div>
                    <button type="button" disabled={props.busy === 'modifier-link'} onClick={props.onLinkModifier}>Attach to Item</button>
                    <div className="data-list">
                        {itemModifierRows.length === 0 && <p className="empty-copy">Attach add-ons to items so servers see the right choices during ordering.</p>}
                        {itemModifierRows.map((row) => (
                            <DataRow
                                key={`${row.link.itemId}-${row.link.modifierId}`}
                                label={row.item?.name ?? ''}
                                value={row.modifier?.name ?? ''}
                            />
                        ))}
                    </div>
                </div>
            </Panel>

            <Panel title="Menu Upload" action="CSV">
                <div className="editor-stack">
                    <textarea
                        className="menu-import"
                        value={props.menuImportText}
                        onChange={(event) => props.onMenuImportText(event.target.value)}
                        spellCheck={false}
                    />
                    <button type="button" disabled={props.busy === 'menu-import'} onClick={props.onImportMenu}>Import Rows</button>
                </div>
            </Panel>

            <Panel title="Item Master" action={`${props.dashboard.menuItems.length} items`}>
                <div className="data-list">
                    {props.dashboard.menuItems.map((item) => (
                        <article className="item-master-row" key={item.id}>
                            <div className="item-master-title">
                                <strong>{item.name}</strong>
                                <span>{item.category} / {item.routeName}</span>
                            </div>
                            <div className="item-master-controls">
                                <select
                                    aria-label={`${item.name} category`}
                                    value={item.category}
                                    onChange={(event) => props.onSaveMenuItemPatch(item, {category: event.target.value})}
                                >
                                    {activeCategories.length === 0 && <option value={item.category}>{item.category}</option>}
                                    {activeCategories.map((category) => (
                                        <option value={category.name} key={category.id}>{category.name}</option>
                                    ))}
                                    {!activeCategories.some((category) => category.name === item.category) && <option value={item.category}>{item.category}</option>}
                                </select>
                                <select
                                    aria-label={`${item.name} kitchen route`}
                                    value={item.routeId}
                                    onChange={(event) => props.onSaveMenuItemPatch(item, {routeId: event.target.value})}
                                >
                                    {props.dashboard.kitchenRoutes.map((route) => (
                                        <option value={route.id} key={route.id}>{route.name}</option>
                                    ))}
                                </select>
                                <input
                                    aria-label={`${item.name} price`}
                                    type="number"
                                    min="0"
                                    defaultValue={item.price}
                                    onBlur={(event) => {
                                        const value = Number(event.target.value);
                                        if (Number.isFinite(value) && value !== item.price) props.onSaveMenuItemPrice(item, value);
                                    }}
                                />
                                <input
                                    aria-label={`${item.name} cost`}
                                    type="number"
                                    min="0"
                                    defaultValue={item.cost}
                                    onBlur={(event) => {
                                        const value = Number(event.target.value);
                                        if (Number.isFinite(value) && value !== item.cost) props.onSaveMenuItemPatch(item, {cost: value});
                                    }}
                                />
                                <input
                                    aria-label={`${item.name} GST percent`}
                                    type="number"
                                    min="0"
                                    defaultValue={Math.round(item.taxRate * 10000) / 100}
                                    onBlur={(event) => {
                                        const value = Number(event.target.value);
                                        if (Number.isFinite(value) && value / 100 !== item.taxRate) props.onSaveMenuItemPatch(item, {taxRate: value / 100});
                                    }}
                                />
                                <select
                                    aria-label={`${item.name} status`}
                                    value={item.status}
                                    onChange={(event) => props.onSaveMenuItemPatch(item, {status: event.target.value})}
                                >
                                    <option value="active">Active</option>
                                    <option value="out_of_stock">Out of stock</option>
                                    <option value="hidden">Hidden</option>
                                </select>
                            </div>
                        </article>
                    ))}
                </div>
            </Panel>

            <Panel title="Recipe Editor" action="BOM">
                <div className="editor-stack">
                    <select value={props.activeRecipeId} onChange={(event) => props.onActiveRecipe(event.target.value)}>
                        {props.dashboard.menuItems.map((item) => (
                            <option value={item.id} key={item.id}>{item.name}</option>
                        ))}
                    </select>
                    <div className="recipe-line-list">
                        {props.recipeDraft.map((line, index) => {
                            const ingredient = props.dashboard.ingredients.find((item) => item.id === line.ingredientId);
                            return (
                                <article className="recipe-line" key={`${line.ingredientId}-${index}`}>
                                    <select value={line.ingredientId} onChange={(event) => props.onUpdateLine(index, {ingredientId: event.target.value})}>
                                        {props.dashboard.ingredients.map((item) => (
                                            <option value={item.id} key={item.id}>{item.name}</option>
                                        ))}
                                    </select>
                                    <input type="number" min="0.01" value={line.quantity} onChange={(event) => props.onUpdateLine(index, {quantity: Number(event.target.value)})}/>
                                    <span>{ingredient?.unit ?? ''}</span>
                                    <button type="button" className="ghost-button" onClick={() => props.onRemoveLine(index)}>Remove</button>
                                </article>
                            );
                        })}
                    </div>
                    <div className="button-row">
                        <button type="button" onClick={props.onAddLine}>Add Line</button>
                        <button type="button" onClick={props.onSaveRecipe} disabled={props.busy === 'save-recipe'}>Save Recipe</button>
                    </div>
                </div>
            </Panel>

            <Panel title="Menu Economics" action="Margin">
                <div className="data-list">
                    {props.dashboard.menuItems.map((item) => (
                        <DataRow key={item.id} label={item.name} value={`${Math.round(((item.price - item.cost) / item.price) * 100)}%`}/>
                    ))}
                </div>
            </Panel>
        </section>
    );
}

function CustomersPage({customers, signals}: {customers: Customer[]; signals: Signal[]}) {
    return (
        <section className="page-grid two">
            <Panel title="Customers" action={`${customers.length} profiles`}>
                <div className="data-list">
                    {customers.map((customer) => (
                        <article className="customer-row" key={customer.id}>
                            <div>
                                <strong>{customer.name}</strong>
                                <span>{customer.favoriteItem || customer.phone}</span>
                            </div>
                            <em>{customer.visitCount} visits</em>
                        </article>
                    ))}
                </div>
            </Panel>
            <Panel title="Signals" action={`${signals.length} live`}>
                <div className="signal-list">
                    {signals.map((signal) => (
                        <article className={`signal-card priority-${signal.priority}`} key={signal.id}>
                            <span>{signal.kind}</span>
                            <strong>{signal.title}</strong>
                            <p>{signal.detail}</p>
                            <button type="button">{signal.action}</button>
                        </article>
                    ))}
                </div>
            </Panel>
        </section>
    );
}

function MarketingPage({drafts, notifications, busy, onMarkRead}: {
    drafts: MarketingDraft[];
    notifications: NotificationRecord[];
    busy: string;
    onMarkRead: (id: string) => void;
}) {
    return (
        <section className="page-grid two">
            <Panel title="Drafts" action={`${drafts.length} queued`}>
                <div className="draft-list">
                    {drafts.map((draft) => (
                        <article className="draft-card" key={draft.id}>
                            <span>{draft.channel}</span>
                            <strong>{draft.title}</strong>
                            <p>{draft.caption}</p>
                        </article>
                    ))}
                </div>
            </Panel>
            <Panel title="WhatsApp Queue" action={`${notifications.length} records`}>
                <div className="notification-list">
                    {notifications.length === 0 && <p className="empty-copy">No queued messages</p>}
                    {notifications.map((notification) => (
                        <article className="notification-row" key={notification.id}>
                            <div>
                                <strong>{notification.template}</strong>
                                <span>{notification.recipient || 'No phone'} / {notification.status}</span>
                            </div>
                            <button type="button" className="ghost-button" disabled={Boolean(notification.readAt) || busy === `notification-${notification.id}`} onClick={() => onMarkRead(notification.id)}>
                                {notification.readAt ? 'Read' : 'Mark'}
                            </button>
                        </article>
                    ))}
                </div>
            </Panel>
        </section>
    );
}

function IntegrationsPage(props: {
    pilot: PilotWorkspace;
    invoices: InvoiceSummary[];
    busy: string;
    printerTarget: string;
    onCreatePayment: (invoiceID: string) => void;
    onMarkPaymentPaid: (requestID: string) => void;
    onMarkPaymentFailed: (requestID: string) => void;
    onCancelPayment: (requestID: string) => void;
    onQueuePrint: (kind: string, referenceID: string) => void;
    onMarkPrintPrinted: (jobID: string) => void;
    onMarkPrintFailed: (jobID: string) => void;
    onRetryPrint: (jobID: string) => void;
    onStoreIntegration: (provider: string) => void;
    onPrinterTarget: (value: string) => void;
    onSavePrinter: () => void;
}) {
    const eligibleInvoices = props.invoices.filter((invoice) => ['draft', 'open', 'kot_sent'].includes(invoice.status)).slice(0, 5);
    return (
        <section className="page-grid two">
            <Panel title="Provider Health" action={`${props.pilot.integrations.length} providers`}>
                <div className="data-list">
                    {props.pilot.integrations.map((integration) => (
                        <article className="integration-card" key={integration.id}>
                            <div>
                                <strong>{integration.displayName}</strong>
                                <span>{integration.mode} / {integration.baseUrl}</span>
                            </div>
                            <StatusBadge status={integration.healthStatus}/>
                        </article>
                    ))}
                </div>
                <div className="button-row">
                    <button type="button" onClick={() => props.onStoreIntegration('meta_whatsapp')}>Meta Test</button>
                    <button type="button" onClick={() => props.onStoreIntegration('razorpay')}>Razorpay Test</button>
                    <button type="button" onClick={() => props.onStoreIntegration('escpos_printer')}>Printer Test</button>
                </div>
            </Panel>

            <Panel title="Razorpay" action={`${props.pilot.paymentRequests.length} requests`}>
                <div className="data-list">
                    {eligibleInvoices.map((invoice) => (
                        <article className="payment-row" key={invoice.id}>
                            <div>
                                <strong>{invoice.invoiceNumber}</strong>
                                <span>{invoice.customerName || 'Walk-in'} / {currency.format(invoice.total)}</span>
                            </div>
                            <button type="button" disabled={props.busy === `payment-${invoice.id}`} onClick={() => props.onCreatePayment(invoice.id)}>Create</button>
                        </article>
                    ))}
                    {props.pilot.paymentRequests.map((request) => (
                        <article className="payment-row" key={request.id}>
                            <div>
                                <strong>{request.reference}</strong>
                                <span>{humanStatus(request.status)} / {currency.format(request.amount)}</span>
                                <small>{request.checkoutUrl}</small>
                            </div>
                            <div className="compact-actions">
                                <button className="ghost-button" type="button" disabled={request.status === 'paid' || request.status === 'cancelled'} onClick={() => props.onMarkPaymentPaid(request.id)}>Paid</button>
                                <button className="ghost-button" type="button" disabled={request.status === 'paid' || request.status === 'failed' || request.status === 'cancelled'} onClick={() => props.onMarkPaymentFailed(request.id)}>Fail</button>
                                <button className="ghost-button" type="button" disabled={request.status === 'paid' || request.status === 'cancelled'} onClick={() => props.onCancelPayment(request.id)}>Cancel</button>
                            </div>
                        </article>
                    ))}
                </div>
            </Panel>

            <Panel title="Receipt Printing" action="TCP / USB / Bluetooth route">
                <div className="printer-config">
                    <select value={printerModeFromTarget(props.printerTarget)} onChange={(event) => props.onPrinterTarget(printerTargetForMode(event.target.value, props.printerTarget))}>
                        <option value="tcp">Network TCP</option>
                        <option value="system">System print dialog</option>
                        <option value="usb">USB adapter</option>
                        <option value="bluetooth">Bluetooth adapter</option>
                    </select>
                    <input value={props.printerTarget} onChange={(event) => props.onPrinterTarget(event.target.value)} placeholder="tcp://192.168.1.50:9100"/>
                    <button type="button" disabled={props.busy === 'save-printer'} onClick={props.onSavePrinter}>Save Printer</button>
                </div>
                <p className="helper-copy">Network printers use <strong>tcp://ip:9100</strong>. USB and Bluetooth save a local adapter path now, then the native bridge can claim it in production.</p>
                <InvoiceStack
                    invoices={props.invoices.slice(0, 5)}
                    empty="No invoices"
                    onOpen={(_id) => undefined}
                    renderActions={(invoice) => (
                        <button type="button" onClick={() => props.onQueuePrint('invoice', invoice.id)}>Receipt</button>
                    )}
                />
            </Panel>

            <Panel title="Print Queue" action={`${props.pilot.printJobs.length} jobs`}>
                <div className="data-list">
                    {props.pilot.printJobs.length === 0 && <p className="empty-copy">No print jobs yet</p>}
                    {props.pilot.printJobs.map((job) => (
                        <article className="print-job" key={job.id}>
                            <div className="print-job-head">
                                <div>
                                    <strong>{humanStatus(job.kind)} / {job.referenceId}</strong>
                                    <span>{job.target} / {job.attempts} attempts</span>
                                </div>
                                <StatusBadge status={job.status}/>
                            </div>
                            {job.lastError && <p className="error-copy">{job.lastError}</p>}
                            <pre>{job.payload}</pre>
                            <div className="button-row">
                                <button type="button" className="ghost-button" disabled={job.status === 'printed'} onClick={() => props.onMarkPrintPrinted(job.id)}>Printed</button>
                                <button type="button" className="ghost-button" disabled={job.status === 'printed'} onClick={() => props.onMarkPrintFailed(job.id)}>Failed</button>
                                <button type="button" className="ghost-button" disabled={job.status === 'printed'} onClick={() => props.onRetryPrint(job.id)}>Retry</button>
                            </div>
                        </article>
                    ))}
                </div>
            </Panel>
        </section>
    );
}

function DayClosePage({summary, cashCounted, closePIN, busy, onCashCounted, onClosePIN, onCloseDay, onExportPDF}: {
    summary: DayCloseSummary;
    cashCounted: number;
    closePIN: string;
    busy: string;
    onCashCounted: (value: number) => void;
    onClosePIN: (value: string) => void;
    onCloseDay: () => void;
    onExportPDF: () => void;
}) {
    return (
        <section className="page-grid two">
            <Panel title="Day Close" action={summary.businessDate}>
                <div className="dayclose-grid">
                    <Metric label="Sales" value={currency.format(summary.salesTotal)} tone="ink"/>
                    <Metric label="Cash" value={currency.format(summary.cashExpected)} tone="green"/>
                    <Metric label="Razorpay" value={currency.format(summary.razorpayTotal)} tone="gold"/>
                    <Metric label="Refunds" value={currency.format(summary.refundTotal)} tone="coral"/>
                </div>
                <div className="form-grid compact">
                    <Field label="Cash counted">
                        <input type="number" min="0" value={cashCounted} onChange={(event) => onCashCounted(Number(event.target.value))}/>
                    </Field>
                    <Field label="Variance">
                        <input readOnly value={currency.format(cashCounted - summary.cashExpected)}/>
                    </Field>
                    <Field label="Close PIN">
                        <input type="password" value={closePIN} onChange={(event) => onClosePIN(event.target.value)} placeholder="Manager PIN"/>
                    </Field>
                </div>
                <div className="button-row">
                    <button type="button" disabled={busy === 'close-day' || summary.status === 'closed'} onClick={onCloseDay}>
                        {summary.status === 'closed' ? 'Closed' : 'Close Day'}
                    </button>
                    <button type="button" className="ghost-button" disabled={busy === 'dayclose-pdf'} onClick={onExportPDF}>Save PDF</button>
                </div>
            </Panel>
            <Panel title="Tender Summary" action={summary.status}>
                <div className="settings-stack">
                    <DataRow label="UPI" value={currency.format(summary.upiTotal)}/>
                    <DataRow label="Card" value={currency.format(summary.cardTotal)}/>
                    <DataRow label="Discounts" value={currency.format(summary.discountTotal)}/>
                    <DataRow label="Voids" value={String(summary.voidCount)}/>
                </div>
            </Panel>
        </section>
    );
}

function AccountingPage({snapshot, busy, onRebuild, onExportCSV}: {
    snapshot: AccountingSnapshot;
    busy: string;
    onRebuild: () => void;
    onExportCSV: () => void;
}) {
    return (
        <section className="page-grid two">
            <Panel title="Accounting Shell" action={`${snapshot.voucherCount} vouchers`}>
                <div className="dayclose-grid">
                    <Metric label="Sales" value={currency.format(snapshot.salesTotal)} tone="ink"/>
                    <Metric label="GST" value={currency.format(snapshot.taxPayable)} tone="gold"/>
                    <Metric label="Cash/Bank" value={currency.format(snapshot.cashAndBank)} tone="green"/>
                    <Metric label="Refunds" value={currency.format(snapshot.refundTotal)} tone="coral"/>
                </div>
                <div className="button-row">
                    <button type="button" disabled={busy === 'rebuild-accounting'} onClick={onRebuild}>Rebuild</button>
                    <button type="button" className="ghost-button" disabled={busy === 'accounting-csv'} onClick={onExportCSV}>Download CSV</button>
                </div>
            </Panel>
            <Panel title="Trial Balance" action={`${snapshot.trialBalance.length} ledgers`}>
                <div className="data-list">
                    {snapshot.trialBalance.map((row) => (
                        <DataRow
                            key={row.ledgerId}
                            label={`${row.ledgerName} / ${row.groupName}`}
                            value={`${currency.format(row.balance)} ${row.balanceSide}`}
                        />
                    ))}
                </div>
            </Panel>
        </section>
    );
}

function AdminCommandCenter({analytics, dashboard, pilot, range, onRange, onRefresh, onNavigate}: {
    analytics: AdminAnalytics | null;
    dashboard: Dashboard;
    pilot: PilotWorkspace;
    range: string;
    onRange: (value: string) => void;
    onRefresh: () => void;
    onNavigate: (page: PageKey) => void;
}) {
    if (!analytics) {
        return (
            <section className="admin-command">
                <Panel title="Command Center" action="Loading">
                    <p className="empty-copy">Preparing business analytics</p>
                </Panel>
            </section>
        );
    }
    const stockRisk = analytics.inventoryHealth.filter((row) => row.status !== 'healthy').slice(0, 7);
    const failedJobs = pilot.printJobs.filter((job) => job.status === 'failed').length;
    return (
        <section className="admin-command">
            <div className="admin-hero">
                <div>
                    <span className="eyebrow">Administration / {analytics.rangeLabel}</span>
                    <h2>{dashboard.restaurant.name} operating picture</h2>
                    <p>{analytics.demoFallback ? 'Demo baseline is active until more closed-sale history is available.' : `Snapshot refreshed ${formatTime(analytics.snapshotStatus.updatedAt)} with ${analytics.snapshotStatus.dailyRows} day rows.`}</p>
                </div>
                <div className="admin-hero-actions">
                    <div className="segmented">
                        {['7d', '30d', '90d', 'all'].map((option) => (
                            <button type="button" className={range === option ? 'active' : ''} key={option} onClick={() => onRange(option)}>
                                {option === 'all' ? 'All' : option.toUpperCase()}
                            </button>
                        ))}
                    </div>
                    <button type="button" onClick={onRefresh}>Refresh Analytics</button>
                </div>
            </div>

            <section className="admin-metric-grid">
                {analytics.executive.map((metric) => (
                    <article className={`admin-metric ${metric.tone}`} key={metric.id}>
                        <span>{metric.label}</span>
                        <strong>{formatAnalyticsValue(metric.value, metric.format)}</strong>
                        <em>{metric.detail}</em>
                    </article>
                ))}
            </section>

            <section className="admin-grid">
                <Panel title="Sales Trend" action={`${analytics.salesTrend.length} points`}>
                    <LineChart points={analytics.salesTrend} format="currency"/>
                </Panel>
                <Panel title="Tender Mix" action="Settlement">
                    <BarList points={analytics.tenderMix} format="currency"/>
                </Panel>
                <Panel title="Category Mix" action="Menu demand">
                    <BarList points={analytics.categoryMix} format="currency"/>
                </Panel>
                <Panel title="Hourly Heatmap" action="Rush pattern">
                    <Heatmap points={analytics.hourlyHeatmap}/>
                </Panel>
                <Panel title="Item Velocity" action="Quantity">
                    <BarList points={analytics.itemVelocity} format="number"/>
                </Panel>
                <Panel title="Contribution Margin" action="Profit pool">
                    <BarList points={analytics.contributionMargin} format="currency"/>
                </Panel>
                <Panel title="Kitchen Aging" action={`${analytics.kitchenPerformance.length} stations`}>
                    <div className="station-health-list">
                        {analytics.kitchenPerformance.length === 0 && <p className="empty-copy">No active kitchen tickets</p>}
                        {analytics.kitchenPerformance.map((row) => (
                            <article className="station-health" key={row.routeName}>
                                <div>
                                    <strong>{row.routeName}</strong>
                                    <span>{row.openTickets} open / {row.readyTickets} ready / {row.servedTickets} served</span>
                                </div>
                                <em>{Math.round(row.oldestAgeMin)}m oldest</em>
                            </article>
                        ))}
                    </div>
                </Panel>
                <Panel title="Inventory Risk" action={`${stockRisk.length} watch`}>
                    <div className="inventory-risk-list">
                        {stockRisk.length === 0 && <p className="empty-copy">No stock items are under pressure</p>}
                        {stockRisk.map((row) => (
                            <article className={`inventory-risk ${row.status}`} key={row.id}>
                                <div>
                                    <strong>{row.name}</strong>
                                    <span>{compactNumber.format(row.onHandQty)} {row.unit} / reorder {compactNumber.format(row.reorderPoint)}</span>
                                </div>
                                <em>{Math.round(row.riskScore)}%</em>
                            </article>
                        ))}
                    </div>
                </Panel>
                <Panel title="Purchase And Rejection Trend" action="Vendors">
                    <BarList points={analytics.purchaseTrend} format="currency"/>
                </Panel>
                <Panel title="Settlement Health" action="Cash / tax / vendors">
                    <div className="settlement-grid">
                        <DataRow label="Cash expected" value={currency.format(analytics.settlement.cashExpected)}/>
                        <DataRow label="Cash variance" value={currency.format(analytics.settlement.cashVariance)}/>
                        <DataRow label="Razorpay clearing" value={currency.format(analytics.settlement.razorpayClearing)}/>
                        <DataRow label="GST payable" value={currency.format(analytics.settlement.taxPayable)}/>
                        <DataRow label="Vendor payables" value={currency.format(analytics.settlement.vendorPayables)}/>
                        <DataRow label="Print failures" value={String(failedJobs)}/>
                    </div>
                </Panel>
                <Panel title="Item Performance Matrix" action="Margin x velocity">
                    <MatrixQuadrants buckets={analytics.itemMatrix}/>
                </Panel>
                <Panel title="Focus Here" action={`${analytics.recommendations.length} moves`}>
                    <div className="recommendation-list">
                        {analytics.recommendations.map((item) => (
                            <article className="recommendation-card" key={item.id}>
                                <span>Priority {item.priority}</span>
                                <strong>{item.title}</strong>
                                <p>{item.detail}</p>
                                <button type="button" onClick={() => onNavigate(item.page || 'admin')}>Open {pageLabel(item.page)}</button>
                            </article>
                        ))}
                    </div>
                </Panel>
                <Panel title="Exception Board" action={`${analytics.exceptions.length} live`}>
                    <div className="exception-list">
                        {analytics.exceptions.length === 0 && <p className="empty-copy">No major exceptions</p>}
                        {analytics.exceptions.map((item) => (
                            <article className={`exception-card ${item.severity}`} key={item.id}>
                                <span>{item.kind}</span>
                                <strong>{item.title}</strong>
                                <p>{item.detail}</p>
                            </article>
                        ))}
                    </div>
                </Panel>
            </section>
        </section>
    );
}

function SettingsPage({
    metrics,
    notifications,
    pilot,
    syncStatus,
    settingsDraft,
    staffDraft,
    busy,
    readinessItems,
    onSettingsDraft,
    onSaveSettings,
    onStaffDraft,
    onSaveStaff,
    onUpdateStaffStatus,
    onExportBackup,
}: {
    metrics: Metrics;
    notifications: NotificationRecord[];
    pilot: PilotWorkspace;
    syncStatus: SyncStatus | null;
    settingsDraft: RestaurantSettings | null;
    staffDraft: StaffDraft;
    busy: string;
    readinessItems: ReadinessItem[];
    onSettingsDraft: (value: RestaurantSettings) => void;
    onSaveSettings: () => void;
    onStaffDraft: (value: StaffDraft | ((current: StaffDraft) => StaffDraft)) => void;
    onSaveStaff: () => void;
    onUpdateStaffStatus: (member: StaffMember, status: string) => void;
    onExportBackup: () => void;
}) {
    const activeStaff = pilot.staff.filter((member) => member.status === 'active');
    return (
        <section className="page-grid two">
            <Panel title="Restaurant Setup" action="Primary">
                {settingsDraft && (
                    <div className="editor-stack">
                        <div className="form-grid">
                            <Field label="Restaurant">
                                <input value={settingsDraft.restaurantName} onChange={(event) => onSettingsDraft({...settingsDraft, restaurantName: event.target.value})}/>
                            </Field>
                            <Field label="GSTIN">
                                <input value={settingsDraft.gstin} onChange={(event) => onSettingsDraft({...settingsDraft, gstin: event.target.value})}/>
                            </Field>
                            <Field label="State">
                                <input value={settingsDraft.state} onChange={(event) => onSettingsDraft({...settingsDraft, state: event.target.value})}/>
                            </Field>
                            <Field label="Invoice prefix">
                                <input value={settingsDraft.invoicePrefix} onChange={(event) => onSettingsDraft({...settingsDraft, invoicePrefix: event.target.value})}/>
                            </Field>
                            <Field label="GST rate">
                                <input type="number" step="0.01" min="0" value={settingsDraft.defaultTaxRate} onChange={(event) => onSettingsDraft({...settingsDraft, defaultTaxRate: Number(event.target.value)})}/>
                            </Field>
                            <Field label="Service">
                                <input type="number" step="0.01" min="0" value={settingsDraft.serviceChargeRate} onChange={(event) => onSettingsDraft({...settingsDraft, serviceChargeRate: Number(event.target.value)})}/>
                            </Field>
                            <Field label="Backup path">
                                <input value={settingsDraft.backupPath} onChange={(event) => onSettingsDraft({...settingsDraft, backupPath: event.target.value})}/>
                            </Field>
                        </div>
                        <button type="button" disabled={busy === 'save-settings'} onClick={onSaveSettings}>Save Setup</button>
                    </div>
                )}
            </Panel>
            <Panel title="Manager" action="Local PIN">
                <div className="settings-stack">
                    <DataRow label="Protected actions" value="Void, refund, line void, day close"/>
                    <DataRow label="Staff" value={String(pilot.staff.length)}/>
                    <DataRow label="Audit records" value={String(pilot.auditLog.length)}/>
                </div>
            </Panel>
            <Panel title="Launch Checklist" action={`${readinessItems.filter((item) => item.done).length}/${readinessItems.length} ready`}>
                <div className="data-list">
                    {readinessItems.map((item) => (
                        <article className="readiness-row" key={item.id}>
                            <div>
                                <strong>{item.title}</strong>
                                <span>{item.detail}</span>
                            </div>
                            <StatusBadge status={item.done ? 'ready' : 'needs_check'}/>
                        </article>
                    ))}
                </div>
            </Panel>
            <Panel title="Staff Directory" action={`${activeStaff.length} active`}>
                <div className="editor-stack">
                    <div className="form-grid">
                        <input
                            placeholder="Name"
                            value={staffDraft.name}
                            onChange={(event) => onStaffDraft((current) => ({...current, name: event.target.value}))}
                        />
                        <select value={staffDraft.role} onChange={(event) => onStaffDraft((current) => ({...current, role: event.target.value}))}>
                            <option value="waiter">Waiter</option>
                            <option value="runner">Runner</option>
                            <option value="cashier">Cashier</option>
                            <option value="barista">Barista</option>
                            <option value="kitchen">Kitchen</option>
                            <option value="manager">Manager</option>
                            <option value="owner">Owner</option>
                        </select>
                        <input
                            placeholder="PIN"
                            type="password"
                            value={staffDraft.pin}
                            onChange={(event) => onStaffDraft((current) => ({...current, pin: event.target.value}))}
                        />
                        <select value={staffDraft.status} onChange={(event) => onStaffDraft((current) => ({...current, status: event.target.value}))}>
                            <option value="active">Active</option>
                            <option value="inactive">Inactive</option>
                        </select>
                    </div>
                    <button type="button" disabled={busy === 'save-staff'} onClick={onSaveStaff}>Save Staff</button>
                    <div className="data-list">
                        {pilot.staff.map((member) => (
                            <article className="staff-row" key={member.id}>
                                <div>
                                    <strong>{member.name}</strong>
                                    <span>{humanStatus(member.role)} / {humanStatus(member.status)}{member.pinHash ? ' / PIN set' : ''}</span>
                                </div>
                                <button
                                    type="button"
                                    className="ghost-button"
                                    disabled={busy === `staff-${member.id}`}
                                    onClick={() => onUpdateStaffStatus(member, member.status === 'active' ? 'inactive' : 'active')}
                                >
                                    {member.status === 'active' ? 'Archive' : 'Restore'}
                                </button>
                            </article>
                        ))}
                    </div>
                </div>
            </Panel>
            <Panel title="Local Health" action="Offline first">
                <div className="settings-stack">
                    <DataRow label="Sync queue" value={String(syncStatus?.pendingCount ?? metrics.pendingSyncItems)}/>
                    <DataRow label="Sync failures" value={String(syncStatus?.failedCount ?? 0)}/>
                    <DataRow label="Database" value={formatBytes(syncStatus?.databaseBytes ?? 0)}/>
                    <DataRow label="Open KOTs" value={String(metrics.openKots)}/>
                    <DataRow label="Message queue" value={String(notifications.length)}/>
                    <DataRow label="Tables" value={String(pilot.tables.length)}/>
                </div>
                <div className="button-row">
                    <button type="button" disabled={busy === 'backup'} onClick={onExportBackup}>Backup Now</button>
                </div>
            </Panel>
            <Panel title="Security Audit" action={`${pilot.auditLog.length} recent`}>
                <div className="data-list">
                    {pilot.auditLog.length === 0 && <p className="empty-copy">No protected actions yet</p>}
                    {pilot.auditLog.map((entry) => (
                        <article className="audit-row" key={entry.id}>
                            <div>
                                <strong>{humanStatus(entry.eventType)}</strong>
                                <span>{entry.detail || `${entry.targetType} ${entry.targetId}`}</span>
                            </div>
                            <small>{formatTime(entry.createdAt)}</small>
                        </article>
                    ))}
                </div>
            </Panel>
        </section>
    );
}

function InvoiceStack({invoices, empty, onOpen, renderActions}: {
    invoices: InvoiceSummary[];
    empty: string;
    onOpen: (id: string) => void;
    renderActions?: (invoice: InvoiceSummary) => ReactNode;
}) {
    return (
        <div className="invoice-stack">
            {invoices.length === 0 && <p className="empty-copy">{empty}</p>}
            {invoices.map((invoice) => (
                <article className="invoice-row" key={invoice.id}>
                    <button type="button" className="invoice-main" onClick={() => onOpen(invoice.id)}>
                        <div>
                            <strong>{invoice.invoiceNumber || 'Draft'}</strong>
                            <span>{invoice.customerName || 'Walk-in'} / {invoice.tableName || invoice.orderType}</span>
                        </div>
                        <div>
                            <em>{currency.format(invoice.total)}</em>
                            <StatusBadge status={invoice.status}/>
                        </div>
                    </button>
                    {renderActions && <div className="invoice-actions">{renderActions(invoice)}</div>}
                </article>
            ))}
        </div>
    );
}

function LineChart({points, format}: {points: AnalyticsPoint[]; format: string}) {
    const safePoints = points.length > 0 ? points : [{label: 'No data', value: 0, count: 0, tone: 'muted'}];
    const max = Math.max(...safePoints.map((point) => point.value), 1);
    const width = 640;
    const height = 210;
    const step = safePoints.length > 1 ? width / (safePoints.length - 1) : width;
    const plot = safePoints.map((point, index) => {
        const x = safePoints.length > 1 ? index * step : width / 2;
        const y = height - ((point.value / max) * (height - 28)) - 12;
        return `${x},${y}`;
    }).join(' ');
    return (
        <div className="line-chart">
            <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Trend chart">
                <polyline points={plot} fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round"/>
                {safePoints.map((point, index) => {
                    const x = safePoints.length > 1 ? index * step : width / 2;
                    const y = height - ((point.value / max) * (height - 28)) - 12;
                    return <circle key={`${point.label}-${index}`} cx={x} cy={y} r="5"/>;
                })}
            </svg>
            <div className="chart-label-row">
                {safePoints.slice(0, 7).map((point, index) => (
                    <span key={`${point.label}-${index}`}>{point.label}<strong>{formatAnalyticsValue(point.value, format)}</strong></span>
                ))}
            </div>
        </div>
    );
}

function BarList({points, format}: {points: AnalyticsPoint[]; format: string}) {
    const max = Math.max(...points.map((point) => point.value), 1);
    return (
        <div className="bar-list">
            {points.length === 0 && <p className="empty-copy">No data in this range</p>}
            {points.map((point, index) => (
                <div className={`bar-row ${point.tone || ''}`} key={`${point.label}-${index}`}>
                    <span>{point.label}</span>
                    <i><b style={{width: `${Math.max(3, (point.value / max) * 100)}%`}}/></i>
                    <strong>{formatAnalyticsValue(point.value, format)}</strong>
                </div>
            ))}
        </div>
    );
}

function Heatmap({points}: {points: AnalyticsPoint[]}) {
    const max = Math.max(...points.map((point) => point.value), 1);
    return (
        <div className="heatmap">
            {points.length === 0 && <p className="empty-copy">No hourly pattern yet</p>}
            {points.map((point, index) => {
                const intensity = Math.max(0.14, point.value / max);
                return (
                    <div style={{'--heat': intensity} as CSSProperties} key={`${point.label}-${index}`}>
                        <strong>{point.label}</strong>
                        <span>{currency.format(point.value)}</span>
                        <em>{point.count} orders</em>
                    </div>
                );
            })}
        </div>
    );
}

function MatrixQuadrants({buckets}: {buckets: ItemMatrixBucket[]}) {
    return (
        <div className="matrix-grid">
            {buckets.map((bucket) => (
                <article className={`matrix-bucket ${bucket.id}`} key={bucket.id}>
                    <strong>{bucket.label}</strong>
                    <span>{bucket.description}</span>
                    <div>
                        {bucket.items.length === 0 && <p className="empty-copy">No items</p>}
                        {bucket.items.map((item) => (
                            <button type="button" key={item.itemId}>
                                <span>{item.name}</span>
                                <em>{item.quantity} sold / {Math.round(item.marginPct)}%</em>
                            </button>
                        ))}
                    </div>
                </article>
            ))}
        </div>
    );
}

function Metric({label, value, tone}: {label: string; value: string; tone: string}) {
    return (
        <article className={`metric-card ${tone}`}>
            <span>{label}</span>
            <strong>{value}</strong>
        </article>
    );
}

function Panel({title, action, children}: {title: string; action: ReactNode; children: ReactNode}) {
    return (
        <section className="panel">
            <div className="panel-heading">
                <h2>{title}</h2>
                <span>{action}</span>
            </div>
            {children}
        </section>
    );
}

function MiniSection({title, children}: {title: string; children: ReactNode}) {
    return (
        <section className="mini-section">
            <h3>{title}</h3>
            {children}
        </section>
    );
}

function Field({label, children}: {label: string; children: ReactNode}) {
    return (
        <label className="field">
            <span>{label}</span>
            {children}
        </label>
    );
}

function MoneyRow({label, value, strong = false}: {label: string; value: number; strong?: boolean}) {
    return (
        <div className={strong ? 'money-row strong' : 'money-row'}>
            <span>{label}</span>
            <strong>{currency.format(value)}</strong>
        </div>
    );
}

function DataRow({label, value}: {label: string; value: string}) {
    return (
        <div className="data-row">
            <span>{label}</span>
            <strong>{value}</strong>
        </div>
    );
}

function LifecycleRow({title, detail, tone}: {title: string; detail: string; tone: string}) {
    return (
        <article className={`lifecycle-row ${tone}`}>
            <strong>{title}</strong>
            <span>{detail}</span>
        </article>
    );
}

function StatusBadge({status}: {status: string}) {
    return <i className={`status-badge ${status}`}>{humanStatus(status)}</i>;
}

function ToastStack({toasts}: {toasts: Toast[]}) {
    return (
        <div className="toast-stack" aria-live="polite">
            {toasts.map((toast) => (
                <article className={`toast ${toast.kind}`} key={toast.id}>
                    <strong>{toast.title}</strong>
                    <span>{toast.detail}</span>
                </article>
            ))}
        </div>
    );
}

function calculateDiscount(subtotal: number, type: string, value: number) {
    if (subtotal <= 0 || value <= 0) return 0;
    if (type === 'percent') return Math.min(subtotal, subtotal * (value / 100));
    if (type === 'fixed') return Math.min(subtotal, value);
    return 0;
}

function normalizeDashboard(input: Dashboard): Dashboard {
    return {
        ...input,
        menuItems: input.menuItems ?? [],
        ingredients: input.ingredients ?? [],
        kitchenRoutes: input.kitchenRoutes ?? [],
        recipes: (input.recipes ?? []).map((recipe) => ({
            ...recipe,
            components: recipe.components ?? [],
        })),
        customers: input.customers ?? [],
        recentSales: input.recentSales ?? [],
        kitchenTickets: (input.kitchenTickets ?? []).map((ticket) => ({
            ...ticket,
            lines: ticket.lines ?? [],
        })),
        deliveries: (input.deliveries ?? []).map((delivery) => ({
            ...delivery,
            lines: delivery.lines ?? [],
        })),
        signals: input.signals ?? [],
        marketingDrafts: input.marketingDrafts ?? [],
    };
}

function normalizePilotWorkspace(input: PilotWorkspace): PilotWorkspace {
    return {
        ...input,
        integrations: input.integrations ?? [],
        staff: input.staff ?? [],
        categories: input.categories ?? [],
        modifiers: input.modifiers ?? [],
        itemModifiers: input.itemModifiers ?? [],
        floorSections: input.floorSections ?? [],
        tables: input.tables ?? [],
        orderSessions: (input.orderSessions ?? []).map((session) => ({
            ...session,
            lineCount: session.lineCount ?? 0,
            readyLineCount: session.readyLineCount ?? 0,
            preparingLineCount: session.preparingLineCount ?? 0,
            queuedLineCount: session.queuedLineCount ?? 0,
            notSentLineCount: session.notSentLineCount ?? 0,
        })),
        paymentRequests: input.paymentRequests ?? [],
        printJobs: input.printJobs ?? [],
        vendors: input.vendors ?? [],
        purchaseOrders: (input.purchaseOrders ?? []).map((order) => ({...order, lines: order.lines ?? []})),
        debitNotes: input.debitNotes ?? [],
        auditLog: input.auditLog ?? [],
        accounting: {
            ...input.accounting,
            trialBalance: input.accounting?.trialBalance ?? [],
        },
    };
}

function normalizeAdminAnalytics(input: AdminAnalytics): AdminAnalytics {
    return {
        ...input,
        executive: input.executive ?? [],
        salesTrend: input.salesTrend ?? [],
        hourlyHeatmap: input.hourlyHeatmap ?? [],
        tenderMix: input.tenderMix ?? [],
        categoryMix: input.categoryMix ?? [],
        itemVelocity: input.itemVelocity ?? [],
        contributionMargin: input.contributionMargin ?? [],
        inventoryHealth: input.inventoryHealth ?? [],
        kitchenPerformance: input.kitchenPerformance ?? [],
        purchaseTrend: input.purchaseTrend ?? [],
        itemMatrix: (input.itemMatrix ?? []).map((bucket) => ({...bucket, items: bucket.items ?? []})),
        settlement: input.settlement ?? {
            cashExpected: 0,
            cashVariance: 0,
            upiTotal: 0,
            cardTotal: 0,
            razorpayTotal: 0,
            razorpayClearing: 0,
            taxPayable: 0,
            vendorPayables: 0,
            refundTotal: 0,
        },
        exceptions: input.exceptions ?? [],
        recommendations: (input.recommendations ?? []).map((item) => ({
            ...item,
            page: pages.some((page) => page.key === item.page) ? item.page : 'admin',
        })),
        snapshotStatus: input.snapshotStatus ?? {dailyRows: 0, itemRows: 0, hourlyRows: 0, updatedAt: ''},
    };
}

function normalizeStaffSession(input: StaffSession): StaffSession {
    return {
        ...input,
        permissions: input.permissions ?? [],
        workspaceAccess: (input.workspaceAccess ?? []).filter((workspace): workspace is WorkspaceKey => workspaces.some((item) => item.key === workspace)),
    };
}

function normalizeSyncStatus(input: SyncStatus): SyncStatus {
    return {
        pendingCount: input?.pendingCount ?? 0,
        syncedCount: input?.syncedCount ?? 0,
        failedCount: input?.failedCount ?? 0,
        oldestPendingAt: input?.oldestPendingAt ?? '',
        lastSyncedAt: input?.lastSyncedAt ?? '',
        lastError: input?.lastError ?? '',
        databasePath: input?.databasePath ?? '',
        databaseBytes: input?.databaseBytes ?? 0,
        walBytes: input?.walBytes ?? 0,
        updatedAt: input?.updatedAt ?? '',
    };
}

function normalizeInvoiceDetail(input: InvoiceDetail): InvoiceDetail {
    return {
        ...input,
        lines: input.lines ?? [],
        payments: input.payments ?? [],
        refunds: input.refunds ?? [],
        events: input.events ?? [],
        kitchenTickets: (input.kitchenTickets ?? []).map((ticket) => ({...ticket, lines: ticket.lines ?? []})),
        notifications: input.notifications ?? [],
    };
}

function buildReadinessItems(dashboard: Dashboard, pilot: PilotWorkspace): ReadinessItem[] {
    const hasPrinter = pilot.integrations.some((integration) => integration.provider === 'escpos_printer' && integration.healthStatus !== 'missing' && integration.baseUrl);
    const hasWhatsApp = pilot.integrations.some((integration) => integration.provider === 'meta_whatsapp' && integration.credentialStatus === 'stored');
    const hasRazorpay = pilot.integrations.some((integration) => integration.provider === 'razorpay' && integration.credentialStatus === 'stored');
    const recipeCount = dashboard.recipes.filter((recipe) => recipe.components.length > 0).length;
    return [
        {
            id: 'restaurant',
            title: 'Restaurant profile',
            detail: 'Legal name, GST, invoice prefix, service charge, and business defaults.',
            done: Boolean(pilot.settings.restaurantName && pilot.settings.invoicePrefix && pilot.settings.defaultTaxRate >= 0),
            page: 'settings',
        },
        {
            id: 'staff',
            title: 'Staff directory',
            detail: 'Add waiters, cashiers, managers, and floor runners before assigning tables.',
            done: pilot.staff.some((member) => member.status === 'active'),
            page: 'settings',
        },
        {
            id: 'tables',
            title: 'Dining floor',
            detail: 'Create sections and tables so service starts from a clear floor map.',
            done: pilot.tables.length > 0 && pilot.floorSections.length > 0,
            page: 'floor',
        },
        {
            id: 'menu',
            title: 'Menu and modifiers',
            detail: 'Groups, sale items, add-ons, GST, route, cost, and item availability.',
            done: dashboard.menuItems.length >= 5 && pilot.categories.length > 0 && pilot.modifiers.length > 0,
            page: 'recipes',
        },
        {
            id: 'recipes',
            title: 'Recipe-linked stock',
            detail: 'Attach ingredients to best-selling items so inventory updates itself.',
            done: recipeCount >= Math.min(3, dashboard.menuItems.length),
            page: 'recipes',
        },
        {
            id: 'vendors',
            title: 'Vendors and purchases',
            detail: 'Set supplier records and purchase orders for receiving and bad batches.',
            done: pilot.vendors.length > 0 && pilot.purchaseOrders.length > 0,
            page: 'vendors',
        },
        {
            id: 'printer',
            title: 'Receipt and KOT printer',
            detail: 'Configure a network, system, USB, or Bluetooth print route.',
            done: Boolean(hasPrinter),
            page: 'integrations',
        },
        {
            id: 'payments',
            title: 'Razorpay test mode',
            detail: 'Store test credentials before trying payment links and QR collection.',
            done: Boolean(hasRazorpay),
            page: 'integrations',
        },
        {
            id: 'whatsapp',
            title: 'WhatsApp receipt queue',
            detail: 'Store Meta credentials and templates before live customer messages.',
            done: Boolean(hasWhatsApp),
            page: 'integrations',
        },
        {
            id: 'dayclose',
            title: 'Day close report',
            detail: 'Close cash, UPI, card, refunds, voids, discounts, and export a PDF.',
            done: pilot.dayClose.status === 'closed' || pilot.dayClose.salesTotal > 0,
            page: 'dayclose',
        },
    ];
}

function normalizeOrderSessionDetail(input: OrderSessionDetail): OrderSessionDetail {
    return {
        ...input,
        lines: (input.lines ?? []).map((line) => ({
            ...line,
            modifierIds: line.modifierIds ?? [],
            modifierNames: line.modifierNames ?? [],
        })),
        events: input.events ?? [],
    };
}

function humanStatus(status: string) {
    return status.replace(/_/g, ' ');
}

function workspaceAccessForRole(role: string): WorkspaceKey[] {
    switch (role) {
        case 'owner':
        case 'manager':
            return ['front', 'kitchenOps', 'backoffice', 'admin'];
        case 'kitchen':
        case 'barista':
            return ['kitchenOps'];
        case 'cashier':
        case 'waiter':
        case 'runner':
            return ['front'];
        default:
            return ['front'];
    }
}

function workspaceLabel(workspaceKey: WorkspaceKey) {
    return workspaces.find((workspace) => workspace.key === workspaceKey)?.label ?? workspaceKey;
}

function pageLabel(pageKey: PageKey) {
    return pages.find((page) => page.key === pageKey)?.label ?? 'Command Center';
}

function formatAnalyticsValue(value: number, format: string) {
    if (format === 'currency') return currency.format(value || 0);
    if (format === 'percent') return `${Math.round(value || 0)}%`;
    if (format === 'minutes') return `${Math.round(value || 0)}m`;
    return compactNumber.format(value || 0);
}

function formatBytes(value: number) {
    if (!value) return '0 B';
    if (value < 1024) return `${value} B`;
    if (value < 1024 * 1024) return `${Math.round(value / 1024)} KB`;
    return `${(value / (1024 * 1024)).toFixed(1)} MB`;
}

function ticketAgeMinutes(createdAt: string) {
    if (!createdAt) return 0;
    const created = new Date(createdAt).getTime();
    if (!Number.isFinite(created)) return 0;
    return Math.max(0, (Date.now() - created) / 60000);
}

function printerModeFromTarget(target: string) {
    if (target.startsWith('system://')) return 'system';
    if (target.startsWith('usb://')) return 'usb';
    if (target.startsWith('bluetooth://')) return 'bluetooth';
    return 'tcp';
}

function printerTargetForMode(mode: string, current: string) {
    if (mode === 'system') return 'system://default';
    if (mode === 'usb') return current.startsWith('usb://') ? current : 'usb://adapter/receipt';
    if (mode === 'bluetooth') return current.startsWith('bluetooth://') ? current : 'bluetooth://adapter/receipt';
    return current.startsWith('tcp://') ? current : 'tcp://192.168.1.50:9100';
}

function sessionKitchenStatus(session: OrderSession) {
    if (session.lineCount <= 0) return 'open';
    if (session.readyLineCount >= session.lineCount) return 'ready';
    if (session.readyLineCount > 0) return 'ready';
    if (session.preparingLineCount > 0) return 'preparing';
    if (session.queuedLineCount > 0) return 'queued';
    return 'open';
}

function sessionKitchenLabel(session: OrderSession) {
    if (session.lineCount <= 0) return 'No items added';
    if (session.readyLineCount >= session.lineCount) return 'All items ready';
    if (session.readyLineCount > 0) return `${session.readyLineCount}/${session.lineCount} ready for pickup`;
    if (session.preparingLineCount > 0) return `${session.preparingLineCount}/${session.lineCount} preparing`;
    if (session.queuedLineCount > 0) return `${session.queuedLineCount}/${session.lineCount} sent to kitchen`;
    return `${session.notSentLineCount || session.lineCount} not sent`;
}

function formatDate(value: string) {
    if (!value) return '';
    return new Intl.DateTimeFormat('en-IN', {dateStyle: 'medium', timeStyle: 'short'}).format(new Date(value));
}

function formatTime(value: string) {
    if (!value) return '';
    return new Intl.DateTimeFormat('en-IN', {timeStyle: 'short'}).format(new Date(value));
}

export default App;
