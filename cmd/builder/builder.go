package builder

import (
	"github.com/spf13/cobra"
)

type SubCommandGroup struct {
	Commands []*cobra.Command
}

func (g SubCommandGroup) AddGroup(c *cobra.Command) {
	for _, cmd := range g.Commands {
		c.AddCommand(cmd)
	}
}
