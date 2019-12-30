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
	"sync"
	"time"
)

var subReqTime time.Duration = 5 * time.Second
var SubDelReqTime time.Duration = 5 * time.Second

type Control struct {
	e2ap         *E2ap
	registry     *Registry
	rtmgrClient  *RtmgrClient
	tracker      *Tracker
	timerMap     *TimerMap
	rmrSendMutex sync.Mutex
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
	xapp.Logger.Info("SUBMGR /ric-plt-submgr:r3-test-v4")
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
	newSubId, isIdValid := c.registry.ReserveSequenceNumber()
	if isIdValid != true {
		xapp.Logger.Error("SubReq: Failed to reserve sequence number. Dropping this msg. SubId: %v, Xid: %s", params.SubId, params.Xid)
		return
	}

	err := c.e2ap.SetSubscriptionRequestSequenceNumber(params.Payload, newSubId)
	if err != nil {
		xapp.Logger.Error("SubReq: Unable to set Sequence Number in Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		c.registry.releaseSequenceNumber(newSubId)
		return
	}

	srcAddr, srcPort, err := c.rtmgrClient.SplitSource(params.Src)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to update routing-manager. Dropping this msg. Err: %s, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		c.registry.releaseSequenceNumber(newSubId)
		return
	}

	/* Create transatcion records for every subscription request */
	transaction, err := c.tracker.TrackTransaction(newSubId, CREATE, *srcAddr, *srcPort, params)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to create transaction record. Dropping this msg. Err: %v SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		c.registry.releaseSequenceNumber(newSubId)
		return
	}

	/* Update routing manager about the new subscription*/
	subRouteAction := transaction.SubRouteInfo()
	xapp.Logger.Info("SubReq: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)

	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to update routing manager. Dropping this SubReq msg. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}

	// Setting new subscription ID in the RMR header
	params.SubId = int(newSubId)
	xapp.Logger.Info("Forwarding SubReq to E2T: Mtype: %v, SubId: %v, Xid %s, Meid %v", params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("SubReq: Failed to send request to E2T %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	} else {
		c.timerMap.StartTimer("RIC_SUB_REQ", int(newSubId), subReqTime, c.handleSubscriptionRequestTimer)
	}
	xapp.Logger.Debug("SubReq: Debugging transaction table = %v", c.tracker.transactionTable)
	return
}

func (c *Control) handleSubscriptionResponse(params *xapp.RMRParams) {
	xapp.Logger.Info("SubResp received from Src: %s, Mtype: %v, SubId: %v, Meid: %v", params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionResponseSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubResp: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubResp: Received payloadSeqNum: %v", payloadSeqNum)

	if !c.registry.IsValidSequenceNumber(payloadSeqNum) {
		xapp.Logger.Error("SubResp: Unknown payloadSeqNum. Dropping this msg. PayloadSeqNum: %v, SubId: %v", payloadSeqNum, params.SubId)
		return
	}

	c.timerMap.StopTimer("RIC_SUB_REQ", int(payloadSeqNum))

	c.registry.setSubscriptionToConfirmed(payloadSeqNum)
	transaction, err := c.tracker.RetriveTransaction(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Error("SubResp: Failed to retrive transaction record. Dropping this msg. Err: %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubResp: SubId: %v, from address: %v:%v. Retrieved old subId", int(payloadSeqNum), transaction.Xappkey.Addr, transaction.Xappkey.Port)

	params.SubId = int(payloadSeqNum)
	params.Xid = transaction.OrigParams.Xid

	xapp.Logger.Info("SubResp: Forwarding Subscription Response to xApp Mtype: %v, SubId: %v, Meid: %v", params.Mtype, params.SubId, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("SubResp: Failed to send response to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	xapp.Logger.Info("SubResp: SubId: %v, from address: %v:%v. Deleting transaction record", int(payloadSeqNum), transaction.Xappkey.Addr, transaction.Xappkey.Port)
	transaction, err = c.tracker.completeTransaction(payloadSeqNum, CREATE)
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
		xapp.Logger.Error("SubFail: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubFail: Received payloadSeqNum: %v", payloadSeqNum)

	c.timerMap.StopTimer("RIC_SUB_REQ", int(payloadSeqNum))

	transaction, err := c.tracker.RetriveTransaction(payloadSeqNum, CREATE)
	if err != nil {
		xapp.Logger.Error("SubFail: Failed to retrive transaction record. Dropping this msg. Err: %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubFail: SubId: %v, from address: %v:%v. Forwarding response to xApp", int(payloadSeqNum), transaction.Xappkey.Addr, transaction.Xappkey.Port)

	params.SubId = int(payloadSeqNum)
	params.Xid = transaction.OrigParams.Xid

	xapp.Logger.Info("Forwarding SubFail to xApp: Mtype: %v, SubId: %v, Xid: %v, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("Failed to send response to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("SubFail: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := transaction.SubRouteInfo()
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubFail: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	xapp.Logger.Info("SubFail: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, CREATE)
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

func (c *Control) handleSubscriptionRequestTimer(strId string, nbrId int) {
	newSubId := uint16(nbrId)
	xapp.Logger.Info("SubReq timer expired. newSubId: %v", newSubId)
	//	var causeContent uint8 = 1  // just some random cause. To be checked later. Should be no respose or something
	//	var causeVal uint8 = 1  // just some random val. To be checked later. Should be no respose or something
	//	c.sendSubscriptionFailure(newSubId, causeContent, causeVal)
}

/*
func (c *Control) sendSubscriptionFailure(subId uint16, causeContent uint8, causeVal uint8) {

	transaction, err := c.tracker.completeTransaction(subId, CREATE)
	if err != nil {
		xapp.Logger.Error("SendSubFail: Failed to delete transaction record. Err:%v. SubId: %v", err, subId)
		return
	}
	xapp.Logger.Info("SendSubFail: SubId: %v, Xid %v, Meid: %v", subId, transaction.OrigParams.Xid, transaction.OrigParams.Meid)

	var params xapp.RMRParams
	params.Mtype = 12012 //xapp.RICMessageTypes["RIC_SUB_FAILURE"]
	params.SubId = int(subId)
	params.Meid = transaction.OrigParams.Meid
	params.Xid = transaction.OrigParams.Xid

//	newPayload, packErr := c.e2ap.PackSubscriptionFailure(transaction.OrigParams.Payload, subId, causeContent, causeVal)
//	if packErr != nil {
//		xapp.Logger.Error("SendSubFail: PackSubscriptionFailure() due to %v", packErr)
//		return
//	}

	newPayload := []byte("40CA4018000003EA7E00050000010016EA6300020021EA74000200C0")  // Temporary solution

	params.PayloadLen = len(newPayload)
	params.Payload = newPayload

	xapp.Logger.Info("SendSubFail: Forwarding failure to xApp: Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(&params)
	if err != nil {
		xapp.Logger.Error("SendSubFail: Failed to send response to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("SendSubFail: SubId: %v, from address: %v:%v. Deleting transaction record", int(subId), transaction.Xappkey.Addr, transaction.Xappkey.Port)

	xapp.Logger.Info("SubReqTimer: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := SubRouteInfo{DELETE, transaction.Xappkey.Addr, transaction.Xappkey.Port, subId}
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SendSubFail: Failed to update routing manager %v. SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SendSubFail: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(subId) {
		transaction, err = c.tracker.completeTransaction(subId, CREATE)
		if err != nil {
			xapp.Logger.Error("SendSubFail: Failed to delete transaction record. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
			return
		}
	} else {
		xapp.Logger.Error("SendSubFail: Failed to release sequency number. SubId: %v, Xid: %s", params.SubId, params.Xid)
	}
	return
}
*/

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
		xapp.Logger.Error("SubDelReq: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}
	xapp.Logger.Info("SubDelReq: Received payloadSeqNum: %v", payloadSeqNum)

	if c.registry.IsValidSequenceNumber(payloadSeqNum) {
		c.registry.deleteSubscription(payloadSeqNum)
		_, err = c.trackDeleteTransaction(params, payloadSeqNum)
		if err != nil {
			xapp.Logger.Error("SubDelReq: Failed to create transaction record. Dropping this msg. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
			return
		}
	} else {
		xapp.Logger.Error("SubDelReq: Not valid sequence number. Dropping this msg. SubId: %v, Xid: %s", params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SubDelReq: Forwarding Request to E2T. Mtype: %v, SubId: %v, Xid: %s, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
	c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("SubDelReq: Failed to send request to E2T. Err %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	} else {
		c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(payloadSeqNum), subReqTime, c.handleSubscriptionDeleteRequestTimer)
	}
	return
}

func (c *Control) trackDeleteTransaction(params *xapp.RMRParams, payloadSeqNum uint16) (transaction *Transaction, err error) {
	srcAddr, srcPort, err := c.rtmgrClient.SplitSource(params.Src)
	if err != nil {
		xapp.Logger.Error("SubDelReq: Failed to update routing-manager. Err: %s, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}
	transaction, err = c.tracker.TrackTransaction(payloadSeqNum, DELETE, *srcAddr, *srcPort, params)
	return
}

func (c *Control) handleSubscriptionDeleteResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Info("SubDelResp received from Src: %s, Mtype: %v, SubId: %v, Meid: %v", params.Src, params.Mtype, params.SubId, params.Meid)
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil

	payloadSeqNum, err := c.e2ap.GetSubscriptionDeleteResponseSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubDelResp: Received payloadSeqNum: %v", payloadSeqNum)

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(payloadSeqNum))

	transaction, err := c.tracker.RetriveTransaction(payloadSeqNum, DELETE)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Failed to retrive transaction record. Dropping this msg. Err: %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubDelResp: SubId: %v, from address: %v:%v. Forwarding response to xApp", int(payloadSeqNum), transaction.Xappkey.Addr, transaction.Xappkey.Port)

	params.SubId = int(payloadSeqNum)
	params.Xid = transaction.OrigParams.Xid
	xapp.Logger.Info("Forwarding SubDelResp to xApp: Mtype: %v, SubId: %v, Xid: %v, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Failed to send response to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		//		return
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("SubDelResp: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := SubRouteInfo{DELETE, transaction.Xappkey.Addr, transaction.Xappkey.Port, payloadSeqNum}
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SubDelResp: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
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
		xapp.Logger.Error("SubDelFail: Unable to get Sequence Number from Payload. Dropping this msg. Err: %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubDelFail: Received payloadSeqNum: %v", payloadSeqNum)

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(payloadSeqNum))

	transaction, err := c.tracker.RetriveTransaction(payloadSeqNum, DELETE)
	if err != nil {
		xapp.Logger.Error("SubDelFail: Failed to retrive transaction record. Dropping msg. Err %v, SubId: %v", err, params.SubId)
		return
	}
	xapp.Logger.Info("SubDelFail: SubId: %v, from address: %v:%v. Forwarding response to xApp", int(payloadSeqNum), transaction.Xappkey.Addr, transaction.Xappkey.Port)

	params.SubId = int(payloadSeqNum)
	params.Xid = transaction.OrigParams.Xid
	xapp.Logger.Info("Forwarding SubDelFail to xApp: Mtype: %v, SubId: %v, Xid: %v, Meid: %v", params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(params)
	if err != nil {
		xapp.Logger.Error("Failed to send SubDelFail to xApp. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		//		return
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("SubDelFail: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := SubRouteInfo{DELETE, transaction.Xappkey.Addr, transaction.Xappkey.Port, payloadSeqNum}
	c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SubDelFail: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SubDelFail: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(payloadSeqNum) {
		transaction, err = c.tracker.completeTransaction(payloadSeqNum, DELETE)
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

func (c *Control) handleSubscriptionDeleteRequestTimer(strId string, nbrId int) {
	newSubId := uint16(nbrId)
	xapp.Logger.Info("SubDelReq timer expired. newSubId: %v", newSubId)
	//	var causeContent uint8 = 1  // just some random cause. To be checked later. Should be no respose or something
	//	var causeVal uint8 = 1  // just some random val. To be checked later. Should be no respose or something
	//	c.sendSubscriptionDeleteFailure(newSubId, causeContent, causeVal)
}

/*
func (c *Control) sendSubscriptionDeleteFailure(subId uint16, causeContent uint8, causeVal uint8) {
	transaction, err := c.tracker.completeTransaction(subId, DELETE)
	if err != nil {
		xapp.Logger.Error("SendSubDelFail: Failed to delete transaction record. Err: %v, newSubId: %v", err, subId)
		return
	}
	xapp.Logger.Info("SendSubDelFail: SubId: %v, Xid %v, Meid: %v",subId, transaction.OrigParams.Xid, transaction.OrigParams.Meid)

	var params xapp.RMRParams
	params.Mtype = 12022 //xapp.RICMessageTypes["RIC_SUB_DEL_FAILURE"]
	params.SubId = int(subId)
	params.Meid = transaction.OrigParams.Meid
	params.Xid = transaction.OrigParams.Xid

//	newPayload, packErr := c.e2ap.PackSubscriptionDeleteFailure(transaction.OrigParams.Payload, subId, causeContent, causeVal)
//	if packErr != nil {
//		xapp.Logger.Error("SendSubDelFail: PackSubscriptionDeleteFailure(). Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid))
//		return
//	}

	newPayload := []byte("40CA4018000003EA7E00050000010016EA6300020021EA74000200C0")  // Temporary solution

	params.PayloadLen = len(newPayload)
	params.Payload = newPayload

	xapp.Logger.Info("SendSubDelFail: Forwarding failure to xApp: Mtype: %v, SubId: %v, Xid: %s, Meid: %v",params.Mtype, params.SubId, params.Xid, params.Meid)
	err = c.rmrReplyToSender(&params)
	if err != nil {
		xapp.Logger.Error("SendSubDelFail: Failed to send response to xApp: Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
	}

	time.Sleep(3 * time.Second)

	xapp.Logger.Info("SendSubDelFail: SubId: %v, from address: %v:%v. Deleting transaction record", int(subId), transaction.Xappkey.Addr, transaction.Xappkey.Port)

	xapp.Logger.Info("SendSubDelFail: Starting routing manager update. SubId: %v, Xid: %s", params.SubId, params.Xid)
	subRouteAction := SubRouteInfo{DELETE, transaction.Xappkey.Addr, transaction.Xappkey.Port, subId}
	err = c.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		xapp.Logger.Error("SendSubDelFail: Failed to update routing manager. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
		return
	}

	xapp.Logger.Info("SendSubDelFail: Deleting transaction record. SubId: %v, Xid: %s", params.SubId, params.Xid)
	if c.registry.releaseSequenceNumber(subId) {
		transaction, err = c.tracker.completeTransaction(subId, DELETE)
		if err != nil {
			xapp.Logger.Error("SendSubDelFail: Failed to delete transaction record. Err: %v, SubId: %v, Xid: %s", err, params.SubId, params.Xid)
			return
		}
	} else {
		xapp.Logger.Error("SendSubDelFail: Failed to release sequency number. SubId: %v, Xid: %s", params.SubId, params.Xid)
	}
	return
}
*/
