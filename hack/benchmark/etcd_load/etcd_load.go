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
    "code.google.com/p/go.crypto/ssh"
    "io/ioutil"
    "bytes"
)

/*
Declarations :::: 
    actions   : for passing otherfunctions as arguments to handler function
    operation : which operation to perform
    keycount  : number of keys to be added, retrieved, deleted or updated
    threads   : number of total threads
    pct       : each entry represents percentage of values lying in (value_range[i],value_range[i+1]) 
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
var key ssh.Signer
var config *ssh.ClientConfig
var ssh_client *ssh.Client
var session *ssh.Session
var pidetcd_s string
var etcdmem_s int
var etcdmem_e int
///////////////////
/// TESTING
///////////////////
//var thread_arr []int

//////////////////

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
            Remote_Flag bool
            Remote_Host string
            Ssh_Port string
            Remote_Host_User string
        }
    }{}
    
    // Config File
    conf_file := os.Args[1]

    err := gcfg.ReadFileInto(&cfg, conf_file)
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

    remote_flag := cfg.Section_Params.Remote_Flag
    remote_host := cfg.Section_Params.Remote_Host
    ssh_port := cfg.Section_Params.Ssh_Port
    remote_host_user := cfg.Section_Params.Remote_Host_User
    if remote_flag {
        etcdhost = remote_host
        t_key, _ := getKeyFile()
        if  err !=nil {
            panic(err)
        }
        key = t_key
        config := &ssh.ClientConfig{
            User: remote_host_user,
            Auth: []ssh.AuthMethod{
            ssh.PublicKeys(key),
            },
        }

        t_client,_ := ssh.Dial("tcp", remote_host+":"+ssh_port, config)
        if err != nil {
            panic("Failed to dial: " + err.Error())
        }
        ssh_client = t_client

    }


    // Getting Memory Info for etcd instance
    if remote_flag {
        var bits bytes.Buffer
        mem_cmd := "pmap -x $(pidof etcd) | tail -n1 | awk '{print $4}'"
        session, err := ssh_client.NewSession()
        if err != nil {
            panic("Failed to create session: " + err.Error())
        }
        defer session.Close()
        session.Stdout = &bits
        if err := session.Run(mem_cmd); err != nil {
            panic("Failed to run: " + err.Error())
        }

        pidetcd_s = bits.String()
        pidetcd_s = strings.TrimSpace(pidetcd_s)
        etcdmem_i, _ := strconv.Atoi(pidetcd_s)
        etcdmem_s = etcdmem_i
    } else{
        pidtemp, _ := exec.Command("pidof","etcd").Output()
        pidetcd = string(pidtemp)
        pidetcd = strings.TrimSpace(pidetcd)
        pidetcd_s = getMemUse(pidetcd)
        etcdmem_i, _  := strconv.Atoi(pidetcd_s)
        etcdmem_s = etcdmem_i
    }
    fmt.Println("This is the current memory usage by etcd before execution: " + pidetcd_s)

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
    if remote_flag {
        var bits bytes.Buffer
        mem_cmd := "pmap -x $(pidof etcd) | tail -n1 | awk '{print $4}'"
        session, err := ssh_client.NewSession()
        if err != nil {
            panic("Failed to create session: " + err.Error())
        }
        defer session.Close()
        session.Stdout = &bits
        if err := session.Run(mem_cmd); err != nil {
            panic("Failed to run: " + err.Error())
        }
        pidetcd_s = bits.String()
        pidetcd_s = strings.TrimSpace(pidetcd_s)
        etcdmem_i,_ := strconv.Atoi(pidetcd_s)
        etcdmem_e = etcdmem_i
    } else {
        pidetcd_s = getMemUse(pidetcd)
        etcdmem_i, _ := strconv.Atoi(pidetcd_s)
        etcdmem_e = etcdmem_i
    }
    fmt.Println("This is the current memory usage by etcd after execution: " + pidetcd_s)
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

func getKeyFile() (key ssh.Signer, err error){
    //fmt.Println("getkey file funciton")
    file := os.Getenv("HOME") + "/.ssh/id_rsa"
    buf, err := ioutil.ReadFile(file)
    if err != nil {
        return
    }
    key, err = ssh.ParsePrivateKey(buf)
    if err != nil {
        return
     }
    return
}
