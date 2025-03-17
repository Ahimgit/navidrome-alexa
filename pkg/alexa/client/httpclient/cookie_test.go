package httpclient

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"testing"
)

func TestCookieHelper(t *testing.T) {

	t.Run("CookiesSaved", func(t *testing.T) {
		t.Run("the cookie file does not exist", func(t *testing.T) {
			cookieHelper := NewCookieHelper("does not exist")

			assert.False(t, cookieHelper.CookiesSaved())
		})

		t.Run("the cookie file exists", func(t *testing.T) {
			tempFile := createTempFile(t)
			defer os.Remove(tempFile.Name())
			cookieHelper := NewCookieHelper(tempFile.Name())

			assert.True(t, cookieHelper.CookiesSaved())
		})
	})

	t.Run("Save and Load cookies", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile.Name())
		cookieHelper := NewCookieHelper(tempFile.Name())

		testBaseDomain := "example.com"

		//save
		savedJar, err := cookiejar.New(nil)
		require.NoError(t, err)
		savedJar.SetCookies(&url.URL{Scheme: "https", Host: testBaseDomain, Path: "/"}, []*http.Cookie{
			{Name: "test1", Value: "value1"},
			{Name: "test2", Value: "value2"},
		})
		require.NoError(t, cookieHelper.SaveCookies(savedJar, testBaseDomain))

		//load
		loadedJar, err := cookiejar.New(nil)
		require.NoError(t, err)
		require.NoError(t, cookieHelper.LoadCookies(loadedJar, testBaseDomain))

		//assert
		cookies := loadedJar.Cookies(&url.URL{Scheme: "https", Host: "alexa." + testBaseDomain, Path: "/"})
		require.Len(t, cookies, 2)
		assert.Equal(t, "test1", cookies[0].Name)
		assert.Equal(t, "value1", cookies[0].Value)
		assert.Equal(t, "test2", cookies[1].Name)
		assert.Equal(t, "value2", cookies[1].Value)
	})
}

func TestExtractCSRF(t *testing.T) {

	t.Run("Extract existing csrf cookie", func(t *testing.T) {
		baseDomain := "example.com"
		expectedCSRF := "csrf_token_value"
		cookieDomain := &url.URL{Scheme: "https", Host: "alexa." + baseDomain, Path: "/"}
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(cookieDomain, []*http.Cookie{{Name: "csrf", Value: expectedCSRF}})

		actualCSRF := NewCookieHelper("unused").ExtractCSRF(jar, baseDomain)

		assert.Equal(t, expectedCSRF, actualCSRF)
	})

	t.Run("Extract empty csrf cookie", func(t *testing.T) {
		baseDomain := "example.com"
		emptyJar, _ := cookiejar.New(nil)

		actualCSRF := NewCookieHelper("unused").ExtractCSRF(emptyJar, baseDomain)

		assert.Empty(t, actualCSRF)
	})

}

func TestExtractLoginForm(t *testing.T) {
	pageHtml := `
		<form name="signInWithPasskeyButton" method="post" action="" class="a-spacing-small">
			<input type="hidden" name="formInput21" value="val21">
			<input type="text" name="username2">
			<input type="hidden" name="formInput22" value="val22">
			<input type="text" name="password2">
	    </form>
		<form method="post" 
		 name="signIn" novalidate action="https://www.amazon.com/ap/signin" 
         class="auth-validate-form auth-clearable-form auth-validate-form">
			<input type="hidden" name="formInput1" value="val1">
			<input type="text" name="username">
			<input type="hidden" name="formInput2" value="val2">
			<input type="text" name="password">
		</form><form></form>`

	formHtml := NewCookieHelper("unused").ExtractLoginForm(pageHtml)

	assert.NotNil(t, formHtml, "Extracted form data should not be nil")
	assert.Equal(t, `<form method="post" 
		 name="signIn" novalidate action="https://www.amazon.com/ap/signin" 
         class="auth-validate-form auth-clearable-form auth-validate-form">
			<input type="hidden" name="formInput1" value="val1">
			<input type="text" name="username">
			<input type="hidden" name="formInput2" value="val2">
			<input type="text" name="password">
		</form>`, formHtml)
}

func TestExtractLoginFormInputs(t *testing.T) {
	formHtml := `
		<form>
			<input type="hidden" name="formInput1" value="val1">
			<input type="text" name="username">
			<input type="hidden" name="formInput2" value="val2">
			<input type="text" name="password">
		</form>`

	formData := NewCookieHelper("unused").ExtractLoginFormInputs(formHtml)

	assert.NotNil(t, formData, "Extracted form data should not be nil")
	assert.Equal(t, "val1", formData.Get("formInput1"))
	assert.Equal(t, "val2", formData.Get("formInput2"))
}

func createTempFile(t *testing.T) *os.File {
	tempFile, err := os.CreateTemp("", "test_cookies.*.data")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())
	return tempFile
}
