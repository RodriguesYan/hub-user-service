package container

import (
	"hub-user-service/internal/application/usecase"
	"hub-user-service/internal/domain/repository"
	"hub-user-service/internal/domain/service"
	"hub-user-service/internal/infra/persistence"
	grpcPresentation "hub-user-service/internal/presentation/grpc"
	httpPresentation "hub-user-service/internal/presentation/http"
	"hub-user-service/shared/config"
	"hub-user-service/shared/database"
	"log"
)

// Container holds all application dependencies
type Container struct {
	// Configuration
	Config *config.Config

	// Database
	DB *database.Database

	// Repositories
	UserRepository repository.IUserRepository

	// Domain Services
	TokenService service.ITokenService
	AuthService  service.IAuthService

	// Use Cases
	LoginUseCase          usecase.ILoginUseCase
	RegisterUserUseCase   usecase.IRegisterUserUseCase
	GetUserProfileUseCase usecase.IGetUserProfileUseCase
	ValidateTokenUseCase  usecase.IValidateTokenUseCase

	// Presentation Layers
	UserGRPCServer  *grpcPresentation.UserGRPCServer
	UserHTTPHandler *httpPresentation.UserHandler
}

// NewContainer creates and initializes a new dependency injection container
func NewContainer() (*Container, error) {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Database connection established")

	// Initialize repositories
	userRepo := persistence.NewUserRepository(db.DB)

	// Initialize domain services
	tokenService := service.NewTokenService(cfg.JWTSecret, cfg.TokenExpiration)
	authService := service.NewAuthService(userRepo, tokenService)

	// Initialize use cases
	loginUseCase := usecase.NewLoginUseCase(authService)
	registerUserUseCase := usecase.NewRegisterUserUseCase(userRepo)
	getUserProfileUseCase := usecase.NewGetUserProfileUseCase(userRepo)
	validateTokenUseCase := usecase.NewValidateTokenUseCase(authService)

	// Initialize presentation layers
	userGRPCServer := grpcPresentation.NewUserGRPCServer(
		loginUseCase,
		registerUserUseCase,
		getUserProfileUseCase,
		validateTokenUseCase,
	)

	userHTTPHandler := httpPresentation.NewUserHandler(
		loginUseCase,
		registerUserUseCase,
		getUserProfileUseCase,
		validateTokenUseCase,
	)

	log.Println("✅ Dependency injection container initialized")

	return &Container{
		Config:                cfg,
		DB:                    db,
		UserRepository:        userRepo,
		TokenService:          tokenService,
		AuthService:           authService,
		LoginUseCase:          loginUseCase,
		RegisterUserUseCase:   registerUserUseCase,
		GetUserProfileUseCase: getUserProfileUseCase,
		ValidateTokenUseCase:  validateTokenUseCase,
		UserGRPCServer:        userGRPCServer,
		UserHTTPHandler:       userHTTPHandler,
	}, nil
}

// Close closes all resources in the container
func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
