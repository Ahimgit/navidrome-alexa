package httpclient

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

var _ = Describe("CookieHelper", func() {
	var (
		cookieHelper   ICookieHelper
		tempCookieFile *os.File
		testBaseDomain = "example.com"
	)

	BeforeEach(func() {
		tempFile, err := os.CreateTemp("", "test_cookies.*.data")
		Expect(err).NotTo(HaveOccurred())
		Expect(tempFile.Close()).To(Succeed())
		tempCookieFile = tempFile
		cookieHelper = NewCookieHelper(tempCookieFile.Name())
	})

	AfterEach(func() {
		os.Remove(tempCookieFile.Name())
	})

	Describe("CookiesSaved", func() {
		Context("when the cookie file does not exist", func() {
			It("should return false", func() {
				os.Remove(tempCookieFile.Name())
				Expect(cookieHelper.CookiesSaved()).To(BeFalse())
			})
		})

		Context("when the cookie file exists", func() {
			It("should return true", func() {
				Expect(cookieHelper.CookiesSaved()).To(BeTrue())
			})
		})
	})

	Describe("Save and Load cookies", func() {
		Context("when loading cookies", func() {
			It("should save and load cookies from the file", func() {
				savedJar, err := cookiejar.New(nil)
				Expect(err).NotTo(HaveOccurred())
				savedJar.SetCookies(&url.URL{Scheme: "https", Host: testBaseDomain, Path: "/"}, []*http.Cookie{
					{Name: "test1", Value: "value1"},
					{Name: "test2", Value: "value2"},
				})
				Expect(cookieHelper.SaveCookies(savedJar, testBaseDomain)).To(Succeed())

				loadedJar, err := cookiejar.New(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(cookieHelper.LoadCookies(loadedJar, testBaseDomain)).To(Succeed())
				cookies := loadedJar.Cookies(&url.URL{Scheme: "https", Host: "alexa." + testBaseDomain, Path: "/"})
				Expect(cookies).To(HaveLen(2))
				Expect(cookies[0].Name).To(Equal("test1"))
				Expect(cookies[0].Value).To(Equal("value1"))
				Expect(cookies[1].Name).To(Equal("test2"))
				Expect(cookies[1].Value).To(Equal("value2"))
			})
		})
	})

})
