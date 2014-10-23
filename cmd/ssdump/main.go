package main

import (
	"flag"
	"fmt"
	"html/template"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kisielk/shapeshifter"
	"github.com/youpy/go-wav"
)

func fatal(err error) {
	fmt.Printf("error: %s\n", err)
	os.Exit(1)
}

var banks []shapeshifter.Bank

func main() {
	flag.Parse()
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	banks, err = shapeshifter.Read(f)
	if err != nil {
		fatal(err)
	}
	pngf, err := os.Create("output.png")
	if err != nil {
		fatal(err)
	}
	img := shapeshifter.DrawWave(banks[0].Waves[6])
	png.Encode(pngf, img)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/draw", handleDraw)
	http.HandleFunc("/play", handlePlay)
	http.ListenAndServe(":8080", nil)
}

var indexTemplate = template.Must(template.New("index").Parse(`
<html>
<head><title>Banks</title></head>
<body>
<ol>
{{ range . }}
<li><a href="/bank/{{.Name}}">{{.Name}}</a></li>
{{ end }}
</ol>
</body>
</html>
`))

func handleIndex(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, banks)
}

func handleDraw(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	bank, _ := strconv.ParseInt(r.Form.Get("bank"), 10, 64)
	wave, _ := strconv.ParseInt(r.Form.Get("wave"), 10, 64)
	w.Header().Set("Content-Type", "image/png")
	img := shapeshifter.DrawWave(banks[int(bank)].Waves[int(wave)])
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

	waveSamples := banks[int(bank)].Waves[int(wave)]
	numSamples := sampleRate * int(duration.Seconds())
	writer := wav.NewWriter(w, uint32(numSamples), 1, 44000, 16)
	samples := make([]wav.Sample, numSamples)
	for i := range samples {
		samples[i].Values[0] = int(waveSamples[i%len(waveSamples)])
	}
	writer.WriteSamples(samples)
}
