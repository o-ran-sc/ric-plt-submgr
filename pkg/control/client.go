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
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
	"strings"
)

type RtmgrClient struct {
	rtClient         *rtmgrclient.RoutingManager
	xappHandleParams *rtmgrhandle.ProvideXappSubscriptionHandleParams
	xappDeleteParams *rtmgrhandle.DeleteXappSubscriptionHandleParams
}

func (rc *RtmgrClient) SubscriptionRequestUpdate(subRouteAction SubRouteInfo) error {
	xapp.Logger.Debug("SubscriptionRequestUpdate() invoked")
	subID := int32(subRouteAction.SubID)
	xapp.Logger.Debug("Subscription action details received. subRouteAction.Command: %v, Address %s, Port %v, subID %v", int16(subRouteAction.Command), subRouteAction.Address, subRouteAction.Port, subID)
	xappSubReq := rtmgr_models.XappSubscriptionData{&subRouteAction.Address, &subRouteAction.Port, &subID}

	switch subRouteAction.Command {
	case CREATE:
		_, postErr := rc.rtClient.Handle.ProvideXappSubscriptionHandle(rc.xappHandleParams.WithXappSubscriptionData(&xappSubReq))
		if postErr != nil && !(strings.Contains(postErr.Error(), "status 200")) {
			xapp.Logger.Error("Updating routing manager about subscription id = %d failed with error: %v", subID, postErr)
			return postErr
		} else {
			xapp.Logger.Info("Succesfully updated routing manager about the subscription: %d", subID)
			return nil
		}
	case DELETE:
		_, _, deleteErr := rc.rtClient.Handle.DeleteXappSubscriptionHandle(rc.xappDeleteParams.WithXappSubscriptionData(&xappSubReq))
		if deleteErr != nil && !(strings.Contains(deleteErr.Error(), "status 200")) {
			xapp.Logger.Error("Deleting subscription id = %d  in routing manager, failed with error: %v", subID, deleteErr)
			return deleteErr
		} else {
			xapp.Logger.Info("Succesfully deleted subscription: %d in routing manager.", subID)
			return nil
		}
	default:
		xapp.Logger.Debug("Unknown subRouteAction.Command: %v, Address %s, Port %v, subID: %v", subRouteAction.Command, subRouteAction.Address, subRouteAction.Port, subID)
		return nil
	}
}

func (rc *RtmgrClient) SplitSource(src string) (*string, *uint16, error) {
	tcpSrc := strings.Split(src, ":")
	if len(tcpSrc) != 2 {
		err := errors.New("unable to get the source details of the xapp - check the source string received from the rmr")
		return nil, nil, err
	}
	srcAddr := tcpSrc[0]
	xapp.Logger.Debug("Debugging Inside splitsource tcpsrc[0] = %s and tcpsrc[1]= %s ", tcpSrc[0], tcpSrc[1])
	srcPort, err := strconv.ParseUint(tcpSrc[1], 10, 16)
	if err != nil {
		return nil, nil, err
	}
	srcPortInt := uint16(srcPort)
	return &srcAddr, &srcPortInt, nil
}
