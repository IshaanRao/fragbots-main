package main

import (
	"io"
	"net/http"
)

var httpClient = http.Client{}

// get Sends get requests easily
func get(url string, headers *http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for key, vals := range *headers {
			for _, val := range vals {
				headers.Add(key, val)
			}

		}
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// post sends post requests easily
func post(url string, headers *http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for key, vals := range *headers {
			for _, val := range vals {
				headers.Add(key, val)
				println(key, ":", val)
			}

		}
	}
	w, _ := io.ReadAll(req.Body)
	println(string(w))

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	we, _ := io.ReadAll(response.Body)
	println(string(we))
	return response, nil
}
