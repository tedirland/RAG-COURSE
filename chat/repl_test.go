package chat

import "testing"

// TestFrameAt guards the spinner index math. The original bug used
// frames[i&len(frames)] (bitwise AND), which panicked once i reached 4.
// frameAt must instead cycle through every frame and never go out of range,
// no matter how large i grows.
func TestFrameAt(t *testing.T) {
	cases := []struct {
		i    int
		want string
	}{
		{0, "|"},
		{1, "/"},
		{2, "-"},
		{3, "\\"},
		{4, "|"},
	}

	for _, c := range cases {
		if got := frameAt(c.i); got != c.want {
			t.Errorf("frameAt(%d) = %q, want %q", c.i, got, c.want)
		}
	}
}

// TestFrameAtNeverPanics walks the counter well past the frame-boundary that
// used to crash, asserting frameAt stays in range for many ticks.
func TestFrameAtNeverPanics(t *testing.T) {
	for i := range 1000 {
		_ = frameAt(i) // must not panic (index out of range)
	}
}
