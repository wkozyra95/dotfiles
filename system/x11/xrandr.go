package x11

import (
	"bytes"
	"regexp"
	"strconv"

	"github.com/wkozyra95/dotfiles/utils/exec"
)

type Display struct {
	resX      int
	resY      int
	name      string
	isPrimary bool
	isActive  bool
}

// match block of output for specific outputs with the exception of first block
// e.g.
// DisplayPort-0 connected primary 2560x1440+0+0 (normal left inverted right x axis y axis) 597mm x 336mm
//
//	2560x1440     59.95*+
//	1920x1200     59.95
//	2048x1080     60.00    24.00
//	1920x1080     60.00    50.00    59.94
var xrandrDisplayBlockRegexp = regexp.MustCompile("[a-zA-Z].*(?:\n .*)*")

// match first line of output info and extract
// name, isActive, isPrimary, resolution+offset
// e.g.
// DisplayPort-0 connected primary 2560x1440+0+0 (normal left inverted right x axis y axis) 597mm x 336mm\n
var xrandrDisplayInfoRegexp = regexp.MustCompile(
	"^([-a-zA-Z0-9]*) (connected|disconnected)\\ ?(primary|)\\ ?([0-9]+x[0-9]+\\+[0-9]+\\+[0-9]+|)\\ ?.*\n?",
)

var xrandrDisplayResolutionOffsetRegexp = regexp.MustCompile(
	"([0-9]+)x([0-9]+)\\+([0-9]+)\\+([0-9]+)",
)

type xrandrDetectResult struct {
	displays []Display
}

func xrandrDetect() (xrandrDetectResult, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	if err := exec.Command().WithBufout(&stdout, &stderr).Run("xrandr"); err != nil {
		return xrandrDetectResult{}, err
	}
	return xrandrParseOutput(stdout.Bytes())
}

func xrandrParseOutput(output []byte) (xrandrDetectResult, error) {
	var result xrandrDetectResult
	// match (name, connectedStatus) for block from regex above
	blocks := xrandrDisplayBlockRegexp.FindAllSubmatch(output, -1)
	result.displays = make([]Display, len(blocks)-1)
	for i, display := range blocks[1:] {
		displayInfo := xrandrDisplayInfoRegexp.FindSubmatch(display[0])
		result.displays[i].name = string(displayInfo[1])
		result.displays[i].isActive = string(displayInfo[2]) == "connected"
		result.displays[i].isPrimary = string(displayInfo[3]) == "primary"
		if string(displayInfo[4]) != "" {
			resolutionRaw := xrandrDisplayResolutionOffsetRegexp.FindSubmatch(displayInfo[4])
			if res, err := strconv.Atoi(string(resolutionRaw[1])); err != nil {
				return result, err
			} else {
				result.displays[i].resX = res
			}
			if res, err := strconv.Atoi(string(resolutionRaw[2])); err != nil {
				return result, err
			} else {
				result.displays[i].resY = res
			}

		}
		log.Printf("display|||%v", result.displays[i])
	}
	return result, nil
}

type xrandrCmd struct{}

func XrandrCommand() xrandrCmd {
	return xrandrCmd{}
}
