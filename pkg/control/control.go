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

	transport := httptransport.New(viper.GetString("rtmgr.HostAddr")+":"+viper.GetString("rtmgr.port"), viper.GetString("rtmgr.baseUrl"), []string{"http"})
	client := rtmgrclient.New(transport, strfmt.Default)
	handle := rtmgrhandle.NewProvideXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	deleteHandle := rtmgrhandle.NewDeleteXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	rtmgrClient := RtmgrClient{client, handle, deleteHandle}

	registry := new(Registry)
	registry.Initialize(seedSN)
	registry.rtmgrClient = &rtmgrClient

	tracker := new(Tracker)
	tracker.Init()

	timerMap := new(TimerMap)
	timerMap.Init()

	return &Control{e2ap: new(E2ap),
		registry:   registry,
		tracker:    tracker,
		timerMap:   timerMap,
		msgCounter: 0,
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
		subs.Release()
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
		subs.Release()
		trans.Release()
		return
	}

	//Optimize and store packed message to be sent (for retransmission). Again owned by subscription?
	trans.Payload = packedData.Buf
	trans.PayloadLen = len(packedData.Buf)

	c.rmrSend("SubReq: SubReq to E2T", subs, trans, trans.Payload, trans.PayloadLen)

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
	c.rmrReplyToSender("SubResp: SubResp to xapp", subs, trans, 12011, trans.Payload, trans.PayloadLen)
	return
}

func (c *Control) handleSubscriptionFailure(params *RMRParams) {
	xapp.Logger.Info("SubFail from E2T: %s", params.String())

	//
	//
	//
	SubFailMsg, err := c.e2ap.UnpackSubscriptionFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubFail: %s Dropping this msg. %s", err.Error(), params.String())
		return
	}

	//
	//
	//
	subs := c.registry.GetSubscription(uint16(SubFailMsg.RequestId.Seq))
	if subs == nil && params.SubId > 0 {
		subs = c.registry.GetSubscription(uint16(params.SubId))
	}

	if subs == nil {
		xapp.Logger.Error("SubFail: Not valid subscription found payloadSeqNum: %d, SubId: %d. Dropping this msg. %s", SubFailMsg.RequestId.Seq, params.SubId, params.String())
		return
	}
	xapp.Logger.Info("SubFail: subscription found payloadSeqNum: %d, SubId: %d", SubFailMsg.RequestId.Seq, subs.GetSubId())

	//
	//
	//
	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubFail: Unknown trans. Dropping this msg. SubId: %d", subs.GetSubId())
		return
	}
	trans.SubFailMsg = SubFailMsg

	//
	//
	//
	c.timerMap.StopTimer("RIC_SUB_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if err != nil {
		return
	}

	if responseReceived == true {
		// Subscription timer already received
		return
	}

	packedData, err := c.e2ap.PackSubscriptionFailure(trans.SubFailMsg)
	if err == nil {
		//Optimize and store packed message to be sent.
		trans.Payload = packedData.Buf
		trans.PayloadLen = len(packedData.Buf)
		c.rmrReplyToSender("SubFail: SubFail to xapp", subs, trans, 12012, trans.Payload, trans.PayloadLen)
		time.Sleep(3 * time.Second)
	} else {
		//TODO error handling improvement
		xapp.Logger.Error("SubFail: %s for trans %s (continuing cleaning)", err.Error(), trans)
	}

	trans.Release()
	subs.Release()
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

		c.rmrSend("SubReq timeout: SubReq to E2T", subs, trans, trans.Payload, trans.PayloadLen)

		tryCount++
		c.timerMap.StartTimer("RIC_SUB_REQ", int(subs.GetSubId()), subReqTime, tryCount, c.handleSubscriptionRequestTimer)
		return
	}

	// Release CREATE transaction
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
		subs.Release()
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
		subs.Release()
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

	c.rmrSend("SubReq timer: SubDelReq to E2T", subs, deltrans, deltrans.Payload, deltrans.PayloadLen)
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

	c.rmrSend("SubDelReq: SubDelReq to E2T", subs, trans, trans.Payload, trans.PayloadLen)

	c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subDelReqTime, FirstTry, c.handleSubscriptionDeleteRequestTimer)
	return
}

func (c *Control) handleSubscriptionDeleteResponse(params *RMRParams) (err error) {
	xapp.Logger.Info("SubDelResp from E2T:%s", params.String())

	//
	//
	//
	SubDelRespMsg, err := c.e2ap.UnpackSubscriptionDeleteResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelResp: %s Dropping this msg. %s", err.Error(), params.String())
		return
	}

	//
	//
	//
	subs := c.registry.GetSubscription(uint16(SubDelRespMsg.RequestId.Seq))
	if subs == nil && params.SubId > 0 {
		subs = c.registry.GetSubscription(uint16(params.SubId))
	}

	if subs == nil {
		xapp.Logger.Error("SubDelResp: Not valid subscription found payloadSeqNum: %d, SubId: %d. Dropping this msg. %s", SubDelRespMsg.RequestId.Seq, params.SubId, params.String())
		return
	}
	xapp.Logger.Info("SubDelResp: subscription found payloadSeqNum: %d, SubId: %d", SubDelRespMsg.RequestId.Seq, subs.GetSubId())

	//
	//
	//
	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubDelResp: Unknown trans. Dropping this msg. SubId: %d", subs.GetSubId())
		return
	}

	trans.SubDelRespMsg = SubDelRespMsg

	//
	//
	//
	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}

	c.sendSubscriptionDeleteResponse("SubDelResp", trans, subs)
	return
}

func (c *Control) handleSubscriptionDeleteFailure(params *RMRParams) {
	xapp.Logger.Info("SubDelFail from E2T:%s", params.String())

	//
	//
	//
	SubDelFailMsg, err := c.e2ap.UnpackSubscriptionDeleteFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelFail: %s Dropping this msg. %s", err.Error(), params.String())
		return
	}

	//
	//
	//
	subs := c.registry.GetSubscription(uint16(SubDelFailMsg.RequestId.Seq))
	if subs == nil && params.SubId > 0 {
		subs = c.registry.GetSubscription(uint16(params.SubId))
	}

	if subs == nil {
		xapp.Logger.Error("SubDelFail: Not valid subscription found payloadSeqNum: %d, SubId: %d. Dropping this msg. %s", SubDelFailMsg.RequestId.Seq, params.SubId, params.String())
		return
	}
	xapp.Logger.Info("SubDelFail: subscription found payloadSeqNum: %d, SubId: %d", SubDelFailMsg.RequestId.Seq, subs.GetSubId())

	//
	//
	//
	trans := subs.GetTransaction()
	if trans == nil {
		xapp.Logger.Error("SubDelFail: Unknown trans. Dropping this msg. SubId: %d", subs.GetSubId())
		return
	}
	trans.SubDelFailMsg = SubDelFailMsg

	//
	//
	//
	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}

	c.sendSubscriptionDeleteResponse("SubDelFail", trans, subs)
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
		// Set possible to handle new response for the subId
		trans.RetryTransaction()
		c.rmrSend("SubDelReq timeout: SubDelReq to E2T", subs, trans, trans.Payload, trans.PayloadLen)
		tryCount++
		c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subReqTime, tryCount, c.handleSubscriptionDeleteRequestTimer)
		return
	}

	c.sendSubscriptionDeleteResponse("SubDelReq(timer)", trans, subs)
	return
}

func (c *Control) sendSubscriptionDeleteResponse(desc string, trans *Transaction, subs *Subscription) {

	if trans.ForwardRespToXapp == true {
		//Always generate SubDelResp
		trans.SubDelRespMsg = &e2ap.E2APSubscriptionDeleteResponse{}
		trans.SubDelRespMsg.RequestId.Id = trans.SubDelReqMsg.RequestId.Id
		trans.SubDelRespMsg.RequestId.Seq = uint32(subs.GetSubId())
		trans.SubDelRespMsg.FunctionId = trans.SubDelReqMsg.FunctionId

		packedData, err := c.e2ap.PackSubscriptionDeleteResponse(trans.SubDelRespMsg)
		if err == nil {
			trans.Payload = packedData.Buf
			trans.PayloadLen = len(packedData.Buf)
			c.rmrReplyToSender(desc+": SubDelResp to xapp", subs, trans, 12021, trans.Payload, trans.PayloadLen)
			time.Sleep(3 * time.Second)
		} else {
			//TODO error handling improvement
			xapp.Logger.Error("%s: %s for trans %s (continuing cleaning)", desc, err.Error(), trans)
		}
	}

	trans.Release()
	subs.Release()
}
