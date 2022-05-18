package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/overlorddamygod/go-concurrent-video-downloader/downloader"
	"github.com/webview/webview"
)

//go:embed www
var fs embed.FS

func main() {
	var now time.Time
	var workers int64 = 20
	channel := make(chan downloader.Status)

	w := webview.New(true)
	defer w.Destroy()
	w.SetSize(600, 420, webview.HintNone)

	if err := os.Mkdir(".downloadertemp", os.ModePerm); err != nil {
		log.Println("Directory already exists")
	}

	// Create a GoLang function callable from JS
	w.Bind("download", func(url string) bool {
		fmt.Println("Downloading", url)
		now = time.Now()

		go downloader.DownloadConcurrently(workers, url, channel)
		return false
	})
	w.Bind("getParts", func() {
		w.Dispatch(func() {
			code := fmt.Sprintf(`setParts(%d)`, workers)
			w.Eval(code)
		})
	})

	go func() {
		for {
			s, ok := <-channel
			if ok {
				if s.Main {
					fmt.Println("COMPLETEDDD")
					w.Dispatch(func() {
						code := fmt.Sprintf(`setProgress(%d)`, s.Value)
						w.Eval(code)
						code = fmt.Sprintf(`setMessage(%s)`, fmt.Sprintf("Download completed %f", time.Since(now).Seconds()))
						w.Eval(code)
					})
					if s.Err != nil {
						w.Dispatch(func() {
							code := fmt.Sprintf(`setMessage(%s)`, s.Err.Error())
							w.Eval(code)
						})
						return
					}
					return
				} else {
					code := fmt.Sprintf(`setPartProgress(%d, %t)`, s.Index, s.Err != nil)

					w.Dispatch(func() {
						w.Eval(code)
					})
				}
			} else {
				return
			}
		}
	}()

	w.SetTitle("Go Concurrent Video Downloader")

	// Create UI with data URI
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(http.FS(fs)))

	w.Navigate(fmt.Sprintf("http://%s/www", ln.Addr()))

	w.Run()
}
