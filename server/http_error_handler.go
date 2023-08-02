package server

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	clienterrorspb "github.com/requiemofthesouls/client-errors/pb"
	"github.com/requiemofthesouls/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	ErrorHandler interface {
		HTTPError(context.Context, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, error)
	}

	errorHandler struct {
		l             logger.Wrapper
		httpStatusMap HTTPStatusMap
	}

	HTTPStatusMap map[codes.Code]int

	httpError struct {
		Code    codes.Code        `json:"code"`
		Msg     string            `json:"msg"`
		Details *httpErrorDetails `json:"details,omitempty"`
	}
	httpErrorDetails struct {
		Msg              string            `json:"msg,omitempty"`
		TKey             string            `json:"tKey,omitempty"`
		Params           map[string]string `json:"params,omitempty"`
		ValidationErrors *validationErrors `json:"validationErrors,omitempty"`
	}
	validationErrors struct {
		Items []*validationErrorItem `json:"items"`
	}
	validationErrorItem struct {
		Field string `json:"field"`
		Msg   string `json:"msg"`
		TKey  string `json:"tKey,omitempty"`
	}
)

func NewErrorHandler(l logger.Wrapper, httpStatusMap HTTPStatusMap) ErrorHandler {
	return &errorHandler{l: l, httpStatusMap: httpStatusMap}
}

func (h *errorHandler) HTTPError(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	req *http.Request,
	err error,
) {
	w.Header().Del("Trailer")
	w.Header().Set("Content-Type", marshaler.ContentType(req.Body))

	var (
		ok bool
		s  *status.Status
	)
	if s, ok = status.FromError(err); !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	body := httpError{Code: s.Code(), Msg: s.Message()}

	for _, detail := range s.Details() {
		if d, ok := detail.(*clienterrorspb.ErrorDetails); ok {
			body.Details = newHTTPErrorDetails(d)
			break
		}
	}

	var (
		buff []byte
		mErr error
	)
	if buff, mErr = marshaler.Marshal(body); mErr != nil {
		h.l.Error("failed to marshal error message", logger.Reflect("body", body), logger.Error(mErr))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, `{"exception": {"error": "failed to marshal error message"}}`); err != nil {
			h.l.Error("failed to write response", logger.Error(err))
		}
		return
	}

	var statusCode int
	if statusCode, ok = h.httpStatusMap[s.Code()]; !ok {
		statusCode = runtime.HTTPStatusFromCode(s.Code())
	}

	w.WriteHeader(statusCode)
	if _, err := w.Write(buff); err != nil {
		h.l.Error("failed to write response", logger.Error(err))
	}
}

func newHTTPErrorDetails(pbErrorDetails *clienterrorspb.ErrorDetails) *httpErrorDetails {
	if pbErrorDetails == nil {
		return nil
	}

	errorDetails := &httpErrorDetails{
		Msg:    pbErrorDetails.Msg,
		Params: pbErrorDetails.Params,
	}

	errorDetails.ValidationErrors = newValidationErrors(pbErrorDetails.ValidationErrors)

	return errorDetails
}

func newValidationErrors(pbValidationErrors *clienterrorspb.ErrorDetails_ValidationErrors) *validationErrors {
	if pbValidationErrors == nil {
		return nil
	}

	vErrors := &validationErrors{
		Items: make([]*validationErrorItem, 0, len(pbValidationErrors.Items)),
	}

	for _, pbValidationErrorItem := range pbValidationErrors.Items {
		vError := validationErrorItem{
			Field: pbValidationErrorItem.Field,
			Msg:   pbValidationErrorItem.Msg,
		}

		vErrors.Items = append(vErrors.Items, &vError)
	}

	return vErrors
}
