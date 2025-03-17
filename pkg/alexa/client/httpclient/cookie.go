package httpclient

import (
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type ICookieHelper interface {
	CookiesSaved() (cookiesExist bool)
	SaveCookies(jar http.CookieJar, baseDomain string) (err error)
	LoadCookies(jar http.CookieJar, baseDomain string) (err error)
	ExtractCSRF(jar http.CookieJar, baseDomain string) (csrf string)
	ExtractLoginForm(pageHtml string) (formHtml string)
	ExtractLoginFormInputs(formHtml string) (formData *url.Values)
}

type CookieHelper struct {
	filePath string
}

var formExtractor = regexp.MustCompile(`(?s)<form[^>]+name="signIn"[^>]*>(.*?)</form>`)
var formInputExtractor = regexp.MustCompile(`name="([^"]+)".*?value="([^"]+)"`)

func NewCookieHelper(filePath string) ICookieHelper {
	return &CookieHelper{
		filePath: filePath,
	}
}

func (c *CookieHelper) CookiesSaved() (cookiesExist bool) {
	info, err := os.Stat(c.filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (c *CookieHelper) SaveCookies(jar http.CookieJar, baseDomain string) (err error) {
	cookieFile, err := os.Create(c.filePath)
	if err != nil {
		return errors.Wrap(err, "unable to create cookie file")
	}
	defer cookieFile.Close()
	cookies := jar.Cookies(&url.URL{Scheme: "https", Host: baseDomain, Path: "/"})
	for _, cookie := range cookies {
		if _, err = cookieFile.WriteString(cookie.String() + "\n"); err != nil {
			return errors.Wrap(err, "unable to write cookie file")
		}
	}
	return nil
}

func (c *CookieHelper) LoadCookies(jar http.CookieJar, baseDomain string) (err error) {
	lines, err := os.ReadFile(c.filePath)
	if err != nil {
		return errors.Wrap(err, "unable to read cookie file")
	}
	var cookies []*http.Cookie
	for _, line := range strings.Split(string(lines), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				cookies = append(cookies, &http.Cookie{Name: parts[0], Value: parts[1]})
			}
		}
	}
	jar.SetCookies(&url.URL{Scheme: "https", Host: "alexa." + baseDomain, Path: "/"}, cookies)
	return nil
}

func (c *CookieHelper) ExtractCSRF(jar http.CookieJar, baseDomain string) (csrf string) {
	cookies := jar.Cookies(&url.URL{Scheme: "https", Host: "alexa." + baseDomain, Path: "/"})
	for _, cookie := range cookies {
		if strings.EqualFold(cookie.Name, "csrf") {
			return cookie.Value
		}
	}
	return ""
}

func (c *CookieHelper) ExtractLoginForm(pageHtml string) (formHtml string) {
	match := formExtractor.FindStringSubmatch(pageHtml)
	if len(match) > 1 {
		return match[0]
	} else {
		return ""
	}
}

func (c *CookieHelper) ExtractLoginFormInputs(formHtml string) (formData *url.Values) {
	formData = &url.Values{}
	matches := formInputExtractor.FindAllStringSubmatch(formHtml, -1)
	for _, match := range matches {
		if len(match) == 3 {
			formData.Add(match[1], match[2])
		}
	}
	return formData
}
