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
	lines []string
}

func (rrt *RmrRouteTable) AddEntry(mtype int, src string, subid int, trg string) {

	line := "mse|"
	line += strconv.FormatInt(int64(mtype), 10)
	if len(src) > 0 {
		line += "," + src
	}
	line += "|"
	line += strconv.FormatInt(int64(subid), 10)
	line += "|"
	line += trg
	rrt.lines = append(rrt.lines, line)
}

func (rrt *RmrRouteTable) GetRt() string {
	allrt := "newrt|start\n"
	for _, val := range rrt.lines {
		allrt += val + "\n"
	}
	allrt += "newrt|end\n"
	return allrt
}
