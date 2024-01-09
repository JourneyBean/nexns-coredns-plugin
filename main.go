package nexns

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
)

type NexnsPlugin struct {
	Next          plugin.Handler
	ControllerURL string
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
		log.Println("[Nexns] Connecting to notification channel.")
		controllerURL := strings.Replace(p.ControllerURL, "http", "ws", 1)
		conn, _, err := websocket.DefaultDialer.Dial(controllerURL+"api/v1/ws/client-notify/", nil)
		if err != nil {
			log.Println("[Nexns] Failed to connect to notification channel:", err)
			return err
		}
		log.Println("[Nexns] Successfully connected to notification channel.")
		defer conn.Close()

		for {
			// 从上游服务器读取消息
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			notificationData := WSNotification{}
			err = json.Unmarshal(msg, &notificationData)
			if err != nil {
				log.Println("[Nexns] Error parsing notification data:", err)
			}

			log.Println("[Nexns] Loading domain id:", notificationData.Domain)
			p.loadDomainDataFromURL(notificationData.Domain)
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
