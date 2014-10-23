package main

import (
	"flag"
	"fmt"
	"html/template"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/kisielk/shapeshifter"
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
	log.Println(bank, wave)
	w.Header().Set("Content-Type", "image/png")
	img := shapeshifter.DrawWave(banks[int(bank)].Waves[int(wave)])
	png.Encode(w, img)
}
