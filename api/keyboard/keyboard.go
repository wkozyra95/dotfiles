// Package keyboard draws the keymap of the keyboard used with the environment
// and shows it in an image viewer.
package keyboard

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/wkozyra95/dotfiles/api/context"
	"github.com/wkozyra95/dotfiles/api/sway"
	"github.com/wkozyra95/dotfiles/utils/exec"
)

const (
	// Floated and left without a border by a for_window rule in
	// configs/sway/config
	viewerAppID = "mycli-keyboard"
	// The image is drawn again on every call, it is a file only for the viewer
	cacheDir = ".cache/mycli/keyboard"
	// Part of the output the preview is allowed to cover
	maxOutputCoverage = 0.95
	// Limit of enlarging a keymap with a few layers to fill the output
	maxScale = 1.5
)

// keyboards are the keyboards used with the environments (CURRENT_ENV), there
// is nothing to show for an environment that is not listed. They are defined
// in dactyl.go.
var keyboards = map[string]keyboard{
	"home": dactylZmk,
	"work": dactylQmk,
}

// keyboard is all that is drawn of a keyboard, written down by hand: the
// firmware is not read.
type keyboard struct {
	name string
	// places the keys, in the order the layers list them
	layout []keyPosition
	// The sets of layers that are used together: a base layer followed by the
	// layers reached from it. Each one is drawn in a frame of its own.
	layerGroups []layerGroup
}

// keyPosition is where the center of a key is, measured in keys, and the angle
// the key is turned by (degrees, clockwise).
type keyPosition struct {
	x, y, rotation float64
}

type layerGroup struct {
	// drawn above the group, optional
	name   string
	layers []layer
}

type layer struct {
	name string
	// The legends of the keys separated by whitespace, in the order of the
	// layout. How they are split into strings does not matter, a string per row
	// of the keyboard reads best. A legend is one of:
	//   - `Q`: what the key does, drawn in the middle of it. The name of a layer
	//     is highlighted
	//   - `BT|0`: the same in more lines, one under another
	//   - `LALT||2xtap|F5`: after the `||` what else the key does, in small
	//     lines under the legend. For a key that does nothing else the room of
	//     the legend stays empty: `||2xtap|F5`
	//   - `___`: a key that falls through to the layer below
	//   - `***`: the key that is held to get to this layer
	keys []string
}

// Preview shows the keymap of the keyboard used with the environment. An
// already open preview is closed instead, so one key binding toggles it.
func Preview(ctx context.Context) error {
	keyboard, hasKeyboard := keyboards[ctx.Environment]
	if !hasKeyboard {
		return fmt.Errorf("no keyboard is configured for the %q environment", ctx.Environment)
	}
	if closePreview() {
		return nil
	}
	// stays unknown without sway
	var screen size
	if width, height, err := sway.FocusedOutputSize(); err == nil {
		screen = size{width, height}
	}
	svg, imageSize, err := drawKeymap(keyboard, screen)
	if err != nil {
		return err
	}
	outputDir := ctx.FromHome(cacheDir)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	image := path.Join(outputDir, keyboard.name+".svg")
	if err := os.WriteFile(image, svg, 0o644); err != nil {
		return err
	}
	return show(image, windowSize(imageSize, screen))
}

// size of an image, or of an output in its logical pixels. Zero when it is
// not known.
type size struct{ width, height int }

// fitScale is the scale the image is shown with on the screen: as large as the
// part of it the preview is allowed to cover, up to maxScale.
func fitScale(image, screen size) float64 {
	return min(
		maxScale,
		maxOutputCoverage*float64(screen.width)/float64(image.width),
		maxOutputCoverage*float64(screen.height)/float64(image.height),
	)
}

// windowSize fits the image to the screen, keeping its proportions.
func windowSize(image, screen size) size {
	if screen == (size{}) {
		return image
	}
	scale := fitScale(image, screen)
	return size{int(float64(image.width) * scale), int(float64(image.height) * scale)}
}

// show opens the image in swayimg, the viewer that can leave the window
// transparent around and behind the image.
func show(image string, window size) error {
	if !exec.CommandExists("swayimg") {
		opener := "xdg-open"
		if runtime.GOOS == "darwin" {
			opener = "open"
		}
		_, err := exec.Command().Args(opener, image).Start()
		return err
	}
	args := []string{"swayimg"}
	for _, option := range []string{
		"general.app_id=" + viewerAppID,
		fmt.Sprintf("general.size=%d,%d", window.width, window.height),
		"general.decoration=no",
		// nothing of the viewer shows: no background, no grid behind the
		// transparent parts of the image, no file details over it
		"viewer.window=#00000000",
		"viewer.transparency=#00000000",
		"info.show=no",
		"viewer.scale=fit",
		// only this image, not the rest of its directory
		"list.all=no",
	} {
		args = append(args, "--config="+option)
	}
	_, err := exec.Command().Args(append(args, image)...).Start()
	return err
}

// closePreview closes the preview and reports whether there was one. Only
// works with sway, anywhere else every call opens another preview.
func closePreview() bool {
	var stdout, stderr bytes.Buffer
	// fails when nothing matches
	err := exec.Command().WithBufout(&stdout, &stderr).
		Args("swaymsg", fmt.Sprintf(`[app_id="^%s$"] kill`, viewerAppID)).Run()
	return err == nil
}
