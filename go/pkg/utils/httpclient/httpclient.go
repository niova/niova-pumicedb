package httpclient

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func service_Request(request *http.Request) ([]byte, error) {

	request.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}

	response, err := httpClient.Do(request)
	if err != nil {
		log.Error("(HTTP CLIENT DO)", err)
		return nil, err
	}

	switch response.StatusCode {
	case 200:
		//Serviced
		defer response.Body.Close()
		return ioutil.ReadAll(response.Body)
	case 503:
		//Service not found, returned for timeout
		return nil, errors.New("Server timed out")
	}
	return nil, nil
}

func HTTP_Request(requestBody []byte, address string, put bool) ([]byte, error) {
	var request *http.Request
	var err error

	connectionAddress := "http://" + address
	if put {
		request, err = http.NewRequest(http.MethodPut, connectionAddress, bytes.NewBuffer(requestBody))
	} else {
		request, err = http.NewRequest(http.MethodGet, connectionAddress, bytes.NewBuffer(requestBody))
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return service_Request(request)
}

// REST_Request issues an HTTP request with an explicit method, optional headers
// and request body, returning the response body together with the HTTP status
// code. Unlike HTTP_Request it makes no assumption about a CPResp envelope and
// surfaces every status code (including 4xx/5xx) to the caller so that REST
// error bodies are preserved instead of being swallowed.
func REST_Request(method, address string, body []byte, headers map[string]string) ([]byte, int, error) {
	connectionAddress := "http://" + address
	request, err := http.NewRequest(method, connectionAddress, bytes.NewBuffer(body))
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		log.Error("(HTTP CLIENT DO)", err)
		return nil, 0, err
	}
	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, response.StatusCode, err
	}
	return respBody, response.StatusCode, nil
}
