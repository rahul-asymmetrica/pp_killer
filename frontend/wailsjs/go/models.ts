export namespace nexus {

	export class AccountingBalanceRow {
	    ledgerId: string;
	    ledgerName: string;
	    groupName: string;
	    normalBalance: string;
	    debitTotal: number;
	    creditTotal: number;
	    balance: number;
	    balanceSide: string;

	    static createFrom(source: any = {}) {
	        return new AccountingBalanceRow(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ledgerId = source["ledgerId"];
	        this.ledgerName = source["ledgerName"];
	        this.groupName = source["groupName"];
	        this.normalBalance = source["normalBalance"];
	        this.debitTotal = source["debitTotal"];
	        this.creditTotal = source["creditTotal"];
	        this.balance = source["balance"];
	        this.balanceSide = source["balanceSide"];
	    }
	}
	export class AccountingSnapshot {
	    salesTotal: number;
	    taxPayable: number;
	    cashAndBank: number;
	    razorpayClearing: number;
	    vendorPayables: number;
	    refundTotal: number;
	    voucherCount: number;
	    trialBalance: AccountingBalanceRow[];

	    static createFrom(source: any = {}) {
	        return new AccountingSnapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.salesTotal = source["salesTotal"];
	        this.taxPayable = source["taxPayable"];
	        this.cashAndBank = source["cashAndBank"];
	        this.razorpayClearing = source["razorpayClearing"];
	        this.vendorPayables = source["vendorPayables"];
	        this.refundTotal = source["refundTotal"];
	        this.voucherCount = source["voucherCount"];
	        this.trialBalance = this.convertValues(source["trialBalance"], AccountingBalanceRow);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AddOrderSessionLineInput {
	    sessionId: string;
	    itemId: string;
	    quantity: number;
	    notes: string;
	    modifierIds: string[];

	    static createFrom(source: any = {}) {
	        return new AddOrderSessionLineInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.itemId = source["itemId"];
	        this.quantity = source["quantity"];
	        this.notes = source["notes"];
	        this.modifierIds = source["modifierIds"];
	    }
	}
	export class AnalyticsSnapshotStatus {
	    dailyRows: number;
	    itemRows: number;
	    hourlyRows: number;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new AnalyticsSnapshotStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dailyRows = source["dailyRows"];
	        this.itemRows = source["itemRows"];
	        this.hourlyRows = source["hourlyRows"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class AnalyticsRecommendation {
	    id: string;
	    title: string;
	    detail: string;
	    priority: number;
	    page: string;

	    static createFrom(source: any = {}) {
	        return new AnalyticsRecommendation(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.priority = source["priority"];
	        this.page = source["page"];
	    }
	}
	export class AnalyticsException {
	    id: string;
	    kind: string;
	    title: string;
	    detail: string;
	    severity: string;
	    value: number;

	    static createFrom(source: any = {}) {
	        return new AnalyticsException(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.severity = source["severity"];
	        this.value = source["value"];
	    }
	}
	export class SettlementHealth {
	    cashExpected: number;
	    cashVariance: number;
	    upiTotal: number;
	    cardTotal: number;
	    razorpayTotal: number;
	    razorpayClearing: number;
	    taxPayable: number;
	    vendorPayables: number;
	    refundTotal: number;

	    static createFrom(source: any = {}) {
	        return new SettlementHealth(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cashExpected = source["cashExpected"];
	        this.cashVariance = source["cashVariance"];
	        this.upiTotal = source["upiTotal"];
	        this.cardTotal = source["cardTotal"];
	        this.razorpayTotal = source["razorpayTotal"];
	        this.razorpayClearing = source["razorpayClearing"];
	        this.taxPayable = source["taxPayable"];
	        this.vendorPayables = source["vendorPayables"];
	        this.refundTotal = source["refundTotal"];
	    }
	}
	export class ItemMatrixEntry {
	    itemId: string;
	    name: string;
	    category: string;
	    quantity: number;
	    sales: number;
	    margin: number;
	    marginPct: number;

	    static createFrom(source: any = {}) {
	        return new ItemMatrixEntry(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.quantity = source["quantity"];
	        this.sales = source["sales"];
	        this.margin = source["margin"];
	        this.marginPct = source["marginPct"];
	    }
	}
	export class ItemMatrixBucket {
	    id: string;
	    label: string;
	    description: string;
	    items: ItemMatrixEntry[];

	    static createFrom(source: any = {}) {
	        return new ItemMatrixBucket(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.items = this.convertValues(source["items"], ItemMatrixEntry);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class KitchenPerformanceRow {
	    routeName: string;
	    openTickets: number;
	    readyTickets: number;
	    servedTickets: number;
	    averageAgeMin: number;
	    oldestAgeMin: number;

	    static createFrom(source: any = {}) {
	        return new KitchenPerformanceRow(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.routeName = source["routeName"];
	        this.openTickets = source["openTickets"];
	        this.readyTickets = source["readyTickets"];
	        this.servedTickets = source["servedTickets"];
	        this.averageAgeMin = source["averageAgeMin"];
	        this.oldestAgeMin = source["oldestAgeMin"];
	    }
	}
	export class InventoryHealthRow {
	    id: string;
	    name: string;
	    onHandQty: number;
	    reorderPoint: number;
	    unit: string;
	    riskScore: number;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new InventoryHealthRow(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.onHandQty = source["onHandQty"];
	        this.reorderPoint = source["reorderPoint"];
	        this.unit = source["unit"];
	        this.riskScore = source["riskScore"];
	        this.status = source["status"];
	    }
	}
	export class AnalyticsPoint {
	    label: string;
	    value: number;
	    count: number;
	    tone: string;

	    static createFrom(source: any = {}) {
	        return new AnalyticsPoint(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	        this.count = source["count"];
	        this.tone = source["tone"];
	    }
	}
	export class AnalyticsMetric {
	    id: string;
	    label: string;
	    value: number;
	    format: string;
	    detail: string;
	    tone: string;

	    static createFrom(source: any = {}) {
	        return new AnalyticsMetric(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.value = source["value"];
	        this.format = source["format"];
	        this.detail = source["detail"];
	        this.tone = source["tone"];
	    }
	}
	export class AdminAnalytics {
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

	    static createFrom(source: any = {}) {
	        return new AdminAnalytics(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generatedAt = source["generatedAt"];
	        this.rangeKey = source["rangeKey"];
	        this.rangeLabel = source["rangeLabel"];
	        this.demoFallback = source["demoFallback"];
	        this.executive = this.convertValues(source["executive"], AnalyticsMetric);
	        this.salesTrend = this.convertValues(source["salesTrend"], AnalyticsPoint);
	        this.hourlyHeatmap = this.convertValues(source["hourlyHeatmap"], AnalyticsPoint);
	        this.tenderMix = this.convertValues(source["tenderMix"], AnalyticsPoint);
	        this.categoryMix = this.convertValues(source["categoryMix"], AnalyticsPoint);
	        this.itemVelocity = this.convertValues(source["itemVelocity"], AnalyticsPoint);
	        this.contributionMargin = this.convertValues(source["contributionMargin"], AnalyticsPoint);
	        this.inventoryHealth = this.convertValues(source["inventoryHealth"], InventoryHealthRow);
	        this.kitchenPerformance = this.convertValues(source["kitchenPerformance"], KitchenPerformanceRow);
	        this.purchaseTrend = this.convertValues(source["purchaseTrend"], AnalyticsPoint);
	        this.itemMatrix = this.convertValues(source["itemMatrix"], ItemMatrixBucket);
	        this.settlement = this.convertValues(source["settlement"], SettlementHealth);
	        this.exceptions = this.convertValues(source["exceptions"], AnalyticsException);
	        this.recommendations = this.convertValues(source["recommendations"], AnalyticsRecommendation);
	        this.snapshotStatus = this.convertValues(source["snapshotStatus"], AnalyticsSnapshotStatus);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}





	export class ApprovalToken {
	    token: string;
	    approvedBy: string;
	    approvedAt: string;
	    expiresAt: string;

	    static createFrom(source: any = {}) {
	        return new ApprovalToken(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.approvedBy = source["approvedBy"];
	        this.approvedAt = source["approvedAt"];
	        this.expiresAt = source["expiresAt"];
	    }
	}
	export class AuditLogEntry {
	    id: string;
	    eventType: string;
	    targetType: string;
	    targetId: string;
	    detail: string;
	    actor: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new AuditLogEntry(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.eventType = source["eventType"];
	        this.targetType = source["targetType"];
	        this.targetId = source["targetId"];
	        this.detail = source["detail"];
	        this.actor = source["actor"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class BackupInput {
	    destination: string;

	    static createFrom(source: any = {}) {
	        return new BackupInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.destination = source["destination"];
	    }
	}
	export class CloseInvoiceInput {
	    invoiceId: string;
	    paymentMethod: string;
	    paymentTendered: number;

	    static createFrom(source: any = {}) {
	        return new CloseInvoiceInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoiceId = source["invoiceId"];
	        this.paymentMethod = source["paymentMethod"];
	        this.paymentTendered = source["paymentTendered"];
	    }
	}
	export class CloseOrderSessionInput {
	    sessionId: string;
	    customerName: string;
	    customerPhone: string;
	    paymentMethod: string;
	    paymentTendered: number;

	    static createFrom(source: any = {}) {
	        return new CloseOrderSessionInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.customerName = source["customerName"];
	        this.customerPhone = source["customerPhone"];
	        this.paymentMethod = source["paymentMethod"];
	        this.paymentTendered = source["paymentTendered"];
	    }
	}
	export class Customer {
	    id: string;
	    name: string;
	    phone: string;
	    totalSpend: number;
	    visitCount: number;
	    lastVisitAt: string;
	    favoriteItem: string;

	    static createFrom(source: any = {}) {
	        return new Customer(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.totalSpend = source["totalSpend"];
	        this.visitCount = source["visitCount"];
	        this.lastVisitAt = source["lastVisitAt"];
	        this.favoriteItem = source["favoriteItem"];
	    }
	}
	export class MarketingDraft {
	    id: string;
	    title: string;
	    channel: string;
	    caption: string;
	    status: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new MarketingDraft(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.channel = source["channel"];
	        this.caption = source["caption"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Signal {
	    id: string;
	    kind: string;
	    title: string;
	    detail: string;
	    action: string;
	    priority: number;

	    static createFrom(source: any = {}) {
	        return new Signal(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.action = source["action"];
	        this.priority = source["priority"];
	    }
	}
	export class DeliveryLine {
	    id: string;
	    batchId: string;
	    ingredientId: string;
	    ingredientName: string;
	    unit: string;
	    orderedQty: number;
	    acceptedQty: number;
	    rejectedQty: number;
	    unitCost: number;
	    rejectionReason: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new DeliveryLine(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.batchId = source["batchId"];
	        this.ingredientId = source["ingredientId"];
	        this.ingredientName = source["ingredientName"];
	        this.unit = source["unit"];
	        this.orderedQty = source["orderedQty"];
	        this.acceptedQty = source["acceptedQty"];
	        this.rejectedQty = source["rejectedQty"];
	        this.unitCost = source["unitCost"];
	        this.rejectionReason = source["rejectionReason"];
	        this.status = source["status"];
	    }
	}
	export class DeliveryBatch {
	    id: string;
	    vendorName: string;
	    invoiceNumber: string;
	    status: string;
	    createdAt: string;
	    lines: DeliveryLine[];

	    static createFrom(source: any = {}) {
	        return new DeliveryBatch(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.vendorName = source["vendorName"];
	        this.invoiceNumber = source["invoiceNumber"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.lines = this.convertValues(source["lines"], DeliveryLine);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class KitchenTicketLine {
	    id: string;
	    ticketId: string;
	    itemId: string;
	    itemName: string;
	    quantity: number;
	    notes: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new KitchenTicketLine(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ticketId = source["ticketId"];
	        this.itemId = source["itemId"];
	        this.itemName = source["itemName"];
	        this.quantity = source["quantity"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	    }
	}
	export class KitchenTicket {
	    id: string;
	    saleId: string;
	    invoiceNumber: string;
	    ticketNumber: string;
	    routeId: string;
	    routeName: string;
	    status: string;
	    createdAt: string;
	    lines: KitchenTicketLine[];

	    static createFrom(source: any = {}) {
	        return new KitchenTicket(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.saleId = source["saleId"];
	        this.invoiceNumber = source["invoiceNumber"];
	        this.ticketNumber = source["ticketNumber"];
	        this.routeId = source["routeId"];
	        this.routeName = source["routeName"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	        this.lines = this.convertValues(source["lines"], KitchenTicketLine);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaleSummary {
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

	    static createFrom(source: any = {}) {
	        return new SaleSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceNumber = source["invoiceNumber"];
	        this.customerName = source["customerName"];
	        this.customerPhone = source["customerPhone"];
	        this.channel = source["channel"];
	        this.orderType = source["orderType"];
	        this.tableName = source["tableName"];
	        this.subtotal = source["subtotal"];
	        this.discountTotal = source["discountTotal"];
	        this.taxTotal = source["taxTotal"];
	        this.total = source["total"];
	        this.paymentMethod = source["paymentMethod"];
	        this.status = source["status"];
	        this.sourceInvoiceId = source["sourceInvoiceId"];
	        this.splitGroupId = source["splitGroupId"];
	        this.voidReason = source["voidReason"];
	        this.refundReason = source["refundReason"];
	        this.approvedBy = source["approvedBy"];
	        this.approvedAt = source["approvedAt"];
	        this.createdAt = source["createdAt"];
	        this.closedAt = source["closedAt"];
	        this.kotSentAt = source["kotSentAt"];
	    }
	}
	export class RecipeComponent {
	    itemId: string;
	    ingredientId: string;
	    ingredientName: string;
	    unit: string;
	    purchaseUnit: string;
	    purchaseToUsage: number;
	    quantity: number;
	    wasteFactorOverride?: number;

	    static createFrom(source: any = {}) {
	        return new RecipeComponent(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.ingredientId = source["ingredientId"];
	        this.ingredientName = source["ingredientName"];
	        this.unit = source["unit"];
	        this.purchaseUnit = source["purchaseUnit"];
	        this.purchaseToUsage = source["purchaseToUsage"];
	        this.quantity = source["quantity"];
	        this.wasteFactorOverride = source["wasteFactorOverride"];
	    }
	}
	export class RecipeCard {
	    itemId: string;
	    itemName: string;
	    routeName: string;
	    components: RecipeComponent[];

	    static createFrom(source: any = {}) {
	        return new RecipeCard(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.itemName = source["itemName"];
	        this.routeName = source["routeName"];
	        this.components = this.convertValues(source["components"], RecipeComponent);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class KitchenRoute {
	    id: string;
	    name: string;
	    printerName: string;
	    color: string;

	    static createFrom(source: any = {}) {
	        return new KitchenRoute(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.printerName = source["printerName"];
	        this.color = source["color"];
	    }
	}
	export class Ingredient {
	    id: string;
	    name: string;
	    unit: string;
	    purchaseUnit: string;
	    purchaseToUsage: number;
	    onHandQty: number;
	    reorderPoint: number;
	    wasteFactor: number;
	    lastPurchaseCost: number;
	    lastAuditAt: string;

	    static createFrom(source: any = {}) {
	        return new Ingredient(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.purchaseUnit = source["purchaseUnit"];
	        this.purchaseToUsage = source["purchaseToUsage"];
	        this.onHandQty = source["onHandQty"];
	        this.reorderPoint = source["reorderPoint"];
	        this.wasteFactor = source["wasteFactor"];
	        this.lastPurchaseCost = source["lastPurchaseCost"];
	        this.lastAuditAt = source["lastAuditAt"];
	    }
	}
	export class MenuItem {
	    id: string;
	    name: string;
	    category: string;
	    price: number;
	    cost: number;
	    status: string;
	    routeId: string;
	    routeName: string;
	    taxRate: number;

	    static createFrom(source: any = {}) {
	        return new MenuItem(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.price = source["price"];
	        this.cost = source["cost"];
	        this.status = source["status"];
	        this.routeId = source["routeId"];
	        this.routeName = source["routeName"];
	        this.taxRate = source["taxRate"];
	    }
	}
	export class Metrics {
	    ordersToday: number;
	    salesToday: number;
	    stockAlerts: number;
	    pendingSyncItems: number;
	    customerCount: number;
	    averageOrder: number;
	    openKots: number;

	    static createFrom(source: any = {}) {
	        return new Metrics(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ordersToday = source["ordersToday"];
	        this.salesToday = source["salesToday"];
	        this.stockAlerts = source["stockAlerts"];
	        this.pendingSyncItems = source["pendingSyncItems"];
	        this.customerCount = source["customerCount"];
	        this.averageOrder = source["averageOrder"];
	        this.openKots = source["openKots"];
	    }
	}
	export class Restaurant {
	    id: string;
	    name: string;
	    phone: string;
	    website: string;
	    brandVoice: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new Restaurant(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.website = source["website"];
	        this.brandVoice = source["brandVoice"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Dashboard {
	    restaurant: Restaurant;
	    metrics: Metrics;
	    menuItems: MenuItem[];
	    ingredients: Ingredient[];
	    kitchenRoutes: KitchenRoute[];
	    recipes: RecipeCard[];
	    customers: Customer[];
	    recentSales: SaleSummary[];
	    kitchenTickets: KitchenTicket[];
	    deliveries: DeliveryBatch[];
	    signals: Signal[];
	    marketingDrafts: MarketingDraft[];

	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.restaurant = this.convertValues(source["restaurant"], Restaurant);
	        this.metrics = this.convertValues(source["metrics"], Metrics);
	        this.menuItems = this.convertValues(source["menuItems"], MenuItem);
	        this.ingredients = this.convertValues(source["ingredients"], Ingredient);
	        this.kitchenRoutes = this.convertValues(source["kitchenRoutes"], KitchenRoute);
	        this.recipes = this.convertValues(source["recipes"], RecipeCard);
	        this.customers = this.convertValues(source["customers"], Customer);
	        this.recentSales = this.convertValues(source["recentSales"], SaleSummary);
	        this.kitchenTickets = this.convertValues(source["kitchenTickets"], KitchenTicket);
	        this.deliveries = this.convertValues(source["deliveries"], DeliveryBatch);
	        this.signals = this.convertValues(source["signals"], Signal);
	        this.marketingDrafts = this.convertValues(source["marketingDrafts"], MarketingDraft);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DayCloseInput {
	    businessDate: string;
	    cashCounted: number;
	    notes: string;
	    staffId: string;
	    pin: string;

	    static createFrom(source: any = {}) {
	        return new DayCloseInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.businessDate = source["businessDate"];
	        this.cashCounted = source["cashCounted"];
	        this.notes = source["notes"];
	        this.staffId = source["staffId"];
	        this.pin = source["pin"];
	    }
	}
	export class DayCloseSummary {
	    id: string;
	    businessDate: string;
	    closedAt: string;
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

	    static createFrom(source: any = {}) {
	        return new DayCloseSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.businessDate = source["businessDate"];
	        this.closedAt = source["closedAt"];
	        this.status = source["status"];
	        this.salesTotal = source["salesTotal"];
	        this.cashExpected = source["cashExpected"];
	        this.cashCounted = source["cashCounted"];
	        this.cashVariance = source["cashVariance"];
	        this.upiTotal = source["upiTotal"];
	        this.cardTotal = source["cardTotal"];
	        this.razorpayTotal = source["razorpayTotal"];
	        this.refundTotal = source["refundTotal"];
	        this.voidCount = source["voidCount"];
	        this.discountTotal = source["discountTotal"];
	        this.notes = source["notes"];
	    }
	}

	export class DeliveryLineInput {
	    ingredientId: string;
	    orderedQty: number;
	    acceptedQty: number;
	    rejectedQty: number;
	    unitCost: number;
	    rejectionReason: string;

	    static createFrom(source: any = {}) {
	        return new DeliveryLineInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ingredientId = source["ingredientId"];
	        this.orderedQty = source["orderedQty"];
	        this.acceptedQty = source["acceptedQty"];
	        this.rejectedQty = source["rejectedQty"];
	        this.unitCost = source["unitCost"];
	        this.rejectionReason = source["rejectionReason"];
	    }
	}
	export class DeliveryInput {
	    vendorName: string;
	    invoiceNumber: string;
	    lines: DeliveryLineInput[];

	    static createFrom(source: any = {}) {
	        return new DeliveryInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vendorName = source["vendorName"];
	        this.invoiceNumber = source["invoiceNumber"];
	        this.lines = this.convertValues(source["lines"], DeliveryLineInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}


	export class DemoSeedResult {
	    months: number;
	    businessDays: number;
	    tables: number;
	    waiters: number;
	    customers: number;
	    invoices: number;
	    kitchenTickets: number;
	    dayCloses: number;
	    purchaseOrders: number;
	    refunds: number;
	    voids: number;
	    salesTotal: number;
	    startedAt: string;
	    endedAt: string;

	    static createFrom(source: any = {}) {
	        return new DemoSeedResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.months = source["months"];
	        this.businessDays = source["businessDays"];
	        this.tables = source["tables"];
	        this.waiters = source["waiters"];
	        this.customers = source["customers"];
	        this.invoices = source["invoices"];
	        this.kitchenTickets = source["kitchenTickets"];
	        this.dayCloses = source["dayCloses"];
	        this.purchaseOrders = source["purchaseOrders"];
	        this.refunds = source["refunds"];
	        this.voids = source["voids"];
	        this.salesTotal = source["salesTotal"];
	        this.startedAt = source["startedAt"];
	        this.endedAt = source["endedAt"];
	    }
	}
	export class DiningTable {
	    id: string;
	    sectionId: string;
	    sectionName: string;
	    name: string;
	    seats: number;
	    status: string;
	    activeSessionId: string;

	    static createFrom(source: any = {}) {
	        return new DiningTable(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sectionId = source["sectionId"];
	        this.sectionName = source["sectionName"];
	        this.name = source["name"];
	        this.seats = source["seats"];
	        this.status = source["status"];
	        this.activeSessionId = source["activeSessionId"];
	    }
	}
	export class DiningTableInput {
	    id: string;
	    sectionId: string;
	    name: string;
	    seats: number;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new DiningTableInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sectionId = source["sectionId"];
	        this.name = source["name"];
	        this.seats = source["seats"];
	        this.status = source["status"];
	    }
	}
	export class ExportResult {
	    kind: string;
	    path: string;
	    fileName: string;
	    mimeType: string;
	    bytes: number;

	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.path = source["path"];
	        this.fileName = source["fileName"];
	        this.mimeType = source["mimeType"];
	        this.bytes = source["bytes"];
	    }
	}
	export class FloorSection {
	    id: string;
	    name: string;
	    sortOrder: number;

	    static createFrom(source: any = {}) {
	        return new FloorSection(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class FloorSectionInput {
	    id: string;
	    name: string;
	    sortOrder: number;

	    static createFrom(source: any = {}) {
	        return new FloorSectionInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sortOrder = source["sortOrder"];
	    }
	}

	export class IngredientInput {
	    id: string;
	    name: string;
	    unit: string;
	    purchaseUnit: string;
	    purchaseToUsage: number;
	    onHandQty: number;
	    reorderPoint: number;
	    wasteFactor: number;
	    lastPurchaseCost: number;

	    static createFrom(source: any = {}) {
	        return new IngredientInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.unit = source["unit"];
	        this.purchaseUnit = source["purchaseUnit"];
	        this.purchaseToUsage = source["purchaseToUsage"];
	        this.onHandQty = source["onHandQty"];
	        this.reorderPoint = source["reorderPoint"];
	        this.wasteFactor = source["wasteFactor"];
	        this.lastPurchaseCost = source["lastPurchaseCost"];
	    }
	}
	export class IngredientUpdateInput {
	    ingredientId: string;
	    reorderPoint: number;
	    purchaseUnit: string;
	    purchaseToUsage: number;
	    lastPurchaseCost: number;

	    static createFrom(source: any = {}) {
	        return new IngredientUpdateInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ingredientId = source["ingredientId"];
	        this.reorderPoint = source["reorderPoint"];
	        this.purchaseUnit = source["purchaseUnit"];
	        this.purchaseToUsage = source["purchaseToUsage"];
	        this.lastPurchaseCost = source["lastPurchaseCost"];
	    }
	}
	export class IntegrationSetting {
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

	    static createFrom(source: any = {}) {
	        return new IntegrationSetting(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.provider = source["provider"];
	        this.mode = source["mode"];
	        this.displayName = source["displayName"];
	        this.baseUrl = source["baseUrl"];
	        this.credentialStatus = source["credentialStatus"];
	        this.healthStatus = source["healthStatus"];
	        this.lastCheckedAt = source["lastCheckedAt"];
	        this.lastError = source["lastError"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class IntegrationSettingInput {
	    provider: string;
	    mode: string;
	    displayName: string;
	    baseUrl: string;
	    secret: string;
	    healthStatus: string;

	    static createFrom(source: any = {}) {
	        return new IntegrationSettingInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.mode = source["mode"];
	        this.displayName = source["displayName"];
	        this.baseUrl = source["baseUrl"];
	        this.secret = source["secret"];
	        this.healthStatus = source["healthStatus"];
	    }
	}

	export class NotificationRecord {
	    id: string;
	    invoiceId: string;
	    channel: string;
	    recipient: string;
	    template: string;
	    payload: string;
	    status: string;
	    error: string;
	    createdAt: string;
	    sentAt: string;
	    readAt: string;

	    static createFrom(source: any = {}) {
	        return new NotificationRecord(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceId = source["invoiceId"];
	        this.channel = source["channel"];
	        this.recipient = source["recipient"];
	        this.template = source["template"];
	        this.payload = source["payload"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.createdAt = source["createdAt"];
	        this.sentAt = source["sentAt"];
	        this.readAt = source["readAt"];
	    }
	}
	export class InvoiceEvent {
	    id: string;
	    invoiceId: string;
	    type: string;
	    title: string;
	    detail: string;
	    actor: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new InvoiceEvent(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceId = source["invoiceId"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	        this.actor = source["actor"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class RefundRecord {
	    id: string;
	    invoiceId: string;
	    amount: number;
	    reason: string;
	    approvedBy: string;
	    approvedAt: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new RefundRecord(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceId = source["invoiceId"];
	        this.amount = source["amount"];
	        this.reason = source["reason"];
	        this.approvedBy = source["approvedBy"];
	        this.approvedAt = source["approvedAt"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class PaymentRecord {
	    id: string;
	    invoiceId: string;
	    method: string;
	    amount: number;
	    tendered: number;
	    changeDue: number;
	    status: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new PaymentRecord(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceId = source["invoiceId"];
	        this.method = source["method"];
	        this.amount = source["amount"];
	        this.tendered = source["tendered"];
	        this.changeDue = source["changeDue"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class InvoiceLine {
	    id: string;
	    invoiceId: string;
	    itemId: string;
	    itemName: string;
	    quantity: number;
	    unitPrice: number;
	    lineTotal: number;
	    notes: string;
	    routeId: string;
	    routeName: string;

	    static createFrom(source: any = {}) {
	        return new InvoiceLine(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceId = source["invoiceId"];
	        this.itemId = source["itemId"];
	        this.itemName = source["itemName"];
	        this.quantity = source["quantity"];
	        this.unitPrice = source["unitPrice"];
	        this.lineTotal = source["lineTotal"];
	        this.notes = source["notes"];
	        this.routeId = source["routeId"];
	        this.routeName = source["routeName"];
	    }
	}
	export class InvoiceDetail {
	    summary: SaleSummary;
	    lines: InvoiceLine[];
	    payments: PaymentRecord[];
	    refunds: RefundRecord[];
	    events: InvoiceEvent[];
	    kitchenTickets: KitchenTicket[];
	    notifications: NotificationRecord[];

	    static createFrom(source: any = {}) {
	        return new InvoiceDetail(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = this.convertValues(source["summary"], SaleSummary);
	        this.lines = this.convertValues(source["lines"], InvoiceLine);
	        this.payments = this.convertValues(source["payments"], PaymentRecord);
	        this.refunds = this.convertValues(source["refunds"], RefundRecord);
	        this.events = this.convertValues(source["events"], InvoiceEvent);
	        this.kitchenTickets = this.convertValues(source["kitchenTickets"], KitchenTicket);
	        this.notifications = this.convertValues(source["notifications"], NotificationRecord);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class InvoiceFilter {
	    status: string;
	    search: string;
	    paymentMethod: string;
	    dateFrom: string;
	    dateTo: string;

	    static createFrom(source: any = {}) {
	        return new InvoiceFilter(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.search = source["search"];
	        this.paymentMethod = source["paymentMethod"];
	        this.dateFrom = source["dateFrom"];
	        this.dateTo = source["dateTo"];
	    }
	}








	export class MenuCategory {
	    id: string;
	    name: string;
	    sortOrder: number;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new MenuCategory(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sortOrder = source["sortOrder"];
	        this.status = source["status"];
	    }
	}
	export class MenuCategoryInput {
	    id: string;
	    name: string;
	    sortOrder: number;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new MenuCategoryInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sortOrder = source["sortOrder"];
	        this.status = source["status"];
	    }
	}
	export class MenuImportInput {
	    text: string;

	    static createFrom(source: any = {}) {
	        return new MenuImportInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	    }
	}
	export class MenuImportResult {
	    imported: number;
	    updated: number;
	    skipped: string[];

	    static createFrom(source: any = {}) {
	        return new MenuImportResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = source["imported"];
	        this.updated = source["updated"];
	        this.skipped = source["skipped"];
	    }
	}

	export class MenuItemInput {
	    id: string;
	    name: string;
	    category: string;
	    price: number;
	    cost: number;
	    status: string;
	    routeId: string;
	    taxRate: number;

	    static createFrom(source: any = {}) {
	        return new MenuItemInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.price = source["price"];
	        this.cost = source["cost"];
	        this.status = source["status"];
	        this.routeId = source["routeId"];
	        this.taxRate = source["taxRate"];
	    }
	}
	export class MenuItemModifier {
	    itemId: string;
	    modifierId: string;

	    static createFrom(source: any = {}) {
	        return new MenuItemModifier(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.modifierId = source["modifierId"];
	    }
	}
	export class MenuModifier {
	    id: string;
	    name: string;
	    priceDelta: number;
	    routeId: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new MenuModifier(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.priceDelta = source["priceDelta"];
	        this.routeId = source["routeId"];
	        this.status = source["status"];
	    }
	}
	export class MenuModifierInput {
	    id: string;
	    name: string;
	    priceDelta: number;
	    routeId: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new MenuModifierInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.priceDelta = source["priceDelta"];
	        this.routeId = source["routeId"];
	        this.status = source["status"];
	    }
	}


	export class OpenOrderSessionInput {
	    tableId: string;
	    waiterId: string;
	    guestCount: number;
	    serviceMode: string;

	    static createFrom(source: any = {}) {
	        return new OpenOrderSessionInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tableId = source["tableId"];
	        this.waiterId = source["waiterId"];
	        this.guestCount = source["guestCount"];
	        this.serviceMode = source["serviceMode"];
	    }
	}
	export class OrderSession {
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

	    static createFrom(source: any = {}) {
	        return new OrderSession(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.tableId = source["tableId"];
	        this.tableName = source["tableName"];
	        this.sectionName = source["sectionName"];
	        this.waiterId = source["waiterId"];
	        this.waiterName = source["waiterName"];
	        this.guestCount = source["guestCount"];
	        this.serviceMode = source["serviceMode"];
	        this.status = source["status"];
	        this.invoiceId = source["invoiceId"];
	        this.subtotal = source["subtotal"];
	        this.taxTotal = source["taxTotal"];
	        this.serviceCharge = source["serviceCharge"];
	        this.total = source["total"];
	        this.openedAt = source["openedAt"];
	        this.closedAt = source["closedAt"];
	        this.lineCount = source["lineCount"];
	        this.readyLineCount = source["readyLineCount"];
	        this.preparingLineCount = source["preparingLineCount"];
	        this.queuedLineCount = source["queuedLineCount"];
	        this.notSentLineCount = source["notSentLineCount"];
	    }
	}
	export class OrderSessionEvent {
	    id: string;
	    sessionId: string;
	    type: string;
	    detail: string;
	    actor: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new OrderSessionEvent(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.type = source["type"];
	        this.detail = source["detail"];
	        this.actor = source["actor"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class OrderSessionLine {
	    id: string;
	    sessionId: string;
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

	    static createFrom(source: any = {}) {
	        return new OrderSessionLine(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.itemId = source["itemId"];
	        this.itemName = source["itemName"];
	        this.quantity = source["quantity"];
	        this.unitPrice = source["unitPrice"];
	        this.modifierTotal = source["modifierTotal"];
	        this.lineTotal = source["lineTotal"];
	        this.notes = source["notes"];
	        this.status = source["status"];
	        this.kotStatus = source["kotStatus"];
	        this.modifierIds = source["modifierIds"];
	        this.modifierNames = source["modifierNames"];
	    }
	}
	export class OrderSessionDetail {
	    session: OrderSession;
	    lines: OrderSessionLine[];
	    events: OrderSessionEvent[];

	    static createFrom(source: any = {}) {
	        return new OrderSessionDetail(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session = this.convertValues(source["session"], OrderSession);
	        this.lines = this.convertValues(source["lines"], OrderSessionLine);
	        this.events = this.convertValues(source["events"], OrderSessionEvent);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}



	export class PaymentRequest {
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

	    static createFrom(source: any = {}) {
	        return new PaymentRequest(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.invoiceId = source["invoiceId"];
	        this.provider = source["provider"];
	        this.amount = source["amount"];
	        this.currency = source["currency"];
	        this.status = source["status"];
	        this.reference = source["reference"];
	        this.checkoutUrl = source["checkoutUrl"];
	        this.qrPayload = source["qrPayload"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class VendorDebitNote {
	    id: string;
	    vendorId: string;
	    vendorName: string;
	    purchaseOrderId: string;
	    amount: number;
	    reason: string;
	    status: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new VendorDebitNote(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.vendorId = source["vendorId"];
	        this.vendorName = source["vendorName"];
	        this.purchaseOrderId = source["purchaseOrderId"];
	        this.amount = source["amount"];
	        this.reason = source["reason"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class PurchaseOrderLine {
	    id: string;
	    purchaseOrderId: string;
	    ingredientId: string;
	    ingredientName: string;
	    orderedQty: number;
	    acceptedQty: number;
	    rejectedQty: number;
	    unitCost: number;
	    rejectionReason: string;

	    static createFrom(source: any = {}) {
	        return new PurchaseOrderLine(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.purchaseOrderId = source["purchaseOrderId"];
	        this.ingredientId = source["ingredientId"];
	        this.ingredientName = source["ingredientName"];
	        this.orderedQty = source["orderedQty"];
	        this.acceptedQty = source["acceptedQty"];
	        this.rejectedQty = source["rejectedQty"];
	        this.unitCost = source["unitCost"];
	        this.rejectionReason = source["rejectionReason"];
	    }
	}
	export class PurchaseOrder {
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

	    static createFrom(source: any = {}) {
	        return new PurchaseOrder(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.vendorId = source["vendorId"];
	        this.vendorName = source["vendorName"];
	        this.poNumber = source["poNumber"];
	        this.status = source["status"];
	        this.expectedDate = source["expectedDate"];
	        this.subtotal = source["subtotal"];
	        this.rejectedTotal = source["rejectedTotal"];
	        this.createdAt = source["createdAt"];
	        this.lines = this.convertValues(source["lines"], PurchaseOrderLine);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Vendor {
	    id: string;
	    name: string;
	    phone: string;
	    gstin: string;
	    paymentTerms: string;
	    qualityScore: number;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new Vendor(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.gstin = source["gstin"];
	        this.paymentTerms = source["paymentTerms"];
	        this.qualityScore = source["qualityScore"];
	        this.status = source["status"];
	    }
	}
	export class PrintJob {
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

	    static createFrom(source: any = {}) {
	        return new PrintJob(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.referenceId = source["referenceId"];
	        this.printerId = source["printerId"];
	        this.target = source["target"];
	        this.payload = source["payload"];
	        this.status = source["status"];
	        this.attempts = source["attempts"];
	        this.lastError = source["lastError"];
	        this.createdAt = source["createdAt"];
	        this.printedAt = source["printedAt"];
	    }
	}
	export class StaffMember {
	    id: string;
	    name: string;
	    role: string;
	    pinHash: string;
	    status: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new StaffMember(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.pinHash = source["pinHash"];
	        this.status = source["status"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class RestaurantSettings {
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

	    static createFrom(source: any = {}) {
	        return new RestaurantSettings(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.restaurantName = source["restaurantName"];
	        this.gstin = source["gstin"];
	        this.legalName = source["legalName"];
	        this.address = source["address"];
	        this.state = source["state"];
	        this.taxMode = source["taxMode"];
	        this.defaultTaxRate = source["defaultTaxRate"];
	        this.serviceChargeRate = source["serviceChargeRate"];
	        this.invoicePrefix = source["invoicePrefix"];
	        this.businessHours = source["businessHours"];
	        this.backupPath = source["backupPath"];
	        this.receiptPrinterId = source["receiptPrinterId"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class PilotWorkspace {
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

	    static createFrom(source: any = {}) {
	        return new PilotWorkspace(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = this.convertValues(source["settings"], RestaurantSettings);
	        this.integrations = this.convertValues(source["integrations"], IntegrationSetting);
	        this.staff = this.convertValues(source["staff"], StaffMember);
	        this.categories = this.convertValues(source["categories"], MenuCategory);
	        this.modifiers = this.convertValues(source["modifiers"], MenuModifier);
	        this.itemModifiers = this.convertValues(source["itemModifiers"], MenuItemModifier);
	        this.floorSections = this.convertValues(source["floorSections"], FloorSection);
	        this.tables = this.convertValues(source["tables"], DiningTable);
	        this.orderSessions = this.convertValues(source["orderSessions"], OrderSession);
	        this.paymentRequests = this.convertValues(source["paymentRequests"], PaymentRequest);
	        this.printJobs = this.convertValues(source["printJobs"], PrintJob);
	        this.vendors = this.convertValues(source["vendors"], Vendor);
	        this.purchaseOrders = this.convertValues(source["purchaseOrders"], PurchaseOrder);
	        this.debitNotes = this.convertValues(source["debitNotes"], VendorDebitNote);
	        this.dayClose = this.convertValues(source["dayClose"], DayCloseSummary);
	        this.accounting = this.convertValues(source["accounting"], AccountingSnapshot);
	        this.auditLog = this.convertValues(source["auditLog"], AuditLogEntry);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class PrinterConnectionInput {
	    target: string;
	    displayName: string;
	    mode: string;

	    static createFrom(source: any = {}) {
	        return new PrinterConnectionInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.displayName = source["displayName"];
	        this.mode = source["mode"];
	    }
	}

	export class PurchaseOrderLineInput {
	    ingredientId: string;
	    orderedQty: number;
	    unitCost: number;

	    static createFrom(source: any = {}) {
	        return new PurchaseOrderLineInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ingredientId = source["ingredientId"];
	        this.orderedQty = source["orderedQty"];
	        this.unitCost = source["unitCost"];
	    }
	}
	export class PurchaseOrderInput {
	    vendorId: string;
	    expectedDate: string;
	    lines: PurchaseOrderLineInput[];

	    static createFrom(source: any = {}) {
	        return new PurchaseOrderInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vendorId = source["vendorId"];
	        this.expectedDate = source["expectedDate"];
	        this.lines = this.convertValues(source["lines"], PurchaseOrderLineInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}


	export class ReceivePurchaseOrderLineInput {
	    lineId: string;
	    acceptedQty: number;
	    rejectedQty: number;
	    rejectionReason: string;

	    static createFrom(source: any = {}) {
	        return new ReceivePurchaseOrderLineInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lineId = source["lineId"];
	        this.acceptedQty = source["acceptedQty"];
	        this.rejectedQty = source["rejectedQty"];
	        this.rejectionReason = source["rejectionReason"];
	    }
	}
	export class ReceivePurchaseOrderInput {
	    purchaseOrderId: string;
	    lines: ReceivePurchaseOrderLineInput[];

	    static createFrom(source: any = {}) {
	        return new ReceivePurchaseOrderInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.purchaseOrderId = source["purchaseOrderId"];
	        this.lines = this.convertValues(source["lines"], ReceivePurchaseOrderLineInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}



	export class RecipeLineUpdateInput {
	    ingredientId: string;
	    quantity: number;
	    wasteFactorOverride?: number;

	    static createFrom(source: any = {}) {
	        return new RecipeLineUpdateInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ingredientId = source["ingredientId"];
	        this.quantity = source["quantity"];
	        this.wasteFactorOverride = source["wasteFactorOverride"];
	    }
	}
	export class RecipeUpdateInput {
	    itemId: string;
	    components: RecipeLineUpdateInput[];

	    static createFrom(source: any = {}) {
	        return new RecipeUpdateInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.components = this.convertValues(source["components"], RecipeLineUpdateInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ReconcileInput {
	    ingredientId: string;
	    physicalQty: number;
	    note: string;

	    static createFrom(source: any = {}) {
	        return new ReconcileInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ingredientId = source["ingredientId"];
	        this.physicalQty = source["physicalQty"];
	        this.note = source["note"];
	    }
	}
	export class RefundInvoiceInput {
	    invoiceId: string;
	    pin: string;
	    amount: number;
	    reason: string;

	    static createFrom(source: any = {}) {
	        return new RefundInvoiceInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoiceId = source["invoiceId"];
	        this.pin = source["pin"];
	        this.amount = source["amount"];
	        this.reason = source["reason"];
	    }
	}



	export class SaleLineInput {
	    itemId: string;
	    quantity: number;
	    notes: string;

	    static createFrom(source: any = {}) {
	        return new SaleLineInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.itemId = source["itemId"];
	        this.quantity = source["quantity"];
	        this.notes = source["notes"];
	    }
	}
	export class SaleInput {
	    customerName: string;
	    customerPhone: string;
	    channel: string;
	    orderType: string;
	    tableName: string;
	    discountType: string;
	    discountValue: number;
	    taxRate: number;
	    paymentMethod: string;
	    paymentTendered: number;
	    lines: SaleLineInput[];

	    static createFrom(source: any = {}) {
	        return new SaleInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.customerName = source["customerName"];
	        this.customerPhone = source["customerPhone"];
	        this.channel = source["channel"];
	        this.orderType = source["orderType"];
	        this.tableName = source["tableName"];
	        this.discountType = source["discountType"];
	        this.discountValue = source["discountValue"];
	        this.taxRate = source["taxRate"];
	        this.paymentMethod = source["paymentMethod"];
	        this.paymentTendered = source["paymentTendered"];
	        this.lines = this.convertValues(source["lines"], SaleLineInput);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}


	export class SessionLineVoidInput {
	    sessionId: string;
	    lineId: string;
	    staffId: string;
	    pin: string;
	    reason: string;

	    static createFrom(source: any = {}) {
	        return new SessionLineVoidInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.lineId = source["lineId"];
	        this.staffId = source["staffId"];
	        this.pin = source["pin"];
	        this.reason = source["reason"];
	    }
	}


	export class SplitLineInput {
	    lineId: string;
	    quantity: number;

	    static createFrom(source: any = {}) {
	        return new SplitLineInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lineId = source["lineId"];
	        this.quantity = source["quantity"];
	    }
	}
	export class SplitInvoiceInput {
	    invoiceId: string;
	    mode: string;
	    lines: SplitLineInput[];
	    amounts: number[];

	    static createFrom(source: any = {}) {
	        return new SplitInvoiceInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoiceId = source["invoiceId"];
	        this.mode = source["mode"];
	        this.lines = this.convertValues(source["lines"], SplitLineInput);
	        this.amounts = source["amounts"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class StaffActionApprovalInput {
	    staffId: string;
	    pin: string;
	    action: string;
	    targetId: string;

	    static createFrom(source: any = {}) {
	        return new StaffActionApprovalInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.staffId = source["staffId"];
	        this.pin = source["pin"];
	        this.action = source["action"];
	        this.targetId = source["targetId"];
	    }
	}
	export class StaffInput {
	    id: string;
	    name: string;
	    role: string;
	    pin: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new StaffInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.pin = source["pin"];
	        this.status = source["status"];
	    }
	}
	export class StaffLoginInput {
	    staffId: string;
	    pin: string;
	    workspace: string;

	    static createFrom(source: any = {}) {
	        return new StaffLoginInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.staffId = source["staffId"];
	        this.pin = source["pin"];
	        this.workspace = source["workspace"];
	    }
	}

	export class StaffSession {
	    staffId: string;
	    name: string;
	    role: string;
	    permissions: string[];
	    workspaceAccess: string[];
	    issuedAt: string;
	    expiresAt: string;

	    static createFrom(source: any = {}) {
	        return new StaffSession(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.staffId = source["staffId"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.permissions = source["permissions"];
	        this.workspaceAccess = source["workspaceAccess"];
	        this.issuedAt = source["issuedAt"];
	        this.expiresAt = source["expiresAt"];
	    }
	}
	export class SyncStatus {
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

	    static createFrom(source: any = {}) {
	        return new SyncStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pendingCount = source["pendingCount"];
	        this.syncedCount = source["syncedCount"];
	        this.failedCount = source["failedCount"];
	        this.oldestPendingAt = source["oldestPendingAt"];
	        this.lastSyncedAt = source["lastSyncedAt"];
	        this.lastError = source["lastError"];
	        this.databasePath = source["databasePath"];
	        this.databaseBytes = source["databaseBytes"];
	        this.walBytes = source["walBytes"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class TableMoveInput {
	    sessionId: string;
	    targetTableId: string;

	    static createFrom(source: any = {}) {
	        return new TableMoveInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.targetTableId = source["targetTableId"];
	    }
	}


	export class VendorInput {
	    id: string;
	    name: string;
	    phone: string;
	    gstin: string;
	    paymentTerms: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new VendorInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.phone = source["phone"];
	        this.gstin = source["gstin"];
	        this.paymentTerms = source["paymentTerms"];
	        this.status = source["status"];
	    }
	}
	export class VoidInvoiceInput {
	    invoiceId: string;
	    pin: string;
	    reason: string;

	    static createFrom(source: any = {}) {
	        return new VoidInvoiceInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.invoiceId = source["invoiceId"];
	        this.pin = source["pin"];
	        this.reason = source["reason"];
	    }
	}

}

