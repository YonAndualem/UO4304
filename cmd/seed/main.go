// seed populates the database with one application in every workflow state.
//
// This gives testers an immediate data set so they can exercise every role's
// endpoints without having to manually drive an application through each stage.
//
// Run with:
//
//	docker compose run --rm seed
//
// Printed IDs can be pasted directly into Postman or curl.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/enterprise/trade-license/src/application/auth"
	"github.com/enterprise/trade-license/src/config"
	"github.com/enterprise/trade-license/src/domain/tradelivense"
	postgresrepo "github.com/enterprise/trade-license/src/infrastructure/persistence/postgres"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	db, err := postgresrepo.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("seed: failed to connect: %v", err)
	}

	userRepo := postgresrepo.NewUserRepository(db)
	authSvc  := auth.NewService(userRepo, cfg.JWTSecret)

	repo := postgresrepo.NewApplicationRepository(db)
	ctx := context.Background()

	fmt.Println("\n========================================")
	fmt.Println("  Trade License — Database Seed")
	fmt.Println("========================================")

	// ── Seed user accounts ────────────────────────────────────────────────────
	// All demo users share the password "demo" for easy exploration.
	demoUsers := []struct {
		userID string
		role   string
	}{
		{"customer-seed-001", "CUSTOMER"},
		{"customer-seed-002", "CUSTOMER"},
		{"customer-seed-003", "CUSTOMER"},
		{"customer-seed-004", "CUSTOMER"},
		{"customer-seed-005", "CUSTOMER"},
		{"customer-seed-006", "CUSTOMER"},
		{"customer-seed-007", "CUSTOMER"},
		{"customer-seed-008", "CUSTOMER"},
		{"reviewer-seed-001", "REVIEWER"},
		{"approver-seed-001", "APPROVER"},
	}

	fmt.Println("\n── User accounts ──")
	for _, u := range demoUsers {
		err := authSvc.Register(u.userID, "demo", u.role)
		if err != nil && !errors.Is(err, auth.ErrUserExists) {
			log.Fatalf("seed: register %s: %v", u.userID, err)
		}
		status := "created"
		if errors.Is(err, auth.ErrUserExists) {
			status = "already exists"
		}
		fmt.Printf("  %-30s [%s]  %s\n", u.userID, u.role, status)
	}
	fmt.Println("  (all demo accounts use password: demo)")

	// ── 1. PENDING ────────────────────────────────────────────────────────────
	// Customer has started the process but not yet submitted.
	pendingApp := buildBase("customer-seed-001")
	if err := repo.Save(ctx, pendingApp); err != nil {
		log.Fatalf("seed: PENDING: %v", err)
	}
	fmt.Printf("\n[PENDING]   customer-seed-001  →  %s\n", pendingApp.ID.String())

	// ── 2. SUBMITTED ──────────────────────────────────────────────────────────
	// Customer has submitted; waiting for reviewer.
	submittedApp := buildReady("customer-seed-002")
	if err := submittedApp.Submit(); err != nil {
		log.Fatalf("seed: Submit: %v", err)
	}
	if err := repo.Save(ctx, submittedApp); err != nil {
		log.Fatalf("seed: SUBMITTED: %v", err)
	}
	fmt.Printf("[SUBMITTED] customer-seed-002  →  %s\n", submittedApp.ID.String())

	// ── 3. ACCEPTED ───────────────────────────────────────────────────────────
	// Reviewer has accepted; waiting for approver.
	acceptedApp := buildReady("customer-seed-003")
	_ = acceptedApp.Submit()
	if err := acceptedApp.Accept("reviewer-seed-001"); err != nil {
		log.Fatalf("seed: Accept: %v", err)
	}
	if err := repo.Save(ctx, acceptedApp); err != nil {
		log.Fatalf("seed: ACCEPTED: %v", err)
	}
	fmt.Printf("[ACCEPTED]  customer-seed-003  →  %s\n", acceptedApp.ID.String())

	// ── 4. APPROVED ───────────────────────────────────────────────────────────
	// Fully approved — terminal success state.
	approvedApp := buildReady("customer-seed-004")
	_ = approvedApp.Submit()
	_ = approvedApp.Accept("reviewer-seed-001")
	if err := approvedApp.Approve("approver-seed-001"); err != nil {
		log.Fatalf("seed: Approve: %v", err)
	}
	if err := repo.Save(ctx, approvedApp); err != nil {
		log.Fatalf("seed: APPROVED: %v", err)
	}
	fmt.Printf("[APPROVED]  customer-seed-004  →  %s\n", approvedApp.ID.String())

	// ── 5. REJECTED (by reviewer) ─────────────────────────────────────────────
	rejectedByReviewerApp := buildReady("customer-seed-005")
	_ = rejectedByReviewerApp.Submit()
	if err := rejectedByReviewerApp.ReviewReject("reviewer-seed-001", "Passport copy is not legible"); err != nil {
		log.Fatalf("seed: ReviewReject: %v", err)
	}
	if err := repo.Save(ctx, rejectedByReviewerApp); err != nil {
		log.Fatalf("seed: REJECTED(reviewer): %v", err)
	}
	fmt.Printf("[REJECTED]  customer-seed-005  →  %s  (rejected by reviewer)\n", rejectedByReviewerApp.ID.String())

	// ── 6. ADJUSTED ───────────────────────────────────────────────────────────
	// Reviewer flagged the application for correction.
	adjustedApp := buildReady("customer-seed-006")
	_ = adjustedApp.Submit()
	if err := adjustedApp.Adjust("reviewer-seed-001", "Please upload a higher-resolution scan of the trade certificate"); err != nil {
		log.Fatalf("seed: Adjust: %v", err)
	}
	if err := repo.Save(ctx, adjustedApp); err != nil {
		log.Fatalf("seed: ADJUSTED: %v", err)
	}
	fmt.Printf("[ADJUSTED]  customer-seed-006  →  %s\n", adjustedApp.ID.String())

	// ── 7. REJECTED (by approver) ─────────────────────────────────────────────
	rejectedByApproverApp := buildReady("customer-seed-007")
	_ = rejectedByApproverApp.Submit()
	_ = rejectedByApproverApp.Accept("reviewer-seed-001")
	if err := rejectedByApproverApp.ApproveReject("approver-seed-001", "Does not comply with municipal zoning regulation 4.2"); err != nil {
		log.Fatalf("seed: ApproveReject: %v", err)
	}
	if err := repo.Save(ctx, rejectedByApproverApp); err != nil {
		log.Fatalf("seed: REJECTED(approver): %v", err)
	}
	fmt.Printf("[REJECTED]  customer-seed-007  →  %s  (rejected by approver)\n", rejectedByApproverApp.ID.String())

	// ── 8. REREVIEW ───────────────────────────────────────────────────────────
	// Approver sent the application back to the reviewer for further scrutiny.
	rereviewApp := buildReady("customer-seed-008")
	_ = rereviewApp.Submit()
	_ = rereviewApp.Accept("reviewer-seed-001")
	if err := rereviewApp.Rereview("approver-seed-001", "Commodity description is too vague — please clarify the product category"); err != nil {
		log.Fatalf("seed: Rereview: %v", err)
	}
	if err := repo.Save(ctx, rereviewApp); err != nil {
		log.Fatalf("seed: REREVIEW: %v", err)
	}
	fmt.Printf("[REREVIEW]  customer-seed-008  →  %s\n", rereviewApp.ID.String())

	fmt.Println("\n========================================")
	fmt.Println("  Seed complete — 8 applications created")
	fmt.Println("========================================")
	fmt.Println("\nUseful test IDs:")
	fmt.Printf("  Submit new app    →  POST /api/customer/applications  (X-User-ID: customer-seed-001)\n")
	fmt.Printf("  Review queue      →  GET  /api/reviewer/applications  (shows SUBMITTED + REREVIEW)\n")
	fmt.Printf("  Accept this one   →  POST /api/reviewer/applications/%s/action\n", submittedApp.ID.String())
	fmt.Printf("  Approve queue     →  GET  /api/approver/applications\n")
	fmt.Printf("  Approve this one  →  POST /api/approver/applications/%s/action\n", acceptedApp.ID.String())
	fmt.Println()
}

// buildBase creates a PENDING application with commodity but no documents or payment.
func buildBase(applicantID string) *tradelivense.TradeLicenseApplication {
	lt, _ := tradelivense.NewLicenseType(tradelivense.TradeLicense)
	app := tradelivense.NewTradeLicenseApplication(applicantID, lt)
	app.SelectCommodity(tradelivense.NewCommodity("General Trading", "Import and export of consumer goods", "Commerce"))
	return app
}

// buildReady creates a PENDING application with all pre-conditions satisfied
// so that Submit() will succeed immediately.
// A UUID-based transaction ID is used so re-running the seed never violates
// the unique constraint on payments.transaction_id.
func buildReady(applicantID string) *tradelivense.TradeLicenseApplication {
	app := buildBase(applicantID)
	app.AttachDocument(tradelivense.NewDocument(
		"Passport Copy",
		fmt.Sprintf("https://storage.example.com/seed/%s/passport.pdf", applicantID),
		"application/pdf",
	))
	app.AttachDocument(tradelivense.NewDocument(
		"Business Registration",
		fmt.Sprintf("https://storage.example.com/seed/%s/business-reg.pdf", applicantID),
		"application/pdf",
	))
	app.SettlePayment(tradelivense.NewPayment(
		500.00,
		"USD",
		"TXN-"+uuid.New().String(), // unique per run — prevents duplicate key on re-seed
	))
	return app
}
