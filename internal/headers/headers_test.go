package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Content-Type:    application/json   \r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, "application/json", headers["content-type"])
		assert.Equal(t, 38, n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["existing-header"] = "existing-value"

		// Parse first header
		data1 := []byte("Host: localhost:8080\r\n")
		n1, done1, err1 := headers.Parse(data1)
		require.NoError(t, err1)
		assert.Equal(t, "localhost:8080", headers["host"])
		assert.Equal(t, 22, n1)
		assert.False(t, done1)

		// Parse second header
		data2 := []byte("Content-Type: text/html\r\n")
		n2, done2, err2 := headers.Parse(data2)
		require.NoError(t, err2)
		assert.Equal(t, "text/html", headers["content-type"])
		assert.Equal(t, 25, n2)
		assert.False(t, done2)

		// Verify existing header still exists
		assert.Equal(t, "existing-value", headers["existing-header"])
		assert.Len(t, headers, 3)
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\nSome body content here")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 2, n)
		assert.True(t, done)
		assert.Empty(t, headers)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host : localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid header")
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Incomplete data without CRLF", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
		assert.Empty(t, headers)
	})

	t.Run("Invalid field-name value", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("H©st: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.Equal(t, false, done)
	})

	t.Run("Duplicate field-name, combine values", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Set-Person: Zack\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 18, n)
		assert.False(t, done)
		assert.Equal(t, "Zack", headers["set-person"])

		data = []byte("Set-Person: Jimmy\r\n")
		n, done, err = headers.Parse(data)
		require.NoError(t, err)
		assert.Equal(t, 19, n)
		assert.False(t, done)
		assert.Equal(t, "Zack, Jimmy", headers["set-person"])
	})
}
