//go:build windows

package app

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

func legacyProtected(b []byte) (string, []byte, error) {
	in := windows.DataBlob{Size: uint32(len(b)), Data: &b[0]}
	var out windows.DataBlob
	err := windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	if err != nil {
		return "", nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return "dpapi", append([]byte(nil), unsafe.Slice(out.Data, int(out.Size))...), nil
}
