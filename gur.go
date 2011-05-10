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

// Global vars
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
	search    = flag.Bool("s", true, "search aur for packages")
	help      = flag.Bool("h", false, "displays usage")
	quiet     = flag.Bool("q", false, "only output package names")
	test      = flag.Bool("t", false, "run tests")
	download  = flag.Bool("d", false, "download and extract tarball into working path")
	debug     = flag.Bool("dh", false, "debug http headers")
	//aur       *Aur
)

// Prints usage detains
func usage() {
	printDefaults()
	os.Exit(1)
}

// Program entry
func main() {
	flag.Parse()
	flag.Usage = printDefaults
	var err os.Error
	aur, err := NewAur(host)
	handleError(err)
	defer aur.Close()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *test {
		doTest()
		os.Exit(0)
	}
	if *download {
		loadSyncCache()
		*search = false
		if len(flag.Args()) == 0 {
			err := os.NewError("no packages specified")
			handleError(err)
		}
		checkDepends(flag.Arg(0))
		doDownload(flag.Arg(0))
		os.Exit(0)
	}
	if *search {
		doSearch()
		os.Exit(0)
	}
	flag.Usage()
}

func printDefaults() {
	fprintf(tw, "Usage of %s:\n\n", os.Args[0])
	flag.VisitAll(vFlag)
	tw.Flush()
}

func vFlag(f *flag.Flag) {
	//format := "\t-%s=%s:\t%s\n"
	format := "\t-%s:\t%s\n"
	fprintf(tw, format, f.Name, f.Usage)
}

func doTest() {
}

//TODO: fix all the crazy err handling
func doDownload(arg string) {
	if fileExists(arg) {
		//return
	}
	buf, err := getResults("info", arg)
	handleError(err)
	err = checkInfoError(buf)
	if err != nil {
		return
	}
	info := new(Info)
	err = json.Unmarshal(buf, info)
	handleError(err)
	aur, _ := NewAur(host)
	res, err := aur.GetTarBall(info.Results.URLPath)
	handleError(err)
	zbuf := new(bytes.Buffer)
	io.Copy(zbuf, res.Body)
	zip := NewZip()
	gzip, err := gzip.NewReader(zbuf)
	handleError(err)
	err = zip.Decompress("./", gzip)
	handleError(err)
	printf("./%v\n", arg)
}

func checkDepends(name string) {
	aur, _ := NewAur(host)
	res, err := aur.GetPkgbuild(name)
	handleError(err)
	gzip, err := gzip.NewReader(res.Body)
	handleError(err)
	pb := new(bytes.Buffer)
	pb.ReadFrom(gzip)
	dbuf := new(bytes.Buffer)
	for _, v := range []string{"depends", "makedepends"} {
		dbuf.Write(parseBashArray(pb.Bytes(), v))
		dbuf.WriteString(" ")
	}
	count := 0
	c := make(chan int)
	for _, b := range bytes.Split(dbuf.Bytes(), []byte(" "), -1) {
		if len(b) == 0 {
			continue
		}
		depend := strings.Trim(string(b), " ")
		repo, pr := whichRepo(depend)
		switch pr {
		case "":
			fprintf(tw, "%s\t%s\t\n", depend, repo)
		default:
			fprintf(tw, "%s\t%s\t(%s)\n", depend, repo, pr)
		}
		if repo == "aur" {
			go func() {
				count++
				doDownload(depend)
				c <- 1
			}()
		}
	}
	for i := 0; i < count; i++ {
		<-c
	}
	tw.Flush()
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
		fprintln(tw, r.Format())
	}
	tw.Flush()
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
	aur, _ := NewAur(host)
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

func parseBashArray(pkgbuild []byte, bvar string) []byte {
	pkg := bytes.NewBuffer(pkgbuild)
	depends := new(bytes.Buffer)
	for {
		line, err := pkg.ReadBytes('\n')
		if err == os.EOF {
			break
		}
		if bytes.HasPrefix(line, []byte(bvar)) {
			// if line ends with ) then we have what we need
			if line[len(line)-2] == ')' {
				depends.Write(line[len(bvar)+2 : len(line)-2])
				break
			}
			depends.Write(line[len(bvar)+2:])
			// find end of array and write it to depends buffer
			rest, _ := pkg.ReadBytes(')')
			depends.Write(rest[0 : len(rest)-2])
		}
	}
	if len(depends.Bytes()) == 0 {
		return nil
	}
	b := depends.Bytes()
	depends.Reset()
	for _, d := range bytes.Split(b, []byte(" "), -1) {
		if len(d) == 0 {
			continue
		}
		d = bytes.Replace(d, []byte("'"), nil, -1)
		d = bytes.Replace(d, []byte("\n"), nil, -1)
		d = bytes.Replace(d, []byte(" "), nil, -1)
		switch {
		case strings.Contains(string(d), ">"):
			s := bytes.Split(d, []byte(">"), -1)
			depends.WriteString(string(s[0]) + " ")
		case strings.Contains(string(d), "<"):
			s := bytes.Split(d, []byte("<"), -1)
			depends.WriteString(string(s[0]) + " ")
		case strings.Contains(string(d), "="):
			s := bytes.Split(d, []byte("="), -1)
			depends.WriteString(string(s[0]) + " ")
		default:
			depends.WriteString(string(d) + " ")
		}
	}
	return depends.Bytes()[0 : len(depends.Bytes())-1]
}
