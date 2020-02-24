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
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststubdummy"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststube2ap"
	"os"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CaseBegin(desc string) *teststub.TestWrapper {
	tent := teststub.NewTestWrapper(desc)
	tent.Logger.Info(desc)
	return tent
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

var xappConn1 *teststube2ap.E2Stub
var xappConn2 *teststube2ap.E2Stub
var e2termConn1 *teststube2ap.E2Stub
var e2termConn2 *teststube2ap.E2Stub
var rtmgrHttp *testingHttpRtmgrStub
var mainCtrl *testingSubmgrControl

var dummystub *teststubdummy.RmrDummyStub

func ut_test_init() {
	tent := CaseBegin("ut_test_init")

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
	tent.Logger.Info("Using cfg file %s", os.Getenv("CFG_FILE"))

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
	// 15560   e2term1 stub
	// 15660   e2term2 stub
	// 13560   xapp1 stub
	// 13660   xapp2 stub
	// 16560   dummy stub
	//
	//---------------------------------
	rt := &teststub.RmrRouteTable{}
	rt.AddRoute(12010, "", -1, "localhost:14560")
	rt.AddRoute(12010, "localhost:14560", -1, "%meid")
	rt.AddRoute(12011, "localhost:15560", -1, "localhost:14560")
	rt.AddRoute(12012, "localhost:15560", -1, "localhost:14560")
	rt.AddRoute(12011, "localhost:15660", -1, "localhost:14560")
	rt.AddRoute(12012, "localhost:15660", -1, "localhost:14560")
	rt.AddRoute(12011, "localhost:14560", -1, "localhost:13660;localhost:13560")
	rt.AddRoute(12012, "localhost:14560", -1, "localhost:13660;localhost:13560")
	rt.AddRoute(12020, "", -1, "localhost:14560")
	rt.AddRoute(12020, "localhost:14560", -1, "%meid")
	rt.AddRoute(12021, "localhost:15560", -1, "localhost:14560")
	rt.AddRoute(12022, "localhost:15560", -1, "localhost:14560")
	rt.AddRoute(12021, "localhost:15660", -1, "localhost:14560")
	rt.AddRoute(12022, "localhost:15660", -1, "localhost:14560")
	rt.AddRoute(12021, "localhost:14560", -1, "localhost:13660;localhost:13560")
	rt.AddRoute(12022, "localhost:14560", -1, "localhost:13660;localhost:13560")
	rt.AddRoute(55555, "", -1, "localhost:13660;localhost:13560;localhost:15560;localhost:15660;localhost:16560")

	rt.AddMeid("localhost:15560", []string{"RAN_NAME_1", "RAN_NAME_2"})
	rt.AddMeid("localhost:15660", []string{"RAN_NAME_11", "RAN_NAME_12"})

	rtfilename, _ := teststub.CreateTmpFile(rt.GetTable())
	defer os.Remove(rtfilename)
	tent.Logger.Info("table[%s]", rt.GetTable())
	//---------------------------------
	//
	//---------------------------------
	tent.Logger.Info("### submgr ctrl run ###")
	mainCtrl = createSubmgrControl(rtfilename, "14560")

	//---------------------------------
	//
	//---------------------------------
	tent.Logger.Info("### xapp1 stub run ###")
	xappConn1 = teststube2ap.CreateNewE2Stub("xappstub1", rtfilename, "13560", "RMRXAPP1STUB", 55555)

	//---------------------------------
	//
	//---------------------------------
	tent.Logger.Info("### xapp2 stub run ###")
	xappConn2 = teststube2ap.CreateNewE2Stub("xappstub2", rtfilename, "13660", "RMRXAPP2STUB", 55555)

	//---------------------------------
	//
	//---------------------------------
	tent.Logger.Info("### e2term1 stub run ###")
	e2termConn1 = teststube2ap.CreateNewE2termStub("e2termstub1", rtfilename, "15560", "RMRE2TERMSTUB1", 55555)

	//---------------------------------
	//
	//---------------------------------
	tent.Logger.Info("### e2term2 stub run ###")
	e2termConn2 = teststube2ap.CreateNewE2termStub("e2termstub2", rtfilename, "15660", "RMRE2TERMSTUB2", 55555)

	//---------------------------------
	// Just to test dummy stub
	//---------------------------------
	tent.Logger.Info("### dummy stub run ###")
	dummystub = teststubdummy.CreateNewRmrDummyStub("dummystub", rtfilename, "16560", "DUMMYSTUB", 55555)

	//---------------------------------
	// Testing message sending
	//---------------------------------
	if teststub.RmrStubControlWaitAlive(10, 55555, mainCtrl.c) == false {
		os.Exit(1)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestMain(m *testing.M) {
	CaseBegin("TestMain start")
	ut_test_init()
	code := m.Run()
	os.Exit(code)
}
