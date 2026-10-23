package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vlahanam/terra-clone/common"
	"github.com/vlahanam/terra-clone/dto"
	"github.com/vlahanam/terra-clone/middleware"
	"github.com/vlahanam/terra-clone/services"
)

// [VN] CompanyController điều phối các yêu cầu HTTP liên quan đến quản lý công ty
type CompanyController struct {
	companyService services.CompanyService
}

// [VN] NewCompanyController khởi tạo CompanyController với CompanyService
func NewCompanyController(companyService services.CompanyService) *CompanyController {
	return &CompanyController{companyService: companyService}
}

// [VN] Create xử lý yêu cầu tạo công ty mới
func (ctrl *CompanyController) Create(c *gin.Context) {
	var req dto.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Invalid company creation data").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	var ownerID *uuid.UUID
	if currentUserID, err := middleware.GetCurrentUserID(c); err == nil && currentUserID != uuid.Nil {
		ownerID = &currentUserID
	}

	res, err := ctrl.companyService.CreateCompany(c.Request.Context(), &req, ownerID)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusCreated, common.ResponseData(res))
}

// [VN] GetByID lấy thông tin công ty theo ID, PublicID hoặc Code
func (ctrl *CompanyController) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("Company identifier is required in URL path")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.companyService.GetCompanyByID(c.Request.Context(), id)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

// [VN] List trả về danh sách công ty kèm phân trang cursor và tìm kiếm
func (ctrl *CompanyController) List(c *gin.Context) {
	var query dto.ListCompaniesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Invalid list query parameters").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	companies, paging, err := ctrl.companyService.ListCompanies(c.Request.Context(), &query)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(companies, paging, nil))
}

// [VN] Update cập nhật thông tin công ty và cấu hình metadata
func (ctrl *CompanyController) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("Company identifier is required in URL path")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := common.ErrBadRequest.
			WithReason("Invalid company update data").
			WithTrace(err)
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	res, err := ctrl.companyService.UpdateCompany(c.Request.Context(), id, &req)
	if err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(res))
}

// [VN] Delete thực hiện soft-delete hoặc hard-delete công ty
func (ctrl *CompanyController) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		errResp := common.ErrBadRequest.WithReason("Company identifier is required in URL path")
		c.JSON(http.StatusBadRequest, errResp)
		return
	}

	hardQuery := c.Query("hard")
	hard, _ := strconv.ParseBool(hardQuery)

	if err := ctrl.companyService.DeleteCompany(c.Request.Context(), id, hard); err != nil {
		reqID := c.GetString("request_id")
		de := common.ToDefaultError(err, reqID)
		c.JSON(de.StatusCode(), de)
		return
	}

	c.JSON(http.StatusOK, common.ResponseData(gin.H{
		"message": "Company deleted successfully",
	}))
}
