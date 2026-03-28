package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/router"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/audit"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/cache"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/config"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/database"
	mongorepo "github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/mongo"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/pg"
	authuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/auth"
	beneficiaryuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/beneficiary"
	debtuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
	programuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/program"
	propertyuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"
	tenantuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/tenant"
	transactionuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
	useruc "github.com/ISubamariner/guimba-go/backend/internal/usecase/user"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
	"github.com/ISubamariner/guimba-go/backend/pkg/logger"
)

// @title           Guimba-GO API
// @version         1.0
// @description     Social Protection Management Information System API
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Set up structured logger
	log := logger.New()
	slog.SetDefault(log)

	slog.Info("starting server", "env", cfg.App.Env, "port", cfg.App.Port)

	ctx := context.Background()

	// Connect to PostgreSQL
	pgPool, err := database.NewPostgresPool(ctx, cfg.Postgres.DSN)
	if err != nil {
		slog.Error("failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer database.ClosePostgres(pgPool)

	// Run database migrations
	if err := database.RunMigrations(cfg.Postgres.DSN, "migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Connect to MongoDB
	mongoClient, err := database.NewMongoClient(ctx, cfg.Mongo.URI)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer database.CloseMongo(ctx, mongoClient)

	// Connect to Redis
	redisClient, err := cache.NewRedisClient(ctx, cfg.Redis.Addr, cfg.Redis.Password)
	if err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer cache.CloseRedis(redisClient)

	// Wire Audit infrastructure
	mongoAuditRepo := mongorepo.NewAuditRepoMongo(mongoClient, cfg.Mongo.DB)
	auditRepo := audit.NewBufferedAuditLogger(mongoAuditRepo, redisClient)
	go auditRepo.Start(ctx)
	defer auditRepo.Stop()

	// Wire handlers
	healthHandler := handler.NewHealthHandler(pgPool, mongoClient, redisClient)

	// Wire Program module
	programRepo := pg.NewProgramRepoPG(pgPool)
	createProgramUC := programuc.NewCreateProgramUseCase(programRepo)
	getProgramUC := programuc.NewGetProgramUseCase(programRepo)
	listProgramsUC := programuc.NewListProgramsUseCase(programRepo)
	updateProgramUC := programuc.NewUpdateProgramUseCase(programRepo)
	deleteProgramUC := programuc.NewDeleteProgramUseCase(programRepo)
	programHandler := handler.NewProgramHandler(createProgramUC, getProgramUC, listProgramsUC, updateProgramUC, deleteProgramUC)

	// Wire Auth & User module
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, 15*time.Minute, 7*24*time.Hour)
	tokenBlocklist := cache.NewTokenBlocklist(redisClient)

	userRepo := pg.NewUserRepoPG(pgPool)
	roleRepo := pg.NewRoleRepoPG(pgPool)

	registerUC := authuc.NewRegisterUseCase(userRepo, roleRepo, jwtManager)
	loginUC := authuc.NewLoginUseCase(userRepo, jwtManager)
	refreshUC := authuc.NewRefreshTokenUseCase(userRepo, jwtManager, tokenBlocklist)
	profileUC := authuc.NewGetProfileUseCase(userRepo)
	authHandler := handler.NewAuthHandler(registerUC, loginUC, refreshUC, profileUC, jwtManager, tokenBlocklist)

	listUsersUC := useruc.NewListUsersUseCase(userRepo)
	updateUserUC := useruc.NewUpdateUserUseCase(userRepo)
	deleteUserUC := useruc.NewDeleteUserUseCase(userRepo)
	assignRoleUC := useruc.NewAssignRoleUseCase(userRepo, roleRepo)
	removeRoleUC := useruc.NewRemoveRoleUseCase(userRepo, roleRepo)
	userHandler := handler.NewUserHandler(listUsersUC, updateUserUC, deleteUserUC, assignRoleUC, removeRoleUC)

	// Wire Beneficiary module
	beneficiaryRepo := pg.NewBeneficiaryRepoPG(pgPool)
	createBeneficiaryUC := beneficiaryuc.NewCreateBeneficiaryUseCase(beneficiaryRepo)
	getBeneficiaryUC := beneficiaryuc.NewGetBeneficiaryUseCase(beneficiaryRepo)
	listBeneficiariesUC := beneficiaryuc.NewListBeneficiariesUseCase(beneficiaryRepo)
	updateBeneficiaryUC := beneficiaryuc.NewUpdateBeneficiaryUseCase(beneficiaryRepo)
	deleteBeneficiaryUC := beneficiaryuc.NewDeleteBeneficiaryUseCase(beneficiaryRepo)
	enrollInProgramUC := beneficiaryuc.NewEnrollInProgramUseCase(beneficiaryRepo, programRepo)
	removeFromProgramUC := beneficiaryuc.NewRemoveFromProgramUseCase(beneficiaryRepo)
	beneficiaryHandler := handler.NewBeneficiaryHandler(createBeneficiaryUC, getBeneficiaryUC, listBeneficiariesUC, updateBeneficiaryUC, deleteBeneficiaryUC, enrollInProgramUC, removeFromProgramUC)

	// Wire Tenant module
	tenantRepo := pg.NewTenantRepoPG(pgPool)
	createTenantUC := tenantuc.NewCreateTenantUseCase(tenantRepo, userRepo, auditRepo)
	getTenantUC := tenantuc.NewGetTenantUseCase(tenantRepo)
	listTenantsUC := tenantuc.NewListTenantsUseCase(tenantRepo)
	updateTenantUC := tenantuc.NewUpdateTenantUseCase(tenantRepo, auditRepo)
	deactivateTenantUC := tenantuc.NewDeactivateTenantUseCase(tenantRepo, auditRepo)
	deleteTenantUC := tenantuc.NewDeleteTenantUseCase(tenantRepo, auditRepo)
	tenantHandler := handler.NewTenantHandler(createTenantUC, getTenantUC, listTenantsUC, updateTenantUC, deactivateTenantUC, deleteTenantUC)

	// Wire Debt repo (needed by Property deactivation)
	debtRepo := pg.NewDebtRepoPG(pgPool)

	// Wire Property module
	propertyRepo := pg.NewPropertyRepoPG(pgPool)
	createPropertyUC := propertyuc.NewCreatePropertyUseCase(propertyRepo, userRepo, auditRepo)
	getPropertyUC := propertyuc.NewGetPropertyUseCase(propertyRepo)
	listPropertiesUC := propertyuc.NewListPropertiesUseCase(propertyRepo)
	updatePropertyUC := propertyuc.NewUpdatePropertyUseCase(propertyRepo, auditRepo)
	deactivatePropertyUC := propertyuc.NewDeactivatePropertyUseCase(propertyRepo, debtRepo, auditRepo)
	deletePropertyUC := propertyuc.NewDeletePropertyUseCase(propertyRepo, auditRepo)
	propertyHandler := handler.NewPropertyHandler(createPropertyUC, getPropertyUC, listPropertiesUC, updatePropertyUC, deactivatePropertyUC, deletePropertyUC)

	// Wire Debt module
	createDebtUC := debtuc.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propertyRepo)
	getDebtUC := debtuc.NewGetDebtUseCase(debtRepo)
	listDebtsUC := debtuc.NewListDebtsUseCase(debtRepo)
	updateDebtUC := debtuc.NewUpdateDebtUseCase(debtRepo)
	cancelDebtUC := debtuc.NewCancelDebtUseCase(debtRepo)
	markDebtPaidUC := debtuc.NewMarkDebtPaidUseCase(debtRepo)
	deleteDebtUC := debtuc.NewDeleteDebtUseCase(debtRepo)
	debtHandler := handler.NewDebtHandler(createDebtUC, getDebtUC, listDebtsUC, updateDebtUC, cancelDebtUC, markDebtPaidUC, deleteDebtUC)

	// Wire Transaction module
	transactionRepo := pg.NewTransactionRepoPG(pgPool)
	recordPaymentUC := transactionuc.NewRecordPaymentUseCase(transactionRepo, debtRepo, userRepo, tenantRepo)
	recordRefundUC := transactionuc.NewRecordRefundUseCase(transactionRepo, debtRepo, userRepo, tenantRepo)
	getTransactionUC := transactionuc.NewGetTransactionUseCase(transactionRepo)
	listTransactionsUC := transactionuc.NewListTransactionsUseCase(transactionRepo)
	verifyTransactionUC := transactionuc.NewVerifyTransactionUseCase(transactionRepo)
	transactionHandler := handler.NewTransactionHandler(recordPaymentUC, recordRefundUC, getTransactionUC, listTransactionsUC, verifyTransactionUC)

	// Set up router
	r := router.NewRouter(router.Handlers{
		Health:      healthHandler,
		Program:     programHandler,
		Auth:        authHandler,
		User:        userHandler,
		Beneficiary: beneficiaryHandler,
		Tenant:      tenantHandler,
		Property:    propertyHandler,
		Debt:        debtHandler,
		Transaction: transactionHandler,
	}, cfg.App.FrontendURL, jwtManager, tokenBlocklist)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
	slog.Info("server stopped")
}
