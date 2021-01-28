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
	"testing"
)

func TestAddAllCountersOnce(t *testing.T) {

	// Check that all counters can be added correctly
	mainCtrl.SetTimesCounterWillBeAdded(cSubReqFromXapp, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubRespToXapp, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubFailToXapp, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubReqToE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubReReqToE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubRespFromE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubFailFromE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubReqTimerExpiry, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cRouteCreateFail, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cRouteCreateUpdateFail, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cMergedSubscriptions, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubDelReqFromXapp, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubDelRespToXapp, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubDelReqToE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubDelReReqToE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubDelRespFromE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubDelFailFromE2, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSubDelReqTimerExpiry, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cRouteDeleteFail, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cRouteDeleteUpdateFail, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cUnmergedSubscriptions, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSDLWriteFailure, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSDLReadFailure, 1)
	mainCtrl.SetTimesCounterWillBeAdded(cSDLRemoveFailure, 1)

	mainCtrl.GetCounterValuesBefore(t)

	mainCtrl.c.UpdateCounter(cSubReqFromXapp)
	mainCtrl.c.UpdateCounter(cSubRespToXapp)
	mainCtrl.c.UpdateCounter(cSubFailToXapp)
	mainCtrl.c.UpdateCounter(cSubReqToE2)
	mainCtrl.c.UpdateCounter(cSubReReqToE2)
	mainCtrl.c.UpdateCounter(cSubRespFromE2)
	mainCtrl.c.UpdateCounter(cSubFailFromE2)
	mainCtrl.c.UpdateCounter(cSubReqTimerExpiry)
	mainCtrl.c.UpdateCounter(cRouteCreateFail)
	mainCtrl.c.UpdateCounter(cRouteCreateUpdateFail)
	mainCtrl.c.UpdateCounter(cMergedSubscriptions)
	mainCtrl.c.UpdateCounter(cSubDelReqFromXapp)
	mainCtrl.c.UpdateCounter(cSubDelRespToXapp)
	mainCtrl.c.UpdateCounter(cSubDelReqToE2)
	mainCtrl.c.UpdateCounter(cSubDelReReqToE2)
	mainCtrl.c.UpdateCounter(cSubDelRespFromE2)
	mainCtrl.c.UpdateCounter(cSubDelFailFromE2)
	mainCtrl.c.UpdateCounter(cSubDelReqTimerExpiry)
	mainCtrl.c.UpdateCounter(cRouteDeleteFail)
	mainCtrl.c.UpdateCounter(cRouteDeleteUpdateFail)
	mainCtrl.c.UpdateCounter(cUnmergedSubscriptions)
	mainCtrl.c.UpdateCounter(cSDLWriteFailure)
	mainCtrl.c.UpdateCounter(cSDLReadFailure)
	mainCtrl.c.UpdateCounter(cSDLRemoveFailure)

	mainCtrl.VerifyCounterValues(t)
}
