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

package xapptweaks

import (
	"bytes"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RMRParams struct {
	*xapp.RMRParams
}

func (params *RMRParams) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "params(Src=%s Mtype=%d SubId=%d Xid=%s Meid=%s Paylens=%d/%d)", params.Src, params.Mtype, params.SubId, params.Xid, params.Meid.RanName, params.PayloadLen, len(params.Payload))
	return b.String()
}

func NewParams(params *xapp.RMRParams) *RMRParams {
	if params != nil {
		return &RMRParams{params}
	}
	return &RMRParams{&xapp.RMRParams{}}
}
