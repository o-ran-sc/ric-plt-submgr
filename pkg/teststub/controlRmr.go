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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"os"
	"strings"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrControl struct {
	desc     string
	syncChan chan struct{}
}

func (tc *RmrControl) GetDesc() string {
	return tc.desc
}

func (tc *RmrControl) ReadyCB(data interface{}) {
	xapp.Logger.Info("RmrControl(%s) ReadyCB", tc.GetDesc())
	tc.syncChan <- struct{}{}
	return
}

func (tc *RmrControl) WaitCB() {
	xapp.Logger.Info("RmrControl(%s) WaitCb .... waiting", tc.GetDesc())
	<-tc.syncChan
	xapp.Logger.Info("RmrControl(%s) WaitCb .... done", tc.GetDesc())
}

func (tc *RmrControl) Init(desc string, rtfile string, port string) {
	os.Setenv("RMR_SEED_RT", rtfile)
	os.Setenv("RMR_SRC_ID", "localhost:"+port)
	xapp.Logger.Info("Using rt file %s", os.Getenv("RMR_SEED_RT"))
	xapp.Logger.Info("Using src id  %s", os.Getenv("RMR_SRC_ID"))
	tc.desc = strings.ToUpper(desc)
	tc.syncChan = make(chan struct{})
}
