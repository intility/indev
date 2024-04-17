package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestError struct {
	Message string
}

func (e *RequestError) Error() string {
	return e.Message
}

func doRequest[T any](client *http.Client, req *http.Request, result *T) error {
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not perform request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if !isSuccessfulStatusCode(resp.StatusCode) {
		// read body as plain text
		if resp.ContentLength > 0 {
			var body []byte

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("could not read response body: %w", err)
			}

			return &RequestError{
				Message: resp.Status + ": " + string(body),
			}
		}

		return &RequestError{
			Message: resp.Status,
		}
	}

	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("could not decode response: %w", err)
		}
	}

	return nil
}

func isSuccessfulStatusCode(status int) bool {
	return status >= 200 && status <= 299
}
