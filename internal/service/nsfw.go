package service

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type NsfwService struct {
	Timeout time.Duration
}

func NewNsfwService(timeout time.Duration) *NsfwService {
	return &NsfwService{Timeout: timeout}
}

func (n *NsfwService) RunScan(targetDir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), n.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "video_analysis.pex", targetDir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pex execution failed: %w\nOutput:\n%s", err, output)
	}

	return nil
}
