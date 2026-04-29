package nexus

import (
	"errors"
	"strings"
)

func (s *Store) UpdateRecipe(input RecipeUpdateInput) (Dashboard, error) {
	if strings.TrimSpace(input.ItemID) == "" {
		return Dashboard{}, errors.New("menu item is required")
	}
	if len(input.Components) == 0 {
		return Dashboard{}, errors.New("recipe requires at least one component")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	if _, err := menuItemByID(tx, input.ItemID); err != nil {
		return Dashboard{}, err
	}

	if _, err := tx.Exec("DELETE FROM recipes WHERE item_id = ?", input.ItemID); err != nil {
		return Dashboard{}, err
	}

	seen := map[string]bool{}
	for _, component := range input.Components {
		if strings.TrimSpace(component.IngredientID) == "" {
			return Dashboard{}, errors.New("ingredient is required")
		}
		if component.Quantity <= 0 {
			return Dashboard{}, errors.New("recipe quantity must be positive")
		}
		if seen[component.IngredientID] {
			return Dashboard{}, errors.New("recipe cannot contain the same ingredient twice")
		}
		seen[component.IngredientID] = true

		var exists int
		if err := tx.QueryRow("SELECT COUNT(*) FROM ingredients WHERE id = ?", component.IngredientID).Scan(&exists); err != nil {
			return Dashboard{}, err
		}
		if exists == 0 {
			return Dashboard{}, errors.New("ingredient not found")
		}

		var override any
		if component.WasteFactorOverride != nil {
			override = *component.WasteFactorOverride
		}
		if _, err := tx.Exec(
			"INSERT INTO recipes (id, item_id, ingredient_id, quantity, waste_factor_override) VALUES (?, ?, ?, ?, ?)",
			mustID(), input.ItemID, component.IngredientID, component.Quantity, override,
		); err != nil {
			return Dashboard{}, err
		}
	}

	createdAt := nowUTC()
	if err := insertSyncLog(tx, "recipe", input.ItemID, "update", input, createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}

func (s *Store) UpdateIngredientSettings(input IngredientUpdateInput) (Dashboard, error) {
	if strings.TrimSpace(input.IngredientID) == "" {
		return Dashboard{}, errors.New("ingredient is required")
	}
	if input.ReorderPoint < 0 {
		return Dashboard{}, errors.New("reorder point cannot be negative")
	}
	if input.PurchaseToUsage <= 0 {
		return Dashboard{}, errors.New("purchase conversion must be positive")
	}
	if strings.TrimSpace(input.PurchaseUnit) == "" {
		return Dashboard{}, errors.New("purchase unit is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Dashboard{}, err
	}
	defer rollback(tx)

	result, err := tx.Exec(`
		UPDATE ingredients
		SET reorder_point = ?, purchase_unit = ?, purchase_to_usage = ?, last_purchase_cost = ?
		WHERE id = ?
	`, input.ReorderPoint, strings.TrimSpace(input.PurchaseUnit), input.PurchaseToUsage, input.LastPurchaseCost, input.IngredientID)
	if err != nil {
		return Dashboard{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Dashboard{}, err
	}
	if affected == 0 {
		return Dashboard{}, errors.New("ingredient not found")
	}

	createdAt := nowUTC()
	if err := insertSyncLog(tx, "ingredient_settings", input.IngredientID, "update", input, createdAt); err != nil {
		return Dashboard{}, err
	}
	if err := tx.Commit(); err != nil {
		return Dashboard{}, err
	}
	return s.Dashboard()
}
