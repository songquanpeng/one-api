package helper

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAudioDuration(t *testing.T) {
	t.Run("should return correct duration for a valid audio file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_audio*.mp3")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// download test audio file
		resp, err := http.Get("https://s3.laisky.com/uploads/2025/01/audio-sample.m4a")
		require.NoError(t, err)
		defer resp.Body.Close()

		_, err = io.Copy(tmpFile, resp.Body)
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())

		duration, err := GetAudioDuration(context.Background(), tmpFile.Name())
		require.NoError(t, err)
		require.Equal(t, duration, 3.904)
	})

	t.Run("should return an error for a non-existent file", func(t *testing.T) {
		_, err := GetAudioDuration(context.Background(), "non_existent_file.mp3")
		require.Error(t, err)
	})
}

func TestGetAudioTokens(t *testing.T) {
	t.Run("should return correct tokens for a valid audio file", func(t *testing.T) {
		// download test audio file
		resp, err := http.Get("https://s3.laisky.com/uploads/2025/01/audio-sample.m4a")
		require.NoError(t, err)
		defer resp.Body.Close()

		tokens, err := GetAudioTokens(context.Background(), resp.Body, 50)
		require.NoError(t, err)
		require.Equal(t, tokens, 200)
	})

	t.Run("should return an error for a non-existent file", func(t *testing.T) {
		_, err := GetAudioTokens(context.Background(), nil, 1)
		require.Error(t, err)
	})
}
