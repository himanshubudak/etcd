package main

import (
    "strconv"
    "strings"
    "os"
    "fmt"
    "github.com/coreos/go-etcd/etcd"
    "crypto/rand"
    "time"
    "sync"
    "os/exec"
    //"encoding/binary"
    "math/big"
    "log"
    //"io/ioutil"
)

/*
Declarations :::: 
    actions : for passing otherfunctions as arguments to handler function
    operation : which operation to perform
    keycount : number of keys to be added, retrieved, deleted or updated
    threads : number of total threads
    pct : each entry represents percentage of values lying in (value_range[i],value_range[i+1]) 
*/

type actions func(int,int)

var n int32
var wg sync.WaitGroup
var operation string
var keycount int
var operation_count int
var threads int
var client *etcd.Client
var pct [5]int
var pct_count [5]int
var value_range [6]int
var pidetcd string
var start time.Time
var f *os.File
var err error
/////////////
// FOR TESTING ONLY
var mycount int
/////////////
func main() {
    etcdhost := os.Args[1]
    etcdport := os.Args[2]
    operation = os.Args[3]
    keycount = int(toInt(os.Args[4],10,64))
    operation_count = int(toInt(os.Args[5],10,64))
    log_file := os.Args[6]
    
    pidtemp, _ := exec.Command("pidof","etcd").Output()
    pidetcd = string(pidtemp)
    pidetcd = strings.TrimSpace(pidetcd)
    etcdmem_s, _  := strconv.Atoi(getMemUse(pidetcd))
    fmt.Println("This is the current memory usage by etcd before execution: " + getMemUse(pidetcd))


    //c := etcd.NewClient(["http://127.0.0.1:2379",]) // default binds to http://0.0.0.0:4001
    
    ///////

    mycount = 0

    ////////
    var machines = []string{"http://"+etcdhost+":"+etcdport}
 //fmt.Println(machines)
    client =  etcd.NewClient(machines)
 //Pre-defined keysize distribution in percentage
    
    //Log File
    f, err = os.OpenFile(log_file, os.O_RDWR | os.O_CREATE , 0666)
    if err != nil {
        log.Fatalf("error opening file: %v", err)
    }
    
    log.SetOutput(f)
    log.Println("Starting #####")
    log.Println("Keycount = %s , operation_count = %s",os.Args[4],os.Args[5])


    // Percentage distribution of key-values
    pct_temp := [5]int{5, 74, 10, 10, 1}
    pct = pct_temp
    value_range_temp := [6]int{0, 256, 512, 1024, 8192, 204800}
    value_range = value_range_temp

    var pct_count [5] int
    for i:=0;i<5;i++{
        pct_count[i] = pct[i] * keycount / 100
    }

    // Number of threads
    threads = 5

    // Keep track of the goroutines
    wg.Add(len(pct))
    
    switch{
    case operation == "create":
        log.Println("Operation : create")
        var values [2]int
        base := 0
        for i:=0;i<len(pct);i++{
            values[0] = value_range[i]
            values[1] = value_range[i+1]
            go create_keys(base,pct_count[i],values)
        }
        wg.Wait()
    case operation == "get":
        log.Println("Operation : get")
        //fmt.Println("Inside get")
        handler(get_values)
        wg.Wait()
    case operation == "update":
        log.Println("Operation : update")
        //fmt.Println("Inside update")
        handler(update_values)
        wg.Wait()
    case operation == "delete":
        log.Println("Operation : delete")
        //fmt.Println("Inside delete")
        handler(delete_values)
        wg.Wait()
    }
    etcdmem_e, _ := strconv.Atoi(getMemUse(pidetcd))
    fmt.Println("This is the current memory usage by etcd after execution: " + getMemUse(pidetcd))
    fmt.Println("Difference := " + strconv.Itoa(etcdmem_e - etcdmem_s))
    fmt.Println(mycount)
    log.Println("Difference in memory use, after and before load testing := " + strconv.Itoa(etcdmem_e - etcdmem_s))
    defer f.Close()
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
    //fmt.Println("per_thread,base = ",strconv.Itoa(per_thread)," ", strconv.Itoa(base))
    var key int
    limit := base + (keycount / threads)
    fmt.Println("limit is = ",strconv.Itoa(limit))
    for i:=0;i<per_thread;i++{
	    mycount++
        //fmt.Println(i)
        //fmt.Println("per_threadmycount,base inside get : ",strconv.Itoa(per_thread)," " , strconv.Itoa(mycount)," ",strconv.Itoa(base))
        m,_ := rand.Int(rand.Reader,big.NewInt(int64(limit-base)))
        key = int(m.Int64()) + base
        //fmt.Println(key)
        start = time.Now()
        result, _ := client.Get(toString(key),false,false)
        elapsed := time.Since(start)
        log.Println("key %s took %s", key, elapsed)
        //defer timeTrack(time.Now(), strconv.Itoa(key))
        fmt.Println(result)
    } 
    //time.Sleep(10000 * time.Millisecond)
    defer wg.Done()
}

func create_keys(base int, count int, r [2]int){
    //fmt.Println("function create called")
    var key int
    for i:=0;i<count;i++{
        mycount++
        //fmt.Println(i)
        key = base + i
        m,_ := rand.Int(rand.Reader,big.NewInt(int64(r[1]-r[0])))
        r1 := int(m.Int64()) + r[0]
        value := RandStringBytesRmndr(r1)
        //fmt.Println(key)
        result, _ := client.Set(toString(key),value,1000)
        _ = result
    }
    defer wg.Done()
}

func update_values(base int, per_thread int){
    var key int
    val := "UpdatedValue"
    limit := base + (keycount / threads)
    for i:=0;i<per_thread;i++{
        m,_ := rand.Int(rand.Reader,big.NewInt(int64(limit-base)))
        key = int(m.Int64()) + base
        result, _ := client.Set(toString(key),val,1000) 
        fmt.Println(result)
    }
    defer wg.Done()
}

func delete_values(base int, per_thread int){
    var key int
    limit := base + (keycount / threads)
    for i:=0;i<per_thread;i++{
        m,_ := rand.Int(rand.Reader,big.NewInt(int64(limit-base)))
        key = int(m.Int64()) + base
        result, _ := client.Delete(toString(key),false)
        fmt.Println(result)
    }
    defer wg.Done()
}


func handler(fn actions){
    per_thread := operation_count/threads
    base := 0
    for i:=0;i<threads;i++{
        go fn(base,per_thread)
        base = base + (keycount/threads)
        //time.Sleep(100 * time.Millisecond)
    }
}

func RandStringBytesRmndr(n int) string {
    const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    b := make([]byte, n)
    for i := range b {
        m,_ := rand.Int(rand.Reader,big.NewInt(int64(len(letterBytes))))
        b[i] = letterBytes[int(m.Int64())]
    }
    return string(b)
}

func getMemUse(pid string) string{
    mem, _ := exec.Command("pmap","-x",pid).Output()
    mempmap := string(mem)
    temp := strings.Split(mempmap,"\n")
    temp2 := temp[len(temp)-2]
    //fmt.Println(temp2)
    temp3 := strings.Fields(temp2)
    memory := temp3[3]
    return memory
    /*
    memuse := exec.Command("pmap","-x",pid).Output()

    pipe1cmd := exec.Command("tail","-n1")
    pipe1cmd.Stdin, _ = memuse.StdoutPipe()
    pipe1cmd.Stdout = os.Stdout

    _ = pipe1cmd.Start()
    _ = memuse.Run()
    _ = pipe1cmd.Wait()

    b, _ := ioutil.ReadAll(os.Stdout)
    memarray := strings.Fields(string(b))
    return memarray[3]
    */
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Println("key %s took %s", name, elapsed)
}

