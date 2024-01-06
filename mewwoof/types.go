package nexns

// DomainData 包含了一个域下的所有信息
type DomainData struct {
	Domain Domain
	Zones  []Zone
}

// Domain 包含了域名、SOA、DNSSEC信息
type Domain struct {
	ID      int    `json:"id"`
	Name    string `json:"domain"`
	Mname   string `json:"mname"`
	Rname   string `json:"rname"`
	Serial  string `json:"serial"`
	Refresh int    `json:"refresh"`
	Retry   int    `json:"retry"`
	Expire  int    `json:"expire"`
	TTL     int    `json:"ttl"`
}

// Zone 包含了区域（zone）的规则信息
type Zone struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Rules  []string `json:"rules"`
	RRsets []RRSet  `json:"rrsets"`
}

// RRSet 包含了DNS资源记录集的信息
type RRSet struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Records []Record `json:"records"`
}

// Record 包含了DNS资源记录的信息
type Record struct {
	ID   int    `json:"id"`
	TTL  int    `json:"ttl"`
	Data string `json:"data"`
}
