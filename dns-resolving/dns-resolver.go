package dns_resolving

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/ui-kreinhard/go-dns-rabbit/activeQuery"
	"log"
	"strconv"
)

type DNSResolver struct {
	table *DNSTable
	querySender *activeQuery.ActiveQueryHandler
}

func NewDNSResolver(table *DNSTable, handler *activeQuery.ActiveQueryHandler) *DNSResolver {
	return &DNSResolver{
		table,
		handler,
	}
}

func (d *DNSResolver) parseQuery(m *dns.Msg) {
	for _, query := range m.Question {
		switch query.Qtype {
		case dns.TypeA:
			log.Printf("Query for %s\n", query.Name)
			ip, err := d.table.Lookup(query.Name)
			if err == nil {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", query.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			} else {
				log.Println("failed to find entry", query.Name)
				log.Println("Try to resolve it via explicit call")
				response := make(chan activeQuery.DNSResponse)
				d.querySender.Query(query.Name, response)
				dnsResponse := <- response
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", query.Name, dnsResponse.IP))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}	
			}
		}
	}
}



func (d *DNSResolver) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	log.Println("Got query")
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		d.parseQuery(msg)
	}

	w.WriteMsg(msg)	
}

func (d *DNSResolver) Listen() {
	// attach request handler func
	dns.HandleFunc(".", d.handleDnsRequest)

	// start server
	port := 5553
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}

}