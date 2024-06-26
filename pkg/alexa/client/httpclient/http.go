package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"
)

type Header struct {
	Key   string
	Value string
}

type Headers []Header

type Response struct {
	Status   int
	Body     string
	Redirect string
}

type IHttpClient interface {
	RestGET(url string, rqHeaders *Headers, rs any) (err error)          // rs by ref (no generic in types)
	RestPOST(url string, rqHeaders *Headers, rq any, rs any) (err error) // rs by ref (no generic in types)
	SimpleGET(url string, rqHeaders *Headers) (rs *Response, err error)
	SimplePOST(url string, rqHeaders *Headers, formData *url.Values) (rs *Response, err error)
	GetCookieJar() (jar http.CookieJar)
	ResetCookieJar()
}

type HttpClient struct {
	*http.Client
	requestLogger  func(rq *http.Request, rqBody []byte)
	responseLogger func(rq *http.Request, rqBody []byte, rs *http.Response, rsBody []byte, err error, start time.Time)
}

func NewHttpClient() *HttpClient {
	client := &HttpClient{
		requestLogger:  func(rq *http.Request, rqBody []byte) {},
		responseLogger: func(rq *http.Request, rqBody []byte, rs *http.Response, rsBody []byte, err error, start time.Time) {},
		Client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	client.ResetCookieJar()
	return client
}

func (httpClient *HttpClient) ResetCookieJar() {
	jar, _ := cookiejar.New(nil)
	httpClient.Client.Jar = jar
}

func (httpClient *HttpClient) WithTimeout(duration time.Duration) *HttpClient {
	httpClient.Timeout = duration
	return httpClient
}

func (httpClient *HttpClient) WithRequestLogger(f func(rq *http.Request, rqBody []byte)) *HttpClient {
	httpClient.requestLogger = f
	return httpClient
}

func (httpClient *HttpClient) WithResponseLogger(f func(rq *http.Request, rqBody []byte, rs *http.Response, rsBody []byte, err error, start time.Time)) *HttpClient {
	httpClient.responseLogger = f
	return httpClient
}

func (httpClient *HttpClient) RestGET(url string, rqHeaders *Headers, rs any) (err error) {
	rsBytes, _, err := httpClient.runHttpRequest("GET", url, rqHeaders, nil)
	if err != nil {
		return err
	}
	if rs == nil {
		return nil // not expecting rs, don't unmarshal
	}
	if err = json.Unmarshal(rsBytes, rs); err != nil {
		return NewHttpError("error unmarshalling rs: "+string(rsBytes), err)
	}
	return nil
}

func (httpClient *HttpClient) RestPOST(url string, rqHeaders *Headers, rq any, rs any) (err error) {
	rqBytes, err := json.Marshal(rq)
	if err != nil {
		return NewHttpError("error marshalling rq", err)
	}
	rsBytes, _, err := httpClient.runHttpRequest("POST", url, rqHeaders, rqBytes)
	if err != nil {
		return err
	}
	if rs == nil {
		return nil // not expecting rs, don't unmarshal
	}
	if err = json.Unmarshal(rsBytes, rs); err != nil {
		return NewHttpError("error unmarshalling rs: "+string(rsBytes), err)
	}
	return nil
}

func (httpClient *HttpClient) SimpleGET(url string, rqHeaders *Headers) (rs *Response, err error) {
	responseBytes, httpResponse, err := httpClient.runHttpRequest("GET", url, rqHeaders, nil)
	if err != nil {
		return nil, err
	}
	return &Response{
		Status:   httpResponse.StatusCode,
		Body:     string(responseBytes),
		Redirect: httpResponse.Header.Get("Location"),
	}, nil
}

func (httpClient *HttpClient) SimplePOST(url string, rqHeaders *Headers, formData *url.Values) (rs *Response, err error) {
	responseBytes, httpResponse, err := httpClient.runHttpRequest("POST", url, rqHeaders, []byte(formData.Encode()))
	if err != nil {
		return nil, err
	}
	return &Response{
		Status:   httpResponse.StatusCode,
		Body:     string(responseBytes),
		Redirect: httpResponse.Header.Get("Location"),
	}, nil
}

func (httpClient *HttpClient) GetCookieJar() http.CookieJar {
	return httpClient.Jar
}

func (httpClient *HttpClient) runHttpRequest(
	rqMethod string,
	rqURL string,
	rqHeaders *Headers,
	rqBody []byte,
) (rsBody []byte, rs *http.Response, err error) {
	startTime := time.Now()
	var rq *http.Request
	if rqBody == nil {
		rq, err = http.NewRequest(rqMethod, rqURL, nil)
	} else {
		rq, err = http.NewRequest(rqMethod, rqURL, bytes.NewBuffer(rqBody))
	}
	if err != nil {
		return nil, nil, NewHttpError("error creating http request", err)
	}
	if rqHeaders != nil {
		for _, header := range *rqHeaders {
			rq.Header.Set(header.Key, header.Value)
		}
	}

	httpClient.requestLogger(rq, rqBody)
	rs, err = httpClient.Client.Do(rq)
	if err != nil {
		err = NewHttpError("error doing http call", err)
		httpClient.responseLogger(rq, rqBody, nil, nil, err, startTime)
		return nil, nil, err
	}
	if rs.StatusCode >= 400 {
		err = NewHttpErrorWithStatus("error status code "+strconv.Itoa(rs.StatusCode), rs.Status, rs.StatusCode)
		httpClient.responseLogger(rq, rqBody, rs, nil, nil, startTime) // err not propagated
		return nil, rs, err
	}
	rsBody, err = io.ReadAll(rs.Body)
	defer nopClose(rs.Body)
	if err != nil {
		err = NewHttpErrorWithStatus("error reading response body", rs.Status, rs.StatusCode)
		httpClient.responseLogger(rq, rqBody, rs, rsBody, err, startTime)
		return nil, rs, err
	}
	httpClient.responseLogger(rq, rqBody, rs, rsBody, nil, startTime)
	return rsBody, rs, nil
}

func nopClose(closer io.Closer) {
	_ = closer.Close()
}
