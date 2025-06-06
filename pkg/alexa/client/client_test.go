package client

import (
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client/httpclient"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func TestAlexaClientLogIn(t *testing.T) {

	t.Run("LogIn with no saved cookies, happy path", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		cookieJar := new(MockCookieJar)

		mockCookieHelper.On("CookiesSaved").Return(false)

		loginStepsSuccess(mockHttpClient, mockCookieHelper)

		expectedDomain := "example.com"
		expectedDevicesCallURL := "https://alexa.example.com/api/devices-v2/device?cached=false"
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders(""), &model.DevicesResponse{}).Return(noError())
		mockHttpClient.On("GetCookieJar").Return(cookieJar)
		mockCookieHelper.On("SaveCookies", cookieJar, expectedDomain).Return(noError())
		mockCookieHelper.On("ExtractCSRF", cookieJar, expectedDomain).Return("csrfToken")
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders("csrfToken"), &model.DevicesResponse{}).Return(noError())

		err := alexaClient.LogIn(false)
		require.NoError(t, err)
		_, err = alexaClient.GetDevices() // verify csrf token is set after login
		require.NoError(t, err)

		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with saved cookies but with re-login, happy path", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		cookieJar := new(MockCookieJar)

		mockHttpClient.On("ResetCookieJar").Return()

		loginStepsSuccess(mockHttpClient, mockCookieHelper)

		expectedDomain := "example.com"
		expectedDevicesCallURL := "https://alexa.example.com/api/devices-v2/device?cached=false"
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders(""), &model.DevicesResponse{}).Return(noError())
		mockHttpClient.On("GetCookieJar").Return(cookieJar)
		mockCookieHelper.On("SaveCookies", cookieJar, expectedDomain).Return(noError())
		mockCookieHelper.On("ExtractCSRF", cookieJar, expectedDomain).Return("csrfToken")
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders("csrfToken"), &model.DevicesResponse{}).Return(noError())

		err := alexaClient.LogIn(true)
		require.NoError(t, err)
		_, err = alexaClient.GetDevices() // verify csrf token is set after login
		require.NoError(t, err)

		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with saved cookies, happy path", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		cookieJar := new(MockCookieJar)

		expectedDomain := "example.com"
		mockCookieHelper.On("CookiesSaved").Return(true)
		mockHttpClient.On("GetCookieJar").Return(cookieJar)
		mockCookieHelper.On("LoadCookies", cookieJar, expectedDomain).Return(noError())
		mockCookieHelper.On("ExtractCSRF", cookieJar, expectedDomain).Return("csrfToken")

		err := alexaClient.LogIn(false)

		require.NoError(t, err)
		mockCookieHelper.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, get form failed", func(t *testing.T) {
		expectedError := errors.New("mock error")
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		mockCookieHelper.On("CookiesSaved").Return(false)
		mockHttpClient.On("SimpleGET", expectedGetFormURL(), expectedGetFormHeaders()).Return(nil, expectedError)

		err := alexaClient.LogIn(false)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn getting form failed: getting login form failed: mock error")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, get form failed with status code", func(t *testing.T) {
		expectedFormGetResponse := &httpclient.Response{Body: "?", Status: 401}
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		mockCookieHelper.On("CookiesSaved").Return(false)
		mockHttpClient.On("SimpleGET", expectedGetFormURL(), expectedGetFormHeaders()).Return(expectedFormGetResponse, noError())

		err := alexaClient.LogIn(false)
		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn getting form failed: getting login form returned wrong status: 401")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, submit form failed", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()

		expectedStep0PageHtml := "mockpagehtml0"
		expectedStep0FormHtml := "mockformhtml0"
		expectedStep0FormGetResponse := &httpclient.Response{Body: expectedStep0PageHtml, Status: 200}

		expectedStep1FormData := &url.Values{"email": {"testUser"}, "password": {""}}
		expectedStep1FormPostURL := "https://www.example.com/ap/signin"
		expectedStep1Error := errors.New("mock error")

		mockCookieHelper.On("CookiesSaved").Return(false)
		mockHttpClient.On("SimpleGET", expectedGetFormURL(), expectedGetFormHeaders()).Return(expectedStep0FormGetResponse, noError())
		mockCookieHelper.On("ExtractLoginForm", expectedStep0PageHtml).Return(expectedStep0FormHtml)
		mockCookieHelper.On("ExtractLoginFormInputs", expectedStep0FormHtml).Return(expectedStep1FormData)
		mockHttpClient.On("SimplePOST", expectedStep1FormPostURL, expectedPostFormHeaders(), expectedStep1FormData).Return(nil, expectedStep1Error)

		err := alexaClient.LogIn(false)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn submit step 1 login form failed: submit failed: mock error")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, submit form failed with status code", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()

		expectedStep0PageHtml := "mockpagehtml0"
		expectedStep0FormHtml := "mockformhtml0"
		expectedStep0FormGetResponse := &httpclient.Response{Body: expectedStep0PageHtml, Status: 200}

		expectedStep1PageHtml := "mockpagehtml1"
		expectedStep1FormHtml := "mockformhtml1"
		expectedStep1FormData := &url.Values{"email": {"testUser"}, "password": {""}}
		expectedStep1FormPostURL := "https://www.example.com/ap/signin"
		expectedStep1FormPostResponse := &httpclient.Response{Body: expectedStep1PageHtml, Status: 200}

		expectedStep2FormData := &url.Values{"email": {"testUser"}, "password": {"testPassword"}}
		expectedStep2FormPostURL := "https://www.example.com/ap/signin"
		expectedStep2FormPostResponse := &httpclient.Response{Status: 200, Body: "enter captcha"}

		mockCookieHelper.On("CookiesSaved").Return(false)
		// step 0
		mockHttpClient.On("SimpleGET", expectedGetFormURL(), expectedGetFormHeaders()).Return(expectedStep0FormGetResponse, noError())
		// step 1
		mockCookieHelper.On("ExtractLoginForm", expectedStep0PageHtml).Return(expectedStep0FormHtml)
		mockCookieHelper.On("ExtractLoginFormInputs", expectedStep0FormHtml).Return(expectedStep1FormData)
		mockHttpClient.On("SimplePOST", expectedStep1FormPostURL, expectedPostFormHeaders(), expectedStep1FormData).Return(expectedStep1FormPostResponse, noError())
		// step 2 fail (no redirect)
		mockCookieHelper.On("ExtractLoginForm", expectedStep1PageHtml).Return(expectedStep1FormHtml)
		mockCookieHelper.On("ExtractLoginFormInputs", expectedStep1FormHtml).Return(expectedStep2FormData)
		mockHttpClient.On("SimplePOST", expectedStep2FormPostURL, expectedPostFormHeaders(), expectedStep2FormData).Return(expectedStep2FormPostResponse, noError())

		err := alexaClient.LogIn(false)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn submit step 2 login form failed: submit failed, wrong status: 200, successful login submit should be a redirect")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, submit form failed with wrong redirect", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()

		expectedStep0PageHtml := "mockpagehtml0"
		expectedStep0FormHtml := "mockformhtml0"
		expectedStep0FormGetResponse := &httpclient.Response{Body: expectedStep0PageHtml, Status: 200}

		expectedStep1PageHtml := "mockpagehtml1"
		expectedStep1FormHtml := "mockformhtml1"
		expectedStep1FormData := &url.Values{"email": {"testUser"}, "password": {""}}
		expectedStep1FormPostURL := "https://www.example.com/ap/signin"
		expectedStep1FormPostResponse := &httpclient.Response{Body: expectedStep1PageHtml, Status: 200}

		expectedStep2FormData := &url.Values{"email": {"testUser"}, "password": {"testPassword"}}
		expectedStep2FormPostURL := "https://www.example.com/ap/signin"
		expectedStep2FormPostResponse := &httpclient.Response{Status: 302, Redirect: "wrong/redirect/url"}

		mockCookieHelper.On("CookiesSaved").Return(false)
		// step 0
		mockHttpClient.On("SimpleGET", expectedGetFormURL(), expectedGetFormHeaders()).Return(expectedStep0FormGetResponse, noError())
		// step 1
		mockCookieHelper.On("ExtractLoginForm", expectedStep0PageHtml).Return(expectedStep0FormHtml)
		mockCookieHelper.On("ExtractLoginFormInputs", expectedStep0FormHtml).Return(expectedStep1FormData)
		mockHttpClient.On("SimplePOST", expectedStep1FormPostURL, expectedPostFormHeaders(), expectedStep1FormData).Return(expectedStep1FormPostResponse, noError())
		// step 2 fail (wrong redirect)
		mockCookieHelper.On("ExtractLoginForm", expectedStep1PageHtml).Return(expectedStep1FormHtml)
		mockCookieHelper.On("ExtractLoginFormInputs", expectedStep1FormHtml).Return(expectedStep2FormData)
		mockHttpClient.On("SimplePOST", expectedStep2FormPostURL, expectedPostFormHeaders(), expectedStep2FormData).Return(expectedStep2FormPostResponse, noError())

		err := alexaClient.LogIn(false)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn submit step 2 login form failed: submit failed, try logining in from an app on the same network: wrong/redirect/url")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, get devices failed", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()

		expectedDevicesCallURL := "https://alexa.example.com/api/devices-v2/device?cached=false"
		expectedError := errors.New("mock error")

		mockCookieHelper.On("CookiesSaved").Return(false)
		loginStepsSuccess(mockHttpClient, mockCookieHelper)
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders(""), &model.DevicesResponse{}).Return(expectedError)

		err := alexaClient.LogIn(false)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn getting devices failed: Alexa.GetDevices failed: mock error")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, save cookies failed", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		cookieJar := new(MockCookieJar)

		expectedDomain := "example.com"
		expectedDevicesCallURL := "https://alexa.example.com/api/devices-v2/device?cached=false"
		expectedError := errors.New("mock error")

		mockCookieHelper.On("CookiesSaved").Return(false)
		loginStepsSuccess(mockHttpClient, mockCookieHelper)
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders(""), &model.DevicesResponse{}).Return(noError())
		mockHttpClient.On("GetCookieJar").Return(cookieJar)
		mockCookieHelper.On("SaveCookies", cookieJar, expectedDomain).Return(expectedError)

		err := alexaClient.LogIn(false)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn saving cookies failed: mock error")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("LogIn with no saved cookies, could not extract csrf", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		cookieJar := new(MockCookieJar)

		expectedDomain := "example.com"
		expectedDevicesCallURL := "https://alexa.example.com/api/devices-v2/device?cached=false"

		mockCookieHelper.On("CookiesSaved").Return(false)
		loginStepsSuccess(mockHttpClient, mockCookieHelper)

		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders(""), &model.DevicesResponse{}).Return(noError())
		mockHttpClient.On("GetCookieJar").Return(cookieJar)
		mockCookieHelper.On("SaveCookies", cookieJar, expectedDomain).Return(noError())
		mockCookieHelper.On("ExtractCSRF", cookieJar, expectedDomain).Return("")

		err := alexaClient.LogIn(false)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn empty csrf cookie")
		mockCookieHelper.AssertExpectations(t)
		mockHttpClient.AssertExpectations(t)
	})
}

func TestAlexaClientAPIs(t *testing.T) {

	t.Run("PostSequenceCmd", func(t *testing.T) {
		mockHttpClient, _, alexaClient := initClient()
		expectedURL := "https://alexa.example.com/api/behaviors/preview"
		expectedRequest := model.AlexaCmd{BehaviorID: "mockCommand"}
		mockHttpClient.
			On("RestPOST", expectedURL, expectedHeaders(""), expectedRequest, nil).
			Return(noError())

		err := alexaClient.PostSequenceCmd(expectedRequest)

		require.NoError(t, err)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("PostSequenceCmd, error", func(t *testing.T) {
		mockHttpClient, _, alexaClient := initClient()
		expectedURL := "https://alexa.example.com/api/behaviors/preview"
		expectedRequest := model.AlexaCmd{BehaviorID: "mockCommand"}
		expectedError := errors.New("mock error")
		mockHttpClient.On("RestPOST", expectedURL, expectedHeaders(""), expectedRequest, nil).Return(expectedError)

		err := alexaClient.PostSequenceCmd(expectedRequest)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.PostSequenceCmd failed: mock error")
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("GetDevices", func(t *testing.T) {
		mockHttpClient, _, alexaClient := initClient()
		expectedURL := "https://alexa.example.com/api/devices-v2/device?cached=false"
		expectedResponse := model.DevicesResponse{Devices: []model.Device{{AccountName: "device1"}}}
		mockHttpClient.
			On("RestGET", expectedURL, expectedHeaders(""), &model.DevicesResponse{}).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*model.DevicesResponse)
				*arg = expectedResponse
			}).
			Return(noError())

		actualResponse, err := alexaClient.GetDevices()

		require.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("GetDevices, error", func(t *testing.T) {
		mockHttpClient, _, alexaClient := initClient()
		expectedURL := "https://alexa.example.com/api/devices-v2/device?cached=false"
		expectedError := errors.New("mock error")
		mockHttpClient.On("RestGET", expectedURL, expectedHeaders(""), &model.DevicesResponse{}).Return(expectedError)

		_, err := alexaClient.GetDevices()

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.GetDevices failed: mock error")
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("GetDevices, auth error retry succeeds", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		cookieJar := new(MockCookieJar)

		expectedInitialError := errors.Wrap(httpclient.NewHttpErrorWithStatus(
			"mock auth error", "401 Unauthorized", 401), "mock wrap")
		expectedResponse := model.DevicesResponse{Devices: []model.Device{{AccountName: "device1"}}}

		expectedDomain := "example.com"
		expectedDevicesCallURL := "https://alexa.example.com/api/devices-v2/device?cached=false"

		// initial fail get devices with auth
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders(""), &model.DevicesResponse{}).Return(expectedInitialError).Once()

		// recover auth via re-login
		mockHttpClient.On("ResetCookieJar").Return()
		loginStepsSuccess(mockHttpClient, mockCookieHelper)
		mockHttpClient.On("RestGET", expectedDevicesCallURL, expectedHeaders(""), &model.DevicesResponse{}).Return(noError()) // get d for csrf token in login
		mockHttpClient.On("GetCookieJar").Return(cookieJar)
		mockCookieHelper.On("SaveCookies", cookieJar, expectedDomain).Return(noError())
		mockCookieHelper.On("ExtractCSRF", cookieJar, expectedDomain).Return("csrfToken")

		// recover get devices
		mockHttpClient.
			On("RestGET", expectedDevicesCallURL, expectedHeaders("csrfToken"), &model.DevicesResponse{}).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*model.DevicesResponse)
				*arg = expectedResponse
			}).
			Return(noError())

		actualResponse, err := alexaClient.GetDevices()

		require.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)

		mockHttpClient.AssertExpectations(t)
		mockCookieHelper.AssertExpectations(t)
	})

	t.Run("GetDevices, auth error retry failed", func(t *testing.T) {
		mockHttpClient, mockCookieHelper, alexaClient := initClient()
		expectedURL := "https://alexa.example.com/api/devices-v2/device?cached=false"
		expectedError1 := errors.Wrap(httpclient.NewHttpErrorWithStatus(
			"mock auth error", "401 Unauthorized", 401), "mock wrap")
		expectedError2 := errors.Wrap(httpclient.NewHttpErrorWithStatus(
			"mock auth error in re-login", "401 Unauthorized", 401), "mock wrap")

		mockHttpClient.On("RestGET", expectedURL, expectedHeaders(""), &model.DevicesResponse{}).Return(expectedError1)
		mockHttpClient.On("ResetCookieJar").Return()
		mockHttpClient.On("SimpleGET", expectedGetFormURL(), expectedGetFormHeaders()).Return(nil, expectedError2)

		_, err := alexaClient.GetDevices()

		mockHttpClient.AssertExpectations(t)
		mockCookieHelper.AssertExpectations(t)

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.LogIn getting form failed: getting login form failed: mock wrap: mock auth error in re-login")
	})

	t.Run("GetVolume", func(t *testing.T) {
		mockHttpClient, _, alexaClient := initClient()
		expectedURL := "https://alexa.example.com/api/devices/deviceType/dsn/audio/v1/allDeviceVolumes"
		expectedResponse := model.VolumeResponse{Volumes: []model.Volume{{Dsn: "device1"}}}
		mockHttpClient.
			On("RestGET", expectedURL, expectedHeaders(""), &model.VolumeResponse{}).
			Run(func(args mock.Arguments) {
				arg := args.Get(2).(*model.VolumeResponse)
				*arg = expectedResponse
			}).
			Return(noError())

		actualResponse, err := alexaClient.GetVolume()

		require.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)
		mockHttpClient.AssertExpectations(t)
	})

	t.Run("GetVolume, error", func(t *testing.T) {
		mockHttpClient, _, alexaClient := initClient()
		expectedURL := "https://alexa.example.com/api/devices/deviceType/dsn/audio/v1/allDeviceVolumes"
		expectedError := errors.New("mock error")
		mockHttpClient.On("RestGET", expectedURL, expectedHeaders(""), &model.VolumeResponse{}).Return(expectedError)

		_, err := alexaClient.GetVolume()

		require.Error(t, err)
		assert.ErrorContains(t, err, "Alexa.GetVolume failed: mock error")
		mockHttpClient.AssertExpectations(t)
	})

}

func initClient() (mockHttpClient *MockIHttpClient, mockCookieHelper *MockICookieHelper, alexaClient IAlexaClient) {
	mockHttpClient = new(MockIHttpClient)
	mockCookieHelper = new(MockICookieHelper)
	alexaClient = NewAlexaClientWithHttpClient(
		"example.com", "testUser", "testPassword",
		mockCookieHelper,
		mockHttpClient)
	return mockHttpClient, mockCookieHelper, alexaClient
}

func noError() error {
	return nil
}

func expectedGetFormURL() string {
	return "https://www.example.com/ap/signin" +
		"?openid.pape.max_auth_age=0" +
		"&openid.identity=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select" +
		"&accountStatusPolicy=P1" +
		"&language=en_US" +
		"&openid.return_to=https%3A%2F%2Fwww.example.com%2Fap%2Fmaplanding" +
		"&openid.assoc_handle=amzn_dp_project_dee_android" +
		"&openid.oa2.response_type=code" +
		"&openid.mode=checkid_setup" +
		"&openid.ns.pape=http%3A%2F%2Fspecs.openid.net%2Fextensions%2Fpape%2F1.0" +
		"&openid.oa2.code_challenge_method=S256" +
		"&openid.ns.oa2=http%3A%2F%2Fwww.amazon.com%2Fap%2Fext%2Foauth%2F2" +
		"&openid.oa2.code_challenge=" +
		"&openid.oa2.scope=device_auth_access" +
		"&openid.claimed_id=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select" +
		"&openid.oa2.client_id=" +
		"&disableLoginPrepopulate=0" +
		"&openid.ns=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0"
}

func expectedGetFormHeaders() *httpclient.Headers {
	return &httpclient.Headers{
		httpclient.Header{Key: "Connection", Value: "keep-alive"},
		httpclient.Header{Key: "Cache-Control", Value: "max-age=0"},
		httpclient.Header{Key: "Upgrade-Insecure-Requests", Value: "1"},
		httpclient.Header{Key: "Content-Type", Value: "application/x-www-form-urlencoded"},
		httpclient.Header{Key: "User-Agent", Value: "Mozilla/5.0 (Linux; Android 13; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/112.0.5615.136 Mobile Safari/537.36"},
		httpclient.Header{Key: "X-Requested-With", Value: "com.amazon.dee.app"},
		httpclient.Header{Key: "Accept-Language", Value: "en-US"},
	}
}

func expectedPostFormHeaders() *httpclient.Headers {
	return &httpclient.Headers{
		httpclient.Header{Key: "Connection", Value: "keep-alive"},
		httpclient.Header{Key: "Cache-Control", Value: "max-age=0"},
		httpclient.Header{Key: "Upgrade-Insecure-Requests", Value: "1"},
		httpclient.Header{Key: "Content-Type", Value: "application/x-www-form-urlencoded"},
		httpclient.Header{Key: "User-Agent", Value: "Mozilla/5.0 (Linux; Android 13; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/112.0.5615.136 Mobile Safari/537.36"}, httpclient.Header{Key: "X-Requested-With", Value: "com.amazon.dee.app"},
		httpclient.Header{Key: "Accept-Language", Value: "en-US"},
		httpclient.Header{Key: "Referer", Value: expectedGetFormURL()}}
}

func expectedHeaders(csrf string) *httpclient.Headers {
	return &httpclient.Headers{
		httpclient.Header{Key: "Accept", Value: "application/json; charset=utf-8"},
		httpclient.Header{Key: "csrf", Value: csrf},
		httpclient.Header{Key: "User-Agent", Value: "PitanguiBridge/2.2.527420.0-[PLATFORM=Android][MANUFACTURER=samsung][RELEASE=13][BRAND=samsung][SDK=33][MODEL=S2]"},
		httpclient.Header{Key: "Connection", Value: "keep-alive"},
		httpclient.Header{Key: "Upgrade-Insecure-Requests", Value: "1"},
		httpclient.Header{Key: "Accept-Language", Value: "en-US"}}
}

func loginStepsSuccess(mockHttpClient *MockIHttpClient, mockCookieHelper *MockICookieHelper) {
	expectedStep0PageHtml := "mockpagehtml0"
	expectedStep0FormHtml := "mockformhtml0"
	expectedStep0FormGetResponse := &httpclient.Response{Body: expectedStep0PageHtml, Status: 200}

	expectedStep1PageHtml := "mockpagehtml1"
	expectedStep1FormHtml := "mockformhtml1"
	expectedStep1FormData := &url.Values{"email": {"testUser"}, "password": {""}}
	expectedStep1FormPostURL := "https://www.example.com/ap/signin"
	expectedStep1FormPostResponse := &httpclient.Response{Body: expectedStep1PageHtml, Status: 200}

	expectedStep2FormData := &url.Values{"email": {"testUser"}, "password": {"testPassword"}}
	expectedStep2FormPostURL := "https://www.example.com/ap/signin"
	expectedStep2FormPostResponse := &httpclient.Response{Status: 302, Redirect: "https://www.example.com/maplanding"}

	// step 0
	mockHttpClient.On("SimpleGET", expectedGetFormURL(), expectedGetFormHeaders()).Return(expectedStep0FormGetResponse, noError())
	// step 1
	mockCookieHelper.On("ExtractLoginForm", expectedStep0PageHtml).Return(expectedStep0FormHtml)
	mockCookieHelper.On("ExtractLoginFormInputs", expectedStep0FormHtml).Return(expectedStep1FormData)
	mockHttpClient.On("SimplePOST", expectedStep1FormPostURL, expectedPostFormHeaders(), expectedStep1FormData).Return(expectedStep1FormPostResponse, noError())
	// step 2
	mockCookieHelper.On("ExtractLoginForm", expectedStep1PageHtml).Return(expectedStep1FormHtml)
	mockCookieHelper.On("ExtractLoginFormInputs", expectedStep1FormHtml).Return(expectedStep2FormData)
	mockHttpClient.On("SimplePOST", expectedStep2FormPostURL, expectedPostFormHeaders(), expectedStep2FormData).Return(expectedStep2FormPostResponse, noError())
}

// todo consider codegen for mocks

type MockCookieJar struct {
	mock.Mock
}

func (m *MockCookieJar) Cookies(u *url.URL) []*http.Cookie {
	args := m.Called(u)
	return args.Get(0).([]*http.Cookie)
}

func (m *MockCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	m.Called(u, cookies)
}

type MockIHttpClient struct {
	mock.Mock
}

func (m *MockIHttpClient) GetCookieJar() (jar http.CookieJar) {
	args := m.Called()
	return args.Get(0).(http.CookieJar)
}

func (m *MockIHttpClient) ResetCookieJar() {
	m.Called()
}

func (m *MockIHttpClient) SimpleGET(url string, headers *httpclient.Headers) (*httpclient.Response, error) {
	args := m.Called(url, headers)
	ret1 := args.Get(0)
	ret2 := args.Error(1)
	if ret1 == nil {
		return nil, ret2
	}
	return ret1.(*httpclient.Response), ret2
}

func (m *MockIHttpClient) SimplePOST(url string, headers *httpclient.Headers, formData *url.Values) (*httpclient.Response, error) {
	args := m.Called(url, headers, formData)
	ret1 := args.Get(0)
	ret2 := args.Error(1)
	if ret1 == nil {
		return nil, ret2
	}
	return ret1.(*httpclient.Response), ret2
}

func (m *MockIHttpClient) RestGET(url string, headers *httpclient.Headers, response interface{}) error {
	args := m.Called(url, headers, response)
	return args.Error(0)
}

func (m *MockIHttpClient) RestPOST(url string, headers *httpclient.Headers, request interface{}, response interface{}) error {
	args := m.Called(url, headers, request, response)
	return args.Error(0)
}

type MockICookieHelper struct {
	mock.Mock
}

func (m *MockICookieHelper) CookiesSaved() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockICookieHelper) SaveCookies(jar http.CookieJar, baseDomain string) error {
	args := m.Called(jar, baseDomain)
	return args.Error(0)
}

func (m *MockICookieHelper) LoadCookies(jar http.CookieJar, baseDomain string) error {
	args := m.Called(jar, baseDomain)
	return args.Error(0)
}

func (m *MockICookieHelper) ExtractCSRF(jar http.CookieJar, baseDomain string) (csrf string) {
	args := m.Called(jar, baseDomain)
	return args.String(0)
}

func (m *MockICookieHelper) ExtractLoginForm(pageHtml string) (formHtml string) {
	args := m.Called(pageHtml)
	return args.String(0)
}

func (m *MockICookieHelper) ExtractLoginFormInputs(formHtml string) *url.Values {
	args := m.Called(formHtml)
	return args.Get(0).(*url.Values)
}
