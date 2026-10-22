package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/middleware"
	"github.com/vlahanam/terra-clone/services"
)

// [VN] AuthController điều phối tiếp nhận và xử lý HTTP request liên quan đến xác thực
type AuthController struct {
	authService services.AuthService
}

// [VN] NewAuthController khởi tạo AuthController với AuthService
func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// [VN] Register tiếp nhận request đăng ký tài khoản mới
func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Invalid registration data. Please verify email format and password (minimum 8 characters)").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.authService.Register(c.Request.Context(), &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusCreated, common.ResponseData(res))
}

// [VN] Login tiếp nhận request xác thực người dùng
func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Email and password cannot be empty").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

// [VN] RefreshToken cấp lại cặp token mới khi Access Token hết hạn
func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Refresh token cannot be empty").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

// [VN] LogoutAll thu hồi toàn bộ phiên đăng nhập của người dùng
func (ctrl *AuthController) LogoutAll(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		errResp := common.ErrUnauthorized.WithReason("Invalid session").WithTrace(err)
		c.JSON(http.StatusUnauthorized, errResp)
		return
	}

	if err := ctrl.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(gin.H{
		"message": "All sessions have been revoked successfully",
	}))
}

// [VN] GetMe trả về thông tin chi tiết của người dùng đang đăng nhập
func (ctrl *AuthController) GetMe(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		errResp := common.ErrUnauthorized.WithReason("Invalid session").WithTrace(err)
		c.JSON(http.StatusUnauthorized, errResp)
		return
	}

	res, err := ctrl.authService.GetMe(c.Request.Context(), userID)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}
