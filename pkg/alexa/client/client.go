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
	LogIn() (err error)
	PostSequenceCmd(command model.AlexaCmd) (err error)
	GetDevices() (devices model.DevicesResponse, err error)
}

type AlexaClient struct {
	client       httpclient.IHttpClient
	cookieHelper httpclient.ICookieHelper
	baseDomain   string
	user         string
	password     string
	csrf         string
}

func NewAlexaClient(baseDomain string, user string, password string, cookieFile string) IAlexaClient {
	return &AlexaClient{
		client:       httpclient.NewHttpClient(),
		cookieHelper: httpclient.NewCookieHelper(cookieFile),
		baseDomain:   baseDomain,
		user:         user,
		password:     password,
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
	}
}

func (c *AlexaClient) LogIn() (err error) {
	if !c.cookieHelper.CookiesSaved() {
		if c.user == "" || c.password == "" {
			return errors.New("Alexa.LogIn no saved cookies, user and password are required but empty")
		}
		// get login form
		formBody, referer, err := getLoginForm(c.baseDomain, c.client)
		if err != nil {
			return errors.Wrap(err, "Alexa.LogIn getting form failed")
		}
		// submit login form
		formData := c.cookieHelper.ExtractLoginFormInputsCSRF(formBody)
		formData.Add("email", c.user)
		formData.Add("password", c.password)
		if err = submitLoginForm(c.baseDomain, referer, formData, c.client); err != nil {
			return errors.Wrap(err, "Alexa.LogIn submit form failed")
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
	c.csrf = c.cookieHelper.ExtractCSRF(c.client.GetCookieJar(), c.baseDomain) // set csrf param
	return nil
}

func (c *AlexaClient) PostSequenceCmd(command model.AlexaCmd) (err error) {
	apiUrl := fmt.Sprintf("https://alexa.%s/api/behaviors/preview", c.baseDomain)
	if err = c.client.RestPOST(apiUrl, buildAppHeaders(c.csrf), command, nil); err != nil {
		return errors.Wrap(err, "Alexa.PostSequenceCmd failed")
	}
	return nil
}

func (c *AlexaClient) GetDevices() (devices model.DevicesResponse, err error) {
	apiUrl := fmt.Sprintf("https://alexa.%s/api/devices-v2/device?cached=false", c.baseDomain)
	if err = c.client.RestGET(apiUrl, buildAppHeaders(c.csrf), &devices); err != nil {
		return devices, errors.Wrap(err, "Alexa.GetDevices failed")
	}
	return devices, nil
}

func getLoginForm(baseDomain string, client httpclient.IHttpClient) (formBody string, referer string, err error) {
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

func submitLoginForm(baseDomain string, referer string, formData *url.Values, client httpclient.IHttpClient) (err error) {
	formUrl := fmt.Sprintf("https://www.%s/ap/signin", baseDomain)
	response, err := client.SimplePOST(formUrl, buildWebViewHeaders(referer), formData)
	if err != nil {
		return errors.Wrapf(err, "submit failed: %d", response.Status)
	}
	if response.Status != 302 || response.Redirect == "" {
		return errors.Errorf("submit failed, wrong status: %d, successful login submit should be a redirect", response.Status)
	}
	if !strings.Contains(response.Redirect, "maplanding") {
		return errors.Errorf("submit failed, try logining in form an app on the same network: %s", response.Redirect)
	}
	return nil
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
