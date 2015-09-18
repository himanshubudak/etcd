This is a etcd load test module written in go-language. It basically creates 
artificial load on a running etcd instance. The test runs for a random set of
keys and values, which can be specified in the configuration file. You can 
configure many other parameters in the config file. See the sample config file 
--"etcd_load.cfg" for more information

Before you proceed there are a few things that you might need to set up.

 - Make sure that the following packages are in your GOPATH. Set your 
   GOPATH if its not already set.
   In go, to get a "package" you can simply do : go get package. 
   In this use case do it under your GOPATH.

	Packages :
		"github.com/coreos/go-etcd/etcd"
		"code.google.com/p/gcfg"
		"code.google.com/p/go.crypto/ssh"


 - Set up a default config file, like the one available in the repo.
 	The details on how to configure it are in the sample -- "etcd_laod.cfg" 
 	itself

Now, to run etcd_load test, use the following steps
 - go build etcd_load.go report.go
 - ./etcd_load -c "default-config-file" --other-optional-flags

	Note that the "-c" flag is compulsory, that is you need to have a default 
	config file that must be input using the -c flag
	To know more about the flags :: do -- go run etcd_load.go -h

You can find runtime details in the log file. 
On the commandline the output looks like :
**********************************************************************
	Summary:
	  Total:	2.8454 secs.
	  Slowest:	0.1493 secs.
	  Fastest:	0.0001 secs.
	  Average:	0.0258 secs.
	  Requests/sec:	35.1449

	Response time histogram:
	  0.000 [1]		|
	  0.015 [17]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎
	  0.030 [48]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
	  0.045 [27]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
	  0.060 [3]		|∎∎
	  0.075 [2]		|∎
	  0.090 [0]		|
	  0.105 [1]		|
	  0.119 [0]		|
	  0.134 [0]		|
	  0.149 [1]		|

	Latency distribution:
	  10% in 0.0006 secs.
	  25% in 0.0226 secs.
	  50% in 0.0240 secs.
	  75% in 0.0308 secs.
	  90% in 0.0332 secs.
	  95% in 0.0481 secs.
	  99% in 0.1493 secs.

**********************************************************************
