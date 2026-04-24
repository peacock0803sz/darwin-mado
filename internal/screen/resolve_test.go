package screen_test

import (
	"errors"
	"testing"

	"github.com/peacock0803sz/mado/internal/ax"
	"github.com/peacock0803sz/mado/internal/screen"
)

var (
	screenBuiltin = ax.Screen{ID: 42, Name: "Built-in Retina Display", UUID: "37D8832A-2D66-02CA-B9F7-8F30A301B230", IsPrimary: true}
	screenDellA   = ax.Screen{ID: 101, Name: "DELL U2720Q", UUID: "12345678-90AB-CDEF-1234-567890ABCDEF"}
	screenDellB   = ax.Screen{ID: 102, Name: "DELL U2720Q", UUID: "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE"}
	screenNoUUID  = ax.Screen{ID: 200, Name: "Sidecar Display", UUID: ""}
)

func TestResolve_UUIDHit(t *testing.T) {
	screens := []ax.Screen{screenBuiltin, screenDellA}
	got, err := screen.Resolve(screenBuiltin.UUID, screens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != screenBuiltin.ID {
		t.Errorf("got.ID = %d, want %d", got.ID, screenBuiltin.ID)
	}
}

func TestResolve_NameHit(t *testing.T) {
	screens := []ax.Screen{screenBuiltin, screenDellA}
	got, err := screen.Resolve("dell u2720q", screens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != screenDellA.ID {
		t.Errorf("got.ID = %d, want %d", got.ID, screenDellA.ID)
	}
}

func TestResolve_NumericIDHit(t *testing.T) {
	screens := []ax.Screen{screenBuiltin, screenDellA}
	got, err := screen.Resolve("101", screens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != screenDellA.ID {
		t.Errorf("got.ID = %d, want %d", got.ID, screenDellA.ID)
	}
}

func TestResolve_NotFound(t *testing.T) {
	screens := []ax.Screen{screenBuiltin}
	_, err := screen.Resolve("NoSuchDisplay", screens)
	var notFound *screen.ScreenNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("got err = %v (%T), want *ScreenNotFoundError", err, err)
	}
	if len(notFound.Available) != 1 {
		t.Errorf("len(Available) = %d, want 1", len(notFound.Available))
	}
}

func TestResolve_AmbiguousName(t *testing.T) {
	screens := []ax.Screen{screenDellA, screenDellB}
	_, err := screen.Resolve("DELL U2720Q", screens)
	var amb *screen.AmbiguousScreenError
	if !errors.As(err, &amb) {
		t.Fatalf("got err = %v (%T), want *AmbiguousScreenError", err, err)
	}
	if len(amb.Candidates) != 2 {
		t.Errorf("len(Candidates) = %d, want 2", len(amb.Candidates))
	}
}

func TestResolve_UUIDWinsOverName(t *testing.T) {
	// A screen whose Name happens to equal another screen's UUID form — the UUID
	// filter must stage-1 hit the owner of that UUID, not stage-2 match by name.
	uuid := "37D8832A-2D66-02CA-B9F7-8F30A301B230"
	s1 := ax.Screen{ID: 1, Name: uuid, UUID: "00000000-0000-0000-0000-000000000000"}
	s2 := ax.Screen{ID: 2, Name: "DELL", UUID: uuid}
	got, err := screen.Resolve(uuid, []ax.Screen{s1, s2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != s2.ID {
		t.Errorf("got.ID = %d, want %d (UUID stage should win)", got.ID, s2.ID)
	}
}

func TestResolve_EmptyFilter(t *testing.T) {
	_, err := screen.Resolve("", []ax.Screen{screenBuiltin})
	var notFound *screen.ScreenNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("got err = %v (%T), want *ScreenNotFoundError", err, err)
	}
}

func TestResolve_IgnoresEmptyUUIDScreens(t *testing.T) {
	// A screen with empty UUID must not be matched by stage 1 even if the
	// filter is also empty-string-looking (e.g. " ").
	screens := []ax.Screen{screenNoUUID, screenBuiltin}
	_, err := screen.Resolve("", screens)
	var notFound *screen.ScreenNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("empty filter should not match the empty-UUID screen")
	}
}
