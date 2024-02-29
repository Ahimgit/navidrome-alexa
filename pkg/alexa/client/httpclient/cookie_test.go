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

func createTempFile(t *testing.T) *os.File {
	tempFile, err := os.CreateTemp("", "test_cookies.*.data")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())
	return tempFile
}
