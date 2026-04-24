package cli

import (
	"context"

	"github.com/peacock0803sz/mado/internal/ax"
)

// ResolveScreenFilterForTest exposes resolveScreenFilterErr so external-test
// packages can verify the error classification (not found / ambiguous) without
// triggering os.Exit.
func ResolveScreenFilterForTest(ctx context.Context, svc ax.WindowService, filter string) (string, error) {
	return resolveScreenFilterErr(ctx, svc, filter)
}
