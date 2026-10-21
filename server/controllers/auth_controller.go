package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/middleware"
	"github.com/vlahanam/terra-clone/services"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Dữ liệu đăng ký không hợp lệ, vui lòng kiểm tra lại email (đúng định dạng) và mật khẩu (tối thiểu 8 ký tự)").
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

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Email và mật khẩu không được để trống").
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

func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Refresh token không được để trống").
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

func (ctrl *AuthController) LogoutAll(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		errResp := common.ErrUnauthorized.WithReason("Phiên làm việc không hợp lệ").WithTrace(err)
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
		"message": "Đã thu hồi tất cả phiên đăng nhập thành công",
	}))
}

func (ctrl *AuthController) GetMe(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		errResp := common.ErrUnauthorized.WithReason("Phiên làm việc không hợp lệ").WithTrace(err)
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
