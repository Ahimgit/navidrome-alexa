package skill

import (
	"context"
	"encoding/json"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/request"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/response"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSkillAPI(t *testing.T) {
	t.Run("SkillAPI, happy path", func(t *testing.T) {
		rq := `{
			"context": { "System": { "application": { "applicationId": "amzn1.ask.skill.xxxxx" } } },
			"request": { "type": "IntentRequest", "intent": { "name": "AMAZON.ResumeIntent" }}
		}`
		rs := `{"version":"1.0","userAgent":"na/1.0","response":{}}`

		mockGinContext, responseRecorder := mockGin(mockJSONPost(rq))
		mockRequest := mockRequestJSONObject(rq)
		mockResponse := response.NewResponseBuilder().Build()
		mockContext := log.CreateLoggerContext(mockGinContext)
		mockHandler := new(MockIHandlerSelector)
		mockHandler.On("HandleRequest", mockRequest, mockContext).Return(mockResponse)

		skillAPI := NewSkillAPI(mockHandler, "amzn1.ask.skill.xxxxx")
		skillAPI.Post(mockGinContext)

		assert.Equal(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)

		mockHandler.AssertExpectations(t)
	})

	t.Run("SkillAPI, request parse error", func(t *testing.T) {
		rq := `{`
		rs := `{"message":"unexpected EOF","status":"error"}`

		mockGinContext, responseRecorder := mockGin(mockJSONPost(rq))
		skillAPI := NewSkillAPI(nil, "amzn1.ask.skill.xxxxx")
		skillAPI.Post(mockGinContext)

		assert.Equal(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 400, responseRecorder.Code)
	})

	t.Run("SkillAPI, auth error", func(t *testing.T) {
		rq := `{
			"context": { "System": { "application": { "applicationId": "amzn1.ask.skill.xxxxx" } } },
			"request": { "type": "IntentRequest", "intent": { "name": "AMAZON.ResumeIntent" }}
		}`
		rs := `{"message":"Unauthorized","status":"error"}`

		mockGinContext, responseRecorder := mockGin(mockJSONPost(rq))

		skillAPI := NewSkillAPI(nil, "amzn1.ask.skill.yyyyy")
		skillAPI.Post(mockGinContext)

		assert.Equal(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 401, responseRecorder.Code)
	})
}

func mockRequestJSONObject(body string) *request.RequestEnvelope {
	var rq = new(request.RequestEnvelope)
	_ = json.Unmarshal([]byte(body), rq)
	return rq
}

func mockJSONPost(body string) *http.Request {
	return httptest.NewRequest("POST", "/", strings.NewReader(body))
}

func mockGin(request *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	mockGinContext, _ := gin.CreateTestContext(recorder)
	mockGinContext.Request = request
	log.InitWithLogger(slog.Default())
	log.CreateRequestContextLogger(mockGinContext)
	return mockGinContext, recorder
}

type MockIHandlerSelector struct {
	mock.Mock
}

func (m *MockIHandlerSelector) HandleRequest(rqe *request.RequestEnvelope, c context.Context) (rs *response.ResponseEnvelope) {
	args := m.Called(rqe, c)
	return args.Get(0).(*response.ResponseEnvelope)
}
