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
    "math/big"
    "log"
    "code.google.com/p/gcfg"
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
var pct []int
var pct_count []int
var value_range []int
var pidetcd string
var start time.Time
var f *os.File
var err error


func main() {

    // Configuration Structure
    cfg := struct {
        Section_Args struct {
            Etcdhost string
            Etcdport string
            Operation string
            Keycount string
            Operation_Count string
            Log_File string
        }
        Section_Params struct {
            Threads int
            Pct string
            Value_Range string
        }
    }{}
    
    // Config File
    filename := os.Args[1]

    err := gcfg.ReadFileInto(&cfg, filename)
    if err != nil {
        log.Fatalf("Failed to parse gcfg data: %s", err)
    }
    
    // Reading from config file
    ////////////////////////////////////////////////////////
    etcdhost := cfg.Section_Args.Etcdhost
    etcdport := cfg.Section_Args.Etcdport
    operation = cfg.Section_Args.Operation
    keycount = int(toInt(cfg.Section_Args.Keycount,10,64))
    operation_count = int(toInt(cfg.Section_Args.Operation_Count,10,64))
    log_file := cfg.Section_Args.Log_File

    // The Distribution of keys
    percents := cfg.Section_Params.Pct
    temp := strings.Split(percents,",")
    pct = make([]int,len(temp))
    pct_count = make([]int,len(temp))
    for i:=0;i<len(temp);i++ {
        pct[i] = int(toInt(temp[i],10,64))
        pct_count[i] = pct[i] * keycount / 100
    }
    // Percentage distribution of key-values
    value_r := cfg.Section_Params.Value_Range
    temp = strings.Split(value_r,",")
    value_range = make([]int,len(temp))
    for i:=0;i<len(temp);i++ {
        value_range[i] = int(toInt(temp[i],10,64))
    }
    //Maximum threads
    threads = cfg.Section_Params.Threads
    fmt.Println("!!!!@@@@@@@@@@#######",threads)
    ///////////////////////////////////////////////////////////

    // Getting Memory Info for etcd instance
    pidtemp, _ := exec.Command("pidof","etcd").Output()
    pidetcd = string(pidtemp)
    pidetcd = strings.TrimSpace(pidetcd)
    etcdmem_s, _  := strconv.Atoi(getMemUse(pidetcd))
    fmt.Println("This is the current memory usage by etcd before execution: " + getMemUse(pidetcd))

    // Creating a new client for handling requests
    var machines = []string{"http://"+etcdhost+":"+etcdport}
    client =  etcd.NewClient(machines)
    
    //Log File
    f, err = os.OpenFile(log_file, os.O_RDWR | os.O_CREATE , 0666)
    if err != nil {
        log.Fatalf("error opening file: %v", err)
    }
    // Log file set
    log.SetOutput(f)
    log.Println("Starting #####")
    log.Println("Keycount = %s , operation_count = %s",cfg.Section_Args.Keycount,cfg.Section_Args.Operation_Count)

    

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
        handler(get_values)
        wg.Wait()
    case operation == "update":
        log.Println("Operation : update")
        handler(update_values)
        wg.Wait()
    case operation == "delete":
        log.Println("Operation : delete")
        handler(delete_values)
        wg.Wait()
    }
    etcdmem_e, _ := strconv.Atoi(getMemUse(pidetcd))
    fmt.Println("This is the current memory usage by etcd after execution: " + getMemUse(pidetcd))
    fmt.Println("Difference := " + strconv.Itoa(etcdmem_e - etcdmem_s))
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
    var key int
    limit := base + (keycount / threads)
    fmt.Println("limit is = ",strconv.Itoa(limit))
    for i:=0;i<per_thread;i++{
        m,_ := rand.Int(rand.Reader,big.NewInt(int64(limit-base)))
        key = int(m.Int64()) + base
        start = time.Now()
        result, _ := client.Get(toString(key),false,false)
        elapsed := time.Since(start)
        log.Println("key %s took %s", key, elapsed)
        fmt.Println(result)
    } 
    defer wg.Done()
}

func create_keys(base int, count int, r [2]int){
    var key int
    for i:=0;i<count;i++{
        key = base + i
        m,_ := rand.Int(rand.Reader,big.NewInt(int64(r[1]-r[0])))
        r1 := int(m.Int64()) + r[0]
        value := RandStringBytesRmndr(r1)
        start = time.Now()
        result, _ := client.Set(toString(key),value,1000)
        elapsed := time.Since(start)
        log.Println("key %s took %s", key, elapsed)
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
        start = time.Now()
        result, _ := client.Set(toString(key),val,1000) 
        elapsed := time.Since(start)
        log.Println("key %s took %s", key, elapsed)
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
        start = time.Now()
        result, _ := client.Delete(toString(key),false)
        elapsed := time.Since(start)
        log.Println("key %s took %s", key, elapsed)
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
    temp3 := strings.Fields(temp2)
    memory := temp3[3]
    return memory
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Println("key %s took %s", name, elapsed)
}

