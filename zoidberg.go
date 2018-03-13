package zoidberg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Zoidberg encapsulates the testing environment and provides access to
// the resources needed to read and parse the API calls.
type Zoidberg struct {
	w          io.WriteCloser
	ts         *httptest.Server
	t          *testing.T
	reqHeaders map[string]string
}

// A Request contains the possible parameters used in making an API call.
type Request struct {
	Method      string
	Path        string
	RequestPath string

	Body                interface{}
	RequestHeaders      map[string]string
	BasicAuthLogin      [2]string
	Description         string
	Write               bool
	ResponseCodes       map[int]string
	ResponseJSONObjects map[string]string
	ParameterValues     map[string]string
}

// NewZoidberg returns a new zoidberg instance
func NewZoidberg(w io.WriteCloser, ts *httptest.Server, t *testing.T, requestHeaders map[string]string) *Zoidberg {
	return &Zoidberg{w: w, ts: ts, t: t, reqHeaders: requestHeaders}
}

// Head Creates a header section
func (z *Zoidberg) Head(title, underline string) {
	fmt.Fprintf(z.w, "%s\n", title)
	fmt.Fprintf(z.w, "%s\n\n", strings.Repeat(underline, len(title)))
}

// Says outputs some paragraph text.
func (z *Zoidberg) Says(text string) {
	fmt.Fprintf(z.w, "  %s\n\n", text)
}

// Ask is a helper function that takes a Request, executes and asserts the response
func (z *Zoidberg) Ask(r Request) {
	// t.Log("TestRequest", method, path, body)
	var bodyReader io.Reader
	if r.Body != nil {
		bodyBytes, err := json.Marshal(r.Body)
		require.NoError(z.t, err)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	if r.RequestPath == "" {
		r.RequestPath = r.Path
	}

	req, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", z.ts.URL, r.RequestPath), bodyReader)
	for k, v := range z.reqHeaders {
		req.Header.Set(k, v)
	}

	req.SetBasicAuth(r.BasicAuthLogin[0], r.BasicAuthLogin[1])
	require.NoError(z.t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(z.t, err)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(z.t, err)

	if r.Write {
		z.getIt(z.t, req, r.Body, resp, b, r)
	}
}

// getIt executes the request
func (z *Zoidberg) getIt(t *testing.T, req *http.Request, reqBody interface{}, resp *http.Response, body []byte, r Request) {
	query := ""
	if req.URL.RawQuery != "" {
		query = fmt.Sprintf("?%s", req.URL.RawQuery)
	}
	fmt.Fprintf(z.w, ".. http:%s:: %s\n\n", strings.ToLower(req.Method), req.URL.Path)
	fmt.Fprintf(z.w, "   %s\n\n", r.Description)

	// Write in the response codes
	if r.ResponseCodes != nil {
		responseCodesOrdered := []int{}
		for k := range r.ResponseCodes {
			responseCodesOrdered = append(responseCodesOrdered, k)
		}
		sort.Ints(responseCodesOrdered)
		fmt.Fprintf(z.w, "     **Response Code**\n\n")
		for _, code := range responseCodesOrdered {
			fmt.Fprintf(z.w, "     - %d: %s\n\n", code, r.ResponseCodes[code])
		}
	}
	fmt.Fprintf(z.w, "\n\n")

	// Write in the parameters
	if r.ParameterValues != nil {
		parameterValuesOrdered := []string{}
		for k := range r.ParameterValues {
			parameterValuesOrdered = append(parameterValuesOrdered, k)
		}
		sort.Strings(parameterValuesOrdered)
		fmt.Fprintf(z.w, "     **Query Parameters**\n\n")
		for _, param := range parameterValuesOrdered {
			fmt.Fprintf(z.w, "     - **%s**: %s\n\n", param, r.ParameterValues[param])
		}
	}
	fmt.Fprintf(z.w, "\n\n")

	// Write in the response codes
	if r.ResponseJSONObjects != nil {
		responseJSONObjectsOrdered := []string{}
		for k := range r.ResponseJSONObjects {
			responseJSONObjectsOrdered = append(responseJSONObjectsOrdered, k)
		}
		sort.Strings(responseJSONObjectsOrdered)
		fmt.Fprintf(z.w, "     **Response JSON Object**\n\n")
		for _, code := range responseJSONObjectsOrdered {
			fmt.Fprintf(z.w, "     - **%s**: %s\n\n", code, r.ResponseJSONObjects[code])
		}
	}
	fmt.Fprintf(z.w, "\n\n")

	fmt.Fprintf(z.w, "   Example request:\n\n")
	fmt.Fprintf(z.w, "   .. sourcecode:: http\n\n")
	fmt.Fprintf(z.w, "      %s %s%s %s\n", req.Method, req.URL.Path, query, req.Proto)
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
