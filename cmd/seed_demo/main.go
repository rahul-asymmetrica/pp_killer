package main

import (
	"flag"
	"fmt"
	"os"

	"NEXUS/internal/nexus"
)

func main() {
	months := flag.Int("months", 6, "months of closed operational history to seed")
	dbPath := flag.String("db", "", "SQLite database path; defaults to the app database")
	flag.Parse()

	path := *dbPath
	if path == "" {
		var err error
		path, err = nexus.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve db path: %v\n", err)
			os.Exit(1)
		}
	}

	store, err := nexus.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open store: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	result, err := store.SeedDemoOperations(*months)
	if err != nil {
		fmt.Fprintf(os.Stderr, "seed demo operations: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Seeded %s\n", path)
	fmt.Printf("Months: %d\n", result.Months)
	fmt.Printf("Business days: %d\n", result.BusinessDays)
	fmt.Printf("Tables: %d\n", result.Tables)
	fmt.Printf("Waiters: %d\n", result.Waiters)
	fmt.Printf("Customers: %d\n", result.Customers)
	fmt.Printf("Invoices: %d\n", result.Invoices)
	fmt.Printf("Kitchen tickets: %d\n", result.KitchenTickets)
	fmt.Printf("Day closes: %d\n", result.DayCloses)
	fmt.Printf("Purchase orders: %d\n", result.PurchaseOrders)
	fmt.Printf("Refunds: %d\n", result.Refunds)
	fmt.Printf("Voids: %d\n", result.Voids)
	fmt.Printf("Sales total: %.2f\n", result.SalesTotal)
	fmt.Printf("Range: %s to %s\n", result.StartedAt, result.EndedAt)
}
