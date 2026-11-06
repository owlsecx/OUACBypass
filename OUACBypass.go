// OUACBypass — UAC Bypass Technique Catalog
// OWL Batch-3 offensive security toolkit.
// Author : khaled.s.haddad    Website : owlsec.org
// Language: Go (standard library only; static, CGO-free build).
// AUTHORIZED TESTING ONLY. See --howto for the legal & ethical warning.

package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ---- OWL UI KIT (red/black) — stdlib only. Paste into each tool's main.go ----
const (
	R   = "\033[91m"
	DR  = "\033[96m"
	GRN = "\033[92m"
	YLW = "\033[93m"
	CY  = "\033[96m"
	W   = "\033[97m"
	B   = "\033[1m"
	X   = "\033[0m"
	MG  = "\033[95m"
	DIM = "\033[2m"
	OR  = "\033[38;5;208m"
	GY  = "\033[90m"
)

const (
	SUITE   = "Owl Windows Suite"
	AUTHOR  = "khaled.s.haddad"
	WEBSITE = "owlsec.org"
)

var noColor = os.Getenv("NO_COLOR") != "" || !isatty()

func isatty() bool { fi, _ := os.Stdout.Stat(); return (fi.Mode() & os.ModeCharDevice) != 0 }
func col(s string, c ...string) string {
	if noColor {
		return s
	}
	return strings.Join(c, "") + s + X
}
func clear() {
	if !noColor {
		fmt.Print("\033[H\033[2J")
	}
}
func rule()            { fmt.Println(col(strings.Repeat("─", 66), CY)) }
func line(w int)       { fmt.Println(col(strings.Repeat("─", w), CY)) }
func okMsg(s string)   { fmt.Println("  " + col("[✔] ", GRN+B) + col(s, W)) }
func errMsg(s string)  { fmt.Println("  " + col("[✘] ", R+B) + col(s, R)) }
func warnMsg(s string) { fmt.Println("  " + col("[!] ", OR+B) + col(s, YLW)) }
func info(s string)    { fmt.Println("  " + col("[*] ", CY) + col(s, DIM+W)) }
func header(title string) {
	fmt.Println()
	rule()
	fmt.Println("  " + col("⬡  "+title, B+CY))
	rule()
	fmt.Println()
}

var stdin = bufio.NewReader(os.Stdin)

func prompt(label string) string {
	fmt.Print("  " + col("[»] ", YLW) + col(label+": ", W))
	s, err := stdin.ReadString('\n')
	if err != nil && s == "" {
		fmt.Println()
		os.Exit(0)
	}
	return strings.TrimSpace(s)
}
func pause() {
	fmt.Print(col("\n  [↩] Press Enter to continue...", YLW))
	stdin.ReadString('\n')
}
func stripANSI(s string) string {
	out := []rune{}
	esc := false
	for _, r := range s {
		if r == '\033' {
			esc = true
			continue
		}
		if esc {
			if r == 'm' {
				esc = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
func visw(s string) int { return len([]rune(stripANSI(s))) }
func pad(s string, n int) string {
	if visw(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-visw(s))
}
func table(headers []string, rows [][]string) {
	w := make([]int, len(headers))
	for i, h := range headers {
		w[i] = visw(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if i < len(w) && visw(c) > w[i] {
				w[i] = visw(c)
			}
		}
	}
	var hb strings.Builder
	for i, h := range headers {
		hb.WriteString(col(pad(h, w[i]), B+CY))
		if i < len(headers)-1 {
			hb.WriteString(col(" │ ", DIM))
		}
	}
	fmt.Println("  " + hb.String())
	seg := make([]string, len(w))
	for i := range w {
		seg[i] = strings.Repeat("─", w[i])
	}
	fmt.Println("  " + col(strings.Join(seg, "─┼─"), DIM))
	for _, r := range rows {
		var rb strings.Builder
		for i := range headers {
			v := ""
			if i < len(r) {
				v = r[i]
			}
			rb.WriteString(col(pad(v, w[i]), W))
			if i < len(headers)-1 {
				rb.WriteString(col(" │ ", DIM))
			}
		}
		fmt.Println("  " + rb.String())
	}
}
func banner() {
	fmt.Println(col(BANNER, CY))
	fmt.Println("  " + col(SUITE+"  ", B+MG) + col("v"+VERSION, YLW))
	fmt.Println("  " + col(TAGLINE, W))
	fmt.Println()
	fmt.Println("  " + col("Developer : ", DIM) + col(AUTHOR, CY) +
		col("   |   Website : ", DIM) + col(WEBSITE, CY))
	rule()
	fmt.Println()
}
func menu(title string, items []string) string {
	clear()
	banner()
	fmt.Println("  " + col("⬡  "+title, B+CY))
	fmt.Println()
	for i, it := range items {
		fmt.Println("  " + col(pad(fmt.Sprintf("[%d]", i+1), 4), CY+B) + "  " + col(it, W))
	}
	fmt.Println()
	fmt.Println("  " + col(pad("[H]", 4), YLW+B) + "  " + col("Help & Usage", W))
	fmt.Println("  " + col(pad("[0]", 4), R+B) + "  " + col("Exit", W))
	fmt.Println()
	rule()
	return prompt("Select option")
}

// ---- OWL house help (shared layout) ----
type owlHelpItem struct{ key, text string }

func owlHelpWrap(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	out := []string{}
	cur := ""
	for _, w := range words {
		if cur != "" && len([]rune(cur))+1+len([]rune(w)) > width {
			out = append(out, cur)
			cur = w
		} else if cur == "" {
			cur = w
		} else {
			cur += " " + w
		}
	}
	out = append(out, cur)
	return out
}
func owlHelpTitle(t string) { fmt.Println(); fmt.Println("  " + col("▸ "+t, B+CY)) }
func owlHelpPara(lines []string) {
	for _, l := range lines {
		for _, w := range owlHelpWrap(l, 60) {
			fmt.Println("    " + col(w, W))
		}
	}
}
func owlHelpRows(items []owlHelpItem) {
	kw := 0
	for _, it := range items {
		if n := len([]rune(it.key)); n > kw {
			kw = n
		}
	}
	for _, it := range items {
		k := "[" + it.key + "]" + strings.Repeat(" ", kw-len([]rune(it.key)))
		for i, w := range owlHelpWrap(it.text, 56-kw) {
			if i == 0 {
				fmt.Println("    " + col(k, YLW+B) + "  " + col(w, W))
			} else {
				fmt.Println("    " + strings.Repeat(" ", kw+2) + "  " + col(w, W))
			}
		}
	}
}
func tuiHelp() {
	clear()
	banner()
	rule()
	fmt.Println("  " + col("⬡  Help & Usage — "+APPNAME, B+CY))
	rule()
	owlHelpTitle("WHAT IT DOES")
	owlHelpPara(owlHelpAbout)
	owlHelpTitle("HOW TO USE THE MENU")
	owlHelpPara(owlHelpFlow)
	owlHelpTitle("MENU OPTIONS")
	owlHelpRows(owlHelpMenu)
	owlHelpTitle("TIPS")
	owlHelpRows(owlHelpTips)
	fmt.Println()
	rule()
	fmt.Println("  " + col("Next page: command-line flags & examples", DIM))
	pause()
	clear()
	usage()
	fmt.Println()
	fmt.Println("  " + col("Tip: the same reference is printed by  "+APPNAME+" --help", DIM))
	pause()
}

const APPNAME = "OUACBypass"
const TAGLINE = "UAC Bypass Technique Catalog"
const VERSION = "1.0.0"
const CATEGORY = "OAttack"

const BANNER = `
 ██████╗ ██╗   ██╗ █████╗  ██████╗ ██████╗ ██╗   ██╗██████╗  █████╗  ███████╗ ███████╗
██╔═══██╗██║   ██║██╔══██╗██╔════╝ ██╔══██╗╚██╗ ██╔╝██╔══██╗██╔══██╗██╔════╝ ██╔════╝ 
██║   ██║██║   ██║███████║██║      ██████╔╝ ╚████╔╝ ██████╔╝███████║███████╗ ███████╗ 
██║   ██║██║   ██║██╔══██║██║      ██╔══██╗  ╚██╔╝  ██╔═══╝ ██╔══██║╚════██║ ╚════██║ 
╚██████╔╝╚██████╔╝██║  ██║╚██████╗ ██████╔╝   ██║   ██║     ██║  ██║███████║ ███████║ 
 ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═════╝    ╚═╝   ╚═╝     ╚═╝  ╚═╝╚══════╝ ╚══════╝ 
`

// ============================================================================
//  SHARED ENGINE  —  types, CLI, export (txt/md/json/csv/sarif), grading,
//  HOW TO USE renderer, interactive red/black TUI, and domain utilities.
//  This block is identical across every OWL Batch-3 tool; each tool supplies
//  its own deep domain engine (catalogs, parsers, generators, scoring) that
//  plugs into the hooks declared at the bottom of this file.
// ============================================================================

// Finding is a single analysis result with MITRE ATT&CK mapping and a score.
type Finding struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Severity    string   `json:"severity"`
	Category    string   `json:"category"`
	Detail      string   `json:"detail"`
	Evidence    string   `json:"evidence"`
	Remediation string   `json:"remediation"`
	Mitre       []string `json:"mitre"`
	Score       int      `json:"score"`
	Ref         string   `json:"ref,omitempty"`
}

// Section is a titled block of lines used by HOW TO USE and catalog views.
type Section struct {
	Title string
	Lines []string
}

// MenuItem is one interactive action in the TUI.
type MenuItem struct {
	Label  string
	Action func()
}

// AnalysisResult is the full output of a domain analysis or demo run.
type AnalysisResult struct {
	Target   string            `json:"target"`
	Tool     string            `json:"tool"`
	When     string            `json:"when"`
	Findings []Finding         `json:"findings"`
	Stats    map[string]string `json:"stats"`
	Grade    string            `json:"grade"`
	Notes    []string          `json:"notes,omitempty"`
}

// ---------------------------------------------------------------------------
//  Severity model
// ---------------------------------------------------------------------------

const (
	SevCrit = "CRITICAL"
	SevHigh = "HIGH"
	SevMed  = "MEDIUM"
	SevLow  = "LOW"
	SevInfo = "INFO"
)

func sevRank(s string) int {
	switch strings.ToUpper(s) {
	case SevCrit:
		return 5
	case SevHigh:
		return 4
	case SevMed:
		return 3
	case SevLow:
		return 2
	case SevInfo:
		return 1
	}
	return 0
}

func sevWeight(s string) int {
	switch strings.ToUpper(s) {
	case SevCrit:
		return 40
	case SevHigh:
		return 25
	case SevMed:
		return 12
	case SevLow:
		return 5
	case SevInfo:
		return 1
	}
	return 0
}

func sevColor(s string) string {
	switch strings.ToUpper(s) {
	case SevCrit:
		return R + B
	case SevHigh:
		return R
	case SevMed:
		return YLW
	case SevLow:
		return W
	default:
		return GY
	}
}

// sarifLevel maps a severity to a SARIF result level.
func sarifLevel(s string) string {
	switch strings.ToUpper(s) {
	case SevCrit, SevHigh:
		return "error"
	case SevMed:
		return "warning"
	case SevLow, SevInfo:
		return "note"
	}
	return "none"
}

// ---------------------------------------------------------------------------
//  Grading — convert a set of findings into an A+..F posture grade.
// ---------------------------------------------------------------------------

func gradeFromFindings(fs []Finding) string {
	risk := 0
	for _, f := range fs {
		risk += sevWeight(f.Severity)
	}
	switch {
	case risk == 0:
		return "A+"
	case risk <= 5:
		return "A"
	case risk <= 15:
		return "B"
	case risk <= 30:
		return "C"
	case risk <= 60:
		return "D"
	case risk <= 100:
		return "E"
	default:
		return "F"
	}
}

func gradeMeaning(g string) string {
	switch g {
	case "A+":
		return "no material exposure detected"
	case "A":
		return "minimal exposure; strong posture"
	case "B":
		return "minor issues; generally solid"
	case "C":
		return "moderate exposure; remediation advised"
	case "D":
		return "significant exposure; prioritize fixes"
	case "E":
		return "severe exposure; urgent remediation"
	default:
		return "critical exposure; immediate action required"
	}
}

func gradeColor(g string) string {
	switch g {
	case "A+", "A":
		return GRN + B
	case "B", "C":
		return YLW + B
	default:
		return R + B
	}
}

// ---------------------------------------------------------------------------
//  Global flags / state
// ---------------------------------------------------------------------------

var (
	flHelp    = flag.Bool("help", false, "show rich usage help and exit")
	flH       = flag.Bool("h", false, "alias for --help")
	flVersion = flag.Bool("version", false, "print version and exit")
	flHowto   = flag.Bool("howto", false, "show the HOW TO USE screen and exit")
	flDemo    = flag.Bool("demo", false, "run a self-contained synthetic demonstration")
	flInput   = flag.String("input", "", "path to an input file to analyze")
	flPath    = flag.String("path", "", "alias for --input (subject path/file)")
	flTarget  = flag.String("target", "", "logical target name/label for the report")
	flFormat  = flag.String("format", "txt", "export format: txt|md|json|csv|sarif")
	flOutput  = flag.String("output", "", "write export to FILE instead of stdout")
	flNoColor = flag.Bool("no-color", false, "disable ANSI colors")
	flSelf    = flag.Bool("selftest", false, "run internal self-test and exit")
	flCatalog = flag.Bool("catalog", false, "print the built-in reference catalog and exit")
	flList    = flag.Bool("list", false, "alias for --catalog")
	flQuiet   = flag.Bool("quiet", false, "suppress banner in CLI export mode")
)

var lastResult *AnalysisResult

// ---------------------------------------------------------------------------
//  main
// ---------------------------------------------------------------------------

func main() {
	flag.Usage = usage
	flag.Parse()

	if *flNoColor {
		noColor = true
	}

	switch {
	case *flHelp || *flH:
		usage()
		return
	case *flVersion:
		fmt.Printf("%s %s  (author khaled.s.haddad, owlsec.org)\n", APPNAME, VERSION)
		return
	case *flSelf:
		os.Exit(runSelfTest())
	case *flHowto:
		printHowto()
		return
	case *flCatalog || *flList:
		printCatalog()
		return
	case *flDemo:
		res := domainDemoResult()
		finalize(&res)
		lastResult = &res
		if !*flQuiet {
			renderResult(&res)
		}
		if *flOutput != "" || *flFormat != "txt" {
			emitExport(&res)
		}
		return
	}

	in := firstNonEmpty(*flInput, *flPath)
	if in != "" {
		res, err := domainAnalyze(in)
		if err != nil {
			errMsg("analysis failed: " + err.Error())
			os.Exit(1)
		}
		finalize(&res)
		lastResult = &res
		if !*flQuiet {
			renderResult(&res)
		}
		emitExport(&res)
		return
	}

	// No actionable flags -> interactive TUI.
	interactive()
}

func firstNonEmpty(a ...string) string {
	for _, s := range a {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

// finalize fills defaulted fields (timestamp, tool, grade) on a result.
func finalize(r *AnalysisResult) {
	if r.When == "" {
		r.When = time.Now().Format("2006-01-02 15:04:05 -0700")
	}
	if r.Tool == "" {
		r.Tool = APPNAME
	}
	if r.Stats == nil {
		r.Stats = map[string]string{}
	}
	if r.Grade == "" {
		r.Grade = gradeFromFindings(r.Findings)
	}
	sortFindings(r.Findings)
}

func sortFindings(fs []Finding) {
	sort.SliceStable(fs, func(i, j int) bool {
		if sevRank(fs[i].Severity) != sevRank(fs[j].Severity) {
			return sevRank(fs[i].Severity) > sevRank(fs[j].Severity)
		}
		return fs[i].ID < fs[j].ID
	})
}

// ---------------------------------------------------------------------------
//  Usage / help
// ---------------------------------------------------------------------------

func usage() {
	if !*flQuiet {
		banner()
	}
	fmt.Println("  " + col("USAGE", R+B))
	fmt.Printf("    %s [flags]\n", strings.ToLower(APPNAME))
	fmt.Println()
	fmt.Println("  " + col("DESCRIPTION", R+B))
	for _, ln := range wrap(domainSummary(), 68) {
		fmt.Println("    " + ln)
	}
	fmt.Println()
	fmt.Println("  " + col("CORE FLAGS", R+B))
	helpRow("-h, --help", "show this rich usage help")
	helpRow("--version", "print version and exit")
	helpRow("--howto", "show the HOW TO USE screen (legal, steps, nav)")
	helpRow("--demo", "run a self-contained synthetic demonstration")
	helpRow("--catalog, --list", "print the built-in reference catalog")
	helpRow("--input FILE", "analyze FILE (also --path)")
	helpRow("--target NAME", "label the engagement/target in the report")
	helpRow("--format FMT", "export format: txt|md|json|csv|sarif")
	helpRow("--output FILE", "write export to FILE (default stdout)")
	helpRow("--no-color", "disable ANSI colors")
	helpRow("--quiet", "suppress banner in export mode")
	fmt.Println()
	fmt.Println("  " + col("EXAMPLES", R+B))
	fmt.Printf("    %s                      # interactive red/black TUI\n", strings.ToLower(APPNAME))
	fmt.Printf("    %s --demo                # synthetic walkthrough\n", strings.ToLower(APPNAME))
	fmt.Printf("    %s --howto               # how-to-use + legal warning\n", strings.ToLower(APPNAME))
	fmt.Printf("    %s --input data --format sarif --output out.sarif\n", strings.ToLower(APPNAME))
	fmt.Println()
	fmt.Println("  " + col("MITRE ATT&CK", R+B))
	fmt.Println("    primary: " + MITRE_PRIMARY + "  (" + mitreName(MITRE_PRIMARY) + ")")
	fmt.Println("    " + strings.Join(MITRE_TAGS, ", "))
	fmt.Println()
	line(64)
	fmt.Println("  " + col("Authorized testing only. See --howto for the legal warning.", GY))
}

func helpRow(a, b string) {
	fmt.Printf("    %s %s\n", col(pad(a, 22), W), col(b, GY))
}

// ---------------------------------------------------------------------------
//  HOW TO USE screen  (MANDATORY — menu item + --howto flag)
// ---------------------------------------------------------------------------

func howtoSections() []Section {
	var secs []Section
	secs = append(secs, Section{
		Title: "LEGAL & ETHICAL WARNING",
		Lines: []string{
			APPNAME + " is an offensive-security tool for AUTHORIZED testing only.",
			"Use it solely on systems and data you own or have EXPLICIT, WRITTEN",
			"permission to assess. Unauthorized access, interception, or attack is",
			"illegal in most jurisdictions (e.g. CFAA, Computer Misuse Act, GDPR).",
			"This build performs OFFLINE analysis, parsing, planning and lab",
			"generation only; it does not attack third parties. You are solely",
			"responsible for staying within the scope of your engagement and law.",
			"Keep all artifacts, hashes and credentials in your authorized vault.",
		},
	})
	steps := domainHowtoSteps()
	secs = append(secs, Section{Title: "HOW TO USE", Lines: numbered(steps)})
	secs = append(secs, Section{
		Title: "NAVIGATION",
		Lines: []string{
			"Launch with no arguments for the interactive red/black TUI.",
			"At each menu type the [NN] number of an option and press Enter.",
			"[00] exits the current menu or the program.",
			"HOW TO USE is always available from the main menu and via --howto.",
			"Every result view can be exported (txt/md/json/csv/sarif).",
			"Use --input FILE for non-interactive (CLI) analysis and export.",
		},
	})
	secs = append(secs, Section{Title: "REQUIREMENTS", Lines: domainRequirements()})
	secs = append(secs, domainHowtoNotes()...)
	secs = append(secs, Section{
		Title: "EXPORT & REPORTING",
		Lines: []string{
			"Formats: txt (console report), md (Markdown), json (machine),",
			"csv (spreadsheet), sarif (SARIF 2.1.0 for CI/code-scanning).",
			"Findings carry MITRE ATT&CK technique IDs and an A+..F grade.",
			"Example: --input data --format sarif --output report.sarif",
		},
	})
	return secs
}

func printHowto() {
	clear()
	banner()
	fmt.Println("  " + col("⬡  HOW TO USE — "+APPNAME, B+CY))
	fmt.Println()
	renderSections(howtoSections())
}

func renderSections(secs []Section) {
	for _, s := range secs {
		fmt.Println("  " + col("⬡ "+s.Title, B+CY))
		line(64)
		for _, ln := range s.Lines {
			fmt.Println("   " + col(ln, W))
		}
		fmt.Println()
	}
}

func numbered(items []string) []string {
	out := make([]string, 0, len(items))
	for i, it := range items {
		out = append(out, fmt.Sprintf("%2d. %s", i+1, it))
	}
	return out
}

// ---------------------------------------------------------------------------
//  Catalog printing (CLI)
// ---------------------------------------------------------------------------

func printCatalog() {
	if !*flQuiet {
		clear()
		banner()
	}
	fmt.Println("  " + col("⬡  REFERENCE CATALOG — "+APPNAME, B+CY))
	fmt.Println()
	renderSections(domainCatalogSections())
}

// ---------------------------------------------------------------------------
//  Result rendering (console)
// ---------------------------------------------------------------------------

func renderResult(r *AnalysisResult) {
	clear()
	banner()
	fmt.Println("  " + col("⬡  RESULT — "+r.Tool, B+CY))
	fmt.Println("  " + col("target : ", DIM) + col(nz(r.Target, "(none)"), W))
	fmt.Println("  " + col("time   : ", DIM) + col(r.When, W))
	gc := gradeColor(r.Grade)
	fmt.Println("  " + col("grade  : ", DIM) + col(r.Grade, gc) + col("  ("+gradeMeaning(r.Grade)+")", GY))
	fmt.Println()
	if len(r.Stats) > 0 {
		fmt.Println("  " + col("⬡ STATISTICS", B+CY))
		keys := make([]string, 0, len(r.Stats))
		for k := range r.Stats {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("   %s %s\n", col(pad(k+":", 22), GY), col(r.Stats[k], W))
		}
		fmt.Println()
	}
	renderSevHistogram(r.Findings)
	if len(r.Findings) == 0 {
		okMsg("no findings")
	} else {
		fmt.Println("  " + col(fmt.Sprintf("⬡ FINDINGS (%d)", len(r.Findings)), B+CY))
		line(64)
		for _, f := range r.Findings {
			renderFinding(f)
		}
	}
	if tm := topMitre(r.Findings, 6); len(tm) > 0 {
		fmt.Println()
		fmt.Println("  " + col("⬡ TOP MITRE ATT&CK", B+CY))
		for _, m := range tm {
			info(m)
		}
	}
	if len(r.Notes) > 0 {
		fmt.Println()
		fmt.Println("  " + col("⬡ NOTES", B+CY))
		for _, n := range r.Notes {
			info(n)
		}
	}
}

func renderFinding(f Finding) {
	tag := col("["+f.Severity+"]", sevColor(f.Severity))
	fmt.Printf("  %s %s %s\n", tag, col(f.ID, DR), col(f.Title, W+B))
	if f.Category != "" {
		fmt.Println("      " + col("category : ", GY) + col(f.Category, W))
	}
	for _, ln := range wrap(f.Detail, 60) {
		fmt.Println("      " + col(ln, W))
	}
	if f.Evidence != "" {
		for _, ln := range wrap("evidence: "+f.Evidence, 60) {
			fmt.Println("      " + col(ln, GY))
		}
	}
	if f.Remediation != "" {
		for _, ln := range wrap("fix: "+f.Remediation, 60) {
			fmt.Println("      " + col(ln, GRN))
		}
	}
	if len(f.Mitre) > 0 {
		fmt.Println("      " + col("mitre: "+strings.Join(f.Mitre, ", "), R))
	}
	fmt.Println()
}

// ---------------------------------------------------------------------------
//  Export engine — TXT / MD / JSON / CSV / SARIF 2.1.0
// ---------------------------------------------------------------------------

func emitExport(r *AnalysisResult) {
	format := strings.ToLower(strings.TrimSpace(*flFormat))
	var data string
	switch format {
	case "", "txt":
		data = exportTXT(r)
	case "md", "markdown":
		data = exportMD(r)
	case "json":
		data = exportJSON(r)
	case "csv":
		data = exportCSV(r)
	case "sarif":
		data = exportSARIF(r)
	default:
		errMsg("unknown format: " + format + " (use txt|md|json|csv|sarif)")
		return
	}
	if *flOutput == "" {
		if format != "txt" {
			fmt.Println(data)
		}
		return
	}
	if err := os.WriteFile(*flOutput, []byte(data), 0644); err != nil {
		errMsg("write failed: " + err.Error())
		return
	}
	okMsg(fmt.Sprintf("wrote %s (%d bytes, %s)", *flOutput, len(data), format))
}

func exportTXT(r *AnalysisResult) string {
	var b strings.Builder
	b.WriteString("OWL " + APPNAME + " — " + TAGLINE + "\n")
	b.WriteString("author khaled.s.haddad  owlsec.org\n")
	b.WriteString(strings.Repeat("=", 64) + "\n")
	b.WriteString("target : " + nz(r.Target, "(none)") + "\n")
	b.WriteString("time   : " + r.When + "\n")
	b.WriteString("grade  : " + r.Grade + " (" + gradeMeaning(r.Grade) + ")\n")
	b.WriteString("mitre  : " + MITRE_PRIMARY + " " + strings.Join(MITRE_TAGS, " ") + "\n\n")
	if len(r.Stats) > 0 {
		b.WriteString("STATISTICS\n" + strings.Repeat("-", 64) + "\n")
		for _, k := range sortedKeys(r.Stats) {
			b.WriteString("  " + pad(k+":", 24) + r.Stats[k] + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("FINDINGS (%d)\n", len(r.Findings)))
	b.WriteString(strings.Repeat("-", 64) + "\n")
	for _, f := range r.Findings {
		b.WriteString(fmt.Sprintf("[%s] %s  %s\n", f.Severity, f.ID, f.Title))
		if f.Category != "" {
			b.WriteString("    category   : " + f.Category + "\n")
		}
		if f.Detail != "" {
			b.WriteString("    detail     : " + f.Detail + "\n")
		}
		if f.Evidence != "" {
			b.WriteString("    evidence   : " + f.Evidence + "\n")
		}
		if f.Remediation != "" {
			b.WriteString("    remediation: " + f.Remediation + "\n")
		}
		if len(f.Mitre) > 0 {
			b.WriteString("    mitre      : " + strings.Join(f.Mitre, ", ") + "\n")
		}
		b.WriteString("    score      : " + strconv.Itoa(f.Score) + "\n\n")
	}
	for _, n := range r.Notes {
		b.WriteString("note: " + n + "\n")
	}
	return b.String()
}

func exportMD(r *AnalysisResult) string {
	var b strings.Builder
	b.WriteString("# OWL " + APPNAME + " — " + TAGLINE + "\n\n")
	b.WriteString("*author khaled.s.haddad · owlsec.org*\n\n")
	b.WriteString("| field | value |\n|---|---|\n")
	b.WriteString("| target | " + mdEsc(nz(r.Target, "(none)")) + " |\n")
	b.WriteString("| time | " + mdEsc(r.When) + " |\n")
	b.WriteString("| grade | **" + r.Grade + "** (" + gradeMeaning(r.Grade) + ") |\n")
	b.WriteString("| mitre | " + MITRE_PRIMARY + " " + strings.Join(MITRE_TAGS, " ") + " |\n\n")
	if len(r.Stats) > 0 {
		b.WriteString("## Statistics\n\n| metric | value |\n|---|---|\n")
		for _, k := range sortedKeys(r.Stats) {
			b.WriteString("| " + mdEsc(k) + " | " + mdEsc(r.Stats[k]) + " |\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(fmt.Sprintf("## Findings (%d)\n\n", len(r.Findings)))
	for _, f := range r.Findings {
		b.WriteString("### [" + f.Severity + "] " + mdEsc(f.Title) + " `" + f.ID + "`\n\n")
		if f.Category != "" {
			b.WriteString("- **category:** " + mdEsc(f.Category) + "\n")
		}
		if f.Detail != "" {
			b.WriteString("- **detail:** " + mdEsc(f.Detail) + "\n")
		}
		if f.Evidence != "" {
			b.WriteString("- **evidence:** `" + f.Evidence + "`\n")
		}
		if f.Remediation != "" {
			b.WriteString("- **remediation:** " + mdEsc(f.Remediation) + "\n")
		}
		if len(f.Mitre) > 0 {
			b.WriteString("- **mitre:** " + strings.Join(f.Mitre, ", ") + "\n")
		}
		b.WriteString("- **score:** " + strconv.Itoa(f.Score) + "\n\n")
	}
	return b.String()
}

func exportJSON(r *AnalysisResult) string {
	payload := map[string]interface{}{
		"tool":         APPNAME,
		"tagline":      TAGLINE,
		"version":      VERSION,
		"author":       "khaled.s.haddad",
		"website":      "owlsec.org",
		"mitrePrimary": MITRE_PRIMARY,
		"mitreTags":    MITRE_TAGS,
		"result":       r,
	}
	buf, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "{\"error\":\"" + err.Error() + "\"}"
	}
	return string(buf)
}

func exportCSV(r *AnalysisResult) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write([]string{"id", "severity", "category", "title", "detail", "evidence", "remediation", "mitre", "score"})
	for _, f := range r.Findings {
		w.Write([]string{
			f.ID, f.Severity, f.Category, f.Title, f.Detail, f.Evidence,
			f.Remediation, strings.Join(f.Mitre, " "), strconv.Itoa(f.Score),
		})
	}
	w.Flush()
	return buf.String()
}

func exportSARIF(r *AnalysisResult) string {
	rules := map[string]bool{}
	var ruleList []map[string]interface{}
	var results []map[string]interface{}
	for _, f := range r.Findings {
		if !rules[f.ID] {
			rules[f.ID] = true
			ruleList = append(ruleList, map[string]interface{}{
				"id":               f.ID,
				"name":             f.Title,
				"shortDescription": map[string]string{"text": f.Title},
				"fullDescription":  map[string]string{"text": nz(f.Detail, f.Title)},
				"helpUri":          "https://owlsec.org",
				"properties": map[string]interface{}{
					"category": f.Category,
					"mitre":    f.Mitre,
				},
			})
		}
		results = append(results, map[string]interface{}{
			"ruleId": f.ID,
			"level":  sarifLevel(f.Severity),
			"message": map[string]string{
				"text": f.Title + ". " + f.Detail + nzPrefix(" Evidence: ", f.Evidence),
			},
			"properties": map[string]interface{}{
				"severity": f.Severity,
				"score":    f.Score,
				"mitre":    f.Mitre,
			},
			"locations": []map[string]interface{}{
				{
					"physicalLocation": map[string]interface{}{
						"artifactLocation": map[string]string{"uri": nz(r.Target, APPNAME)},
					},
				},
			},
		})
	}
	sarif := map[string]interface{}{
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"version": "2.1.0",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name":           APPNAME,
						"fullName":       "OWL " + APPNAME + " — " + TAGLINE,
						"version":        VERSION,
						"informationUri": "https://owlsec.org",
						"organization":   "owlsec.org / khaled.s.haddad",
						"rules":          ruleList,
					},
				},
				"results": results,
				"properties": map[string]interface{}{
					"grade":        r.Grade,
					"mitrePrimary": MITRE_PRIMARY,
					"target":       r.Target,
				},
			},
		},
	}
	buf, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return "{\"error\":\"" + err.Error() + "\"}"
	}
	return string(buf)
}

// ---------------------------------------------------------------------------
//  Interactive TUI
// ---------------------------------------------------------------------------

func interactive() {
	for {
		items := []string{
			"Run DEMO (synthetic walkthrough)",
			"Analyze an input file",
			"Reference catalog",
		}
		dm := domainMenuItems()
		for _, m := range dm {
			items = append(items, m.Label)
		}
		items = append(items,
			"View last result",
			"Export last result",
			"HOW TO USE",
			"About / MITRE mapping",
		)
		choice := menu("MAIN MENU — "+TAGLINE, items)
		switch strings.ToLower(choice) {
		case "h", "help", "?":
			tuiHelp()
		case "0", "00", "q", "exit":
			clear()
			okMsg("Stay authorized. Goodbye.")
			return
		case "1":
			res := domainDemoResult()
			finalize(&res)
			lastResult = &res
			renderResult(&res)
			pause()
		case "2":
			actionAnalyze()
		case "3":
			clear()
			banner()
			fmt.Println("  " + col("⬡  REFERENCE CATALOG", B+CY))
			fmt.Println()
			renderSections(domainCatalogSections())
			pause()
		default:
			handleDynamicChoice(choice, dm)
		}
	}
}

// handleDynamicChoice resolves menu numbers for domain items and trailing items.
func handleDynamicChoice(choice string, dm []MenuItem) {
	n, err := strconv.Atoi(choice)
	if err != nil {
		errMsg("invalid selection")
		pause()
		return
	}
	base := 3 // first three fixed items
	if n > base && n <= base+len(dm) {
		dm[n-base-1].Action()
		return
	}
	idx := n - base - len(dm)
	switch idx {
	case 1: // view last
		if lastResult == nil {
			warnMsg("no result yet — run DEMO or analyze a file first")
			pause()
			return
		}
		renderResult(lastResult)
		pause()
	case 2: // export last
		actionExport()
	case 3: // howto
		printHowto()
		pause()
	case 4: // about
		actionAbout()
	default:
		errMsg("invalid selection")
		pause()
	}
}

func actionAnalyze() {
	clear()
	banner()
	fmt.Println("  " + col("⬡  ANALYZE INPUT", B+CY))
	fmt.Println()
	p := prompt("path to input file")
	if p == "" {
		warnMsg("cancelled")
		pause()
		return
	}
	res, err := domainAnalyze(p)
	if err != nil {
		errMsg(err.Error())
		pause()
		return
	}
	finalize(&res)
	lastResult = &res
	renderResult(&res)
	pause()
}

func actionExport() {
	if lastResult == nil {
		warnMsg("no result yet — run DEMO or analyze a file first")
		pause()
		return
	}
	clear()
	banner()
	fmt.Println("  " + col("⬡  EXPORT LAST RESULT", B+CY))
	fmt.Println()
	f := prompt("format txt|md|json|csv|sarif")
	if f == "" {
		f = "txt"
	}
	o := prompt("output file (blank = screen)")
	*flFormat = f
	*flOutput = o
	if o == "" {
		var data string
		switch strings.ToLower(f) {
		case "md":
			data = exportMD(lastResult)
		case "json":
			data = exportJSON(lastResult)
		case "csv":
			data = exportCSV(lastResult)
		case "sarif":
			data = exportSARIF(lastResult)
		default:
			data = exportTXT(lastResult)
		}
		fmt.Println()
		fmt.Println(data)
	} else {
		emitExport(lastResult)
	}
	pause()
}

func actionAbout() {
	clear()
	banner()
	fmt.Println("  " + col("⬡  ABOUT", B+CY))
	fmt.Println()
	for _, ln := range wrap(domainSummary(), 64) {
		fmt.Println("   " + col(ln, W))
	}
	fmt.Println()
	fmt.Println("  " + col("⬡ MITRE ATT&CK", B+CY))
	fmt.Println("   primary : " + col(MITRE_PRIMARY+"  "+mitreName(MITRE_PRIMARY), W))
	for _, t := range MITRE_TAGS {
		fmt.Println("   " + col(pad(t, 12), R) + col(mitreName(t), GY))
	}
	fmt.Println()
	fmt.Println("  " + col("author khaled.s.haddad  ·  owlsec.org  ·  "+VERSION, GY))
	pause()
}

// ---------------------------------------------------------------------------
//  MITRE reference table (subset relevant to Batch-3 offensive tooling)
// ---------------------------------------------------------------------------

var mitreTable = map[string]string{
	"T1027":     "Obfuscated Files or Information",
	"T1027.009": "Embedded Payloads",
	"T1027.010": "Command Obfuscation",
	"T1059":     "Command and Scripting Interpreter",
	"T1059.001": "PowerShell",
	"T1059.003": "Windows Command Shell",
	"T1059.004": "Unix Shell",
	"T1059.005": "Visual Basic",
	"T1059.006": "Python",
	"T1059.007": "JavaScript",
	"T1105":     "Ingress Tool Transfer",
	"T1090":     "Proxy",
	"T1087":     "Account Discovery",
	"T1087.002": "Domain Account Discovery",
	"T1110":     "Brute Force",
	"T1110.003": "Password Spraying",
	"T1134":     "Access Token Manipulation",
	"T1134.001": "Token Impersonation/Theft",
	"T1134.002": "Create Process with Token",
	"T1134.005": "SID-History Injection",
	"T1135":     "Network Share Discovery",
	"T1548":     "Abuse Elevation Control Mechanism",
	"T1548.002": "Bypass User Account Control",
	"T1548.003": "Sudo and Sudo Caching",
	"T1558":     "Steal or Forge Kerberos Tickets",
	"T1558.003": "Kerberoasting",
	"T1558.004": "AS-REP Roasting",
	"T1611":     "Escape to Host",
	"T1210":     "Exploitation of Remote Services",
	"T1021":     "Remote Services",
	"T1021.002": "SMB/Windows Admin Shares",
	"T1069":     "Permission Groups Discovery",
	"T1482":     "Domain Trust Discovery",
	"T1078":     "Valid Accounts",
	"T1003":     "OS Credential Dumping",
	"TA0001":    "Initial Access",
	"TA0002":    "Execution",
	"TA0003":    "Persistence",
	"TA0004":    "Privilege Escalation",
	"TA0005":    "Defense Evasion",
	"TA0006":    "Credential Access",
	"TA0007":    "Discovery",
	"TA0008":    "Lateral Movement",
	"TA0010":    "Exfiltration",
	"TA0011":    "Command and Control",
}

func mitreName(id string) string {
	if n, ok := mitreTable[id]; ok {
		return n
	}
	return "technique"
}

// ---------------------------------------------------------------------------
//  Domain utilities shared by every tool (entropy, hashing, encoders, text)
// ---------------------------------------------------------------------------

// shannonEntropy returns bits/byte (0..8) for a byte slice.
func shannonEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	var counts [256]int
	for _, b := range data {
		counts[b]++
	}
	n := float64(len(data))
	ent := 0.0
	for _, c := range counts {
		if c == 0 {
			continue
		}
		p := float64(c) / n
		ent -= p * math.Log2(p)
	}
	return ent
}

func entropyVerdict(e float64) string {
	switch {
	case e >= 7.5:
		return "very high (encrypted/packed/compressed)"
	case e >= 6.5:
		return "high (encoded/compressed)"
	case e >= 5.0:
		return "moderate (mixed/base64-ish)"
	case e >= 3.0:
		return "low (plain text/code)"
	default:
		return "very low (repetitive/structured)"
	}
}

func b64Encode(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
func b64Decode(s string) (string, error) {
	d, err := base64.StdEncoding.DecodeString(strings.TrimSpace(s))
	return string(d), err
}
func hexEncode(s string) string { return hex.EncodeToString([]byte(s)) }

func gzipB64(s string) string {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write([]byte(s))
	w.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func md5Hex(s string) string    { h := md5.Sum([]byte(s)); return hex.EncodeToString(h[:]) }
func sha1Hex(s string) string   { h := sha1.Sum([]byte(s)); return hex.EncodeToString(h[:]) }
func sha256Hex(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

// utf16leB64 encodes a string as UTF-16LE then base64 (PowerShell -EncodedCommand).
func utf16leB64(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		if r < 0x10000 {
			buf.WriteByte(byte(r))
			buf.WriteByte(byte(r >> 8))
		} else {
			r -= 0x10000
			hi := 0xD800 + (r >> 10)
			lo := 0xDC00 + (r & 0x3FF)
			buf.WriteByte(byte(hi))
			buf.WriteByte(byte(hi >> 8))
			buf.WriteByte(byte(lo))
			buf.WriteByte(byte(lo >> 8))
		}
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func xorBytes(data []byte, key []byte) []byte {
	if len(key) == 0 {
		return data
	}
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ key[i%len(key)]
	}
	return out
}

// classifyCharset reports character-class composition of a string.
func classifyCharset(s string) (lower, upper, digit, space, punct int) {
	for _, r := range s {
		switch {
		case unicode.IsLower(r):
			lower++
		case unicode.IsUpper(r):
			upper++
		case unicode.IsDigit(r):
			digit++
		case unicode.IsSpace(r):
			space++
		default:
			punct++
		}
	}
	return
}

func randHex(n int) string {
	const h = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = h[rand.Intn(16)]
	}
	return string(b)
}

func randID(prefix string) string {
	return prefix + "-" + randHex(6)
}

// ---------------------------------------------------------------------------
//  Small text/formatting helpers
// ---------------------------------------------------------------------------

func nz(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func nzPrefix(prefix, s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	return prefix + s
}

func mdEsc(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func sortedKeys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func wrap(s string, width int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
		} else if len(cur)+1+len(w) <= width {
			cur += " " + w
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func readFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 8*1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func readFileString(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func baseName(p string) string { return filepath.Base(p) }

func mustInt(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func reFind(pattern, s string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re.FindStringSubmatch(s)
}

func reAll(pattern, s string) [][]string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re.FindAllStringSubmatch(s, -1)
}

func contains(hay, needle string) bool {
	return strings.Contains(strings.ToLower(hay), strings.ToLower(needle))
}

func titleOf(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func mkFinding(id, title, sev, cat, detail string, mitre []string) Finding {
	return Finding{
		ID:       id,
		Title:    title,
		Severity: sev,
		Category: cat,
		Detail:   detail,
		Mitre:    mitre,
		Score:    sevWeight(sev),
	}
}

// ---------------------------------------------------------------------------
//  Binary read helpers (little/big endian) for protocol parsers.
// ---------------------------------------------------------------------------

func u16LE(b []byte, off int) uint16 {
	if off < 0 || off+2 > len(b) {
		return 0
	}
	return binary.LittleEndian.Uint16(b[off : off+2])
}

func u32LE(b []byte, off int) uint32 {
	if off < 0 || off+4 > len(b) {
		return 0
	}
	return binary.LittleEndian.Uint32(b[off : off+4])
}

func u16BE(b []byte, off int) uint16 {
	if off < 0 || off+2 > len(b) {
		return 0
	}
	return binary.BigEndian.Uint16(b[off : off+2])
}

func u32BE(b []byte, off int) uint32 {
	if off < 0 || off+4 > len(b) {
		return 0
	}
	return binary.BigEndian.Uint32(b[off : off+4])
}

func hexDump(b []byte, max int) string {
	if len(b) > max {
		b = b[:max]
	}
	var sb strings.Builder
	for i := 0; i < len(b); i += 16 {
		end := i + 16
		if end > len(b) {
			end = len(b)
		}
		sb.WriteString(fmt.Sprintf("%08x  ", i))
		for j := i; j < i+16; j++ {
			if j < end {
				sb.WriteString(fmt.Sprintf("%02x ", b[j]))
			} else {
				sb.WriteString("   ")
			}
		}
		sb.WriteString(" |")
		for j := i; j < end; j++ {
			c := b[j]
			if c >= 32 && c < 127 {
				sb.WriteByte(c)
			} else {
				sb.WriteByte('.')
			}
		}
		sb.WriteString("|\n")
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
//  Severity analytics (used by result rendering and stats)
// ---------------------------------------------------------------------------

func sevHistogram(fs []Finding) map[string]int {
	h := map[string]int{SevCrit: 0, SevHigh: 0, SevMed: 0, SevLow: 0, SevInfo: 0}
	for _, f := range fs {
		k := strings.ToUpper(f.Severity)
		if _, ok := h[k]; ok {
			h[k]++
		} else {
			h[SevInfo]++
		}
	}
	return h
}

func renderSevHistogram(fs []Finding) {
	if len(fs) == 0 {
		return
	}
	h := sevHistogram(fs)
	order := []string{SevCrit, SevHigh, SevMed, SevLow, SevInfo}
	max := 1
	for _, k := range order {
		if h[k] > max {
			max = h[k]
		}
	}
	fmt.Println("  " + col("⬡ SEVERITY BREAKDOWN", B+CY))
	for _, k := range order {
		n := h[k]
		barLen := (n * 30) / max
		if n > 0 && barLen == 0 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen)
		fmt.Printf("   %s %s %s\n", col(pad(k, 9), sevColor(k)), col(pad(strconv.Itoa(n), 3), W), col(bar, sevColor(k)))
	}
	fmt.Println()
}

// mitreTally counts technique IDs across findings.
func mitreTally(fs []Finding) map[string]int {
	t := map[string]int{}
	for _, f := range fs {
		for _, m := range f.Mitre {
			t[m]++
		}
	}
	return t
}

func topMitre(fs []Finding, n int) []string {
	t := mitreTally(fs)
	type kv struct {
		k string
		v int
	}
	var arr []kv
	for k, v := range t {
		arr = append(arr, kv{k, v})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].v != arr[j].v {
			return arr[i].v > arr[j].v
		}
		return arr[i].k < arr[j].k
	})
	var out []string
	for i, e := range arr {
		if i >= n {
			break
		}
		out = append(out, fmt.Sprintf("%s x%d (%s)", e.k, e.v, mitreName(e.k)))
	}
	return out
}

// ---------------------------------------------------------------------------
//  Generic parsing & text helpers (shared by domain engines)
// ---------------------------------------------------------------------------

func parseKVLine(line, sep string) (string, string, bool) {
	i := strings.Index(line, sep)
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+len(sep):]), true
}

// tokenizeArgs splits a command line honoring simple single/double quotes.
func tokenizeArgs(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	inTok := false
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
			inTok = true
		case r == '\'' || r == '"':
			quote = r
			inTok = true
		case r == ' ' || r == '\t':
			if inTok {
				out = append(out, cur.String())
				cur.Reset()
				inTok = false
			}
		default:
			cur.WriteRune(r)
			inTok = true
		}
	}
	if inTok {
		out = append(out, cur.String())
	}
	return out
}

func splitList(s string) []string {
	f := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t'
	})
	var out []string
	for _, x := range f {
		x = strings.TrimSpace(x)
		if x != "" {
			out = append(out, x)
		}
	}
	return out
}

func atoiOr(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func pct(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) * 100.0 / float64(total)
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func isHexString(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func rot13(s string) string {
	b := []byte(s)
	for i, c := range b {
		switch {
		case c >= 'a' && c <= 'z':
			b[i] = 'a' + (c-'a'+13)%26
		case c >= 'A' && c <= 'Z':
			b[i] = 'A' + (c-'A'+13)%26
		}
	}
	return string(b)
}

const b32alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func base32Encode(data []byte) string {
	var sb strings.Builder
	var buf uint64
	var bits int
	for _, b := range data {
		buf = (buf << 8) | uint64(b)
		bits += 8
		for bits >= 5 {
			bits -= 5
			sb.WriteByte(b32alpha[(buf>>uint(bits))&0x1f])
		}
	}
	if bits > 0 {
		sb.WriteByte(b32alpha[(buf<<uint(5-bits))&0x1f])
	}
	for sb.Len()%8 != 0 {
		sb.WriteByte('=')
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
//  Generic directed graph with weighted BFS/Dijkstra shortest path.
//  Used by attack-path / reachability domains.
// ---------------------------------------------------------------------------

type GraphEdge struct {
	To     string
	Label  string
	Weight int
}

type Graph struct {
	Adj   map[string][]GraphEdge
	Nodes map[string]bool
}

func NewGraph() *Graph {
	return &Graph{Adj: map[string][]GraphEdge{}, Nodes: map[string]bool{}}
}

func (g *Graph) AddNode(n string) {
	if n == "" {
		return
	}
	g.Nodes[n] = true
	if _, ok := g.Adj[n]; !ok {
		g.Adj[n] = nil
	}
}

func (g *Graph) AddEdge(from, to, label string, weight int) {
	g.AddNode(from)
	g.AddNode(to)
	g.Adj[from] = append(g.Adj[from], GraphEdge{To: to, Label: label, Weight: weight})
}

// ShortestPath returns the min-weight path and total cost, or nil if none.
func (g *Graph) ShortestPath(src, dst string) ([]string, []string, int) {
	const inf = int(^uint(0) >> 1)
	dist := map[string]int{}
	prev := map[string]string{}
	prevLabel := map[string]string{}
	for n := range g.Nodes {
		dist[n] = inf
	}
	if _, ok := g.Nodes[src]; !ok {
		return nil, nil, -1
	}
	dist[src] = 0
	visited := map[string]bool{}
	for {
		u := ""
		best := inf
		for n, d := range dist {
			if !visited[n] && d < best {
				best = d
				u = n
			}
		}
		if u == "" {
			break
		}
		if u == dst {
			break
		}
		visited[u] = true
		for _, e := range g.Adj[u] {
			nd := dist[u] + e.Weight
			if nd < dist[e.To] {
				dist[e.To] = nd
				prev[e.To] = u
				prevLabel[e.To] = e.Label
			}
		}
	}
	if dist[dst] == inf {
		return nil, nil, -1
	}
	var path []string
	var labels []string
	cur := dst
	for cur != src {
		path = append([]string{cur}, path...)
		labels = append([]string{prevLabel[cur]}, labels...)
		p, ok := prev[cur]
		if !ok {
			return nil, nil, -1
		}
		cur = p
	}
	path = append([]string{src}, path...)
	return path, labels, dist[dst]
}

// ReachableFrom returns all nodes reachable from src (BFS).
func (g *Graph) ReachableFrom(src string) []string {
	seen := map[string]bool{src: true}
	queue := []string{src}
	var out []string
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, e := range g.Adj[u] {
			if !seen[e.To] {
				seen[e.To] = true
				out = append(out, e.To)
				queue = append(queue, e.To)
			}
		}
	}
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------
//  SID / RID helpers (Windows / AD domains)
// ---------------------------------------------------------------------------

// parseSIDString validates and splits an SDDL-style SID string.
func parseSIDString(s string) (parts []string, ok bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "S-") {
		return nil, false
	}
	parts = strings.Split(s, "-")
	if len(parts) < 3 {
		return nil, false
	}
	for i := 1; i < len(parts); i++ {
		if _, err := strconv.ParseUint(parts[i], 10, 64); err != nil {
			return nil, false
		}
	}
	return parts, true
}

func sidRID(s string) (uint64, bool) {
	parts, ok := parseSIDString(s)
	if !ok || len(parts) < 3 {
		return 0, false
	}
	n, err := strconv.ParseUint(parts[len(parts)-1], 10, 64)
	return n, err == nil
}

func wellKnownRID(rid uint64) string {
	switch rid {
	case 500:
		return "Administrator"
	case 501:
		return "Guest"
	case 502:
		return "KRBTGT"
	case 512:
		return "Domain Admins"
	case 513:
		return "Domain Users"
	case 514:
		return "Domain Guests"
	case 515:
		return "Domain Computers"
	case 516:
		return "Domain Controllers"
	case 518:
		return "Schema Admins"
	case 519:
		return "Enterprise Admins"
	case 520:
		return "Group Policy Creator Owners"
	case 544:
		return "Administrators (builtin)"
	case 545:
		return "Users (builtin)"
	case 546:
		return "Guests (builtin)"
	case 547:
		return "Power Users (builtin)"
	case 548:
		return "Account Operators"
	case 549:
		return "Server Operators"
	case 550:
		return "Print Operators"
	case 551:
		return "Backup Operators"
	case 552:
		return "Replicator"
	case 553:
		return "RAS and IAS Servers"
	case 554:
		return "Pre-Windows 2000 Compatible Access"
	case 555:
		return "Remote Desktop Users"
	case 556:
		return "Network Configuration Operators"
	case 557:
		return "Incoming Forest Trust Builders"
	case 558:
		return "Performance Monitor Users"
	case 559:
		return "Performance Log Users"
	case 560:
		return "Windows Authorization Access Group"
	case 561:
		return "Terminal Server License Servers"
	case 562:
		return "Distributed COM Users"
	case 568:
		return "IIS_IUSRS"
	case 569:
		return "Cryptographic Operators"
	case 573:
		return "Event Log Readers"
	case 574:
		return "Certificate Service DCOM Access"
	case 575:
		return "RDS Remote Access Servers"
	case 576:
		return "RDS Endpoint Servers"
	case 577:
		return "RDS Management Servers"
	case 578:
		return "Hyper-V Administrators"
	case 579:
		return "Access Control Assistance Operators"
	case 580:
		return "Remote Management Users"
	}
	return ""
}

func isPrivilegedRID(rid uint64) bool {
	switch rid {
	case 500, 512, 518, 519, 520, 544, 548, 549, 550, 551, 578:
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
//  Hash format detectors (NTLM / Kerberos / Unix)
// ---------------------------------------------------------------------------

type hashKind struct {
	Name   string
	Mode   string // hashcat mode
	Sample string
	Re     *regexp.Regexp
}

var hashKinds = []hashKind{
	{"NTLM", "1000", "aad3b435b51404eeaad3b435b51404ee:31d6cfe0d16ae931b73c59d7e0c089c0", regexp.MustCompile(`(?i)^[0-9a-f]{32}:[0-9a-f]{32}$`)},
	{"NTLM-hash-only", "1000", "31d6cfe0d16ae931b73c59d7e0c089c0", regexp.MustCompile(`(?i)^[0-9a-f]{32}$`)},
	{"NetNTLMv1", "5500", "user::domain:LMRESP:NTRESP:challenge", regexp.MustCompile(`^[^:]+:[^:]*:[^:]*:[0-9a-fA-F]{48}:[0-9a-fA-F]{48}:[0-9a-fA-F]+$`)},
	{"NetNTLMv2", "5600", "user::domain:challenge:response:blob", regexp.MustCompile(`^[^:]+:[^:]*:[^:]*:[0-9a-fA-F]{16}:[0-9a-fA-F]{32}:.+`)},
	{"Kerberos-23-TGS", "13100", "$krb5tgs$23$*user$realm$spn*$checksum$edata", regexp.MustCompile(`(?i)^\$krb5tgs\$23\$`)},
	{"Kerberos-18-TGS", "19700", "$krb5tgs$18$user$realm$*spn*$checksum$edata", regexp.MustCompile(`(?i)^\$krb5tgs\$18\$`)},
	{"Kerberos-17-TGS", "19600", "$krb5tgs$17$user$realm$*spn*$checksum$edata", regexp.MustCompile(`(?i)^\$krb5tgs\$17\$`)},
	{"Kerberos-23-ASREP", "18200", "$krb5asrep$23$user@realm$hash", regexp.MustCompile(`(?i)^\$krb5asrep\$23\$`)},
	{"Kerberos-18-ASREP", "19800", "$krb5asrep$18$user@realm$hash", regexp.MustCompile(`(?i)^\$krb5asrep\$18\$`)},
	{"md5crypt", "500", "$1$salt$hash", regexp.MustCompile(`^\$1\$[^$]+\$[./0-9A-Za-z]{22}$`)},
	{"sha256crypt", "7400", "$5$salt$hash", regexp.MustCompile(`^\$5\$(rounds=\d+\$)?[^$]+\$[./0-9A-Za-z]{43}$`)},
	{"sha512crypt", "1800", "$6$salt$hash", regexp.MustCompile(`^\$6\$(rounds=\d+\$)?[^$]+\$[./0-9A-Za-z]{86}$`)},
	{"bcrypt", "3200", "$2y$10$...", regexp.MustCompile(`^\$2[aby]?\$\d{2}\$[./A-Za-z0-9]{53}$`)},
	{"yescrypt", "n/a", "$y$j9T$...", regexp.MustCompile(`^\$y\$`)},
	{"LM", "3000", "aad3b435b51404eeaad3b435b51404ee", regexp.MustCompile(`(?i)^[0-9a-f]{32}$`)},
	{"mscash2", "2100", "$DCC2$10240#user#hash", regexp.MustCompile(`(?i)^\$DCC2\$\d+#.+#[0-9a-f]{32}$`)},
	{"krb5pa-23", "7500", "$krb5pa$23$user$realm$salt$hash", regexp.MustCompile(`(?i)^\$krb5pa\$23\$`)},
	{"krb5pa-18", "19900", "$krb5pa$18$user$realm$salt$hash", regexp.MustCompile(`(?i)^\$krb5pa\$18\$`)},
}

func detectHash(line string) *hashKind {
	line = strings.TrimSpace(line)
	for i := range hashKinds {
		if hashKinds[i].Re.MatchString(line) {
			return &hashKinds[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
//  Minimal LDIF parser (shared by AD/LDAP tools)
// ---------------------------------------------------------------------------

type LDIFEntry struct {
	DN     string
	Attrs  map[string][]string
	Object string // first objectClass
}

func parseLDIF(text string) []LDIFEntry {
	var out []LDIFEntry
	var cur *LDIFEntry
	flush := func() {
		if cur != nil && cur.DN != "" {
			out = append(out, *cur)
		}
		cur = nil
	}
	lines := strings.Split(text, "\n")
	var cont string
	for _, raw := range lines {
		if strings.HasPrefix(raw, " ") && cont != "" {
			cont += strings.TrimPrefix(raw, " ")
			continue
		}
		if cont != "" {
			applyLDIFLine(&cur, cont)
			cont = ""
		}
		if strings.TrimSpace(raw) == "" {
			flush()
			continue
		}
		if strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "version:") {
			continue
		}
		cont = strings.TrimRight(raw, "\r")
	}
	if cont != "" {
		applyLDIFLine(&cur, cont)
	}
	flush()
	return out
}

func applyLDIFLine(cur **LDIFEntry, line string) {
	if strings.HasPrefix(line, "dn:") {
		if *cur == nil {
			*cur = &LDIFEntry{Attrs: map[string][]string{}}
		}
		(*cur).DN = strings.TrimSpace(strings.TrimPrefix(line, "dn:"))
		if strings.HasPrefix((*cur).DN, ":") { // base64 dn
			(*cur).DN = strings.TrimSpace(strings.TrimPrefix((*cur).DN, ":"))
			if d, err := base64.StdEncoding.DecodeString((*cur).DN); err == nil {
				(*cur).DN = string(d)
			}
		}
		return
	}
	if *cur == nil {
		*cur = &LDIFEntry{Attrs: map[string][]string{}}
	}
	i := strings.Index(line, ":")
	if i < 0 {
		return
	}
	attr := strings.TrimSpace(line[:i])
	val := strings.TrimSpace(line[i+1:])
	if strings.HasPrefix(val, ":") { // base64
		val = strings.TrimSpace(strings.TrimPrefix(val, ":"))
		if d, err := base64.StdEncoding.DecodeString(val); err == nil {
			val = string(d)
		}
	} else if strings.HasPrefix(val, "<") {
		// URL reference; keep as-is
	}
	(*cur).Attrs[attr] = append((*cur).Attrs[attr], val)
	if strings.EqualFold(attr, "objectClass") && (*cur).Object == "" {
		(*cur).Object = val
	}
}

func (e LDIFEntry) HasClass(c string) bool {
	for _, v := range e.Attrs["objectClass"] {
		if strings.EqualFold(v, c) {
			return true
		}
	}
	return false
}

func (e LDIFEntry) Attr(name string) string {
	vs := e.Attrs[name]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// ---------------------------------------------------------------------------
//  CSV helpers
// ---------------------------------------------------------------------------

func parseCSVText(text string) (headers []string, rows [][]string, err error) {
	r := csv.NewReader(strings.NewReader(text))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(all) == 0 {
		return nil, nil, nil
	}
	return all[0], all[1:], nil
}

func rowMap(headers, row []string) map[string]string {
	m := map[string]string{}
	for i, h := range headers {
		if i < len(row) {
			m[h] = row[i]
		}
	}
	return m
}

// ---------------------------------------------------------------------------
//  Score helpers used by graders
// ---------------------------------------------------------------------------

func riskBucket(score int) string {
	switch {
	case score >= 100:
		return "CRITICAL"
	case score >= 60:
		return "HIGH"
	case score >= 30:
		return "MEDIUM"
	case score >= 10:
		return "LOW"
	default:
		return "INFO"
	}
}

func combineGrades(gs ...string) string {
	worst := "A+"
	order := map[string]int{"A+": 0, "A": 1, "B": 2, "C": 3, "D": 4, "E": 5, "F": 6}
	for _, g := range gs {
		if order[g] > order[worst] {
			worst = g
		}
	}
	return worst
}

// ============================================================================
//  DOMAIN ENGINE — OUACBypass : UAC Bypass Technique Catalog + Config Scanner
//  Catalogs auto-elevate / mock-dir / COM / registry UAC bypass techniques and
//  scans offline Windows config exports for exposure.  MITRE T1548.002
// ============================================================================

var MITRE_PRIMARY = "T1548.002"
var MITRE_TAGS = []string{"T1548", "T1548.002", "T1546.015", "T1112", "T1218", "TA0004", "TA0005"}

func domainSummary() string {
	return "OUACBypass is an offline catalog and configuration scanner for Windows User " +
		"Account Control bypass techniques. It documents auto-elevate binaries, mock " +
		"directories, COM interface abuses, registry hijacks and environment-variable " +
		"tricks with detection guidance, and scans exported registry / autostart / " +
		"AppInfo policy data for exposure indicators. Analysis only — it does not " +
		"execute bypasses against live hosts."
}

type uacTech struct {
	ID         string
	Name       string
	Method     string // autoelevate | com | registry | env | mockdir | dllhijack | other
	Severity   string
	WinVer     string
	Binary     string
	Summary    string
	Detection  string
	Mitigation string
	Mitre      []string
}

var uacCatalog = []uacTech{
	{"UAC-FODHELPER", "fodhelper.exe Registry Hijack", "registry", SevHigh, "Win10+", "fodhelper.exe",
		"fodhelper.exe is auto-elevated; hijacking HKCU\\Software\\Classes\\ms-settings\\Shell\\Open\\command runs attacker code elevated.",
		"Sysmon/ETW: fodhelper.exe spawning unusual children; registry set on ms-settings\\Shell\\Open\\command.",
		"BePUIAccessApps / EnableLUA; monitor the ms-settings registry key; remove unused auto-elevate binaries.",
		[]string{"T1548.002", "T1546.015"}},
	{"UAC-EVENTVWR", "eventvwr.exe MSC Hijack", "registry", SevHigh, "Win7-Win10", "eventvwr.exe",
		"eventvwr.exe (auto-elevate) launches mmc with eventvwr.msc; hijack HKCU\\Software\\Classes\\mscfile\\shell\\open\\command.",
		"eventvwr.exe → cmd/powershell child; mscfile registry modification.",
		"Patched in later Win10; still audit mscfile association and auto-elevate usage.",
		[]string{"T1548.002", "T1546.015"}},
	{"UAC-SDCLT", "sdclt.exe IsolatedCommand", "registry", SevHigh, string("Win10"), "sdclt.exe",
		"sdclt.exe /KickOffElev reads HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\App Paths / IsolatedCommand.",
		"sdclt with unusual args; App Paths registry writes under HKCU.",
		"Block sdclt from non-admin contexts; monitor App Paths writes.",
		[]string{"T1548.002"}},
	{"UAC-COMPMGMT", "CompMgmtLauncher / MMC", "registry", SevMed, "Win7+", "CompMgmtLauncher.exe",
		"Computer Management launcher is auto-elevated; similar mscfile / COM hijacks apply.",
		"CompMgmtLauncher spawning shells; mscfile association changes.",
		"Constrain auto-elevate; prefer AlwaysNotify UAC.",
		[]string{"T1548.002"}},
	{"UAC-SLUI", "slui.exe File Association", "registry", SevHigh, "Win10", "slui.exe",
		"Windows Activation UI (slui.exe) is auto-elevated; hijack HKCU file-assoc for the launched handler.",
		"slui.exe child processes; HKCU\\Software\\Classes changes near activation UI use.",
		"Patch level + registry monitoring of Classes keys under HKCU.",
		[]string{"T1548.002", "T1546.015"}},
	{"UAC-WSRESET", "WSReset.exe", "registry", SevHigh, "Win10", "WSReset.exe",
		"Store reset tool auto-elevates; hijack ms-windows-store\\Shell\\Open\\command under HKCU.",
		"WSReset.exe → cmd; registry writes to ms-windows-store association.",
		"Monitor that Classes key; restrict AppLocker/WDAC for WSReset if unused.",
		[]string{"T1548.002"}},
	{"UAC-CMSTPLUA", "ICMLuaUtil elevated COM", "com", SevCrit, "Win7+", "dllhost.exe / cmstplua",
		"The CMSTPLUA COM elevation moniker launches ICMLuaUtil in a high-integrity host; ShellExec runs elevated.",
		"dllhost.exe hosting CMSTPLUA CLSID; unusual ShellExec from that host.",
		"UAC Always Notify; remove unused elevation COM classes where possible; EDR on elevation moniker use.",
		[]string{"T1548.002"}},
	{"UAC-COLORDATA", "ColorDataProxy COM", "com", SevHigh, "Win10", "ColorAdapterClient / dllhost",
		"ICMColorData / ColorDataProxy COM interface LaunchApplication method runs elevated.",
		"COM calls to ColorDataProxy from medium IL processes.",
		"Monitor elevation COM interfaces; WDAC policies.",
		[]string{"T1548.002"}},
	{"UAC-EDITIONUPGRADE", "EditionUpgradeManager COM", "com", SevHigh, "Win10", "dllhost.exe",
		"IEditionUpgradeManager::ChangeProductKey / ElevateWindow handles elevation abuse.",
		"COM elevation via EditionUpgradeManager CLSID.",
		"Patch + COM telemetry.",
		[]string{"T1548.002"}},
	{"UAC-MOCKDIR", "Trusted Directory Mock (winsxs)", "mockdir", SevHigh, "Win7+", "various auto-elevate",
		"Create a mock 'C:\\Windows \\System32' (trailing space) or similar; copy an auto-elevate binary + malicious DLL; WinSxS/DLL search loads the attacker DLL elevated.",
		"Directory names with trailing spaces/dots; auto-elevate binary loading DLLs from non-system paths.",
		"Kernel/OS patches for trailing-space paths; monitor creation of lookalike Windows dirs.",
		[]string{"T1548.002", "T1574.001"}},
	{"UAC-ENV-SYSTEMROOT", "SYSTEMROOT Environment Hijack", "env", SevHigh, "Win7-early Win10", "taskmgr / others",
		"Some auto-elevate binaries trusted %SYSTEMROOT%; a user-writable SYSTEMROOT led to elevated DLL/exe load (patched).",
		"Unusual SYSTEMROOT values in process environment blocks of elevated processes.",
		"Keep OS patched; deny user control of SYSTEMROOT.",
		[]string{"T1548.002", "T1574"}},
	{"UAC-DLL-HIJACK", "Auto-elevate DLL search-order", "dllhijack", SevHigh, "various", "various",
		"Auto-elevate apps that load DLLs insecurely from CWD or user-writable paths elevate the hijacker's code.",
		"DLL loads from non-System32 paths into auto-elevate processes; Sysmon ImageLoaded.",
		"SafeDLLSearchMode; remove CWD from search; WDAC.",
		[]string{"T1548.002", "T1574.001"}},
	{"UAC-TOKEN-DUP", "Token Duplication / UIAccess", "other", SevMed, "Win7+", "n/a",
		"UIAccess or elevated token duplication tricks (where integrity checks are weak) can yield elevated rights.",
		"DuplicateHandle / ImpersonateLoggedOnUser from medium IL into elevated tokens.",
		"AlwaysNotify; restrict UIAccess; monitor token APIs.",
		[]string{"T1548.002", "T1134"}},
	{"UAC-SILENTCLEANUP", "SilentCleanup scheduled task", "other", SevHigh, "Win8+", "cleanmgr / DiskCleanup",
		"SilentCleanup task runs with highest privileges and honors user-writable environment variables (windir), enabling elevated code.",
		"schtasks SilentCleanup invocation; windir env override; cleanmgr children.",
		"Lock down the task / harden env; EDR on windir overrides.",
		[]string{"T1548.002", "T1053.005"}},
	{"UAC-COMPUTERDEFAULTS", "ComputerDefaults.exe", "registry", SevHigh, "Win10+", "ComputerDefaults.exe",
		"Auto-elevate binary reading ms-settings\\Shell\\Open\\command (same family as fodhelper).",
		"ComputerDefaults.exe → unusual child; ms-settings registry writes.",
		"Same controls as fodhelper.",
		[]string{"T1548.002", "T1546.015"}},
	{"UAC-MMCDDE", "MMC DDE / .msc", "other", SevMed, "Win7+", "mmc.exe",
		"Crafted .msc / DDE can chain into elevated mmc contexts when auto-elevate components are involved.",
		"mmc.exe launching with untrusted .msc; DDE traffic.",
		"Block untrusted .msc; AppLocker.",
		[]string{"T1548.002", "T1170"}},
	{"UAC-PERFMON", "perfmon.exe / MMC", "registry", SevMed, "Win7+", "perfmon.exe",
		"Performance Monitor auto-elevates via MMC; association hijacks apply.",
		"perfmon → shell; Classes hijacks.",
		"Registry monitoring + reduce auto-elevate set.",
		[]string{"T1548.002"}},
	{"UAC-SYSREP", "SystemReset / Sysprep abuse", "other", SevMed, "Win7-8", "sysprep.exe",
		"Historic DLL hijacks (cryptbase/dwmapi/etc.) against auto-elevate sysprep.",
		"sysprep loading unexpected DLLs.",
		"Patched on modern OS; keep updated.",
		[]string{"T1548.002", "T1574.001"}},
	{"UAC-CLICONFG", "cliconfg.exe DLL", "dllhijack", SevMed, "Win7", "cliconfg.exe",
		"SQL Server Client Network Utility auto-elevates and was DLL-hijackable.",
		"cliconfg loading NTWDBLIB from CWD.",
		"Remove unused tools; patch.",
		[]string{"T1548.002", "T1574.001"}},
	{"UAC-INETMGR", "InetMgr.exe (IIS)", "dllhijack", SevMed, "Win with IIS", "InetMgr.exe",
		"IIS Manager auto-elevates; DLL search-order issues historically abused.",
		"InetMgr unexpected DLL loads.",
		"Install only when needed; WDAC.",
		[]string{"T1548.002"}},
}

func lookupUAC(idOrName string) *uacTech {
	q := strings.ToLower(strings.TrimSpace(idOrName))
	for i := range allUAC() {
		t := &allUAC()[i]
		if strings.ToLower(t.ID) == q || strings.Contains(strings.ToLower(t.Name), q) ||
			strings.Contains(strings.ToLower(t.Binary), q) {
			return t
		}
	}
	return nil
}

// Scan offline config exports (reg / txt) for exposure indicators.
func scanUACConfig(text, target string) AnalysisResult {
	res := AnalysisResult{Target: target, Tool: APPNAME, Stats: map[string]string{}}
	var findings []Finding
	n := 0
	checks := []struct {
		id, sev, title, detail, rem string
		re                          *regexp.Regexp
		mitre                       []string
	}{
		{"UAC-CFG-MSSETTINGS", SevHigh, "ms-settings Shell\\Open\\command present",
			"HKCU ms-settings association is the fodhelper/ComputerDefaults hijack surface.",
			"Clear the key; alert on writes to Software\\Classes\\ms-settings.",
			regexp.MustCompile(`(?i)ms-settings\\Shell\\Open\\command`), []string{"T1548.002", "T1546.015"}},
		{"UAC-CFG-MSCFILE", SevHigh, "mscfile shell\\open\\command hijack surface",
			"eventvwr / mmc association hijack indicator.",
			"Reset mscfile association; monitor Classes\\mscfile.",
			regexp.MustCompile(`(?i)mscfile\\shell\\open\\command`), []string{"T1548.002"}},
		{"UAC-CFG-ENABLELUA-0", SevCrit, "EnableLUA disabled (UAC off)",
			"ConsentPromptBehavior / EnableLUA=0 means UAC is effectively off.",
			"Set EnableLUA=1 and ConsentPromptBehaviorAdmin=2 (Always Notify).",
			regexp.MustCompile(`(?i)EnableLUA\s*[:=]\s*0\b`), []string{"T1548.002"}},
		{"UAC-CFG-CONSENT-0", SevHigh, "ConsentPromptBehaviorAdmin=0 (elevate without prompt)",
			"Admin approval mode will auto-elevate without consent.",
			"Set ConsentPromptBehaviorAdmin to 2 (Always Notify).",
			regexp.MustCompile(`(?i)ConsentPromptBehaviorAdmin\s*[:=]\s*0\b`), []string{"T1548.002"}},
		{"UAC-CFG-PROMPT-1", SevMed, "ConsentPromptBehaviorAdmin=1 (prompt on secure desktop off variant)",
			"Weaker prompt posture than Always Notify.",
			"Prefer value 2.",
			regexp.MustCompile(`(?i)ConsentPromptBehaviorAdmin\s*[:=]\s*1\b`), []string{"T1548.002"}},
		{"UAC-CFG-FILTERADMIN-0", SevHigh, "FilterAdministratorToken disabled",
			"Built-in Administrator runs at high IL without filtering.",
			"Enable FilterAdministratorToken.",
			regexp.MustCompile(`(?i)FilterAdministratorToken\s*[:=]\s*0\b`), []string{"T1548.002"}},
		{"UAC-CFG-VIRTUAL-0", SevLow, "EnableVirtualization disabled",
			"File/registry virtualization off for legacy apps (posture signal).",
			"Usually leave enabled unless app-compat requires otherwise.",
			regexp.MustCompile(`(?i)EnableVirtualization\s*[:=]\s*0\b`), []string{"T1548.002"}},
		{"UAC-CFG-WINDIR", SevHigh, "windir / SYSTEMROOT override",
			"User-controlled windir enables SilentCleanup / env-based elevation.",
			"Ensure windir and SYSTEMROOT are not user-writable overrides.",
			regexp.MustCompile(`(?i)(windir|SYSTEMROOT)\s*[:=]\s*[A-Z]:\\Users\\`), []string{"T1548.002", "T1574"}},
		{"UAC-CFG-AUTOELEV", SevMed, "autoElevate manifest reference in export",
			"Presence of autoElevate true in captured manifests — inventory these binaries.",
			"Reduce the set of auto-elevate apps via WDAC/AppLocker where unused.",
			regexp.MustCompile(`(?i)autoElevate\s*>?\s*true|<autoElevate>true`), []string{"T1548.002"}},
		{"UAC-CFG-SILENTCLEANUP", SevHigh, "SilentCleanup task referenced",
			"SilentCleanup is a known UAC bypass vehicle via windir.",
			"Audit the scheduled task ACLs and windir integrity.",
			regexp.MustCompile(`(?i)SilentCleanup`), []string{"T1548.002", "T1053.005"}},
	}
	for _, c := range checks {
		if c.re.MatchString(text) {
			n++
			ev := c.re.FindString(text)
			findings = append(findings, Finding{
				ID: fmt.Sprintf("%s-%03d", c.id, n), Title: c.title, Severity: c.sev,
				Category: "config", Detail: c.detail, Evidence: ev, Remediation: c.rem,
				Mitre: c.mitre, Score: sevWeight(c.sev),
			})
		}
	}
	// also flag any technique binary mentioned
	for _, t := range allUAC() {
		if t.Binary != "" && t.Binary != "n/a" && t.Binary != "various" && t.Binary != "various auto-elevate" {
			if contains(text, t.Binary) {
				findings = append(findings, Finding{
					ID: "UAC-BIN-" + t.ID, Title: "Auto-elevate / bypass binary present: " + t.Binary,
					Severity: SevInfo, Category: "inventory", Detail: t.Name + " — " + t.Summary,
					Evidence: t.Binary, Remediation: t.Mitigation, Mitre: t.Mitre, Score: 1,
				})
			}
		}
	}
	res.Findings = findings
	res.Stats["techniques_cataloged"] = strconv.Itoa(len(uacCatalog))
	res.Stats["config_hits"] = strconv.Itoa(n)
	res.Stats["bytes"] = strconv.Itoa(len(text))
	res.Grade = gradeFromFindings(findings)
	return res
}

func domainAnalyze(path string) (AnalysisResult, error) {
	text, err := readFileString(path)
	if err != nil {
		return AnalysisResult{}, err
	}
	return scanUACConfig(text, baseName(path)), nil
}

func domainDemoResult() AnalysisResult {
	demo := strings.Join([]string{
		`Windows Registry Editor Version 5.00`,
		`[HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System]`,
		`"EnableLUA"=dword:00000000`,
		`"ConsentPromptBehaviorAdmin"=dword:00000000`,
		`"FilterAdministratorToken"=dword:00000000`,
		`[HKEY_CURRENT_USER\Software\Classes\ms-settings\Shell\Open\command]`,
		`@="cmd.exe"`,
		`"DelegateExecute"="|"`,
		`windir=C:\Users\bob\AppData\Local\Temp`,
		`TaskName=SilentCleanup`,
	}, "\n")
	res := scanUACConfig(demo, "demo-uac.reg")
	res.Notes = append(res.Notes, "Demo parses a synthetic .reg export; no live host was modified.")
	return res
}

func domainHowtoSteps() []string {
	return []string{
		"Export the target's UAC-related registry (Policies\\System) and HKCU\\Software\\Classes.",
		"Also capture scheduled tasks (SilentCleanup) and process inventories of auto-elevate binaries.",
		"Run OUACBypass --input export.reg (or use the TUI Analyze option).",
		"Review findings: disabled UAC, weak consent prompts, hijackable Classes keys, env overrides.",
		"Consult the technique catalog for detection and mitigation per bypass family.",
		"Export a graded SARIF/JSON report for the engagement.",
	}
}

func domainRequirements() []string {
	return []string{
		"Offline exports: .reg / .txt of Policies\\System, HKCU Classes, schtasks, optionally manifests.",
		"No admin rights needed to run OUACBypass itself.",
		"Authorization to assess the Windows host's UAC posture.",
	}
}

func domainHowtoNotes() []Section {
	return []Section{
		{Title: "OPERATOR NOTES", Lines: ouacbypassRefNotes()},
		{Title: "TECHNIQUE FAMILIES", Lines: []string{
			"registry  — Classes / App Paths hijacks (fodhelper, eventvwr, sdclt, WSReset…)",
			"com       — elevation monikers (CMSTPLUA, ColorDataProxy…)",
			"mockdir   — trusted-path lookalikes with trailing space/dot",
			"env       — SYSTEMROOT / windir overrides (SilentCleanup)",
			"dllhijack — search-order hijacks of auto-elevate apps",
		}},
		{Title: "HARDENING TARGETS", Lines: []string{
			"EnableLUA=1 ; ConsentPromptBehaviorAdmin=2 ; FilterAdministratorToken=1",
			"Always Notify; WDAC/AppLocker to constrain auto-elevate binaries;",
			"Monitor HKCU\\Software\\Classes\\ms-settings and mscfile.",
		}},
	}
}

func domainCatalogSections() []Section {
	by := map[string][]string{}
	for _, t := range allUAC() {
		by[t.Method] = append(by[t.Method], fmt.Sprintf("%-18s [%-8s] %s — %s", t.ID, t.Severity, t.Name, t.Binary))
	}
	var secs []Section
	for _, m := range []string{"registry", "com", "mockdir", "env", "dllhijack", "other"} {
		if lines, ok := by[m]; ok {
			secs = append(secs, Section{Title: "METHOD: " + strings.ToUpper(m), Lines: lines})
		}
	}
	secs = append(secs, sigmaSection())
	return secs
}

func domainMenuItems() []MenuItem {
	return []MenuItem{
		{Label: "Lookup a UAC bypass technique", Action: actionUACLookup},
		{Label: "List auto-elevate binaries", Action: actionAutoElev},
		{Label: "Hardening checklist", Action: actionUACHarden},
	}
}

func actionUACLookup() {
	clear()
	banner()
	fmt.Println("  " + col("⬡  TECHNIQUE LOOKUP", B+CY))
	fmt.Println()
	q := prompt("id / name / binary")
	t := lookupUAC(q)
	if t == nil {
		warnMsg("not found")
		pause()
		return
	}
	okMsg(t.ID + " — " + t.Name)
	fmt.Println("   " + col("method   : ", GY) + col(t.Method, W))
	fmt.Println("   " + col("severity : ", GY) + col(t.Severity, sevColor(t.Severity)))
	fmt.Println("   " + col("windows  : ", GY) + col(t.WinVer, W))
	fmt.Println("   " + col("binary   : ", GY) + col(t.Binary, W))
	for _, ln := range wrap(t.Summary, 60) {
		fmt.Println("   " + col(ln, W))
	}
	fmt.Println("   " + col("detect   : ", GY) + col(t.Detection, YLW))
	fmt.Println("   " + col("mitigate : ", GY) + col(t.Mitigation, GRN))
	fmt.Println("   " + col("mitre    : ", GY) + col(strings.Join(t.Mitre, ", "), R))
	pause()
}

func actionAutoElev() {
	clear()
	banner()
	fmt.Println("  " + col("⬡  AUTO-ELEVATE / BYPASS BINARIES", B+CY))
	fmt.Println()
	seen := map[string]bool{}
	var rows [][]string
	for _, t := range allUAC() {
		if t.Binary == "" || t.Binary == "n/a" || seen[t.Binary] {
			continue
		}
		seen[t.Binary] = true
		rows = append(rows, []string{t.Binary, t.Method, t.Severity, t.ID})
	}
	table([]string{"BINARY", "METHOD", "SEV", "ID"}, rows)
	pause()
}

func actionUACHarden() {
	clear()
	banner()
	fmt.Println("  " + col("⬡  HARDENING CHECKLIST", B+CY))
	fmt.Println()
	renderSections([]Section{{Title: "UAC HARDENING", Lines: []string{
		"1. EnableLUA = 1",
		"2. ConsentPromptBehaviorAdmin = 2 (Always Notify)",
		"3. ConsentPromptBehaviorUser = 0 (deny elevation for standard users)",
		"4. FilterAdministratorToken = 1",
		"5. PromptOnSecureDesktop = 1",
		"6. Deploy WDAC/AppLocker — block unused auto-elevate binaries",
		"7. Alert on writes to HKCU\\Software\\Classes\\ms-settings and mscfile",
		"8. Alert on windir / SYSTEMROOT overrides in user environments",
		"9. Audit SilentCleanup scheduled task ACLs",
		"10. Keep Windows fully patched (mock-dir / env hijack fixes)",
	}}})
	pause()
}

var uacExtra = []uacTech{
	{"UAC-DCCW", "dccw.exe / Color Management", "dllhijack", SevMed, "Win7+", "dccw.exe",
		"Display Color Calibration auto-elevates; historic GDI DLL hijacks.",
		"dccw loading unexpected DLLs from CWD.", "Remove unused; WDAC.", []string{"T1548.002"}},
	{"UAC-BITLOCKERWIZ", "BitLocker Wizard FJ", "other", SevMed, "Win10", "BitLockerWizardElev.exe",
		"Auto-elevate BitLocker wizard abused via fileless / reg tricks in research.",
		"BitLockerWizardElev unusual children.", "Monitor + constrain.", []string{"T1548.002"}},
	{"UAC-PKGMANAGER", "PkgMgr / Dism hijack family", "dllhijack", SevMed, "Win7", "pkgmgr.exe",
		"Package Manager auto-elevates and historically loaded hijackable DLLs.",
		"pkgmgr unexpected ImageLoaded events.", "Keep patched; AppLocker.", []string{"T1548.002"}},
	{"UAC-WINSAT", "WinSAT DLL hijack", "dllhijack", SevMed, "Win7", "WinSAT.exe",
		"Windows System Assessment Tool auto-elevates; DLL hijack via path.",
		"WinSAT loading from non-system paths.", "Patched; inventory still useful.", []string{"T1548.002", "T1574.001"}},
	{"UAC-MIGWIZ", "migwiz / GUP", "dllhijack", SevLow, "Win7", "migwiz.exe",
		"Legacy migration wizard auto-elevate DLL issues.",
		"migwiz ImageLoaded anomalies.", "Remove legacy tools.", []string{"T1548.002"}},
}

func allUAC() []uacTech { return append(append([]uacTech{}, uacCatalog...), uacExtra...) }

var uacSigmaHints = []string{
	"Alert: Image=fodhelper.exe AND ParentImage NOT IN (trusted) AND CommandLine contains unusual args",
	"Alert: Registry set Key ends with ms-settings\\Shell\\Open\\command",
	"Alert: Registry set Key ends with mscfile\\shell\\open\\command",
	"Alert: Image=dllhost.exe CommandLine contains CMSTPLUA CLSID {3E5FC7F9-9A51-4367-9063-A120244FBEC7}",
	"Alert: Environment windir set under User hive to a path under %USERPROFILE%",
	"Alert: Task Scheduler SilentCleanup executed AND recent windir change",
	"Alert: Process integrity High AND Parent integrity Medium AND Parent NOT consent.exe",
	"Alert: EnableLUA registry value changed to 0",
}

func sigmaSection() Section { return Section{Title: "DETECTION HINTS (SIEM)", Lines: uacSigmaHints} }

func ouacbypassRefNotes() []string {
	return []string{
		"Authorized testing only — do not bypass UAC/escape containers outside of scope.",
		"Prefer detection engineering: convert catalog entries into SIEM/EDR analytics.",
		"Re-test after patches; many historic techniques are version-specific.",
		"Combine with least-privilege hardening checklists shipped in this tool.",
		"Export SARIF into the engagement's vulnerability management pipeline.",
		"Keep offline copies of configs; avoid live probing of production control planes.",
		"Document residual risk when a bypass/escape vector cannot be fully removed.",
		"Cross-reference MITRE ATT&CK technique pages for detection data sources.",
		"For containers: admission controllers beat after-the-fact scanning alone.",
		"For UAC: Always Notify + WDAC is stronger than hunting individual hijacks.",
		"Validate findings with a lab reproduction before filing critical risk.",
		"Track OS build numbers alongside each technique's affected range.",
		"Treat docker.sock and privileged as equivalent to host root compromise.",
		"Token/COM elevation monikers deserve the same scrutiny as registry hijacks.",
		"Schedule periodic re-scans; config drift re-introduces escape surfaces.",
	}
}

func initOUACBypassNotes() {
	// referenced by howto notes expansion
	_ = ouacbypassRefNotes
}

// ---- OWL house help content + self-test ----
var owlHelpAbout = []string{
	"OUACBypass is an offline reference catalog and config auditor for Windows UAC (User Account Control) bypass techniques. It documents the well-known primitives — fodhelper / ComputerDefaults (ms-settings), eventvwr / mmc (mscfile), SilentCleanup (windir), auto-elevate manifests — and maps each to MITRE ATT&CK.",
	"Given a captured config (registry export, whoami, manifest or similar text), it scans for the registry keys and settings that enable those bypasses — EnableLUA=0, weak ConsentPromptBehaviorAdmin, FilterAdministratorToken off, windir overrides — and grades the exposure. It reads files only; it never changes a live system.",
}
var owlHelpFlow = []string{
	"Type the number of an option and press Enter. Start with [1] DEMO for a synthetic walkthrough, [2] Analyze to scan a captured config, or [3] Reference catalog to study the techniques. [H] opens this help, [0] or q quits.",
	"Everything is also available from the command line: --input FILE --format txt|md|json|csv|sarif --output FILE. See page 2.",
}
var owlHelpMenu = []owlHelpItem{
	{"1", "Run DEMO: analyse a built-in synthetic config so you can see the findings and grading."},
	{"2", "Analyze an input file: scan a captured registry/config export for UAC-bypass indicators."},
	{"3", "Reference catalog: the UAC-bypass techniques OUACBypass knows about, with ATT&CK IDs and hardening."},
	{"*", "View / Export last result: review or write the most recent analysis as txt/md/json/csv/sarif."},
	{"*", "HOW TO USE / About: the usage screen and the MITRE ATT&CK mapping."},
}
var owlHelpTips = []owlHelpItem{
	{"*", "This is a defensive auditor: it flags the registry state that ENABLES a bypass so you can harden it, not a tool that performs one."},
	{"*", "EnableLUA=0 is graded CRITICAL — UAC is effectively off. Set it back to 1 with ConsentPromptBehaviorAdmin=2."},
	{"*", "Export a target's HKLM\\...\\System policies and HKCU Classes keys to a text file, then analyse it here."},
	{"*", "Run --selftest to confirm the config scanner, grader and all five export formats."},
}

func runSelfTest() int {
	banner()
	header("Self-Test — UAC config scanner / grader / export")
	pass, fail := 0, 0
	check := func(name string, cond bool) {
		if cond {
			pass++
			fmt.Println("  " + col("[✔] ", GRN) + col(name, W))
		} else {
			fail++
			fmt.Println("  " + col("[✘] ", R) + col(name, W))
		}
	}

	// Demo pipeline.
	res := domainDemoResult()
	finalize(&res)
	check("demo result has findings", len(res.Findings) > 0)
	check("demo result graded", res.Grade != "")

	// Scan a synthetic config with known bypass indicators.
	sample := "EnableLUA=0\r\n" +
		"ConsentPromptBehaviorAdmin=0\r\n" +
		"HKCU\\Software\\Classes\\ms-settings\\Shell\\Open\\command\r\n" +
		"SilentCleanup task enabled\r\n"
	ar := scanUACConfig(sample, "selftest")
	finalize(&ar)
	check("scanUACConfig finds indicators", len(ar.Findings) >= 3)
	hasCrit := false
	for _, f := range ar.Findings {
		if f.Severity == SevCrit {
			hasCrit = true
		}
	}
	check("EnableLUA=0 raises CRITICAL", hasCrit)

	// Clean config yields no high findings.
	clean := scanUACConfig("EnableLUA=1\r\nConsentPromptBehaviorAdmin=2\r\n", "clean")
	highs := 0
	for _, f := range clean.Findings {
		if f.Severity == SevHigh || f.Severity == SevCrit {
			highs++
		}
	}
	check("hardened config has no HIGH/CRITICAL", highs == 0)

	// Export round-trip.
	check("exportTXT non-empty", len(exportTXT(&ar)) > 0)
	check("exportMD non-empty", len(exportMD(&ar)) > 0)
	check("exportCSV non-empty", len(exportCSV(&ar)) > 0)
	check("exportSARIF non-empty", len(exportSARIF(&ar)) > 0)
	var v interface{}
	check("exportJSON is valid JSON", json.Unmarshal([]byte(exportJSON(&ar)), &v) == nil)

	check("reference catalog non-empty", len(domainCatalogSections()) > 0)

	fmt.Println()
	rule()
	if fail == 0 {
		okMsg(fmt.Sprintf("ALL %d CHECKS PASSED", pass))
		return 0
	}
	errMsg(fmt.Sprintf("%d passed, %d FAILED", pass, fail))
	return 1
}
