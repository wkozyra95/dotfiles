package keyboard

import (
	"strings"
)

const (
	transparentLegend = "___"
	heldLegend        = "***"
	lineSeparator     = "|"
	// Between the legend of a key and what else the key does
	secondarySeparator = "||"
	// Drawn on the keys that fall through to the layer below
	transparentSymbol = "▽"
)

type key struct {
	// Lines of the legend: what the key does when it is pressed. Empty when
	// that is nothing.
	legend []string
	// Lines under the legend, in a smaller font: what else the key does
	secondary []string
	// Falls through to the layer below
	isTransparent bool
	// Held to get to the layer
	isHeld bool
}

// parseKeys reads the legends of a layer, the layer type describes them.
func parseKeys(layer layer) []key {
	var keys []key
	for _, legend := range strings.Fields(strings.Join(layer.keys, " ")) {
		switch {
		case legend == transparentLegend:
			keys = append(keys, key{legend: []string{transparentSymbol}, isTransparent: true})
		case legend == heldLegend:
			keys = append(keys, key{isHeld: true})
		case strings.Trim(legend, lineSeparator) == "":
			// the separator is the legend
			keys = append(keys, key{legend: []string{legend}})
		default:
			primary, secondary, _ := strings.Cut(legend, secondarySeparator)
			keys = append(keys, key{legend: legendLines(primary), secondary: legendLines(secondary)})
		}
	}
	return keys
}

// legendLines splits a part of a legend into its lines, none for an empty one.
func legendLines(part string) []string {
	if part == "" {
		return nil
	}
	return strings.Split(part, lineSeparator)
}
