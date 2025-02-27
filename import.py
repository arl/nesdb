#!/usr/bin/env python3

import sys
import sqlite3
import xml.etree.ElementTree as ET

DB_SCHEMA = """
-- Table: game
CREATE TABLE game (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL,
    altname       TEXT,
    class         TEXT,
    subclass      TEXT,
    catalog       TEXT,
    publisher     TEXT,
    developer     TEXT,
    region        TEXT,
    players       TEXT,   -- or INTEGER, but the DB can hold it as text if it’s not always numeric
    date          TEXT    -- storing as TEXT to hold YYYY-MM-DD or partial
);

-- Table: cartridge
CREATE TABLE cartridge (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id      INTEGER NOT NULL,
    system       TEXT,
    revision     TEXT,
    crc          TEXT,
    sha1         TEXT,
    dump         TEXT,
    dumper       TEXT,
    datedumped   TEXT,
    FOREIGN KEY(game_id) REFERENCES game(id)
);

-- Table: device
CREATE TABLE device (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id   INTEGER NOT NULL,
    type      TEXT,
    name      TEXT,
    FOREIGN KEY(game_id) REFERENCES game(id)
);

-- Table: chip_pin
CREATE TABLE chip_pin (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    chip_id   INTEGER NOT NULL,
    number    TEXT,
    function  TEXT,
    FOREIGN KEY(chip_id) REFERENCES chip(id)
);

-- Table: board
CREATE TABLE board (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    cartridge_id  INTEGER NOT NULL,
    type       TEXT,
    pcb        TEXT,
    mapper     TEXT,
    FOREIGN KEY(cartridge_id) REFERENCES cartridge(id)
);

-- Table: prg
CREATE TABLE prg (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id   INTEGER NOT NULL,
    name       TEXT,
    size       TEXT,
    crc        TEXT,
    sha1       TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);

-- Table: chr
CREATE TABLE chr (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id   INTEGER NOT NULL,
    name       TEXT,
    size       TEXT,
    crc        TEXT,
    sha1       TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);

-- Table: vram
CREATE TABLE vram (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id  INTEGER NOT NULL,
    size      TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);

-- Table: wram
CREATE TABLE wram (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id  INTEGER NOT NULL,
    size      TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);

-- Table: chip
CREATE TABLE chip (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id  INTEGER NOT NULL,
    type      TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);

-- Table: cic
CREATE TABLE cic (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id  INTEGER NOT NULL,
    type      TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);

-- Table: pad
CREATE TABLE pad (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id  INTEGER NOT NULL,
    h         TEXT,
    v         TEXT,
    FOREIGN KEY(board_id) REFERENCES board(id)
);
"""

def create_schema(conn):
    """Create all necessary tables if they do not exist."""
    conn.executescript(DB_SCHEMA)
    conn.commit()

def insert_game(conn, game_el):
    """Insert a <game> element into 'game' table."""
    # Extract attributes
    name = game_el.get("name")
    altname = game_el.get("altname")
    _class = game_el.get("class")
    subclass = game_el.get("subclass")
    catalog = game_el.get("catalog")
    publisher = game_el.get("publisher")
    developer = game_el.get("developer")
    region = game_el.get("region")
    players = game_el.get("players")
    date = game_el.get("date")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO game(name, altname, class, subclass, catalog, publisher, developer, region, players, date)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    """, (name, altname, _class, subclass, catalog, publisher, developer, region, players, date))
    conn.commit()
    print(name)
    return cursor.lastrowid

def insert_cartridge(conn, game_id, cart_el):
    """Insert a <cartridge> element into 'cartridge' table."""
    system = cart_el.get("system")
    revision = cart_el.get("revision")
    crc = cart_el.get("crc")
    sha1 = cart_el.get("sha1")
    dump = cart_el.get("dump")
    dumper = cart_el.get("dumper")
    datedumped = cart_el.get("datedumped")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO cartridge(game_id, system, revision, crc, sha1, dump, dumper, datedumped)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    """, (game_id, system, revision, crc, sha1, dump, dumper, datedumped))
    conn.commit()
    return cursor.lastrowid

def insert_board(conn, cartridge_id, board_el):
    """Insert a <board> element into 'board' table."""
    btype = board_el.get("type")
    pcb = board_el.get("pcb")
    mapper = board_el.get("mapper")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO board(cartridge_id, type, pcb, mapper)
        VALUES (?, ?, ?, ?)
    """, (cartridge_id, btype, pcb, mapper))
    conn.commit()
    return cursor.lastrowid

def insert_prg(conn, board_id, prg_el):
    """Insert a <prg> element into 'prg' table."""
    name = prg_el.get("name")
    size = prg_el.get("size")
    crc = prg_el.get("crc")
    sha1 = prg_el.get("sha1")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO prg(board_id, name, size, crc, sha1)
        VALUES (?, ?, ?, ?, ?)
    """, (board_id, name, size, crc, sha1))
    conn.commit()

def insert_chr(conn, board_id, chr_el):
    """Insert a <chr> element into 'chr' table."""
    name = chr_el.get("name")
    size = chr_el.get("size")
    crc = chr_el.get("crc")
    sha1 = chr_el.get("sha1")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO chr(board_id, name, size, crc, sha1)
        VALUES (?, ?, ?, ?, ?)
    """, (board_id, name, size, crc, sha1))
    conn.commit()

def insert_vram(conn, board_id, vram_el):
    """Insert <vram> info into 'vram' table."""
    size = vram_el.get("size")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO vram(board_id, size)
        VALUES (?, ?)
    """, (board_id, size))
    conn.commit()

def insert_wram(conn, board_id, wram_el):
    """Insert <wram> info into 'wram' table."""
    size = wram_el.get("size")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO wram(board_id, size)
        VALUES (?, ?)
    """, (board_id, size))
    conn.commit()

def insert_chip(conn, board_id, chip_el):
    """Insert <chip> info into 'chip' table."""
    ctype = chip_el.get("type")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO chip(board_id, type)
        VALUES (?, ?)
    """, (board_id, ctype))
    conn.commit()
    return cursor.lastrowid

def insert_device(conn, game_id, device_el):
    """Insert <device> info into 'device' table."""
    dtype = device_el.get("type")
    dname = device_el.get("name")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO device(game_id, type, name)
        VALUES (?, ?, ?)
    """, (game_id, dtype, dname))
    conn.commit()

def insert_chip_pin(conn, chip_id, pin_el):
    """Insert <pin> info into 'chip_pin' table."""
    number = pin_el.get("number")
    function = pin_el.get("function")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO chip_pin(chip_id, number, function)
        VALUES (?, ?, ?)
    """, (chip_id, number, function))
    conn.commit()

def insert_cic(conn, board_id, cic_el):
    """Insert <cic> info into 'cic' table."""
    ctype = cic_el.get("type")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO cic(board_id, type)
        VALUES (?, ?)
    """, (board_id, ctype))
    conn.commit()

def insert_pad(conn, board_id, pad_el):
    """Insert <pad> info into 'pad' table."""
    h = pad_el.get("h")
    v = pad_el.get("v")

    cursor = conn.cursor()
    cursor.execute("""
        INSERT INTO pad(board_id, h, v)
        VALUES (?, ?, ?)
    """, (board_id, h, v))
    conn.commit()

def process_peripherals(conn, game_id, game_el):
    """Handle <peripherals> -> <device> elements."""
    peripherals_el = game_el.find("peripherals")
    if peripherals_el is not None:
        for device_el in peripherals_el.findall("device"):
            insert_device(conn, game_id, device_el)

def process_game(conn, game_el):
    """Handle one <game> node: insert the game, then process its cartridges, boards, etc."""
    game_id = insert_game(conn, game_el)

    # Process <peripherals> after or before cartridges, as you prefer
    process_peripherals(conn, game_id, game_el)

    for cartridge_el in game_el.findall("cartridge"):
        cart_id = insert_cartridge(conn, game_id, cartridge_el)

        # Each cartridge has one or more <board> children.
        for board_el in cartridge_el.findall("board"):
            board_id = insert_board(conn, cart_id, board_el)

            # Insert <prg> if present
            for prg_el in board_el.findall("prg"):
                insert_prg(conn, board_id, prg_el)

            # Insert <chr> if present
            for chr_el in board_el.findall("chr"):
                insert_chr(conn, board_id, chr_el)

            # Insert <vram> if present
            for vram_el in board_el.findall("vram"):
                insert_vram(conn, board_id, vram_el)

            # Insert <wram> if present
            for wram_el in board_el.findall("wram"):
                insert_wram(conn, board_id, wram_el)

            # Insert <chip> if present
            for chip_el in board_el.findall("chip"):
                chip_id = insert_chip(conn, board_id, chip_el)
                # Insert <pin> inside <chip>
                for pin_el in chip_el.findall("pin"):
                    insert_chip_pin(conn, chip_id, pin_el)

            # Insert <cic> if present
            for cic_el in board_el.findall("cic"):
                insert_cic(conn, board_id, cic_el)

            # Insert <pad> if present
            for pad_el in board_el.findall("pad"):
                insert_pad(conn, board_id, pad_el)

def main():
    if len(sys.argv) < 3:
        print("Usage: {} cartdb.xml output.db".format(sys.argv[0]))
        sys.exit(1)

    xml_file = sys.argv[1]
    db_file  = sys.argv[2]

    # 1) Parse the XML
    tree = ET.parse(xml_file)
    root = tree.getroot()  # <database> is root

    # 2) Create/connect SQLite DB
    conn = sqlite3.connect(db_file)

    # 3) Create schema
    create_schema(conn)

    # 4) Iterate over <game> elements
    for game_el in root.findall("game"):
        process_game(conn, game_el)

    print("Data imported successfully into", db_file)
    conn.close()

if __name__ == "__main__":
    main()
