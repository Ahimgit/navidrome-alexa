package mid

import (
	"bytes"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var apiKeyRegexFilter = regexp.MustCompile(`(apiKey=)[^&]*`)

type ResponseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w ResponseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogsMiddleware(logRequests bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		logger := log.CreateRequestContextLogger(context)
		if shouldSkipURL(context.Request.URL.Path) || !logRequests {
			context.Next()
			return
		}
		start := time.Now()
		var requestBodyBuffer bytes.Buffer
		var responseBodyBuffer bytes.Buffer
		requestBodyReader := io.NopCloser(io.TeeReader(context.Request.Body, &requestBodyBuffer))
		responseBodyWriter := &ResponseBodyWriter{body: &responseBodyBuffer, ResponseWriter: context.Writer}
		context.Request.Body = requestBodyReader
		context.Writer = responseBodyWriter

		context.Next() // chain

		requestBody := requestBodyBuffer.String()
		responseBody := responseBodyWriter.body.String()

		var lastError error
		if len(context.Errors) > 0 {
			lastError = context.Errors.Last().Err
		}

		logRequest("endpoint",
			context.Request.Method,
			context.Request.URL.String(),
			context.Request.ContentLength,
			int64(context.Writer.Size()),
			context.Request.Header,
			context.Writer.Header(),
			requestBody, responseBody,
			context.Writer.Status(),
			lastError, start, logger,
		)
	}
}

func RequestLogsForClients() func(rq *http.Request, rqBody []byte, rs *http.Response, rsBody []byte, Error error, start time.Time) {
	return func(rq *http.Request, rqBody []byte, rs *http.Response, rsBody []byte, error error, start time.Time) {
		logRequest("client",
			rq.Method,
			rq.URL.String(),
			rq.ContentLength,
			rs.ContentLength,
			rq.Header,
			rs.Header,
			string(rqBody),
			string(rsBody),
			rs.StatusCode,
			error,
			start, log.Logger(),
		)
	}
}

func logRequest(message string,
	rqMethod string, rqUrl string,
	rqContentLength int64, rsContentLength int64,
	rqHeaders http.Header, rsHeaders http.Header,
	rqBody string, rsBody string, rsStatus int,
	error error, start time.Time, log *slog.Logger) {
	attrs := []any{
		slog.String("RequestMethod", rqMethod),
		slog.String("RequestURL", maskURLAuth(rqUrl)),
		slog.String("RequestContentLength", strconv.FormatInt(rqContentLength, 10)),
		slog.String("ResponseContentLength", strconv.FormatInt(rsContentLength, 10)),
		slog.String("ResponseStatus", strconv.Itoa(rsStatus)),
	}
	if error != nil {
		attrs = append(attrs, slog.String("Error", error.Error()))
	}
	attrs = append(attrs,
		slog.Duration("Duration", time.Since(start)),
		slog.Any("@RequestHeaders", headersToAttr(rqHeaders)),
		slog.Any("@ResponseHeaders", headersToAttr(rsHeaders)))
	if rqBody != "" {
		attrs = append(attrs, slog.String("@RequestPayload", rqBody))
	}
	if rsBody != "" {
		attrs = append(attrs, slog.String("@ResponsePayload", rsBody))
	}
	log.Info(message, attrs...)
}

func headersToAttr(headers http.Header) []interface{} {
	var attrs []interface{}
	for name, values := range headers {
		if maskHeaderAuth(name) {
			attrs = append(attrs, slog.String(name, "***"))
			continue
		}
		if len(values) > 0 {
			for _, value := range values {
				attrs = append(attrs, slog.String(name, value))
			}
		}
	}
	return attrs
}

func maskURLAuth(queryString string) string {
	return apiKeyRegexFilter.ReplaceAllStringFunc(queryString, func(match string) string {
		return apiKeyRegexFilter.ReplaceAllString(match, "${1}***")
	})
}

func maskHeaderAuth(headerName string) bool {
	headerName = strings.ToLower(headerName)
	return headerName == "authorization" || headerName == "sec-websocket-protocol"
}
