The current code runs only a simple test case to just add some keys and then get them
It has only been tested on localhost
to run simply type the command (after starting etcd on your machine)
go run etcd_load.go 127.0.0.1 port operation_type keycount operation_count logfile

here port = port where etcd is listening for client requests, on my machine it is 2379,4001
