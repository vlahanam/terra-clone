package common

import (
	stderr "errors"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type DefaultError struct {
	// [VN] Mã lỗi logic
	IDField string `json:"id,omitempty"`

	// [VN] Mã trạng thái
	//
	// [VN] Ví dụ: 404
	CodeField int `json:"code,omitempty"`

	// [VN] Mô tả trạng thái
	//
	// [VN] Ví dụ: Not Found
	StatusField string `json:"status,omitempty"`

	// [VN] Request ID
	//
	// [VN] Ví dụ: d7ef54b1-ec15-46e6-bccb-524b82c035e6
	RIDField string `json:"request,omitempty"`

	// [VN] Lý do gây ra lỗi, dùng để trả về cho người dùng
	//
	// [VN] ví dụ: Người dùng có ID 1234 không tồn tại
	ReasonField string `json:"reason,omitempty"`

	// [VN] Thông tin debug
	//
	// [VN] Không trả về cho người dùng
	//
	// [VN] Ví dụ: SQL field "foo" is not a bool.
	DebugField string `json:"debug,omitempty"`

	// [VN] Thông báo lỗi
	//
	// [VN] Ví dụ: Không tìm thấy tài nguyên
	// Required: true
	ErrorField string `json:"message"`

	// [VN] Chi tiết lỗi bổ sung
	DetailsField map[string]interface{} `json:"details,omitempty"`

	err error
}

func (e *DefaultError) StackTrace() (trace errors.StackTrace) {
	if e.err == e {
		return
	}

	if st := stackTracer(nil); stderr.As(e.err, &st) {
		trace = st.StackTrace()
	}

	return
}

func (e DefaultError) Unwrap() error {
	return e.err
}

func (e *DefaultError) Wrap(err error) {
	e.err = err
}

func (e DefaultError) WithWrap(err error) *DefaultError {
	e.err = err
	return &e
}

func (e DefaultError) WithID(id string) *DefaultError {
	e.IDField = id
	return &e
}

func (e *DefaultError) WithTrace(err error) *DefaultError {
	if st := stackTracer(nil); !stderr.As(e.err, &st) {
		e.Wrap(errors.WithStack(err))
	} else {
		e.Wrap(err)
	}
	return e
}

func (e DefaultError) Is(err error) bool {
	switch te := err.(type) {
	case DefaultError:
		return e.ErrorField == te.ErrorField &&
			e.StatusField == te.StatusField &&
			e.IDField == te.IDField &&
			e.CodeField == te.CodeField
	case *DefaultError:
		return e.ErrorField == te.ErrorField &&
			e.StatusField == te.StatusField &&
			e.IDField == te.IDField &&
			e.CodeField == te.CodeField
	default:
		return false
	}
}

func (e DefaultError) Status() string {
	return e.StatusField
}

func (e DefaultError) ID() string {
	return e.IDField
}

func (e DefaultError) Error() string {
	return e.ErrorField
}

func (e DefaultError) RequestID() string {
	return e.RIDField
}

func (e DefaultError) Reason() string {
	return e.ReasonField
}

func (e DefaultError) Debug() string {
	return e.DebugField
}

func (e DefaultError) Details() map[string]interface{} {
	return e.DetailsField
}

func (e DefaultError) StatusCode() int {
	return e.CodeField
}

func (e DefaultError) WithReason(reason string) *DefaultError {
	e.ReasonField = reason
	return &e
}

func (e DefaultError) WithReasonf(reason string, args ...interface{}) *DefaultError {
	return e.WithReason(fmt.Sprintf(reason, args...))
}

func (e DefaultError) WithError(message string) *DefaultError {
	e.ErrorField = message
	return &e
}

func (e DefaultError) WithErrorf(message string, args ...interface{}) *DefaultError {
	return e.WithError(fmt.Sprintf(message, args...))
}

func (e DefaultError) WithDebugf(debug string, args ...interface{}) *DefaultError {
	return e.WithDebug(fmt.Sprintf(debug, args...))
}

func (e DefaultError) WithDebug(debug string) *DefaultError {
	e.DebugField = debug
	return &e
}

func (e DefaultError) WithDetail(key string, detail interface{}) *DefaultError {
	if e.DetailsField == nil {
		e.DetailsField = map[string]interface{}{}
	}
	e.DetailsField[key] = detail
	return &e
}

func (e DefaultError) WithDetailf(key string, message string, args ...interface{}) *DefaultError {
	if e.DetailsField == nil {
		e.DetailsField = map[string]interface{}{}
	}
	e.DetailsField[key] = fmt.Sprintf(message, args...)
	return &e
}

func (e DefaultError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "id=%s\n", e.IDField)
			_, _ = fmt.Fprintf(s, "rid=%s\n", e.RIDField)
			_, _ = fmt.Fprintf(s, "error=%s\n", e.ErrorField)
			_, _ = fmt.Fprintf(s, "reason=%s\n", e.ReasonField)
			_, _ = fmt.Fprintf(s, "details=%+v\n", e.DetailsField)
			_, _ = fmt.Fprintf(s, "debug=%s\n", e.DebugField)
			e.StackTrace().Format(s, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.ErrorField)
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.ErrorField)
	}
}

func ToDefaultError(err error, requestID string) *DefaultError {
	de := &DefaultError{
		RIDField:     requestID,
		CodeField:    http.StatusInternalServerError,
		DetailsField: map[string]interface{}{},
		ErrorField:   err.Error(),
	}
	de.Wrap(err)

	if c := ReasonCarrier(nil); stderr.As(err, &c) {
		de.ReasonField = c.Reason()
	}
	if c := RequestIDCarrier(nil); stderr.As(err, &c) && c.RequestID() != "" {
		de.RIDField = c.RequestID()
	}
	if c := DetailsCarrier(nil); stderr.As(err, &c) && c.Details() != nil {
		de.DetailsField = c.Details()
	}
	if c := StatusCarrier(nil); stderr.As(err, &c) && c.Status() != "" {
		de.StatusField = c.Status()
	}
	if c := StatusCodeCarrier(nil); stderr.As(err, &c) && c.StatusCode() != 0 {
		de.CodeField = c.StatusCode()
	}
	if c := DebugCarrier(nil); stderr.As(err, &c) {
		de.DebugField = c.Debug()
	}
	if c := IDCarrier(nil); stderr.As(err, &c) {
		de.IDField = c.ID()
	}

	if de.StatusField == "" {
		de.StatusField = http.StatusText(de.StatusCode())
	}

	return de
}

type StatusCodeCarrier interface {
	StatusCode() int
}

type RequestIDCarrier interface {
	RequestID() string
}

type ReasonCarrier interface {
	Reason() string
}

type DebugCarrier interface {
	Debug() string
}

type StatusCarrier interface {
	Status() string
}

type DetailsCarrier interface {
	Details() map[string]interface{}
}

type IDCarrier interface {
	ID() string
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

var ErrNotFound = DefaultError{
	StatusField: http.StatusText(http.StatusNotFound),
	ErrorField:  "[VN] Không tìm thấy tài nguyên được yêu cầu",
	CodeField:   http.StatusNotFound,
}

var ErrUnauthorized = DefaultError{
	StatusField: http.StatusText(http.StatusUnauthorized),
	ErrorField:  "[VN] Yêu cầu không thể được cấp quyền",
	CodeField:   http.StatusUnauthorized,
}

var ErrForbidden = DefaultError{
	StatusField: http.StatusText(http.StatusForbidden),
	ErrorField:  "[VN] Bạn không có quyền thực hiện hành động này",
	CodeField:   http.StatusForbidden,
}

var ErrInternalServerError = DefaultError{
	StatusField: http.StatusText(http.StatusInternalServerError),
	ErrorField:  "[VN] Đã xảy ra lỗi máy chủ nội bộ, vui lòng liên hệ với quản trị viên hệ thống",
	CodeField:   http.StatusInternalServerError,
}

var ErrBadRequest = DefaultError{
	StatusField: http.StatusText(http.StatusBadRequest),
	ErrorField:  "[VN] Yêu cầu không hợp lệ hoặc chứa các tham số sai định dạng",
	CodeField:   http.StatusBadRequest,
}

var ErrUnsupportedMediaType = DefaultError{
	StatusField: http.StatusText(http.StatusUnsupportedMediaType),
	ErrorField:  "[VN] Yêu cầu đang sử dụng định dạng nội dung không được hỗ trợ",
	CodeField:   http.StatusUnsupportedMediaType,
}

var ErrConflict = DefaultError{
	StatusField: http.StatusText(http.StatusConflict),
	ErrorField:  "[VN] Không thể tạo tài nguyên do xảy ra xung đột dữ liệu",
	CodeField:   http.StatusConflict,
}

var ErrRecordNotFound = errors.New("[VN] Không tìm thấy bản ghi")
