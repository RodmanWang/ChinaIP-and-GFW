package main

import (
	"bufio"
	"fmt"
	"math/big"
	"net/http"
	"net/netip"
	"os"
	"strconv"
)

func main() {
	err := processIPRanges()
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func processIPRanges() error {
	u := "https://github.com/Hackl0us/GeoIP2-CN/raw/release/CN-ip-cidr.txt"
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	br := bufio.NewScanner(resp.Body)
	var ipv4Cidr, ipv6Cidr []string
	lastV4, lastV6 := []*big.Int{big.NewInt(0), big.NewInt(0)}, []*big.Int{big.NewInt(0), big.NewInt(0)}
	fillBuf := make([]byte, 16)

	for br.Scan() {
		ip, err := netip.ParsePrefix(br.Text())
		if err != nil {
			continue
		}

		var rangeIp [3]*big.Int
		rangeIp[0] = new(big.Int).SetBytes(ip.Addr().AsSlice())
		rangeIp[1] = new(big.Int).Set(rangeIp[0])

		var ipCidr *[]string
		var setBit int
		if ip.Addr().Is6() {
			setBit, ipCidr = 127, &ipv6Cidr
		} else {
			setBit, ipCidr = 31, &ipv4Cidr
		}

		for i := setBit - ip.Bits(); i >= 0; i-- {
			rangeIp[1].SetBit(rangeIp[1], i, 1)
		}
		rangeIp[1].Add(rangeIp[1], rangeIp[2])

		if lastV4[1].Cmp(rangeIp[0]) == 0 && !ip.Addr().Is6() {
			lastV4[1].Set(rangeIp[1])
		} else if lastV6[1].Cmp(rangeIp[0]) == 0 && ip.Addr().Is6() {
			lastV6[1].Set(rangeIp[1])
		} else {
			if lastV4[1].BitLen() > 0 {
				*ipCidr = append(*ipCidr, ipRangeToCIDR(fillBuf[:4], lastV4[0], lastV4[1].Sub(lastV4[1], rangeIp[2]))...)
			}
			if lastV6[1].BitLen() > 0 {
				*ipCidr = append(*ipCidr, ipRangeToCIDR(fillBuf, lastV6[0], lastV6[1].Sub(lastV6[1], rangeIp[2]))...)
			}
			lastV4[0].Set(rangeIp[0])
			lastV4[1].Set(rangeIp[1])
			lastV6[0].Set(rangeIp[0])
			lastV6[1].Set(rangeIp[1])
		}
	}

	if lastV4[1].BitLen() > 0 {
		ipv4Cidr = append(ipv4Cidr, ipRangeToCIDR(fillBuf[:4], lastV4[0], lastV4[1].Sub(lastV4[1], big.NewInt(1)))...)
	}
	if lastV6[1].BitLen() > 0 {
		ipv6Cidr = append(ipv6Cidr, ipRangeToCIDR(fillBuf, lastV6[0], lastV6[1].Sub(lastV6[1], big.NewInt(1)))...)
	}

	err := writeCIDRToFile("ipv4.txt", ipv4Cidr)
	if err != nil {
		return err
	}
	err = writeCIDRToFile("ipv6.txt", ipv6Cidr)
	if err != nil {
		return err
	}

	fmt.Println("Optimized, sorted, and concatenated IP ranges saved to ipv4.txt and ipv6.txt")

	return nil
}

func ipRangeToCIDR(buf []byte, ipsInt, ipeInt *big.Int) []string {
	var (
		tmpInt = new(big.Int)
		mask   = new(big.Int)
		one    = big.NewInt(1)
		bits   uint
		maxBit = uint(len(buf) * 8)
		cidr   []string
	)

	for {
		bits = 1
		mask.SetUint64(1)
		for bits < maxBit {
			if (tmpInt.Or(ipsInt, mask).Cmp(ipeInt) > 0) ||
				(tmpInt.Lsh(tmpInt.Rsh(ipsInt, bits), bits).Cmp(ipsInt) != 0) {
				bits--
				mask.Rsh(mask, 1)
				break
			}
			bits++
			mask.Add(mask.Lsh(mask, 1), one)
		}

		addr, _ := netip.AddrFromSlice(ipsInt.FillBytes(buf))
		cidr = append(cidr, addr.String()+"/"+strconv.FormatUint(uint64(maxBit-bits), 10))

		if tmpInt.Or(ipsInt, mask); tmpInt.Cmp(ipeInt) >= 0 {
			break
		}
		ipsInt.Add(tmpInt, one)
	}
	return cidr
}

func writeCIDRToFile(fileName string, cidr []string) error {
	fw, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fw.Close()
	for _, v := range cidr {
		fw.WriteString(v + "\n")
	}
	return nil
}
