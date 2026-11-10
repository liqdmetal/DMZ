// Package signal defines the public DERO intent record used by the signal-first DEX and validates it
// before any board read is shown or acted on. It performs no chain calls: the caller reads the record
// from DERO (or a transport) and decides separately whether to verify the source chain.
//
// The record is advisory discovery data. Validation here says the record is well formed and live; it
// does not prove that the maker owns the NFT, approved a marketplace, or can settle.
// The browser mirror is app/signal.js; both readers are held to testdata/signal_v1_vectors.json.
package signal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Version is the only signal schema version this package reads. A new layout must bump both the
// document and this constant; an older reader must reject it rather than guess.
const Version = 1

// Chain is the NFT's native chain. Settlement happens there, never on DERO.
type Chain string

const (
	ChainETH Chain = "eth"
	ChainSOL Chain = "sol"
)

// Side is what the maker offers to do.
type Side string

const (
	SideSell Side = "sell"
	SideBuy  Side = "buy"
)

// ExpiryKind distinguishes an absolute wall-clock deadline from a source-chain height.
type ExpiryKind string

const (
	ExpiryTime   ExpiryKind = "time"
	ExpiryHeight ExpiryKind = "height"
)

// Intent is the public signal record. All integer-valued fields are decimal strings so a JavaScript
// reader never loses precision; the canonical form is JSON with these fields in this order and no
// whitespace.
type Intent struct {
	V       int    `json:"v"`
	Chain   Chain  `json:"chain"`
	Asset   string `json:"asset"`
	Token   string `json:"token"`
	Side    Side   `json:"side"`
	Price   string `json:"price"`
	Pay     string `json:"pay"`
	Expiry  string `json:"expiry"` // "time:<unix-seconds>" or "height:<source-height>"
	Maker   string `json:"maker"`
	Nonce   string `json:"nonce"` // 32-byte lowercase hex, unique per maker/chain
	Adapter string `json:"adapter"`
	Ref     string `json:"ref"` // 32-byte lowercase hex of the full terms document
}

// Expiry is the parsed form of Intent.Expiry.
type Expiry struct {
	Kind  ExpiryKind
	Value uint64
}

var adapterRE = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}/[0-9]+\.[0-9]+\.[0-9]+$`)

// Parse strictly decodes and validates one public signal. Unknown fields, duplicate keys, trailing
// JSON and non-canonical values are refused, because two readers must never disagree about what a
// record says.
func Parse(data []byte) (Intent, error) {
	var in Intent
	if err := checkNoDuplicateKeys(data); err != nil {
		return in, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return in, fmt.Errorf("signal: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return in, errors.New("signal: trailing data after the record")
		}
		return in, fmt.Errorf("signal: trailing data: %w", err)
	}
	if err := in.Validate(); err != nil {
		return in, err
	}
	return in, nil
}

// Validate checks the record's structure and canonical spelling. It does not reject an expired
// record: use Live for that, because expiry depends on a clock or a source-chain height.
func (in Intent) Validate() error {
	if in.V != Version {
		return fmt.Errorf("v: got %d, this reader accepts %d", in.V, Version)
	}
	switch in.Chain {
	case ChainETH, ChainSOL:
	default:
		return fmt.Errorf("chain: %q is not eth or sol", in.Chain)
	}
	if !validAddress(in.Chain, in.Asset) {
		return fmt.Errorf("asset: %q is not a canonical %s address", in.Asset, in.Chain)
	}
	if !validToken(in.Chain, in.Token) {
		return fmt.Errorf("token: %q is not a canonical %s token id", in.Token, in.Chain)
	}
	switch in.Side {
	case SideSell, SideBuy:
	default:
		return fmt.Errorf("side: %q is not sell or buy", in.Side)
	}
	if !validUint(in.Price, 256, false) {
		return fmt.Errorf("price: %q is not a positive canonical uint256", in.Price)
	}
	if in.Pay != "native" && !validAddress(in.Chain, in.Pay) {
		return fmt.Errorf("pay: %q is neither native nor a canonical %s address", in.Pay, in.Chain)
	}
	if _, err := ParseExpiry(in.Expiry); err != nil {
		return fmt.Errorf("expiry: %w", err)
	}
	if !validAddress(in.Chain, in.Maker) {
		return fmt.Errorf("maker: %q is not a canonical %s address", in.Maker, in.Chain)
	}
	if !validHash(in.Nonce) {
		return fmt.Errorf("nonce: %q is not 32 non-zero lowercase hex bytes", in.Nonce)
	}
	if !adapterRE.MatchString(in.Adapter) {
		return fmt.Errorf("adapter: %q is not <name>/<major>.<minor>.<patch>", in.Adapter)
	}
	if !validHash(in.Ref) {
		return fmt.Errorf("ref: %q is not 32 non-zero lowercase hex bytes", in.Ref)
	}
	return nil
}

// Canonical returns the bytes that are hashed or compared when two readers need the same record.
// It is JSON with the fields in Intent order, no whitespace, and no HTML escaping.
func (in Intent) Canonical() ([]byte, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(in); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n")), nil
}

// ParseExpiry parses "time:<unix-seconds>" or "height:<source-height>". Both values are positive.
func ParseExpiry(s string) (Expiry, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return Expiry{}, fmt.Errorf("%q is not time:<seconds> or height:<blocks>", s)
	}
	v, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || v == 0 {
		return Expiry{}, fmt.Errorf("%q has no positive integer value", s)
	}
	switch ExpiryKind(parts[0]) {
	case ExpiryTime:
		if v > math.MaxInt64 {
			return Expiry{}, fmt.Errorf("%q is beyond a Unix timestamp", s)
		}
		return Expiry{Kind: ExpiryTime, Value: v}, nil
	case ExpiryHeight:
		return Expiry{Kind: ExpiryHeight, Value: v}, nil
	default:
		return Expiry{}, fmt.Errorf("%q is not time:<seconds> or height:<blocks>", s)
	}
}

// Live validates the record and then rejects it once its expiry has been reached. A time expiry is
// dead at now >= expiry; a height expiry is dead at source height >= expiry. Pass height 0 when the
// source height is unknown; a height-expiring record then remains live and the caller must check it
// against the source chain before acting.
func (in Intent) Live(now time.Time, height uint64) error {
	if err := in.Validate(); err != nil {
		return err
	}
	e, _ := ParseExpiry(in.Expiry) // Validate already established this
	switch e.Kind {
	case ExpiryTime:
		if now.Unix() >= int64(e.Value) {
			return fmt.Errorf("expiry: time %d has passed", e.Value)
		}
	case ExpiryHeight:
		if height != 0 && height >= e.Value {
			return fmt.Errorf("expiry: source height %d has reached %d", height, e.Value)
		}
	}
	return nil
}

func validToken(chain Chain, s string) bool {
	switch chain {
	case ChainETH:
		return validUint(s, 256, true) // token id 0 is legal
	case ChainSOL:
		return validAddress(chain, s)
	default:
		return false
	}
}

// validUint accepts a canonical unsigned decimal with no sign, no whitespace and no leading zero
// (except the single character "0"). maxBits is 64 or 256; allowZero controls whether 0 is legal.
func validUint(s string, maxBits int, allowZero bool) bool {
	if s == "" || (s[0] == '0' && len(s) > 1) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	n, ok := new(big.Int).SetString(s, 10)
	if !ok || n.BitLen() > maxBits {
		return false
	}
	return allowZero || n.Sign() != 0
}

func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	nonzero := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
		if c != '0' {
			nonzero = true
		}
	}
	return nonzero
}

func validAddress(chain Chain, s string) bool {
	switch chain {
	case ChainETH:
		if len(s) != 42 || !strings.HasPrefix(s, "0x") {
			return false
		}
		body := s[2:]
		for i := 0; i < len(body); i++ {
			c := body[i]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
		return body != strings.Repeat("0", 40)
	case ChainSOL:
		b, ok := decodeBase58(s)
		if !ok || len(b) != 32 {
			return false
		}
		for _, c := range b {
			if c != 0 {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func decodeBase58(s string) ([]byte, bool) {
	const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	if s == "" {
		return nil, false
	}
	n := new(big.Int)
	base := big.NewInt(58)
	for i := 0; i < len(s); i++ {
		d := strings.IndexByte(alphabet, s[i])
		if d < 0 {
			return nil, false
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(d)))
	}
	b := n.Bytes()
	pad := 0
	for pad < len(s) && s[pad] == '1' {
		pad++
	}
	out := make([]byte, pad+len(b))
	copy(out[pad:], b)
	return out, true
}

// checkNoDuplicateKeys walks the JSON token stream before decoding. encoding/json keeps the last
// duplicate, which means two readers can see two different records from the same bytes.
func checkNoDuplicateKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := walkValue(dec, "$"); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			return errors.New("signal: trailing data after the record")
		}
		return fmt.Errorf("signal: trailing data: %w", err)
	}
	return nil
}

func walkValue(dec *json.Decoder, path string) error {
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("signal: invalid JSON at %s: %w", path, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return fmt.Errorf("signal: invalid object at %s: %w", path, err)
			}
			key, ok := kt.(string)
			if !ok {
				return fmt.Errorf("signal: non-string key at %s", path)
			}
			if seen[key] {
				return fmt.Errorf("signal: duplicate key %q at %s", key, path)
			}
			seen[key] = true
			if err := walkValue(dec, path+"."+key); err != nil {
				return err
			}
		}
		if _, err := dec.Token(); err != nil { // closing }
			return fmt.Errorf("signal: invalid object at %s: %w", path, err)
		}
	case '[':
		i := 0
		for dec.More() {
			if err := walkValue(dec, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
			i++
		}
		if _, err := dec.Token(); err != nil { // closing ]
			return fmt.Errorf("signal: invalid array at %s: %w", path, err)
		}
	default:
		return fmt.Errorf("signal: unexpected delimiter %q at %s", delim, path)
	}
	return nil
}
