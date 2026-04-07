package config

import (
	"os"
	"testing"

	"github.com/SomniSom/docker-ops/internal/locale"
)

func TestMain(m *testing.M) {
	locale.Set("en")
	os.Exit(m.Run())
}
