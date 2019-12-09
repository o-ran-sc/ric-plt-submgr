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
	"sync"	
)

var rmrSendMutex = &sync.Mutex{}

var subReqTime time.Duration = 2 * time.Second
var SubDelReqTime time.Duration = 2 * time.Second

type Control struct {
	e2ap        *E2ap
	registry    *Registry
	rtmgrClient *RtmgrClient
	tracker     *Tracker
	rcChan      chan *xapp.RMRParams
	timerMap	*TimerMap
}

type RMRMeid struct {
	PlmnID string
	EnbID  string
	RanName string
}

var seedSN uint16

const (
	CREATE Action = 0
	MERGE  Action = 1
	DELETE Action = 3
)

func init() {
	xapp.Logger.Info("SUBMGR /ric-plt-submgr:r3-test-v2")
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

	timerMap := new(TimerMap)
	timerMap.Init()

	transport := httptransport.New(viper.GetString("rtmgr.HostAddr")+":"+viper.GetString("rtmgr.port"), viper.GetString("rtmgr.baseUrl"), []string{"http"})
	client := rtmgrclient.New(transport, strfmt.Default)
	handle := rtmgrhandle.NewProvideXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	deleteHandle := rtmgrhandle.NewDeleteXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	rtmgrClient := RtmgrClient{client, handle, deleteHandle}

	return Control{new(E2ap), registry, &rtmgrClient, tracker, make(chan *xapp.RMRParams),timerMap}
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
	status := false
	i := 1
	rmrSendMutex.Lock()
	for ; i <= 10 && status == false; i++ { 
		status = xapp.Rmr.Send(params, false)
		if status == false {
			xapp.Logger.Info("rmr.Send() failed. Retry count %v, Mtype: %v, SubId: %v, Xid %s",i, params.Mtype, params.SubId, params.Xid)
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = errors.New("rmr.Send() failed")
		xapp.Rmr.Free(params.Mbuf)
	}
	rmrSendMutex.Unlock()
	
	/*
	if !xapp.Rmr.Send(params, false) {
		err = errors.New("rmr.Send() failed")
		xapp.Rmr.Free(params.Mbuf)
	}
	*/	
	return
}

func (c *Control) rmrReplyToSender(params *xapp.RMRParams) (err error) {
	c.rmrSend(params)
	return
}

func (c *Control) controlLoop() {
	for {
		msg := <-c.rcChan
		switch msg.Mtype {
		case xapp.RICMessageTypes["RIC_SUB_REQ"]:
			go c.handleSubscriptionRequest(msg)
		case xapp.RICMessageTypes["RIC_SUB_RESP"]:
			go c.handleSubscriptionResponse(msg)
		case xapp.RICMessageTypes["RIC_SUB_FAILURE"]:
			go c.handleSubscriptionFailure(msg)
		case xapp.RICMessageTypes["RIC_SUB_DEL_REQ"]:
			go c.handleSubscriptionDeleteRequest(msg)
		case xapp.RICMessageTypes["RIC_SUB_DEL_RESP"]:
			go c.handleSubscriptionDeleteResponse(msg)
		default:
			err := errors.New("Message Type " + strconv.Itoa(msg.Mtype) + " is discarded")
			xapp.Logger.Error("Unknown message type: %v", err)
		}
	}
}

func (c *Control) handleSubscriptionRequest(params *xapp.RMRParams) (err error) {
	xapp.Logger.Info("Subscription Request Received from Src: %s, Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Src, params.Mtype, params.SubId, params.Xid, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	/* Reserve a sequence number and set it in the payload */
	newSubId, isIdValid := c.registry.ReserveSequenceNumber()
	if isIdValid != true {
		xapp.Logger.Info("Further processing of this SubscriptionRequest stopped. SubId: %v, Xid: %s",params.SubId, params.Xid)
		return 
	}

	err = c.e2ap.SetSubscriptionRequestSequenceNumber(params.Payload, newSubId)
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
	xapp.Logger.Info("Starting routing manager update")
	c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)

	//time.Sleep(3 * time.Second)

	// Setting new subscription ID in the RMR header
	params.SubId = int(newSubId)
	xapp.Logger.Info("Forwarding Subscription Request to E2T: Mtype: %v, SubId: %v, Xid %s, Meid %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("Failed to send request to E2T %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	} /*else {
		c.timerMap.StartTimer(newSubId, subReqTime, c.handleSubscriptionRequestTimer)
	}*/
	xapp.Logger.Debug("--- Debugging transaction table = %v", c.tracker.transactionTable)
	return
}

func (c *Control) handleSubscriptionResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Info("Subscription Response Received from Src: %s, Mtype: %v, SubId: %v, Meid: %v",params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionResponseSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}

	xapp.Logger.Info("Received payloadSeqNum: %v",payloadSeqNum)
	if !c.registry.IsValidSequenceNumber(payloadSeqNum) {
		err = errors.New("Unknown Subscription ID: " + strconv.Itoa(int(payloadSeqNum)) + " in Subscritpion Response. Message discarded.")
		return
	}

//	c.timerMap.StopTimer(payloadSeqNum)

	c.registry.setSubscriptionToConfirmed(payloadSeqNum)
	var transaction Transaction
	transaction, err = c.tracker.RetriveTransaction(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Error("Failed to retrive transaction record. Err: %v", err)
		xapp.Logger.Info("Further processing of this Subscription Response stopped. SubId: %v, Xid: %s",params.SubId, params.Xid)
		return
	}
	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Retrieved old subId...", int(payloadSeqNum), transaction.XappInstanceAddress, transaction.XappPort)

    params.SubId = int(payloadSeqNum)
    params.Xid = transaction.OrigParams.Xid
	
	xapp.Logger.Info("Forwarding Subscription Response to UEEC: Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("Failed to send response to requestor %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Deleting transaction record", int(payloadSeqNum), transaction.XappInstanceAddress, transaction.XappPort)
	transaction, err = c.tracker.completeTransaction(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Error("Failed to delete a Subscription Request transaction record due to %v", err)
		return
	}
	return
}

func (c *Control) handleSubscriptionFailure(params *xapp.RMRParams) (err error) {
	xapp.Logger.Info("Subscription Failure Received from Src: %s, Mtype: %v, SubId: %v, Meid: %v",params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionFailureSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Received payloadSeqNum: %v", payloadSeqNum)

	// should here be IsValidSequenceNumber check?

//	c.timerMap.StopTimer(payloadSeqNum)

	var transaction Transaction
	transaction, err = c.tracker.RetriveTransaction(payloadSeqNum, CREATE)
	if  err != nil {
		xapp.Logger.Error("Failed to retrive transaction record. Err %v", err)
		xapp.Logger.Info("Further processing of this Subscription Failure stopped. SubId: %v, Xid: %s",params.SubId, params.Xid)
		return
	}
	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Forwarding response to requestor...", int(payloadSeqNum), transaction.XappInstanceAddress, transaction.XappPort)

	params.SubId = int(payloadSeqNum)
	params.Xid = transaction.OrigParams.Xid

	xapp.Logger.Info("Forwarding Subscription Failure to UEEC: Mtype: %v, SubId: %v, Xid: %v, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("Failed to send response to requestor %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("Starting routing manager update")
	subRouteAction := SubRouteInfo{CREATE, transaction.XappInstanceAddress, transaction.XappPort, payloadSeqNum}
	c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)

	xapp.Logger.Info("Deleting trancaction record")
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, CREATE)
		if err != nil {
			xapp.Logger.Error("Failed to delete a Subscription Request transaction record due to %v", err)
			return
		}
	}
	return
}

func (c *Control) handleSubscriptionRequestTimer(subId uint16) {
	xapp.Logger.Info("Subscription Request timer expired. SubId: %v",subId)
/*	
	transaction, err := c.tracker.completeTransaction(subId, CREATE)
	if err != nil {
		xapp.Logger.Error("Failed to delete a Subscription Request transaction record due to %v", err)
		return
	}
	xapp.Logger.Info("SubId: %v, Xid %v, Meid: %v",subId, transaction.OrigParams.Xid, transaction.OrigParams.Meid)

	var params xapp.RMRParams
	params.Mtype = 12012 //xapp.RICMessageTypes["RIC_SUB_FAILURE"]
	params.SubId = int(subId)
	params.Meid = transaction.OrigParams.Meid
	params.Xid = transaction.OrigParams.Xid
	payload := []byte("40C9408098000003EA7E00050000010016EA6300020021EA6E00808180EA6F000400000000EA6F000400010040EA6F000400020080EA6F0004000300C0EA6F000400040100EA6F000400050140EA6F000400060180EA6F0004000701C0EA6F000400080200EA6F000400090240EA6F0004000A0280EA6F0004000B02C0EA6F0004000C0300EA6F0004000D0340EA6F0004000E0380EA6F0004000F03C0")
	params.PayloadLen = len(payload)
	params.Payload = payload

	xapp.Logger.Info("Forwarding Subscription Failure to UEEC: Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(&params)
	if err != nil {
		xapp.Logger.Error("Failed to send response to requestor %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}
*/
/*
	time.Sleep(3 * time.Second)

	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Deleting transaction record", int(subId), transaction.XappInstanceAddress, transaction.XappPort)

	xapp.Logger.Info("Starting routing manager update")
	subRouteAction := SubRouteInfo{DELETE, transaction.XappInstanceAddress, transaction.XappPort, payloadSeqNum}
	c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)

	xapp.Logger.Info("Deleting trancaction record")
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, CREATE)
		if err != nil {
			xapp.Logger.Error("Failed to delete a Subscription Request transaction record due to %v", err)
			return
		}
	}
*/
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
	xapp.Logger.Info("Subscription Delete Request Received from Src: %s, Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Src, params.Mtype, params.SubId, params.Xid, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteRequestSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Received payloadSeqNum: %v", payloadSeqNum)
	if c.registry.IsValidSequenceNumber(payloadSeqNum) {
		c.registry.deleteSubscription(payloadSeqNum)
		trackErr := c.trackDeleteTransaction(params, payloadSeqNum)
		if trackErr != nil {
			xapp.Logger.Error("Failed to create a Subscription Delete Request transaction record due to %v", trackErr)
			return trackErr
		}
	}

	xapp.Logger.Info("Forwarding Delete Subscription Request to E2T: Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("Failed to send request to E2T %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	} /*else {
		c.timerMap.StartTimer(payloadSeqNum, SubDelReqTime, c.handleSubscriptionDeleteRequestTimer)
	}*/
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
	xapp.Logger.Info("Subscription Delete Response Received from Src: %s, Mtype: %v, SubId: %v, Meid: %v",params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteResponseSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Received payloadSeqNum: %v", payloadSeqNum)

	// should here be IsValidSequenceNumber check?
//	c.timerMap.StopTimer(payloadSeqNum)
	
	var transaction Transaction
	transaction, err = c.tracker.RetriveTransaction(payloadSeqNum, DELETE)
	if  err != nil {
		xapp.Logger.Error("Failed to retrive transaction record. Err %v", err)
		xapp.Logger.Info("Further processing of this Subscription Delete Response stopped. SubId: %v, Xid: %s",params.SubId, params.Xid)
		return
	}
	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Forwarding response to requestor...", int(payloadSeqNum), transaction.XappInstanceAddress, transaction.XappPort)

    params.SubId = int(payloadSeqNum)
    params.Xid = transaction.OrigParams.Xid
	xapp.Logger.Info("Forwarding Subscription Delete Response to UEEC: Mtype: %v, SubId: %v, Xid: %v, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("Failed to send response to requestor %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
//		return
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("Starting routing manager update")
	subRouteAction := SubRouteInfo{DELETE, transaction.XappInstanceAddress, transaction.XappPort, payloadSeqNum}
	c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)

	xapp.Logger.Info("Deleting trancaction record")
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
		if err != nil {
			xapp.Logger.Error("Failed to delete a Subscription Delete Request transaction record due to %v", err)
			return
		}
	}
	return
}

func (c *Control) handleSubscriptionDeleteFailure(params *xapp.RMRParams) (err error) {
	xapp.Logger.Info("Subscription Delete Failure Received from Src: %s, Mtype: %v, SubId: %v, Meid: %v",params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteFailureSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Received payloadSeqNum: %v", payloadSeqNum)

	// should here be IsValidSequenceNumber check?
//	c.timerMap.StopTimer(payloadSeqNum)

	var transaction Transaction
	transaction, err = c.tracker.RetriveTransaction(payloadSeqNum, DELETE)
	if  err != nil {
		xapp.Logger.Error("Failed to retrive transaction record. Err %v", err)
		xapp.Logger.Info("Further processing of this Subscription Delete Failure stopped. SubId: %v, Xid: %s",params.SubId, params.Xid)
		return
	}
	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Forwarding response to requestor...", int(payloadSeqNum), transaction.XappInstanceAddress, transaction.XappPort)

    params.SubId = int(payloadSeqNum)
    params.Xid = transaction.OrigParams.Xid
	xapp.Logger.Info("Forwarding Subscription Delete Failure to UEEC: Mtype: %v, SubId: %v, Xid: %v, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("Failed to send response to requestor %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
//		return
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("Starting routing manager update")
	subRouteAction := SubRouteInfo{DELETE, transaction.XappInstanceAddress, transaction.XappPort, payloadSeqNum}
	c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)

	xapp.Logger.Info("Deleting trancaction record")
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
		if err != nil {
			xapp.Logger.Error("Failed to delete a Subscription Delete Request transaction record due to %v", err)
			return
		}
	}
	return
}

func (c *Control) handleSubscriptionDeleteRequestTimer(subId uint16) {
	xapp.Logger.Info("Subscription Delete Request timer expired. SubId: %v",subId)
/*	
	transaction, err := c.tracker.completeTransaction(subId, DELETE)
	if err != nil {
		xapp.Logger.Error("Failed to delete a Subscription Delete Request transaction record due to %v", err)
		return
	}
	xapp.Logger.Info("SubId: %v, Xid %v, Meid: %v",subId, transaction.OrigParams.Xid, transaction.OrigParams.Meid)

	var params xapp.RMRParams
	params.Mtype = 12022 //xapp.RICMessageTypes["RIC_SUB_DEL_FAILURE"]
	params.SubId = int(subId)
	params.Meid = transaction.OrigParams.Meid
	params.Xid = transaction.OrigParams.Xid
	payload := []byte("40CA4018000003EA7E00050000010016EA6300020021EA74000200C0")
	params.PayloadLen = len(payload)
	params.Payload = payload

	xapp.Logger.Info("Forwarding Subscription Delete Failure to UEEC: Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(&params)
	if err != nil {
		xapp.Logger.Error("Failed to send response to requestor %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}
*/	
/*
	time.Sleep(3 * time.Second)
	xapp.Logger.Info("Subscription ID: %v, from address: %v:%v. Deleting transaction record", int(subId), transaction.XappInstanceAddress, transaction.XappPort)

	xapp.Logger.Info("Starting routing manager update")
	subRouteAction := SubRouteInfo{DELETE, transaction.XappInstanceAddress, transaction.XappPort, payloadSeqNum}
	c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)

	xapp.Logger.Info("Deleting trancaction record")
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
		if err != nil {
			xapp.Logger.Error("Failed to delete a Subscription Delete Request transaction record due to %v", err)
			return
		}
	}
*/
	return
	}
