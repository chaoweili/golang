package main

import(
"fmt"
"io/ioutil"
"net"
"net/url"
"os"
"strings"

"runtime"
"time"

"flag"  //console flag
)
//check if url contains http or https
func ParseUrl(s string) int{
    if strings.Contains(s,"http://"){
        return 1
    }else if strings.Contains(s,"https://")   {
        return 2
    }else {
        return -1
    } 
}
//help of this tool
func usage( ss string){
        fmt.Printf("usage:%s is a benchmarktool like ab bench on apache\n -c concurrency, will open X go routine to run this test\n -n total request\n -u URL which will be tested\n", ss)
        os.Exit(1)
}

var(
    concurrency = flag.Int("c",1,"concurrency")
    total_num = flag.Int("n",1,"total request num")
    dest_url = flag.String("u","","destination url")
)

func main(){
    flag.Parse()

    u,err := url.Parse(*dest_url)
    checkError(err)
    if 0 > ParseUrl(*dest_url){
        usage(os.Args[0])
    }
    req_uri := fmt.Sprintf("GET /%s HTTP/1.0\r\n\r\n",u.Path)

    //use channel to get result from child
    type ch chan int
    channels := make([]ch, *concurrency,*concurrency)

    start_time := time.Now().UnixNano() / 1000000
    times := *total_num / *concurrency
    for i:=0; i< *concurrency;i++{
        channels[i] = make(chan int)
        go  start_one_fork(u.Host,req_uri, times, channels[i])
    }
    
    //as a father fork, wait a long time to give child go func have enough time to finish
    runtime.Gosched()
    iSucc := 0
    iRet := 0
    for i:=0; i< *concurrency;i++{
        iRet = <- channels[i]
        iSucc += iRet
    }
    end_time := time.Now().UnixNano() / 1000000
    speed := int64 (*total_num)
    if end_time - start_time > 0 {
        speed = speed *1000 / (end_time - start_time) 
    }
    fmt.Println("total use time:", end_time - start_time,"mili second")
    fmt.Printf("total request:%d, concurrency:%d\n",*total_num,*concurrency)
    fmt.Printf("result:\nsucc:%d, fail:%d\n",iSucc, *total_num - iSucc)
    fmt.Printf("speed: %d RPS\n",speed )
}
func checkError(err error){
    if err != nil{
        fmt.Fprintf(os.Stderr, "Fatal error:%s\n", err.Error())
    }    
}
//each child will continue visit url and record result
func start_one_fork(dest_addr,req_uri string, times int, channel chan int){

    var iSucc,iFail,iResult = 0,0,0
    
    //short conn, close conn each request
    for i :=0; i < times;i++{
        iResult = reqOnetime(dest_addr,req_uri)
        if iResult == 0{
            iSucc++
        }else{
            iFail++
            fmt.Printf("%d time's result is %d\n", i, iResult)
        }
    }
    channel <- iSucc
}
//a http request and response 
func reqOnetime(dest_addr,req_uri string) int{
    tcpAddr,err:= net.ResolveTCPAddr("tcp4",dest_addr)
    checkError(err)
    if err != nil{
       return -5 
    }
    var result []byte

    //connect to dest 
    conn,err :=net.DialTCP("tcp",nil,tcpAddr)
    //defer conn.Close()    //can not be at this place or if DialTcp return conn is nil, conn.Close will get SIGSEGV

    checkError(err)
    if err != nil{
        return -1 
    }

    defer conn.Close()  //this is right, conn is not nil here

    //send http request to dest url
    _,err = conn.Write([]byte(req_uri))
    checkError(err)
    if err != nil{
        return -2 
    }

    //read http response
    result,err =ioutil.ReadAll(conn)
    checkError(err)
    if err != nil{
        return -3 
    }

    //parse http response, http 200 is succ, otherwise fail
    if 200 ==parse_http_response(result){
        return 0
    }else{
        return -4 
    }
    return 0
}

func parse_http_response( resp []byte) int{
    sResp := string(resp)
    if 0 == len(sResp){
        return -1
    }else if strings.HasPrefix(sResp, "HTTP/") && strings.Contains(sResp, " 200 OK"){// HTTP/1.1 200 OK
        return 200
    }else{
        fmt.Println(sResp)
        return -1
    }
}
