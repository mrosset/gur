package main

import (
	"crypto/tls"
	"http"
	"os"
)

const (
	rpc      = "rpc.php?type=%v&arg=%v"
	pkgbuild = "packages/%v/PKGBUILD"
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
	if *quiet {
		return sprintf("%v", r.Name)
	}
	if !*quiet {
		return sprintf("aur/%v %v (%v) \n  %v", r.Name, r.Version, r.NumVotes, r.Description)
	}
	return ""
}

type Aur struct {
	conn *http.ClientConn
	url  *http.URL
}

func NewAur(rawurl string) (*Aur, os.Error) {
	var (
		aur = new(Aur)
		err os.Error
	)
	if aur.url, err = http.ParseURL(rawurl); err != nil {
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

func (aur *Aur) GetPkgbuild(name string) (res *http.Response, err os.Error) {
	req, err := aur.Request("GET", sprintf(pkgbuild, name))
	if err != nil {
		return nil, err
	}
	res, err = aur.doRequest(req)
	return res, err
}

func (aur *Aur) GetTarBall(urlpath string) (res *http.Response, err os.Error) {
	req, err := aur.Request("GET", urlpath)
	if err != nil {
		return nil, err
	}
	res, err = aur.doRequest(req)
	return res, err
}

func (aur *Aur) Method(method, arg string) (res *http.Response, err os.Error) {
	req, err := aur.Request("GET", sprintf(rpc, method, arg))
	if err != nil {
		return nil, err
	}
	res, err = aur.doRequest(req)
	return res, err
}

func (aur *Aur) doRequest(req *http.Request) (res *http.Response, err os.Error) {
	if *debug {
		b, err := http.DumpRequest(req, true)
		if err != nil {
			return nil, err
		}
		os.Stderr.Write(b)
	}
	if res, err = aur.conn.Do(req); err != nil {
		if err != http.ErrPersistEOF {
			return nil, err
		}
		aur.connect()
		aur.conn.Do(req)
	}
	if *debug {
		b, err := http.DumpResponse(res, false)
		if err != nil {
			return nil, err
		}
		os.Stderr.Write(b)
	}
	return res, nil
}

func (aur *Aur) Request(method, rest string) (*http.Request, os.Error) {
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
	url := aur.url.String() + rest
	if req.URL, err = http.ParseURL(url); err != nil {
		return nil, err
	}

	return req, nil
}

func (aur *Aur) Close() {
	aur.conn.Close()
}
