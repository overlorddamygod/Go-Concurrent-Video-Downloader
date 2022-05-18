package downloader

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func DownloadRange(index int, url string, start int64, end int64, size int64) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	fmt.Println("Download Started", index)

	respByte, err := ioutil.ReadAll(resp.Body)
	fmt.Println(start, end, "len", len(respByte))

	if err != nil {
		log.Fatalln(err)
	}

	if err := ioutil.WriteFile(fmt.Sprintf(".downloadertemp/[%d]%d-%d.tmp", index, start, end), respByte, 0644); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Downloaded", index)
}

func GetContentLength(url string) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	// fmt.Println(resp.Header)

	return resp.ContentLength, nil
}

type Status struct {
	Index int
	Err   error
	Main  bool
	Value int
}

func DownloadConcurrently(workers int64, url string, c chan Status) {
	now := time.Now()

	contentLength, err := GetContentLength(url)

	contentRange := int64((float64(contentLength) / float64(workers)))

	fmt.Printf("File size: %dMB\n", contentLength/1000000)

	if err != nil {
		c <- Status{
			Index: -1,
			Err:   err,
			Main:  true,
		}
		return
	}

	var wg sync.WaitGroup

	var i int64
	var ranges [][]int64
	for i = 0; i < int64(workers); i++ {
		wg.Add(1)
		start := i * contentRange
		end := start + contentRange
		end -= 1
		if i == workers-1 {
			end = contentLength - 1
		}

		r := []int64{start, end}
		ranges = append(ranges, r)

		go func(i int) {
			defer func() {
				wg.Done()
				c <- Status{Index: i, Err: nil, Main: false, Value: 100}
			}()
			// fmt.Println(start, end)
			DownloadRange(i, url, start, end, contentLength)
		}(int(i))
	}

	wg.Wait()

	fmt.Println("All parts downloaded Combining")

	file, err := os.OpenFile("vid.mp4", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	// fmt.Println(ranges)

	for i = 0; i < int64(workers); i++ {
		fmt.Println(ranges[i])
		filepath := fmt.Sprintf(".downloadertemp/[%d]%d-%d.tmp", i, ranges[i][0], ranges[i][1])
		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatal(err)
		}
		file.Write(data)
		e := os.Remove(filepath)
		if e != nil {
			log.Fatal(e)
		}
	}
	c <- Status{Index: -1, Err: nil, Main: true, Value: 100}
	fmt.Println("Download completed.... Total Time Elapsed", time.Since(now).Seconds())
}
