package client

import (
	"fmt"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client/httpclient"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client/model"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

const (
	headerUserAgent      = "Mozilla/5.0 (Linux; Android 13; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/112.0.5615.136 Mobile Safari/537.36"
	headerUserAgentApp   = "PitanguiBridge/2.2.527420.0-[PLATFORM=Android][MANUFACTURER=samsung][RELEASE=13][BRAND=samsung][SDK=33][MODEL=S2]"
	headerAcceptLanguage = "en-US"
)

type IAlexaClient interface {
	LogIn(relog bool) (err error)
	PostSequenceCmd(command model.AlexaCmd) (err error)
	GetDevices() (devices model.DevicesResponse, err error)
	GetVolume() (devices model.VolumeResponse, err error)
}

type AlexaClient struct {
	client       httpclient.IHttpClient
	cookieHelper httpclient.ICookieHelper
	baseDomain   string
	user         string
	password     string
	csrf         string
	retries      int
	retriesMax   int
}

func NewAlexaClient(baseDomain string, user string, password string, cookieFile string) IAlexaClient {
	return &AlexaClient{
		client:       httpclient.NewHttpClient(),
		cookieHelper: httpclient.NewCookieHelper(cookieFile),
		baseDomain:   baseDomain,
		user:         user,
		password:     password,
		retriesMax:   1,
	}
}

func NewAlexaClientWithHttpClient(baseDomain string, user string, password string,
	cookieHelper httpclient.ICookieHelper, client httpclient.IHttpClient) IAlexaClient {
	return &AlexaClient{
		client:       client,
		cookieHelper: cookieHelper,
		baseDomain:   baseDomain,
		user:         user,
		password:     password,
		retriesMax:   1,
	}
}

func (c *AlexaClient) LogIn(relog bool) (err error) {
	if relog || !c.cookieHelper.CookiesSaved() {
		if relog {
			c.client.ResetCookieJar()
		}
		if c.user == "" || c.password == "" {
			return errors.New("Alexa.LogIn no saved cookies, user and password are required but empty")
		}

		// step 0: get login form
		pageHtmlFromStep0, referer, err := getLoginForm(c.baseDomain, c.client)
		if err != nil {
			return errors.Wrap(err, "Alexa.LogIn getting form failed")
		}

		// step 1: submit login form with email w/o password
		formHtmlFromStep0 := c.cookieHelper.ExtractLoginForm(pageHtmlFromStep0)
		formDataForStep1 := c.cookieHelper.ExtractLoginFormInputs(formHtmlFromStep0)
		formDataForStep1.Add("email", c.user)
		formDataForStep1.Add("password", "")
		pageHtmlFromStep1, err := submitLoginForm(c.baseDomain, referer, formDataForStep1, c.client)
		if err != nil {
			return errors.Wrap(err, "Alexa.LogIn submit step 1 login form failed")
		}

		// step 2: submit login form with (hidden input in real form) email and password
		formHtmlFromStep1 := c.cookieHelper.ExtractLoginForm(pageHtmlFromStep1)
		formDataForStep2 := c.cookieHelper.ExtractLoginFormInputs(formHtmlFromStep1)
		formDataForStep2.Add("email", c.user)
		formDataForStep2.Add("password", c.password)
		_, err = submitLoginFormFinal(c.baseDomain, referer, formDataForStep2, c.client)
		if err != nil {
			return errors.Wrap(err, "Alexa.LogIn submit step 2 login form failed")
		}

		// get devices (sets csrf cookie) and save cookies
		_, err = c.GetDevices()
		if err != nil {
			return errors.Wrap(err, "Alexa.LogIn getting devices failed")
		}
		if err := c.cookieHelper.SaveCookies(c.client.GetCookieJar(), c.baseDomain); err != nil {
			return errors.Wrap(err, "Alexa.LogIn saving cookies failed")
		}
	} else {
		if err := c.cookieHelper.LoadCookies(c.client.GetCookieJar(), c.baseDomain); err != nil {
			return errors.Wrap(err, "Alexa.LogIn loading cookies failed")
		}
	}
	csrf := c.cookieHelper.ExtractCSRF(c.client.GetCookieJar(), c.baseDomain)
	if csrf == "" {
		return errors.New("Alexa.LogIn empty csrf cookie")
	}
	c.csrf = csrf // sets csrf param
	return nil
}

func (c *AlexaClient) PostSequenceCmd(command model.AlexaCmd) (err error) {
	if err = c.retry(func() error {
		apiUrl := fmt.Sprintf("https://alexa.%s/api/behaviors/preview", c.baseDomain)
		return c.client.RestPOST(apiUrl, buildAppHeaders(c.csrf), command, nil)
	}); err != nil {
		return errors.Wrap(err, "Alexa.PostSequenceCmd failed")
	}
	return nil
}

func (c *AlexaClient) GetDevices() (devices model.DevicesResponse, err error) {
	if err = c.retry(func() error {
		apiUrl := fmt.Sprintf("https://alexa.%s/api/devices-v2/device?cached=false", c.baseDomain)
		return c.client.RestGET(apiUrl, buildAppHeaders(c.csrf), &devices)
	}); err != nil {
		return devices, errors.Wrap(err, "Alexa.GetDevices failed")
	}
	return devices, nil
}

func (c *AlexaClient) GetVolume() (volume model.VolumeResponse, err error) {
	if err = c.retry(func() error {
		apiUrl := fmt.Sprintf("https://alexa.%s/api/devices/deviceType/dsn/audio/v1/allDeviceVolumes", c.baseDomain)
		return c.client.RestGET(apiUrl, buildAppHeaders(c.csrf), &volume)
	}); err != nil {
		return volume, errors.Wrap(err, "Alexa.GetVolume failed")
	}
	return volume, nil
}

func getLoginForm(baseDomain string, client httpclient.IHttpClient) (pageHtml string, referer string, err error) {
	formUrl := "https://www." + baseDomain + "/ap/signin" +
		"?openid.pape.max_auth_age=0" +
		"&openid.identity=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select" +
		"&accountStatusPolicy=P1" +
		"&language=en_US" +
		"&openid.return_to=https%3A%2F%2Fwww." + baseDomain + "%2Fap%2Fmaplanding" +
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
		"&openid.ns=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0" // params order matters ;(
	response, err := client.SimpleGET(formUrl, buildWebViewHeaders(referer))
	if err != nil {
		return "", "", errors.Wrap(err, "getting login form failed")
	}
	if response.Status != 200 {
		return "", "", errors.Errorf("getting login form returned wrong status: %d", response.Status)
	}
	return response.Body, formUrl, nil
}

func submitLoginForm(baseDomain string, referer string, formData *url.Values, client httpclient.IHttpClient) (pageHtml string, err error) {
	formUrl := fmt.Sprintf("https://www.%s/ap/signin", baseDomain)
	response, err := client.SimplePOST(formUrl, buildWebViewHeaders(referer), formData)
	if err != nil {
		return "", errors.Wrap(err, "submit failed")
	}
	if response.Status != 200 {
		return response.Body, errors.Errorf("submit failed, wrong status: %d, successful login submit should be a OK 200", response.Status)
	}
	return response.Body, nil
}

func submitLoginFormFinal(baseDomain string, referer string, formData *url.Values, client httpclient.IHttpClient) (pageHtml string, err error) {
	formUrl := fmt.Sprintf("https://www.%s/ap/signin", baseDomain)
	response, err := client.SimplePOST(formUrl, buildWebViewHeaders(referer), formData)
	if err != nil {
		return "", errors.Wrap(err, "submit failed")
	}
	if response.Status != 302 || response.Redirect == "" {
		return response.Body, errors.Errorf("submit failed, wrong status: %d, successful login submit should be a redirect", response.Status)
	}
	if !strings.Contains(response.Redirect, "maplanding") {
		return "", errors.Errorf("submit failed, try logining in from an app on the same network: %s", response.Redirect)
	}
	return response.Body, nil
}

func buildAppHeaders(csrf string) (headers *httpclient.Headers) {
	return &httpclient.Headers{
		{Key: "Accept", Value: "application/json; charset=utf-8"},
		{Key: "csrf", Value: csrf},
		{Key: "User-Agent", Value: headerUserAgentApp},
		{Key: "Connection", Value: "keep-alive"},
		{Key: "Upgrade-Insecure-Requests", Value: "1"},
		{Key: "Accept-Language", Value: headerAcceptLanguage},
	}
}

func buildWebViewHeaders(referer string) (headers *httpclient.Headers) {
	headersCollection := httpclient.Headers{
		{Key: "Connection", Value: "keep-alive"},
		{Key: "Cache-Control", Value: "max-age=0"},
		{Key: "Upgrade-Insecure-Requests", Value: "1"},
		{Key: "Content-Type", Value: "application/x-www-form-urlencoded"},
		{Key: "User-Agent", Value: headerUserAgent},
		{Key: "X-Requested-With", Value: "com.amazon.dee.app"},
		{Key: "Accept-Language", Value: headerAcceptLanguage},
	}
	if referer != "" {
		headersCollection = append(headersCollection, httpclient.Header{Key: "Referer", Value: referer})
	}
	return &headersCollection
}

func (c *AlexaClient) retry(retryBlock func() error) error {
	err := retryBlock()
	for httpclient.IsAuthError(err) && c.retries < c.retriesMax { // while auth error and have retries
		c.retries++
		if err = c.LogIn(true); err == nil { // re-login and call again
			err = retryBlock()
		}
	}
	if err == nil {
		c.retries = 0
	}
	return err
}
