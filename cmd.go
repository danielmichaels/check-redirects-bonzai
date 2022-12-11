package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	Z "github.com/rwxrob/bonzai/z"
	"github.com/rwxrob/help"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	reqTimeout = 6 * time.Second
	url        = "https://check-redirects.infra.ptco.rocks/api/search/"
)

var Cmd = &Z.Cmd{
	Name:     `check-redirects`,
	Aliases:  []string{"cr"},
	MinArgs:  1,
	MaxArgs:  2,
	Summary:  `check redirect path of given domain and return final destination`,
	Commands: []*Z.Cmd{help.Cmd},
	Usage:    `ds scripts check-redirects`,
	Description: `
		Follow the redirect chain for a given domain.
		
		This command will only return the <scheme>://<host>/<path> for the **final** destination
		of any redirect chain, or an error.
`,
	Call: func(_ *Z.Cmd, args ...string) error {
		fmt.Printf("Running 'check-redirects.com' API engine for %q\n\n", args[0])
		// check args
		userAgent := "chrome"
		if len(args) == 2 {
			userAgent = args[1]
		}
		cl := http.Client{Timeout: reqTimeout}

		body := struct {
			Domain    string `json:"domain"`
			UserAgent string `json:"user_agent"`
		}{
			Domain:    args[0],
			UserAgent: userAgent,
		}

		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "check-redirects-bonzai")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, */*")

		response, err := cl.Do(req)
		if err != nil {
			return err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println("failed to close connection")
			}
		}(response.Body)

		r, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		var result ResponseTypes
		err = json.Unmarshal(r, &result)
		if err != nil {
			return err
		}
		if result.ErrorResponse.Detail.Error != "" {
			fmt.Printf("Error:\t%s\nUrl:\t%s\nUser-Agent:\t%s\n",
				result.ErrorResponse.Detail.Error,
				result.ErrorResponse.Detail.Url,
				result.ErrorResponse.Detail.UserAgent,
			)
			return nil
		}
		last := result.Response[len(result.Response)-1]
		fmt.Printf("Last Redirect Found: %s://%s%s\n", last.Scheme, last.Host, last.Path)
		return nil
	},
}

type ResponseTypes struct {
	Response      []RedirectResponse
	ErrorResponse ErrResponse
}

func (d *ResponseTypes) UnmarshalJSON(data []byte) error {
	var m map[string]any
	var unmarshalErr error

	if err := json.Unmarshal(data, &m); err != nil {
		unmarshalErr = err
	}

	if _, ok := m["detail"]; ok {
		var errData ErrResponse
		err := json.Unmarshal(data, &errData)
		if err != nil {
			return err
		}
		d.ErrorResponse = errData
		return nil
	}

	var arr []map[string]any
	if err := json.Unmarshal(data, &arr); err != nil {
		unmarshalErr = err
	}

	if _, ok := arr[0]["status_code"]; ok {
		var respData []RedirectResponse
		err := json.Unmarshal(data, &respData)
		if err != nil {
			return err
		}
		d.Response = respData
		return nil
	}
	return unmarshalErr
}

type ErrResponse struct {
	Detail struct {
		Error     string `json:"error"`
		Url       string `json:"url"`
		UserAgent string `json:"user-agent"`
	} `json:"detail"`
}
type RedirectResponse struct {
	Id          int    `json:"id,omitempty"`
	Hop         int    `json:"hop,omitempty"`
	Url         string `json:"url,omitempty"`
	HttpVersion string `json:"http_version,omitempty"`
	StatusCode  struct {
		Code   string `json:"code"`
		Phrase string `json:"phrase"`
	} `json:"status_code,omitempty"`
	Headers struct {
		Location                string `json:"location,omitempty"`
		Server                  string `json:"server"`
		XNfRequestId            string `json:"x-nf-request-id"`
		Date                    string `json:"date"`
		ContentLength           string `json:"content-length,omitempty"`
		ContentType             string `json:"content-type"`
		Age                     string `json:"age,omitempty"`
		CacheControl            string `json:"cache-control,omitempty"`
		ContentEncoding         string `json:"content-encoding,omitempty"`
		Etag                    string `json:"etag,omitempty"`
		PermissionsPolicy       string `json:"permissions-policy,omitempty"`
		StrictTransportSecurity string `json:"strict-transport-security,omitempty"`
		Vary                    string `json:"vary,omitempty"`
		TransferEncoding        string `json:"transfer-encoding,omitempty"`
	} `json:"headers,omitempty"`
	Host        string `json:"host,omitempty"`
	Path        string `json:"path,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	Ipaddr      string `json:"ipaddr,omitempty"`
	TimeElapsed int    `json:"time_elapsed,omitempty"`
	Body        string `json:"body,omitempty"`
	Ipinfo      struct {
		Ip          string `json:"ip"`
		Hostname    string `json:"hostname"`
		City        string `json:"city"`
		Region      string `json:"region"`
		Country     string `json:"country"`
		CountryName string `json:"country_name"`
		Latitude    string `json:"latitude"`
		Longitude   string `json:"longitude"`
		Org         string `json:"org"`
		Postal      string `json:"postal"`
		Timezone    string `json:"timezone"`
		Anycast     bool   `json:"anycast"`
	} `json:"ipinfo,omitempty"`
}
