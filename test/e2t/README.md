# E2T Test platform component
Base project: https://gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp
## General Description
__E2T__ is a platform component stub which has following behaviors
* Receives, decodes and prints out the content of RMR messages
* Sends Subscription Response message (12011) to each RMR message using it's sub_id
* Sends Subscription Response with invalid subscription ID in every 14 second
* Sends Subscription Response with malformed payload in every 14 second
* Set `E2T_RAWDATA` to override default encoded payload (use hex dump format)


