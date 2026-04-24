package command_test

// Tests for the ApproveApplicationHandler use case.

import (
	"context"
	"testing"

	"github.com/enterprise/trade-license/src/application/command"
	domainerrors "github.com/enterprise/trade-license/src/domain/errors"
	"github.com/enterprise/trade-license/src/domain/valueobjects"
	"github.com/enterprise/trade-license/src/testutil"
)

// seedAcceptedApplication creates and persists an ACCEPTED application
// (already passed the reviewer stage).
func seedAcceptedApplication(t *testing.T, repo *testutil.MockRepository) string {
	t.Helper()
	id := seedSubmittedApplication(t, repo)

	reviewHandler := command.NewReviewApplicationHandler(repo)
	_ = reviewHandler.Handle(context.Background(), command.ReviewApplicationCommand{
		ApplicationID: id,
		ReviewerID:    "reviewer-1",
		Action:        command.ReviewActionAccept,
	})
	return id
}

func TestApproveApplicationHandler_Approve(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedAcceptedApplication(t, repo)
	handler := command.NewApproveApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ApproveApplicationCommand{
		ApplicationID: id,
		ApproverID:    "approver-1",
		Action:        command.ApproveActionApprove,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	appID, _ := valueobjects.ApplicationIDFrom(id)
	app, _ := repo.FindByID(context.Background(), appID)
	if app.Status != valueobjects.StatusApproved {
		t.Errorf("expected APPROVED, got %s", app.Status)
	}
}

func TestApproveApplicationHandler_Reject(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedAcceptedApplication(t, repo)
	handler := command.NewApproveApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ApproveApplicationCommand{
		ApplicationID: id,
		ApproverID:    "approver-1",
		Action:        command.ApproveActionReject,
		Notes:         "does not meet zoning requirements",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	appID, _ := valueobjects.ApplicationIDFrom(id)
	app, _ := repo.FindByID(context.Background(), appID)
	if app.Status != valueobjects.StatusRejected {
		t.Errorf("expected REJECTED, got %s", app.Status)
	}
}

func TestApproveApplicationHandler_Rereview(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedAcceptedApplication(t, repo)
	handler := command.NewApproveApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ApproveApplicationCommand{
		ApplicationID: id,
		ApproverID:    "approver-1",
		Action:        command.ApproveActionRereview,
		Notes:         "commodity category needs clarification",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	appID, _ := valueobjects.ApplicationIDFrom(id)
	app, _ := repo.FindByID(context.Background(), appID)
	if app.Status != valueobjects.StatusRereview {
		t.Errorf("expected REREVIEW, got %s", app.Status)
	}
}

func TestApproveApplicationHandler_FailsFromSubmitted(t *testing.T) {
	repo := testutil.NewMockRepository()
	// Only SUBMITTED — reviewer hasn't accepted it yet
	id := seedSubmittedApplication(t, repo)
	handler := command.NewApproveApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ApproveApplicationCommand{
		ApplicationID: id,
		ApproverID:    "approver-1",
		Action:        command.ApproveActionApprove,
	})
	if err != domainerrors.ErrInvalidStatusTransition {
		t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestApproveApplicationHandler_UnknownAction(t *testing.T) {
	repo := testutil.NewMockRepository()
	id := seedAcceptedApplication(t, repo)
	handler := command.NewApproveApplicationHandler(repo)

	err := handler.Handle(context.Background(), command.ApproveApplicationCommand{
		ApplicationID: id,
		ApproverID:    "approver-1",
		Action:        "INVALID",
	})
	if err == nil {
		t.Error("expected error for unknown action")
	}
}
