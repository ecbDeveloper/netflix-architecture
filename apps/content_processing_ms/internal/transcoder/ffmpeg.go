package transcoder

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Transcoder struct{}

func NewTranscoder() *Transcoder {
	return &Transcoder{}
}

func (t *Transcoder) ProcessVideoToDASH(inputPath string, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cmd := exec.Command(
		"ffmpeg",
		"-y", "-i", inputPath,
		"-preset", "veryfast", "-g", "48", "-sc_threshold", "0",

		"-map", "0:v:0", "-s:v:0", "1920x1080", "-c:v:0", "libx264", "-b:v:0", "5000k",

		"-map", "0:v:0", "-s:v:1", "854x480", "-c:v:1", "libx264", "-b:v:1", "1000k",

		"-map", "0:v:0", "-s:v:2", "256x144", "-c:v:2", "libx264", "-b:v:2", "100k",

		"-map", "0:a:0", "-c:a:0", "aac", "-b:a:0", "128k", "-ac", "2",

		"-f", "dash",
		"-seg_duration", "4",
		"-use_template", "1",
		"-use_timeline", "1",
		"-window_size", "0",
		"-init_seg_name", "init_$RepresentationID$.m4s",
		"-media_seg_name", "chunk_$RepresentationID$_$Number$.m4s",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		filepath.Join(outputDir, "master.mpd"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (t *Transcoder) GetVideoDuration(inputPath string) (int, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(output))
	}

	trimmedOutput := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(trimmedOutput, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration '%s': %w", trimmedOutput, err)
	}

	return int(math.Round(durationFloat)), nil
}
