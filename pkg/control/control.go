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
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

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

	rtmgrClientPtr := &rtmgrClient

	//TODO: to make this better. Now it is just a hack.
	registry.rtmgrClient = rtmgrClientPtr

	return &Control{e2ap: new(E2ap),
		registry:    registry,
		rtmgrClient: rtmgrClientPtr,
		tracker:     tracker,
		timerMap:    timerMap,
		msgCounter:  0,
	}
}

func (c *Control) Run() {
	xapp.Run(c)
}

func (c *Control) rmrSendRaw(desc string, params *RMRParams) (err error) {

	xapp.Logger.Info("%s: %s", desc, params.String())
	status := false
	i := 1
	for ; i <= 10 && status == false; i++ {
		c.rmrSendMutex.Lock()
		status = xapp.Rmr.Send(params.RMRParams, false)
		c.rmrSendMutex.Unlock()
		if status == false {
			xapp.Logger.Info("rmr.Send() failed. Retry count %d, %s", i, params.String())
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = fmt.Errorf("rmr.Send() failed. Retry count %d, %s", i, params.String())
		xapp.Logger.Error("%s: %s", desc, err.Error())
		xapp.Rmr.Free(params.Mbuf)
	}
	return
}

func (c *Control) rmrSend(desc string, subs *Subscription, trans *Transaction, payload []byte, payloadLen int) (err error) {
	params := &RMRParams{&xapp.RMRParams{}}
	params.Mtype = trans.GetMtype()
	params.SubId = int(subs.GetSubId())
	params.Xid = ""
	params.Meid = subs.GetMeid()
	params.Src = ""
	params.PayloadLen = payloadLen
	params.Payload = payload
	params.Mbuf = nil

	return c.rmrSendRaw(desc, params)
}

func (c *Control) rmrReplyToSender(desc string, subs *Subscription, trans *Transaction, mType int, payload []byte, payloadLen int) (err error) {
	params := &RMRParams{&xapp.RMRParams{}}
	params.Mtype = mType
	params.SubId = int(subs.GetSubId())
	params.Xid = trans.GetXid()
	params.Meid = trans.GetMeid()
	params.Src = ""
	params.PayloadLen = payloadLen
	params.Payload = payload
	params.Mbuf = nil

	return c.rmrSendRaw(desc, params)
}

func (c *Control) Consume(params *xapp.RMRParams) (err error) {
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil
	msg := &RMRParams{params}
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

func (c *Control) handleSubscriptionRequest(params *RMRParams) {
	xapp.Logger.Info("SubReq from xapp: %s", params.String())

	//
	//
	//
	trans, err := c.tracker.TrackTransaction(NewRmrEndpoint(params.Src),
		params.Mtype,
		params.Xid,
		params.Meid,
		false,
		true)

	if err != nil {
		xapp.Logger.Error("SubReq: %s, Dropping this msg. %s", err.Error(), params.String())
		return
	}

	//
	//
	//
	trans.SubReqMsg, err = c.e2ap.UnpackSubscriptionRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubReq: %s Dropping this msg. %s", err.Error(), trans)
		trans.Release()
		return
	}

	//
	//
	//
	subs, err := c.registry.ReserveSubscription(&trans.RmrEndpoint, trans.Meid)
	if err != nil {
		xapp.Logger.Error("SubReq: %s, Dropping this msg. %s", err.Error(), trans)
		trans.Release()
		return
	}

	err = subs.SetTransaction(trans)
	if err != nil {
		xapp.Logger.Error("SubReq: %s, Dropping this msg. %s", err.Error(), trans)
		c.registry.DelSubscription(subs.Seq)
		trans.Release()
		return
	}

	trans.SubReqMsg.RequestId.Seq = uint32(subs.GetSubId())

	//
	// TODO: subscription create is in fact owned by subscription and not transaction.
	//       Transaction is toward xapp while Subscription is toward ran.
	//       In merge several xapps may wake transactions, while only one subscription
	//       toward ran occurs -> subscription owns subscription creation toward ran
	//
	//       This is intermediate solution while improving message handling
	//
	packedData, err := c.e2ap.PackSubscriptionRequest(trans.SubReqMsg)
	if err != nil {
		xapp.Logger.Error("SubReq: %s for trans %s", err.Error(), trans)
		c.registry.DelSubscription(subs.Seq)
		trans.Release()
		return
	}

	//Optimize and store packed message to be sent (for retransmission). Again owned by subscription?
	trans.Payload = packedData.Buf
	trans.PayloadLen = len(packedData.Buf)

	c.rmrSend("SubReq to E2T", subs, trans, packedData.Buf, len(packedData.Buf))

	c.timerMap.StartTimer("RIC_SUB_REQ", int(subs.GetSubId()), subReqTime, FirstTry, c.handleSubscriptionRequestTimer)
	xapp.Logger.Debug("SubReq: Debugging trans table = %v", c.tracker.transactionXappTable)
	return
}

func (c *Control) handleSubscriptionResponse(params *RMRParams) {
	xapp.Logger.Info("SubResp from E2T: %s", params.String())

	//
	//
	//
	SubRespMsg, err := c.e2ap.UnpackSubscriptionResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelReq: %s Dropping this msg. %s", err.Error(), params.String())
		return
	}

	//
	//
	//
	subs := c.registry.GetSubscription(uint16(SubRespMsg.RequestId.Seq))
	if subs == nil && params.SubId > 0 {
		subs = c.registry.GetSubscription(uint16(params.SubId))
	}

	if subs == nil {
		xapp.Logger.Error("SubResp: Not valid subscription found payloadSeqNum: %d, SubId: %d. Dropping this msg. %s", SubRespMsg.RequestId.Seq, params.SubId, params.String())
		return
	}
	xapp.Logger.Info("SubResp: subscription found payloadSeqNum: %d, SubId: %d", SubRespMsg.RequestId.Seq, subs.GetSubId())

	//
	//
	//
	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubResp: Unknown trans. Dropping this msg. SubId: %d", subs.GetSubId())
		return
	}

	trans.SubRespMsg = SubRespMsg

	//
	//
	//
	c.timerMap.StopTimer("RIC_SUB_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription timer already received
		return
	}

	packedData, err := c.e2ap.PackSubscriptionResponse(trans.SubRespMsg)
	if err != nil {
		xapp.Logger.Error("SubResp: %s for trans %s", err.Error(), trans)
		trans.Release()
		return
	}

	//Optimize and store packed message to be sent.
	trans.Payload = packedData.Buf
	trans.PayloadLen = len(packedData.Buf)

	subs.Confirmed()
	trans.Release()
	c.rmrReplyToSender("SubResp to xapp", subs, trans, 12011, trans.Payload, trans.PayloadLen)
	return
}

func (c *Control) handleSubscriptionFailure(params *RMRParams) {
	xapp.Logger.Info("SubFail from E2T: %s", params.String())

	payloadSeqNum, err := c.e2ap.GetSubscriptionFailureSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubFail: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		return
	}
	xapp.Logger.Info("SubFail: Received payloadSeqNum: %v", payloadSeqNum)

	subs := c.registry.GetSubscription(payloadSeqNum)
	if subs == nil {
		xapp.Logger.Error("SubFail: Unknown payloadSeqNum. Dropping this msg. PayloadSeqNum: %v, SubId: %v", payloadSeqNum, params.SubId)
		return
	}

	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubFail: Unknown trans. Dropping this msg. PayloadSeqNum: %v, SubId: %v", payloadSeqNum, params.SubId)
		return
	}

	c.timerMap.StopTimer("RIC_SUB_REQ", int(payloadSeqNum))

	responseReceived := trans.CheckResponseReceived()
	if err != nil {
		xapp.Logger.Info("SubFail: Dropping this msg. Err: %v SubId: %v", err, payloadSeqNum)
		return
	}

	if responseReceived == true {
		// Subscription timer already received
		return
	}
	xapp.Logger.Info("SubFail: SubId: %v, from address: %s. Forwarding response to xApp", payloadSeqNum, trans.RmrEndpoint)

	c.rmrReplyToSender("SubFail to xapp", subs, trans, params.Mtype, params.Payload, params.PayloadLen)

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("SubFail: Deleting trans record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	trans.Release()
	if !c.registry.DelSubscription(payloadSeqNum) {
		xapp.Logger.Error("SubFail: Failed to release sequency number. SubId: %v, Xid: %s", params.SubId, params.Xid)
	}
	return
}

func (c *Control) handleSubscriptionRequestTimer(strId string, nbrId int, tryCount uint64) {
	xapp.Logger.Info("SubReq timeout: subId: %v,  tryCount: %v", nbrId, tryCount)

	subs := c.registry.GetSubscription(uint16(nbrId))
	if subs == nil {
		xapp.Logger.Error("SubReq timeout: Unknown payloadSeqNum. Dropping this msg. SubId: %v", nbrId)
		return
	}

	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubReq timeout: Unknown trans. Dropping this msg. SubId: %v", subs.GetSubId())
		return
	}

	responseReceived := trans.CheckResponseReceived()

	if responseReceived == true {
		// Subscription Response or Failure already received
		return
	}

	if tryCount < maxSubReqTryCount {
		xapp.Logger.Info("SubReq timeout: Resending SubReq to E2T: Mtype: %v, SubId: %v, Xid %s, Meid %v", trans.GetMtype(), subs.GetSubId(), trans.GetXid(), trans.GetMeid())

		trans.RetryTransaction()

		c.rmrSend("SubReq(SubReq timer) to E2T", subs, trans, trans.Payload, trans.PayloadLen)

		tryCount++
		c.timerMap.StartTimer("RIC_SUB_REQ", int(subs.GetSubId()), subReqTime, tryCount, c.handleSubscriptionRequestTimer)
		return
	}

	// Delete CREATE transaction
	trans.Release()

	// Create DELETE transaction (internal and no messages toward xapp)
	deltrans, err := c.tracker.TrackTransaction(&trans.RmrEndpoint,
		12020, // RIC SUBSCRIPTION DELETE
		trans.GetXid(),
		trans.GetMeid(),
		false,
		false)

	if err != nil {
		xapp.Logger.Error("SubReq timeout: %s, Dropping this msg.", err.Error())
		//TODO improve error handling. Important at least in merge
		c.registry.DelSubscription(subs.GetSubId())
		return
	}

	deltrans.SubDelReqMsg = &e2ap.E2APSubscriptionDeleteRequest{}
	deltrans.SubDelReqMsg.RequestId.Id = trans.SubReqMsg.RequestId.Id
	deltrans.SubDelReqMsg.RequestId.Seq = uint32(subs.GetSubId())
	deltrans.SubDelReqMsg.FunctionId = trans.SubReqMsg.FunctionId
	packedData, err := c.e2ap.PackSubscriptionDeleteRequest(deltrans.SubDelReqMsg)
	if err != nil {
		xapp.Logger.Error("SubReq timeout: Packing SubDelReq failed. Err: %v", err)
		//TODO improve error handling. Important at least in merge
		deltrans.Release()
		c.registry.DelSubscription(subs.GetSubId())
		return
	}
	deltrans.PayloadLen = len(packedData.Buf)
	deltrans.Payload = packedData.Buf

	err = subs.SetTransaction(deltrans)
	if err != nil {
		xapp.Logger.Error("SubReq timeout: %s, Dropping this msg.", err.Error())
		//TODO improve error handling. Important at least in merge
		deltrans.Release()
		return
	}

	c.rmrSend("SubDelReq(SubReq timer) to E2T", subs, deltrans, deltrans.Payload, deltrans.PayloadLen)

	c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subDelReqTime, FirstTry, c.handleSubscriptionDeleteRequestTimer)
	return
}

func (c *Control) handleSubscriptionDeleteRequest(params *RMRParams) {
	xapp.Logger.Info("SubDelReq from xapp: %s", params.String())

	//
	//
	//
	trans, err := c.tracker.TrackTransaction(NewRmrEndpoint(params.Src),
		params.Mtype,
		params.Xid,
		params.Meid,
		false,
		true)

	if err != nil {
		xapp.Logger.Error("SubDelReq: %s, Dropping this msg. %s", err.Error(), params.String())
		return
	}

	//
	//
	//
	trans.SubDelReqMsg, err = c.e2ap.UnpackSubscriptionDeleteRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelReq: %s Dropping this msg. %s", err.Error(), trans)
		trans.Release()
		return
	}

	//
	//
	//
	subs := c.registry.GetSubscription(uint16(trans.SubDelReqMsg.RequestId.Seq))
	if subs == nil && params.SubId > 0 {
		subs = c.registry.GetSubscription(uint16(params.SubId))
	}

	if subs == nil {
		xapp.Logger.Error("SubDelReq: Not valid subscription found payloadSeqNum: %d, SubId: %d. Dropping this msg. %s", trans.SubDelReqMsg.RequestId.Seq, params.SubId, trans)
		trans.Release()
		return
	}
	xapp.Logger.Info("SubDelReq: subscription found payloadSeqNum: %d, SubId: %d. %s", trans.SubDelReqMsg.RequestId.Seq, params.SubId, trans)

	err = subs.SetTransaction(trans)
	if err != nil {
		xapp.Logger.Error("SubDelReq: %s, Dropping this msg. %s", err.Error(), trans)
		trans.Release()
		return
	}

	//
	// TODO: subscription delete is in fact owned by subscription and not transaction.
	//       Transaction is toward xapp while Subscription is toward ran.
	//       In merge several xapps may wake transactions, while only one subscription
	//       toward ran occurs -> subscription owns subscription creation toward ran
	//
	//       This is intermediate solution while improving message handling
	//
	packedData, err := c.e2ap.PackSubscriptionDeleteRequest(trans.SubDelReqMsg)
	if err != nil {
		xapp.Logger.Error("SubDelReq: %s for trans %s", err.Error(), trans)
		trans.Release()
		return
	}

	//Optimize and store packed message to be sent (for retransmission). Again owned by subscription?
	trans.Payload = packedData.Buf
	trans.PayloadLen = len(packedData.Buf)

	subs.UnConfirmed()

	c.rmrSend("SubDelReq to E2T", subs, trans, trans.Payload, trans.PayloadLen)

	c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subDelReqTime, FirstTry, c.handleSubscriptionDeleteRequestTimer)
	return
}

func (c *Control) handleSubscriptionDeleteResponse(params *RMRParams) (err error) {
	xapp.Logger.Info("SubDelResp from E2T:%s", params.String())

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteResponseSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s, Payload %X", err, params.SubId, params.Xid, params.Payload)
		return
	}
	xapp.Logger.Info("SubDelResp: Received payloadSeqNum: %v", payloadSeqNum)

	subs := c.registry.GetSubscription(payloadSeqNum)
	if subs == nil {
		xapp.Logger.Error("SubDelResp: Unknown payloadSeqNum. Dropping this msg. PayloadSeqNum: %v, SubId: %v", payloadSeqNum, params.SubId)
		return
	}

	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubDelResp: Unknown trans. Dropping this msg. PayloadSeqNum: %v, SubId: %v", subs.GetSubId(), params.SubId)
		return
	}

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}

	trans.Release()

	if trans.ForwardRespToXapp == true {
		c.rmrReplyToSender("SubDelResp to xapp", subs, trans, params.Mtype, params.Payload, params.PayloadLen)
		time.Sleep(3 * time.Second)
	}

	xapp.Logger.Info("SubDelResp: Deleting trans record. SubId: %v, Xid: %s", subs.GetSubId(), trans.GetXid())
	if !c.registry.DelSubscription(subs.GetSubId()) {
		xapp.Logger.Error("SubDelResp: Failed to release sequency number. SubId: %v, Xid: %s", subs.GetSubId(), trans.GetXid())
		return
	}
	return
}

func (c *Control) handleSubscriptionDeleteFailure(params *RMRParams) {
	xapp.Logger.Info("SubDelFail from E2T:%s", params.String())

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteFailureSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelFail: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, %s", err, params.String())
		return
	}
	xapp.Logger.Info("SubDelFail: Received payloadSeqNum: %v", payloadSeqNum)

	subs := c.registry.GetSubscription(payloadSeqNum)
	if subs == nil {
		xapp.Logger.Error("SubDelFail: Unknown payloadSeqNum. Dropping this msg. PayloadSeqNum: %v, SubId: %v", payloadSeqNum, params.SubId)
		return
	}

	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubDelFail: Unknown trans. Dropping this msg. PayloadSeqNum: %v, SubId: %v", subs.GetSubId(), params.SubId)
		return
	}

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}
	if trans.ForwardRespToXapp == true {
		var subDelRespPayload []byte
		subDelRespPayload, err = c.e2ap.PackSubscriptionDeleteResponseFromSubDelReq(trans.Payload, subs.GetSubId())
		if err != nil {
			xapp.Logger.Error("SubDelFail:Packing SubDelResp failed. Err: %v", err)
			return
		}

		// RIC SUBSCRIPTION DELETE RESPONSE
		c.rmrReplyToSender("SubDelFail to xapp", subs, trans, 12021, subDelRespPayload, len(subDelRespPayload))
		time.Sleep(3 * time.Second)
	}

	xapp.Logger.Info("SubDelFail: Deleting trans record. SubId: %v, Xid: %s", subs.GetSubId(), trans.GetXid())
	trans.Release()
	if !c.registry.DelSubscription(subs.GetSubId()) {
		xapp.Logger.Error("SubDelFail: Failed to release sequency number. Err: %v, SubId: %v, Xid: %s", err, subs.GetSubId(), trans.GetXid())
		return
	}
	return
}

func (c *Control) handleSubscriptionDeleteRequestTimer(strId string, nbrId int, tryCount uint64) {
	xapp.Logger.Info("SubDelReq timeout: subId: %v, tryCount: %v", nbrId, tryCount)

	subs := c.registry.GetSubscription(uint16(nbrId))
	if subs == nil {
		xapp.Logger.Error("SubDelReq timeout: Unknown payloadSeqNum. Dropping this msg. SubId: %v", nbrId)
		return
	}

	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubDelReq timeout: Unknown trans. Dropping this msg. SubId: %v", subs.GetSubId())
		return
	}

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete Response or Failure already received
		return
	}

	if tryCount < maxSubDelReqTryCount {
		xapp.Logger.Info("SubDelReq timeout: Resending SubDelReq to E2T: Mtype: %v, SubId: %v, Xid %s, Meid %v", trans.GetMtype(), subs.GetSubId(), trans.GetXid(), trans.GetMeid())
		// Set possible to handle new response for the subId

		trans.RetryTransaction()

		c.rmrSend("SubDelReq(SubDelReq timer) to E2T", subs, trans, trans.Payload, trans.PayloadLen)

		tryCount++
		c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subReqTime, tryCount, c.handleSubscriptionDeleteRequestTimer)
		return
	}

	if trans.ForwardRespToXapp == true {
		var subDelRespPayload []byte
		subDelRespPayload, err := c.e2ap.PackSubscriptionDeleteResponseFromSubDelReq(trans.Payload, subs.GetSubId())
		if err != nil {
			xapp.Logger.Error("SubDelReq timeout: Unable to pack payload. Dropping this this msg. Err: %v, SubId: %v, Xid: %s, Payload %x", err, subs.GetSubId(), trans.GetXid(), trans.Payload)
			return
		}

		// RIC SUBSCRIPTION DELETE RESPONSE
		c.rmrReplyToSender("SubDelResp(SubDelReq timer) to xapp", subs, trans, 12021, subDelRespPayload, len(subDelRespPayload))

		time.Sleep(3 * time.Second)

	}

	xapp.Logger.Info("SubDelReq timeout: Deleting trans record. SubId: %v, Xid: %s", subs.GetSubId(), trans.GetXid())
	trans.Release()
	if !c.registry.DelSubscription(subs.GetSubId()) {
		xapp.Logger.Error("SubDelReq timeout: Failed to release sequency number. SubId: %v, Xid: %s", subs.GetSubId(), trans.GetXid())
	}
	return
}
