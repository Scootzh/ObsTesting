package controll

import (
	"fyne.io/fyne/v2/widget"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/sources"
)

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
