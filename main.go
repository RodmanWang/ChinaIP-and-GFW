package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	srcFile      string
	dstFile      string
	databaseType string
	cnRecord     = mmdbtype.Map{
		"country": mmdbtype.Map{
			"geoname_id":           mmdbtype.Uint32(1814991),
			"is_in_european_union": mmdbtype.Bool(false),
			"iso_code":             mmdbtype.String("CN"),
			"names": mmdbtype.Map{
				"de":    mmdbtype.String("China"),
				"en":    mmdbtype.String("China"),
				"es":    mmdbtype.String("China"),
				"fr":    mmdbtype.String("Chine"),
				"ja":    mmdbtype.String("中国"),
				"pt-BR": mmdbtype.String("China"),
				"ru":    mmdbtype.String("Китай"),
				"zh-CN": mmdbtype.String("中国"),
			},
		},
	}
)

func init() {
	flag.StringVar(&srcFile, "s", "ipip_cn.txt", "specify source ip list file")
	flag.StringVar(&dstFile, "d", "Country.mmdb", "specify destination mmdb file")
	flag.StringVar(&databaseType, "t", "GeoIP2-Country", "specify MaxMind database type")
	flag.Parse()
}

func main() {
	writer, err := mmdbwriter.New(
		mmdbwriter.Options{
			DatabaseType: databaseType,
			RecordSize:   24,
		},
	)
	if err != nil {
		logrus.Fatalf("fail to new writer %v\n", err)
	}

	var ipTxtList []string
	fh, err := os.Open(srcFile)
	if err != nil {
		logrus.Fatalf("fail to open %s\n", err)
	}
	scanner := bufio.NewScanner(fh)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		ipTxtList = append(ipTxtList, scanner.Text())
	}

	ipv4Cidr, ipv6Cidr := processIPRanges(ipTxtList)
	for _, cidr := range ipv4Cidr {
		err = writer.InsertCIDR(cidr, cnRecord)
		if err != nil {
			logrus.Fatalf("fail to insert IPv4 CIDR to writer %v\n", err)
		}
	}
	for _, cidr := range ipv6Cidr {
		err = writer.InsertCIDR(cidr, cnRecord)
		if err != nil {
			logrus.Fatalf("fail to insert IPv6 CIDR to writer %v\n", err)
		}
	}

	outFh, err := os.Create(dstFile)
	if err != nil {
		logrus.Fatalf("fail to create output file %v\n", err)
	}

	_, err = writer.WriteTo(outFh)
	if err != nil {
		logrus.Fatalf("fail to write to file %v\n", err)
	}
}

func processIPRanges(ipTxtList []string) (ipv4Cidr, ipv6Cidr []string) {
	var (
		rangeIp     = []*big.Int{{}, {}, big.NewInt(1)}
		lastV4      = []*big.Int{{}, {}}
		lastV6      = []*big.Int{{}, {}}
		fillBuf     = make([]byte, 16)
	)

	for _, ipTxt := range ipTxtList {
		ipTxt = strings.TrimSpace(ipTxt)
		ipParts := strings.Split(ipTxt, "/")
		if len(ipParts) != 2 {
			continue
		}

		ip := ipParts[0]
		maskStr := ipParts[1]
		mask, err := strconv.Atoi(maskStr)
		if err != nil {
			continue
		}

		ips := strings.Split(ip, ".")
		if len(ips) != 4 {
			continue
		}

		// Convert IP to big.Int
		for i := 0; i < 4; i++ {
			part, _ := strconv.Atoi(ips[i])
			rangeIp[0].SetBit(rangeIp[0], 8*i, uint(part))
		}

		rangeIp[1].Set(rangeIp[0])

		ipCidr := &ipv4Cidr
		if mask > 32 {
			ipCidr = &ipv6Cidr
		}

		for i := uint(32 - mask); i > 0; i-- {
			rangeIp[1].SetBit(rangeIp[1], 31-i, 1)
		}
		rangeIp[1].Add(rangeIp[1], rangeIp[2])

		if lastV4[1].Cmp(rangeIp[0]) == 0 || lastV6[1].Cmp(rangeIp[0]) == 0 {
			lastV4[1].Set(rangeIp[1])
			lastV6[1].Set(rangeIp[1])
		} else {
			if lastV4[1].BitLen() > 0 {
				cidr := fmt.Sprintf("%s/%d", bigIntToIPv4(rangeIp[0]), mask)
				*ipCidr = append(*ipCidr, cidr)
			}
			lastV4[0].Set(rangeIp[0])
			lastV4[1].Set(rangeIp[1])

			if lastV6[1].BitLen() > 0 {
				cidr := fmt.Sprintf("%s/%d", bigIntToIPv6(rangeIp[0]), mask)
				ipv6Cidr = append(ipv6Cidr, cidr)
			}
			lastV6[0].Set(rangeIp[0])
			lastV6[1].Set(rangeIp[1])
		}
	}

	if lastV4[1].BitLen() > 0 {
		cidr := fmt.Sprintf("%s/%d", bigIntToIPv4(lastV4[0]), 32-int(lastV4[1].BitLen()))
		ipv4Cidr = append(ipv4Cidr, cidr)
	}
	if lastV6[1].BitLen() > 0 {
		cidr := fmt.Sprintf("%s/%d", bigIntToIPv6(lastV6[0]), 128-int(lastV6[1].BitLen()))
		ipv6Cidr = append(ipv6Cidr, cidr)
	}

	return ipv4Cidr, ipv6Cidr
}

func bigIntToIPv4(ip *big.Int) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip.Bits()[3], ip.Bits()[2], ip.Bits()[1], ip.Bits()[0])
}

func bigIntToIPv6(ip *big.Int) string {
	return fmt.Sprintf("%04x:%04x:%04x:%04x:%04x:%04x:%04x:%04x",
		ip.Bits()[7], ip.Bits()[6], ip.Bits()[5], ip.Bits()[4],
		ip.Bits()[3], ip.Bits()[2], ip.Bits()[1], ip.Bits()[0])
}
