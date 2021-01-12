package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const USAGE = `makeRequest
Usage:
  makeRequest --url=<URL> [--profile=<COUNT>]
  makeRequest --help
Flags:
  --url=<URL>        The URL that this program will make a request to
  --profile=<COUNT>  The number of times our program will make a request to <URL> and display stats
  --help             Display the usage of this program
`

const HEADER = "GET %v HTTP/1.0\r\nHost: %v\r\n\r\n"

var dialer = net.Dialer{
	Timeout: time.Minute,
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func validate(displayHelp bool, urlStr string, profileCount int) error {
	if flag.NFlag() > 2 || flag.NFlag() == 0 {
		return errors.New("Invalid number of flags. Add --help to see available flags")
	}
	if flag.NFlag() == 2 && (urlStr == "" || profileCount < 0) {
		return errors.New("Invalid flags supplied. Add --help to see available flags")
	}
	if displayHelp == true {
		fmt.Print(USAGE)
		os.Exit(0)
	}
	if urlStr == "" {
		return errors.New("Invalid flag supplied. Add --help to see available flags")
	}
	return nil
}

func makeRequest(u *url.URL) (string, int) {
	conn, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("%s:https", u.Hostname()), nil)
	checkError(err)
	reqHeader := fmt.Sprintf(HEADER, u.Path, u.Hostname())
	defer conn.Close()
	_, err = conn.Write([]byte(reqHeader))
	checkError(err)
	res, err := ioutil.ReadAll(conn)
	checkError(err)
	resStr := string(res)
	status, _ := strconv.Atoi(strings.Split(resStr, " ")[1])
	return resStr, status
}

func profile(url *url.URL, profileCount int) {
	var times []int
	var errors []int
	maxSize := float64(0)
	minSize := math.MaxFloat64
	totalTime := 0
	for i := 1; i <= profileCount; i++ {
		startTime := time.Now()
		res, status := makeRequest(url)
		resTime := int(time.Since(startTime).Milliseconds())
		resSize := float64(len(res))
		totalTime += resTime
		times = append(times, resTime)
		if status != 200 {
			errors = append(errors, status)
		}
		minSize = math.Min(minSize, resSize)
		maxSize = math.Max(maxSize, resSize)
	}
	sort.Ints(times)
	median := times[profileCount/2]
	if len(times)%2 == 0 {
		median = (times[profileCount/2] + times[profileCount/2-1]) / 2
	} else {
		median = times[profileCount/2]
	}
	meanTime := int(float64(totalTime) / float64(profileCount))
	successPercentage := (float64(profileCount) - float64(len(errors))) / float64(profileCount) * 100
	fmt.Println("Profile information for: ", strings.ToLower(url.Hostname()))
	fmt.Println("Number of requests: ", profileCount)
	fmt.Println("The fastest time (ms): ", times[0])
	fmt.Println("The slowest time (ms): ", times[profileCount-1])
	fmt.Println("The mean time (ms): ", meanTime)
	fmt.Println("The median time (ms): ", median)
	fmt.Println("The percentage of requests that succeeded: ", successPercentage, "%")
	fmt.Println("The error codes returned that weren't a success: ", errors)
	fmt.Println("The size in bytes of the smallest response: ", int(minSize))
	fmt.Println("The size in bytes of the largest response: ", int(maxSize))
}

func main() {
	helpPtr := flag.Bool("help", false, "Help command")
	urlPtr := flag.String("url", "", "URL")
	profilePtr := flag.Int("profile", -1, "URL")
	flag.Parse()
	err := validate(*helpPtr, *urlPtr, *profilePtr)
	checkError(err)

	urlStr := *urlPtr
	if urlStr[:4] != "http" {
		urlStr = "https://" + urlStr
	}
	url, err := url.Parse(urlStr)
	checkError(err)

	if url.Path == "" {
		url.Path = "/"
	}

	if *profilePtr <= 0 {
		res, _ := makeRequest(url)
		fmt.Println(res)
	} else {
		profile(url, *profilePtr)
	}
}
