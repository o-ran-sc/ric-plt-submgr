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

/*
#include <rmr/RIC_message_types.h>
#include <rmr/rmr.h>

#cgo CFLAGS: -I../
#cgo LDFLAGS: -lrmr_nng -lnng
*/
import "C"

import (
	"errors"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/spf13/viper"
	"github.com/go-openapi/strfmt"
	httptransport "github.com/go-openapi/runtime/client"
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	rtmgrhandle "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client/handle"
	"math/rand"
	"strconv"
	"time"
)

type Control struct {
	e2ap        *E2ap
	registry    *Registry
	rtmgrClient *RtmgrClient
	tracker     *Tracker
}

type RMRMeid struct {
	PlmnID string
	EnbID  string
}

type RMRParams struct {
	Mtype           int
	Payload         []byte
	PayloadLen      int
	Meid            *RMRMeid
	Xid             string
	SubId           int
	Src             string
	Mbuf            *C.rmr_mbuf_t
}

var SEEDSN uint16
var SubscriptionReqChan = make(chan subRouteInfo, 10)

const (
	CREATE Action = 0
	MERGE Action = 1
	DELETE Action = 3
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("submgr")
	viper.AllowEmptyEnv(true)
	SEEDSN = uint16(viper.GetInt("seed_sn"))
	if SEEDSN == 0 {
		rand.Seed(time.Now().UnixNano())
		SEEDSN = uint16(rand.Intn(65535))
	}
	if SEEDSN > 65535 {
		SEEDSN = 0
	}
	xapp.Logger.Info("SUBMGR: Initial Sequence Number: %v", SEEDSN)
}

func NewControl() Control {
	registry := new(Registry)
	registry.Initialize(SEEDSN)

	tracker := new(Tracker)
	tracker.Init()

	transport := httptransport.New(viper.GetString("rtmgr.HostAddr") + ":" + viper.GetString("rtmgr.port"), viper.GetString("rtmgr.baseUrl"), []string{"http"})
	client := rtmgrclient.New(transport, strfmt.Default)
	handle := rtmgrhandle.NewProvideXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	rtmgrClient := RtmgrClient{client, handle}

	return Control{new(E2ap), registry, &rtmgrClient, tracker}
}

func (c *Control) Run() {
	xapp.Run(c)
}

func (c *Control) Consume(rp *xapp.RMRParams) (err error) {
	switch rp.Mtype {
	case C.RIC_SUB_REQ:
		err = c.handleSubscriptionRequest(rp)
	case C.RIC_SUB_RESP:
		err = c.handleSubscriptionResponse(rp)
	case C.RIC_SUB_DEL_REQ:
		err = c.handleSubscriptionDeleteRequest(rp)
	default:
		err = errors.New("Message Type " + strconv.Itoa(rp.Mtype) + " is discarded")
	}
	return
}

func (c *Control) rmrSend(params *xapp.RMRParams) (err error) {
	if !xapp.Rmr.Send(params, false) {
		err = errors.New("rmr.Send() failed")
	}
	return
}

func (c *Control) handleSubscriptionRequest(params *xapp.RMRParams) (err error) {
	payload_seq_num, err := c.e2ap.GetSubscriptionRequestSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Request Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payload_seq_num)

	/* Reserve a sequence number and set it in the payload */
	new_sub_id := c.registry.ReserveSequenceNumber()

	_, err = c.e2ap.SetSubscriptionRequestSequenceNumber(params.Payload, new_sub_id)
	if err != nil {
		err = errors.New("Unable to set Subscription Sequence Number in Payload due to: " + err.Error())
		return
	}

	src_addr, src_port, err := c.rtmgrClient.SplitSource(params.Src)
	if err != nil {
		xapp.Logger.Error("Failed to update routing-manager about the subscription request with reason: %s", err)
		return
	}

	/* Create transatcion records for every subscription request */
	xact_key := Transaction_key{new_sub_id, CREATE}
	xact_value := Transaction{*src_addr, *src_port, params.Payload}
	err = c.tracker.Track_transaction(xact_key, xact_value)
	if err != nil {
		xapp.Logger.Error("Failed to create a transaction record due to %v", err)
		return
	}

	/* Update routing manager about the new subscription*/
	sub_route_action := subRouteInfo{CREATE, *src_addr, *src_port, new_sub_id }
	go c.rtmgrClient.SubscriptionRequestUpdate()
	SubscriptionReqChan <- sub_route_action

	// Setting new subscription ID in the RMR header
	params.SubId = int(new_sub_id)

	xapp.Logger.Info("Generated ID: %v. Forwarding to E2 Termination...", int(new_sub_id))
	c.rmrSend(params)
	return
}

func (c *Control) handleSubscriptionResponse(params *xapp.RMRParams) (err error) {
	payload_seq_num, err := c.e2ap.GetSubscriptionResponseSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Response Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payload_seq_num)
	if !c.registry.IsValidSequenceNumber(payload_seq_num) {
		err = errors.New("Unknown Subscription ID: " + strconv.Itoa(int(payload_seq_num)) + " in Subscritpion Response. Message discarded.")
		return
	}
	c.registry.setSubscriptionToConfirmed(payload_seq_num)
	xapp.Logger.Info("Subscription Response Registered. Forwarding to Requestor...")
	c.rmrSend(params)
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
	payload_seq_num, err := c.e2ap.GetSubscriptionDeleteRequestSequenceNumber(params.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Delete Request Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payload_seq_num)
	if c.registry.IsValidSequenceNumber(payload_seq_num) {
		c.registry.deleteSubscription(payload_seq_num)
	}
	xapp.Logger.Info("Subscription ID: %v. Forwarding to E2 Termination...", int(payload_seq_num))
	c.rmrSend(params)
	return
}
