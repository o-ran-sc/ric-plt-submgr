# RCO Test xApp Instance
Base project: https://gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp
## General Description
__RCO__ is an xApp instance which has following behaviors
* sends valid SUB_REQ RMR message periodically in every 2 second. 
* sends invalid message (10000) message in every 14 second
* sends subscription requsts (12010) with malformed payload in every 14 seconds
* receives RMR messages and emmits log on it's standard output
* default initial sequence number is `12345`. Set `RCO_SEED_SN` environment variable to override
* Set `RCO_RAWDATA` to override default encoded payload (use hex dump format)




