package signal

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Randomized differential cases. A fixed seed makes the corpus reproducible: this test generates it,
// compares it with the committed fixture, and runs it through the Go validator. tests/js/signal.random
// runs the same fixture through app/signal.js. Regenerate after changing the generator with:
//
//	SIGNAL_UPDATE=1 go test ./signal -run '^TestRandomVectors$'
const (
	randomVectorsPath = "testdata/signal_v1_random_vectors.json"
	randomSeed        = int64(0x51a1)
	randomRounds      = 180
)

type randomCorpus struct {
	Version int          `json:"version"`
	Seed    int64        `json:"seed"`
	Rounds  int          `json:"rounds"`
	Cases   []randomCase `json:"cases"`
}

type randomCase struct {
	Name      string         `json:"name"`
	Record    map[string]any `json:"record"`
	Valid     bool           `json:"valid"`
	Field     string         `json:"field,omitempty"`
	Canonical string         `json:"canonical,omitempty"`
	Live      []randomLive   `json:"live,omitempty"`
}

type randomLive struct {
	Now    int64  `json:"now"`
	Height uint64 `json:"height"`
	Valid  bool   `json:"valid"`
	Field  string `json:"field,omitempty"`
}

type randomMutation struct {
	name    string
	field   string
	valid   bool
	expired bool
	apply   func(map[string]any)
}

func fieldSet(key string, value any) func(map[string]any) {
	return func(record map[string]any) { record[key] = value }
}

func fieldDrop(key string) func(map[string]any) {
	return func(record map[string]any) { delete(record, key) }
}

// Every mutation except "valid" and "expired time" must be refused by both readers. The literal values
// are invalid on both chains, so a randomized case cannot pass by choosing the one chain where a value
// happens to remain well formed.
var randomMutations = []randomMutation{
	{name: "valid", valid: true},
	{name: "wrong version", field: "v", apply: fieldSet("v", 2)},
	{name: "zero version", field: "v", apply: fieldSet("v", 0)},
	{name: "unknown chain", field: "chain", apply: fieldSet("chain", "base")},
	{name: "empty asset", field: "asset", apply: fieldSet("asset", "")},
	{name: "zero asset", field: "asset", apply: fieldSet("asset", "0x0000000000000000000000000000000000000000")},
	{name: "zero maker", field: "maker", apply: fieldSet("maker", "0x0000000000000000000000000000000000000000")},
	{name: "leading-zero token", field: "token", apply: fieldSet("token", "01")},
	{name: "token overflow", field: "token", apply: fieldSet("token", strings.Repeat("9", 78))},
	{name: "bad side", field: "side", apply: fieldSet("side", "trade")},
	{name: "zero price", field: "price", apply: fieldSet("price", "0")},
	{name: "leading-zero price", field: "price", apply: fieldSet("price", "01")},
	{name: "price overflow", field: "price", apply: fieldSet("price", strings.Repeat("9", 78))},
	{name: "symbol pay", field: "pay", apply: fieldSet("pay", "USDC")},
	{name: "empty pay", field: "pay", apply: fieldSet("pay", "")},
	{name: "zero time expiry", field: "expiry", apply: fieldSet("expiry", "time:0")},
	{name: "zero height expiry", field: "expiry", apply: fieldSet("expiry", "height:0")},
	{name: "unknown expiry kind", field: "expiry", apply: fieldSet("expiry", "block:100")},
	{name: "short nonce", field: "nonce", apply: fieldSet("nonce", "aa")},
	{name: "zero nonce", field: "nonce", apply: fieldSet("nonce", strings.Repeat("0", 64))},
	{name: "short ref", field: "ref", apply: fieldSet("ref", "bb")},
	{name: "zero ref", field: "ref", apply: fieldSet("ref", strings.Repeat("0", 64))},
	{name: "adapter without version", field: "adapter", apply: fieldSet("adapter", "seaport")},
	{name: "unknown field", field: "extra", apply: func(record map[string]any) { record["extra"] = 1 }},
	{name: "missing asset", field: "asset", apply: fieldDrop("asset")},
	{name: "expired time", valid: true, expired: true, apply: fieldSet("expiry", "time:1")},
}

var randomAdapters = []string{"seaport/1.6.0", "magiceden/2.0.0", "blur/1.0.0", "opensea/2.3.1"}

func randomAddress(r *rand.Rand) string {
	b := make([]byte, 20)
	r.Read(b)
	b[0] |= 1
	return "0x" + hex.EncodeToString(b)
}

func randomHash(r *rand.Rand) string {
	b := make([]byte, 32)
	r.Read(b)
	b[0] |= 1
	return hex.EncodeToString(b)
}

func randomExpiry(r *rand.Rand, chain string) string {
	if chain == "sol" || r.Intn(2) == 0 {
		return "height:" + strconv.FormatUint(uint64(r.Intn(1_000_000_000))+2, 10)
	}
	return "time:" + strconv.FormatInt(r.Int63n(2_000_000_000)+2, 10)
}

// randomRecord creates a valid record on either chain. Solana asset/token/maker stay the published
// constants because generating another canonical 32-byte base58 key adds a second encoder to the test,
// not another rule under test. The varied fields are the ones the validator accepts in many spellings.
func randomRecord(r *rand.Rand) map[string]any {
	chain := "eth"
	if r.Intn(2) == 0 {
		chain = "sol"
	}
	record := map[string]any{
		"v":       Version,
		"chain":   chain,
		"side":    []string{"sell", "buy"}[r.Intn(2)],
		"price":   strconv.FormatUint(uint64(r.Intn(1_000_000_000))+1, 10),
		"nonce":   randomHash(r),
		"adapter": randomAdapters[r.Intn(len(randomAdapters))],
		"ref":     randomHash(r),
	}
	if chain == "eth" {
		record["asset"] = randomAddress(r)
		if r.Intn(4) == 0 {
			record["token"] = "0"
		} else {
			record["token"] = strconv.FormatUint(r.Uint64(), 10)
		}
		if r.Intn(3) == 0 {
			record["pay"] = "native"
		} else {
			record["pay"] = randomAddress(r)
		}
		record["maker"] = randomAddress(r)
	} else {
		record["asset"] = solAsset
		record["token"] = solToken
		record["maker"] = solMaker
		if r.Intn(2) == 0 {
			record["pay"] = "native"
		} else {
			record["pay"] = solToken
		}
	}
	record["expiry"] = randomExpiry(r, chain)
	return record
}

func liveCases(in Intent, expired bool) ([]randomLive, error) {
	e, err := ParseExpiry(in.Expiry)
	if err != nil {
		return nil, err
	}
	live := []randomLive{{Now: 0, Height: 0, Valid: true}, {Now: 0, Height: 0, Valid: false, Field: "expiry"}}
	switch e.Kind {
	case ExpiryTime:
		live[0].Now = int64(e.Value) - 1
		live[1].Now = int64(e.Value)
	case ExpiryHeight:
		live[0].Height = e.Value - 1
		live[1].Height = e.Value
	}
	if expired {
		// time:1 is already dead at the epoch, but the before/at pair above still exercises Live.
		live[0].Now, live[1].Now = 0, 1
	}
	return live, nil
}

func generateRandomCorpus() (randomCorpus, error) {
	r := rand.New(rand.NewSource(randomSeed))
	corpus := randomCorpus{Version: Version, Seed: randomSeed, Rounds: randomRounds}
	corpus.Cases = make([]randomCase, 0, randomRounds)
	for i := 0; i < randomRounds; i++ {
		record := randomRecord(r)
		mutation := randomMutations[i%len(randomMutations)]
		if mutation.apply != nil {
			mutation.apply(record)
		}
		c := randomCase{
			Name:   fmt.Sprintf("%03d-%s-%s", i, record["chain"], mutation.name),
			Record: record,
			Valid:  mutation.valid,
			Field:  mutation.field,
		}
		if mutation.valid {
			raw, err := json.Marshal(record)
			if err != nil {
				return corpus, err
			}
			in, err := Parse(raw)
			if err != nil {
				return corpus, fmt.Errorf("%s: generated invalid record: %w", c.Name, err)
			}
			canonical, err := in.Canonical()
			if err != nil {
				return corpus, err
			}
			c.Canonical = string(canonical)
			c.Live, err = liveCases(in, mutation.expired)
			if err != nil {
				return corpus, err
			}
		}
		corpus.Cases = append(corpus.Cases, c)
	}
	return corpus, nil
}

func TestRandomVectors(t *testing.T) {
	corpus, err := generateRandomCorpus()
	if err != nil {
		t.Fatalf("generate random corpus: %v", err)
	}
	b, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, '\n')
	if os.Getenv("SIGNAL_UPDATE") == "1" {
		if err := os.WriteFile(randomVectorsPath, b, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s (%d cases)", randomVectorsPath, len(corpus.Cases))
		return
	}
	want, err := os.ReadFile(randomVectorsPath)
	if err != nil {
		t.Fatalf("random vectors are part of this package and must be present: %v", err)
	}
	if !bytes.Equal(b, want) {
		t.Fatal("random vector corpus is stale; run SIGNAL_UPDATE=1 go test ./signal -run '^TestRandomVectors$'")
	}
	if corpus.Version != Version {
		t.Fatalf("random corpus is for v%d, this reader accepts v%d", corpus.Version, Version)
	}

	var valid, invalid int
	for _, tc := range corpus.Cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			raw, err := json.Marshal(tc.Record)
			if err != nil {
				t.Fatal(err)
			}
			in, err := Parse(raw)
			if !tc.Valid {
				if err == nil {
					t.Fatalf("Parse accepted an invalid record: %s", raw)
				}
				if tc.Field != "" && !strings.Contains(err.Error(), tc.Field) {
					t.Fatalf("Parse error = %q, want it to mention %q", err, tc.Field)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			got, err := in.Canonical()
			if err != nil {
				t.Fatalf("Canonical: %v", err)
			}
			if string(got) != tc.Canonical {
				t.Fatalf("Canonical:\n got: %s\nwant: %s", got, tc.Canonical)
			}
			for _, lc := range tc.Live {
				err := in.Live(time.Unix(lc.Now, 0), lc.Height)
				if lc.Valid {
					if err != nil {
						t.Fatalf("Live(now=%d, height=%d): %v", lc.Now, lc.Height, err)
					}
					continue
				}
				if err == nil {
					t.Fatalf("Live accepted a dead record at now=%d, height=%d", lc.Now, lc.Height)
				}
				if lc.Field != "" && !strings.Contains(err.Error(), lc.Field) {
					t.Fatalf("Live error = %q, want it to mention %q", err, lc.Field)
				}
			}
		})
		if tc.Valid {
			valid++
		} else {
			invalid++
		}
	}
	if valid < 10 || invalid < 100 {
		t.Fatalf("random corpus is not broad enough: %d valid, %d invalid", valid, invalid)
	}
}
