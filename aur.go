package aur

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"http"
	"io"
	"os"
)

const (
	rpc      = "rpc.php?type=%s&arg=%s"
	pkgbuild = "packages/%s/PKGBUILD"
	tarball  = "packages/%s/%s.tar.gz"
	host     = "http://aur.archlinux.org:443/"
)

type SearchResults struct {
	Type    string
	Results []Result
}

type Error struct {
	Type    string
	Results string
}

type Info struct {
	Type    string
	Results Result
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

func (r Result) Format() string {
	return fmt.Sprintf("aur/%v %v (%v) \n  %v", r.Name, r.Version, r.NumVotes, r.Description)
	return ""
}

type Aur struct {
	conn *http.ClientConn
	url  *http.URL
}

func NewAur() (*Aur, os.Error) {
	var (
		aur = new(Aur)
		err os.Error
	)
	if aur.url, err = http.ParseURL(host); err != nil {
		return nil, err
	}
	aur.connect()
	return aur, nil
}

func (aur *Aur) connect() os.Error {
	tcpConn, err := tls.Dial("tcp", aur.url.Host, nil)
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

func (aur *Aur) Method(method, arg string) (res *http.Response, err os.Error) {
	req, err := aur.buildRequest("GET", fmt.Sprintf(rpc, method, arg))
	if err != nil {
		return nil, err
	}
	res, err = aur.doRequest(req)
	return res, err
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
	//req.TransferEncoding = []string{"chunked"}
	req.Header = http.Header{}
	req.Header.Set("Accept-Encoding", "gzip,deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Method = method
	//req.UserAgent = userAgent
	url := aur.url.String() + rest
	if req.URL, err = http.ParseURL(url); err != nil {
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
