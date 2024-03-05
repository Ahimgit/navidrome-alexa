package tests

import (
	"encoding/json"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/request"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
)

func MockRequestJSONObject(body string) *request.RequestEnvelope {
	var rq = new(request.RequestEnvelope)
	_ = json.Unmarshal([]byte(body), rq)
	return rq
}

func MockJSONPost(body string) *http.Request {
	return httptest.NewRequest("POST", "/", strings.NewReader(body))
}

func MockGin(request *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	mockGinContext, _ := gin.CreateTestContext(recorder)
	mockGinContext.Request = request
	log.InitWithLogger(slog.Default())
	log.CreateRequestContextLogger(mockGinContext)
	return mockGinContext, recorder
}
