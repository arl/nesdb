package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	mapper   int
	battery  boolFlag
	verbose  bool
	showChip bool
)

type boolFlag struct {
	val bool
	set bool
}

func (f boolFlag) String() string {
	if !f.set {
		return "<unset>"
	}
	return strconv.FormatBool(f.val)
}

func (f *boolFlag) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid boolean value: %s", value)
	}
	f.val = v
	f.set = true
	return nil
}

const usage = `Usage:
    searchnes [options]

Filters:
    -m, --mapper                Filter by iNES mapper number
    -b, --battery               Filter by presence/absence of battery-packed RAM

Options:
    -v, --verbose               Verbose execution (print sql query)
    -c, --showchip              Show chip column
`

func main() {
	flag.Usage = func() { fmt.Fprintf(os.Stderr, "%s\n", usage) }

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}
	flag.BoolVar(&verbose, "verbose", false, "Verbose execution (print sql query)")
	flag.BoolVar(&verbose, "v", false, "Verbose execution (print sql query)")
	flag.IntVar(&mapper, "mapper", -1, "Filter by iNES mapper")
	flag.IntVar(&mapper, "m", -1, "Filter by iNES mapper")
	flag.Var(&battery, "battery", "Filter by presence of battery-packed RAM")
	flag.Var(&battery, "b", "Filter by presence of battery-packed RAM")
	flag.BoolVar(&showChip, "showchip", false, "Show chip column")
	flag.BoolVar(&showChip, "c", false, "Show chip column")
	flag.Parse()

	const query = `
SELECT
    g.name AS game_name,
    g.region AS region,
    b.mapper,
    b.type as board_type,
    prg.name  AS prg_name,
    prg.size  AS prg_size,
    chr.name  AS chr_name,
    chr.size  AS chr_size,
    vram.size AS vram_size,
    wram.size AS wram_size,
    wram.battery AS battery
	%s
FROM game g
JOIN cartridge c ON c.game_id = g.id
JOIN board b     ON b.cartridge_id = c.id
LEFT JOIN prg    ON prg.board_id   = b.id
LEFT JOIN chr    ON chr.board_id   = b.id
LEFT JOIN vram   ON vram.board_id  = b.id
LEFT JOIN wram   ON wram.board_id  = b.id
LEFT JOIN chip   ON chip.board_id  = b.id
%s
ORDER BY g.name;`

	var whereClauses []string
	if mapper >= 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("b.mapper = %d", mapper))
	}
	if battery.set {
		if battery.val {
			whereClauses = append(whereClauses, "wram.battery = 1")
		} else {
			whereClauses = append(whereClauses, "wram.battery is NULL")
		}
	}

	bin, err := exec.LookPath("sqlite3")
	if err != nil {
		fmt.Printf("sqlite3 not found in PATH: %v\n", err)
		return
	}

	whereClause := ""
	if whereClauses != nil {
		whereClause = fmt.Sprintf("WHERE %s", strings.Join(whereClauses, " AND "))
	} else {
		whereClause = ""
	}
	chipClause := ""
	if showChip {
		chipClause = ",\n\tchip.type as chip_type"
	}
	q := fmt.Sprintf(query, chipClause, whereClause)
	println(q)
	out, err := exec.Command(bin, "nes.db", ".mode table", q).CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing query: %v\n", err)
		fmt.Printf("Output: %s\n", out)
		return
	}
	fmt.Printf("%s\n", out)
}
