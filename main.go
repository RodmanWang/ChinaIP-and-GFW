package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

const maxInputLineSize = 4 * 1024 * 1024

var cnRecord = mmdbtype.Map{
	"country": mmdbtype.Map{
		"geoname_id":           mmdbtype.Uint32(1814991),
		"is_in_european_union": mmdbtype.Bool(false),
		"iso_code":             mmdbtype.String("CN"),
		"names": mmdbtype.Map{
			"de": mmdbtype.String("China"), "en": mmdbtype.String("China"),
			"es": mmdbtype.String("China"), "fr": mmdbtype.String("Chine"),
			"ja": mmdbtype.String("中国"), "pt-BR": mmdbtype.String("China"),
			"ru": mmdbtype.String("Китай"), "zh-CN": mmdbtype.String("中国"),
		},
	},
}

func main() {
	src := flag.String("s", "ipip_cn.txt", "source IP/CIDR list")
	dst := flag.String("d", "Country.mmdb", "destination MMDB file")
	dbType := flag.String("t", "GeoIP2-Country", "MaxMind database type")
	flag.Parse()

	if err := generateMMDB(*src, *dst, *dbType); err != nil {
		log.Fatal(err)
	}
}

func generateMMDB(srcFile, dstFile, databaseType string) error {
	input, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("open source %q: %w", srcFile, err)
	}
	defer input.Close()

	cidrs, err := readCIDRs(input)
	if err != nil {
		return fmt.Errorf("read source %q: %w", srcFile, err)
	}
	if len(cidrs) == 0 {
		return errors.New("source contains no valid CIDR entries; refusing to create an empty database")
	}

	writer, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: databaseType, RecordSize: 24})
	if err != nil {
		return fmt.Errorf("create MMDB writer: %w", err)
	}
	for _, network := range cidrs {
		if err := writer.Insert(network, cnRecord); err != nil {
			return fmt.Errorf("insert %s: %w", network, err)
		}
	}

	outDir := filepath.Dir(dstFile)
	tmp, err := os.CreateTemp(outDir, ".mmdb-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output beside %q: %w", dstFile, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := writer.WriteTo(tmp); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write MMDB: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync MMDB: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close MMDB: %w", err)
	}
	if err := os.Rename(tmpName, dstFile); err != nil {
		return fmt.Errorf("publish MMDB to %q: %w", dstFile, err)
	}

	log.Printf("created %s with %d networks", dstFile, len(cidrs))
	return nil
}

func readCIDRs(input *os.File) ([]*net.IPNet, error) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), maxInputLineSize)
	var lines []string
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if lineNumber == 1 {
			line = trimBOM(line)
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan input near line %d: %w", lineNumber+1, err)
	}
	return parseCIDRs(lines)
}

func trimBOM(s string) string {
	if len(s) >= 3 && s[0] == 0xef && s[1] == 0xbb && s[2] == 0xbf {
		return s[3:]
	}
	return s
}
