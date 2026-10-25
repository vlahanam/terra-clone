package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/services"
)

// [VN] UserCompanyController điều phối các yêu cầu HTTP liên quan đến quan hệ người dùng và công ty
type UserCompanyController struct {
	userCompanyService services.UserCompanyService
}

// [VN] NewUserCompanyController khởi tạo UserCompanyController với UserCompanyService
func NewUserCompanyController(userCompanyService services.UserCompanyService) *UserCompanyController {
	return &UserCompanyController{userCompanyService: userCompanyService}
}

// [VN] Create xử lý yêu cầu liên kết một người dùng vào công ty
func (ctrl *UserCompanyController) Create(c *gin.Context) {
	var req dto.CreateUserCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Invalid user company association data").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// [VN] Nếu được gọi qua sub-route /companies/:id/users và chưa có company_id trong body, gán từ URL param
	if companyIDParam := c.Param("id"); companyIDParam != "" && req.CompanyID == "" {
		req.CompanyID = companyIDParam
	}

	res, err := ctrl.userCompanyService.CreateUserCompany(c.Request.Context(), &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusCreated, common.ResponseData(res))
}

// [VN] GetByID lấy thông tin chi tiết liên kết theo ID hoặc public_id
func (ctrl *UserCompanyController) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("User company identifier is required in URL path")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.userCompanyService.GetUserCompanyByID(c.Request.Context(), id)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

// [VN] List trả về danh sách liên kết có hỗ trợ phân trang Cursor và tìm kiếm
func (ctrl *UserCompanyController) List(c *gin.Context) {
	var query dto.ListUserCompaniesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Invalid list query parameters").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	// [VN] Hỗ trợ filter tự động khi gọi qua các sub-route tiện ích
	if companyIDParam := c.Param("company_id"); companyIDParam != "" && query.CompanyID == "" {
		query.CompanyID = companyIDParam
	} else if idParam := c.Param("id"); idParam != "" && c.FullPath() != "/api/v1/user-companies" {
		// [VN] Phân biệt route /companies/:id/users vs /users/:id/companies
		if c.Request.URL.Path != "" {
			if query.CompanyID == "" && (c.FullPath() == "/api/v1/companies/:id/users") {
				query.CompanyID = idParam
			} else if query.UserID == "" && (c.FullPath() == "/api/v1/users/:id/companies") {
				query.UserID = idParam
			}
		}
	}

	results, paging, err := ctrl.userCompanyService.ListUserCompanies(c.Request.Context(), &query)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(results, paging, nil))
}

// [VN] Update cập nhật thông tin vai trò, chức danh, phòng ban và quyền hạn của liên kết
func (ctrl *UserCompanyController) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("User company identifier is required in URL path")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	var req dto.UpdateUserCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Invalid user company update data").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.userCompanyService.UpdateUserCompany(c.Request.Context(), id, &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

// [VN] Delete thực hiện xóa mềm hoặc xóa cứng liên kết UserCompany
func (ctrl *UserCompanyController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("User company identifier is required in URL path")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	hardQuery := c.Query("hard")
	hard, _ := strconv.ParseBool(hardQuery)

	if err := ctrl.userCompanyService.DeleteUserCompany(c.Request.Context(), id, hard); err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(gin.H{
		"message": "User company association deleted successfully",
	}))
}
