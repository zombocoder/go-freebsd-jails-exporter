//go:build freebsd && cgo

package freebsd

/*
#include <sys/param.h>
#include <sys/sysctl.h>
#include <sys/user.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

// get_all_procs retrieves all processes via sysctl KERN_PROC_ALL.
// Returns the number of processes, or -1 on error.
// The caller must free the returned buffer.
static int get_all_procs(struct kinfo_proc **procs) {
	int mib[3] = {CTL_KERN, KERN_PROC, KERN_PROC_ALL};
	size_t len = 0;

	*procs = NULL;

	// First call to get the size.
	if (sysctl(mib, 3, NULL, &len, NULL, 0) < 0) {
		return -1;
	}

	// Add 10% buffer for race conditions.
	len += len / 10;

	*procs = (struct kinfo_proc *)malloc(len);
	if (*procs == NULL) {
		return -1;
	}

	if (sysctl(mib, 3, *procs, &len, NULL, 0) < 0) {
		free(*procs);
		*procs = NULL;
		return -1;
	}

	return (int)(len / sizeof(struct kinfo_proc));
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// ProcState represents a process state category.
type ProcState int

const (
	ProcStateRunning ProcState = iota
	ProcStateSleeping
	ProcStateStopped
	ProcStateZombie
	ProcStateUnknown
)

func (s ProcState) String() string {
	switch s {
	case ProcStateRunning:
		return "running"
	case ProcStateSleeping:
		return "sleeping"
	case ProcStateStopped:
		return "stopped"
	case ProcStateZombie:
		return "zombie"
	default:
		return "unknown"
	}
}

// JailProcStats holds process statistics for a single jail.
type JailProcStats struct {
	Processes int
	Threads   int
	ByState   map[ProcState]int
}

// GetAllJailProcStats returns process stats for all jails, keyed by JID.
func GetAllJailProcStats() (map[int]*JailProcStats, error) {
	var cprocs *C.struct_kinfo_proc
	nprocs := C.get_all_procs(&cprocs)
	if nprocs < 0 {
		return nil, errors.New("failed to get process list via sysctl")
	}
	if cprocs == nil {
		return make(map[int]*JailProcStats), nil
	}
	defer C.free(unsafe.Pointer(cprocs))

	procs := unsafe.Slice(cprocs, int(nprocs))
	stats := make(map[int]*JailProcStats)

	for i := 0; i < int(nprocs); i++ {
		p := &procs[i]
		jid := int(p.ki_jid)
		if jid <= 0 {
			continue
		}

		s, ok := stats[jid]
		if !ok {
			s = &JailProcStats{ByState: make(map[ProcState]int)}
			stats[jid] = s
		}

		s.Processes++
		s.Threads += int(p.ki_numthreads)

		state := mapProcState(int(p.ki_stat))
		s.ByState[state]++
	}

	return stats, nil
}

func mapProcState(stat int) ProcState {
	// From sys/proc.h: SIDL=1, SRUN=2, SSLEEP=3, SSTOP=4, SZOMB=5, SWAIT=6, SLOCK=7
	switch stat {
	case 2: // SRUN
		return ProcStateRunning
	case 3, 6, 7: // SSLEEP, SWAIT, SLOCK
		return ProcStateSleeping
	case 4: // SSTOP
		return ProcStateStopped
	case 5: // SZOMB
		return ProcStateZombie
	default:
		return ProcStateUnknown
	}
}
