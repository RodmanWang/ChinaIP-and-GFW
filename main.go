package main

import (
	"bufio"
	"math/big"
	"net/http"
	"net/netip"
	"os"
	"strconv"
)

func main() {
	err := test()
	if err != nil {
		panic(err)
	}
}

func test() error {
	u := "https://github.com/Hackl0us/GeoIP2-CN/raw/release/CN-ip-cidr.txt"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	c := &http.Client{
		// Transport: &http.Transport{
		// 	Proxy: func(r *http.Request) (*url.URL, error) {
		// 		return url.Parse("http://127.0.0.1:1080")
		// 	},
		// },
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var (
		cidr, ipv4Cidr, ipv6Cidr []string

		br      = bufio.NewScanner(resp.Body)
		rangeIp = []*big.Int{{}, {}, big.NewInt(1)}
		lastV4  = []*big.Int{{}, {}}
		lastV6  = []*big.Int{{}, {}}
		fillBuf = make([]byte, 16)
	)
	for br.Scan() {
		ip, err := netip.ParsePrefix(br.Text())
		if err != nil {
			continue
		}

		// rangeIp[0] 当前行起始ip地址
		rangeIp[0].SetBytes(ip.Addr().AsSlice())
		rangeIp[1].Set(rangeIp[0])

		last, tmp, ipCidr, setBit := lastV4, fillBuf[:4], &ipv4Cidr, 31
		if ip.Addr().Is6() {
			last, tmp, ipCidr, setBit = lastV6, fillBuf, &ipv6Cidr, 127
		}

		for i := setBit - ip.Bits(); i >= 0; i-- {
			rangeIp[1].SetBit(rangeIp[1], i, 1)
		}
		rangeIp[1].Add(rangeIp[1], rangeIp[2])

		if last[1].Cmp(rangeIp[0]) == 0 {
			// 本行起始ip是上一行结束ip+1,可以组成连续ip范围
			last[1].Set(rangeIp[1])
		} else {
			if last[1].BitLen() > 0 {
				cidr = ipRangeToCIDR(cidr[:0], tmp, last[0], last[1].Sub(last[1], rangeIp[2]))
				*ipCidr = append(*ipCidr, cidr...) // 根据ip起止范围计算CIDR表达式
			}
			last[0].Set(rangeIp[0])
			last[1].Set(rangeIp[1])
		}
	}
	err = br.Err()
	if err != nil {
		return err
	}

	if lastV4[1].BitLen() > 0 {
		cidr = ipRangeToCIDR(cidr[:0], fillBuf[:4], lastV4[0], lastV4[1].Sub(lastV4[1], rangeIp[2]))
		ipv4Cidr = append(ipv4Cidr, cidr...)
	}
	if lastV6[1].BitLen() > 0 {
		cidr = ipRangeToCIDR(cidr[:0], fillBuf, lastV6[0], lastV6[1].Sub(lastV6[1], rangeIp[2]))
		ipv6Cidr = append(ipv6Cidr, cidr...)
	}

	fwCidr := func(name string, cidr []string) error {
		fw, err := os.Create(name)
		if err != nil {
			return err
		}
		defer fw.Close()
		for _, v := range cidr {
			fw.WriteString(v + "\n")
		}
		return nil
	}

	err = fwCidr("ipv4.txt", ipv4Cidr)
	if err != nil {
		return err
	}
	err = fwCidr("ipv6.txt", ipv6Cidr)
	if err != nil {
		return err
	}
	return nil
}

func ipRangeToCIDR(cidr []string, buf []byte, ipsInt, ipeInt *big.Int) []string {
	var (
		tmpInt = new(big.Int)
		mask   = new(big.Int)
		one    = big.NewInt(1)
		bits   uint
		maxBit = uint(len(buf) * 8)
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
