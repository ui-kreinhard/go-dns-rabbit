package main

import (
	"github.com/ui-kreinhard/go-dns-rabbit/activeQuery"
	"github.com/ui-kreinhard/go-dns-rabbit/chamqp"
	"github.com/ui-kreinhard/go-dns-rabbit/common-types"
	"github.com/ui-kreinhard/go-dns-rabbit/dns-resolving"
	"github.com/ui-kreinhard/go-dns-rabbit/heartbeat"
	"log"
	"os"
)

func heartbeatHandling(rabbitChan *chamqp.Channel, dnsResponseChannel chan activeQuery.DNSResponse) *dns_resolving.DNSTable {
	heartbeatChannel := make(chan common_types.DNSHeartbeat)
	go heartbeat.NewHeartbeatListener(rabbitChan, heartbeatChannel).Listen()
	
	dnsTable := dns_resolving.NewDNSTable(60, heartbeatChannel, dnsResponseChannel)
	go dnsTable.ListenForHeartbeat()
	go dnsTable.ListenForQueryResponses()
	return dnsTable
}

func dnsResolvingHandling(rabbitChan *chamqp.Channel) {
	dnsResponseChannel := make(chan activeQuery.DNSResponse)
	activeQueryHandler := activeQuery.NewActiveQueryHandler(rabbitChan, dnsResponseChannel)
	go activeQueryHandler.Listen()
	
	dnsTable := heartbeatHandling(rabbitChan, dnsResponseChannel)
	go dnsTable.ListenForQueryResponses()
	dns_resolving.NewDNSResolver(dnsTable, activeQueryHandler).Listen()
}

func heartbeatSender(rabbitChan *chamqp.Channel, iface string) {
	go activeQuery.NewActiveQueryResponder(rabbitChan, iface).Listen()
	heartbeat.NewHeartbeatSender(rabbitChan, iface).Send()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	rabbitUrl := os.Args[1]
	mode := os.Args[2]
	conn := chamqp.Dial(rabbitUrl)
	channel := conn.Channel()
	if mode == "server" {
		dnsResolvingHandling(channel)
	} else {
		iface := os.Args[3]
		// assume client mode, I'm lazy...
		heartbeatSender(channel, iface)		
		select {}
	}
}
