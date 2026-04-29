package nexus

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

func (s *Store) GetAdminAnalytics(rangeKey string) (AdminAnalytics, error) {
	if err := s.refreshAnalyticsSnapshots(); err != nil {
		return AdminAnalytics{}, err
	}

	startDate, endDate, normalizedRange, label := analyticsRange(rangeKey)
	daily, err := s.analyticsDailyRows(startDate, endDate)
	if err != nil {
		return AdminAnalytics{}, err
	}
	hourly, err := s.analyticsHourlyRows(startDate, endDate)
	if err != nil {
		return AdminAnalytics{}, err
	}
	itemEntries, err := s.analyticsItemEntries(startDate, endDate)
	if err != nil {
		return AdminAnalytics{}, err
	}
	ingredients, err := s.ingredients()
	if err != nil {
		return AdminAnalytics{}, err
	}
	menuItems, err := s.menuItems()
	if err != nil {
		return AdminAnalytics{}, err
	}
	tickets, err := s.kitchenTickets()
	if err != nil {
		return AdminAnalytics{}, err
	}
	printJobs, err := s.printJobs()
	if err != nil {
		return AdminAnalytics{}, err
	}
	paymentRequests, err := s.paymentRequests()
	if err != nil {
		return AdminAnalytics{}, err
	}
	purchaseOrders, err := s.purchaseOrders()
	if err != nil {
		return AdminAnalytics{}, err
	}
	debitNotes, err := s.vendorDebitNotes()
	if err != nil {
		return AdminAnalytics{}, err
	}
	accounting, err := s.GetAccountingSnapshot()
	if err != nil {
		return AdminAnalytics{}, err
	}
	todayClose, err := s.dayClosePreview(todayDate())
	if err != nil {
		return AdminAnalytics{}, err
	}
	latestCashVariance, err := s.latestCashVariance()
	if err != nil {
		return AdminAnalytics{}, err
	}
	status, err := s.analyticsSnapshotStatus()
	if err != nil {
		return AdminAnalytics{}, err
	}

	totals := summarizeDailyRows(daily)
	netRevenue := round2(totals.salesTotal - totals.refundTotal)
	averageOrder := 0.0
	if totals.orderCount > 0 {
		averageOrder = round2(netRevenue / float64(totals.orderCount))
	}
	stockRisk := 0
	for _, ingredient := range ingredients {
		if ingredient.ReorderPoint > 0 && ingredient.OnHandQty <= ingredient.ReorderPoint {
			stockRisk++
		}
	}
	vendorExposure := openVendorExposure(purchaseOrders)
	failureCount := operationalFailureCount(printJobs, paymentRequests)

	analytics := AdminAnalytics{
		GeneratedAt:        nowUTC(),
		RangeKey:           normalizedRange,
		RangeLabel:         label,
		DemoFallback:       totals.orderCount == 0,
		SalesTrend:         dailyRowsToPoints(daily),
		HourlyHeatmap:      hourlyRowsToPoints(hourly),
		TenderMix:          tenderMixPoints(totals),
		CategoryMix:        categoryMixPoints(itemEntries),
		ItemVelocity:       itemVelocityPoints(itemEntries),
		ContributionMargin: contributionMarginPoints(itemEntries),
		InventoryHealth:    inventoryHealthRows(ingredients),
		KitchenPerformance: kitchenPerformanceRows(tickets),
		PurchaseTrend:      purchaseTrendPoints(purchaseOrders),
		ItemMatrix:         itemMatrixBuckets(itemEntries, menuItems),
		Settlement: SettlementHealth{
			CashExpected:     round2(todayClose.CashExpected),
			CashVariance:     round2(latestCashVariance),
			UPITotal:         round2(totals.upiTotal),
			CardTotal:        round2(totals.cardTotal),
			RazorpayTotal:    round2(totals.razorpayTotal),
			RazorpayClearing: round2(accounting.RazorpayClearing),
			TaxPayable:       round2(accounting.TaxPayable),
			VendorPayables:   round2(accounting.VendorPayables),
			RefundTotal:      round2(totals.refundTotal),
		},
		SnapshotStatus: status,
	}
	analytics.Executive = []AnalyticsMetric{
		{ID: "net_revenue", Label: "Net revenue", Value: netRevenue, Format: "currency", Detail: fmt.Sprintf("%d closed orders", totals.orderCount), Tone: "ink"},
		{ID: "orders", Label: "Orders", Value: float64(totals.orderCount), Format: "number", Detail: "Closed bills in range", Tone: "green"},
		{ID: "avg_bill", Label: "Average bill", Value: averageOrder, Format: "currency", Detail: "Revenue per order", Tone: "gold"},
		{ID: "tax_payable", Label: "GST payable", Value: round2(accounting.TaxPayable), Format: "currency", Detail: "Accounting shell", Tone: "blue"},
		{ID: "cash_variance", Label: "Cash variance", Value: round2(latestCashVariance), Format: "currency", Detail: "Latest closed day", Tone: metricToneForVariance(latestCashVariance)},
		{ID: "refunds_voids", Label: "Refunds + voids", Value: round2(totals.refundTotal), Format: "currency", Detail: fmt.Sprintf("%d voided bills", totals.voidCount), Tone: "coral"},
		{ID: "stock_risk", Label: "Stock risk", Value: float64(stockRisk), Format: "number", Detail: "At or below reorder", Tone: metricToneForCount(stockRisk)},
		{ID: "vendor_exposure", Label: "Vendor exposure", Value: vendorExposure, Format: "currency", Detail: "Open purchase orders", Tone: "gold"},
		{ID: "ops_failures", Label: "Ops failures", Value: float64(failureCount), Format: "number", Detail: "Print/payment exceptions", Tone: metricToneForCount(failureCount)},
	}
	analytics.Exceptions = analyticsExceptions(analytics, ingredients, tickets, printJobs, paymentRequests, debitNotes, totals)
	analytics.Recommendations = analyticsRecommendations(analytics)
	if analytics.DemoFallback {
		applyDemoFallback(&analytics, menuItems)
	}
	return analytics, nil
}

type analyticsDailyRow struct {
	date          string
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

type analyticsTotals struct {
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

type analyticsHourlyRow struct {
	label      string
	hour       int
	salesTotal float64
	orderCount int
}

func (s *Store) refreshAnalyticsSnapshots() error {
	dates, err := s.analyticsSourceDates()
	if err != nil {
		return err
	}
	now := nowUTC()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := tx.Exec("DELETE FROM analytics_daily_snapshots"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM analytics_item_snapshots"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM analytics_hourly_snapshots"); err != nil {
		return err
	}
	for _, date := range dates {
		summary, err := analyticsDaySummaryTx(tx, date)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO analytics_daily_snapshots (
				business_date, sales_total, order_count, refund_total, void_count, discount_total,
				cash_total, upi_total, card_total, razorpay_total, updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, date, summary.salesTotal, summary.orderCount, summary.refundTotal, summary.voidCount, summary.discountTotal,
			summary.cashTotal, summary.upiTotal, summary.cardTotal, summary.razorpayTotal, now); err != nil {
			return err
		}
	}
	if err := rebuildItemSnapshotsTx(tx, now); err != nil {
		return err
	}
	if err := rebuildHourlySnapshotsTx(tx, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) analyticsSourceDates() ([]string, error) {
	rows, err := s.db.Query(`
		SELECT business_date FROM (
			SELECT substr(created_at, 1, 10) AS business_date FROM sales
			UNION
			SELECT substr(created_at, 1, 10) AS business_date FROM refunds
			UNION
			SELECT substr(created_at, 1, 10) AS business_date FROM payments
			UNION
			SELECT substr(created_at, 1, 10) AS business_date FROM purchase_orders
		)
		WHERE business_date <> ''
		ORDER BY business_date
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dates := make([]string, 0)
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}
	return dates, rows.Err()
}

func analyticsDaySummaryTx(tx *sql.Tx, date string) (analyticsDailyRow, error) {
	start, end := dayBounds(date)
	row := analyticsDailyRow{date: date}
	if err := tx.QueryRow(`
		SELECT COALESCE(SUM(total), 0), COUNT(*), COALESCE(SUM(discount_total), 0)
		FROM sales
		WHERE created_at >= ? AND created_at < ? AND status IN ('closed', 'partially_refunded', 'refunded')
	`, start, end).Scan(&row.salesTotal, &row.orderCount, &row.discountTotal); err != nil {
		return row, err
	}
	if err := tx.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM refunds WHERE created_at >= ? AND created_at < ?", start, end).Scan(&row.refundTotal); err != nil {
		return row, err
	}
	if err := tx.QueryRow("SELECT COUNT(*) FROM sales WHERE created_at >= ? AND created_at < ? AND status = 'voided'", start, end).Scan(&row.voidCount); err != nil {
		return row, err
	}
	rows, err := tx.Query("SELECT method, COALESCE(SUM(amount), 0) FROM payments WHERE created_at >= ? AND created_at < ? GROUP BY method", start, end)
	if err != nil {
		return row, err
	}
	defer rows.Close()
	for rows.Next() {
		var method string
		var amount float64
		if err := rows.Scan(&method, &amount); err != nil {
			return row, err
		}
		switch strings.ToLower(method) {
		case "cash":
			row.cashTotal += amount
		case "upi":
			row.upiTotal += amount
		case "card":
			row.cardTotal += amount
		case "razorpay":
			row.razorpayTotal += amount
		}
	}
	return row, rows.Err()
}

func rebuildItemSnapshotsTx(tx *sql.Tx, updatedAt string) error {
	rows, err := tx.Query(`
		SELECT substr(s.created_at, 1, 10) AS business_date,
			sl.item_id,
			sl.item_name,
			COALESCE(NULLIF(mi.category, ''), 'Unassigned') AS category,
			COALESCE(SUM(sl.quantity), 0) AS quantity,
			COALESCE(SUM(sl.line_total), 0) AS gross_sales,
			COALESCE(SUM(mi.cost * sl.quantity), 0) AS cost_total
		FROM sale_lines sl
		JOIN sales s ON s.id = sl.sale_id
		LEFT JOIN menu_items mi ON mi.id = sl.item_id
		WHERE s.status IN ('closed', 'partially_refunded', 'refunded')
		GROUP BY business_date, sl.item_id, sl.item_name, category
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var date, itemID, itemName, category string
		var quantity int
		var grossSales, costTotal float64
		if err := rows.Scan(&date, &itemID, &itemName, &category, &quantity, &grossSales, &costTotal); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO analytics_item_snapshots (
				business_date, item_id, item_name, category, quantity, gross_sales, cost_total, margin, updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, date, itemID, itemName, category, quantity, round2(grossSales), round2(costTotal), round2(grossSales-costTotal), updatedAt); err != nil {
			return err
		}
	}
	return rows.Err()
}

func rebuildHourlySnapshotsTx(tx *sql.Tx, updatedAt string) error {
	rows, err := tx.Query(`
		SELECT substr(created_at, 1, 10) AS business_date,
			CAST(substr(created_at, 12, 2) AS INTEGER) AS hour,
			COALESCE(SUM(total), 0),
			COUNT(*)
		FROM sales
		WHERE status IN ('closed', 'partially_refunded', 'refunded')
		GROUP BY business_date, hour
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		var hour, count int
		var salesTotal float64
		if err := rows.Scan(&date, &hour, &salesTotal, &count); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO analytics_hourly_snapshots (business_date, hour, sales_total, order_count, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`, date, hour, round2(salesTotal), count, updatedAt); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *Store) analyticsDailyRows(startDate, endDate string) ([]analyticsDailyRow, error) {
	rows, err := s.db.Query(`
		SELECT business_date, sales_total, order_count, refund_total, void_count, discount_total,
			cash_total, upi_total, card_total, razorpay_total
		FROM analytics_daily_snapshots
		WHERE business_date >= ? AND business_date <= ?
		ORDER BY business_date
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]analyticsDailyRow, 0)
	for rows.Next() {
		var row analyticsDailyRow
		if err := rows.Scan(&row.date, &row.salesTotal, &row.orderCount, &row.refundTotal, &row.voidCount, &row.discountTotal,
			&row.cashTotal, &row.upiTotal, &row.cardTotal, &row.razorpayTotal); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (s *Store) analyticsHourlyRows(startDate, endDate string) ([]analyticsHourlyRow, error) {
	rows, err := s.db.Query(`
		SELECT hour, COALESCE(SUM(sales_total), 0), COALESCE(SUM(order_count), 0)
		FROM analytics_hourly_snapshots
		WHERE business_date >= ? AND business_date <= ?
		GROUP BY hour
		ORDER BY hour
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]analyticsHourlyRow, 0)
	for rows.Next() {
		var row analyticsHourlyRow
		if err := rows.Scan(&row.hour, &row.salesTotal, &row.orderCount); err != nil {
			return nil, err
		}
		row.label = fmt.Sprintf("%02d:00", row.hour)
		items = append(items, row)
	}
	return items, rows.Err()
}

func (s *Store) analyticsItemEntries(startDate, endDate string) ([]ItemMatrixEntry, error) {
	rows, err := s.db.Query(`
		SELECT item_id, item_name, category,
			COALESCE(SUM(quantity), 0),
			COALESCE(SUM(gross_sales), 0),
			COALESCE(SUM(margin), 0)
		FROM analytics_item_snapshots
		WHERE business_date >= ? AND business_date <= ?
		GROUP BY item_id, item_name, category
		ORDER BY SUM(gross_sales) DESC
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ItemMatrixEntry, 0)
	for rows.Next() {
		var item ItemMatrixEntry
		if err := rows.Scan(&item.ItemID, &item.Name, &item.Category, &item.Quantity, &item.Sales, &item.Margin); err != nil {
			return nil, err
		}
		if item.Sales > 0 {
			item.MarginPct = round2((item.Margin / item.Sales) * 100)
		}
		item.Sales = round2(item.Sales)
		item.Margin = round2(item.Margin)
		items = append(items, item)
	}
	return items, rows.Err()
}

func summarizeDailyRows(rows []analyticsDailyRow) analyticsTotals {
	var totals analyticsTotals
	for _, row := range rows {
		totals.salesTotal += row.salesTotal
		totals.orderCount += row.orderCount
		totals.refundTotal += row.refundTotal
		totals.voidCount += row.voidCount
		totals.discountTotal += row.discountTotal
		totals.cashTotal += row.cashTotal
		totals.upiTotal += row.upiTotal
		totals.cardTotal += row.cardTotal
		totals.razorpayTotal += row.razorpayTotal
	}
	totals.salesTotal = round2(totals.salesTotal)
	totals.refundTotal = round2(totals.refundTotal)
	totals.discountTotal = round2(totals.discountTotal)
	return totals
}

func dailyRowsToPoints(rows []analyticsDailyRow) []AnalyticsPoint {
	points := make([]AnalyticsPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, AnalyticsPoint{
			Label: shortDateLabel(row.date),
			Value: round2(row.salesTotal - row.refundTotal),
			Count: row.orderCount,
			Tone:  "green",
		})
	}
	return points
}

func hourlyRowsToPoints(rows []analyticsHourlyRow) []AnalyticsPoint {
	points := make([]AnalyticsPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, AnalyticsPoint{
			Label: row.label,
			Value: round2(row.salesTotal),
			Count: row.orderCount,
			Tone:  "blue",
		})
	}
	return points
}

func tenderMixPoints(totals analyticsTotals) []AnalyticsPoint {
	points := []AnalyticsPoint{
		{Label: "Cash", Value: round2(totals.cashTotal), Tone: "green"},
		{Label: "UPI", Value: round2(totals.upiTotal), Tone: "blue"},
		{Label: "Card", Value: round2(totals.cardTotal), Tone: "gold"},
		{Label: "Razorpay", Value: round2(totals.razorpayTotal), Tone: "coral"},
	}
	return nonZeroOrOriginal(points)
}

func categoryMixPoints(entries []ItemMatrixEntry) []AnalyticsPoint {
	byCategory := map[string]float64{}
	counts := map[string]int{}
	for _, entry := range entries {
		category := firstNonEmpty(entry.Category, "Unassigned")
		byCategory[category] += entry.Sales
		counts[category] += entry.Quantity
	}
	points := make([]AnalyticsPoint, 0, len(byCategory))
	for category, value := range byCategory {
		points = append(points, AnalyticsPoint{Label: category, Value: round2(value), Count: counts[category], Tone: "green"})
	}
	sortPointsByValue(points)
	return points
}

func itemVelocityPoints(entries []ItemMatrixEntry) []AnalyticsPoint {
	sorted := append([]ItemMatrixEntry(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Quantity == sorted[j].Quantity {
			return sorted[i].Sales > sorted[j].Sales
		}
		return sorted[i].Quantity > sorted[j].Quantity
	})
	points := make([]AnalyticsPoint, 0, minInt(len(sorted), 8))
	for _, entry := range sorted {
		if len(points) >= 8 {
			break
		}
		points = append(points, AnalyticsPoint{Label: entry.Name, Value: float64(entry.Quantity), Count: entry.Quantity, Tone: "blue"})
	}
	return points
}

func contributionMarginPoints(entries []ItemMatrixEntry) []AnalyticsPoint {
	sorted := append([]ItemMatrixEntry(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Margin > sorted[j].Margin
	})
	points := make([]AnalyticsPoint, 0, minInt(len(sorted), 8))
	for _, entry := range sorted {
		if len(points) >= 8 {
			break
		}
		points = append(points, AnalyticsPoint{Label: entry.Name, Value: round2(entry.Margin), Count: entry.Quantity, Tone: "gold"})
	}
	return points
}

func inventoryHealthRows(ingredients []Ingredient) []InventoryHealthRow {
	rows := make([]InventoryHealthRow, 0, len(ingredients))
	for _, ingredient := range ingredients {
		reorder := math.Max(ingredient.ReorderPoint, 1)
		ratio := ingredient.OnHandQty / reorder
		status := "healthy"
		if ratio <= 1 {
			status = "critical"
		} else if ratio <= 2 {
			status = "watch"
		}
		rows = append(rows, InventoryHealthRow{
			ID:           ingredient.ID,
			Name:         ingredient.Name,
			OnHandQty:    round2(ingredient.OnHandQty),
			ReorderPoint: round2(ingredient.ReorderPoint),
			Unit:         ingredient.Unit,
			RiskScore:    round2(clamp(100-(ratio*50), 0, 100)),
			Status:       status,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].RiskScore > rows[j].RiskScore
	})
	return rows
}

func kitchenPerformanceRows(tickets []KitchenTicket) []KitchenPerformanceRow {
	now := time.Now().UTC()
	type stationAccumulator struct {
		row      KitchenPerformanceRow
		ageTotal float64
		ageCount int
	}
	byRoute := map[string]*stationAccumulator{}
	for _, ticket := range tickets {
		route := firstNonEmpty(ticket.RouteName, "Kitchen")
		acc := byRoute[route]
		if acc == nil {
			acc = &stationAccumulator{row: KitchenPerformanceRow{RouteName: route}}
			byRoute[route] = acc
		}
		switch strings.ToLower(ticket.Status) {
		case "ready":
			acc.row.ReadyTickets++
		case "served":
			acc.row.ServedTickets++
		default:
			acc.row.OpenTickets++
		}
		if !strings.EqualFold(ticket.Status, "served") {
			age := ticketAgeMinutes(ticket.CreatedAt, now)
			acc.ageTotal += age
			acc.ageCount++
			if age > acc.row.OldestAgeMin {
				acc.row.OldestAgeMin = age
			}
		}
	}
	rows := make([]KitchenPerformanceRow, 0, len(byRoute))
	for _, acc := range byRoute {
		if acc.ageCount > 0 {
			acc.row.AverageAgeMin = round2(acc.ageTotal / float64(acc.ageCount))
		}
		acc.row.OldestAgeMin = round2(acc.row.OldestAgeMin)
		rows = append(rows, acc.row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].OpenTickets > rows[j].OpenTickets
	})
	return rows
}

func purchaseTrendPoints(orders []PurchaseOrder) []AnalyticsPoint {
	byDate := map[string]AnalyticsPoint{}
	for _, order := range orders {
		date := firstNonEmpty(safeDate(order.CreatedAt), order.ExpectedDate)
		if date == "" {
			date = todayDate()
		}
		point := byDate[date]
		point.Label = shortDateLabel(date)
		point.Value += order.Subtotal
		point.Count++
		point.Tone = "gold"
		if order.RejectedTotal > 0 {
			point.Tone = "coral"
		}
		byDate[date] = point
	}
	points := make([]AnalyticsPoint, 0, len(byDate))
	for _, point := range byDate {
		point.Value = round2(point.Value)
		points = append(points, point)
	}
	sort.Slice(points, func(i, j int) bool {
		return points[i].Label < points[j].Label
	})
	return points
}

func itemMatrixBuckets(entries []ItemMatrixEntry, menuItems []MenuItem) []ItemMatrixBucket {
	allEntries := append([]ItemMatrixEntry(nil), entries...)
	seen := map[string]bool{}
	for _, entry := range allEntries {
		seen[entry.ItemID] = true
	}
	for _, item := range menuItems {
		if seen[item.ID] {
			continue
		}
		marginPct := 0.0
		if item.Price > 0 {
			marginPct = round2(((item.Price - item.Cost) / item.Price) * 100)
		}
		allEntries = append(allEntries, ItemMatrixEntry{
			ItemID:    item.ID,
			Name:      item.Name,
			Category:  item.Category,
			MarginPct: marginPct,
		})
	}
	velocityThreshold := medianQuantity(allEntries)
	marginThreshold := medianMarginPct(allEntries)
	if velocityThreshold <= 0 {
		velocityThreshold = 1
	}
	if marginThreshold <= 0 {
		marginThreshold = 50
	}
	buckets := []ItemMatrixBucket{
		{ID: "stars", Label: "High margin / high velocity", Description: "Protect prep speed and availability."},
		{ID: "premium_sleepers", Label: "High margin / low velocity", Description: "Merchandise, bundle, and train servers."},
		{ID: "traffic_drivers", Label: "Low margin / high velocity", Description: "Watch cost creep and portioning."},
		{ID: "rework", Label: "Low margin / low velocity", Description: "Reprice, reformulate, or hide."},
	}
	for _, entry := range allEntries {
		highMargin := entry.MarginPct >= marginThreshold
		highVelocity := float64(entry.Quantity) >= velocityThreshold
		switch {
		case highMargin && highVelocity:
			buckets[0].Items = append(buckets[0].Items, entry)
		case highMargin:
			buckets[1].Items = append(buckets[1].Items, entry)
		case highVelocity:
			buckets[2].Items = append(buckets[2].Items, entry)
		default:
			buckets[3].Items = append(buckets[3].Items, entry)
		}
	}
	for index := range buckets {
		sort.Slice(buckets[index].Items, func(i, j int) bool {
			return buckets[index].Items[i].Sales > buckets[index].Items[j].Sales
		})
		if len(buckets[index].Items) > 6 {
			buckets[index].Items = buckets[index].Items[:6]
		}
	}
	return buckets
}

func analyticsExceptions(analytics AdminAnalytics, ingredients []Ingredient, tickets []KitchenTicket, printJobs []PrintJob, paymentRequests []PaymentRequest, debitNotes []VendorDebitNote, totals analyticsTotals) []AnalyticsException {
	exceptions := make([]AnalyticsException, 0)
	for _, ingredient := range ingredients {
		if ingredient.ReorderPoint > 0 && ingredient.OnHandQty <= ingredient.ReorderPoint {
			exceptions = append(exceptions, AnalyticsException{
				ID:       "stock-" + ingredient.ID,
				Kind:     "Stock",
				Title:    ingredient.Name + " is at reorder",
				Detail:   fmt.Sprintf("%.0f %s on hand against %.0f reorder", ingredient.OnHandQty, ingredient.Unit, ingredient.ReorderPoint),
				Severity: "high",
				Value:    ingredient.OnHandQty,
			})
		}
	}
	for _, ticket := range tickets {
		if strings.EqualFold(ticket.Status, "served") {
			continue
		}
		age := ticketAgeMinutes(ticket.CreatedAt, time.Now().UTC())
		if age >= 15 {
			exceptions = append(exceptions, AnalyticsException{
				ID:       "kot-" + ticket.ID,
				Kind:     "Kitchen",
				Title:    ticket.TicketNumber + " is aging",
				Detail:   fmt.Sprintf("%s has been open for %.0f minutes", ticket.RouteName, age),
				Severity: "medium",
				Value:    age,
			})
		}
	}
	for _, job := range printJobs {
		if strings.EqualFold(job.Status, "failed") {
			exceptions = append(exceptions, AnalyticsException{
				ID:       "print-" + job.ID,
				Kind:     "Printer",
				Title:    humanStatusText(job.Kind) + " print failed",
				Detail:   firstNonEmpty(job.LastError, job.Target),
				Severity: "high",
				Value:    float64(job.Attempts),
			})
		}
	}
	for _, request := range paymentRequests {
		if strings.EqualFold(request.Status, "failed") || strings.EqualFold(request.Status, "cancelled") {
			exceptions = append(exceptions, AnalyticsException{
				ID:       "payment-" + request.ID,
				Kind:     "Payment",
				Title:    request.Reference + " needs review",
				Detail:   fmt.Sprintf("%s for %.0f", humanStatusText(request.Status), request.Amount),
				Severity: "medium",
				Value:    request.Amount,
			})
		}
	}
	if math.Abs(analytics.Settlement.CashVariance) > 0.01 {
		exceptions = append(exceptions, AnalyticsException{
			ID:       "cash-variance",
			Kind:     "Settlement",
			Title:    "Cash variance exists",
			Detail:   "Latest closed day has a counted-versus-expected difference.",
			Severity: "medium",
			Value:    analytics.Settlement.CashVariance,
		})
	}
	if totals.voidCount > 0 {
		exceptions = append(exceptions, AnalyticsException{
			ID:       "voids",
			Kind:     "Control",
			Title:    "Voids recorded",
			Detail:   fmt.Sprintf("%d voided bills in the selected range", totals.voidCount),
			Severity: "medium",
			Value:    float64(totals.voidCount),
		})
	}
	for _, note := range debitNotes {
		if !strings.EqualFold(note.Status, "closed") {
			exceptions = append(exceptions, AnalyticsException{
				ID:       "debit-" + note.ID,
				Kind:     "Vendor",
				Title:    note.VendorName + " debit note open",
				Detail:   firstNonEmpty(note.Reason, "Rejected goods pending follow-up"),
				Severity: "medium",
				Value:    note.Amount,
			})
		}
	}
	sort.Slice(exceptions, func(i, j int) bool {
		return severityRank(exceptions[i].Severity) > severityRank(exceptions[j].Severity)
	})
	if len(exceptions) > 10 {
		return exceptions[:10]
	}
	return exceptions
}

func analyticsRecommendations(analytics AdminAnalytics) []AnalyticsRecommendation {
	recommendations := make([]AnalyticsRecommendation, 0)
	if len(analytics.Exceptions) > 0 {
		recommendations = append(recommendations, AnalyticsRecommendation{
			ID:       "clear-exceptions",
			Title:    "Clear the exception queue first",
			Detail:   fmt.Sprintf("%d operational exceptions are visible across kitchen, stock, payment, or printer flows.", len(analytics.Exceptions)),
			Priority: 1,
			Page:     "settings",
		})
	}
	for _, bucket := range analytics.ItemMatrix {
		if bucket.ID == "premium_sleepers" && len(bucket.Items) > 0 {
			recommendations = append(recommendations, AnalyticsRecommendation{
				ID:       "move-premium-sleepers",
				Title:    "Push high-margin sleepers",
				Detail:   bucket.Items[0].Name + " has strong margin but needs menu placement or staff prompts.",
				Priority: 2,
				Page:     "recipes",
			})
			break
		}
	}
	if analytics.Settlement.RazorpayClearing > 0 {
		recommendations = append(recommendations, AnalyticsRecommendation{
			ID:       "reconcile-razorpay",
			Title:    "Reconcile Razorpay clearing",
			Detail:   "Match payment requests to settlements before close.",
			Priority: 3,
			Page:     "integrations",
		})
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, AnalyticsRecommendation{
			ID:       "protect-winners",
			Title:    "Protect the winners",
			Detail:   "Keep best-sellers available and review contribution margin before the next menu change.",
			Priority: 3,
			Page:     "recipes",
		})
	}
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority < recommendations[j].Priority
	})
	return recommendations
}

func applyDemoFallback(analytics *AdminAnalytics, menuItems []MenuItem) {
	analytics.RangeLabel = analytics.RangeLabel + " / demo baseline"
	if len(analytics.SalesTrend) == 0 {
		analytics.SalesTrend = []AnalyticsPoint{
			{Label: "Mon", Value: 18200, Count: 64, Tone: "green"},
			{Label: "Tue", Value: 22100, Count: 78, Tone: "green"},
			{Label: "Wed", Value: 20500, Count: 72, Tone: "green"},
			{Label: "Thu", Value: 24800, Count: 86, Tone: "green"},
			{Label: "Fri", Value: 31400, Count: 104, Tone: "green"},
			{Label: "Sat", Value: 38200, Count: 128, Tone: "green"},
			{Label: "Sun", Value: 29600, Count: 98, Tone: "green"},
		}
	}
	if len(analytics.HourlyHeatmap) == 0 {
		analytics.HourlyHeatmap = []AnalyticsPoint{
			{Label: "10:00", Value: 2800, Count: 9, Tone: "blue"},
			{Label: "12:00", Value: 8200, Count: 27, Tone: "blue"},
			{Label: "14:00", Value: 6400, Count: 19, Tone: "blue"},
			{Label: "18:00", Value: 9600, Count: 31, Tone: "blue"},
			{Label: "21:00", Value: 11800, Count: 38, Tone: "blue"},
		}
	}
	if len(analytics.TenderMix) == 0 || totalPointValue(analytics.TenderMix) == 0 {
		analytics.TenderMix = []AnalyticsPoint{
			{Label: "Cash", Value: 28000, Tone: "green"},
			{Label: "UPI", Value: 72000, Tone: "blue"},
			{Label: "Card", Value: 41000, Tone: "gold"},
			{Label: "Razorpay", Value: 18000, Tone: "coral"},
		}
	}
	if len(analytics.CategoryMix) == 0 {
		byCategory := map[string]float64{}
		for _, item := range menuItems {
			byCategory[firstNonEmpty(item.Category, "Menu")] += item.Price
		}
		for category, value := range byCategory {
			analytics.CategoryMix = append(analytics.CategoryMix, AnalyticsPoint{Label: category, Value: round2(value), Tone: "green"})
		}
		sortPointsByValue(analytics.CategoryMix)
	}
	analytics.Recommendations = append([]AnalyticsRecommendation{{
		ID:       "close-pilot-bills",
		Title:    "Close live bills to replace demo analytics",
		Detail:   "The command center is showing a demo baseline until this restaurant has closed-sale history.",
		Priority: 1,
		Page:     "counter",
	}}, analytics.Recommendations...)
}

func (s *Store) analyticsSnapshotStatus() (AnalyticsSnapshotStatus, error) {
	status := AnalyticsSnapshotStatus{UpdatedAt: nowUTC()}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM analytics_daily_snapshots").Scan(&status.DailyRows); err != nil {
		return status, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM analytics_item_snapshots").Scan(&status.ItemRows); err != nil {
		return status, err
	}
	if err := s.db.QueryRow("SELECT COUNT(*) FROM analytics_hourly_snapshots").Scan(&status.HourlyRows); err != nil {
		return status, err
	}
	var updated sql.NullString
	if err := s.db.QueryRow(`
		SELECT MAX(updated_at) FROM (
			SELECT updated_at FROM analytics_daily_snapshots
			UNION ALL SELECT updated_at FROM analytics_item_snapshots
			UNION ALL SELECT updated_at FROM analytics_hourly_snapshots
		)
	`).Scan(&updated); err != nil {
		return status, err
	}
	if updated.Valid && updated.String != "" {
		status.UpdatedAt = updated.String
	}
	return status, nil
}

func (s *Store) latestCashVariance() (float64, error) {
	var variance sql.NullFloat64
	err := s.db.QueryRow("SELECT cash_variance FROM day_closes ORDER BY business_date DESC LIMIT 1").Scan(&variance)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !variance.Valid {
		return 0, nil
	}
	return variance.Float64, nil
}

func analyticsRange(rangeKey string) (string, string, string, string) {
	key := strings.ToLower(strings.TrimSpace(rangeKey))
	days := 30
	switch key {
	case "7d", "7", "week":
		key = "7d"
		days = 7
	case "90d", "90", "quarter":
		key = "90d"
		days = 90
	case "all", "all_time":
		key = "all"
		return "1970-01-01", todayDate(), key, "All time"
	default:
		key = "30d"
	}
	today := time.Now().UTC()
	start := today.AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	return start, today.Format("2006-01-02"), key, fmt.Sprintf("Last %d days", days)
}

func shortDateLabel(date string) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.Format("02 Jan")
}

func safeDate(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	return ""
}

func sortPointsByValue(points []AnalyticsPoint) {
	sort.Slice(points, func(i, j int) bool {
		return points[i].Value > points[j].Value
	})
}

func nonZeroOrOriginal(points []AnalyticsPoint) []AnalyticsPoint {
	for _, point := range points {
		if point.Value != 0 {
			return points
		}
	}
	return points
}

func totalPointValue(points []AnalyticsPoint) float64 {
	total := 0.0
	for _, point := range points {
		total += point.Value
	}
	return total
}

func openVendorExposure(orders []PurchaseOrder) float64 {
	total := 0.0
	for _, order := range orders {
		status := strings.ToLower(order.Status)
		if status == "received" || status == "closed" || status == "cancelled" {
			continue
		}
		total += order.Subtotal
	}
	return round2(total)
}

func operationalFailureCount(printJobs []PrintJob, paymentRequests []PaymentRequest) int {
	count := 0
	for _, job := range printJobs {
		if strings.EqualFold(job.Status, "failed") {
			count++
		}
	}
	for _, request := range paymentRequests {
		if strings.EqualFold(request.Status, "failed") || strings.EqualFold(request.Status, "cancelled") {
			count++
		}
	}
	return count
}

func metricToneForCount(count int) string {
	if count > 0 {
		return "coral"
	}
	return "green"
}

func metricToneForVariance(value float64) string {
	if math.Abs(value) > 0.01 {
		return "coral"
	}
	return "green"
}

func ticketAgeMinutes(createdAt string, now time.Time) float64 {
	created, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return 0
	}
	return math.Max(0, now.Sub(created).Minutes())
}

func medianQuantity(items []ItemMatrixEntry) float64 {
	values := make([]int, 0, len(items))
	for _, item := range items {
		values = append(values, item.Quantity)
	}
	sort.Ints(values)
	if len(values) == 0 {
		return 0
	}
	mid := len(values) / 2
	if len(values)%2 == 0 {
		return float64(values[mid-1]+values[mid]) / 2
	}
	return float64(values[mid])
}

func medianMarginPct(items []ItemMatrixEntry) float64 {
	values := make([]float64, 0, len(items))
	for _, item := range items {
		values = append(values, item.MarginPct)
	}
	sort.Float64s(values)
	if len(values) == 0 {
		return 0
	}
	mid := len(values) / 2
	if len(values)%2 == 0 {
		return (values[mid-1] + values[mid]) / 2
	}
	return values[mid]
}

func severityRank(severity string) int {
	switch strings.ToLower(severity) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
