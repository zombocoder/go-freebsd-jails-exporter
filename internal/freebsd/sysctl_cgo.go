//go:build freebsd && cgo

package freebsd

/*
#include <sys/types.h>
#include <sys/sysctl.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

// sysctl_raw performs a raw sysctl call and returns the data in a malloc'd buffer.
// The caller must free the returned buffer.
static int sysctl_raw(int *name, unsigned int namelen, void **buf, size_t *bufsz) {
	*bufsz = 0;
	*buf = NULL;

	// First call to get size.
	if (sysctl(name, namelen, NULL, bufsz, NULL, 0) < 0) {
		return errno;
	}

	if (*bufsz == 0) {
		return 0;
	}

	// Add 10% buffer for race conditions (new processes appearing).
	*bufsz = *bufsz + (*bufsz / 10);

	*buf = malloc(*bufsz);
	if (*buf == NULL) {
		return ENOMEM;
	}

	if (sysctl(name, namelen, *buf, bufsz, NULL, 0) < 0) {
		int err = errno;
		free(*buf);
		*buf = NULL;
		*bufsz = 0;
		return err;
	}

	return 0;
}

// sysctl_int retrieves an integer value via sysctl by name.
static int sysctl_int_byname(const char *name, int *val) {
	size_t len = sizeof(*val);
	if (sysctlbyname(name, val, &len, NULL, 0) < 0) {
		return errno;
	}
	return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// SysctlRaw performs a raw sysctl call and returns the result as a byte slice.
func SysctlRaw(mib []int32) ([]byte, error) {
	if len(mib) == 0 {
		return nil, fmt.Errorf("empty mib")
	}

	cmib := make([]C.int, len(mib))
	for i, v := range mib {
		cmib[i] = C.int(v)
	}

	var buf unsafe.Pointer
	var bufsz C.size_t

	rc := C.sysctl_raw((*C.int)(&cmib[0]), C.uint(len(cmib)), &buf, &bufsz)
	if rc != 0 {
		return nil, fmt.Errorf("sysctl failed: errno %d", int(rc))
	}

	if buf == nil || bufsz == 0 {
		return nil, nil
	}
	defer C.free(buf)

	result := make([]byte, int(bufsz))
	copy(result, unsafe.Slice((*byte)(buf), int(bufsz)))
	return result, nil
}

// SysctlIntByName retrieves an integer sysctl value by name.
func SysctlIntByName(name string) (int, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var val C.int
	rc := C.sysctl_int_byname(cname, &val)
	if rc != 0 {
		return 0, fmt.Errorf("sysctlbyname(%s) failed: errno %d", name, int(rc))
	}
	return int(val), nil
}
