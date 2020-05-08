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

package xapptweaks

import (
	"fmt"
	"strings"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func ConstructEndpointAddresses(clientEndpoint string) (string, string, error) {

	// Received clientEndpoint addres could be either: service-ricxapp-xappname-http.ricxapp:8080 or
	// service-ricxapp-xappname-rmr.ricxapp:4560
	if i := strings.Index(clientEndpoint, ":"); i == -1 {
		err := fmt.Errorf("Incorrect ClientEndpoint address format=%s. It should be address:port", clientEndpoint)
		return "", "", err
	}

	// xApp's http address need to be in this form: service-ricxapp-xappname-http.ricxapp
	xAppHttpEndPoint := clientEndpoint
	if i := strings.Index(xAppHttpEndPoint, ":"); i != -1 {
		// Remove port form the address
		xAppHttpEndPoint = xAppHttpEndPoint[0:i]
	}

	// Submgr's test address need to be in this form: localhost:13560
	if i := strings.Index(clientEndpoint, "localhost"); i != -1 {
		// Test address is used. clientEndpoint contains already the RMR address we need
		return xAppHttpEndPoint, clientEndpoint, nil
	}

	// xApp's RMR address should be in this form: service-ricxapp-xappname-rmr.ricxapp:4560
	var xAppRrmEndPoint string
	if i := strings.Index(clientEndpoint, "http"); i != -1 {
		// Fix http -> rmr
		xAppRrmEndPoint = strings.Replace(clientEndpoint, "http", "rmr", -1)
	}

	if i := strings.Index(xAppRrmEndPoint, "8080"); i != -1 {
		// Fix RMR port 8080 -> 4560
		xAppRrmEndPoint = strings.Replace(xAppRrmEndPoint, "8080", "4560", -1)
	}

	xapp.Logger.Info("xAppHttpEndPoint=%v, xAppRrmEndPoint=%v", xAppHttpEndPoint, xAppRrmEndPoint)

	return xAppHttpEndPoint, xAppRrmEndPoint, nil
}
