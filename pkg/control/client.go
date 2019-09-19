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
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	rtmgrhandle "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client/handle"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_models"
	"strings"
	"strconv"
	"errors"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

type RtmgrClient struct {
	rtClient         *rtmgrclient.RoutingManager
	xappHandleParams *rtmgrhandle.ProvideXappSubscriptionHandleParams
}

func (rc *RtmgrClient) SubscriptionRequestUpdate() error {
	xapp.Logger.Debug("SubscriptionRequestUpdate() invoked")
	subRouteAction := <-SubscriptionReqChan
	// Routing manager handles subscription id as int32 to accomodate -1 and uint16 values
	subID := int32(subRouteAction.SubID)

	xapp.Logger.Debug("Subscription action details received: ", subRouteAction)

	xappSubReq := rtmgr_models.XappSubscriptionData{&subRouteAction.Address, &subRouteAction.Port, &subID}

	switch subRouteAction.Command {
	case CREATE:
		_, postErr := rc.rtClient.Handle.ProvideXappSubscriptionHandle(rc.xappHandleParams.WithXappSubscriptionData(&xappSubReq))
		if postErr != nil && !(strings.Contains(postErr.Error(), "status 200"))  {
			xapp.Logger.Error("Updating routing manager about subscription id = %d failed with error: %v", subID, postErr)
			return postErr
		} else {
			xapp.Logger.Info("Succesfully updated routing manager about the subscription: %d", subID)
			return nil
		}
	default:
		return nil
	}
}

func (rc *RtmgrClient) SplitSource(src string) (*string, *uint16, error) {
	tcpSrc := strings.Split(src, ":")
	if len(tcpSrc) != 2 {
		err := errors.New("Unable to get the source details of the xapp. Check the source string received from the rmr.")
		return nil, nil, err
	}
	srcAddr := tcpSrc[0]
	xapp.Logger.Info("---Debugging Inside splitsource tcpsrc[0] = %s and tcpsrc[1]= %s ", tcpSrc[0], tcpSrc[1])
	srcPort, err := strconv.ParseUint(tcpSrc[1], 10, 16)
	if err != nil {
		return nil, nil, err
	}
	srcPortInt := uint16(srcPort)
	return &srcAddr, &srcPortInt, nil
}
