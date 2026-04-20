# Trade License Workflow System

A production-grade enterprise application implementing the three-role Trade License
approval workflow using **Domain-Driven Design (DDD)** and **Clean Architecture**.
Written in Go (backend) and Next.js (frontend).

---

## Table of Contents

1. [What This System Does](#1-what-this-system-does)
2. [Architecture Fundamentals](#2-architecture-fundamentals)
   - 2.1 [What is Clean Architecture?](#21-what-is-clean-architecture)
   - 2.2 [What is Domain-Driven Design?](#22-what-is-domain-driven-design)
   - 2.3 [How They Work Together](#23-how-they-work-together)
   - 2.4 [The Dependency Rule](#24-the-dependency-rule)
   - 2.5 [What is CQRS?](#25-what-is-cqrs)
3. [Project Structure](#3-project-structure)
4. [The Domain Layer](#4-the-domain-layer)
   - 4.1 [Bounded Context](#41-bounded-context)
   - 4.2 [Aggregate Root](#42-aggregate-root)
   - 4.3 [Entities](#43-entities)
   - 4.4 [Value Objects](#44-value-objects)
   - 4.5 [Domain Events](#45-domain-events)
   - 4.6 [Repository Interface (Port)](#46-repository-interface-port)
   - 4.7 [Error Sentinels](#47-error-sentinels)
   - 4.8 [Common Package](#48-common-package)
5. [The Application Layer](#5-the-application-layer)
   - 5.1 [Command Handlers (Write Side)](#51-command-handlers-write-side)
   - 5.2 [Query Handlers (Read Side)](#52-query-handlers-read-side)
6. [The Infrastructure Layer](#6-the-infrastructure-layer)
   - 6.1 [GORM Models](#61-gorm-models)
   - 6.2 [Database Initialization](#62-database-initialization)
   - 6.3 [Mapper](#63-mapper)
   - 6.4 [Repository Implementation](#64-repository-implementation)
7. [The Presentation Layer](#7-the-presentation-layer)
   - 7.1 [Middleware](#71-middleware)
   - 7.2 [HTTP Handlers](#72-http-handlers)
   - 7.3 [Router](#73-router)
8. [The Composition Root](#8-the-composition-root)
9. [The Frontend](#9-the-frontend)
   - 9.1 [Technology Stack](#91-technology-stack)
   - 9.2 [Directory Structure](#92-directory-structure)
   - 9.3 [Pages](#93-pages)
   - 9.4 [Components](#94-components)
   - 9.5 [API Client](#95-api-client)
   - 9.6 [Type Definitions](#96-type-definitions)
10. [Workflow State Machine](#10-workflow-state-machine)
    - 10.1 [Status Diagram](#101-status-diagram)
    - 10.2 [Role Responsibilities](#102-role-responsibilities)
    - 10.3 [The ADJUSTED Cycle](#103-the-adjusted-cycle)
11. [API Reference](#11-api-reference)
12. [Database Schema](#12-database-schema)
13. [Configuration](#13-configuration)
14. [How to Run](#14-how-to-run)
15. [Seed Data](#15-seed-data)
16. [Testing](#16-testing)
17. [Security Notes](#17-security-notes)

---

## 1. What This System Does

This system implements an enterprise workflow for applying for a **Trade License**.
Three distinct roles interact with each application as it moves through the process:

| Role | Responsibility |
|------|---------------|
| **Customer** | Fills out and submits the application (license type, commodity, documents, payment) |
| **Reviewer** | Inspects the submission and either accepts it, rejects it, or asks for adjustments |
| **Approver** | Grants the final approval, rejects the application, or sends it back for re-review |

The workflow is driven by a **state machine** — each role can only perform actions
that are valid for the application's current status. Illegal transitions are refused
by the domain aggregate, not by if-statements scattered across handler code.

---

## 2. Architecture Fundamentals

### 2.1 What is Clean Architecture?

Clean Architecture (Robert C. Martin, 2012) organizes code into concentric rings.
The innermost ring contains the most stable, most important code — the business rules.
Each outer ring can depend on inner rings, but **inner rings must never depend on
outer rings**.

```
┌──────────────────────────────────────────────────────────┐
│  Presentation  (HTTP handlers, Fiber, JSON, middleware)  │
│  ┌──────────────────────────────────────────────────┐    │
│  │  Application  (command/query handlers, DTOs)     │    │
│  │  ┌────────────────────────────────────────────┐  │    │
│  │  │  Domain  (aggregates, entities, events)    │  │    │
│  │  └────────────────────────────────────────────┘  │    │
│  └──────────────────────────────────────────────────┘    │
│  Infrastructure  (PostgreSQL, GORM — plugged in from     │
│                   outside, implements domain interfaces)  │
└──────────────────────────────────────────────────────────┘
```

**Why does this matter?**

- You can swap PostgreSQL for MySQL without touching a single domain or application file.
- You can replace Fiber with standard `net/http` without touching business logic.
- You can test every business rule without a running database.
- New developers can read `domain/` to understand what the system does, without
  having to parse HTTP routing or SQL queries.

### 2.2 What is Domain-Driven Design?

DDD (Eric Evans, 2003) is a methodology for building software that reflects the
real-world business problem it solves. Key concepts used in this project:

**Aggregate Root** — A cluster of related objects treated as a single unit of
consistency. All changes to the cluster go through the root. In this system,
`TradeLicenseApplication` is the aggregate root. It owns `Commodity`, `Document[]`,
`Payment`, and `HistoryEntry[]`. Nothing from outside the aggregate modifies these
directly — all mutations go through methods on `TradeLicenseApplication`.

**Entity** — An object with a persistent identity that can change over time.
`Commodity`, `Document`, `Payment`, and `HistoryEntry` are entities — they each
have an ID and are owned by the aggregate.

**Value Object** — An object defined entirely by its value, not by identity.
`ApplicationID` and `LicenseType` are value objects. Two `ApplicationID`s with
the same UUID string are considered equal. They are immutable and validated at
construction time so corrupt values cannot exist in the system.

**Domain Event** — A record of something that happened in the domain, named in
the past tense. `ApplicationSubmittedEvent`, `ApplicationApprovedEvent`, etc.
Events are raised inside aggregate methods and can be consumed by downstream
systems (notifications, audit logs) without tight coupling to the domain.

**Repository** — An abstraction over the persistence mechanism. The domain defines
a `ApplicationRepository` interface; the infrastructure layer provides the actual
PostgreSQL implementation. The domain never imports the infrastructure package.

**Bounded Context** — A logical boundary within which a specific domain model
applies. This system's bounded context is `tradelivense` — all domain types live
inside this package.

### 2.3 How They Work Together

DDD tells you *what* to model and *where* the business rules live.
Clean Architecture tells you *how* to organize the layers so those rules stay
protected from technical details.

Together, the result is:

- Business rules live in `src/domain/tradelivense/aggregate.go`. They are pure Go.
  No imports from Fiber, GORM, or any external library.
- Orchestration (loading, saving, calling domain methods) lives in `src/application/`.
- HTTP translation lives in `src/presentation/`.
- Database details live in `src/infrastructure/`.

### 2.4 The Dependency Rule

Every Go import in this project follows one rule: **imports only point inward**.

```
presentation → application → domain
infrastructure ──────────→ domain
main.go → everything (it is the only exception — it assembles the pieces)
```

The domain imports **nothing** from this project. It only uses the Go standard
library and `github.com/google/uuid`.

You can verify this by running:
```bash
grep -r "enterprise/trade-license" src/domain/
```
The only results will be the package declaration lines — no cross-layer imports.

### 2.5 What is CQRS?

CQRS (Command Query Responsibility Segregation) means the write operations
(commands) and read operations (queries) are handled by separate code paths.

**Commands** mutate state: `Submit`, `Cancel`, `Accept`, `Approve`, etc.
They accept a command struct, perform a domain operation, and persist the change.
They return an error or a minimal response (e.g., the new application's ID).

**Queries** read state: `GetApplication`, `ListByApplicant`, `ListByStatus`.
They never mutate anything. They read from the repository and return a DTO
(Data Transfer Object) shaped for the API consumer.

This separation makes the system easier to reason about — if a test or reader
knows they are looking at a query handler, they know with certainty that no data
will be changed.

---

## 3. Project Structure

```
UO4304/
├── main.go                          # Composition root — wires all layers together
├── go.mod                           # Go module definition
├── docker-compose.yml               # Runs postgres, app, frontend, seed
├── Dockerfile                       # Backend container
├── Dockerfile.frontend              # Frontend container
├── Dockerfile.seed                  # Seed data container
│
├── src/
│   ├── config/
│   │   └── config.go                # Reads environment variables into Config struct
│   │
│   ├── domain/
│   │   ├── common/
│   │   │   ├── aggregate_root.go    # Base type for aggregate roots (event management)
│   │   │   └── domain_event.go      # DomainEvent interface + BaseDomainEvent
│   │   └── tradelivense/            # The Trade License bounded context
│   │       ├── aggregate.go         # TradeLicenseApplication — all business rules
│   │       ├── entities.go          # Commodity, Document, Payment, HistoryEntry
│   │       ├── value_objects.go     # ApplicationID, LicenseType, ApplicationStatus
│   │       ├── events.go            # Domain events (Submitted, Accepted, etc.)
│   │       ├── repository.go        # ApplicationRepository interface (the port)
│   │       └── errors.go            # Domain error sentinels
│   │
│   ├── application/
│   │   ├── command/                 # Write-side use cases
│   │   │   ├── submit_application.go
│   │   │   ├── cancel_application.go
│   │   │   ├── update_application.go
│   │   │   ├── resubmit_application.go
│   │   │   ├── delete_application.go
│   │   │   ├── review_application.go
│   │   │   └── approve_application.go
│   │   └── query/                   # Read-side use cases
│   │       └── get_application.go   # GetApplication, ListByStatus, ListByApplicant + DTOs
│   │
│   ├── infrastructure/
│   │   └── persistence/
│   │       └── postgres/
│   │           ├── db.go            # Opens DB connection and runs AutoMigrate
│   │           ├── mapper.go        # Converts between domain objects ↔ GORM models
│   │           ├── repository.go    # ApplicationRepository implementation (GORM)
│   │           └── models/
│   │               └── models.go    # GORM-annotated table structs
│   │
│   ├── presentation/
│   │   └── http/
│   │       ├── router.go            # Registers all routes and middleware chains
│   │       ├── handler/
│   │       │   ├── customer_handler.go   # CUSTOMER role endpoints
│   │       │   ├── reviewer_handler.go   # REVIEWER role endpoints
│   │       │   ├── approver_handler.go   # APPROVER role endpoints
│   │       │   └── errors.go             # Domain error → HTTP status code mapper
│   │       └── middleware/
│   │           └── role.go          # RequireRole middleware + UserID extractor
│   │
│   └── testutil/
│       └── mock_repository.go       # In-memory repository for unit tests
│
└── frontend/                        # Next.js 16 App Router frontend
    ├── app/
    │   ├── layout.tsx               # Root layout with navbar and identity context
    │   ├── page.tsx                 # Login / quick-select page
    │   ├── test-flow/page.tsx       # Guided multi-scenario test flow
    │   ├── customer/
    │   │   └── applications/
    │   │       ├── page.tsx         # Customer application list (tabbed)
    │   │       ├── new/page.tsx     # Submit new application form
    │   │       └── [id]/
    │   │           ├── page.tsx     # Application detail + action buttons
    │   │           └── edit/page.tsx# Edit / resubmit form
    │   ├── reviewer/
    │   │   ├── queue/page.tsx       # Reviewer work queue
    │   │   └── applications/[id]/page.tsx
    │   └── approver/
    │       ├── queue/page.tsx       # Approver work queue
    │       └── applications/[id]/page.tsx
    ├── components/
    │   ├── AppCard.tsx              # Application summary card (list view)
    │   ├── AppDetail.tsx            # Full application detail + audit trail
    │   ├── StatusBadge.tsx          # Colored status pill
    │   ├── WorkflowTimeline.tsx     # Step progress indicator
    │   ├── ActionModal.tsx          # Reviewer/approver action dialog
    │   ├── NavBar.tsx               # Top navigation bar
    │   └── ui/                      # shadcn/ui primitive components
    ├── lib/
    │   ├── api.ts                   # Typed API client for all backend endpoints
    │   ├── types.ts                 # TypeScript type definitions mirroring backend DTOs
    │   └── identity.ts              # Identity (userId + role) session helpers
    └── contexts/
        └── IdentityContext.tsx      # React context for the current user session
```

---

## 4. The Domain Layer

**Location:** `src/domain/`

The domain layer is the heart of the system. It contains all business rules and
has **zero dependencies** on any external library, framework, or database. If you
read only this layer, you understand exactly what the business does.

### 4.1 Bounded Context

**File:** `src/domain/tradelivense/`

The package name `tradelivense` is a deliberately coined term for the
**Trade License** bounded context. Bounded contexts prevent naming collisions
between different parts of a large enterprise system — in a multi-service
architecture you might have a `billing` context, an `identity` context, and a
`tradelivense` context, each with its own definition of "application" or "user".

Everything that defines what a Trade License application *is* lives inside this package.

### 4.2 Aggregate Root

**File:** `src/domain/tradelivense/aggregate.go`

`TradeLicenseApplication` is the aggregate root. It is the central object in the
system — all other domain objects (`Commodity`, `Document`, `Payment`, `HistoryEntry`)
are owned by and accessed through it.

**Why aggregate roots matter:**

Without an aggregate root, any part of the codebase could reach into a `Payment`
struct and change the amount without the `Application` knowing. Business rules
would be scattered — some in handlers, some in services, some nowhere. Bugs creep
in because the rules are not enforced consistently.

With an aggregate root, **every state change happens through a method on the aggregate**.
The methods enforce the rules:

```
Submit()         — can only be called when status is PENDING, documents are attached,
                   and payment is settled. Transitions PENDING → SUBMITTED.

Cancel()         — can only be called from PENDING or ADJUSTED. Transitions → CANCELLED.

UpdateDetails()  — can only be called from PENDING or ADJUSTED. Updates commodity + docs.

ReplacePayment() — can only be called from PENDING or ADJUSTED. Updates payment in-place.

Resubmit()       — can only be called from ADJUSTED. Archives reviewer notes, clears them,
                   transitions ADJUSTED → SUBMITTED.

Delete()         — can only be called from PENDING, CANCELLED, or REJECTED.

Accept()         — REVIEWER action. SUBMITTED|REREVIEW → ACCEPTED.
ReviewReject()   — REVIEWER action. SUBMITTED|REREVIEW → REJECTED.
Adjust()         — REVIEWER action. SUBMITTED|REREVIEW → ADJUSTED.

Approve()        — APPROVER action. ACCEPTED → APPROVED.
ApproveReject()  — APPROVER action. ACCEPTED → REJECTED.
Rereview()       — APPROVER action. ACCEPTED → REREVIEW.
```

Every method that transitions status also appends a `HistoryEntry` via the
private `addHistory()` helper, creating a complete append-only audit trail.

### 4.3 Entities

**File:** `src/domain/tradelivense/entities.go`

Entities are domain objects that have identity (an ID) and may change over time.

**`Commodity`**
Represents the specific trade activity or product the customer is applying to
operate. Contains name, description, and category. Has a UUID ID generated at
construction. An application has exactly one commodity.

**`Document`**
Represents one supporting file (e.g. passport copy, business registration) that
the customer attaches. Stores a reference URL (not the file itself — the domain
does not manage bytes). An application can have many documents.

**`Payment`**
Records the fee settlement. Contains amount, currency, external transaction ID
(from the payment gateway), settled timestamp, and status. An application has at
most one payment.

**`HistoryEntry`**
Records a single status transition in the audit trail. Immutable once created.
Fields: who acted (`ActorID`), what action was taken, previous status (`FromStatus`),
new status (`ToStatus`), any notes, and the timestamp. History is append-only —
entries are never modified or deleted.

### 4.4 Value Objects

**File:** `src/domain/tradelivense/value_objects.go`

Value objects are defined entirely by their value. Two instances with the same
data are considered identical. They are immutable — once created, they cannot change.

**`ApplicationID`** — Wraps a UUID string. The compiler enforces that you cannot
accidentally pass a plain string where an ID is expected. `ApplicationIDFrom()`
validates the format and returns an error for non-UUID strings — corrupt IDs
cannot enter the domain.

**`LicenseType`** — Wraps the license type string with an allowlist (`validLicenseTypes`).
Currently only `"TRADE_LICENSE"` is valid. Adding a new license type means adding
one entry to the map — no other code changes.

**`ApplicationStatus`** — Typed string constant for each workflow state.
(`PENDING`, `SUBMITTED`, `CANCELLED`, `ACCEPTED`, `REJECTED`, `ADJUSTED`,
`APPROVED`, `REREVIEW`). Using a type prevents raw string comparisons like
`if status == "submited"` — the compiler catches typos.

### 4.5 Domain Events

**File:** `src/domain/tradelivense/events.go`

Domain events record facts about things that happened. They are raised inside
aggregate methods and collected by `AggregateRoot.AddEvent()`. After the
repository persists the aggregate, the application layer calls `PullEvents()`
to dispatch them.

| Event | Raised When |
|-------|-------------|
| `ApplicationSubmittedEvent` | Customer submits (PENDING → SUBMITTED) |
| `ApplicationCancelledEvent` | Customer cancels |
| `ApplicationAcceptedEvent` | Reviewer accepts |
| `ApplicationRejectedEvent` | Reviewer or approver rejects |
| `ApplicationAdjustedEvent` | Reviewer requests adjustment |
| `ApplicationApprovedEvent` | Approver grants final approval |
| `ApplicationRereviewEvent` | Approver sends back for re-review |
| `ApplicationResubmittedEvent` | Customer resubmits after adjustment |

Each event carries the minimum data consumers need to act. For example,
`ApplicationApprovedEvent` carries `ApplicationID` and `ApproverID` — enough for
a notification service to send a congratulations email without querying the database.

### 4.6 Repository Interface (Port)

**File:** `src/domain/tradelivense/repository.go`

```go
type ApplicationRepository interface {
    Save(ctx, app)          error
    Update(ctx, app)        error
    FindByID(ctx, id)       (*TradeLicenseApplication, error)
    FindByApplicantID(ctx)  ([]*TradeLicenseApplication, error)
    FindByStatus(ctx)       ([]*TradeLicenseApplication, error)
    Delete(ctx, id)         error
}
```

This interface is the **Port** in the Ports & Adapters (Hexagonal Architecture)
pattern. The domain defines what it needs. The infrastructure layer provides the
Adapter (the PostgreSQL implementation). The domain never imports the infrastructure.

This inversion is what makes the system testable without a database — test code
can supply a `MockRepository` that satisfies this interface.

### 4.7 Error Sentinels

**File:** `src/domain/tradelivense/errors.go`

Domain errors are package-level `var` declarations using `errors.New()`. This
pattern is called "error sentinels".

```go
var (
    ErrInvalidStatusTransition = errors.New("...")
    ErrDocumentRequired        = errors.New("...")
    ErrPaymentRequired         = errors.New("...")
    ErrApplicationNotFound     = errors.New("...")
    ErrForbidden               = errors.New("...")
)
```

Callers use `errors.Is()` to check the type — this works correctly even when
errors are wrapped with `fmt.Errorf("...: %w", err)`. The presentation layer's
`domainError()` function maps these to HTTP status codes:

| Domain Error | HTTP Status |
|---|---|
| `ErrApplicationNotFound` | 404 Not Found |
| `ErrForbidden` | 403 Forbidden |
| `ErrInvalidStatusTransition` | 422 Unprocessable Entity |
| `ErrDocumentRequired` | 422 Unprocessable Entity |
| `ErrPaymentRequired` | 422 Unprocessable Entity |

The domain never knows about HTTP codes. The presentation layer never knows about
PostgreSQL. The mapping stays in the presentation layer where it belongs.

### 4.8 Common Package

**Files:** `src/domain/common/aggregate_root.go`, `src/domain/common/domain_event.go`

`AggregateRoot` is the base struct embedded by every aggregate root. It manages
the internal slice of domain events. Two methods:

- `AddEvent(event)` — called inside aggregate methods to record what happened.
- `PullEvents()` — called by the application layer after persistence to drain and
  dispatch the events. Clears the slice after returning, so events are dispatched
  exactly once.

`DomainEvent` is the interface all events satisfy. `BaseDomainEvent` provides the
`OccurredAt()` timestamp implementation so concrete events don't repeat it.

---

## 5. The Application Layer

**Location:** `src/application/`

The application layer orchestrates use cases. It loads aggregates from the
repository, calls domain methods, and persists the result. It does **not** contain
business rules — it delegates to the domain. It does **not** know about HTTP — it
delegates to the presentation layer.

Think of it as a use-case script: "to submit an application, do these steps in this order."

### 5.1 Command Handlers (Write Side)

**Location:** `src/application/command/`

Each file is one use case. The structure is always the same:

1. A **Command struct** — plain data carrying the caller's intent (e.g. `ApplicationID`, `ApplicantID`, fields to update).
2. A **Handler struct** — holds the repository as its only dependency.
3. A **Handle(ctx, cmd) error** method — the use-case orchestration.

---

**`submit_application.go`** — `SubmitApplicationHandler`

Implements: "Applicant requests new application for Trade License."

Steps performed:
1. Validate and construct `LicenseType` value object.
2. Construct a new `TradeLicenseApplication` (starts in PENDING).
3. `SelectCommodity()` — Step 1 of the use case.
4. `AttachDocument()` for each document — Step 2.
5. `SettlePayment()` — Step 3.
6. `Submit()` — Step 4. The domain enforces: documents present, payment settled, status PENDING.
7. `repo.Save()`.

Also defines the shared input structs (`CommodityInput`, `DocumentInput`,
`PaymentInput`) with JSON tags that ensure correct decoding regardless of Go's
field name capitalization rules.

---

**`cancel_application.go`** — `CancelApplicationHandler`

Loads the aggregate, checks ownership, calls `Cancel()` (domain enforces PENDING
or ADJUSTED status), persists. Step 4 of the Customer use case: Cancel action.

---

**`update_application.go`** — `UpdateApplicationHandler`

Allows editing a PENDING or ADJUSTED application.

- Updates commodity and documents via `UpdateDetails()` (domain enforces status).
- Optionally updates payment via `ReplacePayment()` if the caller provided new
  payment data. The payment pointer is nil when the caller omits it — meaning
  "keep existing."
- `ReplacePayment` mutates the existing `Payment` struct in place to preserve
  the database primary key, which makes GORM issue an UPDATE rather than an INSERT.

---

**`resubmit_application.go`** — `ResubmitApplicationHandler`

Closes the ADJUSTED cycle. Updates commodity, documents, and optionally payment,
then calls `Resubmit()` which archives the reviewer's notes into History, clears
the Notes field, and transitions ADJUSTED → SUBMITTED.

---

**`delete_application.go`** — `DeleteApplicationHandler`

Loads, checks ownership, calls `Delete()` (domain enforces PENDING/CANCELLED/REJECTED),
then calls `repo.Delete()` for the soft delete. Active applications cannot be deleted.

---

**`review_application.go`** — `ReviewApplicationHandler`

Implements: "Review Submitted New Application for Trade License."

Dispatches to:
- `Accept()` — SUBMITTED|REREVIEW → ACCEPTED
- `ReviewReject()` — SUBMITTED|REREVIEW → REJECTED (notes required)
- `Adjust()` — SUBMITTED|REREVIEW → ADJUSTED (notes required, returned to customer)

---

**`approve_application.go`** — `ApproveApplicationHandler`

Implements: "Approve Reviewed New Application for Trade License."

Dispatches to:
- `Approve()` — ACCEPTED → APPROVED (workflow complete)
- `ApproveReject()` — ACCEPTED → REJECTED (notes required)
- `Rereview()` — ACCEPTED → REREVIEW (notes required, returned to reviewer)

---

### 5.2 Query Handlers (Read Side)

**File:** `src/application/query/get_application.go`

**DTOs (Data Transfer Objects)** are flat, JSON-serialisable structs designed for
API consumers. They are different from domain aggregates — they do not enforce
rules, they just carry data across the API boundary.

```
ApplicationDTO
├── id, license_type, applicant_id, status, notes
├── commodity: CommodityDTO (nullable)
├── documents: DocumentDTO[]
├── payment: PaymentDTO (nullable)
├── history: HistoryEntryDTO[]   ← full audit trail
├── created_at, updated_at
```

Three handlers:

**`GetApplicationHandler`** — loads a single application by ID and maps it to a DTO.
Used by all three roles when viewing application details.

**`ListByStatusHandler`** — returns all applications with a given status. Used by
reviewers (SUBMITTED + REREVIEW queue) and approvers (ACCEPTED queue).

**`ListByApplicantHandler`** — returns all applications belonging to a specific
customer, ordered newest-first.

The `toDTO()` private function performs the mapping from domain aggregate to DTO.
This mapping boundary is intentional — the API contract can evolve independently
of the domain model.

---

## 6. The Infrastructure Layer

**Location:** `src/infrastructure/`

The infrastructure layer contains everything that talks to the outside world:
databases, file systems, external APIs. In this project, that means PostgreSQL
via GORM.

This layer **implements** the domain's `ApplicationRepository` interface. The domain
knows nothing about GORM or PostgreSQL — it only knows the interface. The
infrastructure package is never imported by domain or application code; only
`main.go` imports it to wire everything together.

### 6.1 GORM Models

**File:** `src/infrastructure/persistence/postgres/models/models.go`

GORM models are plain Go structs with GORM tags that define the database schema.
They are kept strictly inside the infrastructure layer — the domain never sees them.

```
Application        → applications table (soft delete via DeletedAt gorm.DeletedAt)
ApplicationHistory → application_history table (append-only, no update/delete)
Commodity          → commodities table (one-to-one with application)
Document           → documents table (one-to-many with application)
Payment            → payments table (one-to-one; TransactionID has unique index)
```

**Important constraints:**

- `Application.DeletedAt` uses GORM's `gorm.DeletedAt` — this provides soft delete.
  When `repo.Delete()` is called, GORM sets `deleted_at` to the current time.
  All subsequent queries automatically filter out rows where `deleted_at IS NOT NULL`.
  The data is never actually removed from the database.

- `Payment.TransactionID` has `gorm:"uniqueIndex"` — this prevents duplicate payment
  records across applications. The external transaction ID from the payment gateway
  must be unique.

- All models use `gorm:"primaryKey;type:uuid"` — UUIDs as primary keys prevent
  enumeration attacks (an attacker cannot guess the next sequential integer ID).

### 6.2 Database Initialization

**File:** `src/infrastructure/persistence/postgres/db.go`

`NewDB()` opens the database connection using GORM and runs `AutoMigrate`.

AutoMigrate creates tables if they do not exist and adds any new columns that
appear in the model structs. It **never drops columns**, so it is safe to run
against an existing database — existing data is preserved.

In a production environment with strict change control, you would replace
AutoMigrate with versioned SQL migration files (e.g. using `golang-migrate`).
AutoMigrate is retained here because it keeps the development loop fast.

### 6.3 Mapper

**File:** `src/infrastructure/persistence/postgres/mapper.go`

The mapper is the translation layer between the domain world and the persistence world.
It contains two private functions:

**`toModel(app *TradeLicenseApplication) *models.Application`**

Converts a domain aggregate to a GORM model for saving. Copies every field,
flattens associations into their model equivalents, and serializes enums to strings.

**`toDomain(m *models.Application) (*TradeLicenseApplication, error)`**

Converts a GORM model back to a domain aggregate after loading from the database.
Reconstructs value objects (`ApplicationID`, `LicenseType`) from their string
representations, and restores entity identities (IDs) that were originally assigned
by the domain.

Why is this separation necessary? The domain uses strong types (`ApplicationID`,
`LicenseType`, `ApplicationStatus`) but the database stores plain strings. The mapper
translates between these two representations so neither layer has to know about the other.

### 6.4 Repository Implementation

**File:** `src/infrastructure/persistence/postgres/repository.go`

`applicationRepository` implements `tradelivense.ApplicationRepository` using GORM.

**`Save(ctx, app)`** — Creates a new application row and all associations in a single
GORM `Create` call.

**`Update(ctx, app)`** — Persists changes to an existing application inside a
database transaction:
1. `tx.Save(m)` — updates the root record.
2. Deletes all document rows for this application, then re-inserts them (full replace).
   This avoids the complexity of diffing which documents were added/removed.
3. `tx.Save(m.Commodity)` — updates or inserts the commodity.
4. `tx.Save(m.Payment)` — updates or inserts the payment. Because `ReplacePayment`
   preserves the payment's primary key when updating, GORM issues an UPDATE here,
   not an INSERT, avoiding a unique-constraint violation.
5. History is inserted with `ON CONFLICT DO NOTHING` — existing history rows are
   never modified, ensuring the audit trail is truly append-only.

**`FindByID(ctx, id)`** — Loads the application with all associations preloaded.
History is ordered ascending by `occurred_at` so the audit trail reads chronologically.
Returns `ErrApplicationNotFound` for missing records.

**`FindByApplicantID(ctx, applicantID)`** — Returns all non-deleted applications
for a customer, ordered `created_at DESC` (newest first for the list view).

**`FindByStatus(ctx, status)`** — Returns all applications with a specific status,
ordered `updated_at ASC` (oldest first) so the review queues process applications
in arrival order and prevent starvation.

**`Delete(ctx, id)`** — GORM soft-delete: sets `deleted_at = NOW()` on the row.

---

## 7. The Presentation Layer

**Location:** `src/presentation/`

The presentation layer translates between the HTTP world and the application layer.
It knows about:
- Fiber (the HTTP framework) and how to parse requests / write responses.
- HTTP status codes and JSON serialization.
- Input validation at the HTTP boundary (missing fields, invalid formats).

It does **not** know about:
- Business rules (those are in the domain).
- SQL or database details (those are in the infrastructure).

### 7.1 Middleware

**File:** `src/presentation/http/middleware/role.go`

**`RequireRole(role string) fiber.Handler`**

A Fiber middleware factory that returns a handler. The handler checks the `X-Role`
header against the required role and returns 403 if they don't match.

In production, this header would be derived from a validated JWT token rather than
trusted from the client directly. The middleware is kept thin so swapping to JWT
requires changing only this file.

**`UserID(c *fiber.Ctx) string`**

Extracts the `X-User-ID` header. Returns empty string if absent. Callers check
for empty string and return 400. Same note applies — production would verify a JWT.

### 7.2 HTTP Handlers

**Files:** `src/presentation/http/handler/`

**`errors.go`** — `domainError(c, err) error`

Centralizes the mapping from domain errors to HTTP status codes. All handlers
call this instead of checking errors individually. The switch uses `errors.Is()`
which correctly handles wrapped errors.

**`customer_handler.go`** — `CustomerHandler`

All endpoints accessible to the `CUSTOMER` role:

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/customer/applications` | Submit new application (Steps 1–4) |
| `GET` | `/api/customer/applications` | List all my applications |
| `GET` | `/api/customer/applications/:id` | View a single application |
| `PUT` | `/api/customer/applications/:id` | Edit PENDING or ADJUSTED (Steps 1–3) |
| `POST` | `/api/customer/applications/:id/resubmit` | Resubmit ADJUSTED → SUBMITTED |
| `POST` | `/api/customer/applications/:id/cancel` | Cancel PENDING or ADJUSTED |
| `DELETE` | `/api/customer/applications/:id` | Soft-delete PENDING/CANCELLED/REJECTED |

For `PUT` and `POST .../resubmit`, the `payment` field in the request body is
optional. If omitted, the existing payment is preserved. If present, the payment
is re-settled via `ReplacePayment()`.

**`reviewer_handler.go`** — `ReviewerHandler`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/reviewer/applications` | List SUBMITTED + REREVIEW queue |
| `GET` | `/api/reviewer/applications/:id` | View application details |
| `POST` | `/api/reviewer/applications/:id/action` | Take action: ACCEPT / REJECT / ADJUST |

**`approver_handler.go`** — `ApproverHandler`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/approver/applications` | List ACCEPTED queue |
| `GET` | `/api/approver/applications/:id` | View application details |
| `POST` | `/api/approver/applications/:id/action` | Take action: APPROVE / REJECT / REREVIEW |

### 7.3 Router

**File:** `src/presentation/http/router.go`

`NewRouter()` builds the Fiber app with:
- `recover.New()` — prevents a runtime panic from crashing the process. Returns 500 instead.
- `logger.New()` — logs every request and response for observability.
- `GET /health` — no-auth health check for container probes.
- Three role-scoped route groups, each protected by `RequireRole` middleware.

---

## 8. The Composition Root

**File:** `main.go`

The composition root is the only file that imports from every layer. Its job is to
wire everything together, not to contain any logic. Reading `main.go` tells you
exactly which implementations are in use and how they are connected.

The wiring order follows the Dependency Rule from inside out:

1. Load config from environment.
2. Construct the infrastructure adapter (database connection + GORM repository).
3. Construct all application-layer command handlers (inject repository).
4. Construct all application-layer query handlers (inject repository).
5. Construct all presentation-layer HTTP handlers (inject application handlers).
6. Build the router (inject HTTP handlers).
7. Start the server.

---

## 9. The Frontend

**Location:** `frontend/`

### 9.1 Technology Stack

| Technology | Version | Purpose |
|---|---|---|
| **Next.js** | 16 (App Router) | React framework with server-side rendering |
| **TypeScript** | 5 | Type safety across all frontend code |
| **Tailwind CSS** | v4 | Utility-first styling (`@theme inline` syntax) |
| **shadcn/ui** | v3 | Accessible UI component primitives |
| **@base-ui/react** | latest | Underlying primitive library used by shadcn v3 |

**Important: This is Next.js App Router.** All pages under `app/` are Server
Components by default. Pages that use `useState`, `useEffect`, or browser APIs
(like `crypto.randomUUID()`) must have `"use client"` at the top of the file.

**Tailwind v4 syntax change:** The gradient utility is `bg-linear-to-br` (not
`bg-gradient-to-br`). Color tokens use `oklch()` values in `@theme inline` blocks
inside `globals.css`.

### 9.2 Directory Structure

```
frontend/
├── app/                 # Next.js App Router pages
├── components/          # Shared UI components
├── lib/                 # API client, type definitions, utilities
├── contexts/            # React context providers
└── public/              # Static assets
```

### 9.3 Pages

**`app/layout.tsx`**

Root layout. Wraps every page with `IdentityProvider` (makes the current user
available globally via React context) and `NavBar`. Sets `<html lang="en">` and
base body styles.

**`app/page.tsx`** — Login / Quick-select

Two-column layout:
- Left: Manual sign-in form (enter any user ID, pick a role, click Enter Portal).
- Right: Seed credentials panel with one-click login for all 10 pre-seeded users.

Identity is stored in `sessionStorage` so it persists across page navigations
within the tab but resets on close. No passwords — this is a demo system.

**`app/test-flow/page.tsx`** — Guided Test Flow

Scenario selector with 6 pre-built flows:

| Scenario | Steps |
|---|---|
| Happy Path | Submit → Accept → Approve |
| Adjust + Resubmit | Submit → Adjust → Customer Resubmit → Accept → Approve |
| Reviewer Rejects | Submit → Reject |
| Approver Rejects | Submit → Accept → Reject |
| REREVIEW Cycle | Submit → Accept → Rereview → Accept → Approve |
| Customer Cancels | Submit → Cancel |

Each scenario shows a step progress bar. The active step renders the appropriate
form or action panel. Uses the pre-seeded identities (`customer-seed-001`,
`reviewer-seed-001`, `approver-seed-001`) automatically.

**`app/customer/applications/page.tsx`** — Customer Application List

Tabbed view: **Active** (PENDING/SUBMITTED/ACCEPTED/ADJUSTED/REREVIEW),
**Completed** (APPROVED/REJECTED/CANCELLED), **All**.

Shows an orange attention banner at the top when any application is in ADJUSTED
status (the customer needs to take action).

**`app/customer/applications/new/page.tsx`** — Submit New Application

Form matching the use-case steps:
- Step 1 card: commodity name, category, description.
- Step 2 card: document name, URL (validated for http/https), content type.
- Step 3 card: payment amount, currency, transaction ID (auto-generated UUID, regeneratable).

**`app/customer/applications/[id]/page.tsx`** — Application Detail

Renders `AppDetail` (full application data + audit trail history). Shows action
buttons based on the current status:

| Status | Available Actions |
|---|---|
| PENDING | Edit Details, Cancel |
| ADJUSTED | Edit Details, Resubmit, Cancel |
| CANCELLED, REJECTED | Delete |

When ADJUSTED, the reviewer's notes are shown in a prominent orange warning box.

**`app/customer/applications/[id]/edit/page.tsx`** — Edit / Resubmit

Pre-fills all three steps from the current application data. Payment section has
a toggle: "Update payment" is off by default (shows current payment as read-only).
Toggling it on reveals editable amount/currency/transaction ID fields.

When accessed with `?resubmit=1`, the form submits to the `POST .../resubmit`
endpoint instead of `PUT .../:id`, and the button label changes to "Submit for Review."
The reviewer's notes (the reason for adjustment) are shown prominently at the top.

**`app/reviewer/queue/page.tsx`** — Reviewer Queue

Lists all SUBMITTED and REREVIEW applications. Each application is a card linking
to the detail page. REREVIEW applications are marked distinctly so the reviewer
knows they came from an approver re-review request.

**`app/reviewer/applications/[id]/page.tsx`** — Reviewer Detail

Shows all three reviewer inspection steps (commodity, documents, payment) and the
action modal for Step 4 (Accept / Reject / Adjust). Notes field is required for
Reject and Adjust.

**`app/approver/queue/page.tsx`** — Approver Queue

Lists all ACCEPTED applications awaiting final decision.

**`app/approver/applications/[id]/page.tsx`** — Approver Detail

Shows the reviewed application and the action modal for Step 4 (Approve / Reject / Rereview).

### 9.4 Components

**`components/AppCard.tsx`** — Application summary card for list views.
Shows license type, commodity name, applicant ID, status badge, and creation date.
ADJUSTED applications get an orange border and "⚠ Action required" label.

**`components/AppDetail.tsx`** — Full application detail view.
Sections: header (ID, license type, applicant, status badge), workflow timeline,
reviewer notes (when present), commodity, documents (with safe link rendering —
only http/https URLs are rendered as `<a>` tags), payment, audit trail history,
timestamps.

The **Audit Trail** section renders a vertical timeline of every `HistoryEntry`,
showing who acted, what action was taken, what the status transition was,
the timestamp, and any notes.

**`components/StatusBadge.tsx`** — Colored pill for application status.
Each status has a distinct color scheme so visual scanning is fast.

**`components/WorkflowTimeline.tsx`** — Horizontal step progress indicator.
Shows the four workflow stages (Submit, Review, Approve, Done) with color-coded
completion state based on the current application status.

**`components/ActionModal.tsx`** — Dialog for reviewer/approver actions.
Shows available actions as buttons, a notes textarea (required for non-approval
actions), and a confirm button.

**`components/NavBar.tsx`** — Top navigation bar.
Shows the current user ID, role, and links relevant to the current role.
Includes a sign-out button that clears the session.

**`components/ui/`** — shadcn/ui primitives: `Button`, `Card`, `Input`, `Label`,
`Textarea`, `Badge`, `Dialog`, `Select`, `Separator`, `Skeleton`.
These are copied and owned by the project (not imported from a CDN) so you can
customize them freely.

### 9.5 API Client

**File:** `frontend/lib/api.ts`

Typed wrapper around `fetch`. The base `request<T>()` function handles:
- Attaching `Content-Type: application/json`, `X-User-ID`, and `X-Role` headers.
- Treating HTTP 204 as a successful empty response.
- Throwing `ApiResponseError` (with status code and body text) for non-OK responses.

Three API namespaces:

**`customerApi`**
- `submit(identity, payload)` — POST + GET (backend returns only the ID, so a
  second GET is made to return the full DTO).
- `list(identity)`, `get(identity, id)`
- `update(identity, id, payload)` — PUT with optional `payment` field.
- `resubmit(identity, id, payload)` — POST to `.../resubmit`.
- `cancel(identity, id)` — POST to `.../cancel`.
- `delete(identity, id)` — DELETE.

**`reviewerApi`** — `list`, `get`, `takeAction`

**`approverApi`** — `list`, `get`, `takeAction`

### 9.6 Type Definitions

**File:** `frontend/lib/types.ts`

TypeScript interfaces that mirror the Go DTOs. These ensure the frontend compiler
catches any mismatch between what the API returns and what the UI expects.

Key types:
- `ApplicationStatus` — union type of all valid status strings.
- `ApplicationDTO` — full application with history array.
- `HistoryEntryDTO` — single audit trail entry.
- `Identity` — `{ userId: string; role: Role }`.

---

## 10. Workflow State Machine

### 10.1 Status Diagram

```
                        ┌────────────────────────────────────────┐
                        │               CUSTOMER                 │
                        └────────────────────────────────────────┘
                                         │ Submit
                                         ▼
                         ┌──────────────────────────────┐
 Customer creates ──────► PENDING ──Cancel──► CANCELLED │
                         └──────────────────────────────┘
                                         │ Submit
                                         ▼
                                     SUBMITTED ◄──────────────────┐
                                         │                        │
                        ┌────────────────┼──────────────┐        │
                        │                │               │        │
                     Accept            Reject          Adjust     │  (Rereview sends
                        │                │               │        │   back to reviewer)
                        ▼                ▼               ▼        │
                     ACCEPTED        REJECTED         ADJUSTED    │
                        │                              │    │     │
                    ┌───┴──────────────────┐       Cancel  Resubmit
                    │      APPROVER        │           │        │
                    └───┬──────────────────┘           ▼        └──►SUBMITTED
                        │                          CANCELLED
              ┌─────────┼──────────┐
           Approve    Reject   Rereview
              │         │          │
              ▼         ▼          ▼
           APPROVED  REJECTED   REREVIEW ────────────────────────►(back to SUBMITTED)
```

### 10.2 Role Responsibilities

**Customer (Step 4 actions):**
- `Submit` — PENDING → SUBMITTED
- `Cancel` — PENDING|ADJUSTED → CANCELLED
- `Resubmit` — ADJUSTED → SUBMITTED (after editing)
- `Delete` — soft-delete PENDING|CANCELLED|REJECTED

**Reviewer (Step 4 actions):**
- `Accept` — SUBMITTED|REREVIEW → ACCEPTED
- `Reject` — SUBMITTED|REREVIEW → REJECTED (notes required)
- `Adjust` — SUBMITTED|REREVIEW → ADJUSTED (notes required — returns to customer)

**Approver (Step 4 actions):**
- `Approve` — ACCEPTED → APPROVED (workflow complete)
- `Reject` — ACCEPTED → REJECTED (notes required)
- `Rereview` — ACCEPTED → REREVIEW (notes required — returns to reviewer)

### 10.3 The ADJUSTED Cycle

The ADJUSTED cycle was a gap in a naive implementation and deserves explicit
documentation:

1. Reviewer calls `Adjust()` with notes explaining what needs correction.
   Status: SUBMITTED → ADJUSTED. Notes stored on application.

2. Customer sees "⚠ Action required" on their list. The reviewer's notes
   are prominently displayed on the detail and edit pages.

3. Customer edits commodity, documents, and/or payment (Steps 1–3 of the use case).

4. Customer calls `Resubmit()`. This:
   - Validates documents still present, payment still settled.
   - Appends a RESUBMIT history entry that archives the reviewer's notes.
   - **Clears the `Notes` field** so the next reviewer sees a clean slate.
   - Transitions ADJUSTED → SUBMITTED.

5. Application re-enters the reviewer queue.

The customer can also `Cancel()` from ADJUSTED (they choose to withdraw rather
than correct) or `Delete()` after cancelling.

---

## 11. API Reference

All endpoints require role-appropriate headers:
```
X-User-ID: <user-id>
X-Role: CUSTOMER | REVIEWER | APPROVER
Content-Type: application/json
```

### Health

```
GET /health
→ 200 { "status": "ok" }
```

### Customer Endpoints

**Submit New Application**
```
POST /api/customer/applications
Body:
{
  "license_type": "TRADE_LICENSE",
  "commodity": { "name": "...", "description": "...", "category": "..." },
  "documents": [{ "name": "...", "url": "https://...", "content_type": "..." }],
  "payment": { "amount": 500.00, "currency": "USD", "transaction_id": "TXN-..." }
}
→ 201 { "application_id": "<uuid>" }
→ 400 if license_type/documents/payment missing or malformed
→ 422 if domain rule violated
```

**List My Applications**
```
GET /api/customer/applications
→ 200 ApplicationDTO[]
```

**Get Application**
```
GET /api/customer/applications/:id
→ 200 ApplicationDTO (includes history[])
→ 404 if not found
```

**Edit Application (Steps 1–3)**
```
PUT /api/customer/applications/:id
Body:
{
  "commodity": { "name": "...", "description": "...", "category": "..." },
  "documents": [{ "name": "...", "url": "https://...", "content_type": "..." }],
  "payment": { "amount": 600.00, "currency": "USD", "transaction_id": "TXN-..." }  // optional
}
→ 200 ApplicationDTO
→ 403 if caller is not the applicant
→ 422 if application is not PENDING or ADJUSTED
```

**Resubmit Application (ADJUSTED → SUBMITTED)**
```
POST /api/customer/applications/:id/resubmit
Body: same as PUT above (payment optional)
→ 200 ApplicationDTO
→ 403 if caller is not the applicant
→ 422 if application is not ADJUSTED
```

**Cancel Application**
```
POST /api/customer/applications/:id/cancel
→ 204 No Content
→ 403 if caller is not the applicant
→ 422 if application is not PENDING or ADJUSTED
```

**Delete Application**
```
DELETE /api/customer/applications/:id
→ 204 No Content
→ 403 if caller is not the applicant
→ 422 if application is not PENDING, CANCELLED, or REJECTED
```

### Reviewer Endpoints

**List Work Queue**
```
GET /api/reviewer/applications           → SUBMITTED + REREVIEW combined
GET /api/reviewer/applications?status=SUBMITTED
GET /api/reviewer/applications?status=REREVIEW
→ 200 ApplicationDTO[]
```

**Get Application**
```
GET /api/reviewer/applications/:id
→ 200 ApplicationDTO
```

**Take Action (Step 4)**
```
POST /api/reviewer/applications/:id/action
Body: { "action": "ACCEPT" | "REJECT" | "ADJUST", "notes": "..." }
Notes required for REJECT and ADJUST.
→ 204 No Content
→ 422 if invalid status transition
```

### Approver Endpoints

**List Work Queue**
```
GET /api/approver/applications           → defaults to status=ACCEPTED
GET /api/approver/applications?status=ACCEPTED
→ 200 ApplicationDTO[]
```

**Get Application**
```
GET /api/approver/applications/:id
→ 200 ApplicationDTO
```

**Take Action (Step 4)**
```
POST /api/approver/applications/:id/action
Body: { "action": "APPROVE" | "REJECT" | "REREVIEW", "notes": "..." }
Notes required for REJECT and REREVIEW.
→ 204 No Content
→ 422 if invalid status transition
```

---

## 12. Database Schema

```sql
-- applications: the aggregate root table
CREATE TABLE applications (
    id           UUID PRIMARY KEY,
    license_type VARCHAR NOT NULL,
    applicant_id VARCHAR NOT NULL,
    status       VARCHAR NOT NULL,
    notes        TEXT,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ  -- soft delete: NULL = active, non-NULL = deleted
);
CREATE INDEX ON applications (applicant_id);
CREATE INDEX ON applications (status);
CREATE INDEX ON applications (deleted_at);

-- application_history: append-only audit trail
CREATE TABLE application_history (
    id             UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    actor_id       VARCHAR NOT NULL,
    action         VARCHAR NOT NULL,   -- SUBMIT, CANCEL, ACCEPT, REJECT, ADJUST, ...
    from_status    VARCHAR NOT NULL,
    to_status      VARCHAR NOT NULL,
    notes          TEXT,
    occurred_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX ON application_history (application_id);
CREATE INDEX ON application_history (occurred_at);

-- commodities: one per application
CREATE TABLE commodities (
    id             UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name           VARCHAR NOT NULL,
    description    TEXT,
    category       VARCHAR
);
CREATE INDEX ON commodities (application_id);

-- documents: many per application
CREATE TABLE documents (
    id             UUID PRIMARY KEY,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name           VARCHAR NOT NULL,
    url            VARCHAR NOT NULL,
    content_type   VARCHAR,
    uploaded_at    TIMESTAMPTZ
);
CREATE INDEX ON documents (application_id);

-- payments: one per application (one-to-one)
CREATE TABLE payments (
    id             UUID PRIMARY KEY,
    application_id UUID NOT NULL UNIQUE REFERENCES applications(id) ON DELETE CASCADE,
    amount         FLOAT,
    currency       VARCHAR,
    transaction_id VARCHAR UNIQUE,  -- unique across all applications
    paid_at        TIMESTAMPTZ,
    status         VARCHAR
);
```

**Cascade deletes** on all child tables mean that if an `Application` row is
hard-deleted (not normally done — we soft-delete), all its associated rows are
automatically removed.

**Soft delete**: The `deleted_at` column on `applications` is set by GORM's
soft-delete feature. GORM automatically appends `WHERE deleted_at IS NULL` to
all SELECT queries so deleted applications are invisible to the rest of the system.

---

## 13. Configuration

All configuration is read from environment variables. A `.env.example` file is
provided for local development.

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL username |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `trade_license` | PostgreSQL database name |
| `DB_SSL_MODE` | `disable` | SSL mode (set to `require` in production) |

---

## 14. How to Run

### Prerequisites

- Docker and Docker Compose v2

### Start all services

```bash
docker compose up --build -d
```

This starts three containers:
- `postgres` — PostgreSQL 16
- `app` — Go backend on port 8080
- `frontend` — Next.js frontend on port 3000

The database schema is created automatically on first start via GORM's AutoMigrate.

### Load seed data

The seed container is excluded from the default startup. Run it once manually:

```bash
docker compose run --rm seed
```

This creates 10 pre-seeded users and their applications across all workflow
states so you can immediately explore every part of the UI.

### Access the application

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Health check | http://localhost:8080/health |

### Rebuild after code changes

```bash
# Backend only
docker compose up --build -d app

# Frontend only
docker compose up --build -d frontend

# Both
docker compose up --build -d
```

### Run backend tests

```bash
go test ./...
```

### Run a specific test

```bash
go test ./src/domain/tradelivense/ -v -run TestTradeLicenseApplication
```

---

## 15. Seed Data

After running `docker compose run --rm seed`, the following users exist:

| User ID | Role | Application Status | Notes |
|---|---|---|---|
| `customer-seed-001` | CUSTOMER | PENDING | Used by test flow; can be submitted |
| `customer-seed-002` | CUSTOMER | SUBMITTED | Awaiting reviewer action |
| `customer-seed-003` | CUSTOMER | ACCEPTED | Awaiting approver action |
| `customer-seed-004` | CUSTOMER | APPROVED | Fully approved — workflow complete |
| `customer-seed-005` | CUSTOMER | REJECTED | Rejected by reviewer |
| `customer-seed-006` | CUSTOMER | ADJUSTED | Needs customer correction |
| `customer-seed-007` | CUSTOMER | REJECTED | Rejected by approver |
| `customer-seed-008` | CUSTOMER | REREVIEW | Sent back to reviewer by approver |
| `reviewer-seed-001` | REVIEWER | — | Sees SUBMITTED + REREVIEW queues |
| `approver-seed-001` | APPROVER | — | Sees ACCEPTED queue |

Log in as any of these from the home page with one click.

---

## 16. Testing

### Unit tests (domain + application layers)

Unit tests are fast, require no infrastructure, and test business rules in isolation.
The `testutil.MockRepository` provides an in-memory implementation of
`ApplicationRepository` for use in tests.

Key test files:

- `src/domain/tradelivense/aggregate_test.go` — tests every state transition on
  the aggregate, including happy paths, invalid transition attempts, and domain
  event emissions.

- `src/application/command/submit_application_test.go` — tests the full submission
  use case with the mock repository.

- `src/application/command/review_application_test.go` — tests all reviewer actions.

- `src/application/command/approve_application_test.go` — tests all approver actions.

### Test naming convention

```
Test<Type>_<Method>_<Scenario>
```
Example: `TestTradeLicenseApplication_Submit_RequiresDocuments`

### Integration testing (manual)

Use the **Test Flow** page at http://localhost:3000/test-flow to walk through
complete scenarios end-to-end in the UI. Six scenarios cover the full range of
workflow paths.

For API-level testing, a Postman collection is included:
```
trade-license.postman_collection.json
```

---

## 17. Security Notes

**Authentication is a placeholder.** The `X-User-ID` and `X-Role` headers are
trusted directly from the client. In production:

1. Implement JWT authentication. The `middleware/role.go` file is the only
   place that needs to change — extract user ID and role from a verified JWT claim
   instead of from raw headers.

2. All business logic, ownership checks (`app.ApplicantID != cmd.ApplicantID`),
   and domain error mappings are already in place. Only the header-trust
   assumption in the middleware needs to be hardened.

**Document URL validation.** The frontend validates that document URLs start with
`http://` or `https://` before rendering them as links, preventing `javascript:`
URL injection attacks in the document viewer.

**UUID-based IDs.** All primary keys are UUIDs. This prevents attackers from
enumerating resources by guessing sequential integer IDs.

**Soft deletes.** Deleted applications are never removed from the database. This
provides an audit trail even for deleted records and allows recovery if data
is accidentally deleted.

**SQL injection.** GORM uses parameterized queries throughout. No raw SQL strings
are constructed from user input.

**Input validation.** All user-facing inputs are validated at the presentation
layer boundary before reaching the domain. The domain provides a second layer of
protection via aggregate invariants.
