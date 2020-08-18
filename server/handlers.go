package server

import (
	"../config"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var cfg = config.GlobalConfig()

func RecoverPanic(writer http.ResponseWriter) {
	logger := localLogger
	if err := recover(); err != nil {
		logger.Print(err)
		WriteResponse(500, "Server Error", writer)
	}
}

func LogRequest(request *http.Request, status int, tstart time.Time) {
	now := time.Now()
	localLogger.Printf("%s %s %s %s %d %v", request.RemoteAddr, request.Method, request.Host, request.URL, status, now.Sub(tstart))
}

func WriteResponse(code int, message string, writer http.ResponseWriter) {
	mlength := strconv.FormatInt(int64(len(message)+1), 10)
	writer.Header().Set("Content-Length", mlength)
	writer.WriteHeader(code)
	fmt.Fprintf(writer, "%s\n", message)
}

func readFile(filepath string)  (string){
	f, err := os.Open(filepath)
	if err != nil {
		fmt.Println("read file fail", err)
		return ""
	}
	defer f.Close()

	fd, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("read to fd fail", err)
		return ""
	}

	return string(fd)
}

func downloadM3u8(writer http.ResponseWriter, request *http.Request)  {
	defer RecoverPanic(writer)
	params := request.URL.Query()
	m3u8File := params.Get("filename")
	startTime := params.Get("start_time")
	endTime := params.Get("end_time")
	if startTime == "" {
		WriteResponse(400, "Require start_time params", writer)
		LogRequest(request, 400, time.Now())
		return
	}
	if endTime == "" {
		WriteResponse(400, "Require end_time params", writer)
		LogRequest(request, 400, time.Now())
		return
	}
	if m3u8File == "" {
		WriteResponse(400, "Require filename params", writer)
		LogRequest(request, 400, time.Now())
		return
	}
	// 20200801142719_m_159626323578015972c2aa302aa1d2.m3u8 00:06:01 00:10:00
	fileName := m3u8File[15:]
	storePath := fmt.Sprintf("%s/%s/%s", cfg.BaseDir, splitPathByVideoName(fileName), m3u8File)
	fmt.Println(storePath)
	_,err := os.Stat(storePath)
	if os.IsNotExist(err){
		WriteResponse(400, "video not exists", writer)
		LogRequest(request, 400, time.Now())
		return
	}

	fileContent := readFile(storePath)
	if fileContent == "" {
		WriteResponse(400, "video not exists", writer)
		LogRequest(request, 400, time.Now())
		return
	}
	m3u8FileName := generateNewPlayList(fileName, fileContent, modifyTime(startTime), modifyTime(endTime))
	http.Redirect(writer, request, "http://127.0.0.1/temp/"+m3u8FileName, http.StatusFound)
	return
}

func splitPathByVideoName(name string) string {
	ext := path.Ext(name)
	fname := name[:len(name)-len(ext)]
	lname := len(fname)
	if lname <= 4 {
		return fname
	} else if lname <= 6 {
		return fmt.Sprintf("%s/%s", fname[0:4], fname[4:])
	} else {
		return fmt.Sprintf("%s/%s/%s", fname[0:4], fname[4:6], fname[6:])
	}
}

func generateNewPlayList(filename string, fileContent string, startTime int, endTime int) string {
	extinf := strings.Split(fileContent, "\n")
	startSecond := 0.0
	endSecond := 0.0
	appendFile := []string{"#EXTM3U", "#EXT-X-VERSION:3", "#EXT-X-MEDIA-SEQUENCE:0", "#EXT-X-ALLOW-CACHE:YES", "#EXT-X-TARGETDURATION:61"}
	length := len(extinf)
	for i:=0;i<length;i++ {
		if strings.Contains(extinf[i], "#EXTINF") {
			secondStr := strings.Split(strings.Split(extinf[i], ":")[1], ",")[0]
			secondFloat, err := strconv.ParseFloat(secondStr, 64)
			if err != nil {
				return ""
			}
			endSecond = startSecond + secondFloat
			startTimeFloat := float64(startTime)
			endTimeFloat := float64(endTime)
			if (startSecond >= startTimeFloat && startSecond <= endTimeFloat) || (endSecond >= startTimeFloat && endSecond <= endTimeFloat) {
				appendFile = append(appendFile, extinf[i])
				appendFile = append(appendFile, fmt.Sprintf("%s/%s",cfg.TempPrefix, extinf[i+1]))
			}
			startSecond = endSecond
		}
	}
	appendFile = append(appendFile, "#EXT-X-ENDLIST")
	if len(appendFile) <= 5 {
		return ""
	}
	tempFileName, err := generateTempFileName(filename, startTime, endTime, appendFile)
	if err != nil {
		return ""
	}
 	return tempFileName
}
// http://127.0.0.1:8080/downloadM3u8?filename=20200801142719_m_159626323578015972c2aa302aa1d2.m3u8&start_time=00:05:01&end_time=00:10:00
func generateTempFileName(filename string, start int, end int, appendFile []string) (string, error) {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s", filename, start, end))) // 需要加密的字符串为 123456
	cipherStr := h.Sum(nil)
	tempFileName := fmt.Sprintf("%s.m3u8", hex.EncodeToString(cipherStr))
	tempStorePath := fmt.Sprintf("%s/%s", cfg.TempDir, tempFileName)
	fmt.Println(tempStorePath)
	_,err := os.Stat(tempStorePath)
	if !os.IsNotExist(err){
		return tempFileName, nil
	}
	err = ioutil.WriteFile(tempStorePath, []byte(strings.Join(appendFile, "\n")), 777)
	if err != nil {
		return "", err
	}
	return tempFileName, nil
}
