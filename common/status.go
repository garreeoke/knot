package common

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"applariat.io/cluster-manager/types"
	"applariat.io/propeller/kube"

	"strings"
)

const (
	// DefaultTimeOut The default timeout to check master server health
	DefaultTimeOut = 30 * time.Minute
	// DefaultSleepTime The default sleep time between master health checks
	DefaultSleepTime = 30 * time.Second
)

// APIStatusInput input arguments to ApiStatus function
type K8APIStatusInput struct {
	// Set to false for tls true to use basic auth
	Basic bool
	// Required
	// Master is the API server DNS
	Master string
	// Optional
	// UserName is used to authenticate against the API
	// Required for basic auth
	UserName string
	// Optional
	// Password is used to authenticate against the API
	// Required for basic auth
	Password string
	// Optional
	// Certs is used to authenticate against the API
	// Required for tls
	// Keys for map are clientCert, clientKey, caCert
	Certs map[string]string
	// Optional
	// Timeout how many minutes to wait before giving up on healthy master
	TimeOut time.Duration
	// Optional how long to sleep between health checks
	// SleepTime
	SleepTime time.Duration
	// Data being passed around, mainly used to send messages to the channel
	ClusterData *types.ClusterData
}

// APIStatusOutput Return type from status func
type K8APIStatusOutput struct {
	APIHealthy bool
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *K8APIStatusInput) Validate() error {

	if s.Master == "" {
		return errors.New("Missing required parameter Master")
	}
	/*
	if s.UserName == "" {
		return errors.New("Missing required parameter UserName")
	}
	if s.Password == "" {
		return errors.New("Missing required parameter Password")
	}
	*/
	if s.TimeOut == 0 {
		s.TimeOut = DefaultTimeOut
	}
	if s.SleepTime == 0 {
		s.SleepTime = DefaultSleepTime
	}
	return nil
}

// APIStatus will attempt to hit the /healthz endpoint on
// the kubernetes master server. Once it get's a 200ok, it will
// return healthy, otherwise it will either time out or return false
func K8ApiStatus(in *K8APIStatusInput) (*K8APIStatusOutput,error) {

	err := in.Validate()
	if err != nil {
		return &K8APIStatusOutput{APIHealthy: false}, err
	}
	tchan := doTimeOut(in.TimeOut)
	hchan := doHealthCheck(in)

	isHealthy := false
	select {
	case isHealthy = <-hchan:
	case <-tchan:
	}

	return &K8APIStatusOutput{APIHealthy: isHealthy}, nil

}

func doTimeOut(t time.Duration) chan bool {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(t)
		timeout <- true
		close(timeout)
	}()
	return timeout
}

// Runs a health check ping
func doHealthCheck(in *K8APIStatusInput) chan bool {

	healthy := make(chan bool, 1)
	client := kubeClient(in)
	URL := in.Master + "/healthz"
	secureCheck := strings.Split(URL, "https://")
	if len(secureCheck) == 1 {
		URL = "https://" + URL
	}
	request, _ := http.NewRequest("GET", URL, nil)
	if in.Basic {
		request.SetBasicAuth(in.UserName, in.Password)
	}

	go func(c *http.Client, r *http.Request, s time.Duration) {
		i := 1
		for {
			in.ClusterData.Job.Log.Println("Checking kubernetes api")
			if i > 1 {
				time.Sleep(s)
			}
			i++

			response, err := c.Do(r)
			if err != nil {
				in.ClusterData.Job.Log.Println("API not available yet: ", i, err)
				continue
			}

			r, err := ioutil.ReadAll(response.Body)
			if err != nil {
				in.ClusterData.Job.Log.Println("Unable to process result: ", i, err)
				continue
			}

			body := string(r)
			response.Body.Close()
			in.ClusterData.Job.Log.Printf("API_CHECK: url: %v code: %d, body: %s\n", URL, response.StatusCode, body)

			if response.StatusCode == 200 && body == "ok" {
				healthy <- true
				break
			}
		}
		close(healthy)
	}(client, request, in.SleepTime)
	return healthy

}

func kubeClient(in *K8APIStatusInput) *http.Client {

	var client http.Client
	if in.Basic {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		}}
	} else if !in.Basic {
		client.Transport = kube.Transport(in.Certs)
	}

	return &client

}

