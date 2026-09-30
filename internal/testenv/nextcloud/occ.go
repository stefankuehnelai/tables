package nextcloud

import (
	"context"
	"fmt"
	"io"
	"strings"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

// OCC runs a Nextcloud occ command as the web server user.
func (n *Nextcloud) OCC(ctx context.Context, args ...string) (string, error) {
	if n == nil || n.container == nil {
		return "", fmt.Errorf("Nextcloud container is not available")
	}
	command := append([]string{"php", "occ"}, args...)
	code, output, err := n.container.Exec(ctx, command, tcexec.WithUser("www-data"))
	if err != nil {
		return "", fmt.Errorf("execute occ %s: %w", strings.Join(args, " "), err)
	}
	body, readErr := io.ReadAll(output)
	if readErr != nil {
		return "", fmt.Errorf("read occ output: %w", readErr)
	}
	if code != 0 {
		return "", fmt.Errorf("occ %s: exit %d: %s", strings.Join(args, " "), code, strings.TrimSpace(string(body)))
	}
	return strings.TrimSpace(string(body)), nil
}
