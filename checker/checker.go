package checker

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)

type StatusCodeError struct {
	StatusCode int
	Err        error
}

func (s *StatusCodeError) Error() string {
	return fmt.Sprintf("%s: %d", s.Err, s.StatusCode)
}

type DestinationError struct {
	Destination *url.URL
	Err         error
}

func (s *DestinationError) Error() string {
	return fmt.Sprintf("%s: %v", s.Err, s.Destination)
}

func Check(source string, target string, statusCode int) (bool, error) {
	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrAbortHandler // don't follow redirects please
	}
	resp, err := client.Get(source)
	if err != nil && !errors.Is(err, http.ErrAbortHandler) {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != statusCode {
		return false, &StatusCodeError{StatusCode: resp.StatusCode, Err: errors.New("status code")}
	}
	location, err := resp.Location()
	if err != nil {
		return false, err
	}
	dest, err := url.Parse(target)
	if err != nil {
		return false, err
	}
	if !reflect.DeepEqual(location, dest) {
		return false, &DestinationError{Destination: location, Err: errors.New("destination")}
	}
	return true, nil
}
