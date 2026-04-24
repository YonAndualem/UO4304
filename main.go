// main is the composition root — the single place where all layers are wired together.
//
// Clean Architecture mandates that only the outermost entry point is allowed to
// import from every layer. main.go constructs all objects, injects all dependencies,
// and starts the server. It contains zero business logic — it is purely assembly.
//
// Dependency flow (outer → inner, matching the Clean Architecture rings):
//
//	main.go
//	  └─ presentation/http  (HTTP handlers, router, middleware)
//	       └─ application   (command and query handlers)
//	            └─ domain   (aggregates, entities, value objects, events)
//	  └─ infrastructure     (PostgreSQL repository — satisfies domain.ApplicationRepository)
//
// The domain layer has no outward dependencies. Every arrow points inward.
package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"github.com/enterprise/trade-license/src/application/auth"
	"github.com/enterprise/trade-license/src/application/command"
	"github.com/enterprise/trade-license/src/application/query"
	"github.com/enterprise/trade-license/src/config"
	postgresrepo "github.com/enterprise/trade-license/src/infrastructure/persistence/postgres"
	miniostorage "github.com/enterprise/trade-license/src/infrastructure/storage/minio"
	httpserver "github.com/enterprise/trade-license/src/presentation/http"
	"github.com/enterprise/trade-license/src/presentation/http/handler"
)

func main() {
	// Load .env if present. In production, environment variables are injected
	// directly by the container runtime (Docker, Kubernetes) so this is a no-op.
	_ = godotenv.Load()

	cfg := config.Load()

	// ── Infrastructure layer ──────────────────────────────────────────────────
	// Open the PostgreSQL connection and run AutoMigrate to ensure the schema
	// matches the current model definitions. The returned *gorm.DB is the only
	// infrastructure object that crosses layer boundaries.
	db, err := postgresrepo.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// ── Object storage (MinIO) ───────────────────────────────────────────────
	store, err := miniostorage.New(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket, cfg.MinIOUseSSL)
	if err != nil {
		log.Fatalf("failed to connect to MinIO: %v", err)
	}
	if err := store.EnsureBucket(context.Background()); err != nil {
		log.Fatalf("failed to ensure MinIO bucket: %v", err)
	}
	uploadHandler := handler.NewUploadHandler(store)

	// ── Auth ──────────────────────────────────────────────────────────────────
	userRepo    := postgresrepo.NewUserRepository(db)
	authSvc     := auth.NewService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	// The repository adapter satisfies the domain's ApplicationRepository port.
	repo := postgresrepo.NewApplicationRepository(db)

	// ── Application layer: command handlers (write side) ─────────────────────
	// Each handler encapsulates exactly one use case from the requirements document.
	// They depend only on domain types and the repository interface.
	submitHandler   := command.NewSubmitApplicationHandler(repo)   // UC: Submit New Application
	cancelHandler   := command.NewCancelApplicationHandler(repo)   // UC: Cancel Application
	updateHandler   := command.NewUpdateApplicationHandler(repo)   // UC: Edit Application (Steps 1–3)
	resubmitHandler := command.NewResubmitApplicationHandler(repo) // UC: Resubmit Adjusted Application
	deleteHandler   := command.NewDeleteApplicationHandler(repo)   // UC: Delete Application
	reviewHandler   := command.NewReviewApplicationHandler(repo)   // UC: Review Application (Accept/Reject/Adjust)
	approveHandler  := command.NewApproveApplicationHandler(repo)  // UC: Approve Application (Approve/Reject/Rereview)

	// ── Application layer: query handlers (read side) ─────────────────────────
	// Query handlers are read-only — they never mutate state (CQRS read side).
	getHandler      := query.NewGetApplicationHandler(repo)      // Read single application
	listByStatus    := query.NewListByStatusHandler(repo)        // Read work queues (reviewer/approver)
	listByApplicant := query.NewListByApplicantHandler(repo)     // Read customer's own applications

	// ── Presentation layer: HTTP handlers ─────────────────────────────────────
	// HTTP handlers receive injected application-layer handlers. They know about
	// HTTP (Fiber context, JSON, status codes) but nothing about databases.
	customerHandler := handler.NewCustomerHandler(
		submitHandler, cancelHandler, updateHandler, resubmitHandler, deleteHandler,
		getHandler, listByApplicant,
	)
	reviewerHandler := handler.NewReviewerHandler(reviewHandler, getHandler, listByStatus)
	approverHandler := handler.NewApproverHandler(approveHandler, getHandler, listByStatus)

	// Build the Fiber router with all routes registered.
	app := httpserver.NewRouter(authHandler, customerHandler, reviewerHandler, approverHandler, uploadHandler, authSvc)

	log.Printf("server starting on :%s", cfg.ServerPort)
	log.Fatal(app.Listen(":" + cfg.ServerPort))
}
