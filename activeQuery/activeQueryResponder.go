package activeQuery

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/ui-kreinhard/go-dns-rabbit/chamqp"
	"github.com/ui-kreinhard/go-dns-rabbit/heartbeat"
	"log"
	"os"
	"strconv"
)

type ActiveQueryResponder struct {
	queryChannel  chan amqp.Delivery
	rabbitChannel *chamqp.Channel
	ifaceName     string 
	exchangeName  string
}

func NewActiveQueryResponder(rabbitChannel *chamqp.Channel, ifaceName string) *ActiveQueryResponder{
	return &ActiveQueryResponder{
		make(chan amqp.Delivery),
		rabbitChannel,
		ifaceName,
		"dns.query",
	} 
}

func (a *ActiveQueryResponder) Listen() {
	log.Println("Listening for queries")
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("Cannot get hostname", err)
	}
	routingKey := "query"
	queue := "query." + hostname + strconv.Itoa(os.Getpid())
	
	a.rabbitChannel.ExchangeDeclare(a.exchangeName, "topic", false, false, false, false, nil, nil)
	a.rabbitChannel.QueueDeclare(queue, false, true, false, false, nil, nil, nil)
	a.rabbitChannel.QueueBind(queue, routingKey, a.exchangeName, false, nil, nil)
	a.rabbitChannel.Consume(queue, "", false, false, false, false, nil, a.queryChannel, nil)
	for {
		select {
		case rawQuery := <- a.queryChannel:
			query := DNSQuery{}
			err := json.Unmarshal(rawQuery.Body, &query)
			if err != nil {
				log.Println("cannot unmarshal", err)
				continue
			}
			log.Println("Received query for", query)
			if hostname + "." + a.ifaceName + "." == query.NameToResolve {
				a.respond(query)
				rawQuery.Ack(false)
			} else {
				log.Println("Not for me - skipping")
			}
		}
	}
}

func (a *ActiveQueryResponder) respond(query DNSQuery) error {
	ip, err := heartbeat.GetInternalIP(a.ifaceName)
	if err != nil {
		return err
	}
	response :=DNSResponse{
		query.UUID,
		query.NameToResolve,
		ip,
	}
	rawData, err := json.Marshal(&response)
	if err != nil {
		return err
	}
	return a.rabbitChannel.Publish(a.exchangeName,
		"response",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        rawData,
		})
}
