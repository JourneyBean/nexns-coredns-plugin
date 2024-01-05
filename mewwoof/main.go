package mewwoof

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type MewwoofPlugin struct {
	Next          plugin.Handler
	ControllerURL string
	Database      Trie
}

func (p MewwoofPlugin) Name() string {
	return "mewwoof"
}

func (p *MewwoofPlugin) Init() error {
	err := p.loadAllDataFromURL()
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

	queryName := state.QName()
	queryType := dns.TypeToString[state.QType()]
	sourceIP := net.ParseIP(state.IP())

	rrDataset := make([]dns.RR, 0)
	rrExtraset := make([]dns.RR, 0)

	// regular response
	domain, rrset := p.searchRRset(queryName, queryType, sourceIP)
	ds, es := p.parseRRset(domain, rrset, sourceIP)
	rrDataset = append(rrDataset, ds...)
	rrExtraset = append(rrExtraset, es...)

	// CNAME response, limit depth=1
	if state.QType() != dns.TypeCNAME {
		cnameDomain, cnameRRset := p.searchRRset(queryName, "CNAME", sourceIP)
		if cnameRRset != nil {
			ds, _ := p.parseRRset(cnameDomain, cnameRRset, sourceIP)

			rrDataset = append(rrDataset, ds...)

			for _, cnameRR := range ds {
				domain, rrset := p.searchRRset(cnameRR.(*dns.CNAME).Target, queryType, sourceIP)
				ds, es := p.parseRRset(domain, rrset, sourceIP)
				rrDataset = append(rrDataset, ds...)
				rrExtraset = append(rrExtraset, es...)
			}
		}
	}

	code, msg := p.writeAnswer(&rrDataset, &rrExtraset, r)
	w.WriteMsg(msg)
	return code, nil
}
