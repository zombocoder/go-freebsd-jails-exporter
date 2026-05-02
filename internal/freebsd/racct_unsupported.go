//go:build !freebsd || !cgo

package freebsd

type RACCTResource struct {
	Name  string
	Value int64
}

type JailRACCT struct {
	Resources map[string]int64
}

func IsRACCTEnabled() bool {
	return false
}

func GetJailRACCT(_ string) (*JailRACCT, error) {
	return nil, errUnsupported
}
