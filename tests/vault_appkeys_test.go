package tests

import (
	"testing"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/veil/tests/harness"
)

// WT-17, DVM half: the app can only report a Register or Burn as executed when it reads back the
// exact keys the contract wrote. vault.bas v3 keys uc_, ul_ and w_<j>_by by the RAW 33-byte SIGNER()
// value; the app used to ask for the dero1 text instead and therefore never saw its own index, deposit
// address or withdrawal. This test performs the two calls and reads each record with the keys
// app/vault.js builds, so a contract change back to a text key (or an app change back to one) fails here
// rather than silently disabling Register and Burn.
func TestVaultAppKeys(t *testing.T) {
	c := harness.New()
	owner, treasury, user := c.NewAccount(), c.NewAccount(), c.NewAccount()
	cfg := vaultCfg{k: 2, cap: 9000, minbond: 1000, minburn: 100, wfee: 10, mindep: 5, regfee: 0, signers: 3}
	scid, sg := installVault(t, c, owner, treasury, cfg)
	if r := c.Call(scid, "Seal", nil, nil, owner); !r.OK() {
		t.Fatalf("seal: %v", r.Err)
	}
	if r := c.Call(scid, "Register", nil, nil, user); !r.OK() {
		t.Fatalf("register: %v", r.Err)
	}

	// Register: ua_<j> holds the raw key, ul_<raw>_<c> names the index, uc_<raw> is the count.
	raw := string(user.Compressed())
	if got := c.S(scid, "ua_0"); got != raw {
		t.Fatalf("ua_0 is %x, want the raw SIGNER() %x", got, raw)
	}
	if got := c.U(scid, "ul_"+raw+"_0"); got != 0 {
		t.Fatalf("ul_<raw>_0 = %d, want index 0", got)
	}
	if got := c.U(scid, "uc_"+raw); got != 1 {
		t.Fatalf("uc_<raw> = %d, want 1", got)
	}
	// The text spelling is a different key: asking for it is exactly the bug WT-17 fixes.
	if c.Exists(scid, "uc_"+user.String()) {
		t.Fatalf("uc_<dero1> exists; vault.bas is expected to key per-address records by the raw SIGNER()")
	}

	// Attest and mint so the user has units to burn.
	for i := 0; i < 2; i++ {
		if r := c.Call(scid, "Attest", rpc.Arguments{harness.U("j", 0), harness.S("addr", "4Axmr-user")}, nil, sg[i]); !r.OK() {
			t.Fatalf("attest %d: %v", i, r.Err)
		}
	}
	for i := 0; i < 2; i++ {
		if r := c.Call(scid, "Mint", mintArgs(0, 1, 0, 2000), nil, sg[i]); !r.OK() {
			t.Fatalf("mint %d: %v", i, r.Err)
		}
	}
	if c.U(scid, "sup") != 2000 {
		t.Fatalf("sup %d, want 2000", c.U(scid, "sup"))
	}

	// Burn: the record the Burn probe reads is w_<j>_to, w_<j>_v and w_<j>_by, the last keyed raw.
	dest := "4AxmrDestinationAddressForTheAppKeysTest"
	if r := c.Call(scid, "Burn", rpc.Arguments{harness.S("dest", dest), harness.S("refund", user.String())}, map[crypto.Hash]uint64{scid: 2000}, user); !r.OK() {
		t.Fatalf("burn: %v", r.Err)
	}
	if got := c.S(scid, "w_0_to"); got != dest {
		t.Fatalf("w_0_to = %q, want %q", got, dest)
	}
	if got := c.U(scid, "w_0_v"); got != 1990 {
		t.Fatalf("w_0_v = %d, want 1990 (2000 units - 10 wfee)", got)
	}
	if got := c.S(scid, "w_0_by"); got != raw {
		t.Fatalf("w_0_by is %x, want the raw refund key %x", got, raw)
	}
	// This is the exact check app/vault.js's burnOpened makes: the record opened in [0,1), names this
	// wallet and the signed destination, and holds the net units. A wrong net amount must not match.
	if !(c.S(scid, "w_0_to") == dest && c.S(scid, "w_0_by") == raw && c.U(scid, "w_0_v") == 1990) {
		t.Fatal("the Burn probe's record check does not match the contract's own record")
	}
}
