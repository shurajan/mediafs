package media

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	resolutionCache sync.Map
	durationCache   sync.Map
)

// GetVideoResolution возвращает строку вида "1920x1080"
func GetVideoResolution(filePath string) *string {
	if val, ok := resolutionCache.Load(filePath); ok {
		if cached, ok := val.(*string); ok {
			return cached
		}
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	res := strings.TrimSpace(string(output))
	if matched, _ := regexp.MatchString(`^\d+x\d+$`, res); matched {
		resolutionCache.Store(filePath, &res)
		return &res
	}

	return nil
}

// GetVideoDuration возвращает продолжительность видео в секундах
func GetVideoDuration(filePath string) *int {
	if val, ok := durationCache.Load(filePath); ok {
		if cached, ok := val.(*int); ok {
			return cached
		}
	}

	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	durationStr := strings.TrimSpace(string(output))
	floatVal, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return nil
	}

	rounded := int(floatVal + 0.5)
	durationCache.Store(filePath, &rounded)
	return &rounded
}
