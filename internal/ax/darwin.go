//go:build darwin

package ax

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics -framework CoreFoundation

#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <stdint.h>

// --- CGS Private API (SkyLight.framework via dlsym) ---

#define kCGSAllSpacesMask 0x7

typedef int CGSConnectionID;
typedef CFDictionaryRef (*CGSCopySpacesForWindows_f)(CGSConnectionID, int, CFArrayRef);
typedef CFArrayRef      (*CGSCopyManagedDisplaySpaces_f)(CGSConnectionID);
typedef CGSConnectionID (*CGSMainConnectionID_f)(void);

static CGSCopySpacesForWindows_f     _cgs_spaces_for_windows;
static CGSCopyManagedDisplaySpaces_f _cgs_managed_display_spaces;
static CGSMainConnectionID_f         _cgs_main_connection_id;
static int _cgsAvailable;

static void cgs_init(void) {
    void *sl = dlopen(
        "/System/Library/PrivateFrameworks/SkyLight.framework/SkyLight",
        RTLD_LAZY | RTLD_GLOBAL);
    if (!sl) return;
    _cgs_main_connection_id      = (CGSMainConnectionID_f)dlsym(sl, "CGSMainConnectionID");
    _cgs_spaces_for_windows      = (CGSCopySpacesForWindows_f)dlsym(sl, "CGSCopySpacesForWindows");
    _cgs_managed_display_spaces  = (CGSCopyManagedDisplaySpaces_f)dlsym(sl, "CGSCopyManagedDisplaySpaces");
    if (_cgs_main_connection_id && _cgs_spaces_for_windows && _cgs_managed_display_spaces) {
        _cgsAvailable = 1;
    }
}

static int cgs_is_available(void) { return _cgsAvailable; }

static CGSConnectionID cgs_get_cid(void) { return _cgs_main_connection_id(); }

// Batch: returns CFDictionaryRef{windowID -> [spaceID,...]}. Caller must CFRelease.
static CFDictionaryRef cgs_copy_spaces_for_windows(CGSConnectionID cid, CFArrayRef wids) {
    return _cgs_spaces_for_windows(cid, kCGSAllSpacesMask, wids);
}

// Returns CFArrayRef of display info dicts. Caller must CFRelease.
static CFArrayRef cgs_copy_managed_display_spaces(CGSConnectionID cid) {
    return _cgs_managed_display_spaces(cid);
}

// Build a CFArray of uint32 window IDs. Caller must CFRelease.
static CFArrayRef cg_make_wid_array(const uint32_t *ids, int n) {
    CFMutableArrayRef a = CFArrayCreateMutable(NULL, n, &kCFTypeArrayCallBacks);
    if (!a) return NULL;
    for (int i = 0; i < n; i++) {
        CFNumberRef num = CFNumberCreate(NULL, kCFNumberSInt32Type, (const int32_t*)&ids[i]);
        if (num) { CFArrayAppendValue(a, num); CFRelease(num); }
    }
    return a;
}

// Look up per-window space IDs in batch result dict by window ID.
// Returns NULL if window not present in dict. Do NOT CFRelease the result.
static CFArrayRef cgs_window_space_ids(CFDictionaryRef dict, uint32_t wid) {
    if (!dict) return NULL;
    CFNumberRef key = CFNumberCreate(NULL, kCFNumberSInt32Type, (const int32_t*)&wid);
    if (!key) return NULL;
    CFTypeRef val = CFDictionaryGetValue(dict, key);
    CFRelease(key);
    if (!val) return NULL;
    return (CFArrayRef)val;
}

// Get a space ID (int64) from a space IDs array at index.
static int cgs_get_space_id(CFArrayRef arr, int idx, int64_t *out) {
    if (!arr || idx >= (int)CFArrayGetCount(arr)) return 0;
    CFNumberRef num = (CFNumberRef)CFArrayGetValueAtIndex(arr, (CFIndex)idx);
    if (!num) return 0;
    return CFNumberGetValue(num, kCFNumberSInt64Type, out) ? 1 : 0;
}

// Get "Spaces" CFArrayRef from a display dict. Do NOT CFRelease the result.
static CFArrayRef cg_display_spaces(CFDictionaryRef d) {
    if (!d) return NULL;
    CFTypeRef val = CFDictionaryGetValue(d, CFSTR("Spaces"));
    if (!val) return NULL;
    return (CFArrayRef)val;
}

// Get id64 (space ID) from a space info dict.
static int cg_space_id(CFDictionaryRef d, int64_t *out) {
    if (!d) return 0;
    CFNumberRef num = (CFNumberRef)CFDictionaryGetValue(d, CFSTR("id64"));
    if (!num) return 0;
    return CFNumberGetValue(num, kCFNumberSInt64Type, out) ? 1 : 0;
}

// Safe CFRelease helpers for types not covered by existing helpers.
static void cf_release_dict(CFDictionaryRef d) { if (d) CFRelease(d); }

// Check Accessibility permission
int ax_is_trusted() {
    return AXIsProcessTrusted() ? 1 : 0;
}

// Retrieve all window info from CGWindowList (onscreen only)
CFArrayRef cg_list_windows() {
    return CGWindowListCopyWindowInfo(
        kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements,
        kCGNullWindowID
    );
}

// Retrieve an int32 value from a CFDictionary
int cg_dict_int(CFDictionaryRef dict, CFStringRef key, int32_t *out) {
    CFNumberRef num = (CFNumberRef)CFDictionaryGetValue(dict, key);
    if (!num) return 0;
    return CFNumberGetValue(num, kCFNumberSInt32Type, out) ? 1 : 0;
}

// Retrieve a CFStringRef value from a CFDictionary
CFStringRef cg_dict_string(CFDictionaryRef dict, CFStringRef key) {
    return (CFStringRef)CFDictionaryGetValue(dict, key);
}

// Retrieve the kCGWindowBounds value as a CFDictionaryRef (NULL if absent)
CFDictionaryRef cg_dict_bounds(CFDictionaryRef dict) {
    CFTypeRef val = CFDictionaryGetValue(dict, kCGWindowBounds);
    if (!val) return NULL;
    return (CFDictionaryRef)val;
}

// Parse a bounds dictionary as a CGRect
int cg_parse_bounds(CFDictionaryRef boundsDict, int *x, int *y, int *w, int *h) {
    CGRect rect;
    if (!CGRectMakeWithDictionaryRepresentation(boundsDict, &rect)) return 0;
    *x = (int)rect.origin.x;
    *y = (int)rect.origin.y;
    *w = (int)rect.size.width;
    *h = (int)rect.size.height;
    return 1;
}

// Convert a CFStringRef to a UTF-8 C string (caller must free)
char* cf_to_cstr(CFStringRef s) {
    if (!s) return NULL;
    CFIndex len = CFStringGetMaximumSizeForEncoding(
        CFStringGetLength(s), kCFStringEncodingUTF8) + 1;
    char *buf = (char*)malloc(len);
    if (!buf) return NULL;
    if (!CFStringGetCString(s, buf, len, kCFStringEncodingUTF8)) {
        free(buf);
        return NULL;
    }
    return buf;
}

// AX API: retrieve the window array for a PID and return a CFArrayRef that the caller must CFRelease
CFArrayRef ax_windows_for_pid(pid_t pid) {
    AXUIElementRef app = AXUIElementCreateApplication(pid);
    if (!app) return NULL;
    CFTypeRef ref = NULL;
    AXError err = AXUIElementCopyAttributeValue(app, kAXWindowsAttribute, &ref);
    CFRelease(app);
    if (err != kAXErrorSuccess || !ref) return NULL;
    return (CFArrayRef)ref;
}

// AX API: return the window title (caller must free)
char* ax_window_title(AXUIElementRef win) {
    CFTypeRef ref = NULL;
    if (AXUIElementCopyAttributeValue(win, kAXTitleAttribute, &ref) != kAXErrorSuccess) return NULL;
    char *s = cf_to_cstr((CFStringRef)ref);
    CFRelease(ref);
    return s;
}

// AX API: return whether the window is minimized (1=minimized, 0=not)
int ax_is_minimized(AXUIElementRef win) {
    CFTypeRef ref = NULL;
    if (AXUIElementCopyAttributeValue(win, kAXMinimizedAttribute, &ref) != kAXErrorSuccess) return 0;
    int result = (ref == kCFBooleanTrue) ? 1 : 0;
    CFRelease(ref);
    return result;
}

// AX API: set the window position (returns 0 on success)
int ax_set_position(AXUIElementRef win, double x, double y) {
    CGPoint p = CGPointMake(x, y);
    AXValueRef val = AXValueCreate(kAXValueCGPointType, &p);
    if (!val) return (int)kAXErrorFailure;
    int err = (int)AXUIElementSetAttributeValue(win, kAXPositionAttribute, val);
    CFRelease(val);
    return err;
}

// AX API: set the window size (returns 0 on success)
int ax_set_size(AXUIElementRef win, double w, double h) {
    CGSize s = CGSizeMake(w, h);
    AXValueRef val = AXValueCreate(kAXValueCGSizeType, &s);
    if (!val) return (int)kAXErrorFailure;
    int err = (int)AXUIElementSetAttributeValue(win, kAXSizeAttribute, val);
    CFRelease(val);
    return err;
}

// Null-check helpers (CF types cannot be compared directly to nil in cgo)
int cf_array_is_null(CFArrayRef a)       { return a == NULL ? 1 : 0; }
int cf_string_is_null(CFStringRef s)     { return s == NULL ? 1 : 0; }
int cf_dict_is_null(CFDictionaryRef d)   { return d == NULL ? 1 : 0; }
int cf_type_is_null(CFTypeRef t)         { return t == NULL ? 1 : 0; }
int cstr_is_null(const char *s)          { return s == NULL ? 1 : 0; }

// Safe CFRelease helpers: avoid Go-side CFArrayRef→CFTypeRef cast which can
// produce a NULL pointer on ARM64 due to cgo opaque-struct reinterpretation.
void cf_release_array(CFArrayRef a)      { if (a) CFRelease(a); }
*/
import "C"

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// cgsOnce guards one-time initialization of the SkyLight CGS private API.
var cgsOnce sync.Once

func ensureCGS() {
	cgsOnce.Do(func() { C.cgs_init() })
}

// windowEntry pairs a resolved Window with its CGWindowID for the CGS batch call.
type windowEntry struct {
	win  *Window
	cgID uint32
}

// buildSpaceMap queries CGSCopyManagedDisplaySpaces and returns a map from
// Space ID (int64) to 1-based desktop number (Mission Control order).
// Returns nil on API failure; a nil map lookup returns (0, false) safely in Go.
func buildSpaceMap(cid C.CGSConnectionID) map[int64]int {
	displaySpaces := C.cgs_copy_managed_display_spaces(cid)
	if C.cf_array_is_null(displaySpaces) != 0 {
		return nil
	}
	defer C.cf_release_array(displaySpaces)

	spaceMap := make(map[int64]int)
	desktopNum := 1
	displayCount := int(C.CFArrayGetCount(displaySpaces))

	for i := 0; i < displayCount; i++ {
		displayDict := C.CFDictionaryRef(C.CFArrayGetValueAtIndex(displaySpaces, C.CFIndex(i)))
		if C.cf_dict_is_null(displayDict) != 0 {
			continue
		}
		spacesArr := C.cg_display_spaces(displayDict)
		if C.cf_array_is_null(spacesArr) != 0 {
			continue
		}
		spaceCount := int(C.CFArrayGetCount(spacesArr))
		for j := 0; j < spaceCount; j++ {
			spaceDict := C.CFDictionaryRef(C.CFArrayGetValueAtIndex(spacesArr, C.CFIndex(j)))
			if C.cf_dict_is_null(spaceDict) != 0 {
				desktopNum++
				continue
			}
			var sid C.int64_t
			if C.cg_space_id(spaceDict, &sid) != 0 {
				spaceMap[int64(sid)] = desktopNum
			}
			desktopNum++
		}
	}
	return spaceMap
}

const (
	axRetryCount    = 3
	axRetryInterval = 100 * time.Millisecond
)

// withRetry executes fn up to axRetryCount times, sleeping axRetryInterval between attempts.
// It aborts early if ctx is cancelled.
func withRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := 0; i < axRetryCount; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(axRetryInterval):
			}
		}
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

// darwinService is the darwin implementation of WindowService using the AX API.
type darwinService struct{}

// NewWindowService returns a WindowService backed by the real AX API.
func NewWindowService() WindowService {
	return &darwinService{}
}

func (s *darwinService) CheckPermission() error {
	if C.ax_is_trusted() == 0 {
		return &PermissionError{}
	}
	return nil
}

// ListWindows retrieves all windows using CGWindowList and the AX API (T018).
func (s *darwinService) ListWindows(ctx context.Context) ([]Window, error) {
	if err := s.CheckPermission(); err != nil {
		return nil, err
	}

	screens, err := s.ListScreens(ctx)
	if err != nil {
		return nil, err
	}

	infoList := C.cg_list_windows()
	if C.cf_array_is_null(infoList) != 0 {
		return nil, nil
	}
	defer C.CFRelease(C.CFTypeRef(infoList))

	count := int(C.CFArrayGetCount(infoList))
	entries := make([]windowEntry, 0, count)

	// Cache the AX window array per PID to avoid redundant calls
	axCache := make(map[uint32]C.CFArrayRef)
	defer func() {
		for _, arr := range axCache {
			C.cf_release_array(arr)
		}
	}()

	// CGWindowList has one entry per window.
	// When a PID has multiple windows, the AX-side index must be tracked.
	// AX API window list takes precedence; CGWindowList is used as supplemental info.
	pidWindowIndex := make(map[uint32]int) // per-PID AX index counter

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		dictRef := C.CFArrayGetValueAtIndex(infoList, C.CFIndex(i))
		if dictRef == nil {
			continue
		}
		dict := C.CFDictionaryRef(dictRef)

		entry := windowFromCGInfo(dict, screens, axCache, pidWindowIndex)
		if entry == nil {
			continue
		}
		entries = append(entries, *entry)
	}

	// Populate Desktop field via CGS batch call (T005b).
	// Desktop defaults to -1 (unknown) from windowFromCGInfo; only overwrite on success.
	ensureCGS()
	if C.cgs_is_available() != 0 && len(entries) > 0 {
		cgIDs := make([]C.uint32_t, len(entries))
		for i, e := range entries {
			cgIDs[i] = C.uint32_t(e.cgID)
		}
		widArray := C.cg_make_wid_array(&cgIDs[0], C.int(len(cgIDs)))
		if C.cf_array_is_null(widArray) == 0 {
			defer C.cf_release_array(widArray)

			cid := C.cgs_get_cid()
			spaceMap := buildSpaceMap(cid)

			batchResult := C.cgs_copy_spaces_for_windows(cid, widArray)
			if C.cf_dict_is_null(batchResult) == 0 {
				defer C.cf_release_dict(batchResult)
				for i := range entries {
					spaceIDs := C.cgs_window_space_ids(batchResult, C.uint32_t(entries[i].cgID))
					if C.cf_array_is_null(spaceIDs) != 0 {
						continue
					}
					switch int(C.CFArrayGetCount(spaceIDs)) {
					case 0:
						// Window is not on any space (e.g., minimized). Desktop remains -1.
					case 1:
						var sid C.int64_t
						if C.cgs_get_space_id(spaceIDs, 0, &sid) != 0 {
							if dn, ok := spaceMap[int64(sid)]; ok {
								entries[i].win.Desktop = dn
							}
						}
					default:
						// Present on multiple spaces = "all desktops"
						entries[i].win.Desktop = 0
					}
				}
			}
		}
	}

	windows := make([]Window, len(entries))
	for i, e := range entries {
		windows[i] = *e.win
	}
	return windows, nil
}

// ListScreens returns all connected screens using CGGetActiveDisplayList (T058).
func (s *darwinService) ListScreens(ctx context.Context) ([]Screen, error) {
	var displayIDs [32]C.CGDirectDisplayID
	var count C.uint32_t

	if C.CGGetActiveDisplayList(32, &displayIDs[0], &count) != C.kCGErrorSuccess {
		return nil, fmt.Errorf("CGGetActiveDisplayList failed")
	}

	primaryID := C.CGMainDisplayID()
	screens := make([]Screen, 0, int(count))

	for i := 0; i < int(count); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		id := displayIDs[i]
		bounds := C.CGDisplayBounds(id)

		screens = append(screens, Screen{
			ID:        uint32(id),
			Name:      fmt.Sprintf("Display %d", uint32(id)),
			X:         int(bounds.origin.x),
			Y:         int(bounds.origin.y),
			Width:     int(bounds.size.width),
			Height:    int(bounds.size.height),
			IsPrimary: id == primaryID,
		})
	}

	return screens, nil
}

// MoveWindow moves the specified window to a new position (T025).
func (s *darwinService) MoveWindow(ctx context.Context, pid uint32, title string, x, y int) error {
	if err := s.CheckPermission(); err != nil {
		return err
	}

	return withRetry(ctx, func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		win, err := findAXWindow(pid, title)
		if err != nil {
			return err
		}
		defer C.CFRelease(C.CFTypeRef(win))

		if ret := C.ax_set_position(win, C.double(x), C.double(y)); ret != 0 {
			return fmt.Errorf("AXUIElementSetAttributeValue(position) failed: %d", ret)
		}
		return nil
	})
}

// ResizeWindow resizes the specified window (T025).
func (s *darwinService) ResizeWindow(ctx context.Context, pid uint32, title string, w, h int) error {
	if err := s.CheckPermission(); err != nil {
		return err
	}

	return withRetry(ctx, func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		win, err := findAXWindow(pid, title)
		if err != nil {
			return err
		}
		defer C.CFRelease(C.CFTypeRef(win))

		if ret := C.ax_set_size(win, C.double(w), C.double(h)); ret != 0 {
			return fmt.Errorf("AXUIElementSetAttributeValue(size) failed: %d", ret)
		}
		return nil
	})
}

// --- Helper functions ---

// windowFromCGInfo builds a windowEntry from a CGWindowInfo dictionary.
// The entry pairs the resolved Window with its CGWindowID for the CGS batch lookup.
func windowFromCGInfo(
	dict C.CFDictionaryRef,
	screens []Screen,
	axCache map[uint32]C.CFArrayRef,
	pidWindowIndex map[uint32]int,
) *windowEntry {
	// Extract CGWindowID via kCGWindowNumber (needed for CGS Space lookup).
	var cgWinNum C.int32_t
	C.cg_dict_int(dict, C.kCGWindowNumber, &cgWinNum) // ignore return: 0 is safe fallback

	// Retrieve the app PID via kCGWindowOwnerPID
	var pid C.int32_t
	if C.cg_dict_int(dict, C.kCGWindowOwnerPID, &pid) == 0 || pid == 0 {
		return nil
	}
	appPID := uint32(pid)

	// Retrieve the app name via kCGWindowOwnerName
	nameRef := C.cg_dict_string(dict, C.kCGWindowOwnerName)
	if C.cf_string_is_null(nameRef) != 0 {
		return nil
	}
	appNameCS := C.cf_to_cstr(nameRef)
	if C.cstr_is_null(appNameCS) != 0 {
		return nil
	}
	appName := C.GoString(appNameCS)
	C.free(unsafe.Pointer(appNameCS))
	if appName == "" {
		return nil
	}

	// Retrieve position and size via kCGWindowBounds
	boundsDict := C.cg_dict_bounds(dict)
	var x, y, width, height int
	if C.cf_dict_is_null(boundsDict) == 0 {
		var cx, cy, cw, ch C.int
		if C.cg_parse_bounds(boundsDict, &cx, &cy, &cw, &ch) != 0 {
			x, y, width, height = int(cx), int(cy), int(cw), int(ch)
		}
	}

	// Zero-size windows are menu-bar-only apps or similar; skip them
	if width == 0 && height == 0 {
		return nil
	}

	// Retrieve window title and state via the AX API
	axArr, ok := axCache[appPID]
	if !ok {
		axArr = C.ax_windows_for_pid(C.pid_t(appPID))
		axCache[appPID] = axArr // cache even if nil
	}

	title := ""
	state := StateNormal

	if C.cf_array_is_null(axArr) == 0 {
		idx := pidWindowIndex[appPID]
		pidWindowIndex[appPID]++

		if idx < int(C.CFArrayGetCount(axArr)) {
			win := C.AXUIElementRef(C.CFArrayGetValueAtIndex(axArr, C.CFIndex(idx)))

			titleCS := C.ax_window_title(win)
			if titleCS != nil {
				title = C.GoString(titleCS)
				C.free(unsafe.Pointer(titleCS))
			}

			if C.ax_is_minimized(win) != 0 {
				state = StateMinimized
			}
		}
	}

	// Derive screen_id from the screen with the largest intersection area
	screenID, screenName := deriveScreen(x, y, width, height, screens)
	if state == StateMinimized || state == StateHidden {
		screenID = 0
		screenName = ""
	}

	return &windowEntry{
		win: &Window{
			AppName:    appName,
			Title:      title,
			PID:        appPID,
			X:          x,
			Y:          y,
			Width:      width,
			Height:     height,
			State:      state,
			ScreenID:   screenID,
			ScreenName: screenName,
			Desktop:    -1,
		},
		cgID: uint32(cgWinNum),
	}
}

// deriveScreen returns the screen with the largest intersection area with the given window rectangle.
func deriveScreen(wx, wy, ww, wh int, screens []Screen) (uint32, string) {
	maxArea := 0
	var bestID uint32
	var bestName string

	for _, s := range screens {
		ix1 := max(wx, s.X)
		iy1 := max(wy, s.Y)
		ix2 := min(wx+ww, s.X+s.Width)
		iy2 := min(wy+wh, s.Y+s.Height)

		if ix2 > ix1 && iy2 > iy1 {
			area := (ix2 - ix1) * (iy2 - iy1)
			if area > maxArea {
				maxArea = area
				bestID = s.ID
				bestName = s.Name
			}
		}
	}
	return bestID, bestName
}

// findAXWindow searches for an AXUIElementRef by PID and title (caller must CFRelease).
func findAXWindow(pid uint32, title string) (C.AXUIElementRef, error) {
	arr := C.ax_windows_for_pid(C.pid_t(pid))
	if C.cf_array_is_null(arr) != 0 {
		return 0, fmt.Errorf("no AX windows for pid %d", pid)
	}
	defer C.CFRelease(C.CFTypeRef(arr))

	count := int(C.CFArrayGetCount(arr))
	for i := 0; i < count; i++ {
		win := C.AXUIElementRef(C.CFArrayGetValueAtIndex(arr, C.CFIndex(i)))

		titleCS := C.ax_window_title(win)
		if titleCS == nil {
			continue
		}
		t := C.GoString(titleCS)
		C.free(unsafe.Pointer(titleCS))

		if t == title {
			C.CFRetain(C.CFTypeRef(win))
			return win, nil
		}
	}
	return 0, fmt.Errorf("window not found: pid=%d title=%q", pid, title)
}
