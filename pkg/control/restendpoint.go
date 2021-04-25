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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
	"strings"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func ConstructEndpointAddresses(clientEndpoint models.SubscriptionParamsClientEndpoint) (string, string, error) {

	if clientEndpoint.Host == "" {
		err := fmt.Errorf("Incorrect ClientEndpoint Host value =%s.", clientEndpoint.Host)
		return "", "", err
	}

	xAppHttpEndPoint := clientEndpoint.Host
	var xAppRrmEndPoint string
	// Submgr's test address need to be in this form: localhost:13560
	if i := strings.Index(clientEndpoint.Host, "localhost"); i != -1 {
		// Test address is used. clientEndpoint contains already the RMR address we need
		xAppRrmEndPoint = xAppHttpEndPoint + ":" + strconv.FormatInt(*clientEndpoint.HTTPPort, 10)
		return xAppHttpEndPoint, xAppRrmEndPoint, nil

	}

	if i := strings.Index(clientEndpoint.Host, "http"); i != -1 {
		// Fix http -> rmr
		xAppRrmEndPoint = strings.Replace(clientEndpoint.Host, "http", "rmr", -1)
	}
	xAppRrmEndPoint = xAppRrmEndPoint + ":" + strconv.FormatInt(*clientEndpoint.RMRPort, 10)

	xapp.Logger.Info("xAppHttpEndPoint=%v, xAppRrmEndPoint=%v", xAppHttpEndPoint, xAppRrmEndPoint)

	return xAppHttpEndPoint, xAppRrmEndPoint, nil
}
