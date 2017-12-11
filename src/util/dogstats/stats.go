package dogstats

import (
	"fmt"
	"log"

	"github.com/DataDog/datadog-go/statsd"
)

var DogStatsdInstance *DogStatsd

// --------------------------------------------------
// Datadog XXX TODO: move into its own file?
// --------------------------------------------------
type DogStatsd struct {
	c          *statsd.Client
	Addr       string
	Namespace  string
	Environ    string
	Region     string
	sampleRate float64
}

func NewDogStatsd(addr, namespace, environ, region string, sampleRate float64) (*DogStatsd, error) {
	if addr == "" {
		return nil, fmt.Errorf("empty address for dogstatsd")
	}
	log.Printf("connecting to dogstatsd on %s with nampspace %s, environ %s and region %s with sample rate %f.",
		addr, namespace, environ, region, sampleRate)
	c, err := statsd.New(addr)
	if err != nil {
		log.Println("setup dogstatd connection error", err)
		return nil, err
	}
	c.Namespace = namespace + "."
	c.Tags = append(c.Tags, namespace)
	c.Tags = append(c.Tags, "region:"+region)
	c.Tags = append(c.Tags, "env:"+environ+"-"+namespace)

	dogStatsd := &DogStatsd{
		c:          c,
		Addr:       addr,
		Namespace:  namespace,
		Environ:    environ,
		Region:     region,
		sampleRate: sampleRate,
	}
	return dogStatsd, nil
}

// Increments the counter using the `name` and `value` as tag parameters
func Incr(key, value string) {
	if DogStatsdInstance == nil {
		return
	}
	DogStatsdInstance.incr(key, value)
}

func (ds *DogStatsd) incr(key, value string) {
	//log.Println("log stats for", ds.Namespace, ds.Namespace+"-"+ds.Environ, key, key+":"+value)
	if err := ds.c.Incr("count", []string{key, key + ":" + value}, ds.sampleRate); err != nil {
		log.Printf("increase counter for %s:%s with error %v:", key, value, err)
	}
}

// --------------------------------------------------
