package main

import (
    "bufio"
    "flag"
    "github.com/maxmind/mmdbwriter"
    "github.com/maxmind/mmdbwriter/mmdbtype"
    log "github.com/sirupsen/logrus"
    "net"
    "os"
    "strings"
)

var (
    srcFile       string
    dstFile       string
    databaseType  string
    cnRecord = mmdbtype.Map{
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
        log.Fatalf("fail to new writer %v\n", err)
    }

    var ipTxtList []string
    fh, err := os.Open(srcFile)
    if err != nil {
        log.Fatalf("fail to open %s\n", err)
    }
    scanner := bufio.NewScanner(fh)
    scanner.Split(bufio.ScanLines)

    for scanner.Scan() {
        ipTxtList = append(ipTxtList, scanner.Text())
    }

    ipList := parseCIDRs(ipTxtList)

    mergedIPList := mergeAndOptimizeIPs(ipList)

    for _, ip := range mergedIPList {
        err = writer.Insert(ip, cnRecord)
        if err != nil {
            log.Fatalf("fail to insert to writer %v\n", err)
        }
    }

    outFh, err := os.Create(dstFile)
    if err != nil {
        log.Fatalf("fail to create output file %v\n", err)
    }

    _, err = writer.WriteTo(outFh)
    if err != nil {
        log.Fatalf("fail to write to file %v\n", err)
    }
}

func parseCIDRs(ipTxtList []string) []*net.IPNet {
    var ipList []*net.IPNet
    for _, ipTxt := range ipTxtList {
        ipParts := strings.Split(ipTxt, "/")
        ip := net.ParseIP(ipParts[0])
        mask, _ := net.IPMask(net.ParseIP("255.255.255.255").To4())
        prefix := &net.IPNet{IP: ip, Mask: mask}
        ipList = append(ipList, prefix)
    }
    return ipList
}

func mergeAndOptimizeIPs(ipList []*net.IPNet) []*net.IPNet {
    // 实现合并和优化 IP 地址段的逻辑
    // ...
}
