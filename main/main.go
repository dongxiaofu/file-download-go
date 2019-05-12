package main

import (
	"net"
	"log"
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"fmt"
	"encoding/json"
	"os"
	"io/ioutil"
	"flag"
	"time"
)

/**
 go run main.go -host=chugang.net -path=/ -filename=index.html -sleep=0
 go run main.go -host=i1.whymtj.com -path=/uploads/tu/201905/9999/f785f95cc1.jpg?t==3337 -filename=g9.jpg -sleep=0
go run main.go -host=dev.cg.com -path=/t.json -filename=t.json -sleep=0
395135
 */

func main() {

	//var newResponseContentBuf2 []byte
	//fmt.Println(len(newResponseContentBuf2))
	//return


	hostParam := flag.String("host", "", "host, 必填")
	pathParam := flag.String("path", "", "path，必填")
	filenameParam := flag.String("filename", "", "filename，必填")
	isSleepParam := flag.String("sleep", "0", "sleep,选题，默认为0")

	flag.Parse()

	if *hostParam == "" {
		fmt.Println("请输入host")
		return
	}

	if *pathParam == "" {
		fmt.Println("请输入path")
		return
	}

	if *filenameParam == "" {
		fmt.Println("请输入filename")
		return
	}

	var dbFile string = "tmp"

	var host string = *hostParam
	var path string = *pathParam
	var filename = *filenameParam
	var port string = "80"
	address := host + ":" + port

	// 读取下载文件的元数据
	var fileInfo2 fileMeta
	fileInfoString := readFile(dbFile)
	if fileInfoString == "" {
		fileInfo2 = fileMeta{
			Url:        address,
			FileOffset: 0,
			FileSize:0,
		}
	} else {
		var fileInfoJson = []byte(fileInfoString)

		err5 := json.Unmarshal(fileInfoJson, &fileInfo2)
		if err5 != nil {
			panic("error")
			return
		}
	}

	fileSize2Int64 := getFileSize(filename)
	fmt.Println("===========================fileSize2Int64=====start")
	fmt.Println(fileSize2Int64)
	fmt.Println("===========================fileSize2Int64=====end")
	fileSizeStr := strconv.FormatInt(fileSize2Int64, 10)
	fileSize2, _ := strconv.Atoi(fileSizeStr)

	//fmt.Println("===========================length2=====start")
	//fmt.Println(fileInfo2.FileSize)
	//fmt.Println("===========================length2=====end")

	fmt.Println("===========================fileSize2=====start")
	fmt.Println(fileSize2)
	fmt.Println("===========================fileSize2=====end")

	if fileSize2 > 0 && fileInfo2.FileSize == fileSize2 {
		os.Remove(dbFile)
		fmt.Println("下载完毕")
		return
	}



	// 发起 http 请求
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)

	checkError(err)

	tcpConn, err2 := net.DialTCP("tcp", nil, tcpAddr)

	tcpConn2, _ := net.DialTCP("tcp", nil, tcpAddr)

	var fileInfo fileMeta
	fileInfo = fileMeta{
		Url:        "1",
		FileOffset: 0,
		FileSize:0,
	}

	fileInfo.Url = address

	checkError(err2)

	var buffer bytes.Buffer

	requestHeaderGet := "GET " + path + " HTTP/1.1 \r\n"
	buffer.WriteString(requestHeaderGet)

	if fileInfo2.FileOffset != 0 {
		fmt.Println("yes")
		rangeHeader := "Range:bytes=" + strconv.Itoa(fileSize2) + "-\r\n"
		buffer.WriteString(rangeHeader)
	}

	buffer.WriteString("Cache-Control: no-cache \r\n")
	buffer.WriteString("Pragma: no-cache \r\n")
	buffer.WriteString("Connection: keep-alive \r\n")

	requestHeaderHost := "Host: " + host + "\r\n\r\n"
	buffer.WriteString(requestHeaderHost)
	requestHeader := buffer.String()

	//fmt.Println(requestHeader)
	//return

	_, err3 := tcpConn.Write([]byte(requestHeader))

	checkError(err3)


	// 获取头部信息
	_, err33 := tcpConn2.Write([]byte(requestHeader))
	checkError(err33)

	length2, headerLength := getResponseHeader(tcpConn2)

	fmt.Println("===========================length2=====start")
	fmt.Println(length2)
	fmt.Println("===========================length2=====end")

	var fileOffset int
	var fileSize int

	fileOffset = fileInfo2.FileOffset
	if fileInfo2.FileSize == 0 {
		fileInfo2.FileSize = length2
		fileSize = length2
	}else{
		//length2 = fileInfo2.FileSize
		fileSize = fileInfo2.FileSize
	}

	defer tcpConn.Close()

	// 2 是 \r\n\r\n 的长度。为何不是4
	length := length2 + headerLength + strings.Count("\r\n\r\n", "")

	fmt.Println("===========================length=====start")
	fmt.Println(length)
	fmt.Println("===========================length=====end")

	responseContentBuf := make([]byte, length)
	leng := 0
	fmt.Println("===========================leng=====start")
	fmt.Println(leng)
	fmt.Println("===========================leng=====end")

	var k int
	var newOffset int64 = 0
	for {

		var newResponseContentBuf []byte
		//if cap(responseContentBuf) == length {
		//	newResponseContentBuf = make([]byte, len(responseContentBuf)*2, cap(responseContentBuf) * 2)
		//	copy(newResponseContentBuf, responseContentBuf)
		//}

		length++

		fmt.Println("===========================k=====start")
		fmt.Println(k)
		fmt.Println("===========================k=====end")

		var n int
		if len(newResponseContentBuf) > 0 {
			n, _ = tcpConn.Read(newResponseContentBuf[leng:])
		}else{
			// 每次读取的数据，都包括 http header，为何不断点下载的时候，为何上文分配的
			// length := length2 + headerLength + strings.Count("\r\n\r\n", "")
			// 空间够用？应该是缺少很多个 http header 头部长度才正确
			n, _ = tcpConn.Read(responseContentBuf[leng:])
		}

		k++

		fmt.Println("===========================responseContentBuf=====start")
		fmt.Println(string(responseContentBuf))
		fmt.Println("===========================responseContentBuf=====end")

		fmt.Println("===========================leng=====start-for")
		fmt.Println(leng)
		fmt.Println("===========================leng=====end-for")

		end := leng + n
		if n > 0 {

			response := string(responseContentBuf[leng:end])
			//fmt.Println("===========================response=====start")
			//fmt.Println(response)
			//fmt.Println("===========================response=====end")
			s := strings.Split(response, "\r\n\r\n")

			var content string
			if len(s) == 2 {
				content = s[1]
				fileOffset += n - headerLength-1
			}else{
				content = s[0]
				fileOffset += n-1
			}
			//fmt.Println("===========================content=====start")
			//fmt.Println(content)
			//fmt.Println("===========================content=====end")
			newOffset = appendToFile(filename, content, newOffset)

			leng += n

			// 将文件信息保存到文件中
			fileInfo.FileOffset = fileOffset
			fileInfo.FileSize = fileInfo2.FileSize

			b, err := json.Marshal(fileInfo)
			if err != nil {
				fmt.Println("error: ", err, b)
			}
			saveToFile(string(b), dbFile)

		}else{

			fileSize2Int64 := getFileSize(filename)
			//fmt.Println("===========================fileSize2Int64=====start")
			//fmt.Println(fileSize2Int64)
			//fmt.Println("===========================fileSize2Int64=====end")
			fileSizeStr := strconv.FormatInt(fileSize2Int64, 10)
			fileSize2, _ := strconv.Atoi(fileSizeStr)

			//fmt.Println("===========================length2=====start")
			//fmt.Println(length2)
			//fmt.Println("===========================length2=====end")
			//
			//fmt.Println("===========================fileSize2=====start")
			//fmt.Println(fileSize2)
			//fmt.Println("===========================fileSize2=====end")

			if fileSize == fileSize2 {
				os.Remove(dbFile)
			}

			break
		}

		fmt.Printf("%d 次保存\n", k)
		if *isSleepParam != "0" {
			time.Sleep(time.Duration(40) * time.Second)
		}

	}

	return
}

func getResponseContentLength(responseStr string) int {
	pattern := "Content-Length: ([0-9]*?\r\n)"
	reg := regexp.MustCompile(pattern)
	match := reg.FindStringSubmatch(responseStr)
	//fmt.Println(responseStr, match[1])

	var count int = len(match)
	if count > 1 {
		lengthStr := strings.TrimSpace(match[1])
		length, _ := strconv.Atoi(lengthStr)
		return length
	} else {
		pattern := "Content-Range: bytes [0-9].*?-[0-9].*?/([0-9].*?)\r\n"
		reg := regexp.MustCompile(pattern)
		match := reg.FindStringSubmatch(responseStr)
		count := len(match)
		if count > 1 {
			length, _ := strconv.Atoi(match[1])
			return length
		} else {
			return 0
		}
	}
}

func saveToFile(doc string, file string) {
	dstFile, err := os.Create(file)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer dstFile.Close()

	dstFile.WriteString(doc)
}

func readFile(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return ""
	}

	return string(buf)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func getResponseHeader(conn net.Conn) (int, int)  {
	buf := make([]byte, 400, 500)
	len := 0
	for {
		n, _ := conn.Read(buf[len:])
		if n > 0 {
			len += n
		}else{
			break
		}
	}

	header := string(buf)
	header = strings.TrimSpace(header)
	length := getResponseContentLength(header)
	s := strings.Split(header, "\r\n\r\n")
	headerLength := strings.Count(s[0], "")

	return length, headerLength
}

// fileName:文件名字(带全路径)
// content: 写入的内容
func appendToFile(fileName string, content string, offset int64) (int64) {

	//var f *os.File
	//var err error
	//if Exists(fileName) == false {
	//	f, err = os.Create(fileName)
	//}else{
	//	// 以只写的模式，打开文件
	//	f, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	//}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)

	var newOffset int64 = 0

	if err != nil {
		fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
	} else {
		// 查找文件末尾的偏移量
		offset = 0
		n, _ := f.Seek(offset, 2)
		// 从末尾的偏移量开始写入内容
		n2, _ := f.WriteAt([]byte(content), n)
		//_, err = f.WriteString(content)
		newOffset = n + int64(n2)

	}
	defer f.Close()
	return newOffset
}

func getFileSize(filename string) int64  {

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	fileSize := fileInfo.Size()

	return fileSize
}

type fileMeta struct {
	Url        string
	FileOffset int
	FileSize int
}
