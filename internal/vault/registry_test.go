package vault

import (
	"errors"
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()

	// Need a real vault for testing - use temp dir
	root := t.TempDir()
	v, err := Open(Config{Name: "test-vault", Root: root})
	if err != nil {
		t.Fatalf("Open %v", err)
	}

	// Register
	if err := r.Register(v, false); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Get
	got, err := r.Get("test-vault")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != v {
		t.Error("Get returned different vault")
	}

	// First registered becomes default
	def, err := r.Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if def != v {
		t.Error("Default should be first registered vault")
	}
	if r.DefaultName() != "test-vault" {
		t.Errorf("DefaultName = %q, want %q", r.DefaultName(), "test-vault")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()
	root := t.TempDir()

	v1, _ := Open(Config{Name: "dup", Root: root})
	v2, _ := Open(Config{Name: "dup", Root: root})

	if err := r.Register(v1, false); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	if err := r.Register(v2, false); !errors.Is(err, ErrVaultExists) {
		t.Fatalf("second Register error %v, want ErrVaultExists", err)
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("Get error = %v, want ErrVaultNotFound", err)
	}
}

func TestRegistry_DefaultEmpty(t *testing.T) {
	r := NewRegistry()

	_, err := r.Default()
	if !errors.Is(err, ErrNoDefaultVault) {
		t.Errorf("Default error = %v, want ErrNoDefaultVault", err)
	}
	if r.DefaultName() != "" {
		t.Errorf("DefaultName = %q, want empty", r.DefaultName())
	}
}

func TestRegistry_ExplicitDefault(t *testing.T) {
	r := NewRegistry()
	root := t.TempDir()

	v1, _ := Open(Config{Name: "first", Root: root})
	v2, _ := Open(Config{Name: "second", Root: root})
	r.Register(v1, false)
	r.Register(v2, true) // explicit default
	if r.DefaultName() != "second" {
		t.Errorf("DefaultName = %q, want %q", r.DefaultName(), "second")
	}
}
