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

package e2ap

import (
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type IndicationHeader struct {
	InterfaceId
	InterfaceDirection uint32
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
//type InterfaceMessage struct {
//  Data []uint8
//}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type IndicationMessage struct {
	InterfaceMessage packer.PackedData
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APIndication struct {
	RequestId
	FunctionId
	IndicationHeader
	IndicationMessage
	//	CallProcessId
	IndicationSn   int32
	IndicationType uint64
}
