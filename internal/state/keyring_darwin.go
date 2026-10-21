//go:build darwin

package state

import (
	"context"
	"github.com/ebitengine/purego"
	"sync"
	"unsafe"
)

var keychainOnce sync.Once
var keychainError error
var keychainInteraction func(uint8) int32
var keychainFind func(uintptr, uint32, *byte, uint32, *byte, *uint32, *unsafe.Pointer, uintptr) int32
var keychainAdd func(uintptr, uint32, *byte, uint32, *byte, uint32, *byte, uintptr) int32
var keychainFree func(uintptr, unsafe.Pointer) int32

func loadKeychain() {
	lib, err := purego.Dlopen("/System/Library/Frameworks/Security.framework/Security", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		keychainError = err
		return
	}
	purego.RegisterLibFunc(&keychainInteraction, lib, "SecKeychainSetUserInteractionAllowed")
	purego.RegisterLibFunc(&keychainFind, lib, "SecKeychainFindGenericPassword")
	purego.RegisterLibFunc(&keychainAdd, lib, "SecKeychainAddGenericPassword")
	purego.RegisterLibFunc(&keychainFree, lib, "SecKeychainItemFreeContent")
	if keychainInteraction(0) != 0 {
		keychainError = ErrVault
	}
}
func chooseProtection(mode string) string {
	if mode == "file" {
		return "file"
	}
	return "keyring"
}
func keyringKey(ctx context.Context, account string, create, interactive bool) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	keychainOnce.Do(loadKeychain)
	if keychainError != nil {
		return nil, ErrVault
	}
	service, acct := []byte("dev.flashdata.cli"), []byte(account)
	var size uint32
	var ptr unsafe.Pointer
	status := keychainFind(0, uint32(len(service)), &service[0], uint32(len(acct)), &acct[0], &size, &ptr, 0)
	if status == 0 {
		defer keychainFree(0, ptr)
		if size != 32 || ptr == nil {
			return nil, ErrVault
		}
		return append([]byte(nil), unsafe.Slice((*byte)(ptr), int(size))...), nil
	}
	// Only errSecItemNotFound permits creation. Native UI is always disabled.
	if status != -25300 || !create {
		return nil, ErrVault
	}
	key, err := RandomKey()
	if err != nil {
		return nil, err
	}
	if keychainAdd(0, uint32(len(service)), &service[0], uint32(len(acct)), &acct[0], uint32(len(key)), &key[0], 0) != 0 {
		return nil, ErrVault
	}
	return key, nil
}
