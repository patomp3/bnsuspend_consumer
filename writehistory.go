package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// WriteHistoryInfo Object
type WriteHistoryInfo struct {
	ByUser struct {
		ByChannel string `json:"byChannel"`
		ByHost    string `json:"byHost"`
		ByProject string `json:"byProject"`
		ByUser    string `json:"byUser"`
	} `json:"ByUser"`
	CustomerNr int `json:"CustomerNr"`
	EventNr    int `json:"EventNr"`
	Reason     int `json:"Reason"`
}

// ResultRes object
type ResultRes struct {
	ErrorCode   int    `json:"ErrorCode"`
	ErrorDesc   string `json:"ErrorDesc"`
	ResultValue string `json:"ResultValue"`
}

// New for create write history info
/*func New(customerNr int, eventNr int, reason int) *WriteHistoryInfo {
	return &WriteHistoryInfo{CustomerNr: customerNr, EventNr: eventNr, Reason: reason}
}*/

// WriteHistory to create histiry icc
func (w WriteHistoryInfo) WriteHistory() bool {
	myReturn := false

	//## Write user
	var myResult ResultRes
	w.ByUser.ByUser = strconv.Itoa(cfg.historyuser)

	jsonValue, _ := json.Marshal(w)
	log.Printf("%s", string(jsonValue))
	response, err := http.Post(cfg.writeHistoryURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Printf("The HTTP request failed with error %s", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
		//myReturn = json.Unmarshal(string(data))
		err = json.Unmarshal(data, &myResult)
		if err != nil {
			panic(err)
		}
		if myResult.ErrorCode == 0 {
			myReturn = true
		}
	}

	return myReturn
}
