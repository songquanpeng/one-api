package helper

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

// SaveTmpFile saves data to a temporary file. The filename would be apppended with a random string.
func SaveTmpFile(filename string, data io.Reader) (string, error) {
	f, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create temporary file %s", filename)
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	if err != nil {
		return "", errors.Wrapf(err, "failed to copy data to temporary file %s", filename)
	}

	return f.Name(), nil
}

// GetAudioDuration returns the duration of an audio file in seconds.
func GetAudioDuration(ctx context.Context, filename string) (float64, error) {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 {{input}}
	c := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := c.Output()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get audio duration")
	}

	return strconv.ParseFloat(string(bytes.TrimSpace(output)), 64)
}
