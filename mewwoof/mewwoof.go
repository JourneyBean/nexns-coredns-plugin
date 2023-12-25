package mewwoof

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type MewwoofPlugin struct {
	Next          plugin.Handler
	ControllerURL string
	Data          []DomainData
}

func (p MewwoofPlugin) Name() string {
	return "mewwoof"
}

func (p *MewwoofPlugin) Init() error {
	err := p.loadDataFromURL()
	if err != nil {
		return fmt.Errorf("failed to initialize plugin: %v", err)
	}

	log.Println("Mewwoof Plugin Init success. Controller URL:", p.ControllerURL)

	return nil
}

func (p MewwoofPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}
	log.Println("from:", state.IP(), "query:",
		state.QName(), dns.ClassToString[state.QClass()], dns.TypeToString[state.QType()])

	domain, zone, rrset, error := p.findMatchingRRSet(r.Question[0].Name, r.Question[0].Qtype, net.ParseIP(state.IP()))

	// NXDOMAIN
	if rrset == nil || error != nil {
		msg := new(dns.Msg)
		msg.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(msg)
		return dns.RcodeNameError, nil
	}

	// return resource records
	return p.returnRecords(w, r, *domain, *zone, *rrset), nil
}

func (p *MewwoofPlugin) loadDataFromURL() error {

	// Send HTTP GET request
	response, err := http.Get(p.ControllerURL)
	if err != nil {
		return fmt.Errorf("HTTP request error: %v", err)
	}
	defer response.Body.Close()

	// Read response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Read response body error: %v", err)
	}

	// Parse JSON data
	err = json.Unmarshal(body, &p.Data)
	if err != nil {
		return fmt.Errorf("JSON parsing error: %v", err)
	}

	return nil
}

// findMatchingRRSet 根据DNS名称和源IP查找匹配的区域和记录
func (p *MewwoofPlugin) findMatchingRRSet(
	requestName string,
	requestType uint16,
	srcIP net.IP,
) (*Domain, *Zone, *RRSet, error) {

	for _, domain_data := range p.Data {

		domain_fqdn := dns.Fqdn(domain_data.Domain.Name)
		if !strings.HasSuffix(requestName, domain_fqdn) {
			continue
		}

		subdomain := strings.TrimSuffix(requestName, domain_fqdn)
		subdomain = subdomain[:len(subdomain)-1]

		for _, zone := range domain_data.Zones {

			for _, rule := range zone.Rules {

				_, ipNet, err := net.ParseCIDR(rule)
				if err != nil {
					continue
				}

				// 检查源IP是否匹配ACL列表中的规则
				if ipNet.Contains(srcIP) {

					for _, rrset := range zone.RRsets {

						if rrset.Name == subdomain && rrset.Type == dns.TypeToString[requestType] {
							return &domain_data.Domain, &zone, &rrset, nil
						}
					}
				}
			}
		}
	}

	return nil, nil, nil, nil
}

// returnRecords 返回RRset中的记录
func (p *MewwoofPlugin) returnRecords(
	w dns.ResponseWriter,
	r *dns.Msg,
	domain Domain,
	zone Zone,
	rrset RRSet,
) int {

	// request state
	state := request.Request{W: w, Req: r}

	// new msg
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, record := range rrset.Records {
		dnsType := dns.StringToType[rrset.Type]
		rr := dns.TypeToRR[dns.StringToType[rrset.Type]]()
		responseHeader := dns.RR_Header{
			Name:   state.QName(),
			Rrtype: dnsType,
			Class:  state.QClass(),
			Ttl:    uint32(record.TTL),
		}

		switch dnsType {

		case dns.TypeA:
			rr.(*dns.A).Hdr = responseHeader
			rr.(*dns.A).A = net.ParseIP(record.Data)

		case dns.TypeAAAA:
			rr.(*dns.AAAA).Hdr = responseHeader
			rr.(*dns.AAAA).AAAA = net.ParseIP(record.Data)

		case dns.TypeTXT:
			rr.(*dns.TXT).Hdr = responseHeader
			rr.(*dns.TXT).Txt = []string{record.Data}

		// case dns.TypeSRV:
		// 	rr.(*dns.SRV).Hdr = responseHeader
		// rr.(*dns.SRV)

		case dns.TypeMX:
			parts := strings.Fields(record.Data)
			preference, _ := strconv.Atoi(parts[0])

			rr.(*dns.MX).Hdr = responseHeader
			rr.(*dns.MX).Preference = uint16(preference)
			rr.(*dns.MX).Mx = parts[1]

		case dns.TypeNS:
			rr.(*dns.NS).Hdr = responseHeader
			rr.(*dns.NS).Ns = record.Data

		}

		msg.Answer = append(msg.Answer, rr)
	}

	// write msg
	w.WriteMsg(msg)

	return dns.RcodeSuccess
}
