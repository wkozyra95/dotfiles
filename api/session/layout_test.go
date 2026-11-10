package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColumnLayoutCommands(t *testing.T) {
	cases := []struct {
		name     string
		first    int64
		columns  [][]int64
		expected string
	}{
		{
			name:    "row keeps launch order",
			first:   1,
			columns: [][]int64{{1, 2, 3}},
			expected: "[con_id=1] layout splith; " +
				"[con_id=1] mark --add _session_layout; [con_id=2] move container to mark _session_layout; " +
				"[con_id=2] mark --add _session_layout; [con_id=3] move container to mark _session_layout; " +
				"unmark _session_layout",
		},
		{
			name:    "one window per column",
			first:   1,
			columns: [][]int64{{1}, {2}},
			expected: "[con_id=1] layout splith; " +
				"[con_id=1] mark --add _session_layout; [con_id=2] move container to mark _session_layout; " +
				"unmark _session_layout",
		},
		{
			name:    "left column is swapped into first place and both columns stack",
			first:   4,
			columns: [][]int64{{1, 2}, {3, 4}},
			expected: "[con_id=1] layout splith; " +
				"[con_id=1] swap container with con_id 4; " +
				"[con_id=1] mark --add _session_layout; [con_id=3] move container to mark _session_layout; " +
				"[con_id=1] splitv; " +
				"[con_id=1] mark --add _session_layout; [con_id=2] move container to mark _session_layout; " +
				"[con_id=3] splitv; " +
				"[con_id=3] mark --add _session_layout; [con_id=4] move container to mark _session_layout; " +
				"unmark _session_layout",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, strings.Join(columnLayoutCommands(c.first, c.columns), "; "))
		})
	}
}
