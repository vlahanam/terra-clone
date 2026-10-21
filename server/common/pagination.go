package common

// [VN] Paging quản lý phân trang con trỏ (cursor-based pagination)
type Paging struct {
	Total int64 `json:"total" form:"-"`
	FakeCursor string `json:"cursor" form:"cursor"`
	NextCursor string `json:"next_cursor"`
}