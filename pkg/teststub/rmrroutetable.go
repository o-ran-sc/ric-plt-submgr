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
	"strconv"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type RmrRouteTable struct {
	routes []string
	meids  []string
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

func (rrt *RmrRouteTable) GetTable() string {
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
