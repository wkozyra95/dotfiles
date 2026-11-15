package keyboard

import (
	"math"
)

// The keyboards below are copied by hand from the firmware of each build,
// nothing reads the firmware when the preview is drawn. After flashing a
// changed keymap, update the layers here. The sources:
//
//   - dactylZmk:
//     ~/drive/MyProjects/dactyl/zmk_firmware/zmk-config/config/dactyl.keymap
//   - dactylQmk:
//     ~/drive/MyProjects/dactyl/qmk_firmware/keyboards/handwired/dactyl_manuform/4x5_rev2/keymaps/custom/keymap.c
//   - the shape of both (dactylLayout): the photos in
//     ~/drive/MyProjects/dactyl/dactyl-keyboard (github.com/wkozyra95/dactyl-keyboard),
//     things/03-iteration is the ZMK build and things/02-iteration the QMK one
//
// Both keymaps list the keys in the same order: three rows of ten, the two
// extra keys of each half, then the two rows of the thumb blocks. The legends
// are described at the layer type.

// The layers are named by their number, the ones of the alt set with an A in
// front. The firmware calls them differently: `L_0` is default, `L_1` is raise,
// `L_2` is lower, `AL_0` is alt_default and `AL_1` is alt_raise.
//
// What the legends of the keys that are more than a single key say:
//   - the name of a layer: the layer is on while the key is held (`&mo` of
//     ZMK, `MO()` of QMK)
//   - `+Ctrl`, `+Alt` under the name of a layer: held together with the layer
//   - `SWITCH` above the name of a layer: the key switches to the layer and it
//     stays on (`&to`)
//   - `2xtap` and what follows it, in the small lines: what tapping the key
//     twice does. That is all a key with nothing above those lines does
//
// SUPER is the GUI key of the firmware ($mod of sway).

// dactylZmk is the wireless build. Its keymap holds two sets of
// layers: tapping the LALT key twice switches to the alt one (`&to 3`), the
// `SWITCH L_0` key of that one switches back (`&to 0`). L_2 is reached from both.
var dactylZmk = keyboard{
	name:   "dactyl-zmk",
	layout: dactylLayout(),
	layerGroups: []layerGroup{
		{name: "main", layers: []layer{
			// The first two keys of the fourth row are tap dances: td_cad does
			// nothing on a single tap and sends Ctrl+Alt+Del on a double one,
			// td_alt is LALT and switches to AL_0 on a double tap. The last one
			// (ctl_alt_m2) holds Ctrl+Alt and L_2.
			{name: "L_0", keys: []string{
				"Q      W      E      R      T                    Y      U      I      O      P",
				"A      S      D      F      G                    H      J      K      L      ;",
				"Z      X      C      V      B                    N      M      ,      .      /",
				"       ||2xtap|Ctrl+Alt|+Del LALT||2xtap|SWITCH|AL_0            BSPC   L_2|+Ctrl|+Alt",
				"                     ESC    SUPER  L_2     L_2    SPACE  TAB",
				"                     LSHFT  LCTL   RALT    ENTER  L_1    ENTER",
			}},
			{name: "L_1", keys: []string{
				"1      2      3      4      5                    6      7      8      9      0",
				"`      ___    DEL    ___    ___                  LEFT   DOWN   UP     RIGHT  '",
				"___    ___    ___    ___    ___                  -      =      [      ]      \\",
				"       ___    ___                                              DEL    ___",
				"                     ___    ___    ___     ___    ___    ___",
				"                     ___    ___    ___     ___    ***    ___",
			}},
			// the last key of the fourth row is held by the Ctl+Alt one of L_0
			{name: "L_2", keys: []string{
				"F1     F2     F3     F4     F5                   F6     F7     F8     F9     F10",
				"F11    F12    OUT|USB OUT|BLE ___                PSCRN  VOL-   VOL+   MUTE   PLAY|PAUSE",
				"BT|0   BT|1   BT|2   BT|3   BT|CLR               ___    ___    ___    ___    ___",
				"       ___    BOOT                                             BOOT   ***",
				"                     ___    ___    ***     ***    ___    ___",
				"                     ___    ___    ___     ___    ___    ___",
			}},
		}},
		{name: "alt", layers: []layer{
			{name: "AL_0", keys: []string{
				"Q      E      W      R      T                    Y      U      I      O      P",
				"F      A      S      D      G                    H      J      K      L      ;",
				"Z      X      C      V      B                    N      M      ,      .      /",
				"       TAB    ENTER                                            BSPC   SWITCH|L_0",
				"                     ESC    AL_1   SPACE   L_2    SPACE  TAB",
				"                     LSHFT  LCTL   RALT    ENTER  AL_1   ENTER",
			}},
			{name: "AL_1", keys: []string{
				"1      2      3      4      5                    F6     F7     F8     F9     F10",
				"6      7      8      9      0                    LEFT   DOWN   UP     RIGHT  '",
				"F1     F2     F3     F4     F5                   -      =      [      ]      \\",
				"       ___    ___                                              DEL    ___",
				"                     ___    ***    ___     ___    ___    ___",
				"                     ___    ___    ___     ___    ***    ___",
			}},
		}},
	},
}

// dactylQmk is the wired build, with a single set of layers.
var dactylQmk = keyboard{
	name:   "dactyl-qmk",
	layout: dactylLayout(),
	layerGroups: []layerGroup{
		{layers: []layer{
			{name: "L_0", keys: []string{
				"Q      W      E      R      T                    Y      U      I      O      P",
				"A      S      D      F      G                    H      J      K      L      ;",
				"Z      X      C      V      B                    N      M      ,      .      /",
				"       L_2    LALT                                             BSPC   L_2",
				"                     ESC    SUPER  L_2     DEL    SPACE  TAB",
				"                     LSHFT  LCTL   RALT    BSPC   L_1    ENTER",
			}},
			{name: "L_1", keys: []string{
				"1      2      3      4      5                    6      7      8      9      0",
				"`      ___    DEL    ___    ___                  LEFT   DOWN   UP     RIGHT  '",
				"___    ___    ___    ___    ___                  -      =      [      ]      \\",
				"       ___    ___                                              ___    ___",
				"                     ___    ___    ___     ___    ___    ___",
				"                     ___    ___    ___     ___    ***    ___",
			}},
			// BOOT is on a key that is `L_2` on L_0, so it needs one of the other two
			// held
			{name: "L_2", keys: []string{
				"F1     F2     F3     F4     F5                   F6     F7     F8     F9     F10",
				"F11    F12    ___    ___    ___                  PSCRN  VOL-   VOL+   MUTE   PLAY|PAUSE",
				"___    ___    ___    ___    ___                  ___    ___    ___    ___    ___",
				"       BOOT   ___                                              ___    ***",
				"                     ___    ___    ***     ___    ___    ___",
				"                     ___    ___    ___     ___    ___    ___",
			}},
		}},
	},
}

// dactylLayout is the shape both builds share: per half three rows of five
// keys in staggered columns, two more keys under the ring and the middle
// finger, and a block of two rows of three keys for the thumb, turned inwards.
// The right half mirrors the left one.
func dactylLayout() []keyPosition {
	const (
		// of the whole keyboard, in keys: two halves of five columns and the
		// gap between them
		width = 14.5
		// the middle of the left thumb block and the angle it is turned by
		thumbX, thumbY = 5.0, 4.75
		thumbAngle     = 12.0
	)
	// how far down a column starts, from the pinky to the index finger
	stagger := []float64{0.4, 0.15, 0, 0.15, 0.15}
	// a key of a column of the left half, and of the column mirroring it
	left := func(column int, row float64) keyPosition {
		return keyPosition{x: float64(column) + 0.5, y: row + stagger[column] + 0.5}
	}
	right := func(column int, row float64) keyPosition {
		key := left(column, row)
		key.x = width - key.x
		return key
	}
	// a key of a thumb block, column and row are counted from its middle
	thumb := func(middleX, column, row, angle float64) keyPosition {
		sin, cos := math.Sincos(angle * math.Pi / 180)
		return keyPosition{
			x:        middleX + column*cos - row*sin,
			y:        thumbY + column*sin + row*cos,
			rotation: angle,
		}
	}

	var keys []keyPosition
	for row := 0.0; row < 3; row++ {
		for column := 0; column < 5; column++ {
			keys = append(keys, left(column, row))
		}
		for column := 4; column >= 0; column-- {
			keys = append(keys, right(column, row))
		}
	}
	keys = append(keys, left(1, 3), left(2, 3), right(2, 3), right(1, 3))
	for _, row := range []float64{-0.5, 0.5} {
		for _, column := range []float64{-1, 0, 1} {
			keys = append(keys, thumb(thumbX, column, row, thumbAngle))
		}
		for _, column := range []float64{-1, 0, 1} {
			keys = append(keys, thumb(width-thumbX, column, row, -thumbAngle))
		}
	}
	return keys
}
