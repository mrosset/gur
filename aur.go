package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"http"
	"io"
	"json"
	"os"
	"sync"
	"net"
)

const (
	rpc       = "rpc.php?type=%s&arg=%s"
	pkgbuild  = "packages/%s/PKGBUILD"
	tarball   = "packages/%s/%s.tar.gz"
	host      = "http://aur.archlinux.org:80/"
	userAgent = "curl/7.21.4 (x86_64-unknown-linux-gnu) libcurl/7.21.4 OpenSSL/1.0.0d zlib/1.2.5"
)

type SearchResults struct {
	Type       string
	RawResults []*json.RawMessage "results"
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
	conn *http.ClientConn
	sync.RWMutex
}

func NewAur() (*Aur, os.Error) {
	var (
		aur = new(Aur)
	)
	err := aur.connect()
	if err != nil {
		return nil, err
	}
	return aur, nil
}

func (aur *Aur) connect() os.Error {
	url, err := http.ParseURL(host)
	if err != nil {
		return err
	}
	tcpConn, err := net.Dial("tcp", url.Host)
	if err != nil {
		return err
	}
	aur.conn = http.NewClientConn(tcpConn, nil)
	return nil
}

func (aur *Aur) Pkgbuild(name string) ([]byte, os.Error) {
	req, err := aur.buildRequest("GET", fmt.Sprintf(pkgbuild, name))
	if err != nil {
		return nil, err
	}
	res, err := aur.doRequest(req)
	if err != nil {
		return nil, err
	}
	b, err := readBody(res)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (aur *Aur) Tarball(name string) (io.Reader, os.Error) {
	req, err := aur.buildRequest("GET", fmt.Sprintf(tarball, name, name))
	if err != nil {
		return nil, err
	}
	res, err := aur.doRequest(req)
	if err != nil {
		return nil, err
	}
	b, err := readBody(res)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func (aur *Aur) Results(method, arg string) (sr *SearchResults, err os.Error) {
	req, err := aur.buildRequest("GET", fmt.Sprintf(rpc, method, arg))
	if err != nil {
		return nil, err
	}
	res, err := aur.doRequest(req)
	b, err := readBody(res)
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

func (aur *Aur) doRequest(req *http.Request) (res *http.Response, err os.Error) {
	if res, err = aur.conn.Do(req); err != nil {
		if err != http.ErrPersistEOF {
			return nil, err
		}
		aur.connect()
		aur.conn.Do(req)
	}
	return res, nil
}

func (aur *Aur) buildRequest(method, rest string) (*http.Request, os.Error) {
	var (
		err os.Error
	)
	req := new(http.Request)
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.TransferEncoding = []string{"chunked"}
	req.Header = http.Header{}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Method = method
	req.UserAgent = userAgent
	if req.URL, err = http.ParseURL(host + rest); err != nil {
		return nil, err
	}
	return req, nil
}

func readBody(res *http.Response) ([]byte, os.Error) {
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, os.NewError(fmt.Sprintf("Http GET failed for %s with status code %s", res.Request.URL, res.Status))
	}
	gz, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, gz)
	return buf.Bytes(), err
}

func (aur *Aur) Close() {
	aur.conn.Close()
}
