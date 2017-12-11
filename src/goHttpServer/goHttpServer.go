package main

import (
	"context"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"util/dogstats"

	"github.com/bitly/go-simplejson"
)

var (
	HttpCalls expvar.Int
)

func init() {
	expvar.Publish("http_calls", &HttpCalls)
}

func slow(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	log.Println("Header:", req.Header)
	log.Println("URL:", req.URL)
	log.Println("Params:", req.Form)
	time.Sleep(2 * time.Second)
	w.Write([]byte("slow"))
}

// JSONDecodeMap is a method whicg decodes a byte array into a map and
// returns along with decoded json object.  Both map and json object
// point to the same memory location
func JSONDecodeMap(body []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	jsonObj, err := simplejson.NewJson(body)
	if err != nil {
		//return nil, err
		return nil, errors.New("Could not decode with JSON")
	}
	data, err = jsonObj.Map()
	if err != nil {
		//return nil, err
		return nil, errors.New("Could not type cast to map")
	}
	return data, nil // return in JSON
}

func StatusOK(w http.ResponseWriter, req *http.Request) {
	HttpCalls.Add(1)
	req.ParseForm()
	defer req.Body.Close()
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err.Error())
		StatusNotOK(w, req)
		return
	}
	// print out debugging info
	log.Println("Request Header:", req.Header)
	log.Println("Query URL:", req.URL)
	//log.Println("req.RequestURI:", req.RequestURI)
	log.Println("Query Params:", req.Form)

	// send metrics to datadog
	if req.Form.Get("campaign") != "" {
		dogstats.Incr("conversion_type", req.Form.Get("conversion_type"))
	}
	switch req.Method {
	case "GET":
		// Serve the resource.
	case "POST":
		// Create a new record.
		m, err := JSONDecodeMap(reqBody)
		if err != nil {
			log.Println(err.Error())
			StatusNotOK(w, req)
			return
		} else {
			log.Println("Requesst Body:", m)
		}
	case "PUT":
		// Update an existing record.
	case "DELETE":
		// Remove the record.
	default:
		// Give an error message.
	}

	//w.Header().Add("Content-Type", "application/json")
	//w.Write([]byte("{\"status\": 200}\n"))
	w.Write([]byte("ok"))
}

func StatusNotOK(w http.ResponseWriter, req *http.Request) {
	HttpCalls.Add(1)
	req.ParseForm()
	log.Println("Header:", req.Header)
	log.Println("URL:", req.URL)
	log.Println("Params:", req.Form)
	//w.Header().Add("Content-Type", "application/json")
	//w.Write([]byte("{\"status\": 200}\n"))
	//w.Write([]byte("ok"))
	//http.Error(w, "Not Ok", 404)
	http.Error(w, "UPSTREAM ERROR", http.StatusBadGateway)
}

func main() {
	port := flag.Int("port", 5678, "port to listen on")
	dogStatsdAddr := flag.String("dog-stats-address", "", "host:port (eg. 127.0.0.1:8125)")
	dogStatsdNamespace := flag.String("dog-stats-ns", "attribution", "namespace for datadog")
	dogStatsdEnviron := flag.String("dog-stats-env", "dev", "environment tag for datadog")
	dogStatsdRegion := flag.String("dog-stats-region", "us-east-2", "region tag for datadog")
	dogStatsdSampleRate := flag.Float64("dog-stats-sample-rate", 1, "sample rate of dogstatsd (0 < rate <= 1.0)")
	flag.Parse()

	dogstatsInstance, err := dogstats.NewDogStatsd(
		*dogStatsdAddr, *dogStatsdNamespace, *dogStatsdEnviron, *dogStatsdRegion, *dogStatsdSampleRate,
	)
	if err == nil {
		dogstats.DogStatsdInstance = dogstatsInstance
		log.Println("finished setup dogstatd on", dogstats.DogStatsdInstance.Addr)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	url := fmt.Sprintf(":%d", *port)
	srv := &http.Server{Addr: url, Handler: http.DefaultServeMux}

	go func() {
		sig := <-sigs
		log.Printf("terminating on signal %d", sig)
		//fmt.Println("Press enter to shutdown server")
		//fmt.Scanln()
		log.Println("Shutting down server...")
		if err := srv.Shutdown(context.Background()); err != nil { // HL
			log.Fatalf("could not shutdown: %v", err)
		}
	}()

	http.HandleFunc("/", StatusOK)
	http.HandleFunc("/slow", slow)
	http.HandleFunc("/fail", StatusNotOK)

	log.Println("starting goHttpServer on", url)
	err = srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
	log.Println("Server gracefully stopped")
}
