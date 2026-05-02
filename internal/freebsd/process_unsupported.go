//go:build !freebsd || !cgo

package freebsd

type ProcState int

const (
	ProcStateRunning  ProcState = iota
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

type JailProcStats struct {
	Processes int
	Threads   int
	ByState   map[ProcState]int
}

func GetAllJailProcStats() (map[int]*JailProcStats, error) {
	return nil, errUnsupported
}
