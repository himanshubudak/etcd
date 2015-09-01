package main

import (
 "strconv"
 "os"
 "fmt"
 "github.com/coreos/go-etcd/etcd"
)

//var c etcd.NewClient
var operation int
var keycount int
var operation_count int
var threads int

func main() {
 threads = 5
 etcdhost := os.Args[1]
 etcdport := os.Args[2]

// operation,err = strconv.Atoi(os.Args[3])
// keycount = os.Args[4]
// operation_count = os.Args[5]
 //log_file := os.Args[6]
 
    //c := etcd.NewClient(["http://127.0.0.1:2379",]) // default binds to http://0.0.0.0:4001
 var machines = []string{"http://"+etcdhost+":"+etcdport}
 //fmt.Println(machines)
 var c =  etcd.NewClient(machines)
 //Pre-defined keysize distribution in percentage
 pct := [5] int {5, 74, 10, 10, 1}
 //key_range := [6]int{0, 256, 512, 1024, 8192, 204800}

 countn := 10
 for i:=0;i<countn;i++{
  myres, _ := c.Set(strconv.Itoa(i),strconv.Itoa(i),1000)
  fmt.Println(myres)
 }

 for i:=0;i<countn;i++{
  val, _ := c.Get(strconv.Itoa(i),false,false)
  fmt.Println(val)
 }
 
 //pct_count = [x * keycount/100 for x in pct]
 var pct_count [5] int
 for i:=0;i<5;i++{
  pct_count[i] = pct[i] * keycount / 100
 }

    // SET the value "bar" to the key "foo" with zero TTL
    // returns a: *store.Response
  res, _ := c.Set("/frontends/fe5", "10.0.0.5", 0)
  fmt.Printf("set response: %+v\n", res)

    // GET the value that is stored for the key "foo"
    // return a slice: []*store.Response
 //values, _ := c.Get("/frontends/",false,false)
  /*  for i, res := range values { // .. and print them out
        fmt.Printf("[%d] get response: %+v\n", i, res)
    }
*/
    // DELETE the key "foo"
    // returns a: *store.Response
    //res, _ = c.Delete("foo")
    //fmt.Printf("delete response: %+v\n", res)
}

/*
func get_values(base int, per_thread int){
 limit := base + (keycount/threads) - 1 
 var keys [20] int
 for i:=0;i<per_thread;i++{
  num := rand.Intn(limit-base) + base
  key_val := c.Get(strconv.Itoa(key),false,false) 
  fmt.Println(key_val)
 } 
}
*/
