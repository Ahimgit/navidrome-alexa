package api

import (
	"fmt"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client/model"
	"github.com/ahimgit/navidrome-alexa/pkg/util/tests"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type TestCase struct {
	command string
	run     func(playerAPI *PlayerAPI, gin *gin.Context)
}

func TestPostPlayerTextCommands(t *testing.T) {

	testcases := []TestCase{
		{"play", func(playerAPI *PlayerAPI, gin *gin.Context) { playerAPI.PostPlay(gin) }},
		{"stop", func(playerAPI *PlayerAPI, gin *gin.Context) { playerAPI.PostStop(gin) }},
		{"next", func(playerAPI *PlayerAPI, gin *gin.Context) { playerAPI.PostNext(gin) }},
		{"previous", func(playerAPI *PlayerAPI, gin *gin.Context) { playerAPI.PostPrev(gin) }},
	}

	for _, testCase := range testcases {
		t.Run(fmt.Sprintf("%s with correct request", testCase.command), func(t *testing.T) {
			rq := `{"deviceOwnerCustomerId": "cid", "deviceType": "dt", "serialNumber": "sn"}`
			rs := fmt.Sprintf(`{"message": "%s executed", "status": "success"}`, testCase.command)
			expectedText := fmt.Sprintf(`ask skill name to %s`, testCase.command)
			expectedCommand := model.BuildTextCommandCmd(expectedText, "en-US", "dt", "sn", "cid")

			mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))
			mockAlexaClient := new(MockAlexaClient)
			mockAlexaClient.On("PostSequenceCmd", expectedCommand).Return(noError())

			playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
			testCase.run(playerAPI, mockGinContext)

			assert.JSONEq(t, rs, responseRecorder.Body.String())
			assert.Equal(t, 200, responseRecorder.Code)
			mockAlexaClient.AssertExpectations(t)
		})
	}

	for _, testCase := range testcases {
		t.Run(fmt.Sprintf("%s with alexa client error", testCase.command), func(t *testing.T) {
			rq := `{"deviceOwnerCustomerId": "cid", "deviceType": "dt", "serialNumber": "sn"}`
			rs := `{"message":"mock error", "status":"error"}`
			expectedText := fmt.Sprintf(`ask skill name to %s`, testCase.command)
			expectedCommand := model.BuildTextCommandCmd(expectedText, "en-US", "dt", "sn", "cid")

			mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))
			mockAlexaClient := new(MockAlexaClient)
			mockAlexaClient.On("PostSequenceCmd", expectedCommand).Return(errors.New("mock error"))

			playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
			testCase.run(playerAPI, mockGinContext)

			assert.JSONEq(t, rs, responseRecorder.Body.String())
			assert.Equal(t, 500, responseRecorder.Code)
			mockAlexaClient.AssertExpectations(t)
		})
	}

	for _, testCase := range testcases {
		t.Run(fmt.Sprintf("%s with invalid request", testCase.command), func(t *testing.T) {
			rq := `{`
			rs := `{"message":"unexpected EOF", "status":"error"}`

			mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))
			mockAlexaClient := new(MockAlexaClient)

			playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
			testCase.run(playerAPI, mockGinContext)

			assert.JSONEq(t, rs, responseRecorder.Body.String())
			assert.Equal(t, 400, responseRecorder.Code)
			mockAlexaClient.AssertExpectations(t)
		})
	}

}

func TestPostPlayerVolumeCommand(t *testing.T) {
	t.Run("PostVolume with correct request", func(t *testing.T) {
		rq := `{
			"device": {"deviceOwnerCustomerId": "cid", "deviceType": "dt", "serialNumber": "sn"},
		    "volume": 31
		}`
		rs := `{"message": "volume updated", "status": "success"}`

		expectedCommand := model.BuildVolumeCmd(31, "en-US", "dt", "sn", "cid")

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))
		mockAlexaClient := new(MockAlexaClient)
		mockAlexaClient.On("PostSequenceCmd", expectedCommand).Return(noError())

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.PostVolume(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})

	t.Run("PostVolume with alexa client error", func(t *testing.T) {
		rq := `{
			"device": {"deviceOwnerCustomerId": "cid", "deviceType": "dt", "serialNumber": "sn"},
		    "volume": 31
		}`
		rs := `{"message":"mock error", "status":"error"}`

		expectedCommand := model.BuildVolumeCmd(31, "en-US", "dt", "sn", "cid")

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))
		mockAlexaClient := new(MockAlexaClient)
		mockAlexaClient.On("PostSequenceCmd", expectedCommand).Return(errors.New("mock error"))

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.PostVolume(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 500, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})

	t.Run("PostVolume with invalid request", func(t *testing.T) {
		rq := `{`
		rs := `{"message":"unexpected EOF", "status":"error"}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))
		mockAlexaClient := new(MockAlexaClient)

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.PostVolume(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 400, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})
}

type MockAlexaClient struct {
	mock.Mock
}

func (m *MockAlexaClient) LogIn() (err error) {
	args := m.Called()
	return args.Error(1)
}

func (m *MockAlexaClient) PostSequenceCmd(command model.AlexaCmd) (err error) {
	args := m.Called(command)
	return args.Error(0)
}

func (m *MockAlexaClient) GetDevices() (devices model.DevicesResponse, err error) {
	args := m.Called()
	ret1 := args.Get(0)
	ret2 := args.Error(1)
	return ret1.(model.DevicesResponse), ret2
}

func (m *MockAlexaClient) GetVolume() (devices model.VolumeResponse, err error) {
	args := m.Called()
	ret1 := args.Get(0)
	ret2 := args.Error(1)
	return ret1.(model.VolumeResponse), ret2
}

func noError() error {
	return nil
}
