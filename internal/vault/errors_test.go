package vault

import (
	"errors"
	"io/fs"
	"testing"
)

func TestAgentMessage(t *testing.T) {
	tests := []struct {
		err error
		want string
	}{
		{errInvalid, "invalid path or identifier"},
		{errOutsideVault, "path is outside the vault"},
		{errNotPermitted, "path is not in the permitted set"},
		{errNotFound, "not found"},
		{errTooLarge, "file is too large to read"},
		{errors.New("some internal error"), "request failed"},
		{nil, "request failed"},
	}
	for _, tt := range tests {
		got := AgentMessage(tt.err)
		if got != tt.want {
			t.Errorf("AgentMessage(%v) = %q, want %q", tt.err, got, tt.want)
		}
	}
}

func TestMapFSError(t *testing.T) {
	tests := []struct{
		name string
		err error
		want error
	}{
		{"nil", nil, nil},
		{"errInvalid passthrough", errInvalid, errInvalid},
		{"errOutsideVault passthrough", errOutsideVault, errOutsideVault},
		{"errNotPermitted passthrough", errNotPermitted, errNotPermitted},
		{"errNotFound passthrough", errNotFound, errNotFound},
		{"errTooLarge  passthrough", errTooLarge , errTooLarge},
		{"fs.ErrNotExist -> errNotFound", fs.ErrNotExist, errNotFound},
		{"fs.ErrPermission -> errNotFound", fs.ErrPermission, errNotFound},
		{"unknown error -> errOutsideVault", errors.New("unknown"), errOutsideVault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapFSError(tt.err)
			if !errors.Is(got, tt.want) && (got != nil || tt.want != nil) {
				t.Errorf("mapFSError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
