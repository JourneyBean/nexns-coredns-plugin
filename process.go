package nexns

import (
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

const MaxPacketSize = 512
const MaxTxtRecordSize = 255

func (p *NexnsPlugin) searchRRset(queryName string, queryTypeString string, sourceIP net.IP) (*Domain, *RRSet) {
	domainData := p.Database.Search(queryName)
	return p.searchRRsetFromDomainData(domainData, queryName, queryTypeString, sourceIP)
}

// 搜索trie树，匹配domain中的RRset
func (p *NexnsPlugin) searchRRsetFromDomainData(domainData *DomainData, queryName string, queryTypeString string, sourceIP net.IP) (*Domain, *RRSet) {

	if domainData == nil {
		return nil, nil
	}

	// find matching zone
	for _, zone := range domainData.Zones {

		for _, rule := range zone.Rules {

			_, ipNet, err := net.ParseCIDR(rule)
			if err != nil {
				continue
			}

			if ipNet.Contains(sourceIP) {

				// find matching prefix's rrset
				for _, rrset := range zone.RRsets {

					rrsetDomain := rrset.Name + "." + domainData.Domain.Name + "."
					if len(rrset.Name) == 0 {
						rrsetDomain = domainData.Domain.Name + "."
					}
					if rrsetDomain == queryName && rrset.Type == queryTypeString {

						// check if empty
						if len(rrset.Records) == 0 {
							return nil, nil
						}
						return &domainData.Domain, &rrset
					}

				}

			}

		}

	}

	return nil, nil
}

func (p *NexnsPlugin) writeAnswer(rrData *[]dns.RR, rrExtra *[]dns.RR, r *dns.Msg) (int, *dns.Msg) {
	if len(*rrData) == 0 {
		msg := new(dns.Msg)
		msg.SetRcode(r, dns.RcodeNameError)
		msg.Authoritative = true
		return dns.RcodeNameError, msg
	}

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true
	msg.Answer = append(msg.Answer, *rrData...)
	msg.Extra = append(msg.Answer, *rrExtra...)
	return dns.RcodeSuccess, msg
}

func (p *NexnsPlugin) parseRRset(domain *Domain, rrset *RRSet, sourceIP net.IP) ([]dns.RR, []dns.RR) {
	rrDataset := make([]dns.RR, 0)
	rrExtraset := make([]dns.RR, 0)

	if domain == nil || rrset == nil {
		return rrDataset, rrExtraset
	}

	// regular records
	for _, record := range rrset.Records {
		ds, es := p.parseRecordData(domain, rrset, &record, sourceIP)
		for _, rr := range ds {
			rrDataset = append(rrDataset, rr)
		}
		for _, rr := range es {
			rrDataset = append(rrDataset, rr)
		}
	}

	return rrDataset, rrExtraset
}

func (p *NexnsPlugin) parseRecordData(domain *Domain, rrset *RRSet, record *Record, sourceIP net.IP) ([]dns.RR, []dns.RR) {
	dnsType := dns.StringToType[rrset.Type]
	rrDataset := make([]dns.RR, 0)
	rrExtraset := make([]dns.RR, 0)

	if domain == nil || rrset == nil || record == nil {
		return rrDataset, rrExtraset
	}

	responseHeader := dns.RR_Header{
		Name:   getFqdn(rrset.Name, domain.Name),
		Rrtype: dnsType,
		Class:  dns.ClassINET,
		Ttl:    uint32(record.TTL),
	}

	switch dnsType {

	case dns.TypeA:
		rr := dns.TypeToRR[dnsType]()
		rr.(*dns.A).Hdr = responseHeader
		rr.(*dns.A).A = net.ParseIP(record.Data)
		rrDataset = append(rrDataset, rr)

	case dns.TypeAAAA:
		rr := dns.TypeToRR[dnsType]()
		rr.(*dns.AAAA).Hdr = responseHeader
		rr.(*dns.AAAA).AAAA = net.ParseIP(record.Data)
		rrDataset = append(rrDataset, rr)

	case dns.TypeTXT:
		chunks := splitIntoChunks(record.Data, MaxTxtRecordSize)
		for _, chunk := range chunks {
			rr := dns.TypeToRR[dnsType]()
			rr.(*dns.TXT).Hdr = responseHeader
			rr.(*dns.TXT).Txt = []string{chunk}
			rrDataset = append(rrDataset, rr)
		}

	case dns.TypeSOA:
		serial, _ := strconv.Atoi(domain.Serial)
		rr := dns.TypeToRR[dnsType]()
		rr.(*dns.SOA).Hdr = responseHeader
		rr.(*dns.SOA).Ns = getFqdn(domain.Mname, domain.Name)
		rr.(*dns.SOA).Mbox = getFqdn(domain.Rname, domain.Name)
		rr.(*dns.SOA).Serial = uint32(serial)
		rr.(*dns.SOA).Refresh = uint32(domain.Refresh)
		rr.(*dns.SOA).Retry = uint32(domain.Retry)
		rr.(*dns.SOA).Expire = uint32(domain.Expire)
		rr.(*dns.SOA).Minttl = uint32(domain.TTL)
		rrDataset = append(rrDataset, rr)

	case dns.TypeMX:
		parts := strings.Fields(record.Data)
		preference, _ := strconv.Atoi(parts[0])
		mx := getFqdn(parts[1], domain.Name)

		rr := dns.TypeToRR[dnsType]()
		rr.(*dns.MX).Hdr = responseHeader
		rr.(*dns.MX).Preference = uint16(preference)
		rr.(*dns.MX).Mx = mx
		rrDataset = append(rrDataset, rr)

		// add additional A, AAAA records
		ad1, arr1 := p.searchRRset(mx, "A", sourceIP)
		dataA, extraA := p.parseRRset(ad1, arr1, sourceIP)
		for _, r := range dataA {
			rrExtraset = append(rrDataset, r)
		}
		for _, r := range extraA {
			rrExtraset = append(rrDataset, r)
		}

		ad2, arr2 := p.searchRRset(mx, "AAAA", sourceIP)
		dataAAAA, extraAAAA := p.parseRRset(ad2, arr2, sourceIP)
		for _, r := range dataAAAA {
			rrExtraset = append(rrDataset, r)
		}
		for _, r := range extraAAAA {
			rrExtraset = append(rrDataset, r)
		}

	case dns.TypeCNAME:
		rr := dns.TypeToRR[dnsType]()
		rr.(*dns.CNAME).Hdr = responseHeader
		rr.(*dns.CNAME).Target = getFqdn(record.Data, domain.Name)
		rrDataset = append(rrDataset, rr)

	case dns.TypeNS:
		rr := dns.TypeToRR[dnsType]()
		rr.(*dns.NS).Hdr = responseHeader
		rr.(*dns.NS).Ns = getFqdn(record.Data, domain.Name)
		rrDataset = append(rrDataset, rr)

	}

	return rrDataset, rrExtraset
}

// splitIntoChunks splits a string into chunks of a given size
func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func getFqdn(prefixOrFqdn string, domainName string) string {
	fqdn := prefixOrFqdn
	if len(fqdn) == 0 {
		fqdn = domainName + "."
	} else if fqdn[len(fqdn)-1] != '.' {
		fqdn = fqdn + "." + domainName + "."
	}
	return fqdn
}
