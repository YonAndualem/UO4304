// reset deletes all non-demo data from the database.
//
// Demo accounts (customer-seed-*, reviewer-seed-001, approver-seed-001) and
// their applications are preserved. Every user and application created by real
// users is permanently removed.
//
// Run with:
//
//	docker compose run --rm reset
package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/enterprise/trade-license/src/config"
	postgresrepo "github.com/enterprise/trade-license/src/infrastructure/persistence/postgres"
)

// seedUserIDs are the accounts created by the seed command. Everything else is deleted.
var seedUserIDs = []string{
	"customer-seed-001",
	"customer-seed-002",
	"customer-seed-003",
	"customer-seed-004",
	"customer-seed-005",
	"customer-seed-006",
	"customer-seed-007",
	"customer-seed-008",
	"reviewer-seed-001",
	"approver-seed-001",
}

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	db, err := postgresrepo.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("reset: failed to connect: %v", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("  Trade License — Data Reset")
	fmt.Println("========================================")
	fmt.Println("\nDemo seed accounts and their data will be preserved.")
	fmt.Println("All other users and applications will be deleted.")
	fmt.Println()

	// Delete non-seed applications (hard delete — bypasses soft delete).
	res := db.Exec(
		"DELETE FROM applications WHERE applicant_id NOT IN ?",
		seedUserIDs,
	)
	if res.Error != nil {
		log.Fatalf("reset: delete applications: %v", res.Error)
	}
	fmt.Printf("  Applications deleted: %d\n", res.RowsAffected)

	// Delete non-seed user accounts.
	res = db.Exec(
		"DELETE FROM users WHERE user_id NOT IN ?",
		seedUserIDs,
	)
	if res.Error != nil {
		log.Fatalf("reset: delete users: %v", res.Error)
	}
	fmt.Printf("  User accounts deleted: %d\n", res.RowsAffected)

	fmt.Println("\n========================================")
	fmt.Println("  Reset complete")
	fmt.Println("========================================")
	fmt.Println()
}
