package tests

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/veil/tests/harness"
)

// WT-58: the operator documents must describe the contract that ships. These tests pin the claims that
// went stale (halt scope, Tighten, the lock file, gas, finality and custody) and every local prefixed
// citation in the acceptance matrix. They read the real checkout through VEIL_REPO_ROOT because
// tests/run.sh copies only tests/, contracts/ and reference/ into the vendored module.
func repoPath(t *testing.T, rel string) string {
	t.Helper()
	root := os.Getenv("VEIL_REPO_ROOT")
	if root == "" {
		root = "../.."
	}
	return root + string(os.PathSeparator) + strings.ReplaceAll(rel, "/", string(os.PathSeparator))
}

func readRepo(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(repoPath(t, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

func mustSay(t *testing.T, file, text, want string) {
	t.Helper()
	if !strings.Contains(flat(text), flat(want)) {
		t.Errorf("%s no longer says %q", file, want)
	}
}

func mustNotSay(t *testing.T, file, text, gone string) {
	t.Helper()
	if strings.Contains(flat(text), flat(gone)) {
		t.Errorf("%s still says the stale claim %q", file, gone)
	}
}

func flat(s string) string { return strings.Join(strings.Fields(s), " ") }

// TestWT58CitedLines checks the passages WT-58 names, not just that the files exist. Keep these line
// numbers in the row current; a moved paragraph is exactly the drift this test is for.
func TestWT58CitedLines(t *testing.T) {
	cites := []struct {
		file string
		line int
		want string
	}{
		{"research/wrapper/RUNBOOKS.md", 97, "untouched for 30 minutes is taken over"},
		{"research/wrapper/RUNBOOKS.md", 268, "three entrypoints that refuse while halted"},
		{"contracts/README.md", 27, "16,358–16,407 by install shape"},
		{"contracts/README.md", 490, "`Burn` and `Released` work while halted"},
		{"deploy/VAULT_PARAMETERS.md", 32, "tighten only (k votes; can rise, never fall)"},
		{"deploy/VAULT_PARAMETERS.md", 34, "tighten only (k votes; can fall, never rise to 0)"},
		{"docs/HANDOVER.md", 230, "untouched for 30 minutes is taken over"},
		{"docs/HANDOVER.md", 241, "Zebra: tip − 100; Monero: tip − 10"},
		{"docs/HANDOVER.md", 380, "16,358–16,407"},
		{"research/wrapper/README.md", 108, "trusted for custody on the foreign chain"},
		{"docs/WRAP_ACCEPTANCE_MATRIX.md", 123, "Reconciled against v3"},
	}
	for _, c := range cites {
		lines := strings.Split(readRepo(t, c.file), "\n")
		if c.line < 1 || c.line > len(lines) {
			t.Errorf("%s:%d is past the end (%d lines)", c.file, c.line, len(lines))
			continue
		}
		if !strings.Contains(lines[c.line-1], c.want) {
			t.Errorf("%s:%d no longer says %q; reads %q", c.file, c.line, c.want, strings.TrimSpace(lines[c.line-1]))
		}
	}
}

var wt58Docs = []string{
	"research/wrapper/RUNBOOKS.md",
	"research/wrapper/README.md",
	"deploy/VAULT_PARAMETERS.md",
	"docs/HANDOVER.md",
	"contracts/README.md",
	"docs/WRAP_ACCEPTANCE_MATRIX.md",
}

// TestWT58OperatorDocsAgree is the reconciliation half: it fails if one document drifts back to the
// pre-v3 behavior or if the gas/parameter/finality numbers stop naming their shape.
func TestWT58OperatorDocsAgree(t *testing.T) {
	runbook := readRepo(t, "research/wrapper/RUNBOOKS.md")
	mustSay(t, "RUNBOOKS.md", runbook, "`Register`, `Attest` and `Mint` are the three entrypoints that refuse while halted")
	mustSay(t, "RUNBOOKS.md", runbook, "`Burn` and `Released` have no halt guard")
	mustNotSay(t, "RUNBOOKS.md", runbook, "no burns, no releases")
	mustNotSay(t, "RUNBOOKS.md", runbook, "Released` is one of the entrypoints a halt blocks")
	mustSay(t, "RUNBOOKS.md", runbook, "`finalityDepth` 100 on Zcash, 10 on Monero")
	mustSay(t, "RUNBOOKS.md", runbook, "30 minutes is taken over")

	contracts := readRepo(t, "contracts/README.md")
	mustSay(t, "contracts/README.md", contracts, "`Burn` and `Released` work while halted")
	mustSay(t, "contracts/README.md", contracts, "16,358–16,407 by install shape")
	mustSay(t, "contracts/README.md", contracts, "3,593 below at its largest install shape")
	mustSay(t, "contracts/README.md", contracts, "The signers can steal or freeze")

	wrapper := readRepo(t, "research/wrapper/README.md")
	mustSay(t, "research/wrapper/README.md", wrapper, "Only Register, Attest and Mint refuse while halted")
	mustSay(t, "research/wrapper/README.md", wrapper, "`Burn` and `Released` work")
	mustSay(t, "research/wrapper/README.md", wrapper, "per-backend `Finalized()` (Zcash 100, Monero 10)")
	mustNotSay(t, "research/wrapper/README.md", wrapper, "mint, burn or release while halted")
	mustSay(t, "research/wrapper/README.md", wrapper, "trusted for custody on the foreign chain")

	params := readRepo(t, "deploy/VAULT_PARAMETERS.md")
	mustSay(t, "deploy/VAULT_PARAMETERS.md", params, "tighten only (k votes; can rise, never fall)")
	mustSay(t, "deploy/VAULT_PARAMETERS.md", params, "tighten only (k votes; can fall, never rise to 0)")
	mustSay(t, "deploy/VAULT_PARAMETERS.md", params, "| Zcash | tip − 100")
	mustSay(t, "deploy/VAULT_PARAMETERS.md", params, "| Monero | tip − 10")
	mustSay(t, "deploy/VAULT_PARAMETERS.md", params, "16,358–16,407")
	mustNotSay(t, "deploy/VAULT_PARAMETERS.md", params, "| `wdelay` | `SetLimits` before Seal | no |")
	mustNotSay(t, "deploy/VAULT_PARAMETERS.md", params, "| `winmax` | `SetLimits` before Seal | no |")

	handover := readRepo(t, "docs/HANDOVER.md")
	mustSay(t, "docs/HANDOVER.md", handover, "pre-fix contract refused `Released` while halted")
	mustSay(t, "docs/HANDOVER.md", handover, "30 minutes is taken over")
	mustSay(t, "docs/HANDOVER.md", handover, "Zebra: tip − 100")
	mustSay(t, "docs/HANDOVER.md", handover, "Monero: tip − 10")
	mustSay(t, "docs/HANDOVER.md", handover, "16,358–16,407")
	mustNotSay(t, "docs/HANDOVER.md", handover, "removed on every exit path")

	matrix := readRepo(t, "docs/WRAP_ACCEPTANCE_MATRIX.md")
	mustSay(t, "docs/WRAP_ACCEPTANCE_MATRIX.md", matrix, "Reconciled against v3")
	mustSay(t, "docs/WRAP_ACCEPTANCE_MATRIX.md", matrix, "tests/doccites_test.go")
}

// TestVaultHaltScopeMatchesDocs is the matrix's parse check: the three exposure-increasing entrypoints
// refuse while halted, and the two exits do not. It also pins Tighten's safe direction.
func TestVaultHaltScopeMatchesDocs(t *testing.T) {
	src := readRepo(t, "contracts/vault.bas")
	body := func(name string) string {
		t.Helper()
		start := strings.Index(src, "Function "+name+"(")
		if start < 0 {
			t.Fatalf("vault.bas has no Function %s", name)
		}
		end := strings.Index(src[start:], "\nEnd Function")
		if end < 0 {
			t.Fatalf("vault.bas Function %s has no End Function", name)
		}
		return src[start : start+end]
	}
	for _, name := range []string{"Register", "Attest", "Mint"} {
		if !strings.Contains(body(name), "ishalted()") {
			t.Errorf("%s must guard on ishalted()", name)
		}
	}
	for _, name := range []string{"Burn", "Released"} {
		if strings.Contains(body(name), "ishalted()") {
			t.Errorf("%s must not guard on ishalted(); a halt cannot block the exit", name)
		}
	}
	tighten := body("Tighten")
	for _, want := range []string{`STORE("winmax"`, `STORE("wdelay"`, `LOAD("sealed")`} {
		if !strings.Contains(tighten, want) {
			t.Errorf("Tighten no longer contains %q", want)
		}
	}
	setLimits := body("SetLimits")
	if !strings.Contains(setLimits, `LOAD("sealed") != 0`) {
		t.Error("SetLimits is no longer owner-before-Seal only")
	}
}

// TestDocLockMatchesDaemon pins the lock-file wording to the daemon's actual heartbeat and takeover.
func TestDocLockMatchesDaemon(t *testing.T) {
	main := readRepo(t, "tools/vaultsigner/main.go")
	for _, want := range []string{"const lockStale = 30 * time.Minute", "os.Chtimes(lockFile", "os.Remove(lockFile)"} {
		if !strings.Contains(main, want) {
			t.Errorf("vaultsigner/main.go no longer contains %q", want)
		}
	}
	if !strings.Contains(readRepo(t, "research/wrapper/RUNBOOKS.md"), "untouched for 30 minutes is taken over") {
		t.Error("RUNBOOKS.md does not describe the stale-lock takeover")
	}
}

// TestDocFinalityMatchesBackends pins the per-backend defaults the documents now state.
func TestDocFinalityMatchesBackends(t *testing.T) {
	zcash := readRepo(t, "tools/vaultsigner/zcashvault/backend.go")
	if !strings.Contains(zcash, "if r.Depth == 0 {\n\t\tr.Depth = 100") {
		t.Error("Zcash Finalized no longer defaults to 100")
	}
	monero := readRepo(t, "tools/vaultsigner/monerovault/backend.go")
	if !strings.Contains(monero, "cfg.FinalityDepth = 10") || !strings.Contains(monero, "cfg.FinalityDepth < 10") {
		t.Error("Monero finality no longer defaults to and floors at 10")
	}
	handover := readRepo(t, "docs/HANDOVER.md")
	for _, want := range []string{"(Zebra: tip − 100; Monero: tip − 10", "Finalized()` = tip − 10"} {
		if !strings.Contains(handover, want) {
			t.Errorf("HANDOVER.md no longer states %q", want)
		}
	}
}

// TestDocGasMatchesInstall measures the three documented install shapes. A comment or line can move;
// a gas number that no longer matches the DVM cannot.
func TestDocGasMatchesInstall(t *testing.T) {
	shapes := []struct {
		name                        string
		cap, minbond, minburn, wfee uint64
		mindep, capratio            uint64
		bigbps, bigdelay            uint64
		want                        uint64
	}{
		{"smoke", 9000, 1000, 100, 10, 10, 3, 2000, 0, 16358},
		{"zcash", 100_000_000_000, 20000, 1_000_000, 10_000, 100_000, 2_000_000, 2000, 500, 16392},
		{"monero", 100_000_000_000_000, 20000, 10_000_000_000, 100_000_000, 1_000_000_000, 2_000_000_000, 2000, 500, 16407},
	}
	code := readContract(t, "vault.bas")
	for _, tc := range shapes {
		t.Run(tc.name, func(t *testing.T) {
			c := harness.New()
			owner, treasury := c.NewAccount(), c.NewAccount()
			_, res := c.Install(code, nil, rpc.Arguments{
				harness.U("k", 2), harness.S("treasury", treasury.String()),
				harness.U("cap", tc.cap), harness.U("minbond", tc.minbond),
				harness.U("minburn", tc.minburn), harness.U("wfee", tc.wfee),
				harness.U("mindep", tc.mindep), harness.U("regfee", 0),
				harness.U("capratio", tc.capratio), harness.U("bigbps", tc.bigbps),
				harness.U("bigdelay", tc.bigdelay),
			}, owner, harness.MaxStorageGas)
			if !res.OK() {
				t.Fatalf("install: %v", res.Err)
			}
			if res.GasStorage != tc.want {
				t.Errorf("%s install storage gas = %d, documented %d", tc.name, res.GasStorage, tc.want)
			}
			if res.GasStorage >= harness.MaxStorageGas {
				t.Fatalf("%s install storage gas %d is at the cap", tc.name, res.GasStorage)
			}
		})
	}
	docs := []string{"contracts/README.md", "docs/HANDOVER.md", "deploy/VAULT_PARAMETERS.md"}
	for _, rel := range docs {
		text := readRepo(t, rel)
		for _, want := range []string{"16,358", "16,392", "16,407"} {
			if !strings.Contains(text, want) {
				t.Errorf("%s no longer records the measured gas %s", rel, want)
			}
		}
	}
}

var prefixedCite = regexp.MustCompile(`\b(main|wc|zfee|wm|wv|wz):([A-Za-z0-9_./-]+):([0-9]+)(?:-([0-9]+))?`)

// resolveCite maps a matrix prefix to the merged tree. wm: and wv: are unresolved source branches in
// this checkout; they are skipped rather than asserted against the wrong file.
func resolveCite(prefix, rel string) (string, bool) {
	switch prefix {
	case "main":
		for _, p := range []string{
			rel,
			"contracts/" + rel,
			"app/" + rel,
			"tools/vaultsigner/" + rel,
			"tools/simvault/" + rel,
			"research/wrapper/" + rel,
		} {
			if _, err := os.Stat(repoForCite(p)); err == nil {
				return p, true
			}
		}
	case "wc":
		for _, p := range []string{"tools/vaultsigner/" + rel, rel} {
			if _, err := os.Stat(repoForCite(p)); err == nil {
				return p, true
			}
		}
	case "zfee":
		p := "app/" + rel
		if _, err := os.Stat(repoForCite(p)); err == nil {
			return p, true
		}
	case "wz":
		p := "tools/vaultsigner/" + rel
		if _, err := os.Stat(repoForCite(p)); err == nil {
			return p, true
		}
	}
	return "", false
}

func repoForCite(rel string) string {
	root := os.Getenv("VEIL_REPO_ROOT")
	if root == "" {
		root = "../.."
	}
	return root + string(os.PathSeparator) + strings.ReplaceAll(rel, "/", string(os.PathSeparator))
}

// TestDoccitesResolve runs over the WT-58 documents and the acceptance matrix. It checks every local
// prefixed citation (main:, wc:, zfee:, wz:) is inside its file; wm:/wv: point at trees absent here.
func TestDoccitesResolve(t *testing.T) {
	docs := append([]string{}, wt58Docs...)
	checked := 0
	skipped := 0
	for _, rel := range docs {
		text := readRepo(t, rel)
		for _, m := range prefixedCite.FindAllStringSubmatch(text, -1) {
			prefix, path := m[1], m[2]
			target, ok := resolveCite(prefix, path)
			if !ok {
				skipped++
				continue
			}
			start, _ := strconv.Atoi(m[3])
			end := start
			if m[4] != "" {
				end, _ = strconv.Atoi(m[4])
			}
			lines := strings.Split(readRepo(t, target), "\n")
			if start < 1 || end < start || end > len(lines) {
				t.Errorf("%s cites %s:%s-%s, but %s has %d lines", rel, prefix, m[3], orSame(m[4], m[3]), target, len(lines))
				continue
			}
			checked++
		}
	}
	if checked < 50 {
		t.Fatalf("only %d local citations checked (%d skipped): the citation test has gone vacuous", checked, skipped)
	}
	t.Logf("checked %d local prefixed citations; skipped %d citations to absent wm/wv trees", checked, skipped)
}

func orSame(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// TestDoccitesMatrixShape guards the matrix's own header against a silent reversion to the old
// contradictory citation set.
func TestDoccitesMatrixShape(t *testing.T) {
	matrix := readRepo(t, "docs/WRAP_ACCEPTANCE_MATRIX.md")
	for _, want := range []string{
		"| WT-58 |",
		"tests/doccites_test.go",
		"Reconciled against v3",
		"16,358–16,407",
	} {
		if !strings.Contains(matrix, want) {
			t.Errorf("WRAP_ACCEPTANCE_MATRIX.md no longer contains %q", want)
		}
	}
	if strings.Contains(matrix, "`research/wrapper/RUNBOOKS.md:262-267` says a halt blocks") {
		t.Error("the matrix still points at the stale halt-scope citation")
	}
}
