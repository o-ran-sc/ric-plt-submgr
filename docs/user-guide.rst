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

User-Guide (new)
================

.. contents::
   :depth: 3
   :local:

Overview
--------
Subscription Manager is a basic platform service in RIC. It is responsible for managing E2 subscriptions from xApps to the
E2 Node (eNodeB or gNodeB).

xApp can subscribe and unsubscribe messages from E2 Node through Subscription Manager. Subscription Manager manages subscriptions
and message routing of the subscribed messages between E2 Termination and xApps. If one xApp has already made a subscription
and then another xApp initiates identical subscription, Subscription Manager does not forward the subscription to E2 Node but merges the
subscriptions internally. In merge case Subscription Manager just updates the message routing information to Routing Manager and
sends response to xApp.

Interface between xApp and Subscription Manager is HTTP based REST interface. Interface codes are generated with help of Swagger from a
yaml definition file. REST interface is used also between Subscription Manager and Routing Manager. Subscription Manager and
E2 Termination interface is based on RMR messages. xApp should use also Swagger generated code when it implements subscription REST
interface towards Subscription Manager.

    .. image:: images/PlaceInRICSoftwareArchitecture.png
      :width: 600
      :alt: Place in RIC's software architecture picture


One xApp generated REST subscription message can contain multiple E2 subscriptions. For every E2 subscription in the message there is also
xApp generated xApp instance id. In E2 interface there can be only one ongoing E2 subscription or subscription delete procedure towards
E2 Node at any time. That is because Subscription Manager is able to merge new E2 subscriptions only which those it has already received
successful response from E2 Node. E2 subscriptions and delete subscriptions may be therefore queued in Subscription Manager.

Subscription Manager may need to do reties towards E2 Node during subscribe or subscription delete procedure. Retries will increase completion
time of the procedure. This needs to be considered in xApp's implementation. Subscription Manager sends REST notification to xApp for every
completed E2 subscription procedure regardless is the E2 subscription successful or not. Notification is not sent for E2 subscription delete
procedures. Subscription Manager allocates globally unique REST request id for each new REST subscription request. That is returned to xApp in
response. When xApp wants to delete REST subscription, xApp need to use the same id in deletion request.

Subscription Manager allocates unique id also for the E2 subscriptions (E2 instance id). The id called 'InstanceId' in E2 specification.
In successful case the REST notification contains the id generated for the REST request, xApp instance id and E2 instance id. From xApp point
of view xApp instance id identifies received REST notification for the E2 subscription in the REST request. REST notification contains also Subscription
Manager generated E2 instance id. xApp can use that to map received E2 Indication message to E2 subscription. If E2 subscription procedure is unsuccessful
then E2 instance id is 0 and the notification contains non-zero error cause string.

xApp need to be able preserve Subscription Manager allocated REST request id over xApp restart. The id is needed for deletion of the REST
subscription and if there is need to resend the same REST request.  

Three different type of subscriptions are supported. REPORT, POLICY and INSERT. REPORT and INSERT works the same way from subscription point of view.
REPORT and INSERT type REST subscription request can contain content for multiple E2 subscriptions. POLICY subscription can also contain content for multiple
E2 subscriptions but using in that way may not be feasible. REPORT, POLICY and INSERT type subscriptions in the same REST request are not supported supported
in Subscription Manager.

REPORT and INSERT type subscription can contain content for multiple E2 subscriptions. If there is need to resend the same REST request, the request must
contain Subscription Manager allocated id for the REST request, which was returned to xApp when the request was sent first time. The request must also
contain the same content as first time. Reason for xApp to resend the same request could be timeout or some failure which is not permanent in E2 Node or
xApp restart. In retry cases Subscription Manager retries the E2 subscriptions which does not exist in its records. For others Subscription Manager
returns successful REST notification without sending any messages to E2 Node. One REST Subscription request can contain E2 subscriptions to only one E2 Node.

If there is need to change REPORT or INSERT type subscription then previously made subscription need to be deleted first. If there are any REPORT or INSERT
type E2 subscription which need to change frequently, it is not good idea to bundle them with other REPORT or INSERT type E2 subscriptions in the same REST
subscription request.

POLICY type subscription can contain content for multiple E2 subscriptions but it may not be feasible as POLICY subscription may change. Idea in POLICY
subscription is that xApp can send changed contend to E2 Node without making new subscription but just send update. In such case it is not good idea to bundle
the POLICY type E2 subscription with any other POLICY type E2 subscriptions in the same REST subscription request.

In xApp restart case only mandatory thing what is required xApp to be able preserve is the Subscription Manager allocated REST requests ids. That is if xApp
can generate the equal requests otherwise as were done first time before restart. xApp can resent the same REST requests to Subscription Manager as first time
before restart. REST request id must be placed in the REST request. That is the only way for Subscription Manager to identify already made subscriptions in it
records and work as expected, i.e.  not run into problems and return successful REST notifications to xApp without sending any messages to E2 Node.

Architecture
------------

  * Message routing

      Subscribed messages from E2 Node are transported to RIC inside RIC Indication message. RIC Indication message is transported to xApp
      inside RMR message in Payload field of the RMR message. RMR message is routed to xApp based on SubId field (E2 instance id) in
      the RMR header. 

      Subscription Manager allocates unique E2 instance id for every E2 subscription during subscription procedure. Subscription Manager
      puts allocated E2 instance id to InstanceId field in the ASN.1 packed RIC Subscription Request message which is sent to E2 Node. That
      E2 instance id is then used for the E2 subscription in RIC and E2 Node as long the E2 subscription lives. xApp gets the
      allocated E2 instance id in REST notification message when E2 subscription procedure is completed.
      
      Subscribed messages are routed to xApps based on InstanceId in E2 Indication message. InstanceId is placed in the SubId field of the RMR
      message header when E2 Termination sends the subscribed message to xApp.

      RIC Subscription Request and RIC Subscription delete Request messages are pre configured to be routed to E2 Termination and responses
      to those messages back to Subscription Manager.

      Subscription Manager allocates RIC Requestor Id for E2 interface communication. Currently the id value is always 123. E2 Node gets the Request
      of the xApp who makes the first subscription. E2 Node uses Subscription Manager allocated RIC Requestor ID in all RIC Indication messages it sends
      to RIC for the subscription. In merge case subscription in E2 Node is created only for the first requestor.

  * Subscription procedure
      
    * Successful case

      xApp sends REST Subscription Request message to Subscription Manager. The request can contain multiple E2 subscriptions. It contains also
      xApp generated xApp instance id for every E2 subscription. Subscription Manager checks does the message contain Subscription Manager allocated
      REST request id for the request. When xApp sends the request first time there is no REST request id and Subscription Manager allocates it.

      Then Subscription Manager makes simple validation for data in the request and copies data to Golang data types. When all data is copied successfully
      Subscription Manager sends successful respond to the REST request. Response contains Subscription Manager allocated REST request id.
      Then Subscription Manager sends route create request to Routing Manager over REST interface. When route is created successfully, Subscription Manager
      ASN.1 encodes the E2 messages and forwards those to E2 Termination. When RIC Subscription Response arrives from E2 Termination
      Subscription Manager forwards REST notification to xApp. The notification contains REST request id, xApp instance id and E2 instance id.
      
      Subscription Manager supervises route creation and RIC Subscription Request with a timer.

      RIC Indication messages which are used to transport subscribed messages from E2 Node are routed from E2 Termination to xApps directly using
      the routes created during Subscription procedure.

      Subscription Manager supports REPORT, POLICY and INSERT type subscriptions (E2 RICActionTypes). CONTROL is not supported. POLICY type
      subscription can be updated. In update case signaling sequence is the same as above, except route is not created to Routing manager.
      xApp uses initially allocated REST request id, xApp instance id in update case. Route in POLICY type subscription case is needed
      only that Error Indication could be to xApp, but it is not used currently. RIC Subscription Request message contains list of ActionsToBeSetup
      information elements. The list cannot have REPORT, POLICY or INSERT action types at the same time. Subscription Manager checks actions types
      in the message. If different action types is found the REST request is not accepted.


    .. image:: images/Successful_Subscription.png
      :width: 600
      :alt: Successful subscription picture


    * Failure case

      Failure can happen already before REST request reaches Subscription Manager. Swagger make value checks for the message passed to it.
      If values are does not accepted then send function returns "unknown error".

      If failure happens when Subscription Manager validates the REST request then error is returned instantly and processing of request is
      stopped. xApp receives bad request (HTTP response code 400) response.

      If failure response is received from E2 Node then REST notification is forwarded to xApp with appropriate error cause. The notification
      contains REST request id, xApp instance id and zero E2 instance id.

    .. image:: images/Subscription_Failure.png
      :width: 600
      :alt: Subscription failure picture

    * Timeout in Subscription Manager

      In case of timeout in Subscription Manager, Subscription Manager may resend the RIC Subscription Request to E2 Node. By default Subscription
      Manager retries twice. If there is no response after retries, Subscription Manager sends unsuccessful REST notification to xApp. The notification
      contains REST request id, xApp instance id and zero E2 instance id.

    * Timeout in xApp

      xApp can resend the same REST Subscription Request if request timeouts.

      xApp may resend the same request if it does not receive expected notification in expected time. If xApp resends the same request while Subscription
      Manager is still processing previous request then Subscription Manager responds accepts the request and continues processing previous request.

    .. image:: images/Subscription_Timeout.png
      :width: 600
      :alt: Subscription timeout picture

  * Subscription delete procedure

    * Successful case

      xApp sends REST Subscription Delete Request message to Subscription Manager. xApp must use the same REST request id which it received in REST Subscription
      Response. REST delete request will delete all successfully subscribed E2 subscriptions which was subscribed earlier when the REST request id was created.
      When Subscription Manager receives REST Subscription Delete Request it check has it such REST subscription. If it has then Subscription Manager sends successful
      response to xApp and starts sending E2 delete requests to E2 Termination one by one. When RIC Subscription Delete Response arrives from E2 Termination to
      Subscription Manager, Subscription Manager request route deletion from Routing Manager. xApp does not get any notification about deleted E2 subscriptions. 
      
      Subscription Manager supervises RIC Subscription Deletion Request and route delete with a timer.

    .. image:: images/Successful_Subscription_Delete.png
      :width: 600
      :alt: Successful subscription delete picture

    * Failure case

      Delete procedure cannot fail from xApp point of view. Subscription Manager always responds with successful REST Subscription Response to xApp.
      E2 Node could respond with delete failure in case the subscription which Subscription Manager wants to delete does not exist. In this case delete procedure
      ends there.

    .. image:: images/Subscription_Delete_Failure.png
      :width: 600
      :alt: Subscription delete failure picture

    * Timeout in Subscription Manager

      In case of timeout in Subscription Manager, Subscription Manager may resend the RIC Subscription Delete Request to E2 Node. By default Subscription Manager
      retries twice. If there is no response after retry, Subscription Manager stops trying.

    * Timeout in xApp

      xApp can resend the same REST Subscription Delete Request if request timeouts.

    .. image:: images/Subscription_Delete_Timeout.png
      :width: 600
      :alt: Subscription delete timeout picture

    * Unknown REST request id

      If Subscription Manager receives RIC Subscription Delete Request for a REST request id which does not exist, Subscription Manager sends
      successful REST response to xApp.

  * Subscription merge procedure

    * Successful case

      Merge is possible only for REPORT type subscription. It is possible only when Action Type and Event Trigger Definition of subscriptions are equal.

      xApp sends REST Subscription Request message to Subscription Manager. The request can contain multiple E2 subscriptions as in normal Subscription
      procedure but some of the E2 subscriptions in the list are already subscribed from E2 Node. For those which are not yet subscribed Subscription Manager
      applies normal Subscription procedure. E2 subscriptions in the list which are already subscribed are just assigned to existing subscriptions and Subscription
      Manager just sends route create to Routing Manager and then forwards successful REST notification to xApp for the E2 subscriptions. The notification
      contains REST request id, xApp instance id and E2 instance id.

      One thing to note! REST Subscription request and returned REST notification goes through different TCP ports. For that reason there is no guarantee that
      response for REST Subscription request arrives to xApp before first REST notification. That is possible mostly in merge case where subscription already exist
      in Subscription Manager records. Successful REST notification is returned to xApp without making subscription from E2 Node which would cause some delay before
      REST notification can be sent.
      
      Route create is supervised with a timer.

      ``Only REPORT type subscriptions can be be merged.``

    .. image:: images/Successful_Subscription_Merge.png
      :width: 600
      :alt: Successful subscription merge picture

    * Failure case

      Failure can happen already before REST request reaches Subscription Manager. Swagger make value checks for the message passed to it.
      If values are does not accept then send function returns "unknown error".

      If failure happens when Subscription Manager validates the REST request then error is returned instantly and processing of request is
      stopped. xApp receives bad request (HTTP response code 400) response.
      
      If error happens during route create then Subscription Manager forwards REST notification toxApp with appropriate error cause. The notification contains
      also REST request id, xApp instance id and zero E2 instance id.

    * Timeout in Subscription Manager

      Timeout can come only in route create during merge operation. If error happens during route create then Subscription Manager forwards REST
      notification toxApp with appropriate error cause. The notification contains also REST request id, xApp instance id and zero E2 instance id.

    * Timeout in xApp

      xApp can resend the same REST Subscription Request if request timeouts.

  * Subscription delete merge procedure

    * Successful case

      xApp sends REST Subscription Delete Request message to Subscription Manager. If delete concerns merged subscription, Subscription Manager
      responds with REST Subscription Delete Response to xApp and then sends route delete request to Routing manager.
      
      Subscription Manager supervises route delete with a timer.

    .. image:: images/Successful_Subscription_Delete_Merge.png
      :width: 600
      :alt: Successful subscription delete merge picture

    * Failure case

      Delete procedure cannot fail from xApp point of view. Subscription Manager always responds with successful REST Subscription Delete Response to xApp.

    * Timeout in Subscription Manager

      Timeout can only happen in route delete to Routing manager. Subscription Manager always responds with successful REST Subscription Delete Response to xApp.

    * Timeout in xApp

      xApp can resend the same REST Delete Request if request timeouts.

  * xApp restart

    When xApp is restarted for any reason it may resend REST subscription requests for subscriptions which have already been subscribed. If REPORT or INSERT type
    subscription already exists and RMR endpoint of requesting xApp is attached to subscription then successful response is sent to xApp directly without
    updating Routing Manager and E2 Node. If POLICY type subscription already exists, request is forwarded to E2 Node and successful response is sent to xApp.
    E2 Node is expected to accept duplicate POLICY type requests. In restart IP address of the xApp may change but domain service address name does not.
    RMR message routing uses domain service address name.

  * Subscription Manager restart

    Subscription Manager stores REST request ids, E2 subscriptions and their mapping to REST request ids in db (SDL). In start up Subscription Manager restores REST request
    ids, E2 subscriptions and their mapping from db. For E2 subscriptions which were not successfully completed, Subscription Manager sends delete request to E2 Node and
    removes routes created for those. In restart case xApp may need to resend the same REST request to get all E2 subscriptions completed.
    
    Restoring subscriptions from db can be disable via submgr-config.yaml file by setting "readSubsFromDb": "false".

Metrics
-------
 Subscription Manager adds following statistic counters:

 Subscription create counters:
		- SubReqFromXapp: The total number of SubscriptionRequest messages received from xApp
		- SubRespToXapp: The total number of SubscriptionResponse messages sent to xApp
		- SubFailToXapp: The total number of SubscriptionFailure messages sent to xApp
		- SubReqToE2: The total number of SubscriptionRequest messages sent to E2Term
		- SubReReqToE2: The total number of SubscriptionRequest messages resent to E2Term
		- SubRespFromE2: The total number of SubscriptionResponse messages from E2Term
		- SubFailFromE2: The total number of SubscriptionFailure messages from E2Term
		- SubReqTimerExpiry: The total number of SubscriptionRequest timer expires
		- RouteCreateFail: The total number of subscription route create failure
		- RouteCreateUpdateFail: The total number of subscription route create update failure
		- MergedSubscriptions: The total number of merged Subscriptions

 Subscription delete counters:
		- SubDelReqFromXapp: The total number of SubscriptionDeleteResponse messages received from xApp
		- SubDelRespToXapp: The total number of SubscriptionDeleteResponse messages sent to xApp
		- SubDelReqToE2: The total number of SubscriptionDeleteRequest messages sent to E2Term
		- SubDelReReqToE2: The total number of SubscriptionDeleteRequest messages resent to E2Term
		- SubDelRespFromE2: The total number of SubscriptionDeleteResponse messages from E2Term
		- SubDelFailFromE2: The total number of SubscriptionDeleteFailure messages from E2Term
		- SubDelReqTimerExpiry: The total number of SubscriptionDeleteRequest timer expires
		- RouteDeleteFail: The total number of subscription route delete failure
		- RouteDeleteUpdateFail: The total number of subscription route delete update failure
		- UnmergedSubscriptions: The total number of unmerged Subscriptions

 SDL failure counters:
		- SDLWriteFailure: The total number of SDL write failures
		- SDLReadFailure: The total number of SDL read failures
		- SDLRemoveFailure: The total number of SDL read failures

Configurable parameters
-----------------------
 Subscription Manager has following configurable parameters.
   - Retry timeout for RIC Subscription Request message
      - e2tSubReqTimeout_ms: 2000 is the default value

   - Retry timeout for RIC Subscription Delete Request message
      - e2tSubDelReqTime_ms: 2000 is the default value

   - Waiting time for RIC Subscription Response and RIC Subscription Delete Response messages
      - e2tRecvMsgTimeout_ms: 2000 is the default value

   - Try count for RIC Subscription Request message   
      - e2tMaxSubReqTryCount: 2 is the default value

   - Try count for RIC Subscription Delete Request message   
      - e2tMaxSubDelReqTryCount: 2 is the default value
   
   - Are subscriptions read from database in Subscription Manager startup
      - readSubsFromDb: "true"  is the default value
 
 The parameters can be changed on the fly via Kubernetes Configmap. Default parameters values are defined in Helm chart

 Use following command to open Subscription Manager's Configmap in Nano editor. First change parameter and then store the
 change by pressing first Ctrl + o. Close editor by pressing the Ctrl + x. The change is visible in Subscription Manager's
 log after some 20 - 30 seconds.
 
 .. code-block:: none

  KUBE_EDITOR="nano" kubectl edit cm configmap-ricplt-submgr-submgrcfg -n ricplt

REST interface for debugging and testing
----------------------------------------
 Give following commands to get Subscription Manager pod's IP address

 .. code-block:: none

  kubectl get pods -A | grep submgr
  
  ricplt        submgr-75bccb84b6-n9vnt          1/1     Running             0          81m

  Syntax: kubectl exec -t -n ricplt <add-submgr-pod-name> -- cat /etc/hosts | grep submgr | awk '{print $1}'
  
  Example: kubectl exec -t -n ricplt submgr-75bccb84b6-n9vnt -- cat /etc/hosts | grep submgr | awk '{print $1}'

  10.244.0.181

 Get metrics

 .. code-block:: none

  Example: curl -s GET "http://10.244.0.181:8080/ric/v1/metrics"

 Get subscriptions

 .. code-block:: none

  Example: curl -X GET "http://10.244.0.181:8088/ric/v1/subscriptions"

 Delete single subscription from db

 .. code-block:: none

  Syntax: curl -X POST "http://10.244.0.181:8080/ric/v1/test/deletesubid={SubscriptionId}"
  
  Example: curl -X POST "http://10.244.0.181:8080/ric/v1/test/deletesubid=1"

 Remove all subscriptions from db

 .. code-block:: none

  Example: curl -X POST "http://10.244.0.181:8080/ric/v1/test/emptydb"

 Make Subscription Manager restart

 .. code-block:: none

  Example: curl -X POST "http://10.244.0.181:8080/ric/v1/test/restart"

 Use this command to get Subscription Manager's log writings

 .. code-block:: none

   Example: kubectl logs -n ricplt submgr-75bccb84b6-n9vnt

 Logger level in configmap.yaml file in Helm chart is by default 2. It means that only info logs are printed.
 To see debug log writings it has to be changed to 4.

 .. code-block:: none

    "logger":
      "level": 4

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

      - REPORT
      - POLICY
      - INSERT

    * RIC Subscription Delete procedure

    * Merge and delete of equal REPORT type subscriptions.

Recommendations for xApps
-------------------------

   * Recommended retry delay in xApp

     Subscription Manager makes two retries for E2 subscriptions and E2 subscription deletions. xApp should not retry before it has received REST notification for
     all E2 subscriptions sent in REST subscription request. Maximum time to complete all E2 subscriptions in Subscription Manager can be calculated like this:
     t >= 3 * 2s * count_of_subscriptions in the REST request. Length of supervising timers in Subscription Manager for the requests it sends to E2 Node is by
     default 2 seconds. There can be only one ongoing E2 subscription request towards per E2 Node other requests are queued in Subscription Manager.
