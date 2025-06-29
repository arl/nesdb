package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	. "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

var (
	mapper   int
	battery  BoolFlag
	showChip bool
	verbose  bool
)

func init() {
	flag.IntVar(&mapper, "m", -1, "Filter by iNES mapper number")
	flag.IntVar(&mapper, "mapper", -1, "Filter by iNES mapper number")
	flag.Var(&battery, "b", "Filter by presence of battery-packed RAM")
	flag.Var(&battery, "battery", "Filter by presence of battery-packed RAM")
	flag.BoolVar(&showChip, "c", false, "Show chip column")
	flag.BoolVar(&showChip, "showchip", false, "Show chip column")
	flag.BoolVar(&verbose, "v", false, "Verbose execution (print SQL query)")
	flag.BoolVar(&verbose, "verbose", false, "Verbose execution (print SQL query)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  %s [options]

Filters:
  -m, --mapper   Filter by iNES mapper number
  -b, --battery  Filter by presence of battery-packed RAM

Options:
  -v, --verbose  Verbose execution (print SQL query)
  -c, --showchip Show chip column

`, os.Args[0])
	}
}

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Open database
	db, err := sql.Open("sqlite3", "nes.db")
	if err != nil {
		log.Fatalf("opening DB: %v", err)
	}
	defer db.Close()

	// Initialize goqu
	dialect := Dialect("sqlite3")
	d := dialect.From(T("game").As("g")).
		Join(
			T("cartridge").As("c"),
			On(Ex{"c.game_id": C("g.id")}),
		).
		Join(
			T("board").As("b"),
			On(Ex{"b.cartridge_id": C("c.id")}),
		).
		LeftJoin(T("prg"), On(Ex{"prg.board_id": C("b.id")})).
		LeftJoin(T("chr"), On(Ex{"chr.board_id": C("b.id")})).
		LeftJoin(T("vram"), On(Ex{"vram.board_id": C("b.id")})).
		LeftJoin(T("wram"), On(Ex{"wram.board_id": C("b.id")})).
		LeftJoin(T("chip"), On(Ex{"chip.board_id": C("b.id")}))

	// Select columns
	cols := []any{
		C("g.name").As("game_name"),
		C("g.region"),
		C("b.mapper"),
		C("b.type").As("board_type"),
		C("prg.name").As("prg_name"),
		C("prg.size").As("prg_size"),
		C("chr.name").As("chr_name"),
		C("chr.size").As("chr_size"),
		C("vram.size").As("vram_size"),
		C("wram.size").As("wram_size"),
		C("wram.battery").As("battery"),
	}
	if showChip {
		cols = append(cols, C("chip.type").As("chip_type"))
	}
	d = d.Select(cols...).Order(C("g.name").Asc())

	// Build WHERE clauses
	var filters []Expression
	if mapper >= 0 {
		filters = append(filters, C("b.mapper").Eq(mapper))
	}

	if battery.IsSet() {
		if battery.Value() {
			filters = append(filters, C("wram.battery").Eq(1))
		} else {
			filters = append(filters, L("wram.battery").IsNull())
		}
	}
	if len(filters) > 0 {
		d = d.Where(filters...)
	}

	query, args, err := d.ToSQL()
	if err != nil {
		log.Fatalf("building SQL: %v", err)
	}
	if verbose {
		fmt.Println("SQL:", query, args)
	}

	query = strings.ReplaceAll(query, "`", "")

	if err := run(query); err != nil {
		log.Fatal(err)
	}
}

func run(query string) error {
	bin, err := exec.LookPath("sqlite3")
	if err != nil {
		return fmt.Errorf("sqlite3 not found in PATH: %v", err)
	}

	out, err := exec.Command(bin, "nes.db", ".mode table", query).CombinedOutput()
	if err != nil {
		fmt.Printf("Output: %s\n", out)
		return fmt.Errorf("sqlite3 failed: %v", err)
	}
	fmt.Printf("%s\n", out)
	return nil
}

type BoolFlag struct {
	val bool
	set bool
}

func (f BoolFlag) String() string {
	if !f.set {
		return "<unset>"
	}
	return strconv.FormatBool(f.val)
}

func (f *BoolFlag) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid boolean value: %s", value)
	}
	f.val = v
	f.set = true
	return nil
}

func (f *BoolFlag) IsSet() bool { return f.set }
func (f *BoolFlag) Value() bool { return f.val }
