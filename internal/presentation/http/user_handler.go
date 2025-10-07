package http

import (
	"encoding/json"
	"hub-user-service/internal/application/usecase"
	"net/http"
	"strings"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	loginUseCase          usecase.ILoginUseCase
	registerUseCase       usecase.IRegisterUserUseCase
	getUserProfileUseCase usecase.IGetUserProfileUseCase
	validateTokenUseCase  usecase.IValidateTokenUseCase
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(
	loginUC usecase.ILoginUseCase,
	registerUC usecase.IRegisterUserUseCase,
	getUserProfileUC usecase.IGetUserProfileUseCase,
	validateTokenUC usecase.IValidateTokenUseCase,
) *UserHandler {
	return &UserHandler{
		loginUseCase:          loginUC,
		registerUseCase:       registerUC,
		getUserProfileUseCase: getUserProfileUC,
		validateTokenUseCase:  validateTokenUC,
	}
}

// Login handles user login via HTTP
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd usecase.LoginCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.loginUseCase.Execute(&cmd)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// Register handles user registration via HTTP
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd usecase.RegisterUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.registerUseCase.Execute(&cmd)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, result)
}

// GetProfile handles getting user profile via HTTP
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.extractUserIDFromToken(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	result, err := h.getUserProfileUseCase.Execute(userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// ValidateToken handles token validation via HTTP
func (h *UserHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
		return
	}

	result, err := h.validateTokenUseCase.Execute(token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, result)
}

// HealthCheck handles health check requests
func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"healthy": true,
		"version": "1.0.0",
	})
}

// extractUserIDFromToken extracts user ID from JWT token in the request
func (h *UserHandler) extractUserIDFromToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", http.ErrNoCookie
	}

	result, err := h.validateTokenUseCase.Execute(token)
	if err != nil {
		return "", err
	}

	if !result.Valid {
		return "", http.ErrNoCookie
	}

	return result.UserID, nil
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to marshal response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// AuthMiddleware validates JWT tokens for protected routes
func (h *UserHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		result, err := h.validateTokenUseCase.Execute(token)
		if err != nil || !result.Valid {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		next(w, r)
	}
}
