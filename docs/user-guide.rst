..
..  Copyright (c) 2019 AT&T Intellectual Property.
..  Copyright (c) 2019 Nokia.
..
..  Licensed under the Creative Commons Attribution 4.0 International
..  Public License (the "License"); you may not use this file except
..  in compliance with the License. You may obtain a copy of the License at
..
..    https://creativecommons.org/licenses/by/4.0/
..
..  Unless required by applicable law or agreed to in writing, documentation
..  distributed under the License is distributed on an "AS IS" BASIS,
..  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
..
..  See the License for the specific language governing permissions and
..  limitations under the License.
..

User-Guide
==========

.. contents::
   :depth: 3
   :local:

Overview
--------
Subscription Manager is a basic platform service in RIC. It is responsible for managing E2 subscriptions from xApps to the
Radio Access Network (RAN).

xApp can subscribe and unsubscribe messages from gNodeB through Subscription Manager. Subscription Manager manages the subscriptions
and message routing of the subscribed messages between E2 Termination and xApps. If one xApp has already made a subscription and then
another xApp initiates identical subscription, Subscription Manager does not forward the subscription to gNodeB but merges the
subscriptions internally. In merge case Subscription Manager just updates the message routing information to Routing Manager and
sends response to xApp.

There can be only one ongoing RIC Subscription or RIC Subscription Delete procedure towards RAN at any time. That is because Subscription
Manager is able to merge new subscriptions only which those it has already received successful response from RAN. Subscriptions
and delete subscriptions are therefore queued in Subscription Manager. Subscription Manager may need to do reties during subscribe or
unsubscribe procedure. As it can increase completion time of the procedure, this needs to be considered when retries are implemented
in xApp side. xApp's retry delay should not be too short.

    .. image:: images/PlaceInRICSoftwareArchitecture.png
      :width: 600
      :alt: Place in RIC's software architecture picture

Architecture
------------

  * Message routing

      Subscribed messages from RAN are transported to RIC inside RIC Indication message. RIC Indication message is transported to xApp
      inside RMR message, in Payload field of the RMR message. RMR message is routed to xApp based on SubId field (subscription id) in
      the RMR header. Same routing mechanism is used also for response messages from Subscription Manager to xApp. Subscription Manager is
      not able to respond to xApp if route for subscription has not been created.

      When xApp sends message to Subscription Manager it sets -1 in SubId field in the RMR header. It means that messages are routed based
      on message type (Mtype filed in RMR header). RIC Subscription Request and RIC Subscription delete Request messages are pre configured
      to be routed to Subscription Manager.

      Subscription Manager allocates unique RIC Request Sequence Number for every subscription during Subscription procedure. Subscription
      Manager replaces existing ASN.1 decoded RIC Request Sequence Number in the RIC Subscription Request message allocated by the xApp.
      That sequence number (subscription id) is then used for the subscription in RIC and RAN as long the subscription lives. xApp gets
      the sequence number in RIC Subscription Response message from Subscription manager based on the message types.
      
      Subscribed messages are routed to xApps based on sequence number. Sequence number is placed in the SubId field of the RMR message
      header when E2 Termination sends the subscribed message to xApp. When xApp wants to delete the subscription, the same sequence number
      must be included in the ASN.1 encoded RIC Subscription Delete Request message sent to Subscription Manager.

      Subscription Manager responds to xApp with xApp allocated RIC Requestor ID. In merge case subscription is created only for the first
      requestor. RAN gets the Requestor ID of the xApp who makes the first subscription. RAN uses that Requestor ID in all RIC Indication
      messages it sends to RIC for the subscription. Therefore xApp may get Requestor ID in RIC Indication message that belongs to another xApp.
      The xApp whose subscription is merged into the first subscription will also get Requestor ID of the first subscribed xApp in the RIC
      Subscription Response and RIC Subscription Delete Response messages.

      TransactionId (Xid) in RMR message header is used to track messages between xApp and Subscription Manager. xApp allocates it. Subscription
      Manager returns TransactionId received from xApp in response message to xApp. xApp uses it to map response messages to request messages
      it sends.

  * Subscription procedure
      
    * Successful case

      xApp sends RIC Subscription Request message to Subscription Manager. Subscription Manager validates request types in the message and sends
      route create to Routing Manager over REST interface. When route is created successfully Subscription Manager forwards request to E2
      Termination. When RIC Subscription Response arrives from E2 Termination Subscription Manager forwards it to xApp.
      
      Subscription Manager supervises route create and RIC Subscription Request with a timer.

      RIC Indication messages which are used to transport subscribed messages from RAN are routed from E2 Termination to xApps
      directly using the routes created during Subscription procedure.

      ``Routing manager has 1 second delay in routing create in R3 release before it responds to Subscription Manager. That is because of delay in route create to RMR.``

      Subscription Manager supports REPORT and POLICY type subscriptions (RICActionTypes). CONTROL and INSERT are not supported. POLICY type
      subscription can be updated. In update case signaling sequence is the same as above, except route is not created to Routing manager.
      xApp uses initially allocated TransactionId and RIC Request Sequence Number in update case. Route in POLICY type subscription case is needed
      only that Subscription Manager can send response messages to xApp. RIC Subscription Request message contains list of RICaction-ToBeSetup-ItemIEs.
      The list cannot have both REPORT and POLICY action types at the same time. Subscription Manager checks actions types in the message.
      If both action types is found the message is dropped.


    .. image:: images/Successful_Subscription.png
      :width: 600
      :alt: Successful subscription picture


    * Failure case

      In failure case Subscription Manager checks the failure cause and acts based on that. If failure cause is "duplicate" Subscription
      Manager sends delete to RAN and then resends the same subscription. If failure cause is such that Subscription manager cannot do
      anything to fix the problem, it sends delete to RAN and sends RIC Subscription Failure to xApp. Subscription Manager may retry RIC
      Subscription Request and RIC Subscription Delete messages also in this case before it responds to xApp.

    .. image:: images/Subscription_Failure.png
      :width: 600
      :alt: Subscription failure picture

    * Timeout case

     In case of timeout in Subscription Manager, Subscription Manager may resend the RIC Subscription Request to RAN. If there is no response
      after retry, Subscription Manager shall NOT send any response to xApp. xApp may retry RIC Subscription Request, if it wishes to do so.
      Subscription Manager does no handle the retry if Subscription Manager has ongoing subscription procedure for the same subscription.
      Subscription just drops the request.

    .. image:: images/Subscription_Timeout.png
      :width: 600
      :alt: Subscription timeout picture

  * Subscription delete procedure

    * Successful case

      xApp sends RIC Subscription Delete Request message to Subscription Manager. xApp must use the same RIC Request Sequence Number which
      it received in RIC Subscription Response message when subscription is deleted. When Subscription Manager receives RIC Subscription
      Delete Request message, Subscription Manager first forwards the request to E2 Termination. When RIC Subscription Delete Response arrives
      from E2 Termination to Subscription Manager, Subscription Manager forwards it to xApp and then request route deletion from Routing Manager.
      
      Subscription Manager supervises RIC Subscription Deletion Request and route delete with a timer.

    .. image:: images/Successful_Subscription_Delete.png
      :width: 600
      :alt: Successful subscription delete picture

    * Failure case

      Delete procedure cannot fail from xApp point of view. Subscription Manager always responds with RIC Subscription Delete Response to xApp.

    .. image:: images/Subscription_Delete_Failure.png
      :width: 600
      :alt: Subscription delete failure picture

    * Timeout case

      In case of timeout in Subscription Manager, Subscription Manager may resend the RIC Subscription Delete Request to RAN. If there is no
      response after retry, Subscription Manager responds to xApp with RIC Subscription Delete Response.

    .. image:: images/Subscription_Delete_Timeout.png
      :width: 600
      :alt: Subscription delete timeout picture

    * Unknown subscription

      If Subscription Manager receives RIC Subscription Delete Request for a subscription which does not exist, Subscription Manager cannot respond
      to xApp as there is no route for the subscription.

  * Subscription merge procedure

    * Successful case

      xApp sends RIC Subscription Request message to Subscription Manager. Subscription Manager checks is the Subscription mergeable. If not,
      Subscription Manager continues with normal Subscription procedure. If Subscription is mergeable, Subscription Manager sends route create
      to Routing Manager and then responds with RIC Subscription Response to xApp.
      
      Route create is supervised with a timer.

      Merge for REPORT type subscription is possible if Action Type and Event Trigger Definition of subscriptions are equal.

      ``Only REPORT type subscriptions can be be merged.``

    .. image:: images/Successful_Subscription_Merge.png
      :width: 600
      :alt: Successful subscription merge picture

    * Failure case

      Failure case is basically the same as in normal subscription procedure. Failure can come only from RAN when merge is not yet done.
      If error happens during route create Subscription Manager drops the RIC Subscription Request message and xApp does not get any response.

    * Timeout case

      Timeout case is basically the same as in normal subscription procedure but timeout can come only in route create during merge operation.
      If error happens during route create, Subscription Manager drops the RIC Subscription Request message and xApp does not get any response.

  * Subscription delete merge procedure

    * Successful case

      xApp sends RIC Subscription Delete Request message to Subscription Manager. If delete concerns merged subscription, Subscription Manager
      responds with RIC Subscription Delete Response to xApp and then sends route delete request to Routing manager.
      
      Subscription Manager supervises route delete with a timer.

    .. image:: images/Successful_Subscription_Delete_Merge.png
      :width: 600
      :alt: Successful subscription delete merge picture

    * Failure case

      Delete procedure cannot fail from xApp point of view. Subscription Manager responds with RIC Subscription Delete Response message to xApp.

    * Timeout case

      Timeout can only happen in route delete to Routing manager. Subscription Manager responds with RIC Subscription Delete Response message to xApp.

  * Unknown message

     If Subscription Manager receives unknown message, Subscription Manager drops the message.

RAN services explained
----------------------
  RIC hosted xApps may use the following RAN services from a RAN node:

  *  REPORT: RIC requests that RAN sends a REPORT message to RIC and continues further call processing in RAN after each occurrence of a defined SUBSCRIPTION
  *  INSERT: RIC requests that RAN sends an INSERT message to RIC and suspends further call processing in RAN after each occurrence of a defined SUBSCRIPTION
  *  CONTROL: RIC sends a Control message to RAN to initiate or resume call processing in RAN
  *  POLICY: RIC requests that RAN executes a specific POLICY during call processing in RAN after each occurrence of a defined SUBSCRIPTION

Supported E2 procedures and RAN services
----------------------------------------
    * RIC Subscription procedure with following RIC action types:

      - RIC action type REPORT
      - RIC action type POLICY

    * RIC Subscription Delete procedure

    * Merge and delete of equal REPORT type subscriptions.

Recommendations for xApps
-------------------------

   * Recommended retry delay

     Recommended retry delay for xApp is > 10 seconds
