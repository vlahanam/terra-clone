package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/services"
)

type UserController struct {
	userService services.UserService
}

func NewUserController(userService services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (ctrl *UserController) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Dữ liệu tạo người dùng không hợp lệ").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusCreated, common.ResponseData(res))
}

func (ctrl *UserController) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("Thiếu ID người dùng trong URL")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

func (ctrl *UserController) List(c *gin.Context) {
	var query dto.ListUsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Tham số truy vấn danh sách không hợp lệ").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	users, paging, err := ctrl.userService.ListUsers(c.Request.Context(), &query)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(users, paging, nil))
}

func (ctrl *UserController) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("Thiếu ID người dùng trong URL")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Dữ liệu cập nhật người dùng không hợp lệ").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.userService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

func (ctrl *UserController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("Thiếu ID người dùng trong URL")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	hardQuery := c.Query("hard")
	hard, _ := strconv.ParseBool(hardQuery)

	if err := ctrl.userService.DeleteUser(c.Request.Context(), id, hard); err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(gin.H{
		"message": "Đã xóa người dùng thành công",
	}))
}
