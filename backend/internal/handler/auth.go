package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"gm_site/internal/middleware"
	"gm_site/internal/model"
	"gm_site/internal/repository"
	"gm_site/internal/service"
)

// validate is a singleton instance used for struct tag validation.
var validate = validator.New()

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtSvc     *service.JWTService
	emailSvc   service.EmailService
	adminEmail string
}

// NewAuthHandler creates a new AuthHandler with the given dependencies.
func NewAuthHandler(userRepo *repository.UserRepository, jwtSvc *service.JWTService, emailSvc service.EmailService, adminEmail string) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtSvc:     jwtSvc,
		emailSvc:   emailSvc,
		adminEmail: adminEmail,
	}
}

// Login handles POST /api/auth/login.
//
// It validates the user's credentials and returns a token pair on success.
// Users with status "pending" or "rejected" are not allowed to log in.
func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind login request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求数据")
	}

	// Find user by email
	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusUnauthorized, "邮箱或密码错误")
		}
		middleware.GetLogger(c).Error("failed to find user by email", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		middleware.GetLogger(c).Error("password verification failed", "err", err)
		return Error(c, http.StatusUnauthorized, "邮箱或密码错误")
	}

	// Check user status
	switch user.Status {
	case model.UserStatusPending:
		return Error(c, http.StatusForbidden, "账户正在审核中")
	case model.UserStatusRejected:
		return Error(c, http.StatusForbidden, "账户已被拒绝")
	case model.UserStatusApproved:
		// OK
	default:
		return Error(c, http.StatusForbidden, "账户状态异常")
	}

	// Generate token pair
	tokenPair, err := h.jwtSvc.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		middleware.GetLogger(c).Error("failed to generate token pair", "err", err)
		return Error(c, http.StatusInternalServerError, "生成令牌失败")
	}

	return Success(c, tokenPair)
}

// Register handles POST /api/auth/register.
//
// It validates the request, checks for duplicate emails, hashes the password,
// creates a pending user, and asynchronously notifies the admin.
func (h *AuthHandler) Register(c echo.Context) error {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind register request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求数据")
	}

	// Validate request fields using struct tags
	if err := validate.Struct(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			errs := make(map[string]string)
			for _, fe := range ve {
				errs[fe.Field()] = fe.Tag()
			}
			return c.JSON(http.StatusBadRequest, model.APIResponse{
				Code:    http.StatusBadRequest,
				Message: "请求参数校验失败",
				Data:    errs,
			})
		}
		middleware.GetLogger(c).Error("request validation failed", "err", err)
		return Error(c, http.StatusBadRequest, "请求参数校验失败")
	}

	// Check if email is already registered
	existingUser, err := h.userRepo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		middleware.GetLogger(c).Error("failed to check email existence", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}
	if existingUser != nil {
		return Error(c, http.StatusConflict, "该邮箱已被注册")
	}

	// Hash password with bcrypt cost=10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		middleware.GetLogger(c).Error("failed to hash password", "err", err)
		return Error(c, http.StatusInternalServerError, "密码加密失败")
	}

	// Determine role: admin emails get admin role
	role := model.UserRoleUser
	if req.Email == h.adminEmail {
		role = model.UserRoleAdmin
	}

	// Create user: admin auto-approved, others pending
	status := model.UserStatusPending
	if role == model.UserRoleAdmin {
		status = model.UserStatusApproved
	}
	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
		Role:         role,
		Status:       status,
	}

	if err := h.userRepo.Create(user); err != nil {
		middleware.GetLogger(c).Error("failed to create user", "err", err)
		return Error(c, http.StatusInternalServerError, "创建用户失败")
	}

	// Asynchronously notify admin
	go func() {
		if err := h.emailSvc.SendAdminNotification(req.Email, req.Nickname); err != nil {
			logger := middleware.GetLogger(c)
			logger.Error("SendAdminNotification failed", "err", err, "email", req.Email, "nickname", req.Nickname)
		} else {
			logger := middleware.GetLogger(c)
			logger.Info("admin email notification sent", "to", h.adminEmail, "new_user", req.Email)
		}
	}()

	msg := "注册成功，请等待管理员审核"
	if role == model.UserRoleAdmin {
		msg = "注册成功"
	}
	return Created(c, map[string]string{
		"message": msg,
	})
}

// Refresh handles POST /api/auth/refresh.
//
// It validates the refresh token and returns a new token pair.
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req model.RefreshRequest
	if err := c.Bind(&req); err != nil {
		middleware.GetLogger(c).Error("failed to bind refresh request", "err", err)
		return Error(c, http.StatusBadRequest, "无效的请求数据")
	}

	// Validate refresh token
	userID, err := h.jwtSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Error(c, http.StatusUnauthorized, "刷新令牌已过期")
		}
		middleware.GetLogger(c).Error("failed to validate refresh token", "err", err)
		return Error(c, http.StatusUnauthorized, "无效的刷新令牌")
	}

	// Fetch user to get current role
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusUnauthorized, "用户不存在")
		}
		middleware.GetLogger(c).Error("failed to find user for refresh", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	// Generate new token pair
	tokenPair, err := h.jwtSvc.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		middleware.GetLogger(c).Error("failed to generate token pair on refresh", "err", err)
		return Error(c, http.StatusInternalServerError, "生成令牌失败")
	}

	return Success(c, tokenPair)
}

// Me handles GET /api/auth/me.
//
// It returns the authenticated user's profile information.
func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := c.Get(middleware.UserIDKey).(int64)
	if !ok {
		return Error(c, http.StatusUnauthorized, "未登录")
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Error(c, http.StatusUnauthorized, "用户不存在")
		}
		middleware.GetLogger(c).Error("failed to find user by ID", "err", err)
		return Error(c, http.StatusInternalServerError, "服务器内部错误")
	}

	return Success(c, map[string]interface{}{
		"id":       user.ID,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
		"status":   user.Status,
	})
}
