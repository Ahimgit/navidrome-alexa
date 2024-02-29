package httpclient

import (
	"errors"
	"github.com/h2non/gock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestHttpClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HttpClient Suite")
}

var _ = Describe("HttpClient", func() {

	type TestRQ struct {
		RqField1 string `json:"rqField1"`
		RqField2 int    `json:"rqField2"`
	}

	type TestRS struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	var client *HttpClient

	BeforeEach(func() {
		client = NewHttpClient()
	})

	AfterEach(func() {
		gock.Off()
	})

	Describe("SimpleGET", func() {
		Context("when simple GET request is made to a mocked 200 OK", func() {
			It("should be successful and return expected status and body", func() {
				gock.New("http://dummy").Get("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					Reply(200).
					BodyString("simple body1")

				rs, err := client.SimpleGET("http://dummy/url?param=test1", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				})

				Expect(err).To(BeNil())
				Expect(rs).To(Not(BeNil()))
				Expect(rs.Status).To(Equal(200))
				Expect(rs.Body).To(Equal("simple body1"))
				Expect(rs.Redirect).To(BeEmpty())
			})
		})

		Context("when simple GET request is made to a mocked 301 Moved Permanently", func() {
			It("should be successful and return expected status, body and redirect location", func() {
				gock.New("http://dummy").Get("/url").
					MatchParam("param", "test2").
					Reply(301).
					BodyString("simple body2").
					AddHeader("Location", "https://redirect?with=param")

				rs, err := client.SimpleGET("http://dummy/url?param=test2", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				})

				Expect(err).To(BeNil())
				Expect(rs).To(Not(BeNil()))
				Expect(rs.Status).To(Equal(301))
				Expect(rs.Body).To(Equal("simple body2"))
				Expect(rs.Redirect).To(Equal("https://redirect?with=param"))
			})
		})

		Context("when simple GET request is made to a mocked error", func() {
			It("should fail with the expected error", func() {
				gock.New("http://dummy").Get("/url").
					MatchParam("param", "test3").
					ReplyError(errors.New("mock error"))

				rs, err := client.SimpleGET("http://dummy/url?param=test3", nil)

				Expect(rs).To(BeNil())
				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal("error posting: Get \"http://dummy/url?param=test3\": mock error"))
			})
		})
	})

	Describe("SimplePOST", func() {
		Context("when simple POST request is made to a mocked 301 Moved Permanently", func() {
			It("should be successful and return expected status, body and redirect location", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					BodyString("key+1=value+1&key+2=value+2").
					Reply(301).
					BodyString("simple body1").
					AddHeader("Location", "https://redirect?with=param")

				headers := &Headers{{Key: "X-Header1", Value: "header-value-1"}}
				formData := &url.Values{}
				formData.Add("key 1", "value 1")
				formData.Add("key 2", "value 2")
				rs, err := client.SimplePOST("http://dummy/url?param=test1", headers, formData)

				Expect(err).To(BeNil())
				Expect(rs).To(Not(BeNil()))
				Expect(rs.Status).To(Equal(301))
				Expect(rs.Body).To(Equal("simple body1"))
				Expect(rs.Redirect).To(Equal("https://redirect?with=param"))
			})
		})

		Context("when simple POST request is made to a mocked error", func() {
			It("should fail with the expected error", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test2").
					ReplyError(errors.New("mock error"))

				headers := &Headers{{Key: "X-Header1", Value: "header-value-1"}}
				formData := &url.Values{}
				formData.Add("key 1", "value 1")
				formData.Add("key 2", "value 2")
				rs, err := client.SimplePOST("http://dummy/url?param=test2", headers, formData)

				Expect(rs).To(BeNil())
				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal("error posting: Post \"http://dummy/url?param=test2\": mock error"))
			})
		})
	})

	Describe("test RestGET", func() {
		Context("when REST GET request is made to a mocked 200 OK", func() {
			It("should respond with expected JSON parsed into an object", func() {
				gock.New("http://dummy").Get("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					Reply(200).
					JSON(`{"field1":"val1","field2":321}`)

				var rs TestRS
				err := client.RestGET("http://dummy/url?param=test1", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, &rs)

				Expect(err).To(BeNil())
				Expect(rs).To(Not(BeNil()))
				Expect(rs).To(Equal(TestRS{
					Field1: "val1",
					Field2: 321,
				}))
			})
		})

		Context("when REST GET request is made to a mocked invalid JSON response", func() {
			It("should fail with JSON parse error", func() {
				gock.New("http://dummy").Get("/url").
					MatchParam("param", "test2").
					MatchHeader("X-Header1", "header-value-1").
					Reply(200).
					JSON(`{"field1":"val1","field2":321`)

				var rs TestRS
				err := client.RestGET("http://dummy/url?param=test2", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, &rs)

				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal(`error unmarshalling rs: {"field1":"val1","field2":321: unexpected end of JSON input`))
			})
		})

		Context("when REST GET request is made to a mocked empty body 200 OK response", func() {
			It("should complete successfully without a response or error", func() {
				gock.New("http://dummy").Get("/url").
					MatchParam("param", "test3").
					MatchHeader("X-Header1", "header-value-1").
					Reply(200)

				err := client.RestGET("http://dummy/url?param=test3", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, nil)

				Expect(err).To(BeNil())
			})
		})

		Context("when REST GET request is made to a mocked 400 Bad Request", func() {
			It("should fail with an error that has status code and text", func() {
				gock.New("http://dummy").Get("/url").
					MatchParam("param", "test4").
					MatchHeader("X-Header1", "header-value-1").
					Reply(400)

				var rs TestRS
				err := client.RestGET("http://dummy/url?param=test4", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, &rs)

				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal(`error posting status: 400, 400 Bad Request`))
			})
		})

	})

	Describe("test RestPOST", func() {

		Context("when REST POST request is made to a mocked 200 OK", func() {
			It("should respond with expected JSON parsed into an object", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
					Reply(200).
					JSON(`{"field1":"val1","field2":321}`)

				rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
				var rs TestRS
				err := client.RestPOST("http://dummy/url?param=test1", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, rq, &rs)

				Expect(err).To(BeNil())
				Expect(rs).To(Not(BeNil()))
				Expect(rs).To(Equal(TestRS{
					Field1: "val1",
					Field2: 321,
				}))
			})
		})

		Context("when REST POST request is made to a mocked invalid JSON response", func() {
			It("should fail with JSON parse error", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test2").
					MatchHeader("X-Header1", "header-value-1").
					BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
					Reply(200).
					JSON(`{"field1":"val1","field2":321`)

				var rs TestRS
				rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
				err := client.RestPOST("http://dummy/url?param=test2", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, rq, &rs)

				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal(`error unmarshalling rs: {"field1":"val1","field2":321: unexpected end of JSON input`))
			})
		})

		Context("when REST POST request is made with invalid request", func() {
			It("should fail with JSON parse error", func() {
				badRequest := func() {}

				var rs TestRS
				err := client.RestPOST("http://dummy/url?param=test2", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, badRequest, &rs)

				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal(`error marshalling rq: json: unsupported type: func()`))
			})
		})

		Context("when REST POST request is made to a mocked empty body 200 OK response", func() {
			It("should complete successfully without a response or error", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
					Reply(200)
				rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
				err := client.RestPOST("http://dummy/url?param=test1", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, rq, nil)
				Expect(err).To(BeNil())
			})
		})

		Context("when REST POST request is made to a mocked 400 Bad Request", func() {
			It("should fail with an error that has status code and text", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
					Reply(400)

				var rs TestRS
				rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
				err := client.RestPOST("http://dummy/url?param=test1", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, rq, &rs)

				Expect(err).To(Not(BeNil()))
				Expect(err.Error()).To(Equal(`error posting status: 400, 400 Bad Request`))
			})
		})
	})

	Describe("test request logging", func() {
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

		BeforeEach(func() {
			client = client.
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
		})

		Context("when request is performed using a client that has rq/rs logging configured", func() {
			It("should pass rq/rs and thbeir bodies to logger functions", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
					Reply(200).
					JSON(`{"field1":"val1","field2":321}`)

				var rs TestRS
				rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
				err := client.RestPOST("http://dummy/url?param=test1", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, rq, &rs)

				Expect(err).To(BeNil())
				Expect(rs).To(Not(BeNil()))

				Expect(lastRq).To(Not(BeNil()))
				Expect(lastRq.URL.String()).To(Equal("http://dummy/url?param=test1"))
				Expect(lastRq.Header.Get("X-Header1")).To(Equal("header-value-1"))
				Expect(string(lastRqBody)).To(Equal(`{"rqField1":"rqVal1","rqField2":567}`))

				Expect(lastRsRq).To(Equal(lastRq))
				Expect(lastRsRqBody).To(Equal(lastRqBody))
				Expect(lastRs).To(Not(BeNil()))
				Expect(lastRs.StatusCode).To(Equal(200))
				Expect(string(lastRsBody)).To(Equal(`{"field1":"val1","field2":321}`))
				Expect(lastRsTime).To(Not(BeNil()))
				Expect(lastRsErr).To(BeNil())
			})
		})

		Context("when request is performed using a client that has rq/rs logging configured and it results in an error", func() {
			It("should pass error to the response logging function", func() {
				gock.New("http://dummy").Post("/url").
					MatchParam("param", "test1").
					MatchHeader("X-Header1", "header-value-1").
					BodyString(`{"rqField1":"rqVal1","rqField2":567}`).
					ReplyError(errors.New("mock error"))

				var rs TestRS
				rq := TestRQ{RqField1: "rqVal1", RqField2: 567}
				err := client.RestPOST("http://dummy/url?param=test1", &Headers{
					{Key: "X-Header1", Value: "header-value-1"},
				}, rq, &rs)

				Expect(err).To(Not(BeNil()))
				Expect(lastRsErr).To(Not(BeNil()))
				Expect(lastRsErr.Error()).To(Equal(`error posting: Post "http://dummy/url?param=test1": mock error`))
			})
		})
	})
})
