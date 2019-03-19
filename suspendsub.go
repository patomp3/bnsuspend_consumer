package main

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	smslog "github.com/patomp3/smslogs"
	sms "github.com/patomp3/smsservices"
)

// OrderRequest Object
type OrderRequest struct {
	TVSCustomer  string `json:"tvscustomer"`
	OrderTransID string `json:"order_trans_id"`
}

// SubmitOrderRequest for send request to omx
type SubmitOrderRequest struct {
	tvsCustomer           string
	orderID               string
	orderType             int
	channel               string
	effectiveDate         string
	dealerCode            string
	customerID            string
	accountID             string
	subscriberID          string
	OUID                  string
	status                int
	activityReason        string
	userText              string
	payChannelIDPrimary   string
	payChannelIDSecondary string
	resourceCategory      string
	resourceName          string
	valuesArray           string
	identification        string
	identificationType    int
	language              string
	subscriberNumber      string
	subscriberType        string

	title     string
	firstName string
	lastName  string
	gender    int
	houseNo   string
	floor     string

	room               string
	soi                string
	subsoi             string
	moo                string
	amphur             string
	building           string
	streetName         string
	tumbon             string
	city               string
	country            string
	timeAtAddress      string
	typeOfAccomodation string
	zipCode            string
	homePhone          string
}

// SubmitOrderResult for return result from omx
type SubmitOrderResult struct {
	OMXTrackingID string
	RespCode      string
	RespMsg       string
}

// OrderResponse for get response from submit order response
type OrderResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	SOAPENV string   `xml:"SOAP-ENV,attr"`
	Body    struct {
		Text                string `xml:",chardata"`
		SubmitOrderResponse struct {
			Text    string `xml:",chardata"`
			Jms1    string `xml:"jms1,attr"`
			Ns0     string `xml:"ns0,attr"`
			RespMsg struct {
				Text string `xml:",chardata"`
			} `xml:"respMsg"`
			RespCode struct {
				Text string `xml:",chardata"`
			} `xml:"respCode"`
			OMXTrackingId struct {
				Text string `xml:",chardata"`
			} `xml:"OMXTrackingId"`
		} `xml:"submitOrderResponse"`
	} `xml:"Body"`
}

// SubmitOrder for
func (r OrderRequest) SubmitOrder(orderType int) SubmitOrderResult {
	var myReturn SubmitOrderResult

	if orderType == 124 {
		log.Printf("## Process Order Type = %d", orderType)

		var myRequest SubmitOrderRequest

		dbICC := sms.New(cfg.dbICC)

		var suspendResult driver.Rows
		//log.Printf("Execute Store return cursor")
		bResult := dbICC.ExecuteStoreProcedure("begin PK_WFA_BNCONSUMER.GetSuspendSubInfo(:1,:2); end;", r.TVSCustomer, sql.Out{Dest: &suspendResult})
		if bResult && suspendResult != nil {
			log.Printf("## found customer to suspend")

			values := make([]driver.Value, len(suspendResult.Columns()))
			if suspendResult.Next(values) == nil {

				//log.Printf("## in")

				myRequest.tvsCustomer = r.TVSCustomer
				myRequest.channel = values[0].(string)    //fix
				myRequest.orderID = values[1].(string)    //gen
				myRequest.dealerCode = values[2].(string) //fix
				myRequest.orderType = orderType
				myRequest.effectiveDate = time.Now().Format(time.RFC3339Nano)

				myRequest.customerID = values[3].(string)
				myRequest.accountID = values[4].(string)
				myRequest.OUID = values[5].(string)
				myRequest.status, _ = strconv.Atoi(values[6].(string))
				myRequest.activityReason = values[7].(string)
				myRequest.userText = values[8].(string)
				myRequest.payChannelIDPrimary = values[9].(string)
				myRequest.payChannelIDSecondary = values[9].(string)
				myRequest.resourceCategory = values[10].(string)
				myRequest.resourceName = values[11].(string)
				myRequest.valuesArray = values[12].(string)
				//myRequest.amphur = "บางพลัด"
				//myRequest.building = ""
				//myRequest.city = "กรุงเทพมหานคร"
				//myRequest.country = "THA"
				//myRequest.floor = ""
				//myRequest.houseNo = "580/73"
				//myRequest.moo = ""
				//myRequest.room = ""
				//myRequest.soi = "จรัลสนิทวงศ์ 71"
				//myRequest.subsoi = ""
				//myRequest.streetName = "จรัลสนิทวงศ์"
				//myRequest.timeAtAddress = "0101"
				//myRequest.tumbon = "บางพลัด"
				//myRequest.typeOfAccomodation = "OWN"
				//myRequest.zipCode = "10700"
				myRequest.language = values[13].(string)
				myRequest.subscriberID = values[14].(string)
				myRequest.identification = values[15].(string)
				myRequest.identificationType, _ = strconv.Atoi(values[16].(string))
				myRequest.language = values[13].(string)
				myRequest.homePhone = values[17].(string)
				myRequest.title = values[18].(string)
				myRequest.firstName = values[19].(string)
				myRequest.lastName = values[20].(string)
				myRequest.gender, _ = strconv.Atoi(values[21].(string))
				myRequest.subscriberNumber = values[22].(string)
				myRequest.subscriberType = values[23].(string)

				myReturn = submitOrderSuspend(r, myRequest)

				// Write log to stdoutput
				appFunc := "SubmitOrder-124"
				jsonReq, _ := json.Marshal(myRequest)
				jsonRes, _ := json.Marshal(myReturn)

				mLog := smslog.New(cfg.appName)
				mLog.OrderDate = ""
				mLog.OrderNo = ""
				mLog.OrderType = ""
				mLog.TVSNo = ""
				tag := []string{cfg.env, cfg.appName, appFunc, "INFO"}
				mLog.Tags = tag
				mLog.PrintLog(smslog.INFO, appFunc, r.OrderTransID, string(jsonReq), string(jsonRes))
				// End Write log
			} else {
				myReturn.OMXTrackingID = ""
				myReturn.RespCode = "100"
				myReturn.RespMsg = "not found customer info to suspend"
			}
		} else {
			myReturn.OMXTrackingID = ""
			myReturn.RespCode = "100"
			myReturn.RespMsg = "not found customer info to suspend"
		}
	}

	return myReturn
}

// submitOrderSuspend for ...
func submitOrderSuspend(r OrderRequest, req SubmitOrderRequest) SubmitOrderResult {
	var myReturn SubmitOrderResult

	payloadStr := "<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:sub=\"http://services.omx.truecorp.co.th/SubmitOrder\">" +
		"<soapenv:Header/>" +
		"<soapenv:Body>" +
		"<submitOrderRequest xmlns=\"http://services.omx.truecorp.co.th/SubmitOrder\" xmlns:SOAP-ENV=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\">" +
		"<Order>" +
		"<channel>:CHANNEL</channel>" +
		"<orderId>:ORDERID</orderId>" +
		"<orderType>:ORDERTYPE</orderType>" +
		"<effectiveDate>:EFFECTIVEDATE</effectiveDate>" +
		"<dealerCode>:DEALERCODE</dealerCode>" +
		"</Order>" +
		"<Customer>" +
		"<customerId>:CUSTOMERID</customerId>" +
		"<Account>" +
		"<accountId>:ACCOUNTID</accountId>" +
		"</Account>" +
		"<OU>" +
		"<ouId>:OUID</ouId>" +
		"<subscriber>" +
		"<status>:STATUS</status>" +
		"<activityInfo>" +
		"<activityReason>:ACTIVITYREASON</activityReason>" +
		"<userText>:USERTEXT</userText>" +
		"</activityInfo>" +
		"<payChannelIdPrimary>:PAYCHANNELIDPRIMARY</payChannelIdPrimary>" +
		"<payChannelIdSecondary>:PAYCHANNELIDSECONDARY</payChannelIdSecondary>" +
		"<resourceInfo>" +
		"<resourceCategory>:RESOURCECATEGORY</resourceCategory>" +
		"<resourceName>:RESOURCENAME</resourceName>" +
		"<valuesArray>:VALUESARRAY</valuesArray>" +
		"</resourceInfo>" +
		"<subscriberGeneralInfo>" +
		"<language>:LANGUAGE</language>" +
		"<smsLang>:LANGUAGE</smsLang>" +
		"</subscriberGeneralInfo>" +
		"<subscriberId>:SUBSCRIBERID</subscriberId>" +
		"<subscriberIndyName>" +
		"<identification>:IDENTIFICATION</identification>" +
		"<identificationType>:TYPEIDENTIFICATION</identificationType>" +
		"<language>:LANGUAGE</language>" +
		"<homePhone>:HOMEPHONE</homePhone>" +
		"<title>:TITLE</title>" +
		"<firstName>:FIRSTNAME</firstName>" +
		"<lastName>:LASTNAME</lastName>" +
		"<gender>:GENDER</gender>" +
		"</subscriberIndyName>" +
		"<subscriberNumber>:SUBSCRIBERNUMBER</subscriberNumber>" +
		"<subscriberType>:SUBSCRIBERTYPE</subscriberType>" +
		"<ExtendedInfo>" +
		"<name>PRIMARY_RESOURCE_TYPE</name>" +
		"<value>UICC</value>" +
		"</ExtendedInfo>" +
		"</subscriber>" +
		"</OU>" +
		"</Customer>" +
		"</submitOrderRequest>" +
		"</soapenv:Body>" +
		"</soapenv:Envelope>"

	//replace string
	payloadStr = strings.Replace(payloadStr, ":CHANNEL", req.channel, 1)
	payloadStr = strings.Replace(payloadStr, ":ORDERID", req.orderID, 1)
	payloadStr = strings.Replace(payloadStr, ":ORDERTYPE", strconv.Itoa(req.orderType), 1)
	payloadStr = strings.Replace(payloadStr, ":EFFECTIVEDATE", req.effectiveDate, 1)
	payloadStr = strings.Replace(payloadStr, ":DEALERCODE", req.effectiveDate, 1)
	payloadStr = strings.Replace(payloadStr, ":CUSTOMERID", req.customerID, 1)
	payloadStr = strings.Replace(payloadStr, ":ACCOUNTID", req.accountID, 1)
	payloadStr = strings.Replace(payloadStr, ":OUID", req.OUID, 1)
	payloadStr = strings.Replace(payloadStr, ":STATUS", strconv.Itoa(req.status), 1)
	payloadStr = strings.Replace(payloadStr, ":ACTIVITYREASON", req.activityReason, 1)
	payloadStr = strings.Replace(payloadStr, ":USERTEXT", req.userText, 1)
	payloadStr = strings.Replace(payloadStr, ":PAYCHANNELIDPRIMARY", req.payChannelIDPrimary, 1)
	payloadStr = strings.Replace(payloadStr, ":PAYCHANNELIDSECONDARY", req.payChannelIDSecondary, 1)
	payloadStr = strings.Replace(payloadStr, ":RESOURCECATEGORY", req.resourceCategory, 1)
	payloadStr = strings.Replace(payloadStr, ":RESOURCENAME", req.resourceName, 1)
	payloadStr = strings.Replace(payloadStr, ":VALUESARRAY", req.valuesArray, 1)
	payloadStr = strings.Replace(payloadStr, ":SUBSCRIBERID", req.subscriberID, 1)
	payloadStr = strings.Replace(payloadStr, ":IDENTIFICATION", req.identification, 1)
	payloadStr = strings.Replace(payloadStr, ":TYPEIDENTIFICATION", strconv.Itoa(req.identificationType), 1)
	payloadStr = strings.Replace(payloadStr, ":LANGUAGE", req.language, -1)
	payloadStr = strings.Replace(payloadStr, ":GENDER", strconv.Itoa(req.gender), 1)

	payloadStr = strings.Replace(payloadStr, ":HOUSENO", req.houseNo, 1)
	payloadStr = strings.Replace(payloadStr, ":FLOOR", req.floor, 1)
	payloadStr = strings.Replace(payloadStr, ":AMPHUR", req.amphur, 1)
	payloadStr = strings.Replace(payloadStr, ":BUILDING", req.building, 1)
	payloadStr = strings.Replace(payloadStr, ":CITY", req.city, 1)
	payloadStr = strings.Replace(payloadStr, ":COUNTRY", req.country, 1)
	payloadStr = strings.Replace(payloadStr, ":MOO", req.moo, 1)
	payloadStr = strings.Replace(payloadStr, ":ROOM", req.room, 1)
	payloadStr = strings.Replace(payloadStr, ":SOI", req.soi, 1)
	payloadStr = strings.Replace(payloadStr, ":SUBSOI", req.subsoi, 1)
	payloadStr = strings.Replace(payloadStr, ":STREETNAME", req.streetName, 1)
	payloadStr = strings.Replace(payloadStr, ":TIMEATADDRESS", req.timeAtAddress, 1)
	payloadStr = strings.Replace(payloadStr, ":TUMBON", req.tumbon, 1)
	payloadStr = strings.Replace(payloadStr, ":TYPEOFACCOMODATION", req.typeOfAccomodation, 1)
	payloadStr = strings.Replace(payloadStr, ":ZIPCODE", req.zipCode, 1)
	payloadStr = strings.Replace(payloadStr, ":HOMEPHONE", req.homePhone, 1)
	payloadStr = strings.Replace(payloadStr, ":TITLE", req.title, 1)
	payloadStr = strings.Replace(payloadStr, ":FIRSTNAME", req.firstName, 1)
	payloadStr = strings.Replace(payloadStr, ":LASTNAME", req.lastName, 1)

	payload := []byte(payloadStr)

	log.Printf("Post Requerst = %s", string(payload))

	soapAction := "/Services/SubmitOrderOp"
	httpMethod := "POST"

	oReq, err := http.NewRequest(httpMethod, cfg.orderSuspendURL, bytes.NewReader(payload))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
		//return
	}

	oReq.Header.Set("Content-type", "text/xml;charset=UTF-8")
	oReq.Header.Set("SOAPAction", soapAction)
	//req.Header.Set("Accept", "text/xml")

	client := &http.Client{}

	res, err := client.Do(oReq)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
		//return
	}
	defer res.Body.Close()

	//htmlData, err := ioutil.ReadAll(res.Body) //<--- here!

	//fmt.Println()
	//log.Printf("%s", string(htmlData))

	result := new(OrderResponse)
	err = xml.NewDecoder(res.Body).Decode(result)
	if err != nil {
		//log.Fatal("Error on unmarshaling xml. ", err.Error())
		//return
		myReturn.OMXTrackingID = ""
		myReturn.RespCode = "900"
		myReturn.RespMsg = err.Error()
	} else {
		myReturn.OMXTrackingID = result.Body.SubmitOrderResponse.OMXTrackingId.Text
		myReturn.RespCode = result.Body.SubmitOrderResponse.RespCode.Text
		myReturn.RespMsg = result.Body.SubmitOrderResponse.RespMsg.Text
		//log.Printf("%s %s %s", result.Body.SubmitOrderResponse.RespCode.Text, result.Body.SubmitOrderResponse.RespMsg.Text, result.Body.SubmitOrderResponse.OMXTrackingId.Text)
		//acctBalance, _ = strconv.ParseFloat(result.Body.GetAccountBalanceResponse.GetAccountBalanceResult.SearchResult.ArBalanceField.Text, 64)
	}

	// Write log to stdoutput
	appFunc := "SubmitOrderOMX-124"
	jsonReq, _ := json.Marshal(payloadStr)
	jsonRes, _ := json.Marshal(result)

	mLog := smslog.New(cfg.appName)
	mLog.OrderDate = ""
	mLog.OrderNo = ""
	mLog.OrderType = ""
	mLog.TVSNo = ""
	tag := []string{cfg.env, cfg.appName, appFunc, "INFO"}
	mLog.Tags = tag
	mLog.PrintLog(smslog.INFO, appFunc, r.OrderTransID, string(jsonReq), string(jsonRes))
	// End Write log

	// Set Write History ICC
	inCust, _ := strconv.Atoi(req.tvsCustomer)
	wReq := WriteHistoryInfo{CustomerNr: inCust, EventNr: cfg.historyEvent, Reason: cfg.historyReason}
	wRes := wReq.WriteHistory()
	if wRes {
		log.Printf("## Write history success.")
	}

	return myReturn
}
