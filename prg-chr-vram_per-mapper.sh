#!/bin/sh

MAPPER="${1:-0}"
echo $MAPPER

sqlite3 nes.db '
SELECT
    g.name AS game_name,
    b.mapper,
    prg.name  AS prg_name,
    prg.size  AS prg_size,
    prg.crc   AS prg_crc,
    prg.sha1  AS prg_sha1,
    chr.name  AS chr_name,
    chr.size  AS chr_size,
    chr.crc   AS chr_crc,
    chr.sha1  AS chr_sha1,
    vram.size AS vram_size
FROM game g
JOIN cartridge c ON c.game_id = g.id
JOIN board b     ON b.cartridge_id = c.id
LEFT JOIN prg    ON prg.board_id   = b.id
LEFT JOIN chr    ON chr.board_id   = b.id
LEFT JOIN vram   ON vram.board_id  = b.id
WHERE b.mapper = '$MAPPER'
ORDER BY g.name;'
