package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type appConfig struct {
	queueName string
	queueURL  string

	dbICC  string
	dbPED  string
	dbATB2 string

	orderSuspendURL string
	updateOrderURL  string
	writeHistoryURL string

	historyuser   int
	historyEvent  int
	historyReason int

	env     string
	appName string
}

var cfg appConfig

func main() {

	log.Printf("##### Service BNSuspend Started #####")

	// For no assign parameter env. using default to Test
	var env string
	if len(os.Args) > 1 {
		env = strings.ToLower(os.Args[1])
	} else {
		env = "development"
	}

	// Load configuration
	viper.SetConfigName("app")    // no need to include file extension
	viper.AddConfigPath("config") // set the path of your config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("## Config file not found. >> %s\n", err.Error())
	} else {
		// read config file
		cfg.queueName = viper.GetString(env + ".queuename")
		cfg.queueURL = viper.GetString(env + ".queueurl")
		cfg.dbICC = viper.GetString(env + ".DBICC")
		cfg.dbATB2 = viper.GetString(env + ".DBATB2")
		cfg.dbPED = viper.GetString(env + ".DBPED")
		cfg.updateOrderURL = viper.GetString(env + ".updateorderurl")
		cfg.orderSuspendURL = viper.GetString(env + ".ordersuspendurl")
		cfg.writeHistoryURL = viper.GetString(env + ".writehistoryurl")

		cfg.historyuser, _ = strconv.Atoi(viper.GetString(env + ".userapp"))
		cfg.historyEvent, _ = strconv.Atoi(viper.GetString(env + ".historyevent"))
		cfg.historyReason, _ = strconv.Atoi(viper.GetString(env + ".historyreason"))

		cfg.env = viper.GetString("env")
		cfg.appName = viper.GetString("appName")

		log.Printf("## Loading Configuration")
		log.Printf("## Env\t= %s", env)
	}

	q := ReceiveQueue{cfg.queueURL, cfg.queueName}
	ch := q.Connect()
	q.Receive(ch)

	/*notifyStatus := ""
	orderRes := SubmitOrder("130003241", 124, "11111")
	if orderRes.OMXTrackingID != "" {
		notifyStatus = "Z"
	} else {
		notifyStatus = "E"
	}
	log.Printf("Notify Result %s", notifyStatus)
	log.Printf("Response %v", orderRes)*/

	//res := SubmitOrder("130003241", 124)
	//log.Printf("%v", res)
	//t := time.Now().Format("2006-01-02T15:04:05-0700")
	//	t, _ := time.Parse("2006-01-02T15:04:05", string(time.Now))
	//fmt.Println(t)

	//t = time.Now().Format(time.RFC3339)
	//fmt.Printf(t)
	/*res := SubmitOrderSuspend("111")
	log.Printf("%v", res)*/

}
