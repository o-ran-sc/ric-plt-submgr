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
	"encoding/json"
	"errors"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingControl struct {
	desc     string
	syncChan chan struct{}
}

func (tc *testingControl) ReadyCB(data interface{}) {
	xapp.Logger.Info("testingControl(%s) ReadyCB", tc.desc)
	tc.syncChan <- struct{}{}
	return
}

func (tc *testingControl) WaitCB() {
	<-tc.syncChan
}

func initTestingControl(desc string, rtfile string, port string) testingControl {
	tc := testingControl{}
	os.Setenv("RMR_SEED_RT", rtfile)
	os.Setenv("RMR_SRC_ID", "localhost:"+port)
	xapp.Logger.Info("Using rt file %s", os.Getenv("RMR_SEED_RT"))
	xapp.Logger.Info("Using src id  %s", os.Getenv("RMR_SRC_ID"))
	tc.desc = desc
	tc.syncChan = make(chan struct{})
	return tc
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingRmrControl struct {
	testingControl
	rmrClientTest *xapp.RMRClient
}

func (tc *testingRmrControl) RmrSend(params *xapp.RMRParams) (err error) {
	//
	//NOTE: Do this way until xapp-frame sending is improved
	//
	status := false
	i := 1
	for ; i <= 10 && status == false; i++ {
		status = tc.rmrClientTest.SendMsg(params)
		if status == false {
			xapp.Logger.Info("rmr.Send() failed. Retry count %v, Mtype: %v, SubId: %v, Xid %s", i, params.Mtype, params.SubId, params.Xid)
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = errors.New("rmr.Send() failed")
		tc.rmrClientTest.Free(params.Mbuf)
	}
	return
}

func initTestingRmrControl(desc string, rtfile string, port string, stat string, consumer xapp.MessageConsumer) testingRmrControl {
	tc := testingRmrControl{}
	tc.testingControl = initTestingControl(desc, rtfile, port)
	tc.rmrClientTest = xapp.NewRMRClientWithParams("tcp:"+port, 4096, 1, stat)
	tc.rmrClientTest.SetReadyCB(tc.ReadyCB, nil)
	go tc.rmrClientTest.Start(consumer)
	tc.WaitCB()
	return tc
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingMessageChannel struct {
	rmrConChan chan *xapp.RMRParams
}

func initTestingMessageChannel() testingMessageChannel {
	mc := testingMessageChannel{}
	mc.rmrConChan = make(chan *xapp.RMRParams)
	return mc
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type testingXappControl struct {
	testingRmrControl
	testingMessageChannel
	meid *xapp.RMRMeid
	xid  string
}

func (tc *testingXappControl) Consume(msg *xapp.RMRParams) (err error) {
	if msg.Xid == tc.xid {
		xapp.Logger.Info("testingXappControl(%s) Consume mtype=%s subid=%d xid=%s", tc.desc, xapp.RicMessageTypeToName[msg.Mtype], msg.SubId, msg.Xid)
		tc.rmrConChan <- msg
	} else {
		xapp.Logger.Info("testingXappControl(%s) Ignore mtype=%s subid=%d xid=%s, Expected xid=%s", tc.desc, xapp.RicMessageTypeToName[msg.Mtype], msg.SubId, msg.Xid, tc.xid)
	}
	return
}

func createNewXappControl(desc string, rtfile string, port string, stat string, ranname string, xid string) *testingXappControl {
	xappCtrl := &testingXappControl{}
	xappCtrl.testingRmrControl = initTestingRmrControl(desc, rtfile, port, stat, xappCtrl)
	xappCtrl.testingMessageChannel = initTestingMessageChannel()
	xappCtrl.meid = &xapp.RMRMeid{RanName: ranname}
	xappCtrl.xid = xid
	return xappCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingE2termControl struct {
	testingRmrControl
	testingMessageChannel
}

func (tc *testingE2termControl) Consume(msg *xapp.RMRParams) (err error) {
	xapp.Logger.Info("testingE2termControl(%s) Consume mtype=%s subid=%d xid=%s", tc.desc, xapp.RicMessageTypeToName[msg.Mtype], msg.SubId, msg.Xid)
	tc.rmrConChan <- msg
	return
}

func createNewE2termControl(desc string, rtfile string, port string, stat string) *testingE2termControl {
	e2termCtrl := &testingE2termControl{}
	e2termCtrl.testingRmrControl = initTestingRmrControl(desc, rtfile, port, stat, e2termCtrl)
	e2termCtrl.testingMessageChannel = initTestingMessageChannel()
	return e2termCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingMainControl struct {
	testingControl
	c *Control
}

func createNewMainControl(desc string, rtfile string, port string) *testingMainControl {
	mainCtrl = &testingMainControl{}
	mainCtrl.testingControl = initTestingControl(desc, rtfile, port)
	mainCtrl.c = NewControl()
	xapp.SetReadyCB(mainCtrl.ReadyCB, nil)
	go xapp.RunWithParams(mainCtrl.c, false)
	mainCtrl.WaitCB()
	return mainCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func testError(t *testing.T, pattern string, args ...interface{}) {
	xapp.Logger.Error(fmt.Sprintf(pattern, args...))
	t.Errorf(fmt.Sprintf(pattern, args...))
}

func testLog(t *testing.T, pattern string, args ...interface{}) {
	xapp.Logger.Info(fmt.Sprintf(pattern, args...))
	t.Logf(fmt.Sprintf(pattern, args...))
}

func testCreateTmpFile(str string) (string, error) {
	file, err := ioutil.TempFile("/tmp", "*.rt")
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(str)
	if err != nil {
		file.Close()
		return "", err
	}
	return file.Name(), nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

var xappConn1 *testingXappControl
var xappConn2 *testingXappControl
var e2termConn *testingE2termControl
var mainCtrl *testingMainControl

func TestMain(m *testing.M) {
	xapp.Logger.Info("TestMain start")

	//
	//Cfg creation won't work like this as xapp-frame reads it during init.
	//
	/*
	    cfgstr:=`{
	      "local": {
	          "host": ":8080"
	      },
	      "logger": {
	          "level": 4
	      },
	      "rmr": {
	         "protPort": "tcp:14560",
	         "maxSize": 4096,
	         "numWorkers": 1,
	         "txMessages": ["RIC_SUB_REQ", "RIC_SUB_DEL_REQ"],
	         "rxMessages": ["RIC_SUB_RESP", "RIC_SUB_FAILURE", "RIC_SUB_DEL_RESP", "RIC_SUB_DEL_FAILURE", "RIC_INDICATION"]
	      },
	      "db": {
	          "host": "localhost",
	          "port": 6379,
	          "namespaces": ["sdl", "rnib"]
	      },
	         "rtmgr" : {
	           "HostAddr" : "localhost",
	           "port" : "8989",
	           "baseUrl" : "/"
	         }
	   `

	   cfgfilename,_ := testCreateTmpFile(cfgstr)
	   defer os.Remove(cfgfilename)
	   os.Setenv("CFG_FILE", cfgfilename)
	*/
	xapp.Logger.Info("Using cfg file %s", os.Getenv("CFG_FILE"))

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### submgr main run ###")

	subsrt := `newrt|start
mse|12010|-1|localhost:14560
mse|12010,localhost:14560|-1|localhost:15560
mse|12011,localhost:15560|-1|localhost:14560
mse|12011|-1|localhost:13560;localhost:13660
mse|12012,localhost:15560|-1|localhost:14560
mse|12012|-1|localhost:13560;localhost:13660
mse|12020|-1|localhost:14560
mse|12020,localhost:14560|-1|localhost:15560
mse|12021,localhost:15560|-1|localhost:14560
mse|12021|-1|localhost:13560;localhost:13660
mse|12022,localhost:15560|-1|localhost:14560
mse|12022|-1|localhost:13560;localhost:13660
newrt|end
`
	subrtfilename, _ := testCreateTmpFile(subsrt)
	defer os.Remove(subrtfilename)
	mainCtrl = createNewMainControl("main", subrtfilename, "14560")

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### xapp1 rmr run ###")

	xapprt1 := `newrt|start
mse|12010|-1|localhost:14560
mse|12011|-1|localhost:13560
mse|12012|-1|localhost:13560
mse|12020|-1|localhost:14560
mse|12021|-1|localhost:13560
mse|12022|-1|localhost:13560
newrt|end
`

	xapprtfilename1, _ := testCreateTmpFile(xapprt1)
	defer os.Remove(xapprtfilename1)
	xappConn1 = createNewXappControl("xappConn1", xapprtfilename1, "13560", "RMRXAPP1STUB", "RAN_NAME_1", "XID_1")

	//---------------------------------
	//
	//---------------------------------

	xapp.Logger.Info("### xapp2 rmr run ###")

	xapprt2 := `newrt|start
mse|12010|-1|localhost:14560
mse|12011|-1|localhost:13660
mse|12012|-1|localhost:13660
mse|12020|-1|localhost:14560
mse|12021|-1|localhost:13660
mse|12022|-1|localhost:13660
newrt|end
`

	xapprtfilename2, _ := testCreateTmpFile(xapprt2)
	defer os.Remove(xapprtfilename2)
	xappConn2 = createNewXappControl("xappConn2", xapprtfilename2, "13660", "RMRXAPP2STUB", "RAN_NAME_1", "XID_2")

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### e2term rmr run ###")

	e2termrt := `newrt|start
mse|12010|-1|localhost:15560
mse|12011|-1|localhost:14560
mse|12012|-1|localhost:14560
mse|12020|-1|localhost:15560
mse|12021|-1|localhost:14560
mse|12022|-1|localhost:14560
newrt|end
`

	e2termrtfilename, _ := testCreateTmpFile(e2termrt)
	defer os.Remove(e2termrtfilename)
	e2termConn = createNewE2termControl("e2termConn", e2termrtfilename, "15560", "RMRE2TERMSTUB")

	//---------------------------------
	//
	//---------------------------------
	http_handler := func(w http.ResponseWriter, r *http.Request) {
		var req rtmgr_models.XappSubscriptionData
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
		}
		xapp.Logger.Info("(http handler) handling Address=%s Port=%d SubscriptionID=%d", *req.Address, *req.Port, *req.SubscriptionID)

		w.WriteHeader(200)
	}

	go func() {
		http.HandleFunc("/", http_handler)
		http.ListenAndServe("localhost:8989", nil)
	}()

	//---------------------------------
	//
	//---------------------------------
	code := m.Run()
	os.Exit(code)
}
