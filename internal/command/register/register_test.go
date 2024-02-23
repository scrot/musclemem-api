package register

import (
	"testing"

	"github.com/scrot/musclemem-api/internal/cli"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	cs := []struct {
		name string
	}{
		{"registerWithFlags"},
	}

	config := &cli.CLIConfig{}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			_ = NewRegisterCmd(config)
			t.Fatalf("Not implemented")
		})
	}
}
