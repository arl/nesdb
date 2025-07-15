package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	. "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	"github.com/kanmu/go-sqlfmt/sqlfmt"
)

//go:embed nes.db
var nesDB []byte

var (
	mapper   int
	region   string
	battery  BoolFlag
	showChip bool
	dryrun   bool
	order    multiFlag
)

const allRegions = "All"

func cli() {
	flag.IntVar(&mapper, "m", -1, "Filter by iNES mapper number")
	flag.IntVar(&mapper, "mapper", -1, "Filter by iNES mapper number")
	flag.StringVar(&region, "r", allRegions, "Filter by iNES region")
	flag.StringVar(&region, "region", allRegions, "Filter by iNES region")
	flag.Var(&battery, "b", "Filter by presence of battery-packed RAM")
	flag.Var(&battery, "battery", "Filter by presence of battery-packed RAM")
	flag.BoolVar(&showChip, "c", false, "Show chip column")
	flag.BoolVar(&showChip, "showchip", false, "Show chip column")
	flag.BoolVar(&dryrun, "q", false, "Show SQL query (but do not execute it)")
	flag.BoolVar(&dryrun, "query", false, "Show SQL query (but do not execute it)")
	flag.Var(&order, "o", "Order results by column (defaults to name, can be repeated)")
	flag.Var(&order, "order", "Order results by column (defaults to name, can be repeated)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage:
  %s [options]

Filters:
  -m, --mapper   Filter by iNES mapper number
  -b, --battery  Filter by presence of battery-packed RAM
  -r, --region   Filter by iNES region

Options:
  -q, --query    Show SQL Query (but do not execute it)
  -c, --showchip Show chip column
  -o, --order    Order results by column (default to name, can be repeated)


`, os.Args[0])
	}
}

func main() {
	cli()
	flag.Parse()

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
	d = d.Select(cols...)

	// Build filters
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

	region = cleanRegion(region)
	if region != allRegions {
		filters = append(filters, C("g.region").Eq(region))
	}

	if len(filters) > 0 {
		d = d.Where(filters...)
	}

	if len(order) > 0 {
		for _, o := range order {
			d = d.OrderAppend(C(o).Asc())
		}
	} else {
		d = d.Order(C("g.name").Asc())
	}

	query, args, err := d.ToSQL()
	_ = args
	if err != nil {
		log.Fatalf("building SQL: %v", err)
	}

	query1 := strings.ReplaceAll(query, "`", "")
	query2, err := sqlfmt.Format(query1, &sqlfmt.Options{Distance: 8})
	if err != nil {
		query2 = query1
	}
	if dryrun {
		fmt.Println("SQL:", query2)
	} else {
		if err := run(query2); err != nil {
			log.Fatal(err)
		}
	}
}

func run(query string) error {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	if _, err := f.Write(nesDB); err != nil {
		return err
	}

	bin, err := exec.LookPath("sqlite3")
	if err != nil {
		return fmt.Errorf("sqlite3 not found in PATH: %v", err)
	}

	out, err := exec.Command(bin,
		f.Name(),
		"--readonly",
		".mode table", query).CombinedOutput()
	if err != nil {
		fmt.Printf("Output: %s\n", out)
		return fmt.Errorf("sqlite3 failed: %v", err)
	}
	fmt.Printf("%s\n", out)
	return nil
}

func cleanRegion(region string) string {
	region = strings.Title(strings.ToLower(region))
	if region == "Usa" {
		return "USA"
	}
	return region
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

// IsSet reports whether the flag has been set.
func (f *BoolFlag) IsSet() bool { return f.set }

// Value reports the flag value if it has been set, else always false.
func (f *BoolFlag) Value() bool { return f.val }

type multiFlag []string

func (f *multiFlag) String() string { return fmt.Sprint(*f) }

func (f *multiFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}
