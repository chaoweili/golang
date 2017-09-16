# golang
useful golang tools and program

## abbench.go

### What's this:
this is a abbench written by Golang
If you do not know abbench, then 
this is a benchmark tool used to test performance of a http svr.

### How to use it:
./abbench -c 100000 -n 200000 -u http://10.10.10.10:80
<br>-c concurrency, will open X go routine to run this test
<br>-n total request
<br>-u URL which will be tested
