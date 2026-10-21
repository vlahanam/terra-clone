package common

// [VN] successResponse đóng gói cấu trúc phản hồi thành công chuẩn cho tất cả API
type successResponse struct {
	Data interface{} `json:"data"`
	Paging interface{} `json:"paging,omitempty"`
	Extra interface{} `json:"extra,omitempty"`
}

// [VN] SuccessResponse tạo đối tượng phản hồi thành công đầy đủ Data, Paging và Extra
func SuccessResponse(data, paging, extra interface{}) *successResponse {
	return &successResponse{Data: data, Paging: paging, Extra: extra}
}

// [VN] ResponseData tạo đối tượng phản hồi thành công chỉ chứa trường Data
func ResponseData(data interface{}) *successResponse {
	return SuccessResponse(data, nil, nil)
}
