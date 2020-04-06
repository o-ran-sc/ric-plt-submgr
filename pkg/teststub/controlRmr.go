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
package teststub

import (
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"os"
	"strconv"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrControl struct {
	TestWrapper
	syncChan  chan struct{}
	DataPort  uint16
	RoutePort uint16
}

func (tc *RmrControl) ReadyCB(data interface{}) {
	tc.syncChan <- struct{}{}
	return
}

func (tc *RmrControl) WaitCB() {
	<-tc.syncChan
}

func (tc *RmrControl) Init(desc string, rtfile string, port uint16, rtport uint16) {
	tc.TestWrapper.Init(desc)
	tc.DataPort = port
	tc.RoutePort = rtport
	os.Setenv("RMR_SEED_RT", rtfile)
	xapp.Logger.Info("Using rt file %s", os.Getenv("RMR_SEED_RT"))
	if tc.DataPort > 0 {
		os.Setenv("RMR_SRC_ID", "localhost:"+strconv.FormatUint(uint64(tc.DataPort), 10))
		xapp.Logger.Info("Using src id  %s", os.Getenv("RMR_SRC_ID"))
	}
	if tc.RoutePort > 0 {
		os.Setenv("RMR_RTG_SVC", "localhost:"+strconv.FormatUint(uint64(tc.RoutePort), 10))
		xapp.Logger.Info("Using rtg svc  %s", os.Getenv("RMR_RTG_SVC"))
	}
	tc.syncChan = make(chan struct{})
}

func (tc *RmrControl) TestError(t *testing.T, pattern string, args ...interface{}) {
	tc.Logger.Error(fmt.Sprintf(pattern, args...))
	t.Errorf(fmt.Sprintf(pattern, args...))
}

func (tc *RmrControl) TestLog(t *testing.T, pattern string, args ...interface{}) {
	tc.Logger.Info(fmt.Sprintf(pattern, args...))
	t.Logf(fmt.Sprintf(pattern, args...))
}
