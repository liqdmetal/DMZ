package main

import (
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"
	"testing"

	rpc "veil/tools/simrpc"
)

const (
	poolHashCur   = "dced10e0aa5a23adb91fdc01251152ae2ca35cfcac13fa01209a1091c88127dc"
	poolHashOld   = "304ca766363a493edea83f19102091291275eb0ba2573f27cc0680c5a5041580"
	launchHashCur = "ac9dbb1e9cd613cc8022cc958f4d7e17285fc853a6c2598664a5efd53c0771f4"
	// ourTreasury is already in the spelling ADDRESS_STRING writes, which is what MarketTreasury reports and what
	// DVMAddress returns unchanged.
	ourTreasury   = "dero1qy2nxgts7wdn28ckc4l2tewphjcppqjfj69ddkxjn0ay8hlsjx73jqgpyhv40"
	otherTreasury = "dero1qyx7qyvrtrhvaszeej487k2g689fav7h38ay37fja9qf40ycgl0m2qgayajpd"
)

// The sets come from the block named, current and legacy, and never from another template's block.
func TestMarketTemplatesFromConfig(t *testing.T) {
	src := `templates: {
    pool:     { file: "templates/pool.bas", sha256: "` + poolHashCur + `", legacy: ["` + poolHashOld + `"] },
    launch:   { file: "templates/launch.bas", sha256: "` + launchHashCur + `", legacy: [] },
    nfa:      { file: "templates/nfa.bas", sha256: "` + strings.Repeat("9", 64) + `", legacy: [] }
  }`
	pool, launch := marketTemplates(src)
	if len(pool) != 2 || !pool[poolHashCur] || !pool[poolHashOld] {
		t.Fatalf("pool hashes = %v", pool)
	}
	if len(launch) != 1 || !launch[launchHashCur] {
		t.Fatalf("launch hashes = %v", launch)
	}
	if pool[launchHashCur] || launch[poolHashCur] {
		t.Fatal("a hash was taken from another template's block")
	}
	if p, l := marketTemplates(`founders: { sha256: "` + strings.Repeat("1", 64) + `" }`); len(p)+len(l) != 0 {
		t.Fatalf("a config naming neither template verified something: %v %v", p, l)
	}
}

// AUD-15-02/-10: a pool registered as a launch, a contract that is not the shipped template, and a market paying
// another treasury are all refused before any fee key is read; the shipped template is taken on either list, and a
// legacy hash of it is still a market of that list.
func TestListedMarketsRefusesWhatTheRegistryCannotVouchFor(t *testing.T) {
	oldPool, oldLaunch := poolTpl, launchTpl
	t.Cleanup(func() { poolTpl, launchTpl = oldPool, oldLaunch })
	poolTpl = map[string]bool{poolHashCur: true, poolHashOld: true}
	launchTpl = map[string]bool{launchHashCur: true}

	enc := func(s string) string { return hex.EncodeToString([]byte(s)) }
	ids := map[string]string{
		"pool_0":   scid('a'), // shipped pool, our treasury
		"pool_1":   scid('b'), // a legacy version of the pool template
		"pool_2":   scid('c'), // a contract anyone installed
		"pool_3":   scid('d'), // shipped pool, but paying somebody else
		"launch_0": scid('e'), // the POOL template registered as a launch
		"launch_1": scid('f'), // shipped launch, our treasury
	}
	keys := map[string]interface{}{"pc": json.Number("4"), "lc": json.Number("2")}
	for k, id := range ids {
		keys[k] = enc(id)
	}
	ch := newFakeChain(1)
	for k, id := range ids {
		ch.m[id] = &fakeMarket{quote: deroID, asset: tokA}
		ch.treasury[id] = ourTreasury
		ch.code[id] = poolHashCur
		_ = k
	}
	ch.code[ids["pool_1"]] = poolHashOld
	ch.code[ids["pool_2"]] = strings.Repeat("0", 64)
	ch.treasury[ids["pool_3"]] = otherTreasury
	ch.code[ids["launch_0"]] = poolHashCur
	ch.code[ids["launch_1"]] = launchHashCur

	got := listedMarkets(rpc.SCState{StringKeys: keys}, ch, ourTreasury)
	var names []string
	for _, m := range got {
		names = append(names, m.scid)
	}
	want := ids["pool_0"] + "," + ids["pool_1"] + "," + ids["launch_1"]
	if strings.Join(names, ",") != want {
		t.Fatalf("swept %v, want pool_0, pool_1 and launch_1 (in that order)", names)
	}

	// A build that ships no template verifies nothing, whatever the registry says.
	poolTpl, launchTpl = nil, nil
	if got := listedMarkets(rpc.SCState{StringKeys: keys}, ch, ourTreasury); len(got) != 0 {
		t.Fatalf("swept %d markets with no template named in the config", len(got))
	}
}

// AUD-15-02: the credit wraps nothing. An addition that would pass 2^64 is refused, and the journal keeps what it had.
func TestOwedUnitsCannotWrap(t *testing.T) {
	j := newJournal(t.TempDir() + "/journal.json")
	j.Owed[tokA] = math.MaxUint64 - 1
	if err := j.add(tokA, 2); err == nil || !strings.Contains(err.Error(), "wrap") {
		t.Fatalf("an addition past 2^64 was accepted: %v", err)
	}
	if j.Owed[tokA] != math.MaxUint64-1 {
		t.Fatalf("the refused addition changed the total: %d", j.Owed[tokA])
	}
	if err := j.add(tokA, 1); err != nil || j.Owed[tokA] != math.MaxUint64 {
		t.Fatalf("the exact fit was refused: %v, %d", err, j.Owed[tokA])
	}

	// The same guard on the claim path: a market cannot claim the journal into a wrapped total.
	ch := newFakeChain(1)
	ch.m["m"] = &fakeMarket{fd: 10, fa: 1, quote: deroID, asset: tokA}
	k := newJournal(t.TempDir() + "/journal2.json")
	k.Owed[deroID] = math.MaxUint64 - 5
	if _, _, err := k.claimMarket(ch, "m", deroID, tokA); err == nil {
		t.Fatal("a claim that would wrap the journal was counted")
	}
	if k.Owed[deroID] != math.MaxUint64-5 {
		t.Fatalf("the refused claim changed the total: %d", k.Owed[deroID])
	}
}
