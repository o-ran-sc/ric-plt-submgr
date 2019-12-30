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
	"fmt"
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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingRmrControl struct {
	testingControl
	rmrClientTest *xapp.RMRClient
	rmrConChan    chan *xapp.RMRParams
}

func (tc *testingRmrControl) Consume(msg *xapp.RMRParams) (err error) {
	xapp.Logger.Info("testingRmrControl(%s) Consume", tc.desc)
	tc.rmrConChan <- msg
	return
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

func createNewRmrControl(desc string, rtfile string, port string, stat string) *testingRmrControl {
	os.Setenv("RMR_SEED_RT", rtfile)
	os.Setenv("RMR_SRC_ID", "localhost:"+port)
	xapp.Logger.Info("Using rt file %s", os.Getenv("RMR_SEED_RT"))
	xapp.Logger.Info("Using src id  %s", os.Getenv("RMR_SRC_ID"))
	newConn := &testingRmrControl{}
	newConn.desc = desc
	newConn.syncChan = make(chan struct{})
	newConn.rmrClientTest = xapp.NewRMRClientWithParams("tcp:"+port, 4096, 1, stat)
	newConn.rmrConChan = make(chan *xapp.RMRParams)
	newConn.rmrClientTest.SetReadyCB(newConn.ReadyCB, nil)
	go newConn.rmrClientTest.Start(newConn)
	<-newConn.syncChan
	return newConn
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingMainControl struct {
	testingControl
	c *Control
}

func (mc *testingMainControl) wait_subs_clean(e2SubsId int, secs int) bool {
	i := 1
	for ; i <= secs*2; i++ {
		if mc.c.registry.IsValidSequenceNumber(uint16(e2SubsId)) == false {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func testError(t *testing.T, pattern string, args ...interface{}) {
	xapp.Logger.Error(fmt.Sprintf(pattern, args...))
	t.Errorf(fmt.Sprintf(pattern, args...))
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

var xappConn *testingRmrControl
var e2termConn *testingRmrControl
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
mse|12011|-1|localhost:13560
mse|12012,localhost:15560|-1|localhost:14560
mse|12012|-1|localhost:13560
mse|12020|-1|localhost:14560
mse|12020,localhost:14560|-1|localhost:15560
mse|12021,localhost:15560|-1|localhost:14560
mse|12021|-1|localhost:13560
mse|12022,localhost:15560|-1|localhost:14560
mse|12022|-1|localhost:13560
newrt|end
`

	subrtfilename, _ := testCreateTmpFile(subsrt)
	defer os.Remove(subrtfilename)
	os.Setenv("RMR_SEED_RT", subrtfilename)
	os.Setenv("RMR_SRC_ID", "localhost:14560")
	xapp.Logger.Info("Using rt file %s", os.Getenv("RMR_SEED_RT"))
	xapp.Logger.Info("Using src id  %s", os.Getenv("RMR_SRC_ID"))

	mainCtrl = &testingMainControl{}
	mainCtrl.desc = "main"
	mainCtrl.syncChan = make(chan struct{})

	mainCtrl.c = NewControl()
	xapp.SetReadyCB(mainCtrl.ReadyCB, nil)
	go xapp.RunWithParams(mainCtrl.c, false)
	<-mainCtrl.syncChan

	//---------------------------------
	//
	//---------------------------------
	xapp.Logger.Info("### xapp rmr run ###")

	xapprt := `newrt|start
mse|12010|-1|localhost:14560
mse|12011|-1|localhost:13560
mse|12012|-1|localhost:13560
mse|12020|-1|localhost:14560
mse|12021|-1|localhost:13560
mse|12022|-1|localhost:13560
newrt|end
`

	xapprtfilename, _ := testCreateTmpFile(xapprt)
	defer os.Remove(xapprtfilename)
	xappConn = createNewRmrControl("xappConn", xapprtfilename, "13560", "RMRXAPPSTUB")

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
	e2termConn = createNewRmrControl("e2termConn", e2termrtfilename, "15560", "RMRE2TERMSTUB")

	//---------------------------------
	//
	//---------------------------------
	http_handler := func(w http.ResponseWriter, r *http.Request) {
		xapp.Logger.Info("(http handler) handling")
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
