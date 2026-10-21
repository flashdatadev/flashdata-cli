//go:build linux

package state

import (
	"context"
	"errors"
	"testing"
)

func TestSecretServiceNeverAutolaunchesOrUsesRemoteBus(t *testing.T) {
	for _, address := range []string{"", "autolaunch:", "tcp:host=example.com,port=1234", "unix:path=/tmp/flashdata-test-bus;autolaunch:"} {
		t.Setenv("DBUS_SESSION_BUS_ADDRESS", address)
		if _, err := keyringKey(context.Background(), "offline", true, false); !errors.Is(err, ErrVault) {
			t.Fatal("unexpected bus accepted", address, err)
		}
	}
}
