package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	TargetUrl = "http://deviceshifu-plate-reader.deviceshifu.svc.cluster.local/get_measurement"
)

func main() {
	client := &http.Client{}
	timeInterval, err := strconv.Atoi(os.Getenv("poll_interval"))
	if err != nil {
		log.Println("warning: env \"poll_interval\" not found, default 10 seconds.")
		timeInterval = 10
	} else {
		log.Printf("time interval of polling is set as %d seconds.\n", timeInterval)
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second * time.Duration(timeInterval))
	defer ticker.Stop()

	for {
		select {
		case s := <-c:
			log.Println("get signal:", s)
			return
		case <-ticker.C:
			sendHttpReq(client)
		}
	}

}

func sendHttpReq(client *http.Client) {
	req, err := http.NewRequest("GET", TargetUrl, nil)

	if err != nil {
		log.Println("new http request err:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("send http request err:", err)
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var sum float64 = 0
	var cnt int = 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("read http response err:", err)
			return
		}
		line = strings.TrimRight(line, " \r\n")
		for _, num := range strings.Split(line, " ") {
			n, err := strconv.ParseFloat(num, 64)
			if err != nil {
				log.Printf("parse %s to float err: %v\n", num, err)
				return
			}
			sum += n
			cnt++
		}
	}
	sum /= float64(cnt)

	log.Printf("average measurement: %.3f\n", sum)
}
