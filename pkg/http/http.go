// This file is part of MinIO Operator
// Copyright (c) 2022 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package http

import (
	"io"
	"io/ioutil"
	"net/http"
)

// ClientI interface with all functions to be implemented
// by mock when testing, it should include all HttpClient respective api calls
// that are used within this project.
type ClientI interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

// Client is an HTTP Interface implementation
//
// Define the structure of a http client and define the functions that are actually used
type Client struct {
	Client *http.Client
}

// Get implements http.Client.Get()
func (c *Client) Get(url string) (resp *http.Response, err error) {
	return c.Client.Get(url)
}

// Post implements http.Client.Post()
func (c *Client) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return c.Client.Post(url, contentType, body)
}

// Do implement http.Client.Do()
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

// DrainBody close non nil response with any response Body.
// convenient wrapper to drain any remaining data on response body.
//
// Subsequently this allows golang http RoundTripper
// to re-use the same connection for future requests.
func DrainBody(respBody io.ReadCloser) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if respBody != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		//  - http://stackoverflow.com/a/17961593/4465767
		defer respBody.Close()
		io.Copy(ioutil.Discard, respBody)
	}
}
