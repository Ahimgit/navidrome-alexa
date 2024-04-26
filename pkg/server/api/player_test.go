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

func TestPlayerAPIGetVolume(t *testing.T) {

	t.Run("GetVolume", func(t *testing.T) {
		rs := `{"volumes": [
			{"deviceSerialNumber": "dsn1", "muted": false, "volume": 11},
			{"deviceSerialNumber": "dsn2", "muted": true, "volume": 22}
		]}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))
		mockAlexaClient := new(MockAlexaClient)
		mockAlexaClient.On("GetVolume").Return(volume(), noError())

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.GetVolume(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})

	t.Run("GetVolume, client error", func(t *testing.T) {
		rs := `{"message":"mock error", "status":"error"}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))
		mockAlexaClient := new(MockAlexaClient)
		mockAlexaClient.On("GetVolume").Return(volume(), errors.New("mock error"))

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.GetVolume(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 500, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})
}

func TestPlayerAPIGetDevices(t *testing.T) {

	t.Run("GetDevices", func(t *testing.T) {
		rs := `{"devices":[{"name":"an3","deviceOwnerCustomerId":"cid3","deviceType":"dt3","serialNumber":"sn3"}]}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))
		mockAlexaClient := new(MockAlexaClient)
		mockAlexaClient.On("GetDevices").Return(devices(), noError())

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.GetDevices(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})

	t.Run("GetDevices, no devices error", func(t *testing.T) {
		rs := `{"message":"No devices on the account", "status":"error"}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))
		mockAlexaClient := new(MockAlexaClient)
		mockAlexaClient.On("GetDevices").Return(model.DevicesResponse{}, noError())

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.GetDevices(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 404, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})

	t.Run("GetDevices, client error", func(t *testing.T) {
		rs := `{"message":"mock error", "status":"error"}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))
		mockAlexaClient := new(MockAlexaClient)
		mockAlexaClient.On("GetDevices").Return(devices(), errors.New("mock error"))

		playerAPI := NewPlayerAPI(mockAlexaClient, "skill name")
		playerAPI.GetDevices(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 500, responseRecorder.Code)
		mockAlexaClient.AssertExpectations(t)
	})
}

func noError() error {
	return nil
}

func volume() model.VolumeResponse {
	return model.VolumeResponse{
		Volumes: []model.Volume{
			{
				DeviceType:    "dt1",
				Dsn:           "dsn1",
				SpeakerVolume: 11,
			},
			{
				DeviceType:    "dt1",
				Dsn:           "dsn2",
				SpeakerVolume: 22,
				SpeakerMuted:  true,
			},
		},
	}
}

func devices() model.DevicesResponse {
	return model.DevicesResponse{
		Devices: []model.Device{
			{
				AccountName:           "an1",
				Capabilities:          []string{"ANYTHING", "ANYTHING"},
				DeviceAccountId:       "dai1",
				DeviceFamily:          "df1",
				DeviceOwnerCustomerId: "cid1",
				DeviceType:            "dt1",
				SerialNumber:          "sn1",
			},
			{
				AccountName:           "an2",
				Capabilities:          []string{"AUDIO_PLAYER"},
				DeviceAccountId:       "dai2",
				DeviceFamily:          "WHA",
				DeviceOwnerCustomerId: "cid2",
				DeviceType:            "dt2",
				SerialNumber:          "sn2",
			},
			{
				AccountName:           "an3",
				Capabilities:          []string{"ANYTHING", "AUDIO_PLAYER"},
				DeviceAccountId:       "dai3",
				DeviceFamily:          "df3",
				DeviceOwnerCustomerId: "cid3",
				DeviceType:            "dt3",
				SerialNumber:          "sn3",
			},
		},
	}
}

type MockAlexaClient struct {
	mock.Mock
}

func (m *MockAlexaClient) LogIn(relog bool) (err error) {
	args := m.Called(relog)
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
