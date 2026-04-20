package command_test

// Tests for the ReviewApplicationHandler use case.
// Each test seeds the mock repository with an application in the correct
// starting state, then exercises the reviewer action.

import (
	"context"
	"testing"

	"github.com/enterprise/trade-license/src/application/command"
	"github.com/enterprise/trade-license/src/domain/tradelivense"
	"github.com/enterprise/trade-license/src/testutil"
)

// seedSubmittedApplication is a helper that creates and persists a SUBMITTED application.
func seedSubmittedApplication(t *testing.T, repo *testutil.MockRepository) string {
	t.Helper()
	app := testutil.NewReadyToSubmitApplication("customer-1")
	if err := app.Submit(); err != nil {
		t.Fatalf("seed: submit failed: %v", err)
	}
	if err := repo.Save(context.Background(), app); err != nil {
		t.Fatalf("seed: save failed: %v", err)
	}
	return app.ID.String()
}

func TestReviewApplicationHandler_Accept(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedSubmittedApplication(t, repo)
	handler := command.NewReviewApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ReviewApplicationCommand{
		ApplicationID: id,
		ReviewerID:    "reviewer-1",
		Action:        command.ReviewActionAccept,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	appID, _ := tradelivense.ApplicationIDFrom(id)
	app, _ := repo.FindByID(context.Background(), appID)
	if app.Status != tradelivense.StatusAccepted {
		t.Errorf("expected ACCEPTED, got %s", app.Status)
	}
}

func TestReviewApplicationHandler_Reject(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedSubmittedApplication(t, repo)
	handler := command.NewReviewApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ReviewApplicationCommand{
		ApplicationID: id,
		ReviewerID:    "reviewer-1",
		Action:        command.ReviewActionReject,
		Notes:         "incomplete documentation",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	appID, _ := tradelivense.ApplicationIDFrom(id)
	app, _ := repo.FindByID(context.Background(), appID)
	if app.Status != tradelivense.StatusRejected {
		t.Errorf("expected REJECTED, got %s", app.Status)
	}
	if app.Notes != "incomplete documentation" {
		t.Error("expected rejection notes to be persisted")
	}
}

func TestReviewApplicationHandler_Adjust(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedSubmittedApplication(t, repo)
	handler := command.NewReviewApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ReviewApplicationCommand{
		ApplicationID: id,
		ReviewerID:    "reviewer-1",
		Action:        command.ReviewActionAdjust,
		Notes:         "please resubmit page 2",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	appID, _ := tradelivense.ApplicationIDFrom(id)
	app, _ := repo.FindByID(context.Background(), appID)
	if app.Status != tradelivense.StatusAdjusted {
		t.Errorf("expected ADJUSTED, got %s", app.Status)
	}
}

func TestReviewApplicationHandler_UnknownAction(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedSubmittedApplication(t, repo)
	handler := command.NewReviewApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ReviewApplicationCommand{
		ApplicationID: id,
		ReviewerID:    "reviewer-1",
		Action:        "INVALID_ACTION",
	})
	if err == nil {
		t.Error("expected error for unknown action")
	}
}

func TestReviewApplicationHandler_NotFound(t *testing.T) {
	repo := testutil.NewMockRepository()
	handler := command.NewReviewApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ReviewApplicationCommand{
		ApplicationID: "00000000-0000-0000-0000-000000000000",
		ReviewerID:    "reviewer-1",
		Action:        command.ReviewActionAccept,
	})
	if err != tradelivense.ErrApplicationNotFound {
		t.Errorf("expected ErrApplicationNotFound, got %v", err)
	}
}
