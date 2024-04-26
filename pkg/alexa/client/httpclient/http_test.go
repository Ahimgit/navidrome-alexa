package httpclient

import (
	"fmt"
	"github.com/h2non/gock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type TestRQ struct {
	RqField1 string `json:"rqField1"`
	RqField2 int    `json:"rqField2"`
}

type TestRS struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func TestSimpleGET(t *testing.T) {
	client := NewHttpClient()

	t.Run("simple GET request is made to a mocked 200 OK", func(t *testing.T) {
		gock.New("http://dummy").Get("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			Reply(200).
			BodyString("simple body1")
		defer gock.Off()

		rs, err := client.SimpleGET("http://dummy/url?param=test1", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		})

		require.NoError(t, err)
		assert.NotNil(t, rs)
		assert.Equal(t, 200, rs.Status)
		assert.Equal(t, "simple body1", rs.Body)
		assert.Empty(t, rs.Redirect)
	})

	t.Run("simple GET request is made to a mocked 301 Moved Permanently", func(t *testing.T) {
		gock.New("http://dummy").Get("/url").
			MatchParam("param", "test2").
			Reply(301).
			BodyString("simple body2").
			AddHeader("Location", "https://redirect?with=param")
		defer gock.Off()

		rs, err := client.SimpleGET("http://dummy/url?param=test2", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		})

		require.NoError(t, err)
		assert.NotNil(t, rs)
		assert.Equal(t, 301, rs.Status)
		assert.Equal(t, "simple body2", rs.Body)
		assert.Equal(t, "https://redirect?with=param", rs.Redirect)
	})

	t.Run("simple GET request is made to a mocked error", func(t *testing.T) {
		gock.New("http://dummy").Get("/url").
			MatchParam("param", "test3").
			ReplyError(errors.New("mock error"))
		defer gock.Off()

		rs, err := client.SimpleGET("http://dummy/url?param=test3", nil)

		assert.Nil(t, rs)
		assert.Error(t, err)
		assert.Equal(t, `error doing http call: Get "http://dummy/url?param=test3": mock error`, err.Error())
	})
}

func TestSimplePOST(t *testing.T) {
	client := NewHttpClient()

	t.Run("simple POST request is made to a mocked 301 Moved Permanently", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			BodyString("key+1=value+1&key+2=value+2").
			Reply(301).
			BodyString("simple body1").
			AddHeader("Location", "https://redirect?with=param")
		defer gock.Off()

		headers := &Headers{{Key: "X-Header1", Value: "header-value-1"}}
		formData := &url.Values{}
		formData.Add("key 1", "value 1")
		formData.Add("key 2", "value 2")
		rs, err := client.SimplePOST("http://dummy/url?param=test1", headers, formData)

		require.NoError(t, err)
		assert.NotNil(t, rs)
		assert.Equal(t, 301, rs.Status)
		assert.Equal(t, "simple body1", rs.Body)
		assert.Equal(t, "https://redirect?with=param", rs.Redirect)
	})

	t.Run("simple POST request is made to a mocked error", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test2").
			ReplyError(errors.New("mock error"))
		defer gock.Off()

		headers := &Headers{{Key: "X-Header1", Value: "header-value-1"}}
		formData := &url.Values{}
		formData.Add("key 1", "value 1")
		formData.Add("key 2", "value 2")
		rs, err := client.SimplePOST("http://dummy/url?param=test2", headers, formData)

		assert.Nil(t, rs)
		assert.Error(t, err)
		assert.Equal(t, `error doing http call: Post "http://dummy/url?param=test2": mock error`, err.Error())
	})
}

func TestRestGET(t *testing.T) {
	client := NewHttpClient()

	t.Run("REST GET request is made to a mocked 200 OK", func(t *testing.T) {
		gock.New("http://dummy").Get("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			Reply(200).
			JSON(`{"field1":"val1","field2":321}`)
		defer gock.Off()

		var rs TestRS
		err := client.RestGET("http://dummy/url?param=test1", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, &rs)

		require.NoError(t, err)
		assert.Equal(t, TestRS{
			Field1: "val1",
			Field2: 321,
		}, rs)
	})

	t.Run("REST GET request is made to a mocked invalid JSON response", func(t *testing.T) {
		gock.New("http://dummy").Get("/url").
			MatchParam("param", "test2").
			MatchHeader("X-Header1", "header-value-1").
			Reply(200).
			JSON(`{"field1":"val1","field2":321`)
		defer gock.Off()

		var rs TestRS
		err := client.RestGET("http://dummy/url?param=test2", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, &rs)

		assert.Error(t, err)
		assert.Equal(t, `error unmarshalling rs: {"field1":"val1","field2":321: unexpected end of JSON input`, err.Error())
	})

	t.Run("REST GET request is made to a mocked empty body 200 OK response", func(t *testing.T) {
		gock.New("http://dummy").Get("/url").
			MatchParam("param", "test3").
			MatchHeader("X-Header1", "header-value-1").
			Reply(200)
		defer gock.Off()

		err := client.RestGET("http://dummy/url?param=test3", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, nil)

		require.NoError(t, err)
	})

	t.Run("REST GET request is made to a mocked 400 Bad Request", func(t *testing.T) {
		gock.New("http://dummy").Get("/url").
			MatchParam("param", "test4").
			MatchHeader("X-Header1", "header-value-1").
			Reply(400)
		defer gock.Off()

		var rs TestRS
		err := client.RestGET("http://dummy/url?param=test4", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, &rs)

		assert.Error(t, err)
		assert.Equal(t, "error status code 400", err.Error())
		var httpError *HttpError
		if errors.As(err, &httpError) {
			assert.Equal(t, 400, httpError.StatusCode)
			assert.Equal(t, "400 Bad Request", httpError.Status)
		} else {
			assert.Fail(t, "Error isn't of type HttpError")
		}
	})
}

func TestRestPOST(t *testing.T) {
	client := NewHttpClient()

	t.Run("REST POST request is made to a mocked 200 OK", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
			Reply(200).
			JSON(`{"field1":"val1","field2":321}`)
		defer gock.Off()

		rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
		var rs TestRS
		err := client.RestPOST("http://dummy/url?param=test1", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, rq, &rs)

		require.NoError(t, err)
		assert.Equal(t, TestRS{
			Field1: "val1",
			Field2: 321,
		}, rs)
	})

	t.Run("REST POST request is made to a mocked invalid JSON response", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test2").
			MatchHeader("X-Header1", "header-value-1").
			BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
			Reply(200).
			JSON(`{"field1":"val1","field2":321`)
		defer gock.Off()

		var rs TestRS
		rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
		err := client.RestPOST("http://dummy/url?param=test2", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, rq, &rs)

		assert.Error(t, err)
		assert.Equal(t, `error unmarshalling rs: {"field1":"val1","field2":321: unexpected end of JSON input`, err.Error())
	})

	t.Run("REST POST request is made with invalid request", func(t *testing.T) {
		badRequest := func() {}

		var rs TestRS
		err := client.RestPOST("http://dummy/url?param=test2", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, badRequest, &rs)

		assert.Error(t, err)
		assert.Equal(t, `error marshalling rq: json: unsupported type: func()`, err.Error())
	})

	t.Run("REST POST request is made to a mocked empty body 200 OK response", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
			Reply(200)
		defer gock.Off()

		rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
		err := client.RestPOST("http://dummy/url?param=test1", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, rq, nil)

		require.NoError(t, err)
	})

	t.Run("REST POST request is made to a mocked 400 Bad Request", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
			Reply(400)
		defer gock.Off()

		var rs TestRS
		rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
		err := client.RestPOST("http://dummy/url?param=test1", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, rq, &rs)

		assert.Error(t, err)
		var httpError *HttpError
		if errors.As(err, &httpError) {
			fmt.Println(httpError.StatusCode)
			assert.Equal(t, 400, httpError.StatusCode)
			assert.Equal(t, "400 Bad Request", httpError.Status)
		} else {
			assert.Fail(t, "Error isn't of type HttpError")
		}
		assert.Equal(t, "error status code 400", err.Error())
	})
}

func TestRequestLogging(t *testing.T) {
	var (
		lastRq       *http.Request
		lastRqBody   []byte
		lastRsRq     *http.Request
		lastRsRqBody []byte
		lastRs       *http.Response
		lastRsBody   []byte
		lastRsErr    error
		lastRsTime   time.Time
	)

	client := NewHttpClient().
		WithRequestLogger(func(rq *http.Request, rqBody []byte) {
			lastRq = rq
			lastRqBody = rqBody
		}).
		WithResponseLogger(func(rq *http.Request, rqBody []byte, rs *http.Response, rsBody []byte, err error, start time.Time) {
			lastRsRq = rq
			lastRsRqBody = rqBody
			lastRs = rs
			lastRsBody = rsBody
			lastRsErr = err
			lastRsTime = start
		})

	t.Run("request is performed using a client that has rq/rs logging configured", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
			Reply(200).
			JSON(`{"field1":"val1","field2":321}`)
		defer gock.Off()

		var rs TestRS
		rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
		err := client.RestPOST("http://dummy/url?param=test1", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, rq, &rs)

		require.NoError(t, err)
		assert.NotNil(t, rs)

		assert.NotNil(t, lastRq)
		assert.Equal(t, "http://dummy/url?param=test1", lastRq.URL.String())
		assert.Equal(t, "header-value-1", lastRq.Header.Get("X-Header1"))
		assert.Equal(t, `{"rqField1":"rqVal1","rqField2":567}`, string(lastRqBody))

		assert.Equal(t, lastRq, lastRsRq)
		assert.Equal(t, lastRqBody, lastRsRqBody)
		assert.NotNil(t, lastRs)
		assert.Equal(t, 200, lastRs.StatusCode)
		assert.Equal(t, `{"field1":"val1","field2":321}`, string(lastRsBody))
		assert.NotNil(t, lastRsTime)
		assert.NoError(t, lastRsErr)
	})

	t.Run("request is performed using a client that has rq/rs logging configured and it results in an error", func(t *testing.T) {
		gock.New("http://dummy").Post("/url").
			MatchParam("param", "test1").
			MatchHeader("X-Header1", "header-value-1").
			BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
			ReplyError(errors.New("mock error"))
		defer gock.Off()

		var rs TestRS
		rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
		err := client.RestPOST("http://dummy/url?param=test1", &Headers{
			{Key: "X-Header1", Value: "header-value-1"},
		}, rq, &rs)

		assert.Error(t, err)
		assert.Error(t, lastRsErr)
		assert.Equal(t, `error doing http call: Post "http://dummy/url?param=test1": mock error`, lastRsErr.Error())
	})
}
