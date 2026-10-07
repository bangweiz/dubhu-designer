package controller

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bangweiz/dubhu-designer/internal/dto"
	"github.com/bangweiz/dubhu-designer/internal/identity"
	"github.com/bangweiz/dubhu-designer/internal/mapper"
	"github.com/bangweiz/dubhu-designer/internal/models"
	"github.com/bangweiz/dubhu-designer/internal/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthController struct{ authService *service.AuthService }

func NewAuthController(s *service.AuthService) *AuthController {
	return &AuthController{authService: s}
}

// Registration and login are the only public API routes.
func (c *AuthController) RegisterRoutes(rg *gin.RouterGroup) {
	limit := authRateLimit()
	rg.POST("/organisations", limit, c.CreateOrganisation)
	rg.POST("/organisations/:organisationId/auth/login", limit, c.Login)
	authenticated := rg.Group("/organisations/:organisationId")
	authenticated.GET("/auth/me", c.Me)
	authenticated.POST("/auth/logout", c.Logout)
	authenticated.POST("/accounts", c.CreateAccount)
}
func (c *AuthController) CreateOrganisation(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")
	var req dto.CreateOrganisationDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.Trim()
	resp, err := c.authService.CreateOrganisation(ctx.Request.Context(), req)
	if err != nil {
		writeServiceError(ctx, err, errorContext{nameValue: req.Name})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": resp})
}
func (c *AuthController) CreateAccount(ctx *gin.Context) {
	var req dto.CreateAccountDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := c.authService.CreateAccount(ctx.Request.Context(), req)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": resp})
}
func (c *AuthController) Login(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")
	var req dto.LoginDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	resp, err := c.authService.Login(ctx.Request.Context(), ctx.Param("organisationId"), req)
	if err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}
func bearerToken(ctx *gin.Context) string {
	parts := strings.Fields(ctx.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// RequirePermissions protects the entire API. Public routes are explicitly allowlisted;
// every other route must authenticate, match its organisation, and pass a role check.
func (c *AuthController) RequirePermissions() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.FullPath()
		if ctx.Request.Method == http.MethodPost && (path == "/api/v1/organisations" || path == "/api/v1/organisations/:organisationId/auth/login") {
			ctx.Next()
			return
		}
		ctx.Header("Cache-Control", "no-store")
		account, err := c.authService.Authenticate(ctx.Request.Context(), bearerToken(ctx))
		if err != nil {
			writeServiceError(ctx, err)
			ctx.Abort()
			return
		}
		orgID, err := bson.ObjectIDFromHex(ctx.Param("organisationId"))
		if err != nil || orgID != account.OrganisationID {
			writeServiceError(ctx, service.ErrForbidden)
			ctx.Abort()
			return
		}
		ctx.Set("authenticatedAccount", account)
		principal := identity.Principal{OrganisationID: account.OrganisationID, AccountID: account.ID, Role: string(account.Role)}
		ctx.Request = ctx.Request.WithContext(identity.WithPrincipal(ctx.Request.Context(), principal))
		if !permitted(ctx.Request.Method, path, account.Role) {
			writeServiceError(ctx, service.ErrForbidden)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func permitted(method, path string, role models.AccountRole) bool {
	if role != models.RoleRoot && role != models.RoleAdmin && role != models.RoleUser {
		return false
	}
	const prefix = "/api/v1/organisations/:organisationId/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	resource := strings.TrimPrefix(path, prefix)
	if resource == "accounts" {
		return method == http.MethodPost && role == models.RoleRoot
	}
	if resource == "auth/me" {
		return method == http.MethodGet
	}
	if resource == "auth/logout" {
		return method == http.MethodPost
	}
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return role == models.RoleRoot || role == models.RoleAdmin
	default:
		return false
	}
}
func (c *AuthController) Me(ctx *gin.Context) {
	account := ctx.MustGet("authenticatedAccount").(*models.Account)
	ctx.JSON(http.StatusOK, gin.H{"data": mapper.ToAccountResponseDTO(account)})
}
func (c *AuthController) Logout(ctx *gin.Context) {
	if err := c.authService.Logout(ctx.Request.Context(), bearerToken(ctx)); err != nil {
		writeServiceError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// A bounded per-process throttle is a baseline; multi-instance deployments need shared gateway limiting.
func authRateLimit() gin.HandlerFunc {
	type window struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	entries := map[string]window{}
	return func(ctx *gin.Context) {
		now := time.Now()
		key := ctx.ClientIP()
		mu.Lock()
		for k, v := range entries {
			if !v.reset.After(now) {
				delete(entries, k)
			}
		}
		current := entries[key]
		if current.reset.IsZero() {
			current.reset = now.Add(time.Minute)
		}
		if current.count >= 30 || (len(entries) >= 10000 && current.count == 0) {
			mu.Unlock()
			ctx.Header("Retry-After", "60")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many authentication requests"})
			return
		}
		current.count++
		entries[key] = current
		mu.Unlock()
		ctx.Next()
	}
}
