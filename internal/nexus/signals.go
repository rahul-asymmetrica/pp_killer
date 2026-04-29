package nexus

import (
	"fmt"
	"time"
)

func (s *Store) signals(ingredients []Ingredient, customers []Customer, deliveries []DeliveryBatch, items []MenuItem) []Signal {
	signals := make([]Signal, 0, 8)

	for _, ingredient := range ingredients {
		if ingredient.OnHandQty <= ingredient.ReorderPoint {
			signals = append(signals, Signal{
				ID:       "stock_" + ingredient.ID,
				Kind:     "inventory",
				Title:    ingredient.Name + " is near reorder",
				Detail:   fmt.Sprintf("%.0f %s left against a %.0f %s reorder point.", ingredient.OnHandQty, ingredient.Unit, ingredient.ReorderPoint, ingredient.Unit),
				Action:   "Create purchase reminder",
				Priority: 1,
			})
		}
		if ingredient.WasteFactor >= 0.1 {
			signals = append(signals, Signal{
				ID:       "waste_" + ingredient.ID,
				Kind:     "waste",
				Title:    ingredient.Name + " waste is drifting high",
				Detail:   fmt.Sprintf("Current dynamic waste factor is %.1f%% after recent audits.", ingredient.WasteFactor*100),
				Action:   "Review prep and storage",
				Priority: 2,
			})
		}
	}

	for _, delivery := range deliveries {
		if delivery.Status == "partial_rejected" {
			signals = append(signals, Signal{
				ID:       "delivery_" + delivery.ID,
				Kind:     "vendor",
				Title:    "Vendor debit note pending",
				Detail:   fmt.Sprintf("%s has a partially rejected batch that should be adjusted before payment.", delivery.VendorName),
				Action:   "Send WhatsApp note",
				Priority: 1,
			})
		}
	}

	now := time.Now().UTC()
	for _, customer := range customers {
		if customer.VisitCount >= 3 {
			signals = append(signals, Signal{
				ID:       "whale_" + customer.ID,
				Kind:     "crm",
				Title:    customer.Name + " is a high-value regular",
				Detail:   fmt.Sprintf("%d visits, %.0f total spend, favorite item: %s.", customer.VisitCount, customer.TotalSpend, customer.FavoriteItem),
				Action:   "Queue loyalty reward",
				Priority: 2,
			})
			break
		}
		if parsed, err := time.Parse(time.RFC3339, customer.LastVisitAt); err == nil && now.Sub(parsed) > 14*24*time.Hour {
			signals = append(signals, Signal{
				ID:       "risk_" + customer.ID,
				Kind:     "crm",
				Title:    customer.Name + " may be slipping",
				Detail:   "This guest has passed their usual return window.",
				Action:   "Send win-back offer",
				Priority: 2,
			})
		}
	}

	for _, item := range items {
		if item.Price > 0 {
			margin := (item.Price - item.Cost) / item.Price
			if margin >= 0.55 {
				signals = append(signals, Signal{
					ID:       "menu_" + item.ID,
					Kind:     "menu",
					Title:    item.Name + " is margin-friendly",
					Detail:   fmt.Sprintf("Approximate gross margin is %.0f%%. Feature it in combos or specials.", margin*100),
					Action:   "Draft menu push",
					Priority: 3,
				})
				break
			}
		}
	}

	return signals
}
