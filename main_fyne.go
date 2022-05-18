package main

import (
	"image/color"
	"log"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/overlorddamygod/go-concurrent-video-downloader/downloader"
)

func fyna_main() {
	var workers int64 = 20

	a := app.New()
	w := a.NewWindow("Hello World")
	w.Resize(fyne.NewSize(500, 300))
	w.CenterOnScreen()

	var progress = []*widget.ProgressBar{}
	var i int64 = 0
	for i = 0; i < workers; i++ {
		progress = append(progress, widget.NewProgressBar())
	}

	entry := widget.NewEntry()
	finalProgress := widget.NewProgressBar()

	channel := make(chan downloader.Status)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Url", Widget: entry}},
		OnSubmit: func() {
			// exec.Command("cmd", "start", "vid.mp4").Run()
			log.Println("Form submitted:", entry.Text)
			downloader.DownloadConcurrently(workers, entry.Text, channel)
		},
		SubmitText: "Download",
	}

	go func() {
		for {
			select {
			case s := <-channel:
				if s.Main {
					finalProgress.SetValue(float64(s.Value))
					return
				} else {
					progress[s.Index].SetValue(float64(s.Value))
					finalProgress.SetValue(finalProgress.Value + float64(100/workers))
				}
			}
		}
	}()

	progressContainer1 := container.New(layout.NewHBoxLayout(),
		progress[0],
		progress[1],
		progress[2],
		progress[3],
		progress[4],
	)

	progressContainer2 := container.New(layout.NewHBoxLayout(),
		progress[5],
		progress[6],
		progress[7],
		progress[8],
		progress[9],
	)
	progressContainer3 := container.New(layout.NewHBoxLayout(),
		progress[10],
		progress[11],
		progress[12],
		progress[13],
		progress[14],
	)
	progressContainer4 := container.New(layout.NewHBoxLayout(),
		progress[15],
		progress[16],
		progress[17],
		progress[18],
		progress[19],
	)
	header := container.New(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		canvas.NewText("Go Concurrent Video Downloader", color.White),
		layout.NewSpacer(),
	)

	w.SetContent(container.New(layout.NewVBoxLayout(), header, form, progressContainer1, progressContainer2, progressContainer3, progressContainer4, finalProgress))
	w.ShowAndRun()
}
