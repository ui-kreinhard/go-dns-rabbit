package dns_resolving

import (
	"errors"
	"github.com/ui-kreinhard/go-dns-rabbit/activeQuery"
	"github.com/ui-kreinhard/go-dns-rabbit/common-types"
	"log"
	"time"
)

type DNSTable struct {
	entries         map[string]*DNSEntry
	maxTTL          float64
	heartbeatInChan chan common_types.DNSHeartbeat
	queryInChan     chan activeQuery.DNSResponse
}

func NewDNSTable(maxTTL float64, inChan chan common_types.DNSHeartbeat, queryInChan chan activeQuery.DNSResponse) *DNSTable {
	return &DNSTable{
		entries:         map[string]*DNSEntry{},
		maxTTL:          maxTTL,
		heartbeatInChan: inChan,
		queryInChan: queryInChan,
	}
}

type DNSEntry struct {
	name string
	ip string
	lastSeen time.Time
}

func (d *DNSTable) Lookup(name string) (string, error){
	if d.entries[name] != nil {
		return d.entries[name].ip, nil
	}
	return "", errors.New("Entry not found")
}

func (d *DNSTable) addOrUpdate(name, ip string) {
	d.entries[name] = &DNSEntry{
		name,
		ip,
		time.Now(),
	}	
}

func (d *DNSTable) RemoveOutdated() {
	for {
		for key, entry := range d.entries {
			if time.Now().Sub(entry.lastSeen).Seconds() >= d.maxTTL {
				d.remove(key)
			}
		}
		time.Sleep(60 * time.Second)
	}
}

func (d *DNSTable) remove(name string) {
	d.entries[name] = nil
}

func (d *DNSTable) ListenForHeartbeat() {
	log.Println("Listening for heartbeats from clients")
	for {
		select {
		case dnsHeartbeat := <- d.heartbeatInChan:
			d.addOrUpdate(dnsHeartbeat.Name, dnsHeartbeat.IP)
		}
	}
}

func (d *DNSTable) ListenForQueryResponses() {
	log.Println("Listening for query responses")
	for {
		select {
			case response := <- d.queryInChan:
				log.Println("Received query response to add/modify")
				d.addOrUpdate(response.NameToResolve, response.IP)
		}
	}
}