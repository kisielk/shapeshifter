package main

import (
	"flag"
	"fmt"
	"html/template"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kisielk/shapeshifter"
	"github.com/youpy/go-wav"
)

func fatal(err error) {
	fmt.Printf("error: %s\n", err)
	os.Exit(1)
}

var config *shapeshifter.Config

var wavesDir = flag.String("waves", "waves/", "Directory containing available wavebanks")

func main() {
	flag.Parse()
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	config, err = shapeshifter.Read(f)
	if err != nil {
		fatal(err)
	}
	pngf, err := os.Create("output.png")
	if err != nil {
		fatal(err)
	}
	img := shapeshifter.DrawWave(config[0].Waves[6])
	png.Encode(pngf, img)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/bank/", handleBank)
	http.HandleFunc("/draw", handleDraw)
	http.HandleFunc("/play", handlePlay)
	http.ListenAndServe(":8080", nil)
}

var indexTemplate = template.Must(template.New("index").Parse(`
<html>
<head><title>Intellijel Shapeshifter Customizer</title></head>
<body>
<ol>
{{range $i, $bank := .}}
<li><a href="/bank/{{$i}}">{{$bank.Name}}</a></li>
{{end}}
</ol>
</body>
</html>
`))

var bankTemplate = template.Must(template.New("bank").Parse(`
<html>
<head><title>Intellijel Shapeshift Customizer - Bank {{.Name}}</title></head>
<body>
<img src="/draw?bank={{.Num}}&wave=0" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=0" controls preload="none"></audio>
<img src="/draw?bank={{.Num}}&wave=1" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=1" controls preload="none"></audio>
<img src="/draw?bank={{.Num}}&wave=2" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=2" controls preload="none"></audio>
<img src="/draw?bank={{.Num}}&wave=3" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=3" controls preload="none"></audio>
<img src="/draw?bank={{.Num}}&wave=4" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=4" controls preload="none"></audio>
<img src="/draw?bank={{.Num}}&wave=5" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=5" controls preload="none"></audio>
<img src="/draw?bank={{.Num}}&wave=6" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=6" controls preload="none"></audio>
<img src="/draw?bank={{.Num}}&wave=7" width="256" height="128">
<audio src="/play?bank={{.Num}}&wave=7" controls preload="none"></audio>
</body>
</html>
`))

func handleIndex(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, config)
}

func handleBank(w http.ResponseWriter, r *http.Request) {
	num := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	n, _ := strconv.ParseInt(num, 10, 64)
	bankTemplate.Execute(w, map[string]string{
		"Name": config[n].Name,
		"Num":  num,
	})
}

func handleDraw(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	bank, _ := strconv.ParseInt(r.Form.Get("bank"), 10, 64)
	wave, _ := strconv.ParseInt(r.Form.Get("wave"), 10, 64)
	w.Header().Set("Content-Type", "image/png")
	img := shapeshifter.DrawWave(config[int(bank)].Waves[int(wave)])
	png.Encode(w, img)
}

const (
	defaultPlayDuration = 5 * time.Second
	sampleRate          = 44000
)

func handlePlay(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	duration, err := time.ParseDuration(r.Form.Get("duration"))
	if err != nil {
		duration = defaultPlayDuration
	}
	bank, _ := strconv.ParseInt(r.Form.Get("bank"), 10, 64)
	wave, _ := strconv.ParseInt(r.Form.Get("wave"), 10, 64)
	w.Header().Set("Content-Type", "audio/x-wav")

	waveSamples := config[int(bank)].Waves[int(wave)]
	numSamples := sampleRate * int(duration.Seconds())
	writer := wav.NewWriter(w, uint32(numSamples), 1, 44000, 16)
	samples := make([]wav.Sample, numSamples)
	for i := range samples {
		samples[i].Values[0] = int(waveSamples[i%len(waveSamples)])
	}
	writer.WriteSamples(samples)
}
