package mode_test

import (
	"flag"
	"testing"

	"github.com/dobyte/due/mode"
)

func TestGetMode(t *testing.T) {
	flag.Parse()

	t.Log(mode.GetMode())
}
