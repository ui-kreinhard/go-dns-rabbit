package heartbeat

import (
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
	"github.com/ui-kreinhard/go-dns-rabbit/chamqp"
	"github.com/ui-kreinhard/go-dns-rabbit/common-types"
	"log"
	"net"
	"os"
	"time"
)

type HeartbeatSender struct {
	iface string
	rabbitChan *chamqp.Channel
}

func NewHeartbeatSender(rabbitChan *chamqp.Channel, iface string) *HeartbeatSender {
	return &HeartbeatSender{
		iface,
		rabbitChan,
	}
}


func (h *HeartbeatSender) Send() {
	exchangeName := "dns.tunnel"
	h.rabbitChan.ExchangeDeclare(exchangeName, "topic", false, false, false, false, nil, nil)

	for {
		heartbeat, err := heartbeat(h.iface)
		if err != nil {
			log.Println("err creating heartbeat", err)
			continue
		}
		log.Println("Sending heartbeat with", heartbeat)
		rawData, err := json.Marshal(heartbeat)
		if err != nil {
			log.Println("err marshalling", err)
			continue
		}
		err = h.rabbitChan.Publish(
			exchangeName,
			h.iface,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/octet-stream",
				Body:        rawData,
			})
		if err != nil {
			log.Println(err)
		}
		time.Sleep(30 *time.Second)
	}
}

func heartbeat(ifaceName string) (common_types.DNSHeartbeat, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return common_types.DNSHeartbeat{}, err
	}
	ip, err := GetInternalIP(ifaceName)
	return common_types.DNSHeartbeat{
		hostname + "." + ifaceName + ".",
		ip,
	},nil
}

func GetInternalIP(ifaceName string) (string, error) {
	itf, _ := net.InterfaceByName(ifaceName) //here your interface
	item, _ := itf.Addrs()
	var ip net.IP
	for _, addr := range item {
		switch v := addr.(type) {
		case *net.IPNet:
			if !v.IP.IsLoopback() {
				if v.IP.To4() != nil {//Verify if IP is IPV4
					ip = v.IP
				}
			}
		}
	}
	if ip != nil {
		return ip.String(), nil
	} 
	return "", errors.New("iface not found")
}