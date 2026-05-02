//go:build !freebsd || !cgo

package freebsd

func SysctlRaw(_ []int32) ([]byte, error) {
	return nil, errUnsupported
}

func SysctlIntByName(_ string) (int, error) {
	return 0, errUnsupported
}
