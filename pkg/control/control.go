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

import (
	"errors"
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	rtmgrhandle "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client/handle"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
	"math/rand"
	"sync"
	"time"
)

var subReqTime time.Duration = 5 * time.Second
var subDelReqTime time.Duration = 5 * time.Second
var maxSubReqTryCount uint64 = 2    // Initial try + retry
var maxSubDelReqTryCount uint64 = 2 // Initial try + retry

type Control struct {
	e2ap         *E2ap
	registry     *Registry
	rtmgrClient  *RtmgrClient
	tracker      *Tracker
	timerMap     *TimerMap
	rmrSendMutex sync.Mutex
	msgCounter   uint64
}

type RMRMeid struct {
	PlmnID  string
	EnbID   string
	RanName string
}

var seedSN uint16

const (
	CREATE Action = 0
	MERGE  Action = 1
	NONE   Action = 2
	DELETE Action = 3
)

func init() {
	xapp.Logger.Info("SUBMGR")
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

func NewControl() *Control {
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

	return &Control{e2ap: new(E2ap),
		registry:    registry,
		rtmgrClient: &rtmgrClient,
		tracker:     tracker,
		timerMap:    timerMap,
		msgCounter:  0,
	}
}

func (c *Control) Run() {
	xapp.Run(c)
}

func (c *Control) rmrSend(params *xapp.RMRParams) (err error) {
	status := false
	i := 1
	for ; i <= 10 && status == false; i++ {
		c.rmrSendMutex.Lock()
		status = xapp.Rmr.Send(params, false)
		c.rmrSendMutex.Unlock()
		if status == false {
			xapp.Logger.Info("rmr.Send() failed. Retry count %v, Mtype: %v, SubId: %v, Xid %s", i, params.Mtype, params.SubId, params.Xid)
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = errors.New("rmr.Send() failed")
		xapp.Rmr.Free(params.Mbuf)
	}
	return
}

func (c *Control) rmrReplyToSender(params *xapp.RMRParams) (err error) {
	c.rmrSend(params)
	return
}

func (c *Control) Consume(msg *xapp.RMRParams) (err error) {
	c.msgCounter++
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
	case xapp.RICMessageTypes["RIC_SUB_DEL_FAILURE"]:
		go c.handleSubscriptionDeleteFailure(msg)
	default:
		xapp.Logger.Info("Unknown Message Type '%d', discarding", msg.Mtype)
	}
	return nil
}

func (c *Control) handleSubscriptionRequest(params *xapp.RMRParams) {
	xapp.Logger.Info("SubReq received from Src: %s, Mtype: %v, SubId: %v, Xid: %s, Meid: %v", params.Src, params.Mtype, params.SubId, params.Xid, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	/* Reserve a sequence number and set it in the payload */
	subs := c.registry.ReserveSubscription()
	if subs == nil {
		xapp.Logger.Error("SubReq: Failed to reserve sequence number. Dropping this msg. SubId: %v, Xid: %s", params.SubId, params.Xid)
		return
	}

	params.SubId = int(subs.Seq)
	err := c.e2ap.SetSubscriptionRequestSequenceNumber(params.Payload, subs.Seq)
	if err != nil {
		xapp.Logger.Error("SubReq: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		c.registry.releaseSequenceNumber(subs.Seq)
		return
	}

	srcAddr, srcPort, err := c.rtmgrClient.SplitSource(params.Src)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to update routing-manager. Dropping this msg. Err: %s, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		c.registry.releaseSequenceNumber(subs.Seq)
		return
	}

	// Create transatcion record for every subscription request
	var forwardRespToXapp bool = true
	var responseReceived bool = false
	transaction, err := c.tracker.TrackTransaction(subs.Seq, CREATE, *srcAddr, *srcPort, params, responseReceived, forwardRespToXapp)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to create transaction record. Dropping this msg. Err: %v SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		c.registry.releaseSequenceNumber(subs.Seq)
		return
	}

	// Update routing manager about the new subscription
	subRouteAction := transaction.SubRouteInfo()
	xapp.Logger.Info("SubReq: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)

	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to update routing manager. Dropping this msg. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		c.registry.releaseSequenceNumber(subs.Seq)
		return
	}

	// Setting new subscription ID in the RMR header
	xapp.Logger.Info("SubReq: Forwarding SubReq to E2T: Mtype: %v, SubId: %v, Xid %s, Meid %v", params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to send request to E2T %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}
	c.timerMap.StartTimer("RIC_SUB_REQ", int(subs.Seq), subReqTime, FirstTry, c.handleSubscriptionRequestTimer)
	xapp.Logger.Debug("SubReq: Debugging transaction table = %v", c.tracker.transactionTable)
	return
}

func (c *Control) handleSubscriptionResponse(params *xapp.RMRParams) {
	xapp.Logger.Info("SubResp received from Src: %s, Mtype: %v, SubId: %v, Meid: %v", params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionResponseSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubResp: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		return
	}
	xapp.Logger.Info("SubResp: Received payloadSeqNum: %v", payloadSeqNum)

	if !c.registry.IsValidSequenceNumber(payloadSeqNum) {
		xapp.Logger.Error("SubResp: Unknown payloadSeqNum. Dropping this msg. PayloadSeqNum: %v, SubId: %v", payloadSeqNum, params.SubId)
		return
	}

	c.timerMap.StopTimer("RIC_SUB_REQ", int(payloadSeqNum))

	transaction, responseReceived, err := c.tracker.CheckResponseReceived(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Info("SubResp: Dropping this msg. Err: %v SubId: %v", err, payloadSeqNum)
		return
	}

	if responseReceived == true {
		// Subscription timer already received
		return
	}
	xapp.Logger.Info("SubResp: SubId: %v, from address: %v:%v.", payloadSeqNum, transaction.Xappkey.Addr, transaction.Xappkey.Port)

	c.registry.setSubscriptionToConfirmed(payloadSeqNum)

	params.SubId = int(payloadSeqNum)
	params.Xid = transaction.OrigParams.Xid

	xapp.Logger.Info("SubResp: Forwarding Subscription Response to xApp Mtype: %v, SubId: %v, Meid: %v", params.Mtype, params.SubId, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("SubResp: Failed to send response to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	xapp.Logger.Info("SubResp: SubId: %v, from address: %v:%v. Deleting transaction record", payloadSeqNum, transaction.Xappkey.Addr, transaction.Xappkey.Port)
	_, err = c.tracker.completeTransaction(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Error("SubResp: Failed to delete transaction record. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}
	return
}

func (c *Control) handleSubscriptionFailure(params *xapp.RMRParams) {
	xapp.Logger.Info("SubFail received from Src: %s, Mtype: %v, SubId: %v, Meid: %v", params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionFailureSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubFail: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		return
	}
	xapp.Logger.Info("SubFail: Received payloadSeqNum: %v", payloadSeqNum)

	c.timerMap.StopTimer("RIC_SUB_REQ", int(payloadSeqNum))

	transaction, responseReceived, err := c.tracker.CheckResponseReceived(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Info("SubFail: Dropping this msg. Err: %v SubId: %v", err, payloadSeqNum)
		return
	}

	if responseReceived == true {
		// Subscription timer already received
		return
	}
	xapp.Logger.Info("SubFail: SubId: %v, from address: %v:%v. Forwarding response to xApp", payloadSeqNum, transaction.Xappkey.Addr, transaction.Xappkey.Port)

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("SubFail: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := transaction.SubRouteInfo()
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubFail: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	xapp.Logger.Info("SubFail: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		_, err = c.tracker.completeTransaction(payloadSeqNum, CREATE)
		if err != nil {
			xapp.Logger.Error("SubFail: Failed to delete transaction record. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
			return
		}
	} else {
		xapp.Logger.Error("SubFail: Failed to release sequency number. SubId: %v, Xid: %s", params.SubId, params.Xid)
		return
	}
	return
}

func (c *Control) handleSubscriptionRequestTimer(strId string, nbrId int, tryCount uint64) {
	subId := uint16(nbrId)
	xapp.Logger.Info("handleSubTimer: SubReq timer expired. subId: %v,  tryCount: %v", subId, tryCount)

	transaction, responseReceived, err := c.tracker.CheckResponseReceived(subId, CREATE)
	if err != nil {
		xapp.Logger.Info("handleSubTimer: Dropping this timer action. Err: %v SubId: %v", err, subId)
		return
	}

	if responseReceived == true {
		// Subscription Response or Failure already received
		return
	}

	if tryCount < maxSubReqTryCount {
		xapp.Logger.Info("handleSubTimer: Resending SubReq to E2T: Mtype: %v, SubId: %v, Xid %s, Meid %v", transaction.OrigParams.Mtype, transaction.OrigParams.SubId, transaction.OrigParams.Xid, transaction.OrigParams.Meid)
		// Set possible to handle new response for the subId
		err = c.tracker.RetryTransaction(subId, CREATE)
		if err != nil {
			xapp.Logger.Error("handleSubDelTimer: Failed to retry transaction record. Dropping timer action. Err %v, SubId: %v", err, transaction.OrigParams.SubId)
			return
		}

		err = c.rmrSend(transaction.OrigParams)
		if err != nil {
			xapp.Logger.Error("handleSubTimer: Failed to send request to E2T %v, SubId: %v, Xid: %s", err, transaction.OrigParams.SubId, transaction.OrigParams.Xid)
		}

		tryCount++
		c.timerMap.StartTimer("RIC_SUB_REQ", int(subId), subReqTime, tryCount, c.handleSubscriptionRequestTimer)
		return
	}

	var subDelReqPayload []byte
	subDelReqPayload, err = c.e2ap.PackSubscriptionDeleteRequest(transaction.OrigParams.Payload, subId)
	if err != nil {
		xapp.Logger.Error("handleSubTimer: Packing SubDelReq failed. Err: %v", err)
		return
	}

	// Cancel failed subscription
	var params xapp.RMRParams
	params.Mtype = 12020 // RIC SUBSCRIPTION DELETE
	params.SubId = int(subId)
	params.Xid = transaction.OrigParams.Xid
	params.Meid = transaction.OrigParams.Meid
	params.Src = transaction.OrigParams.Src
	params.PayloadLen = len(subDelReqPayload)
	params.Payload = subDelReqPayload
	params.Mbuf = nil

	// Delete CREATE transaction
	_, err = c.tracker.completeTransaction(subId, CREATE)
	if err != nil {
		xapp.Logger.Error("handleSubTimer: Failed to delete create transaction record. Dropping this timer action. Err: %v, SubId: %v, Xid: %s", err, subId, params.Xid)
		return
	}

	// Create DELETE transaction
	var forwardRespToXapp bool = false
	_, err = c.trackDeleteTransaction(&params, subId, forwardRespToXapp)
	if err != nil {
		xapp.Logger.Error("handleSubTimer: Failed to create delete transaction record. Dropping this timer action. Err: %v, SubId: %v, Xid: %s", err, subId, params.Xid)
		return
	}

	xapp.Logger.Info("handleSubTimer: Sending SubDelReq to E2T: Mtype: %v, SubId: %v, Meid: %v", params.Mtype, params.SubId, params.Meid)
	c.rmrSend(&params)
	if err != nil {
		xapp.Logger.Error("handleSubTimer: Failed to send request to E2T %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}
	c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subId), subDelReqTime, FirstTry, c.handleSubscriptionDeleteRequestTimer)
	return
}

func (act Action) String() string {
	actions := [...]string{
		"CREATE",
		"MERGE",
		"NONE",
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

func (c *Control) handleSubscriptionDeleteRequest(params *xapp.RMRParams) {
	xapp.Logger.Info("SubDelReq received from Src: %s, Mtype: %v, SubId: %v, Xid: %s, Meid: %v", params.Src, params.Mtype, params.SubId, params.Xid, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteRequestSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelReq: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		return
	}
	xapp.Logger.Info("SubDelReq: Received payloadSeqNum: %v", payloadSeqNum)

	if c.registry.IsValidSequenceNumber(payloadSeqNum) {
		var forwardRespToXapp bool = true
		_, err = c.trackDeleteTransaction(params, payloadSeqNum, forwardRespToXapp)
		if err != nil {
			xapp.Logger.Error("SubDelReq: Failed to create transaction record. Dropping this msg. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
			return
		}
		c.registry.setSubscriptionToUnConfirmed(payloadSeqNum)
	} else {
		xapp.Logger.Error("SubDelReq: Not valid sequence number. Dropping this msg. SubId: %v, Xid: %s", params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SubDelReq: Forwarding Request to E2T. Mtype: %v, SubId: %v, Xid: %s, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
	c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("SubDelReq: Failed to send request to E2T. Err %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}
	c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(payloadSeqNum), subDelReqTime, FirstTry, c.handleSubscriptionDeleteRequestTimer)
	return
}

func (c *Control) trackDeleteTransaction(params *xapp.RMRParams, payloadSeqNum uint16, forwardRespToXapp bool) (transaction *Transaction, err error) {
	srcAddr, srcPort, err := c.rtmgrClient.SplitSource(params.Src)
	if err != nil {
		xapp.Logger.Error("Failed to split source address. Err: %s, SubId: %v, Xid: %s", err, payloadSeqNum, params.Xid)
	}
	var respReceived bool = false
	transaction, err = c.tracker.TrackTransaction(payloadSeqNum, DELETE, *srcAddr, *srcPort, params, respReceived, forwardRespToXapp)
	return
}

func (c *Control) handleSubscriptionDeleteResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Info("SubDelResp received from Src: %s, Mtype: %v, SubId: %v, Meid: %v", params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteResponseSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		return
	}
	xapp.Logger.Info("SubDelResp: Received payloadSeqNum: %v", payloadSeqNum)

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(payloadSeqNum))

	transaction, responseReceived, err := c.tracker.CheckResponseReceived(payloadSeqNum, DELETE)
	if err != nil {
		xapp.Logger.Info("SubDelResp: Dropping this msg. Err: %v SubId: %v", err, payloadSeqNum)
		return
	}

	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}
	xapp.Logger.Info("SubDelResp: SubId: %v, from address: %v:%v. Forwarding response to xApp", payloadSeqNum, transaction.Xappkey.Addr, transaction.Xappkey.Port)

	if transaction.ForwardRespToXapp == true {
		params.SubId = int(payloadSeqNum)
		params.Xid = transaction.OrigParams.Xid
		xapp.Logger.Info("Forwarding SubDelResp to xApp: Mtype: %v, SubId: %v, Xid: %v, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
		err = c.rmrReplyToSender(params)
		if err != nil {
			xapp.Logger.Error("SubDelResp: Failed to send response to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		}

		time.Sleep(3 * time.Second)
	}

	xapp.Logger.Info("SubDelResp: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := SubRouteInfo{DELETE, transaction.Xappkey.Addr, transaction.Xappkey.Port, payloadSeqNum}
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SubDelResp: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		_, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
		if err != nil {
			xapp.Logger.Error("SubDelResp: Failed to delete transaction record. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
			return
		}
	} else {
		xapp.Logger.Error("SubDelResp: Failed to release sequency number. SubId: %v, Xid: %s", params.SubId, params.Xid)
		return
	}
	return
}

func (c *Control) handleSubscriptionDeleteFailure(params *xapp.RMRParams) {
	xapp.Logger.Info("SubDelFail received from Src: %s, Mtype: %v, SubId: %v, Meid: %v", params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteFailureSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelFail: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		return
	}
	xapp.Logger.Info("SubDelFail: Received payloadSeqNum: %v", payloadSeqNum)

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(payloadSeqNum))

	transaction, responseReceived, err := c.tracker.CheckResponseReceived(payloadSeqNum, DELETE)
	if err != nil {
		xapp.Logger.Info("SubDelFail: Dropping this msg. Err: %v SubId: %v", err, payloadSeqNum)
		return
	}

	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}
	xapp.Logger.Info("SubDelFail: SubId: %v, from address: %v:%v. Forwarding response to xApp", payloadSeqNum, transaction.Xappkey.Addr, transaction.Xappkey.Port)

	if transaction.ForwardRespToXapp == true {
		var subDelRespPayload []byte
		subDelRespPayload, err = c.e2ap.PackSubscriptionDeleteResponse(transaction.OrigParams.Payload, payloadSeqNum)
		if err != nil {
			xapp.Logger.Error("SubDelFail:Packing SubDelResp failed. Err: %v", err)
			return
		}

		params.Mtype = 12021 // RIC SUBSCRIPTION DELETE RESPONSE
		params.SubId = int(payloadSeqNum)
		params.Xid = transaction.OrigParams.Xid
		params.Meid = transaction.OrigParams.Meid
		params.Src = transaction.OrigParams.Src
		params.PayloadLen = len(subDelRespPayload)
		params.Payload = subDelRespPayload
		params.Mbuf = nil
		xapp.Logger.Info("SubDelFail: Forwarding SubDelFail to xApp: Mtype: %v, SubId: %v, Xid: %v, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
		err = c.rmrReplyToSender(params)
		if err != nil {
			xapp.Logger.Error("SubDelFail: Failed to send SubDelFail to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		}

		time.Sleep(3 * time.Second)
	}

	xapp.Logger.Info("SubDelFail: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := SubRouteInfo{DELETE, transaction.Xappkey.Addr, transaction.Xappkey.Port, payloadSeqNum}
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubDelFail: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SubDelFail: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		_, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
		if err != nil {
			xapp.Logger.Error("SubDelFail: Failed to delete transaction record. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
			return
		}
	} else {
		xapp.Logger.Error("SubDelFail: Failed to release sequency number. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}
	return
}

func (c *Control) handleSubscriptionDeleteRequestTimer(strId string, nbrId int, tryCount uint64) {
	subId := uint16(nbrId)
	xapp.Logger.Info("handleSubDelTimer: SubDelReq timer expired. subId: %v, tryCount: %v", subId, tryCount)

	transaction, responseReceived, err := c.tracker.CheckResponseReceived(subId, DELETE)
	if err != nil {
		xapp.Logger.Info("handleSubTimer: Dropping this timer action. Err: %v SubId: %v", err, subId)
		return
	}

	if responseReceived == true {
		// Subscription Delete Response or Failure already received
		return
	}

	if tryCount < maxSubDelReqTryCount {
		xapp.Logger.Info("handleSubDelTimer: Resending SubDelReq to E2T: Mtype: %v, SubId: %v, Xid %s, Meid %v", transaction.OrigParams.Mtype, transaction.OrigParams.SubId, transaction.OrigParams.Xid, transaction.OrigParams.Meid)
		// Set possible to handle new response for the subId
		err = c.tracker.RetryTransaction(subId, DELETE)
		if err != nil {
			xapp.Logger.Error("handleSubDelTimer: Failed to retry transaction record. Dropping  timer action. Err %v, SubId: %v", err, transaction.OrigParams.SubId)
			return
		}

		err = c.rmrSend(transaction.OrigParams)
		if err != nil {
			xapp.Logger.Error("handleSubDelTimer: Failed to send request to E2T %v, SubId: %v, Xid: %s", err, transaction.OrigParams.SubId, transaction.OrigParams.Xid)
		}

		tryCount++
		c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subId), subReqTime, tryCount, c.handleSubscriptionDeleteRequestTimer)
		return
	}

	var params xapp.RMRParams
	if transaction.ForwardRespToXapp == true {
		var subDelRespPayload []byte
		subDelRespPayload, err = c.e2ap.PackSubscriptionDeleteResponse(transaction.OrigParams.Payload, subId)
		if err != nil {
			xapp.Logger.Error("handleSubDelTimer: Unable to pack payload. Dropping this timer action. Err: %v, SubId: %v, Xid: %s, Payload %x", err, subId, transaction.OrigParams.Xid, transaction.OrigParams.Payload)
			return
		}

		params.Mtype = 12021 // RIC SUBSCRIPTION DELETE RESPONSE
		params.SubId = int(subId)
		params.Meid = transaction.OrigParams.Meid
		params.Xid = transaction.OrigParams.Xid
		params.Src = transaction.OrigParams.Src
		params.PayloadLen = len(subDelRespPayload)
		params.Payload = subDelRespPayload
		params.Mbuf = nil

		xapp.Logger.Info("handleSubDelTimer: Sending SubDelResp to xApp: Mtype: %v, SubId: %v, Xid: %s, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
		err = c.rmrReplyToSender(&params)
		if err != nil {
			xapp.Logger.Error("handleSubDelTimer: Failed to send response to xApp: Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		}

		time.Sleep(3 * time.Second)
	}

	xapp.Logger.Info("handleSubDelTimer: Starting routing manager update. SubId: %v, Xid: %s", subId, params.Xid)
	subRouteAction := SubRouteInfo{DELETE, transaction.Xappkey.Addr, transaction.Xappkey.Port, subId}
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("handleSubDelTimer: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, subId, params.Xid)
		return
	}

	xapp.Logger.Info("handleSubDelTimer: Deleting transaction record. SubId: %v, Xid: %s", subId, params.Xid)
	if c.registry.releaseSequenceNumber(subId) {
		_, err = c.tracker.completeTransaction(subId, DELETE)
		if err != nil {
			xapp.Logger.Error("handleSubDelTimer: Failed to delete transaction record. Err: %v, SubId: %v, Xid: %s", err, subId, params.Xid)
			return
		}
	} else {
		xapp.Logger.Error("handleSubDelTimer: Failed to release sequency number. SubId: %v, Xid: %s", subId, params.Xid)
	}
	return
}
