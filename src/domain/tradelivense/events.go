package tradelivense

import "github.com/enterprise/trade-license/src/domain/common"

// Domain events record every meaningful state change in the Trade License workflow.
//
// Each event carries the minimum data needed for downstream consumers to act
// without querying the database. They are raised inside the aggregate methods
// and dispatched by the application layer after a successful persist.

// ApplicationSubmittedEvent is raised when a customer successfully submits
// a new trade license application (Pending → Submitted).
type ApplicationSubmittedEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ApplicantID   string
}

func (e ApplicationSubmittedEvent) EventName() string { return "ApplicationSubmitted" }

// ApplicationCancelledEvent is raised when a customer cancels their pending
// application before it enters the review stage (Pending → Cancelled).
type ApplicationCancelledEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ApplicantID   string
}

func (e ApplicationCancelledEvent) EventName() string { return "ApplicationCancelled" }

// ApplicationAcceptedEvent is raised when a reviewer accepts a submitted
// application, forwarding it to the approval stage (Submitted|Rereview → Accepted).
type ApplicationAcceptedEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ReviewerID    string
}

func (e ApplicationAcceptedEvent) EventName() string { return "ApplicationAccepted" }

// ApplicationRejectedEvent is raised when either a reviewer or an approver
// rejects an application. The ActorID identifies which role took the action.
type ApplicationRejectedEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ActorID       string
	Reason        string
}

func (e ApplicationRejectedEvent) EventName() string { return "ApplicationRejected" }

// ApplicationAdjustedEvent is raised when a reviewer flags an application for
// adjustment — meaning it requires correction before it can be accepted
// (Submitted|Rereview → Adjusted).
type ApplicationAdjustedEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ReviewerID    string
	Notes         string
}

func (e ApplicationAdjustedEvent) EventName() string { return "ApplicationAdjusted" }

// ApplicationApprovedEvent is raised when an approver grants final approval,
// completing the workflow (Accepted → Approved).
type ApplicationApprovedEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ApproverID    string
}

func (e ApplicationApprovedEvent) EventName() string { return "ApplicationApproved" }

// ApplicationRereviewEvent is raised when an approver sends an accepted
// application back to the reviewer for additional scrutiny (Accepted → Rereview).
type ApplicationRereviewEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ApproverID    string
	Notes         string
}

func (e ApplicationRereviewEvent) EventName() string { return "ApplicationRereview" }

// ApplicationResubmittedEvent is raised when a customer resubmits an application
// that was sent back for adjustment (Adjusted → Submitted).
type ApplicationResubmittedEvent struct {
	common.BaseDomainEvent
	ApplicationID string
	ApplicantID   string
}

func (e ApplicationResubmittedEvent) EventName() string { return "ApplicationResubmitted" }
