package nexns

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type NexnsPlugin struct {
	Next          plugin.Handler
	ControllerURL string
	ClientId      string
	ClientSecret  string
	Database      Trie
}

type WSNotification struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Domain int    `json:"domain"`
}

func (p *NexnsPlugin) Name() string {
	return "nexns"
}

func (p *NexnsPlugin) Init() error {
	err := p.loadAllDataFromURL()
	if err != nil {
		return fmt.Errorf("[Nexns] Failed to initialize plugin: %v", err)
	}

	// websocket to recv notifications
	go func() error {
		for {
			err := p.connectToNotificationChannel()
			if err != nil {
				time.Sleep(5 * time.Second) // 等待一段时间后尝试重新连接
			}
		}
	}()

	log.Println("[Nexns] Init success. Controller URL:", p.ControllerURL)

	return nil
}

func (p *NexnsPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {

	state := request.Request{W: w, Req: r}

	queryName := state.QName()
	queryType := dns.TypeToString[state.QType()]
	sourceIP := net.ParseIP(state.IP())

	domainData := p.Database.Search(queryName)

	// if domain not exists, pass to next plugin
	if domainData == nil {
		return plugin.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	}

	rrDataset := make([]dns.RR, 0)
	rrExtraset := make([]dns.RR, 0)

	// SOA
	if queryType == "SOA" {
		ds, _ := p.parseRecordData(&domainData.Domain, &RRSet{Name: "", Type: "SOA"}, &Record{}, sourceIP)
		rrDataset = append(rrDataset, ds...)
	}

	// regular response
	domain, rrset := p.searchRRsetFromDomainData(domainData, queryName, queryType, sourceIP)
	ds, es := p.parseRRset(domain, rrset, sourceIP)
	rrDataset = append(rrDataset, ds...)
	rrExtraset = append(rrExtraset, es...)

	// CNAME response, limit depth=1
	if state.QType() != dns.TypeCNAME {
		cnameDomain, cnameRRset := p.searchRRset(queryName, "CNAME", sourceIP)
		if cnameRRset != nil {
			ds, _ := p.parseRRset(cnameDomain, cnameRRset, sourceIP)

			// Add CNAME record itself
			rrDataset = append(rrDataset, ds...)

			// Add CNAMEd records of same type
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
