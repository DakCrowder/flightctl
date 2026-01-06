package podman

import (
	"context"
	"fmt"
	"os/exec"
)

// EventsSinceCmd returns a command to get podman events since the given time.
// After creating the command, it should be started with exec.Start().
// When the events are in sync with the current time a sync event is emitted.
func (c *Client) EventsSinceCmd(ctx context.Context, events []string, sinceTime string) *exec.Cmd {
	args := []string{"events", "--format", "json", "--since", sinceTime}
	for _, event := range events {
		args = append(args, "--filter", fmt.Sprintf("event=%s", event))
	}

	return c.exec.CommandContext(ctx, podmanCmd, args...)
}
