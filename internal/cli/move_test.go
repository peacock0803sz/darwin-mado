package cli_test

import (
	"context"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/cli"
)

func TestMoveCmd_ResolveScreenFilter_UUIDHit(t *testing.T) {
	// The pure resolve helper shares its implementation with `mado list`;
	// this confirms that `mado move --screen <UUID>` goes through the same
	// seam and does not degrade to stage-2/3 unexpectedly.
	uuid := "BBBB0000-0000-0000-0000-000000000000"
	screens := []ax.Screen{
		{ID: 1, Name: "Built-in", UUID: "AAAA0000-0000-0000-0000-000000000000", IsPrimary: true},
		{ID: 2, Name: "DELL U2720Q", UUID: uuid},
	}
	svc := &ax.MockWindowService{Screens: screens}
	got, err := cli.ResolveScreenFilterForTest(context.Background(), svc, uuid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid {
		t.Errorf("got %q, want %q", got, uuid)
	}
}
