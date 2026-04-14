// Package video provides steganography support for video files by processing
// individual frames using ffmpeg.
package video

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"stego/internal/stego_loader"

	"golang.org/x/sync/errgroup"
)

// CmdRunner defines an interface for executing shell commands, allowing for mocking in tests.
type CmdRunner interface {
	Run(name string, arg ...string) error
}

// DefaultCmdRunner is the default implementation of CmdRunner using os/exec.
type DefaultCmdRunner struct{}

func (r DefaultCmdRunner) Run(name string, arg ...string) error {
	return exec.Command(name, arg...).Run()
}

var runner CmdRunner = DefaultCmdRunner{}

// EncodeVideo encodes a message into a video file by embedding it in every frame.
// It extracts frames to a temporary directory, encodes the message into each
// frame, and then reconstructs the video using ffmpeg.
func EncodeVideo(pathIn string, pathOut string, message string, pass string, messageLen int, addMod int, negMod int) error {
	tmpDir, err := os.MkdirTemp("", "stego-video-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract frames from input video
	if err := runner.Run("ffmpeg", "-i", pathIn, filepath.Join(tmpDir, "frame_%04d.png")); err != nil {
		return fmt.Errorf("ffmpeg extract frames failed: %w", err)
	}

	frames, err := filepath.Glob(filepath.Join(tmpDir, "frame_*.png"))
	if err != nil {
		return fmt.Errorf("failed to list extracted frames: %w", err)
	}

	if len(frames) == 0 {
		return fmt.Errorf("no frames extracted from video")
	}

	// Encode message into ALL frames concurrently
	g, _ := errgroup.WithContext(context.Background())
	// Limit concurrency to avoid resource exhaustion
	g.SetLimit(10)

	for _, framePath := range frames {
		framePath := framePath // capture for closure
		g.Go(func() error {
			if err := stego_loader.EncodeFile(framePath, framePath, message, pass, messageLen, addMod, negMod); err != nil {
				return fmt.Errorf("failed to encode message into frame %s: %w", framePath, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Reconstruct video from frames
	if err := runner.Run("ffmpeg", "-y", "-i", filepath.Join(tmpDir, "frame_%04d.png"), "-i", pathIn, "-map", "0:v", "-map", "1:a?", "-c:v", "libx264", "-pix_fmt", "yuv420p", pathOut); err != nil {
		return fmt.Errorf("ffmpeg reconstruct video failed: %w", err)
	}

	return nil
}

// DecodeVideo decodes a message from a video file.
// It extracts the first frame of the video and decodes the message from it.
func DecodeVideo(path string, pass string, messageLen int) (string, error) {
	tmpDir, err := os.MkdirTemp("", "stego-video-decode-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract only the first frame
	if err := runner.Run("ffmpeg", "-i", path, "-vframes", "1", filepath.Join(tmpDir, "frame_0001.png")); err != nil {
		return "", fmt.Errorf("ffmpeg extract first frame failed: %w", err)
	}

	msg, err := stego_loader.DecodeFile(filepath.Join(tmpDir, "frame_0001.png"), pass, messageLen)
	if err != nil {
		return "", fmt.Errorf("failed to decode message from frame: %w", err)
	}
	return msg, nil
}
