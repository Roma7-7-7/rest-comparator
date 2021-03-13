package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/resty.v0"
	"gopkg.in/yaml.v2"
)

//HAHA

//Header : header
type Header struct {
	Key   string `yaml2:"key"`
	Value string `yaml2:"value"`
}

//Properties : properties
type Properties struct {
	Headers []Header `yaml2:"headers"`
}

//Endpoint : endpoint
type Endpoint struct {
	URL        string     `yaml2:"url"`
	Method     string     `yaml2:"method"`
	Properties Properties `yaml2:"properties"`
}

var endpointsPath = "endpoints"

func checkExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Println(err)
		log.Fatalf("Failed to locate endpoint file [%v]", path)
	}
}

func readEndpointFile(path string) Endpoint {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		log.Println(err)
		log.Fatalf("Failed to read file [%v]", path)
	}

	result := Endpoint{}
	err = yaml.Unmarshal(data, &result)

	if err != nil {
		log.Println(err)
		log.Fatalf("Failed to parse file [%v]", path)
	}

	return result
}

func setProperties(endpoint Endpoint, r *resty.Request) {
	if len(endpoint.Properties.Headers) > 0 {
		for _, h := range endpoint.Properties.Headers {
			r.SetHeader(h.Key, h.Value)
		}
	}
}

func call(endpoint Endpoint) *resty.Response {
	method := endpoint.Method
	if len(method) == 0 {
		method = "GET"
	} else {
		method = strings.ToUpper(method)
	}

	var resp *resty.Response
	var err error

	request := resty.R()

	setProperties(endpoint, request)

	switch method {
	case "GET":
		resp, err = request.Get(endpoint.URL)
		break
	case "POST":
		resp, err = request.Post(endpoint.URL)
		break
	case "PUT":
		resp, err = request.Put(endpoint.URL)
		break
	case "PATCH":
		resp, err = request.Patch(endpoint.URL)
		break
	case "DELETE":
		resp, err = request.Delete(endpoint.URL)
		break
	case "OPTIONS":
		resp, err = request.Options(endpoint.URL)
		break
	default:
		log.Fatalf("Unsupported method [%v] for endpoint [%v]", method, endpoint.URL)
	}

	if err != nil {
		log.Println(err)
		log.Fatalf("Failed to call endpoint [%v]", endpoint.URL)
	}

	return resp
}

func compare(endpoint1 Endpoint, response1 *resty.Response, endpoint2 Endpoint, response2 *resty.Response) {
	url1, url2 := endpoint1.URL, endpoint2.URL
	status1, status2 := response1.Status(), response2.Status()

	if status1 != status2 {
		log.Fatalf("Diferent statsues for endpoints [%v] and [%v]. Statuses: [%v], [%v]", url1, url2, status1, status2)
	}

	body1, body2 := response1.String(), response2.String()

	if body1 != body2 {
		log.Fatalf("Diferent bodies for endpoints [%v] and [%v]. Bodies: [%v], [%v]", url1, url2, body1, body2)
	}
}

func main() {
	e1Path, e2Path := filepath.Join(endpointsPath, "e1.yml"), filepath.Join(endpointsPath, "e2.yml")

	checkExist(e1Path)
	checkExist(e2Path)

	e1, e2 := readEndpointFile(e1Path), readEndpointFile(e2Path)
	r1, r2 := call(e1), call(e2)

	compare(e1, r1, e2, r2)

	log.Println("All calls are equal")
}
