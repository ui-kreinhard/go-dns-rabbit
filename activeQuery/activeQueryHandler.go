package activeQuery

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/ui-kreinhard/go-dns-rabbit/chamqp"
	"github.com/ui-kreinhard/go-dns-rabbit/utils"
	"log"
	"os"
	"strconv"
)

type ActiveQueryHandler struct {
	rabbitChannel *chamqp.Channel
	responseChannal chan amqp.Delivery
	exchangeName string
	responseUUIDMapping map[string]*chan DNSResponse
	commonOutChannel chan DNSResponse
}

func NewActiveQueryHandler(rabbitChan *chamqp.Channel, commonOutChannel chan DNSResponse) *ActiveQueryHandler {
	return &ActiveQueryHandler{
		rabbitChan,
		make(chan amqp.Delivery),
		"dns.query",
		make(map[string]*chan DNSResponse),
		commonOutChannel,
	}
}

func (a *ActiveQueryHandler) Listen() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("Cannot get valid hostname")
	}
	errChan := make(chan error)

	
	routingKey := "response"
	queue := "response." + hostname + strconv.Itoa(os.Getpid())

	a.rabbitChannel.ExchangeDeclare(a.exchangeName, "topic", false, false, false, false, nil, errChan)
	a.rabbitChannel.QueueDeclare(queue, false, true, false, false, nil, nil, nil)
	a.rabbitChannel.QueueBind(queue, routingKey, a.exchangeName, false, nil, nil)
	a.rabbitChannel.Consume(queue, "", false, false, false, false, nil, a.responseChannal, nil)
	for {
		select {
		case rawResponse := <- a.responseChannal:
			response := DNSResponse{}
			err := json.Unmarshal(rawResponse.Body, &response)
			if err != nil {
				log.Println("Cannot unmarshal", err)
				break
			}
			log.Println("Received response", response)
			outChan := a.responseUUIDMapping[response.UUID]
			
			if outChan == nil {
				log.Println("Not for me")	
			} else {
				*outChan <- response
				rawResponse.Ack(false)	
			}
			a.commonOutChannel <- response
		}
	}
}

type DNSQuery struct {
	UUID string
	NameToResolve string
}

type DNSResponse struct {
	UUID string	
	NameToResolve string
	IP string
}

func (a *ActiveQueryHandler) Query(name string, out chan DNSResponse) error {
	query := DNSQuery{
		utils.GenUUID(),
		name,
	}
	a.responseUUIDMapping[query.UUID] = &out	
	rawData, err := json.Marshal(query)
	if err != nil {
		return err
	}
	
	return a.rabbitChannel.Publish(a.exchangeName,
		"query",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        rawData,
		})
}