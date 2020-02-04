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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingRmrControl struct {
	desc     string
	mutex    sync.Mutex
	syncChan chan struct{}
}

func (tc *testingRmrControl) Lock() {
	tc.mutex.Lock()
}

func (tc *testingRmrControl) Unlock() {
	tc.mutex.Unlock()
}

func (tc *testingRmrControl) GetDesc() string {
	return tc.desc
}

func (tc *testingRmrControl) ReadyCB(data interface{}) {
	xapp.Logger.Info("testingRmrControl(%s) ReadyCB", tc.GetDesc())
	tc.syncChan <- struct{}{}
	return
}

func (tc *testingRmrControl) WaitCB() {
	<-tc.syncChan
}

func (tc *testingRmrControl) init(desc string, rtfile string, port string) {
	os.Setenv("RMR_SEED_RT", rtfile)
	os.Setenv("RMR_SRC_ID", "localhost:"+port)
	xapp.Logger.Info("Using rt file %s", os.Getenv("RMR_SEED_RT"))
	xapp.Logger.Info("Using src id  %s", os.Getenv("RMR_SRC_ID"))
	tc.desc = strings.ToUpper(desc)
	tc.syncChan = make(chan struct{})
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingRmrStubControl struct {
	testingRmrControl
	rmrConChan    chan *RMRParams
	rmrClientTest *xapp.RMRClient
	active        bool
	msgCnt        uint64
}

func (tc *testingRmrStubControl) GetMsgCnt() uint64 {
	return tc.msgCnt
}

func (tc *testingRmrStubControl) IncMsgCnt() {
	tc.msgCnt++
}

func (tc *testingRmrStubControl) DecMsgCnt() {
	if tc.msgCnt > 0 {
		tc.msgCnt--
	}
}

func (tc *testingRmrStubControl) TestMsgCnt(t *testing.T) {
	if tc.GetMsgCnt() > 0 {
		testError(t, "(%s) message count expected 0 but is %d", tc.GetDesc(), tc.GetMsgCnt())
	}
}

func (tc *testingRmrStubControl) RmrSend(params *RMRParams) (err error) {
	//
	//NOTE: Do this way until xapp-frame sending is improved
	//
	xapp.Logger.Info("(%s) RmrSend %s", tc.GetDesc(), params.String())
	status := false
	i := 1
	for ; i <= 10 && status == false; i++ {
		status = tc.rmrClientTest.SendMsg(params.RMRParams)
		if status == false {
			xapp.Logger.Info("(%s) RmrSend failed. Retry count %v, %s", tc.GetDesc(), i, params.String())
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = fmt.Errorf("(%s) RmrSend failed. Retry count %v, %s", tc.GetDesc(), i, params.String())
		xapp.Rmr.Free(params.Mbuf)
	}
	return
}

func (tc *testingRmrStubControl) init(desc string, rtfile string, port string, stat string, consumer xapp.MessageConsumer) {
	tc.active = false
	tc.testingRmrControl.init(desc, rtfile, port)
	tc.rmrConChan = make(chan *RMRParams)
	tc.rmrClientTest = xapp.NewRMRClientWithParams("tcp:"+port, 4096, 1, stat)
	tc.rmrClientTest.SetReadyCB(tc.ReadyCB, nil)
	go tc.rmrClientTest.Start(consumer)
	tc.WaitCB()
	allRmrStubs = append(allRmrStubs, tc)
}

var allRmrStubs []*testingRmrStubControl

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

var xappConn1 *testingXappStub
var xappConn2 *testingXappStub
var e2termConn *testingE2termStub
var rtmgrHttp *testingHttpRtmgrStub
var mainCtrl *testingSubmgrControl

func ut_test_init() {
	xapp.Logger.Info("ut_test_init")

	//---------------------------------
	//
	//---------------------------------
	rtmgrHttp = createNewHttpRtmgrStub("RTMGRSTUB", "8989")
	go rtmgrHttp.run()

	//---------------------------------
	//
	//---------------------------------

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
	// Static routetable for rmr
	//
	// NOTE: Routing table is configured so, that responses
	//       are duplicated to xapp1 and xapp2 instances.
	//       If XID is not matching xapp stub will just
	//       drop message. (Messages 12011, 12012, 12021, 12022)
	//
	// NOTE2: 55555 message type is for stub rmr connectivity probing
	//
	// NOTE3: Ports per entity:
	//
	// Port    Entity
	// -------------------
	// 14560   submgr
	// 15560   e2term stub
	// 13560   xapp1 stub
	// 13660   xapp2 stub
	//
	//---------------------------------

	allrt := `newrt|start
mse|12010|-1|localhost:14560
mse|12010,localhost:14560|-1|localhost:15560
mse|12011,localhost:15560|-1|localhost:14560
mse|12012,localhost:15560|-1|localhost:14560
mse|12011,localhost:14560|-1|localhost:13660;localhost:13560
mse|12012,localhost:14560|-1|localhost:13660;localhost:13560
mse|12020|-1|localhost:14560
mse|12020,localhost:14560|-1|localhost:15560
mse|12021,localhost:15560|-1|localhost:14560
mse|12022,localhost:15560|-1|localhost:14560
mse|12021,localhost:14560|-1|localhost:13660;localhost:13560
mse|12022,localhost:14560|-1|localhost:13660;localhost:13560
mse|55555|-1|localhost:13660;localhost:13560,localhost:15560
newrt|end
`

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### submgr ctrl run ###")
	subsrt := allrt
	subrtfilename, _ := testCreateTmpFile(subsrt)
	defer os.Remove(subrtfilename)
	mainCtrl = createSubmgrControl("main", subrtfilename, "14560")

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### xapp1 stub run ###")
	xapprt1 := allrt
	xapprtfilename1, _ := testCreateTmpFile(xapprt1)
	defer os.Remove(xapprtfilename1)
	xappConn1 = createNewXappStub("xappstub1", xapprtfilename1, "13560", "RMRXAPP1STUB")

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### xapp2 stub run ###")
	xapprt2 := allrt
	xapprtfilename2, _ := testCreateTmpFile(xapprt2)
	defer os.Remove(xapprtfilename2)
	xappConn2 = createNewXappStub("xappstub2", xapprtfilename2, "13660", "RMRXAPP2STUB")

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### e2term stub run ###")
	e2termrt := allrt
	e2termrtfilename, _ := testCreateTmpFile(e2termrt)
	defer os.Remove(e2termrtfilename)
	e2termConn = createNewE2termStub("e2termstub", e2termrtfilename, "15560", "RMRE2TERMSTUB")

	//---------------------------------
	// Testing message sending
	//---------------------------------
	var dummyBuf []byte = make([]byte, 100)

	params := &RMRParams{&xapp.RMRParams{}}
	params.Mtype = 55555
	params.SubId = -1
	params.Payload = dummyBuf
	params.PayloadLen = 100
	params.Meid = &xapp.RMRMeid{RanName: "NONEXISTINGRAN"}
	params.Xid = "THISISTESTFORSTUBS"
	params.Mbuf = nil

	status := false
	i := 1
	for ; i <= 10 && status == false; i++ {
		xapp.Rmr.Send(params.RMRParams, false)

		status = true
		for _, val := range allRmrStubs {
			if val.active == false {
				status = false
				break
			}
		}
		if status == true {
			break
		}
		xapp.Logger.Info("Sleep 0.5 secs and try routes again")
		time.Sleep(500 * time.Millisecond)
	}

	if status == false {
		xapp.Logger.Error("Could not initialize routes")
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestMain(m *testing.M) {
	xapp.Logger.Info("TestMain start")
	ut_test_init()
	code := m.Run()
	os.Exit(code)
}
