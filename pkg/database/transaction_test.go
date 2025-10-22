package database

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

// TestDefaultTxOptions tests that DefaultTxOptions returns correct default values
func TestDefaultTxOptions(t *testing.T) {
	opts := DefaultTxOptions()

	// Verify isolation level
	if opts.IsoLevel != pgx.ReadCommitted {
		t.Errorf("expected IsoLevel to be ReadCommitted, got %v", opts.IsoLevel)
	}

	// Verify access mode
	if opts.AccessMode != pgx.ReadWrite {
		t.Errorf("expected AccessMode to be ReadWrite, got %v", opts.AccessMode)
	}

	// Verify read-only flag
	if opts.ReadOnly != false {
		t.Errorf("expected ReadOnly to be false, got %v", opts.ReadOnly)
	}
}

// TestTxOptions_CustomValues tests that we can create custom transaction options
func TestTxOptions_CustomValues(t *testing.T) {
	tests := []struct {
		name       string
		opts       TxOptions
		wantIso    pgx.TxIsoLevel
		wantAccess pgx.TxAccessMode
		wantRO     bool
	}{
		{
			name: "serializable read-write",
			opts: TxOptions{
				IsoLevel:   pgx.Serializable,
				AccessMode: pgx.ReadWrite,
				ReadOnly:   false,
			},
			wantIso:    pgx.Serializable,
			wantAccess: pgx.ReadWrite,
			wantRO:     false,
		},
		{
			name: "read committed read-only",
			opts: TxOptions{
				IsoLevel:   pgx.ReadCommitted,
				AccessMode: pgx.ReadOnly,
				ReadOnly:   true,
			},
			wantIso:    pgx.ReadCommitted,
			wantAccess: pgx.ReadOnly,
			wantRO:     true,
		},
		{
			name: "repeatable read",
			opts: TxOptions{
				IsoLevel:   pgx.RepeatableRead,
				AccessMode: pgx.ReadWrite,
				ReadOnly:   false,
			},
			wantIso:    pgx.RepeatableRead,
			wantAccess: pgx.ReadWrite,
			wantRO:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.IsoLevel != tt.wantIso {
				t.Errorf("IsoLevel = %v, want %v", tt.opts.IsoLevel, tt.wantIso)
			}
			if tt.opts.AccessMode != tt.wantAccess {
				t.Errorf("AccessMode = %v, want %v", tt.opts.AccessMode, tt.wantAccess)
			}
			if tt.opts.ReadOnly != tt.wantRO {
				t.Errorf("ReadOnly = %v, want %v", tt.opts.ReadOnly, tt.wantRO)
			}
		})
	}
}
