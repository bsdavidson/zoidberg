package zoidberg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"testing"

	"net/http/httptest"

	"github.com/stretchr/testify/require"
)

// Zoidberg ...
type Zoidberg struct {
	w         io.WriteCloser
	ts        *httptest.Server
	t         *testing.T
	reqHeader string
}

// Request ...
type Request struct {
	Method       string
	Path         string
	Body         interface{}
	Description  string
	Write        bool
	ReponseCodes map[int]string
}

// NewZoidberg returns a new zoidberg instance
func NewZoidberg(w io.WriteCloser, ts *httptest.Server, t *testing.T, requestHeader string) *Zoidberg {
	return &Zoidberg{w: w, ts: ts, t: t, reqHeader: requestHeader}
}

// WoopWoopWoop ...
func (z *Zoidberg) WoopWoopWoop(t *testing.T, req *http.Request, reqBody interface{}, resp *http.Response, body []byte, description string, responses map[int]string) {
	query := ""
	if req.URL.RawQuery != "" {
		query = fmt.Sprintf("?%s", req.URL.RawQuery)
	}
	fmt.Fprintf(z.w, ".. http:%s:: %s%s\n\n", strings.ToLower(req.Method), req.URL.Path, query)
	fmt.Fprintf(z.w, "   %s\n\n", description)
	if responses != nil {
		responseCodes := []int{}
		for k := range responses {
			responseCodes = append(responseCodes, k)
		}
		sort.Ints(responseCodes)
		fmt.Fprintf(z.w, "     **Response Codes**\n\n")
		for _, code := range responseCodes {
			fmt.Fprintf(z.w, "     - %d: %s\n\n", code, responses[code])
		}
	}
	fmt.Fprintf(z.w, "\n\n")

	fmt.Fprintf(z.w, "   Example request:\n\n")
	fmt.Fprintf(z.w, "   .. sourcecode:: http\n\n")
	fmt.Fprintf(z.w, "      %s %s %s\n", req.Method, req.URL.Path, req.Proto)
	for k := range req.Header {
		fmt.Fprintf(z.w, "      %s: %s\n", k, req.Header.Get(k))
	}

	if reqBody != nil {
		b, err := json.MarshalIndent(reqBody, "      ", "  ")
		require.NoError(t, err)
		fmt.Fprintf(z.w, "\n")
		fmt.Fprintf(z.w, "      %s\n\n", b)
	}

	fmt.Fprintf(z.w, "\n")
	fmt.Fprintf(z.w, "   Example response:\n\n")
	fmt.Fprintf(z.w, "   .. sourcecode:: http\n\n")
	fmt.Fprintf(z.w, "      %s %s\n", resp.Proto, resp.Status)
	for k := range resp.Header {
		fmt.Fprintf(z.w, "      %s: %s\n", k, resp.Header.Get(k))
	}
	fmt.Fprintf(z.w, "\n")

	var jb interface{}
	if len(body) > 0 {
		require.NoError(t, json.Unmarshal(body, &jb))
		b, err := json.MarshalIndent(jb, "      ", "  ")
		require.NoError(t, err)
		fmt.Fprintf(z.w, "      %s\n\n", b)
	}

}

// WhyNot Creates a header section
func (z *Zoidberg) WhyNot(title, underline string) {
	fmt.Fprintf(z.w, "%s\n", title)
	fmt.Fprintf(z.w, "%s\n\n", strings.Repeat(underline, len(title)))
}

// Says outputs some paragraph text.
func (z *Zoidberg) Says(text string) {
	fmt.Fprintf(z.w, "  %s\n\n", text)
}

// Ask is a helper function that takes a number of parmeters, makes a request, and asserts the response
func (z *Zoidberg) Ask(r Request) {
	// t.Log("TestRequest", method, path, body)
	var bodyReader io.Reader
	if r.Body != nil {
		bodyBytes, err := json.Marshal(r.Body)
		require.NoError(z.t, err)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", z.ts.URL, r.Path), bodyReader)
	req.Header.Set("Content-Type", z.reqHeader)
	require.NoError(z.t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(z.t, err)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(z.t, err)

	if r.Write {
		z.WoopWoopWoop(z.t, req, r.Body, resp, b, r.Description, r.ReponseCodes)
	}
}
