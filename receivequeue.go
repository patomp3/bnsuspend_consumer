package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"

	sms "github.com/patomp3/smsservices"
	"github.com/streadway/amqp"
)

// ReceiveQueue struct...
type ReceiveQueue struct {
	URL       string
	QueueName string
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s", msg, err)
	}
}

// Close for
func (r ReceiveQueue) Close() {
	//q.conn.Close()
	//q.ch.Close()
}

// Connect for
func (r ReceiveQueue) Connect() *amqp.Channel {
	conn, err := amqp.Dial(r.URL)
	//defer conn.Close()
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
		return nil
	}

	ch, err := conn.Channel()
	//defer ch.Close()
	if err != nil {
		failOnError(err, "Failed to open a channel")
		return nil
	}

	return ch
}

// Receive for receive message from queue
func (r ReceiveQueue) Receive(ch *amqp.Channel) {

	/*conn, err := amqp.Dial(q.URL)
	defer conn.Close()
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
		return false
	}

	ch, err := conn.Channel()
	defer ch.Close()
	if err != nil {
		failOnError(err, "Failed to open a channel")
		return false
	}*/

	q, err := ch.QueueDeclarePassive(
		r.QueueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
	}

	msgs, err := ch.Consume(
		q.Name, // routing key
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		failOnError(err, "Failed to publish a message")

	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			orderTransID := string(d.Body)
			orderID := d.MessageId

			log.Printf("## Received Order Trans Id : %s", orderTransID)
			log.Printf("## >> Id : %s", orderID)

			//TODO : Process for consumer to
			var orderResult driver.Rows
			var payload map[string]string
			isError := 0
			//log.Printf("Execute Store return cursor")
			dbPED := sms.New(cfg.dbPED)
			bResult := dbPED.ExecuteStoreProcedure("begin PK_WFA_CORE.GetPayloadData(:1,:2); end;", orderTransID, sql.Out{Dest: &orderResult})
			if bResult && orderResult != nil {
				values := make([]driver.Value, len(orderResult.Columns()))
				if orderResult.Next(values) == nil {
					payloadStr := values[1].(string)
					log.Printf("payload = %s", payloadStr)

					err := json.Unmarshal([]byte(payloadStr), &payload)
					if err != nil {
						//panic(err)
						isError = 1
					}
				}
			}

			// not error
			if isError == 0 {
				notifyStatus := ""
				orderReq := OrderRequest{payload["tvscustomer"], orderTransID}
				orderRes := orderReq.SubmitOrder(124)
				if orderRes.OMXTrackingID != "" {
					notifyStatus = "Z"
				} else {
					notifyStatus = "E"
				}

				resStr, _ := json.Marshal(orderRes)
				result := UpdateRequest{orderTransID, orderID, notifyStatus, orderRes.RespCode, orderRes.RespMsg, string(resStr)}
				result.NotifyResult()
				_ = result

			}

		}
	}()

	log.Printf("## [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

/*func (r UpdateRequest) notifyResult() UpdateResponse {
	var resultRes UpdateResponse

	reqPost, _ := json.Marshal(r)

	response, err := http.Post(cfg.updateOrderURL, "application/json", bytes.NewBuffer(reqPost))
	if err != nil {
		log.Printf("The HTTP request failed with error %s", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		//fmt.Println(string(data))
		//myReturn = json.Unmarshal(string(data))
		err = json.Unmarshal(data, &resultRes)
		if err != nil {
			//panic(err)
			log.Printf("The HTTP response failed with error %s", err)
		} else {
			log.Printf("## Result >> %v", resultRes)
		}
	}

	return resultRes
}
*/
