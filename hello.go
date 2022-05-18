package main

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/events"
	"github.com/andreykaipov/goobs/api/requests/sources"
)

type logger struct{}

func (l *logger) Printf(f string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[1;33m"+f+"\033[0m\n", v...)
}

func updateTime(client *goobs.Client, mute *widget.Label) {
	client.Sources.ToggleMute(&sources.ToggleMuteParams{Source: "Game"})
	//	resp, _ := client.Sources.GetVolume(&sources.GetVolumeParams{Source: "Game"})

}
func setAudio(client *goobs.Client, src string, slider *widget.Slider) {
	slider.OnChanged = func(f float64) {
		if _, err := client.Sources.SetVolume(&sources.SetVolumeParams{
			Source: src,
			Volume: slider.Value,
		}); err != nil {
			panic(err)
		}
	}
}

func getSceneList(client *goobs.Client) []string {
	scenes, _ := client.Scenes.GetSceneList()
	var ar []string
	for sc := range scenes.Scenes {
		ar = append(ar, string(scenes.Scenes[sc].Name))
	}
	return ar
}

func main() {

	client, err := goobs.New(
		os.Getenv("WSL_HOST")+":4444",
		goobs.WithPassword("Passwords"),               // optional
		goobs.WithDebug(os.Getenv("OBS_DEBUG") != ""), // optional
		goobs.WithLogger(&logger{}),                   // optional
	)
	if err != nil {
		fmt.Printf("Could not connect to obs-webbsocket\n")
		panic(err)
	}
	defer client.Disconnect()

	version, _ := client.General.GetVersion()
	fmt.Printf("Websocket server version: %s\n", version.ObsWebsocketVersion)
	fmt.Printf("OBS Studio version: %s\n", version.ObsStudioVersion)

	mute := widget.NewLabel("")

	// This event loop is in the foreground now. If you mess around in OBS,
	// you'll see different events popping up! Basically monitoring everything
	go func() {
		for event := range client.IncomingEvents {
			switch e := event.(type) {
			//Monitor volumechange
			case *events.SourceVolumeChanged:
				fmt.Printf("Volume changed for %-25q: %f\n", e.SourceName, e.Volume*100)
			//Monitor mutestate
			case *events.SourceMuteStateChanged:
				resp, _ := client.Sources.GetVolume(&sources.GetVolumeParams{Source: e.SourceName})
				test := fmt.Sprintf("%q is muted? %t\n", e.SourceName, resp.Muted)
				mute.SetText(test)
				fmt.Printf("%q is muted? %t\n", e.SourceName, resp.Muted)
			default:
				log.Printf("Unhandled event: %#v", e.GetUpdateType())
			}
		}
	}()

	a := app.New()
	w := a.NewWindow("Hello World")
	//Get array of scenes for dropDownlist
	ar := getSceneList(client)
	t := widget.NewSelect(
		ar,
		func(s string) {
			fmt.Printf("Selected: %s\n", s)
		})

	t.OnChanged = func(s string) {
		//Should work, but doesnt... Trying to update the scene list with refresh
		ar = getSceneList(client)
		t.Refresh()
	}
	w.SetContent(mute)
	w.Show()
	src := widget.NewEntry()
	src.SetText("Game")

	//slider change OBS source volume
	d, _ := client.Sources.GetVolume(&sources.GetVolumeParams{Source: "Game"})
	data := binding.BindFloat(&d.Volume)
	slider := widget.NewSliderWithData(0.00001, 1, data)
	slider.Step = 0.00001
	setAudio(client, src.Text, slider)
	//Label not working when changing OBS...
	lab := widget.NewLabelWithData(
		binding.FloatToString(data),
	)

	w2 := a.NewWindow("Larger")

	w.SetContent(mute)
	w2.SetContent(
		container.NewVBox(
			widget.NewButton("MuteTest", func() {
				updateTime(client, mute)
			}),
			slider,
			lab,
			src,
			widget.NewButton("Saves", func() {
				setAudio(client, src.Text, slider)
			}),
			t,
			widget.NewButton("UpdateList", func() {
				ar = getSceneList(client)
				t.Refresh()
			}),
		),
	)
	w2.Resize(fyne.NewSize(400, 400))
	w2.Show()

	a.Run()
}
