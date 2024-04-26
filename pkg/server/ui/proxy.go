package ui

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
)

const MaxSize = 3 * 1024 * 1024

func GetWidget(c *gin.Context) {
	proxied := c.Query("proxied")
	if err := validateURL(proxied); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	proxiedContent, err := fetchWithLimit(proxied, MaxSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	widgetContent, err := assets.ReadFile("assets/widget.js")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	combinedContent := string(proxiedContent) + "\n\n\n // NA widget\n" + string(widgetContent)
	c.Header("Content-Type", "application/javascript")
	c.String(http.StatusOK, combinedContent)
}

func validateURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return err
	}

	if !(parsedURL.Scheme == "http" || parsedURL.Scheme == "https") {
		return errors.New("invalid URL scheme")
	}
	return nil
}

func fetchWithLimit(rawURL string, maxSize int64) ([]byte, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	limitedReader := &io.LimitedReader{R: resp.Body, N: maxSize}
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if limitedReader.N <= 0 {
		return nil, errors.New("remote file is too large")
	}
	return content, nil
}
