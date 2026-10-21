//go:build linux

package state

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

// Secret Service calls intentionally never invoke Unlock or Prompt. Locked
// keyrings fail closed, including in noninteractive scheduled tasks.
func chooseProtection(mode string) string {
	if mode == "file" {
		return "file"
	}
	if mode == "auto" && os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		return "file"
	}
	return "keyring"
}
func keyringKey(ctx context.Context, account string, create, interactive bool) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	address := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	if address == "" {
		return nil, ErrVault
	}
	for _, entry := range strings.Split(address, ";") {
		if !strings.HasPrefix(entry, "unix:") {
			return nil, ErrVault
		}
	}
	// Never auto-launch a bus or connect a plaintext Secret Service session over
	// a remote transport. The existing local bus must already be available.
	conn, err := dbus.Connect(address, dbus.WithContext(ctx))
	if err != nil {
		return nil, ErrVault
	}
	defer conn.Close()
	const name = "org.freedesktop.secrets"
	const serviceIF = "org.freedesktop.Secret.Service"
	const collectionIF = "org.freedesktop.Secret.Collection"
	service := conn.Object(name, "/org/freedesktop/secrets")
	attributes := map[string]string{"service": "dev.flashdata.cli", "account": account}
	var unlocked, locked []dbus.ObjectPath
	if err = service.CallWithContext(ctx, serviceIF+".SearchItems", 0, attributes).Store(&unlocked, &locked); err != nil || len(locked) > 0 {
		return nil, ErrVault
	}
	var variant dbus.Variant
	var session dbus.ObjectPath
	if err = service.CallWithContext(ctx, serviceIF+".OpenSession", 0, "plain", dbus.MakeVariant("")).Store(&variant, &session); err != nil {
		return nil, ErrVault
	}
	defer conn.Object(name, session).CallWithContext(ctx, "org.freedesktop.Secret.Session.Close", 0)
	type secret struct {
		Session     dbus.ObjectPath
		Parameters  []byte
		Value       []byte
		ContentType string
	}
	if len(unlocked) > 0 {
		var result map[dbus.ObjectPath]secret
		if err = service.CallWithContext(ctx, serviceIF+".GetSecrets", 0, unlocked, session).Store(&result); err != nil {
			return nil, ErrVault
		}
		v, ok := result[unlocked[0]]
		if !ok {
			return nil, ErrVault
		}
		return DecodeKey(string(v.Value))
	}
	if !create {
		return nil, ErrVault
	}
	collection := conn.Object(name, "/org/freedesktop/secrets/aliases/default")
	if err = collection.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, collectionIF, "Locked").Store(&variant); err != nil {
		return nil, ErrVault
	}
	if isLocked, ok := variant.Value().(bool); !ok || isLocked {
		return nil, ErrVault
	}
	key, err := RandomKey()
	if err != nil {
		return nil, err
	}
	properties := map[string]dbus.Variant{"org.freedesktop.Secret.Item.Label": dbus.MakeVariant("FlashData CLI credential encryption"), "org.freedesktop.Secret.Item.Attributes": dbus.MakeVariant(attributes)}
	var item, prompt dbus.ObjectPath
	if err = collection.CallWithContext(ctx, collectionIF+".CreateItem", 0, properties, secret{session, []byte{}, []byte(EncodeKey(key)), "text/plain"}, true).Store(&item, &prompt); err != nil || prompt != "/" {
		return nil, errors.New("credential_store_locked_or_unavailable")
	}
	return key, nil
}
