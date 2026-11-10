package signal

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

const (
	ethAsset  = "0x1111111111111111111111111111111111111111"
	ethMaker  = "0x2222222222222222222222222222222222222222"
	ethPay    = "0x3333333333333333333333333333333333333333"
	solAsset  = "So11111111111111111111111111111111111111112"
	solToken  = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	solMaker  = "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"
	nonce     = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	termsHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func validETH() Intent {
	return Intent{
		V:       Version,
		Chain:   ChainETH,
		Asset:   ethAsset,
		Token:   "0",
		Side:    SideSell,
		Price:   "1000000000000000000",
		Pay:     ethPay,
		Expiry:  "time:2000000000",
		Maker:   ethMaker,
		Nonce:   nonce,
		Adapter: "seaport/1.6.0",
		Ref:     termsHash,
	}
}

func validSOL() Intent {
	return Intent{
		V:       Version,
		Chain:   ChainSOL,
		Asset:   solAsset,
		Token:   solToken,
		Side:    SideBuy,
		Price:   "2500000",
		Pay:     "native",
		Expiry:  "height:250000000",
		Maker:   solMaker,
		Nonce:   "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		Adapter: "magiceden/2.0.0",
		Ref:     "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
	}
}

func TestValidRecords(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   Intent
	}{
		{"ethereum sell", validETH()},
		{"solana buy", validSOL()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.in.Validate(); err != nil {
				t.Fatalf("Validate: %v", err)
			}
			raw, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Parse(raw)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got != tc.in {
				t.Fatalf("Parse changed the record\n got: %+v\nwant: %+v", got, tc.in)
			}
		})
	}
}

func TestCanonical(t *testing.T) {
	in := validETH()
	got, err := in.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	want := `{"v":1,"chain":"eth","asset":"` + ethAsset + `","token":"0","side":"sell","price":"1000000000000000000","pay":"` + ethPay + `","expiry":"time:2000000000","maker":"` + ethMaker + `","nonce":"` + nonce + `","adapter":"seaport/1.6.0","ref":"` + termsHash + `"}`
	if string(got) != want {
		t.Fatalf("Canonical:\n got: %s\nwant: %s", got, want)
	}

	// Formatting and key order in the input do not change the canonical bytes.
	shuffled := `{
		"ref":"` + termsHash + `",
		"adapter":"seaport/1.6.0",
		"nonce":"` + nonce + `",
		"maker":"` + ethMaker + `",
		"expiry":"time:2000000000",
		"pay":"` + ethPay + `",
		"price":"1000000000000000000",
		"side":"sell",
		"token":"0",
		"asset":"` + ethAsset + `",
		"chain":"eth",
		"v":1
	}`
	parsed, err := Parse([]byte(shuffled))
	if err != nil {
		t.Fatal(err)
	}
	again, err := parsed.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != want {
		t.Fatalf("Canonical after reordered input:\n got: %s\nwant: %s", again, want)
	}
}

func TestValidateRejects(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Intent)
		want string
	}{
		{"wrong version", func(in *Intent) { in.V = 2 }, "v:"},
		{"zero version", func(in *Intent) { in.V = 0 }, "v:"},
		{"unknown chain", func(in *Intent) { in.Chain = "base" }, "chain:"},
		{"empty asset", func(in *Intent) { in.Asset = "" }, "asset:"},
		{"uppercase eth asset", func(in *Intent) { in.Asset = "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" }, "asset:"},
		{"checksummed eth asset", func(in *Intent) { in.Asset = "0xAbCdEf0123456789aBcDeF0123456789AbCdEf01" }, "asset:"},
		{"zero eth asset", func(in *Intent) { in.Asset = "0x0000000000000000000000000000000000000000" }, "asset:"},
		{"sol asset on eth", func(in *Intent) { in.Asset = solAsset }, "asset:"},
		{"eth leading-zero token", func(in *Intent) { in.Token = "01" }, "token:"},
		{"eth negative token", func(in *Intent) { in.Token = "-1" }, "token:"},
		{"eth token overflow", func(in *Intent) { in.Token = strings.Repeat("9", 78) }, "token:"},
		{"bad side", func(in *Intent) { in.Side = "trade" }, "side:"},
		{"zero price", func(in *Intent) { in.Price = "0" }, "price:"},
		{"leading-zero price", func(in *Intent) { in.Price = "01" }, "price:"},
		{"signed price", func(in *Intent) { in.Price = "+1" }, "price:"},
		{"price overflow", func(in *Intent) { in.Price = strings.Repeat("9", 78) }, "price:"},
		{"symbol pay", func(in *Intent) { in.Pay = "USDC" }, "pay:"},
		{"wrong-chain pay", func(in *Intent) { in.Pay = solAsset }, "pay:"},
		{"empty expiry", func(in *Intent) { in.Expiry = "" }, "expiry:"},
		{"zero time expiry", func(in *Intent) { in.Expiry = "time:0" }, "expiry:"},
		{"negative height expiry", func(in *Intent) { in.Expiry = "height:-1" }, "expiry:"},
		{"unknown expiry kind", func(in *Intent) { in.Expiry = "block:100" }, "expiry:"},
		{"zero maker", func(in *Intent) { in.Maker = "0x0000000000000000000000000000000000000000" }, "maker:"},
		{"short nonce", func(in *Intent) { in.Nonce = "aa" }, "nonce:"},
		{"uppercase nonce", func(in *Intent) { in.Nonce = strings.ToUpper(nonce) }, "nonce:"},
		{"zero nonce", func(in *Intent) { in.Nonce = strings.Repeat("0", 64) }, "nonce:"},
		{"adapter without version", func(in *Intent) { in.Adapter = "seaport" }, "adapter:"},
		{"adapter with partial version", func(in *Intent) { in.Adapter = "seaport/1.6" }, "adapter:"},
		{"uppercase adapter", func(in *Intent) { in.Adapter = "Seaport/1.6.0" }, "adapter:"},
		{"short ref", func(in *Intent) { in.Ref = "bb" }, "ref:"},
		{"uppercase ref", func(in *Intent) { in.Ref = strings.ToUpper(termsHash) }, "ref:"},
		{"zero ref", func(in *Intent) { in.Ref = strings.Repeat("0", 64) }, "ref:"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := validETH()
			tc.edit(&in)
			err := in.Validate()
			if err == nil {
				t.Fatal("Validate accepted an invalid record")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestParseRejectsAmbiguousJSON(t *testing.T) {
	raw, err := json.Marshal(validETH())
	if err != nil {
		t.Fatal(err)
	}
	base := string(raw)
	tests := []struct {
		name string
		data string
		want string
	}{
		{"unknown field", strings.TrimSuffix(base, "}") + `,"extra":1}`, "unknown field"},
		{"duplicate key", strings.Replace(base, `"v":1`, `"v":1,"v":1`, 1), "duplicate key"},
		{"trailing object", base + `{}`, "trailing data"},
		{"array", `[` + base + `]`, "cannot unmarshal"},
		{"empty", ``, "invalid JSON"},
		{"trailing comma", strings.TrimSuffix(base, "}") + `,}`, "invalid object"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.data))
			if err == nil {
				t.Fatal("Parse accepted invalid JSON")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Parse error = %q, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestExpiry(t *testing.T) {
	tests := []struct {
		in   string
		kind ExpiryKind
		val  uint64
		ok   bool
	}{
		{"time:2000000000", ExpiryTime, 2000000000, true},
		{"height:250000000", ExpiryHeight, 250000000, true},
		{"time:0", "", 0, false},
		{"height:0", "", 0, false},
		{"height:-1", "", 0, false},
		{"height:1.5", "", 0, false},
		{"time:18446744073709551615", "", 0, false}, // beyond Unix time
		{"date:2000000000", "", 0, false},
		{"time", "", 0, false},
		{"time:", "", 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseExpiry(tc.in)
			if tc.ok {
				if err != nil {
					t.Fatal(err)
				}
				if got.Kind != tc.kind || got.Value != tc.val {
					t.Fatalf("got %+v, want %s:%d", got, tc.kind, tc.val)
				}
				return
			}
			if err == nil {
				t.Fatalf("ParseExpiry accepted %q", tc.in)
			}
		})
	}
}

func TestLive(t *testing.T) {
	before := time.Unix(1999999999, 0)
	at := time.Unix(2000000000, 0)

	in := validETH()
	if err := in.Live(before, 0); err != nil {
		t.Fatalf("live before expiry: %v", err)
	}
	if err := in.Live(at, 0); err == nil {
		t.Fatal("Live accepted a record at its time expiry")
	}

	in = validSOL() // height:250000000
	if err := in.Live(time.Time{}, 249999999); err != nil {
		t.Fatalf("live below height expiry: %v", err)
	}
	if err := in.Live(time.Time{}, 250000000); err == nil {
		t.Fatal("Live accepted a record at its height expiry")
	}
	if err := in.Live(time.Time{}, 0); err != nil {
		t.Fatalf("Live rejected a height-expiring record with an unknown height: %v", err)
	}
}

func TestTokenRules(t *testing.T) {
	in := validETH()
	if err := in.Validate(); err != nil {
		t.Fatalf("token 0 must be legal: %v", err)
	}
	in.Token = "115792089237316195423570985008687907853269984665640564039457584007913129639935" // 2^256-1
	if err := in.Validate(); err != nil {
		t.Fatalf("max uint256 token must be legal: %v", err)
	}

	sol := validSOL()
	sol.Token = "not-a-mint"
	if err := sol.Validate(); err == nil {
		t.Fatal("Solana token accepted a non-address")
	}
}
