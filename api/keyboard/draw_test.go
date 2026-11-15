package keyboard

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKeys(t *testing.T) {
	keys := parseKeys(
		layer{keys: []string{"Q  BT|0 ___", "  *** | \\ a|b|c", "L_1||2xtap|x _ __ ||2xtap|F5 ||"}},
	)
	assert.Equal(t, []key{
		{legend: []string{"Q"}},
		{legend: []string{"BT", "0"}},
		{legend: []string{"▽"}, isTransparent: true},
		{isHeld: true},
		{legend: []string{"|"}},
		{legend: []string{`\`}},
		{legend: []string{"a", "b", "c"}},
		{legend: []string{"L_1"}, secondary: []string{"2xtap", "x"}},
		{legend: []string{"_"}},
		{legend: []string{"__"}},
		{secondary: []string{"2xtap", "F5"}},
		{legend: []string{"||"}},
	}, keys)
}

// The lines on a key are centered on it as a block.
func TestDrawKeyLegend(t *testing.T) {
	layerNames := map[string]bool{"fn": true, "alt": true, "alt numbers": true}
	draw := func(key key) string {
		var svg strings.Builder
		drawKey(&svg, key, point{0, 0}, 0, layerNames)
		_, texts, _ := strings.Cut(strings.TrimSuffix(svg.String(), "</g>\n"), "/>")
		return texts
	}
	assert.Equal(t, `<text class="main">Q</text>`, draw(key{legend: []string{"Q"}}))
	assert.Equal(t, `<text class="main activator">fn</text>`, draw(key{legend: []string{"fn"}}))
	// the lines of a legend are equally large
	assert.Equal(
		t,
		`<text class="main activator" y="-8.5">fn</text><text class="main" y="8.5">+Ctl</text>`,
		draw(key{legend: []string{"fn", "+Ctl"}}),
	)
	// also when one of them is too long for the key
	assert.Equal(
		t,
		`<text class="main activator" y="-8.5" style="font-size: 11.5px">fn</text>`+
			`<text class="main" y="8.5" style="font-size: 11.5px">+Ctl+Alt</text>`,
		draw(key{legend: []string{"fn", "+Ctl+Alt"}}),
	)
	// the whole legend is the name of a layer
	assert.Equal(
		t,
		`<text class="main activator" y="-8.5" style="font-size: 13.1px">alt</text>`+
			`<text class="main activator" y="8.5" style="font-size: 13.1px">numbers</text>`,
		draw(key{legend: []string{"alt", "numbers"}}),
	)
	// what else the key does is under the legend, in small lines
	assert.Equal(
		t,
		`<text class="main" y="-16.5">LALT</text><text class="small" y="-2.5">2xtap</text>`+
			`<text class="small" y="8.5">SWITCH</text><text class="small activator" y="19.5">alt</text>`,
		draw(key{legend: []string{"LALT"}, secondary: []string{"2xtap", "SWITCH", "alt"}}),
	)
	// the room of the legend stays empty on a key without one
	assert.Equal(
		t,
		`<text class="small" y="-2.5">2xtap</text><text class="small" y="8.5">Ctrl+Alt</text>`+
			`<text class="small" y="19.5">+Del</text>`,
		draw(key{secondary: []string{"2xtap", "Ctrl+Alt", "+Del"}}),
	)
	assert.Equal(t, `<text class="small" y="8.5">2xtap</text>`, draw(key{secondary: []string{"2xtap"}}))
	assert.Equal(t, ``, draw(key{isHeld: true}))
}

// layers of a keyboard with nothing on them
func emptyLayers(names ...string) []layer {
	layers := make([]layer, len(names))
	for i, name := range names {
		layers[i] = layer{name: name}
	}
	return layers
}

var (
	// the cell of a layer of the dactyl
	testCell  = size{984, 423}
	twoGroups = []layerGroup{
		{name: "main", layers: emptyLayers("default", "raise", "lower")},
		{name: "alt", layers: emptyLayers("alt_default", "alt_raise")},
	}
	singleGroup = []layerGroup{{layers: emptyLayers("default", "raise", "lower")}}
	landscape   = size{2560, 1440}
	portrait    = size{1440, 2560}
)

func TestGroupsNextToEachOther(t *testing.T) {
	// without a screen, and on one that is wider than tall
	for _, screen := range []size{{}, landscape} {
		assert.Equal(t, arrangement{
			image: size{2008, 1353},
			frames: []frame{
				{point{10, 10}, size{984, 1333}, []point{{10, 50}, {10, 473}, {10, 896}}},
				{point{1014, 10}, size{984, 910}, []point{{1014, 50}, {1014, 473}}},
			},
		}, bestArrangement(twoGroups, testCell, screen))
	}
}

func TestGroupsOneUnderAnotherOnPortraitScreen(t *testing.T) {
	assert.Equal(t, arrangement{
		image: size{1004, 2283},
		frames: []frame{
			{point{10, 10}, size{984, 1333}, []point{{10, 50}, {10, 473}, {10, 896}}},
			{point{10, 1363}, size{984, 910}, []point{{10, 1403}, {10, 1826}}},
		},
	}, bestArrangement(twoGroups, testCell, portrait))
}

func TestSingleGroupFillsTheScreen(t *testing.T) {
	// two columns on a wide screen, the group has no name to make room for
	assert.Equal(t, arrangement{
		image:  size{1988, 898},
		frames: []frame{{point{10, 10}, size{1968, 878}, []point{{10, 18}, {994, 18}, {10, 441}}}},
	}, bestArrangement(singleGroup, testCell, landscape))
	// one column on a tall one, and when the screen is not known
	for _, screen := range []size{{}, portrait} {
		assert.Equal(t, arrangement{
			image:  size{1004, 1321},
			frames: []frame{{point{10, 10}, size{984, 1301}, []point{{10, 18}, {10, 441}, {10, 864}}}},
		}, bestArrangement(singleGroup, testCell, screen))
	}
}

func TestStackedFramesAreEquallyWide(t *testing.T) {
	groups := []layerGroup{{layers: emptyLayers("a", "b")}, {layers: emptyLayers("c")}}
	result := arrange(groups, size{100, 50}, 2, true)
	assert.Equal(t, size{220, 204}, result.image)
	assert.Equal(t, size{200, 82}, result.frames[0].size)
	assert.Equal(t, size{200, 82}, result.frames[1].size)
}

var testKeyboard = keyboard{
	name: "test",
	// the keys take 2.25 by 1.5: the second one is turned, the third is half a key up
	layout: []keyPosition{{x: 0.5, y: 0.75}, {x: 1.75, y: 0.75, rotation: 90}, {x: 0.5, y: 0.25}},
	layerGroups: []layerGroup{
		{name: "a <&> b", layers: []layer{
			{name: "base", keys: []string{"Q upper||2×tap", "<&>"}},
			{name: "upper", keys: []string{"___ *** longlegend"}},
		}},
	},
}

func TestDrawKeymap(t *testing.T) {
	image, imageSize, err := drawKeymap(testKeyboard, size{})
	assert.Nil(t, err)
	svg := string(image)
	// the keys take 144 by 96 pixels
	assert.Equal(t, size{144 + 2*28 + 2*10, 2*10 + 40 + 2*(36+96) + 24}, imageSize)
	assert.True(t, strings.HasPrefix(
		svg, `<svg width="220" height="348" viewBox="0 0 220 348" xmlns="http://www.w3.org/2000/svg">`,
	))
	assert.True(t, strings.HasSuffix(svg, "</g>\n</svg>\n"))
	assert.Contains(t, svg, `<rect class="group" x="10" y="10" width="200" height="328" rx="14"/>`)
	assert.Contains(t, svg, `<text class="group" x="38" y="36">a &lt;&amp;&gt; b</text>`)
	// the layers, one under another
	assert.Contains(t, svg, `<g transform="translate(38, 50)">`+"\n"+`<text class="layer" x="0" y="18">base:</text>`)
	assert.Contains(t, svg, `<g transform="translate(38, 182)">`+"\n"+`<text class="layer" x="0" y="18">upper:</text>`)
	// the keys of the first one
	assert.Contains(
		t, svg,
		`<g transform="translate(32, 100)"><rect class="key" x="-30" y="-30" width="60" height="60" rx="7"/>`+
			`<text class="main">Q</text></g>`,
	)
	assert.Contains(
		t,
		svg,
		`<g transform="translate(112, 100) rotate(90)"><rect class="key" x="-30" y="-30" width="60" height="60" rx="7"/>`+
			`<text class="main activator" y="-5.5">upper</text><text class="small" y="8.5">2×tap</text></g>`,
	)
	assert.Contains(t, svg, `<g transform="translate(32, 68)"><rect class="key" `)
	assert.Contains(t, svg, `<text class="main">&lt;&amp;&gt;</text>`)
	// and of the second
	assert.Contains(t, svg, `<text class="main transparent">▽</text>`)
	assert.Contains(t, svg, `<rect class="key held" x="-30" y="-30" width="60" height="60" rx="7"/></g>`)
	assert.Contains(t, svg, `<text class="main" style="font-size: 9.2px">longlegend</text>`)
	assert.Equal(t, 6, strings.Count(svg, `<rect class="key`))
}

func TestDrawKeymapInvalid(t *testing.T) {
	keyboard := testKeyboard
	keyboard.layerGroups = []layerGroup{{layers: []layer{{name: "base", keys: []string{"Q W"}}}}}
	_, _, err := drawKeymap(keyboard, size{})
	assert.EqualError(t, err, `layer "base" of test lists 2 keys, the keyboard has 3`)

	keyboard.layerGroups = []layerGroup{
		{layers: []layer{{name: "base", keys: []string{"Q W E"}}}},
		{layers: []layer{{name: "base", keys: []string{"Q W E"}}}},
	}
	_, _, err = drawKeymap(keyboard, size{})
	assert.EqualError(t, err, `test has two layers named "base"`)

	keyboard.layerGroups = []layerGroup{{name: "main"}}
	_, _, err = drawKeymap(keyboard, size{})
	assert.EqualError(t, err, `group "main" of test has no layers`)

	keyboard.layerGroups = nil
	_, _, err = drawKeymap(keyboard, size{})
	assert.EqualError(t, err, "keyboard test has no keys or no layers")
}

// The keyboards of the environments: every layer lists all the keys.
func TestConfiguredKeyboards(t *testing.T) {
	assert.Equal(t, "dactyl-zmk", keyboards["home"].name)
	assert.Equal(t, "dactyl-qmk", keyboards["work"].name)
	for _, keyboard := range keyboards {
		assert.Len(t, keyboard.layout, 46, keyboard.name)
		// 14.5 keys wide, the thumb blocks reach a bit over 6 keys down
		assert.Equal(t, size{928, 387}, newBoard(keyboard.layout).size, keyboard.name)
		_, _, err := drawKeymap(keyboard, landscape)
		assert.Nil(t, err, keyboard.name)
	}
}
