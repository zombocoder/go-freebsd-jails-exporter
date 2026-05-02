//go:build freebsd && cgo

package freebsd

/*
#include <sys/types.h>
#include <sys/sysctl.h>
#include <sys/rctl.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>

// racct_get_usage retrieves RACCT usage for a jail by jid.
// The filter format is "jail:<name>" and the output is a comma-separated
// list of "resource=value" pairs.
// Returns 0 on success, errno on failure.
// buf must be pre-allocated by the caller.
static int racct_get_jail_usage(const char *jail_name, char *buf, size_t bufsz) {
	char filter[512];
	snprintf(filter, sizeof(filter), "jail:%s", jail_name);

	memset(buf, 0, bufsz);

	if (rctl_get_racct(filter, strlen(filter) + 1, buf, bufsz) < 0) {
		return errno;
	}

	return 0;
}

// check_racct_enabled checks if RACCT is available by testing kern.racct.enable.
static int check_racct_enabled(int *enabled) {
	size_t len = sizeof(*enabled);
	if (sysctlbyname("kern.racct.enable", enabled, &len, NULL, 0) < 0) {
		return errno;
	}
	return 0;
}
*/
import "C"

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

// RACCTResource represents a single RACCT resource value.
type RACCTResource struct {
	Name  string
	Value int64
}

// JailRACCT holds all RACCT resources for a jail.
type JailRACCT struct {
	Resources map[string]int64
}

// IsRACCTEnabled checks if RACCT is enabled on this system.
func IsRACCTEnabled() bool {
	var enabled C.int
	rc := C.check_racct_enabled(&enabled)
	if rc != 0 {
		return false
	}
	return enabled != 0
}

// GetJailRACCT retrieves RACCT resource usage for a jail by name.
func GetJailRACCT(jailName string) (*JailRACCT, error) {
	cname := C.CString(jailName)
	defer C.free(unsafe.Pointer(cname))

	bufsize := 4096
	buf := (*C.char)(C.malloc(C.size_t(bufsize)))
	defer C.free(unsafe.Pointer(buf))

	rc := C.racct_get_jail_usage(cname, buf, C.size_t(bufsize))
	if rc != 0 {
		return nil, fmt.Errorf("rctl_get_racct failed for jail %s: errno %d", jailName, int(rc))
	}

	result := C.GoString(buf)
	if result == "" {
		return &JailRACCT{Resources: make(map[string]int64)}, nil
	}

	return parseRACCTOutput(result)
}

func parseRACCTOutput(output string) (*JailRACCT, error) {
	racct := &JailRACCT{Resources: make(map[string]int64)}

	pairs := strings.Split(strings.TrimSpace(output), ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			continue
		}

		racct.Resources[name] = value
	}

	return racct, nil
}
