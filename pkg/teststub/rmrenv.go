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
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/xapptweaks"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"os"
	"strconv"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type RmrRouteTable struct {
	tmpfile string
	routes  []string
	meids   []string
}

func (rrt *RmrRouteTable) AddRoute(mtype int, src string, subid int, trg string) {

	line := "mse|"
	line += strconv.FormatInt(int64(mtype), 10)
	if len(src) > 0 {
		line += "," + src
	}
	line += "|"
	line += strconv.FormatInt(int64(subid), 10)
	line += "|"
	line += trg
	rrt.routes = append(rrt.routes, line)
}

func (rrt *RmrRouteTable) AddMeid(trg string, meids []string) {

	line := "mme_ar"
	line += "|"
	line += trg
	line += "|"
	for _, str := range meids {
		line += " " + str
	}
	rrt.meids = append(rrt.meids, line)
}

func (rrt *RmrRouteTable) DelMeid(meids []string) {

	line := "mme_del"
	line += "|"
	for _, str := range meids {
		line += " " + str
	}
	rrt.meids = append(rrt.meids, line)
}

func (rrt *RmrRouteTable) FileName() string {
	return rrt.tmpfile
}

func (rrt *RmrRouteTable) Table() string {
	allrt := "newrt|start\n"
	for _, val := range rrt.routes {
		allrt += val + "\n"
	}
	allrt += "newrt|end\n"
	allrt += "meid_map | start\n"
	for _, val := range rrt.meids {
		allrt += val + "\n"
	}
	allrt += "meid_map | end | " + strconv.FormatInt(int64(len(rrt.meids)), 10) + "\n"
	return allrt
}

func (rrt *RmrRouteTable) Enable() {
	if len(rrt.tmpfile) > 0 {
		os.Remove(rrt.tmpfile)
	}
	rrt.tmpfile, _ = CreateTmpFile(rrt.Table())
	os.Setenv("RMR_SEED_RT", rrt.tmpfile)
	xapp.Logger.Info("Using rt file %s", os.Getenv("RMR_SEED_RT"))
}

func (rrt *RmrRouteTable) Disable() {
	if len(rrt.tmpfile) > 0 {
		os.Remove(rrt.tmpfile)
		os.Unsetenv("RMR_SEED_RT")
		rrt.tmpfile = ""
		xapp.Logger.Info("Not using rt file ")
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type RmrSrcId struct {
	xapptweaks.RmrEndpoint
}

func (rsi *RmrSrcId) Enable() {
	if rsi.Port > 0 {
		os.Setenv("RMR_SRC_ID", rsi.String())
		xapp.Logger.Info("Using src id  %s", os.Getenv("RMR_SRC_ID"))
	}
}

func (rsi *RmrSrcId) Disable() {
	os.Unsetenv("RMR_SRC_ID")
	xapp.Logger.Info("Not using Using src id")
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrRtgSvc struct {
	xapptweaks.RmrEndpoint
}

func (rrs *RmrRtgSvc) Enable() {
	if rrs.Port > 0 {
		os.Setenv("RMR_RTG_SVC", rrs.String())
		xapp.Logger.Info("Using rtg svc  %s", os.Getenv("RMR_SRC_ID"))
	}
}

func (rrs *RmrRtgSvc) Disable() {
	os.Unsetenv("RMR_RTG_SVC")
	xapp.Logger.Info("Not using Using rtg svc")
}
