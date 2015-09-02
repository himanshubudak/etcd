package main

import (
    "strconv"
    "os"
    "fmt"
    "github.com/coreos/go-etcd/etcd"
    "math/rand"
    "time"
)

type actions func(int,int)
var operation string
var keycount int
var operation_count int
var threads int
var client *etcd.Client
var pct [5]int
var pct_count [5]int
var value_range [6]int

func main() {
    etcdhost := os.Args[1]
    etcdport := os.Args[2]

    operation = os.Args[3]
    keycount = int(toInt(os.Args[4],10,64))
    operation_count = int(toInt(os.Args[5],10,64))
    //log_file := os.Args[6]
 
    //c := etcd.NewClient(["http://127.0.0.1:2379",]) // default binds to http://0.0.0.0:4001
    var machines = []string{"http://"+etcdhost+":"+etcdport}
 //fmt.Println(machines)
    client =  etcd.NewClient(machines)
 //Pre-defined keysize distribution in percentage
    pct_temp := [5]int{5, 74, 10, 10, 1}
    pct = pct_temp
    value_range_temp := [6]int{0, 256, 512, 1024, 8192, 204800}
    value_range = value_range_temp

    var pct_count [5] int
    for i:=0;i<5;i++{
        pct_count[i] = pct[i] * keycount / 100
    }
    threads = 5

    switch{
    case operation == "create":
        fmt.Println("Inside create")
        var values [2]int
        base := 0
        for i:=0;i<len(pct);i++{
            values[0] = value_range[i]
            values[1] = value_range[i+1]
            go create_keys(base,pct_count[i],values)
            time.Sleep(100 * time.Millisecond)
        }
    case operation == "get":
        fmt.Println("Inside get")
        handler(get_values)
    case operation == "update":
        fmt.Println("Inside update")
        handler(update_values)
    case operation == "delete":
        fmt.Println("Inside delete")
        handler(delete_values)
    }
    // SET the value "bar" to the key "foo" with zero TTL
    // returns a: *store.Response
    
    /*
    res, _ := client.Set("/frontends/fe6", "10.0.0.6", 0)
    fmt.Printf("set response: %+v\n", res)
    value_range_input := [2]int{8,16}
    create_keys(0,10,value_range_input)
    get_values(1,5)
    */


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

func toInt(s string, base int, bitSize int) int64 {
    i, err := strconv.ParseInt(s, base, bitSize)
    if err != nil {
        panic(err)
    }
    return i
}

func toString(i int) string {
    s := strconv.Itoa(i)
    return s
}

func get_values(base int, per_thread int){
    var key int
    limit := base + (keycount / threads) - 1
    for i:=0;i<per_thread;i++{
        rand.Seed(time.Now().Unix())
        key = rand.Intn(limit-base) + base
        result, _ := client.Get(toString(key),false,false) 
        fmt.Println(result)
    } 
}

func create_keys(base int, count int, r [2]int){
    fmt.Println("function create called")
    var key int
    for i:=0;i<count;i++{
        key = base + i
        rand.Seed(time.Now().Unix())
        r1 := rand.Intn(r[1]-r[0]) + r[0]
        value := RandStringBytesRmndr(r1)
        fmt.Println(value)
        result, _ := client.Set(toString(key),value,1000)
        fmt.Println(result)
    }
}

func update_values(base int, per_thread int){
    var key int
    val := "UpdatedValue"
    limit := base + (keycount / threads) - 1
    for i:=0;i<per_thread;i++{
        rand.Seed(time.Now().Unix())
        key = rand.Intn(limit-base) + base
        result, _ := client.Set(toString(key),val,1000) 
        fmt.Println(result)
    }

}

func delete_values(base int, per_thread int){
    var key int
    limit := base + (keycount / threads) - 1
    for i:=0;i<per_thread;i++{
        rand.Seed(time.Now().Unix())
        key = rand.Intn(limit-base) + base
        result, _ := client.Delete(toString(key),false)
        fmt.Println(result)
    }
}


func handler(fn actions){
    per_thread := operation_count/threads
    base := 0
    for i:=0;i<threads;i++{
        go fn(base,per_thread)
        time.Sleep(100 * time.Millisecond)
    }
}

func RandStringBytesRmndr(n int) string {
    const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
    }
    return string(b)
}
/*
func get_update_delete_jobs()

def get_update_delete_jobs(func):
    per_thread = operation_count / threads
    jobs = []
    base = 0
    for i in range(threads):
        p = multiprocessing.Process(target=func, args=(base, per_thread, ))
        jobs.append(p)
        base = base + (keycount / threads)
        p.start()

*/
