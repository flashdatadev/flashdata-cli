//go:build windows

package state

import (
	"context"
	"golang.org/x/sys/windows"
	"os"
	"time"
	"unsafe"
)

func currentSID() (*windows.SID, error) {
	u, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, err
	}
	return u.User.Sid, nil
}
func privateCheck(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return ErrUnsafe
	}
	sid, err := currentSID()
	if err != nil {
		return ErrUnsafe
	}
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return ErrUnsafe
	}
	owner, _, err := sd.Owner()
	if err != nil || (!owner.Equals(sid) && owner.String() != "S-1-5-32-544") {
		return ErrUnsafe
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil {
		return ErrUnsafe
	}
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if windows.GetAce(acl, i, &ace) != nil {
			return ErrUnsafe
		}
		if ace.Header.AceType == windows.ACCESS_DENIED_ACE_TYPE {
			continue
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return ErrUnsafe
		}
		allowed := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !allowed.Equals(sid) && allowed.String() != "S-1-5-18" && allowed.String() != "S-1-5-32-544" {
			return ErrUnsafe
		}
	}
	return nil
}
func setPrivate(path string) error {
	sid, err := currentSID()
	if err != nil {
		return ErrUnsafe
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;" + sid.String() + ")")
	if err != nil {
		return ErrUnsafe
	}
	acl, _, err := sd.DACL()
	if err != nil {
		return ErrUnsafe
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}
func privateDirectory(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if os.MkdirAll(path, 0700) != nil {
			return ErrUnsafe
		}
		if setPrivate(path) != nil {
			return ErrUnsafe
		}
		info, err = os.Lstat(path)
	}
	if err != nil || !info.IsDir() {
		return ErrUnsafe
	}
	return privateCheck(path, info)
}
func secureCreated(path string) error { return setPrivate(path) }
func createOutput(path string) (*os.File, error) {
	sid, err := currentSID()
	if err != nil {
		return nil, ErrUnsafe
	}
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;FA;;;SY)(A;;FA;;;" + sid.String() + ")")
	if err != nil {
		return nil, ErrUnsafe
	}
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	attributes := windows.SecurityAttributes{SecurityDescriptor: sd}
	attributes.Length = uint32(unsafe.Sizeof(attributes))
	handle, err := windows.CreateFile(name, windows.GENERIC_WRITE, 0, &attributes, windows.CREATE_NEW, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(handle), path), nil
}
func replace(from, to string) error {
	a, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	b, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(a, b, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
func syncDirectory(string) error { return nil }
func lock(ctx context.Context, path string) (func(), error) {
	if info, e := os.Lstat(path); e == nil {
		if !info.Mode().IsRegular() || privateCheck(path, info) != nil {
			return nil, ErrUnsafe
		}
	} else if !os.IsNotExist(e) {
		return nil, ErrUnsafe
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, ErrUnsafe
	}
	if err = secureCreated(path); err != nil {
		f.Close()
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || privateCheck(path, info) != nil {
		f.Close()
		return nil, ErrUnsafe
	}
	ov := &windows.Overlapped{}
	for {
		err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, ov)
		if err == nil {
			return func() { _ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, ov); f.Close() }, nil
		}
		if err != windows.ERROR_LOCK_VIOLATION {
			f.Close()
			return nil, ErrUnsafe
		}
		select {
		case <-ctx.Done():
			f.Close()
			return nil, ErrBusy
		case <-time.After(25 * time.Millisecond):
		}
	}
}
func protectOS(b []byte, decrypt bool) ([]byte, error) {
	if len(b) == 0 {
		return nil, ErrCorrupt
	}
	in := windows.DataBlob{Size: uint32(len(b)), Data: &b[0]}
	var out windows.DataBlob
	var err error
	if decrypt {
		err = windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	} else {
		err = windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	}
	if err != nil {
		return nil, ErrVault
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	result := append([]byte(nil), unsafe.Slice(out.Data, int(out.Size))...)
	return result, nil
}
func chooseProtection(mode string) string {
	if mode == "file" {
		return "file"
	}
	return "dpapi"
}
func keyringKey(context.Context, string, bool, bool) ([]byte, error) { return nil, ErrVault }
