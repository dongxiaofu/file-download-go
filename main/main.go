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
)

/**
 go run main.go -host=chugang.net -path=/ -filename=index.html
 go run main.go -host=i1.whymtj.com -path=/uploads/tu/201904/9999/75880a6ff0.jpg?t==3337 -filename=g.jpg
 */

func main() {

	hostParam := flag.String("host", "", "host, 必填")
	pathParam := flag.String("path", "", "path，必填")
	filenameParam := flag.String("filename", "", "filename，必填")

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
	fmt.Println(address)
	// 发起 http 请求
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)

	checkError(err)

	tcpConn, err2 := net.DialTCP("tcp", nil, tcpAddr)

	tcpConn2, _ := net.DialTCP("tcp", nil, tcpAddr)

	var fileInfo fileMeta
	fileInfo = fileMeta{
		Url:      "1",
		FileSize: 0,
	}

	//fmt.Println(fileInfo)
	fileInfo.Url = address

	checkError(err2)

	var buffer bytes.Buffer

	requestHeaderGet := "GET " + path + " HTTP/1.1 \r\n"
	buffer.WriteString(requestHeaderGet)

	// 读取下载文件的元数据
	var fileInfo2 fileMeta
	fileInfoString := readFile(dbFile)
	if fileInfoString == "" {
		fileInfo2 = fileMeta{
			Url:      address,
			FileSize: 0,
		}
	} else {
		var fileInfoJson = []byte(fileInfoString)

		err5 := json.Unmarshal(fileInfoJson, &fileInfo2)
		if err5 != nil {
			panic("error")
			return
		}
	}

	if fileInfo2.FileSize != 0 {
		fmt.Println("yes")
		rangeHeader := "Range:bytes=" + strconv.Itoa(fileInfo2.FileSize) + "-\r\n"
		buffer.WriteString(rangeHeader)
	}

	buffer.WriteString("Cache-Control: no-cache \r\n")
	buffer.WriteString("Pragma: no-cache \r\n")
	buffer.WriteString("Connection: keep-alive \r\n")

	//buffer.WriteString("Range:bytes=1000-\r\n")
	requestHeaderHost := "Host: " + host + "\r\n\r\n"
	buffer.WriteString(requestHeaderHost)
	//buffer.WriteString("Connection: keep-alive \r\n\r\n")
	//buffer.WriteString("Pragma: no-cache \r\n\r\n")
	//buffer.WriteString("Cache-Control: no-cache \r\n\r\n")
	//buffer.WriteString("Upgrade-Insecure-Requests: 1 \r\n\r\n")
	//buffer.WriteString("User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36 \r\n\r\n")
	//buffer.WriteString("Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3 \r\n\r\n")
	//buffer.WriteString("Accept-Encoding: gzip, deflate \r\n\r\n")
	//buffer.WriteString("Accept-Language: zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7 \r\n\r\n")

	requestHeader := buffer.String()

	_, err3 := tcpConn.Write([]byte(requestHeader))

	checkError(err3)


	// 获取头部信息
	_, err33 := tcpConn2.Write([]byte(requestHeader))
	checkError(err33)

	length2, headerLength := getResponseHeader(tcpConn2)

	var fileSize int

	fileSize = fileInfo2.FileSize

	fmt.Println(fileSize)

	defer tcpConn.Close()

	length := length2 + headerLength
	responseContentBuf := make([]byte, length, length)
	leng := 0
	var k int
	for {
		n, _ := tcpConn.Read(responseContentBuf[leng:])
		k++
		if n > 0 {

			response := string(responseContentBuf[leng:leng + n])
			s := strings.Split(response, "\r\n\r\n")

			var content string
			if len(s) == 2 {
				content = s[1]
				fileSize += n - headerLength
			}else{
				content = s[0]
				fileSize += n
			}

			appendToFile(filename, content)

			leng += n

			// 将文件信息保存到文件中
			fileInfo.FileSize = fileSize

			b, err := json.Marshal(fileInfo)
			if err != nil {
				fmt.Println("error: ", err, b)
			}
			saveToFile(string(b), dbFile)

		}else{

			fileSize2Int64 := getFileSize(filename)
			fmt.Println("===========================fileSize2Int64=====start")
			fmt.Println(fileSize2Int64)
			fmt.Println("===========================fileSize2Int64=====end")
			fileSizeStr := strconv.FormatInt(fileSize2Int64, 10)
			fileSize2, _ := strconv.Atoi(fileSizeStr)

			fmt.Println("===========================length2=====start")
			fmt.Println(length2)
			fmt.Println("===========================length2=====end")

			fmt.Println("===========================fileSize2=====start")
			fmt.Println(fileSize2)
			fmt.Println("===========================fileSize2=====end")

			if length2 - fileSize2 >= 0 && length2 - fileSize2 < 5 {
				os.Remove(dbFile)
			}

			if fileSize2 - length2 >= 0 && fileSize2 - length2 < 5 {
				os.Remove(dbFile)
			}

			break
		}

		//time.Sleep(time.Duration(40) * time.Second)
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
	fmt.Println(header)
	fmt.Println("=================")
	s := strings.Split(header, "\r\n\r\n")
	fmt.Println(s[0])
	headerLength := strings.Count(s[0], "")

	return length, headerLength
}

// fileName:文件名字(带全路径)
// content: 写入的内容
func appendToFile(fileName string, content string) error {

	//var f *os.File
	//var err error
	//if Exists(fileName) == false {
	//	f, err = os.Create(fileName)
	//}else{
	//	// 以只写的模式，打开文件
	//	f, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	//}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)

	if err != nil {
		fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
	} else {
		// 查找文件末尾的偏移量
		n, _ := f.Seek(0, os.SEEK_END)
		// 从末尾的偏移量开始写入内容
		_, err = f.WriteAt([]byte(content), n)
		//_, err = f.WriteString(content)
	}
	defer f.Close()
	return err
}

func getFileSize(filename string) int64  {

	fileInfo, _ := os.Stat(filename)
	fileSize := fileInfo.Size()

	return fileSize
}

type fileMeta struct {
	Url      string
	FileSize int
}
