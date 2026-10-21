package common

// [VN] Paging quản lý phân trang con trỏ (cursor-based pagination)
type Paging struct {
	// [VN] Tổng số lượng bản ghi
	Total int64 `json:"total" form:"-"`
	// [VN] Con trỏ hiện tại truyền từ client
	FakeCursor string `json:"cursor" form:"cursor"`
	// [VN] Con trỏ tiếp theo cho trang kế tiếp
	NextCursor string `json:"next_cursor"`
}