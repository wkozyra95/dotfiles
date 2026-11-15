package keyboard

import (
	"fmt"
	"html"
	"math"
	"strings"
	"unicode/utf8"
)

// Sizes of the drawing, in the pixels of the image.
const (
	// Between the centers of two neighbouring keys
	keyPitch = 64
	// Side of the square drawn for a key
	keySize   = 60
	keyRadius = 7
	// Font of a legend and of the small lines under it, and the
	// room a line of each takes
	mainFontSize    = 14
	mainLineHeight  = 17
	smallFontSize   = 9
	smallLineHeight = 11
	// Room a legend has on a key, a longer one gets a smaller font
	legendWidth = 55
	// Width of a character of a monospace font, as a part of the font size
	characterWidth = 0.6

	// Room for the name of a layer above its keys
	layerHeader = 36
	// Between the keys of a layer and the sides of its cell
	layerPadding = 28

	// Around the frame of a group, so the gap between two frames is twice that
	groupMargin = 10
	// Room for the name of a group above its layers, and what an unnamed one
	// gets instead
	groupHeader    = 40
	groupPadding   = 8
	groupFooter    = 24
	groupRadius    = 14
	groupNameColor = "8f979f"
	// The name of a layer on the key that turns the layer on
	layerNameColor = "e5c07b"
	// The frame of a group, the rest of the image is transparent. The frame lets
	// a bit of what is behind the preview through.
	groupFrameColor   = "272a2e"
	groupFrameOpacity = 0.9
)

const style = `
svg { font-family: Source Code Pro, monospace; font-size: %dpx; fill: #d1d6db; }
text { text-anchor: middle; dominant-baseline: middle; }
rect.group { fill: #%s; fill-opacity: %g; }
rect.key { fill: #3f4750; stroke: #60666c; stroke-width: 1; }
rect.held { fill: #854747; }
text.group { font-size: 20px; font-weight: bold; text-anchor: start; fill: #%s; }
text.layer { font-weight: bold; text-anchor: start; }
text.small { font-size: %dpx; }
text.transparent { fill: #7e8184; }
text.activator { fill: #%s; }
`

type point struct{ x, y int }

// board is the layout of a keyboard in pixels: the centers of the keys,
// counted from the top left corner of the area they take.
type board struct {
	size
	keys []point
}

func newBoard(layout []keyPosition) board {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, key := range layout {
		// half of the side of the square around a turned key
		sin, cos := math.Sincos(key.rotation * math.Pi / 180)
		reach := (math.Abs(sin) + math.Abs(cos)) / 2
		minX, maxX = min(minX, key.x-reach), max(maxX, key.x+reach)
		minY, maxY = min(minY, key.y-reach), max(maxY, key.y+reach)
	}
	pixels := func(keys float64) int { return int(math.Round(keys * keyPitch)) }
	result := board{size: size{pixels(maxX - minX), pixels(maxY - minY)}}
	for _, key := range layout {
		result.keys = append(result.keys, point{pixels(key.x - minX), pixels(key.y - minY)})
	}
	return result
}

// frame is where a group is drawn: the rectangle around it and the top left
// corners of the cells of its layers.
type frame struct {
	point
	size
	layers []point
}

// arrangement places the groups next to each other or one under another, with
// the layers of a group in a number of columns.
type arrangement struct {
	image  size
	frames []frame
}

func arrange(groups []layerGroup, cell size, columns int, isStacked bool) arrangement {
	var result arrangement
	// frames that are one under another are equally wide
	widest := 0
	for _, group := range groups {
		widest = max(widest, min(columns, len(group.layers)))
	}
	// top left corner of the area of a group, the margin included
	var corner point
	for _, group := range groups {
		groupColumns := min(columns, len(group.layers))
		if isStacked {
			groupColumns = widest
		}
		rows := (len(group.layers) + columns - 1) / columns
		header := groupHeader
		if group.name == "" {
			header = groupPadding
		}
		frame := frame{
			point: point{corner.x + groupMargin, corner.y + groupMargin},
			size:  size{groupColumns * cell.width, header + rows*cell.height + groupFooter},
		}
		for i := range group.layers {
			frame.layers = append(frame.layers, point{
				frame.x + (i%columns)*cell.width,
				frame.y + header + (i/columns)*cell.height,
			})
		}
		result.frames = append(result.frames, frame)
		if isStacked {
			corner.y += frame.height + 2*groupMargin
			result.image = size{max(result.image.width, frame.width+2*groupMargin), corner.y}
		} else {
			corner.x += frame.width + 2*groupMargin
			result.image = size{corner.x, max(result.image.height, frame.height+2*groupMargin)}
		}
	}
	return result
}

// bestArrangement is the one that shows the keymap the largest on the screen.
// Without a screen the groups are next to each other, each in a single column.
func bestArrangement(groups []layerGroup, cell size, screen size) arrangement {
	mostLayers := 0
	for _, group := range groups {
		mostLayers = max(mostLayers, len(group.layers))
	}
	best := arrange(groups, cell, 1, false)
	if screen == (size{}) {
		return best
	}
	for _, isStacked := range []bool{false, true} {
		for columns := 1; columns <= mostLayers; columns++ {
			candidate := arrange(groups, cell, columns, isStacked)
			if fitScale(candidate.image, screen) > fitScale(best.image, screen) {
				best = candidate
			}
		}
	}
	return best
}

// drawKeymap draws the layers of the keyboard as an SVG image, arranged for
// the screen it is going to be shown on.
func drawKeymap(keyboard keyboard, screen size) ([]byte, size, error) {
	if len(keyboard.layout) == 0 || len(keyboard.layerGroups) == 0 {
		return nil, size{}, fmt.Errorf("keyboard %s has no keys or no layers", keyboard.name)
	}
	layerNames := map[string]bool{}
	keys := map[string][]key{}
	for _, group := range keyboard.layerGroups {
		if len(group.layers) == 0 {
			return nil, size{}, fmt.Errorf("group %q of %s has no layers", group.name, keyboard.name)
		}
		for _, layer := range group.layers {
			if layerNames[layer.name] {
				return nil, size{}, fmt.Errorf("%s has two layers named %q", keyboard.name, layer.name)
			}
			layerNames[layer.name] = true
			keys[layer.name] = parseKeys(layer)
			if len(keys[layer.name]) != len(keyboard.layout) {
				return nil, size{}, fmt.Errorf(
					"layer %q of %s lists %d keys, the keyboard has %d",
					layer.name, keyboard.name, len(keys[layer.name]), len(keyboard.layout),
				)
			}
		}
	}

	board := newBoard(keyboard.layout)
	cell := size{board.width + 2*layerPadding, layerHeader + board.height}
	arrangement := bestArrangement(keyboard.layerGroups, cell, screen)

	var svg strings.Builder
	fmt.Fprintf(
		&svg, `<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`+"\n",
		arrangement.image.width, arrangement.image.height, arrangement.image.width, arrangement.image.height,
	)
	svg.WriteString("<style>")
	fmt.Fprintf(
		&svg, style,
		mainFontSize, groupFrameColor, groupFrameOpacity, groupNameColor, smallFontSize, layerNameColor,
	)
	svg.WriteString("</style>\n")
	for i, group := range keyboard.layerGroups {
		frame := arrangement.frames[i]
		fmt.Fprintf(
			&svg, `<rect class="group" x="%d" y="%d" width="%d" height="%d" rx="%d"/>`+"\n",
			frame.x, frame.y, frame.width, frame.height, groupRadius,
		)
		if group.name != "" {
			fmt.Fprintf(
				&svg, `<text class="group" x="%d" y="%d">%s</text>`+"\n",
				frame.x+layerPadding, frame.y+groupHeader/2+6, html.EscapeString(group.name),
			)
		}
		for j, layer := range group.layers {
			fmt.Fprintf(
				&svg, `<g transform="translate(%d, %d)">`+"\n",
				frame.layers[j].x+layerPadding, frame.layers[j].y,
			)
			fmt.Fprintf(
				&svg, `<text class="layer" x="0" y="%d">%s:</text>`+"\n",
				layerHeader/2, html.EscapeString(layer.name),
			)
			for k, key := range keys[layer.name] {
				drawKey(&svg, key, board.keys[k], keyboard.layout[k].rotation, layerNames)
			}
			svg.WriteString("</g>\n")
		}
	}
	svg.WriteString("</svg>\n")
	return []byte(svg.String()), arrangement.image, nil
}

func drawKey(svg *strings.Builder, key key, center point, rotation float64, layerNames map[string]bool) {
	transform := fmt.Sprintf("translate(%d, %d)", center.x, layerHeader+center.y)
	if rotation != 0 {
		transform += fmt.Sprintf(" rotate(%g)", rotation)
	}
	class := "key"
	if key.isHeld {
		class += " held"
	}
	fmt.Fprintf(
		svg, `<g transform="%s"><rect class="%s" x="%d" y="%d" width="%d" height="%d" rx="%d"/>`,
		transform, class, -keySize/2, -keySize/2, keySize, keySize, keyRadius,
	)
	// the lines are centered on the key together, the room of the legend stays
	// empty on a key that has only the small ones
	top := -(float64(max(len(key.legend), 1))*mainLineHeight + float64(len(key.secondary))*smallLineHeight) / 2
	isOnlyLine := len(key.legend) == 1 && len(key.secondary) == 0
	// The name of a layer is highlighted: a line that is one, or a legend that
	// is one as a whole.
	isLegendLayerName := layerNames[strings.Join(key.legend, " ")]
	drawLines := func(lines []string, class string, fontSize, height float64, isLayerName bool) {
		if key.isTransparent {
			class += " transparent"
		}
		fontStyle := fontSizeFor(lines, fontSize)
		for _, line := range lines {
			lineClass := class
			if isLayerName || layerNames[line] {
				lineClass += " activator"
			}
			position := ""
			if !isOnlyLine {
				position = fmt.Sprintf(` y="%g"`, top+height/2)
			}
			fmt.Fprintf(svg, `<text class="%s"%s%s>%s</text>`, lineClass, position, fontStyle, html.EscapeString(line))
			top += height
		}
	}
	drawLines(key.legend, "main", mainFontSize, mainLineHeight, isLegendLayerName)
	if len(key.legend) == 0 {
		top += mainLineHeight
	}
	drawLines(key.secondary, "small", smallFontSize, smallLineHeight, false)
	svg.WriteString("</g>\n")
}

// fontSizeFor is the style that makes the lines of a legend fit on a key,
// empty when the font of their kind does. The lines stay equally large.
func fontSizeFor(lines []string, fontSize float64) string {
	longest := 0
	for _, line := range lines {
		longest = max(longest, utf8.RuneCountInString(line))
	}
	if legendWidth >= characterWidth*float64(longest)*fontSize {
		return ""
	}
	return fmt.Sprintf(` style="font-size: %.1fpx"`, legendWidth/(characterWidth*float64(longest)))
}
