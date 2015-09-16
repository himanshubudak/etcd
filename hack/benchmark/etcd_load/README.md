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

Now, to run etcd_load test, use a command of type
 - go run etcd_load.go -c "default-config-file" --other-optional-flags

	Note that the "-c" flag is compulsory, that is you need to have a default 
	config file that must be input using the -c flag
	To know more about the flags :: do -- go run etcd_load.go -h

