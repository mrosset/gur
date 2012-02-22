package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	client = new(http.Client)
)

const (
	host      = "http://aur.archlinux.org:80/"
	rpc       = "%s/rpc.php?type=%s&arg=%s"
	pkgbuild  = "%s/packages/%s/%s/PKGBUILD"
	tarball   = "%s/packages/%s/%s/%s.tar.gz"
	userAgent = "curl/7.21.4 (x86_64-unknown-linux-gnu) libcurl/7.21.4 OpenSSL/1.0.0d zlib/1.2.5"
)

func init() {
	client = new(http.Client)
}

type SearchResults struct {
	Type       string
	RawResults []*json.RawMessage `json:"results"`
}

type Result struct {
	Id          string
	Name        string
	Version     string
	CategoryID  string
	Description string
	URL         string
	URLPath     string
	Licensce    string
	NumVotes    string
	OutOfDate   string
}

func (r Result) String() string {
	return fmt.Sprintf("%v/%v %v (%v) \n%v", "aur", r.Name, r.Version, r.NumVotes, r.Description)
}

type Aur struct {
}

func GetPkgBuild(name string) ([]byte, error) {
	res, err := client.Get(fmt.Sprintf(pkgbuild, host, name[0:2], name))
	err = checkRes(res, err)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GetTarball(name string) (io.Reader, error) {
	res, err := client.Get(fmt.Sprintf(tarball, host, name[0:2], name, name))
	err = checkRes(res, err)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func GetResults(method, arg string) (sr *SearchResults, err error) {
	res, err := client.Get(fmt.Sprintf(rpc, host, method, arg))
	err = checkRes(res, err)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	sr = new(SearchResults)
	err = json.Unmarshal(b, sr)
	if err != nil {
		return nil, err
	}
	return sr, err
}

func checkRes(res *http.Response, err error) error {
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Http GET failed for %s with status code %s", res.Request.URL, res.Status)
	}
	return nil
}

/*
func (aur *Aur) buildRequest(method, rest string) (*http.Request, error) {
	var (
		err error
	)
	req := new(http.Request)
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.TransferEncoding = []string{"chunked"}
	req.Header = http.Header{}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", userAgent)
	req.Method = method
	if req.URL, err = url.Parse(host + rest); err != nil {
		return nil, err
	}
	return req, nil
}
*/

/*
func readBody(res *http.Response) ([]byte, error) {
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Http GET failed for %s with status code %s", res.Request.URL, res.Status)
	}
	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, gz)
	return buf.Bytes(), err
}
*/
