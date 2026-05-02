//go:build freebsd && cgo

package freebsd

/*
#include <sys/param.h>
#include <sys/jail.h>
#include <jail.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#cgo LDFLAGS: -ljail

struct jail_extra {
	char parent[32];
	char persist[32];
	char osrelease[256];
	char osreldate[32];
	int  ok;
};

// get_jail_extra retrieves extra jail parameters using jailparam_import/export.
static void get_jail_extra(int jid, struct jail_extra *extra) {
	struct jailparam params[5];
	int nparams = 0;
	char jid_str[32];

	memset(extra, 0, sizeof(*extra));
	snprintf(jid_str, sizeof(jid_str), "%d", jid);

	// Initialize all params one by one, tracking how many succeeded.
	if (jailparam_init(&params[0], "jid") != 0) goto done;
	nparams = 1;
	if (jailparam_import(&params[0], jid_str) != 0) goto done;

	if (jailparam_init(&params[1], "parent") != 0) goto done;
	nparams = 2;

	if (jailparam_init(&params[2], "persist") != 0) goto done;
	nparams = 3;

	if (jailparam_init(&params[3], "osrelease") != 0) goto done;
	nparams = 4;

	if (jailparam_init(&params[4], "osreldate") != 0) goto done;
	nparams = 5;

	if (jailparam_get(params, 5, 0) < 0) goto done;

	// Export values as strings.
	for (int i = 1; i < 5; i++) {
		char *val = jailparam_export(&params[i]);
		if (val == NULL) continue;
		switch (i) {
		case 1: strlcpy(extra->parent, val, sizeof(extra->parent)); break;
		case 2: strlcpy(extra->persist, val, sizeof(extra->persist)); break;
		case 3: strlcpy(extra->osrelease, val, sizeof(extra->osrelease)); break;
		case 4: strlcpy(extra->osreldate, val, sizeof(extra->osreldate)); break;
		}
		free(val);
	}
	extra->ok = 1;

done:
	if (nparams > 0)
		jailparam_free(params, nparams);
}
*/
import "C"

import (
	"fmt"
	"strconv"
)

// JailExtra contains additional jail parameters not available in the library's Info struct.
type JailExtra struct {
	Parent    int
	Persist   bool
	OSRelease string
	OSRelDate int
}

// GetJailExtra retrieves extra jail parameters (parent, persist, osrelease, osreldate) for a jail by JID.
func GetJailExtra(jid int) (*JailExtra, error) {
	var extra C.struct_jail_extra
	C.get_jail_extra(C.int(jid), &extra)
	if extra.ok == 0 {
		return nil, fmt.Errorf("failed to get extra params for jail %d", jid)
	}

	parent, _ := strconv.Atoi(C.GoString(&extra.parent[0]))
	persist := C.GoString(&extra.persist[0])
	osreldate, _ := strconv.Atoi(C.GoString(&extra.osreldate[0]))

	return &JailExtra{
		Parent:    parent,
		Persist:   persist == "1",
		OSRelease: C.GoString(&extra.osrelease[0]),
		OSRelDate: osreldate,
	}, nil
}
