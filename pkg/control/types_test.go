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
	"testing"
)

func TestAction(t *testing.T) {

	testActionString := func(t *testing.T, val int, str string) {
		if Action(val).String() != str {
			testError(t, "String for value %d expected %s got %s", val, str, Action(val).String())
		}
	}

	testActionString(t, 0, "CREATE")
	testActionString(t, 1, "MERGE")
	testActionString(t, 2, "NONE")
	testActionString(t, 3, "DELETE")
	testActionString(t, 5, "UNKNOWN")
	testActionString(t, 6, "UNKNOWN")
	testActionString(t, 7, "UNKNOWN")
	testActionString(t, 10, "UNKNOWN")
}
