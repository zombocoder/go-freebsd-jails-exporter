//go:build !freebsd || !cgo

package freebsd

type JailExtra struct {
	Parent    int
	Persist   bool
	OSRelease string
	OSRelDate int
}

func GetJailExtra(_ int) (*JailExtra, error) {
	return nil, errUnsupported
}
