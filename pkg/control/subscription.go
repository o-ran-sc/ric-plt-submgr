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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"sync"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Subscription struct {
	mutex      sync.Mutex                     // Lock
	valid      bool                           // valid
	registry   *Registry                      // Registry
	ReqId      e2ap.RequestId                 // ReqId (Requestor Id + Seq Nro a.k.a subsid)
	Meid       *xapp.RMRMeid                  // Meid/ RanName
	EpList     RmrEndpointList                // Endpoints
	TransLock  sync.Mutex                     // Lock transactions, only one executed per time for subs
	TheTrans   *Transaction                   // Ongoing transaction from xapp
	SubReqMsg  *e2ap.E2APSubscriptionRequest  // Subscription information
	SubRespMsg *e2ap.E2APSubscriptionResponse // Subscription information
	SubFailMsg *e2ap.E2APSubscriptionFailure  // Subscription information
}

func (s *Subscription) String() string {
	return "subs(" + s.ReqId.String() + "/" + s.Meid.RanName + "/" + s.EpList.String() + ")"
}

func (s *Subscription) GetReqId() e2ap.RequestId {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.ReqId
}

func (s *Subscription) GetRequestorId() uint32 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.ReqId.Id
}

func (s *Subscription) GetSeq() uint32 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.ReqId.Seq
}

func (s *Subscription) GetSubId() uint32 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.ReqId.Seq
}

func (s *Subscription) GetMeid() *xapp.RMRMeid {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.Meid != nil {
		return s.Meid
	}
	return nil
}

func (s *Subscription) IsTransactionReserved() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.TheTrans != nil {
		return true
	}
	return false

}

func (s *Subscription) GetTransaction() *Transaction {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.TheTrans
}

func (s *Subscription) WaitTransactionTurn(trans *Transaction) {
	s.TransLock.Lock()
	s.mutex.Lock()
	s.TheTrans = trans
	s.mutex.Unlock()
}

func (s *Subscription) ReleaseTransactionTurn(trans *Transaction) {
	s.mutex.Lock()
	if trans != nil && trans == s.TheTrans {
		s.TheTrans = nil
	}
	s.mutex.Unlock()
	s.TransLock.Unlock()
}

func (s *Subscription) IsMergeable(trans *Transaction, subReqMsg *e2ap.E2APSubscriptionRequest) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.valid == false {
		return false
	}

	if s.SubReqMsg == nil {
		return false
	}

	if s.Meid.RanName != trans.Meid.RanName {
		return false
	}

	// EventTrigger check
	if s.SubReqMsg.EventTriggerDefinition.InterfaceDirection != subReqMsg.EventTriggerDefinition.InterfaceDirection ||
		s.SubReqMsg.EventTriggerDefinition.ProcedureCode != subReqMsg.EventTriggerDefinition.ProcedureCode ||
		s.SubReqMsg.EventTriggerDefinition.TypeOfMessage != subReqMsg.EventTriggerDefinition.TypeOfMessage {
		return false
	}

	if s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Val[0] != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Val[0] ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Val[1] != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Val[1] ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Val[2] != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Val[2] {
		return false
	}

	if s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.Present != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.Present ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.NodeId != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.NodeId ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.PlmnIdentity.Val[0] != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.PlmnIdentity.Val[0] ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.PlmnIdentity.Val[1] != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.PlmnIdentity.Val[1] ||
		s.SubReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.PlmnIdentity.Val[2] != subReqMsg.EventTriggerDefinition.InterfaceId.GlobalGnbId.PlmnIdentity.Val[2] {
		return false
	}

	// Actions check
	if len(s.SubReqMsg.ActionSetups) != len(subReqMsg.ActionSetups) {
		return false
	}

	for _, acts := range s.SubReqMsg.ActionSetups {
		for _, actt := range subReqMsg.ActionSetups {
			if acts.ActionId != actt.ActionId {
				return false
			}
			if acts.ActionType != actt.ActionType {
				return false
			}

			if acts.ActionType != e2ap.E2AP_ActionTypeReport {
				return false
			}

			if acts.ActionDefinition.Present != actt.ActionDefinition.Present ||
				acts.ActionDefinition.StyleId != actt.ActionDefinition.StyleId ||
				acts.ActionDefinition.ParamId != actt.ActionDefinition.ParamId {
				return false
			}
			if acts.SubsequentAction.Present != actt.SubsequentAction.Present ||
				acts.SubsequentAction.Type != actt.SubsequentAction.Type ||
				acts.SubsequentAction.TimetoWait != actt.SubsequentAction.TimetoWait {
				return false
			}
		}
	}

	return true
}
