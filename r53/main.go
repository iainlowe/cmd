// Command r53ls lists AWS Route53 hosts
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	flag "github.com/ogier/pflag"
	"launchpad.net/goamz/aws"
)

type A struct {
	XMLName            xml.Name `xml:"ListResourceRecordSetsResponse"`
	ResourceRecordSets RRSets
}

type RRSets struct {
	XMLName xml.Name `xml:"ResourceRecordSets"`
	RRSets  []RRSet  `xml:"ResourceRecordSet"`
}

type RRSet struct {
	XMLName xml.Name `xml:"ResourceRecordSet"`

	Name            string
	Type            string
	TTL             string
	ResourceRecords RRs `xml:"ResourceRecords"`
}

type RRs struct {
	XMLName         xml.Name `xml:"ResourceRecords"`
	ResourceRecords []RR     `xml:"ResourceRecord"`
}

type RR struct {
	XMLName xml.Name `xml:"ResourceRecord"`
	Value   string
}

type RREntry struct {
	Name  string
	Value string
	Type  string
	TTL   string
}

func getRRs() []RREntry {
	var auth aws.Auth
	var req *http.Request
	var err error

	if auth, err = aws.EnvAuth(); err != nil {
		log.Fatal(err)
	}

	if req, err = http.NewRequest("GET", "https://route53.amazonaws.com/2013-04-01/hostedzone/"+zone+"/rrset", bytes.NewBuffer([]byte(""))); err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Date", time.Now().Add(time.Minute*4).UTC().Format("20060102T150405Z"))
	req.Header.Add("Host", "route53.amazonaws.com")

	if err = aws.SignV4(req, auth, aws.USEast.Name); err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	// resptxt, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(resptxt))

	dec := xml.NewDecoder(resp.Body)

	rrset := &A{}
	err = dec.Decode(rrset)

	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	ret := make([]RREntry, 0)

	for _, rr := range rrset.ResourceRecordSets.RRSets {
		values := []string{}

		for _, r := range rr.ResourceRecords.ResourceRecords {
			values = append(values, strings.TrimSuffix(r.Value, "."))
		}

		ret = append(ret, RREntry{
			Name:  strings.TrimSuffix(strings.Replace(rr.Name, "\\052", "*", 1), "."),
			Value: strings.Join(values, ", "),
			TTL:   rr.TTL,
			Type:  rr.Type,
		})
	}

	return ret
}

var zone string
var showPrivate bool
var showAll bool
var quiet bool
var verbose bool

type RRWrapper struct {
	rrs *[]RREntry
}

func (r *RRWrapper) Len() int {
	return len(*r.rrs)
}

func init() {
	flag.BoolVarP(&showPrivate, "show-private", "p", false, "Include private records (192.x and 10.x)")
	flag.StringVarP(&zone, "zone", "z", os.Getenv("R53_ZONE_ID"), "The zone to list for")
	flag.BoolVarP(&quiet, "quiet", "q", false, "Terse output")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Long output")
	flag.BoolVarP(&showAll, "show-all", "A", false, "Show all record types (incl. TXT MX etc.); by default show only CNAME and A records")

	flag.Usage = func() {
		fmt.Println(`You need the following definitions in your env:

	export AWS_ACCESS_KEY_ID=KJHGDKA7876JHG8JH8
	export AWS_SECRET_ACCESS_KEY=kjhKJhkasduhKJH3224kjasd
	export R53_ZONE_ID=AKJHKHJRBDSDA

Usage of r53:`)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	rrTypes := make(map[string]bool)

	if len(flag.Args()) > 0 {
		for _, arg := range flag.Args() {
			rrTypes[strings.ToUpper(arg)] = true
		}
	} else if !showAll {
		rrTypes["CNAME"] = true
		rrTypes["A"] = true
	}

	rrs := getRRs()

	for _, rr := range rrs {
		if len(rrTypes) > 0 && !rrTypes[rr.Type] {
			continue
		}

		if !showPrivate && (strings.HasPrefix(rr.Value, "192.") || strings.HasPrefix(rr.Value, "10.")) {
			continue
		}

		switch {
		case quiet:
			fmt.Println(rr.Value)
		case verbose:
			fmt.Printf("%45s %8s %6s (%s)\n", rr.Name, rr.Type, rr.TTL, rr.Value)
		case len(rrTypes) == 1 && rrTypes["A"]:
			fmt.Printf("%15s %s\n", rr.Value, rr.Name)
		case len(rrTypes) == 1 && rrTypes["CNAME"]:
			fmt.Printf("%45s -> %s\n", rr.Name, rr.Value)
		default:
			fmt.Printf("%45s %8s (%s)\n", rr.Name, rr.Type, rr.Value)
		}
	}
}
