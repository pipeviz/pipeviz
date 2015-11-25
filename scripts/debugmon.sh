#!/bin/sh

# requires expvarmon, go get github.com/divan/expvarmon
# pipeviz must be compiled with -tags debug (`go install -tags debug` from root of repo)
expvarmon -ports=6060 -vars="mem:memstats.Alloc,mem:memstats.Sys,mem:memstats.HeapAlloc,mem:memstats.HeapInuse,duration:memstats.PauseNs,duration:memstats.PauseTotalNs,memstats.NumGC,Goroutines,duration:Uptime,MsgID,WebsockClients,MainBrokerSubs" -i="1s"
