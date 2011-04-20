package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"json"
	"os"
	"strings"
	"tabwriter"
)

// Constants
const (
	program = "gur"
	version = "0.0.1"
	host    = "https://aur.archlinux.org:443/"
	//rawurl = "https://localhost:80/"
)

// Globa vars
var (
	printf   = fmt.Printf
	println  = fmt.Println
	sprintf  = fmt.Sprintf
	fprintln = fmt.Fprintln
	fprintf  = fmt.Fprintf
	tw       = tabwriter.NewWriter(os.Stderr, 1, 4, 1, ' ', 0)
	// FIXME: change to final program name when decided. Use this so as not to give wrong userAgent
	//userAgent = sprintf("%v/%v", program, version)
	userAgent = "curl/7.21.4 (x86_64-unknown-linux-gnu) libcurl/7.21.4 OpenSSL/1.0.0d zlib/1.2.5"
	help      = flag.Bool("h", false, "displays usage")
	quiet     = flag.Bool("q", false, "only output package names")
	test      = flag.Bool("t", false, "run tests")
	search    = flag.Bool("v", true, "search aur for packages")
	download  = flag.Bool("d", false, "download and extract tarball into working path")
	debug     = flag.Bool("dh", false, "debug http headers")
	dumpjson  = flag.Bool("dj", false, "dump json to stderr")
	aur       *Aur
)

// Prints usage detains
func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}

// Program entry
func main() {
	flag.Parse()
	var err os.Error
	aur, err = NewAur(host)
	handleError(err)
	defer aur.Close()
	if *test {
		doTest()
		os.Exit(0)
	}
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *download {
		*search = false
		if len(flag.Args()) == 0 {
			err := os.NewError("no packages specified")
			handleError(err)
		}
		doDownload(flag.Arg(0))
		checkDepends(flag.Arg(0))
		os.Exit(0)
	}
	if *search {
		doSearch()
		os.Exit(0)
	}
	flag.Usage()
}

func doTest() {
	printf("%v\n", isInstalled("xorg-server"))
}

//TODO: fix all the crazy err handling
func doDownload(arg string) {
	buf, err := getResults("info", arg)
	handleError(err)
	err = checkInfoError(buf)
	if err != nil {
		return
	}
	info := new(Info)
	err = json.Unmarshal(buf, info)
	handleError(err)
	res, err := aur.GetTarBall(info.Results.URLPath)
	handleError(err)
	zbuf := new(bytes.Buffer)
	io.Copy(zbuf, res.Body)
	zip := NewZip()
	gzip, err := gzip.NewReader(zbuf)
	handleError(err)
	err = zip.Decompress("./", gzip)
	handleError(err)
	zbuf.Reset()
	printf("./%v\n", arg)
}

func checkDepends(name string) {
	res, err := aur.GetPkgbuild(name)
	handleError(err)
	gzip, err := gzip.NewReader(res.Body)
	handleError(err)
	pb := new(bytes.Buffer)
	pb.ReadFrom(gzip)
	depends := new(bytes.Buffer)
	for _, v := range [][]byte{[]byte("depends"), []byte("makedepends")} {
		depends.Write(parseBashArray(pb.Bytes(), v))
		depends.WriteString(" ")
	}
	for {
		b, err := depends.ReadBytes(' ')
		if err == os.EOF {
			break
		}
		depend := strings.Trim(string(b), " ")
		doDownload(depend)
	}
}

// Calls search rpc and prints results
func doSearch() {
	if len(flag.Args()) == 0 {
		err := os.NewError("no packages specified")
		handleError(err)
	}
	arg := flag.Arg(0)
	sr := new(SearchResults)
	buf, err := getResults("search", arg)
	handleError(err)
	err = checkInfoError(buf)
	handleError(err)
	err = json.Unmarshal(buf, sr)
	handleError(err)
	for _, r := range sr.Results {
		println(r.Format())
	}
}

// Checks if rpc returned a error object
func checkInfoError(buf []byte) os.Error {
	info := new(Info)
	json.Unmarshal(buf, info)
	if info.Type == "error" {
		je := new(Error)
		err := json.Unmarshal(buf, je)
		if err != nil {
			return err
		}
		err = os.NewError(sprintf("gur: json %v", je.Results))
		return err
	}
	return nil
}

// Generic call to rpc methods
func getResults(method, arg string) ([]byte, os.Error) {
	buf := new(bytes.Buffer)
	res, err := aur.Method(method, arg)
	if err != nil {
		return nil, err
	}
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		zr, err := gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(buf, zr)
		handleError(err)
	default:
		_, err := io.Copy(buf, res.Body)
		handleError(err)
	}
	return buf.Bytes(), nil
}

func handleError(err os.Error) {
	if err != nil {
		printf("%v\n", err.String())
		os.Exit(1)
	}
}

func parseBashArray(pkgbuild, bvar []byte) []byte {
	pkg := bytes.NewBuffer(pkgbuild)
	depends := new(bytes.Buffer)
	defer pkg.Reset()
	defer depends.Reset()
	for {
		line, err := pkg.ReadBytes('\n')
		if err == os.EOF {
			break
		}
		if bytes.HasPrefix(line, bvar) {
			// write depend line but remove the var name
			depends.Write(bytes.Replace(line, bvar, nil, 1))
			// if line ends with ) then we have what we need
			if line[len(line)-2] == ')' {
				break
			}
			// find end of array and write it to depends buffer
			rest, _ := pkg.ReadBytes(')')
			depends.Write(rest)
		}
	}
	if len(depends.Bytes()) == 0 {
		return nil
	}
	b := depends.Bytes()
	depends.Reset()
	// loop though bytes and remove unwanted spaces and characters, then write to depends buffer
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case ' ':
			if b[i+1] != ' ' && b[i-1] != ' ' {
				depends.WriteByte(b[i])
			}
		case '\'', '(', ')', '=', '<', '>', '.', '\n', '\t', '\\':
		default:
			depends.WriteByte(b[i])
		}
	}
	return depends.Bytes()
}
