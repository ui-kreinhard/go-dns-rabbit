package heartbeat

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/ui-kreinhard/go-dns-rabbit/chamqp"
	"github.com/ui-kreinhard/go-dns-rabbit/common-types"
	"log"
	"os"
	"strconv"
)

type HeartbeatListener struct {
	rabbitChannel *chamqp.Channel	
	outChannel    chan common_types.DNSHeartbeat
}

func NewHeartbeatListener(rabbitChannel *chamqp.Channel, outChannel chan common_types.DNSHeartbeat) *HeartbeatListener {
	return &HeartbeatListener{
		rabbitChannel,
		outChannel,
	}
}

func (h *HeartbeatListener) Listen() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("Cannot get valid hostname")
	}
	heartbeatAMQPChannel := make(chan amqp.Delivery)
	errChan := make(chan error)
	
	exchangeName := "dns.tunnel"
	routingKey := "#"
	queue := "heartbeatlistener." + hostname + strconv.Itoa(os.Getpid())

	h.rabbitChannel.ExchangeDeclare(exchangeName, "topic", false, false, false, false, nil, errChan)
	h.rabbitChannel.QueueDeclare(queue, false, true, false, false, nil, nil, nil)
	h.rabbitChannel.QueueBind(queue, routingKey, exchangeName, false, nil, nil)
	h.rabbitChannel.Consume(queue, "", true, false, false, false, nil, heartbeatAMQPChannel, nil)
	for {
		select {
		case rawHeartbeat := <- heartbeatAMQPChannel:
			log.Println("Received heartbeat")
			heartbeat := common_types.DNSHeartbeat{}
			err  := json.Unmarshal(rawHeartbeat.Body, &heartbeat)
			if err != nil {
				log.Println("err unmarshalling", err)
				break
			}
			log.Println(heartbeat)
			h.outChannel <- heartbeat
		}
	}
}