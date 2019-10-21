/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package control

import "C"

import (
	"errors"
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	rtmgrhandle "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client/handle"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
	"math/rand"
	"strconv"
	"time"
)

type Control struct {
	e2ap        *E2ap
	registry    *Registry
	rtmgrClient *RtmgrClient
	tracker     *Tracker
	rcChan      chan *xapp.RMRParams
}

type RMRMeid struct {
	PlmnID string
	EnbID  string
}

var seedSN uint16
var SubscriptionReqChan = make(chan SubRouteInfo, 10)

const (
	CREATE Action = 0
	MERGE  Action = 1
	DELETE Action = 3
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("submgr")
	viper.AllowEmptyEnv(true)
	seedSN = uint16(viper.GetInt("seed_sn"))
	if seedSN == 0 {
		rand.Seed(time.Now().UnixNano())
		seedSN = uint16(rand.Intn(65535))
	}
	if seedSN > 65535 {
		seedSN = 0
	}
	xapp.Logger.Info("SUBMGR: Initial Sequence Number: %v", seedSN)
}

func NewControl() Control {
	registry := new(Registry)
	registry.Initialize(seedSN)

	tracker := new(Tracker)
	tracker.Init()

	transport := httptransport.New(viper.GetString("rtmgr.HostAddr")+":"+viper.GetString("rtmgr.port"), viper.GetString("rtmgr.baseUrl"), []string{"http"})
	client := rtmgrclient.New(transport, strfmt.Default)
	handle := rtmgrhandle.NewProvideXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	deleteHandle := rtmgrhandle.NewDeleteXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	rtmgrClient := RtmgrClient{client, handle, deleteHandle}

	return Control{new(E2ap), registry, &rtmgrClient, tracker, make(chan *xapp.RMRParams)}
}

func (c *Control) Run() {
	go c.controlLoop()
	xapp.Run(c)
}

func (c *Control) Consume(rp *xapp.RMRParams) (err error) {
	c.rcChan <- rp
	return
}

func (c *Control) rmrSend(params *xapp.RMRParams) (err error) {
	if !xapp.Rmr.Send(params, false) {
		err = errors.New("rmr.Send() failed")
	}
	return
}

func (c *Control) rmrReplyToSender(params *xapp.RMRParams) (err error) {
	if !xapp.Rmr.Send(params, true) {
		err = errors.New("rmr.Send() failed")
	}
	return
}

func (c *Control) controlLoop() {
	for {
		msg := <-c.rcChan
		switch msg.Mtype {
		case xapp.RICMessageTypes["RIC_SUB_REQ"]:
			c.handleSubscriptionRequest(msg)
		case xapp.RICMessageTypes["RIC_SUB_RESP"]:
			c.handleSubscriptionResponse(msg)
		case xapp.RICMessageTypes["RIC_SUB_DEL_REQ"]:
			c.handleSubscriptionDeleteRequest(msg)
		case xapp.RICMessageTypes["RIC_SUB_DEL_RESP"]:
			c.handleSubscriptionDeleteResponse(msg)
		default:
			err := errors.New("Message Type " + strconv.Itoa(msg.Mtype) + " is discarded")
			xapp.Logger.Error("Unknown message type: %v", err)
		}
	}
}

func (c *Control) handleSubscriptionRequest(params *xapp.RMRParams) (err error) {
	payloadSeqNum, err := c.e2ap.GetSubscriptionRequestSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Request Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payloadSeqNum)

	/* Reserve a sequence number and set it in the payload */
	newSubId := c.registry.ReserveSequenceNumber()

	_, err = c.e2ap.SetSubscriptionRequestSequenceNumber(params.Payload, newSubId)
	if err != nil {
		err = errors.New("Unable to set Subscription Sequence Number in Payload due to: " + err.Error())
		return
	}

	srcAddr, srcPort, err := c.rtmgrClient.SplitSource(params.Src)
	if err != nil {
		xapp.Logger.Error("Failed to update routing-manager about the subscription request with reason: %s", err)
		return
	}

	/* Create transatcion records for every subscription request */
	xactKey := TransactionKey{newSubId, CREATE}
	xactValue := Transaction{*srcAddr, *srcPort, params}
	err = c.tracker.TrackTransaction(xactKey, xactValue)
	if err != nil {
		xapp.Logger.Error("Failed to create a Subscription Request transaction record due to %v", err)
		return
	}

	/* Update routing manager about the new subscription*/
	subRouteAction := SubRouteInfo{CREATE, *srcAddr, *srcPort, newSubId}
	go c.rtmgrClient.SubscriptionRequestUpdate()
	SubscriptionReqChan <- subRouteAction

	// Setting new subscription ID in the RMR header
	params.SubId = int(newSubId)

	xapp.Logger.Info("Generated ID: %v. Forwarding to E2 Termination...", int(newSubId))
	c.rmrSend(params)
	xapp.Logger.Debug("--- Debugging transaction table = %v", c.tracker.transactionTable)
	return
}

func (c *Control) handleSubscriptionResponse(params *xapp.RMRParams) (err error) {
	payloadSeqNum, err := c.e2ap.GetSubscriptionResponseSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Response Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payloadSeqNum)
	if !c.registry.IsValidSequenceNumber(payloadSeqNum) {
		err = errors.New("Unknown Subscription ID: " + strconv.Itoa(int(payloadSeqNum)) + " in Subscritpion Response. Message discarded.")
		return
	}
	c.registry.setSubscriptionToConfirmed(payloadSeqNum)
	xapp.Logger.Info("Subscription Response Registered. Forwarding to Requestor...")
	transaction, err := c.tracker.completeTransaction(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Error("Failed to delete a Subscription Request transaction record due to %v", err)
		return
	}
	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Forwarding to E2 Termination...", int(payloadSeqNum), transaction.XappInstanceAddress, transaction.XappPort)
	params.Mbuf = transaction.OrigParams.Mbuf
	c.rmrReplyToSender(params)
	return
}

func (act Action) String() string {
	actions := [...]string{
		"CREATE",
		"MERGE",
		"DELETE",
	}

	if act < CREATE || act > DELETE {
		return "Unknown"
	}
	return actions[act]
}

func (act Action) valid() bool {
	switch act {
	case CREATE, MERGE, DELETE:
		return true
	default:
		return false
	}
}

func (c *Control) handleSubscriptionDeleteRequest(params *xapp.RMRParams) (err error) {
	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteRequestSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Delete Request Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payloadSeqNum)
	if c.registry.IsValidSequenceNumber(payloadSeqNum) {
		c.registry.deleteSubscription(payloadSeqNum)
		trackErr := c.trackDeleteTransaction(params, payloadSeqNum)
		if trackErr != nil {
			xapp.Logger.Error("Failed to create a Subscription Delete Request transaction record due to %v", trackErr)
			return trackErr
		}
	}
	xapp.Logger.Info("Subscription ID: %v. Forwarding to E2 Termination...", int(payloadSeqNum))
	c.rmrSend(params)
	return
}

func (c *Control) trackDeleteTransaction(params *xapp.RMRParams, payloadSeqNum uint16) (err error) {
	srcAddr, srcPort, err := c.rtmgrClient.SplitSource(params.Src)
	if err != nil {
		xapp.Logger.Error("Failed to update routing-manager about the subscription delete request with reason: %s", err)
	}
	xactKey := TransactionKey{payloadSeqNum, DELETE}
	xactValue := Transaction{*srcAddr, *srcPort, params}
	err = c.tracker.TrackTransaction(xactKey, xactValue)
	return
}

func (c *Control) handleSubscriptionDeleteResponse(params *xapp.RMRParams) (err error) {
	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteResponseSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	var transaction, _ = c.tracker.RetriveTransaction(payloadSeqNum, DELETE)
	subRouteAction := SubRouteInfo{DELETE, transaction.XappInstanceAddress, transaction.XappPort, payloadSeqNum}
	go c.rtmgrClient.SubscriptionRequestUpdate()
	SubscriptionReqChan <- subRouteAction

	xapp.Logger.Info("Subscription Delete Response Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payloadSeqNum)
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
		if err != nil {
			xapp.Logger.Error("Failed to delete a Subscription Delete Request transaction record due to %v", err)
			return
		}
		xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Forwarding to E2 Termination...", int(payloadSeqNum), transaction.XappInstanceAddress, transaction.XappPort)
		//params.Src = xAddress + ":" + strconv.Itoa(int(xPort))
		params.Mbuf = transaction.OrigParams.Mbuf
		c.rmrReplyToSender(params)
	}
	return
}
