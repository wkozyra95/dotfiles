package x11

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type regexpTestMatch struct {
	input      []byte
	matched    string
	matchedSub []string
	expression *regexp.Regexp
}

var xrandrOutputExample1 = `Screen 0: minimum 320 x 200, current 2560 x 1440, maximum 16384 x 16384
DisplayPort-0 connected primary 2560x1440+0+0 (normal left inverted right x axis y axis) 597mm x 336mm
   2560x1440     59.95*+
   1920x1200     59.95  
   2048x1080     60.00    24.00  
   1920x1080     60.00    50.00    59.94  
   1600x1200     60.00  
   1680x1050     59.95  
   1280x1024     75.02    60.02  
   1440x900      59.95  
   1280x800      59.95  
   1152x864      75.00  
   1280x720      60.00    50.00    59.94  
   1024x768      75.03    60.00  
   800x600       75.00    60.32  
   720x576       50.00  
   720x480       60.00    59.94  
   640x480       75.00    60.00    59.94  
   720x400       70.08  
DisplayPort-1 disconnected (normal left inverted right x axis y axis)
DisplayPort-2 disconnected (normal left inverted right x axis y axis)
HDMI-A-0 disconnected (normal left inverted right x axis y axis)
`

var xrandrOutputDisplayExample1 = `DisplayPort-0 connected primary 2560x1440+0+0 (normal left inverted right x axis y axis) 597mm x 336mm
   2560x1440     59.95*+
   1920x1200     59.95  
   2048x1080     60.00    24.00  
   1920x1080     60.00    50.00    59.94  
   1600x1200     60.00  
   1680x1050     59.95  
   1280x1024     75.02    60.02  
   1440x900      59.95  
   1280x800      59.95  
   1152x864      75.00  
   1280x720      60.00    50.00    59.94  
   1024x768      75.03    60.00  
   800x600       75.00    60.32  
   720x576       50.00  
   720x480       60.00    59.94  
   640x480       75.00    60.00    59.94  
   720x400       70.08  `

var (
	xrandrOutputDisplayExample2 = `HDMI-A-0 disconnected (normal left inverted right x axis y axis)`
	xrandrOutputDisplayExample3 = `DisplayPort-0 connected 2560x1440+0+0 (normal left inverted right x axis y axis) 597mm x 336mm
   2560x1440     59.95*+
   1920x1200     59.95  
   2048x1080     60.00    24.00  
   1920x1080     60.00    50.00    59.94  
   1600x1200     60.00  `
)

func TestXrandrDetectExample1(t *testing.T) {
	output, err := xrandrParseOutput([]byte(xrandrOutputExample1))
	require.Nil(t, err)
	assert.Equal(t, xrandrDetectResult{
		displays: []Display{
			{resX: 2560, resY: 1440, name: "DisplayPort-0", isActive: true, isPrimary: true},
			{resX: 0, resY: 0, name: "DisplayPort-1", isActive: false, isPrimary: false},
			{resX: 0, resY: 0, name: "DisplayPort-2", isActive: false, isPrimary: false},
			{resX: 0, resY: 0, name: "HDMI-A-0", isActive: false, isPrimary: false},
		},
	}, output)
}

func TestXrandrRegexpMatch(t *testing.T) {
	matchTests := []regexpTestMatch{
		{
			[]byte(xrandrOutputDisplayExample1),
			"DisplayPort-0 connected primary 2560x1440+0+0 (normal left inverted right x axis y axis) 597mm x 336mm\n",
			[]string{
				"DisplayPort-0",
				"connected",
				"primary",
				"2560x1440+0+0",
			},
			xrandrDisplayInfoRegexp,
		},
		{
			[]byte(xrandrOutputDisplayExample2),
			xrandrOutputDisplayExample2,
			[]string{
				"HDMI-A-0",
				"disconnected",
				"",
				"",
			},
			xrandrDisplayInfoRegexp,
		},
		{
			[]byte(xrandrOutputDisplayExample3),
			"DisplayPort-0 connected 2560x1440+0+0 (normal left inverted right x axis y axis) 597mm x 336mm\n",
			[]string{
				"DisplayPort-0",
				"connected",
				"",
				"2560x1440+0+0",
			},
			xrandrDisplayInfoRegexp,
		},
	}
	for i, testCase := range matchTests {
		t.Run(fmt.Sprintf("testCase %d", i), func(t *testing.T) {
			output := testCase.expression.FindSubmatch(testCase.input)
			t.Log(spew.Sdump(output))
			assert.Equal(t, testCase.matched, string(output[0]))
			for i, outputs := range output[1 : len(output)-1] {
				assert.Equal(t, testCase.matchedSub[i], string(outputs))
			}
			assert.Equal(t, len(testCase.matchedSub), len(output)-1)
		})
	}
}
