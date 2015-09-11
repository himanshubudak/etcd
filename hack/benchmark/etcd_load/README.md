The current code runs only a simple test case to just add some keys and then get them
It has only been tested on localhost
to run simply type the command (after starting etcd on your machine)
go run etcd_load.go etcd_load.cfg

etcd_load.cfg is the config file, use it to customize the test
For running it in remote mode (i.e remote-flag=True), it is being assumed
that the machine from which etcd_load is being called is among the known 
hosts of the remote machine, where etcd is running, and also that the former
can access the latter without the need of inputting password.
