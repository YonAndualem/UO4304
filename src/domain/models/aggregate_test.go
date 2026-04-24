package models_test

// This file tests the TradeLicenseApplication aggregate in isolation.
// Because the domain layer has no external dependencies, these tests require
// no mocks, no database, and no HTTP stack — they run purely in memory.
//
// Test naming convention: Test<Type>_<Method>_<Scenario>

import (
	"testing"

	"github.com/enterprise/trade-license/src/domain/models"
	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
	"github.com/enterprise/trade-license/src/testutil"
)

// ─── Submit ──────────────────────────────────────────────────────────────────

func TestTradeLicenseApplication_Submit_HappyPath(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")

	if err := app.Submit(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusSubmitted {
		t.Errorf("expected status SUBMITTED, got %s", app.Status)
	}
}

func TestTradeLicenseApplication_Submit_EmitsEvent(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()

	events := app.PullEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "ApplicationSubmitted" {
		t.Errorf("expected ApplicationSubmitted event, got %s", events[0].EventName())
	}
}

func TestTradeLicenseApplication_Submit_FailsWithoutDocuments(t *testing.T) {
	app := testutil.NewPendingApplication("applicant-1")
	app.SettlePayment(models.NewPayment(500, "USD", "TXN-001"))
	// No documents attached

	err := app.Submit()
	if err != domainerrors.ErrDocumentRequired {
		t.Errorf("expected ErrDocumentRequired, got %v", err)
	}
}

func TestTradeLicenseApplication_Submit_FailsWithoutPayment(t *testing.T) {
	app := testutil.NewPendingApplication("applicant-1")
	app.AttachDocument(models.NewDocument("Passport", "https://storage/passport.pdf", "application/pdf"))
	// No payment settled

	err := app.Submit()
	if err != domainerrors.ErrPaymentRequired {
		t.Errorf("expected ErrPaymentRequired, got %v", err)
	}
}

func TestTradeLicenseApplication_Submit_FailsWhenNotPending(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit() // Move to SUBMITTED

	// Attempting to submit again should fail
	err := app.Submit()
	if err != domainerrors.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func TestTradeLicenseApplication_Cancel_HappyPath(t *testing.T) {
	app := testutil.NewPendingApplication("applicant-1")

	if err := app.Cancel(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusCancelled {
		t.Errorf("expected status CANCELLED, got %s", app.Status)
	}
}

func TestTradeLicenseApplication_Cancel_EmitsEvent(t *testing.T) {
	app := testutil.NewPendingApplication("applicant-1")
	_ = app.Cancel()

	events := app.PullEvents()
	if len(events) != 1 || events[0].EventName() != "ApplicationCancelled" {
		t.Errorf("expected ApplicationCancelled event")
	}
}

func TestTradeLicenseApplication_Cancel_FailsWhenNotPending(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit() // Move to SUBMITTED

	err := app.Cancel()
	if err != domainerrors.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

// ─── Accept (Reviewer) ────────────────────────────────────────────────────────

func TestTradeLicenseApplication_Accept_FromSubmitted(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()

	if err := app.Accept("reviewer-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusAccepted {
		t.Errorf("expected status ACCEPTED, got %s", app.Status)
	}
}

func TestTradeLicenseApplication_Accept_FromRereview(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()
	_ = app.Accept("reviewer-1")
	_ = app.Rereview("approver-1", "needs more info")

	// Reviewer can accept again after rereview
	if err := app.Accept("reviewer-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusAccepted {
		t.Errorf("expected status ACCEPTED, got %s", app.Status)
	}
}

func TestTradeLicenseApplication_Accept_FailsFromPending(t *testing.T) {
	app := testutil.NewPendingApplication("applicant-1")

	err := app.Accept("reviewer-1")
	if err != domainerrors.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

// ─── ReviewReject ─────────────────────────────────────────────────────────────

func TestTradeLicenseApplication_ReviewReject_HappyPath(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()

	if err := app.ReviewReject("reviewer-1", "missing trade certificate"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusRejected {
		t.Errorf("expected status REJECTED, got %s", app.Status)
	}
	if app.Notes != "missing trade certificate" {
		t.Errorf("expected notes to be set")
	}
}

// ─── Adjust ───────────────────────────────────────────────────────────────────

func TestTradeLicenseApplication_Adjust_HappyPath(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()

	if err := app.Adjust("reviewer-1", "please upload a clearer copy"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusAdjusted {
		t.Errorf("expected status ADJUSTED, got %s", app.Status)
	}
}

func TestTradeLicenseApplication_Adjust_FailsFromAccepted(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()
	_ = app.Accept("reviewer-1")

	err := app.Adjust("reviewer-1", "notes")
	if err != domainerrors.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

// ─── Approve ──────────────────────────────────────────────────────────────────

func TestTradeLicenseApplication_Approve_HappyPath(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()
	_ = app.Accept("reviewer-1")

	if err := app.Approve("approver-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusApproved {
		t.Errorf("expected status APPROVED, got %s", app.Status)
	}
}

func TestTradeLicenseApplication_Approve_EmitsEvent(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()
	_ = app.Accept("reviewer-1")
	app.PullEvents() // Clear preceding events

	_ = app.Approve("approver-1")
	events := app.PullEvents()
	if len(events) != 1 || events[0].EventName() != "ApplicationApproved" {
		t.Errorf("expected ApplicationApproved event")
	}
}

func TestTradeLicenseApplication_Approve_FailsFromSubmitted(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()

	err := app.Approve("approver-1")
	if err != domainerrors.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

// ─── ApproveReject ────────────────────────────────────────────────────────────

func TestTradeLicenseApplication_ApproveReject_HappyPath(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()
	_ = app.Accept("reviewer-1")

	if err := app.ApproveReject("approver-1", "does not meet requirements"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusRejected {
		t.Errorf("expected status REJECTED, got %s", app.Status)
	}
}

// ─── Rereview ─────────────────────────────────────────────────────────────────

func TestTradeLicenseApplication_Rereview_HappyPath(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()
	_ = app.Accept("reviewer-1")

	if err := app.Rereview("approver-1", "commodity description unclear"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if app.Status != valueobjects.StatusRereview {
		t.Errorf("expected status REREVIEW, got %s", app.Status)
	}
}

func TestTradeLicenseApplication_Rereview_FailsFromSubmitted(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")
	_ = app.Submit()

	err := app.Rereview("approver-1", "notes")
	if err != domainerrors.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

// ─── Full workflow ─────────────────────────────────────────────────────────────

// TestFullApprovalWorkflow walks an application through the complete happy path:
// Pending → Submitted → Accepted → Approved.
func TestFullApprovalWorkflow(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")

	steps := []struct {
		name   string
		action func() error
		want   valueobjects.ApplicationStatus
	}{
		{"submit", func() error { return app.Submit() }, valueobjects.StatusSubmitted},
		{"accept", func() error { return app.Accept("reviewer-1") }, valueobjects.StatusAccepted},
		{"approve", func() error { return app.Approve("approver-1") }, valueobjects.StatusApproved},
	}

	for _, step := range steps {
		if err := step.action(); err != nil {
			t.Fatalf("step %q: unexpected error: %v", step.name, err)
		}
		if app.Status != step.want {
			t.Errorf("after step %q: expected %s, got %s", step.name, step.want, app.Status)
		}
	}
}

// TestRereviewWorkflow tests the approver sending the application back for re-examination:
// Pending → Submitted → Accepted → Rereview → Accepted → Approved.
func TestRereviewWorkflow(t *testing.T) {
	app := testutil.NewReadyToSubmitApplication("applicant-1")

	_ = app.Submit()
	_ = app.Accept("reviewer-1")
	_ = app.Rereview("approver-1", "check commodity again")

	if app.Status != valueobjects.StatusRereview {
		t.Fatalf("expected REREVIEW, got %s", app.Status)
	}

	// Reviewer accepts again after rereview
	_ = app.Accept("reviewer-1")
	_ = app.Approve("approver-1")

	if app.Status != valueobjects.StatusApproved {
		t.Errorf("expected APPROVED after rereview path, got %s", app.Status)
	}
}
