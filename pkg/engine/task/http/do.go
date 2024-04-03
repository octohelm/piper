package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/octohelm/x/ptr"

	"github.com/go-courier/logr"
	"github.com/octohelm/piper/pkg/cueflow"
	"github.com/octohelm/piper/pkg/engine/task/client"

	"github.com/octohelm/piper/pkg/engine/task"
	"github.com/octohelm/piper/pkg/engine/task/file"
)

func init() {
	cueflow.RegisterTask(task.Factory, &Do{})
}

// Do http request
type Do struct {
	task.Task

	// http method
	Method string `json:"method"`
	// http request url
	Url string `json:"url"`
	// http headers
	Header map[string]client.StringOrSlice `json:"header,omitempty"`
	// http query
	Query map[string]client.StringOrSlice `json:"query,omitempty"`
	// http request body
	RequestBody file.StringOrFile `json:"body,omitempty"`

	// options
	With DoOption `json:"with,omitempty"`

	// Response
	Response Response `json:"-" output:"response"`
}

type DoOption struct {
	// header keys for result
	ExposeHeaders []string `json:"exposeHeaders"`
}

type Response struct {
	// status code
	Status int `json:"status,omitempty"`
	// response header, only pick headers requests by `with.header`
	Header map[string]client.StringOrSlice `json:"header,omitempty"`
	// auto unmarshal based on content-type
	Data any `json:"data,omitempty"`
}

func (r *Do) Do(ctx context.Context) error {
	size, err := r.RequestBody.Size(ctx)
	if err != nil {
		return err
	}

	var reader io.Reader

	if size > 0 {
		pw := cueflow.NewProcessWriter(size)

		f, err := r.RequestBody.Open(ctx)
		if err != nil {
			return err
		}
		defer f.Close()

		_, l := logr.FromContext(ctx).Start(ctx, "uploading", slog.Int64(cueflow.LogAttrProgressTotal, size))
		defer l.End()

		go func() {
			for p := range pw.Process(ctx) {
				l.WithValues(slog.Int64(cueflow.LogAttrProgressCurrent, p.Current)).Info("")
			}
		}()

		reader = io.TeeReader(f, pw)
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, r.Url, reader)
	if err != nil {
		return err
	}

	if len(r.Query) > 0 {
		q := req.URL.Query()
		for k, vv := range r.Query {
			q[k] = vv
		}
		req.URL.RawQuery = q.Encode()
	}

	for k, vv := range r.Header {
		req.Header[k] = vv
	}

	if size > 0 {
		req.ContentLength = size
	}

	logr.FromContext(ctx).Info(fmt.Sprintf("%s %s", req.Method, req.URL.String()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	r.Ok = ptr.Ptr(resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices)
	r.Response.Status = resp.StatusCode

	if contentType := resp.Header.Get("Content-Type"); strings.Contains(contentType, "json") {
		a := &client.Any{}
		if err := json.NewDecoder(resp.Body).Decode(a); err != nil {
			return err
		}
		r.Response.Data = a.Value
	} else if strings.HasPrefix(contentType, "text/") {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		r.Response.Data = string(data)
	} else {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		r.Response.Data = data
	}

	r.Response.Header = map[string]client.StringOrSlice{}

	for _, k := range r.With.ExposeHeaders {
		r.Response.Header[k] = resp.Header.Values(k)
	}

	return nil
}
