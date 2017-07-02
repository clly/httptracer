package main

import (
	"flag"
	"log"

	"net/http"

	"fmt"

	"gitlab.com/clly/httptracer"
	"gitlab.com/clly/toolbox"
)

var (
	host string
)

func main() {
	flag.StringVar(&host, "Host", "", "")
	flag.Parse()

	u := toolbox.ParseURL(host)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	tracer, err := httptracer.Trace(req)
	fmt.Printf("%+v\n", tracer.IP)
	fmt.Printf("%+v\n", tracer.Timers)
}
