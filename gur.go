package main

import (
	"bytes"
	"flag"
	"fmt"
	"json"
	"os"
	"strings"
	"tabwriter"
	timer "github.com/str1ngs/gotimer"
	"sync"
)

// Constants
const (
	program = "gur"
	version = "0.0.1"
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
	search   = flag.Bool("s", true, "search aur for packages")
	help     = flag.Bool("h", false, "displays usage")
	quiet    = flag.Bool("q", false, "only output package names")
	test     = flag.Bool("t", false, "run tests")
	download = flag.Bool("d", false, "download and extract tarball into working path")
	debug    = flag.Bool("dh", false, "debug http headers")
	//aur      *Aur
)

// Prints usage detains
func usage() {
	printDefaults()
	os.Exit(1)
}

// Program entry
func main() {
	defer timer.From(timer.Now())
	flag.Parse()
	flag.Usage = printDefaults
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *test {
		doTest()
		os.Exit(0)
	}
	if *download {
		if len(flag.Args()) == 0 {
			err := os.NewError("no packages specified")
			handleError(err)
		}
		loadSyncCache()
		readInstalled()
		*search = false
		checkDepends(flag.Arg(0))
		doDownload(flag.Arg(0))
		tw.Flush()
		return
	}
	if *search {
		doSearch()
		return
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
func doDownload(name string) {
	if fileExists(name) {
		//return
	}
	aur, _ := NewAur()
	reader, err := aur.Tarball(name)
	handleError(err)
	tar := NewTar()
	err = tar.Untar("./", reader)
	handleError(err)
	fmt.Fprintf(os.Stderr, "./%v\n", name)
}

func checkDepends(name string) {
	aur, _ := NewAur()
	b, err := aur.Pkgbuild(name)
	handleError(err)
	pb := bytes.NewBuffer(b)
	dbuf := new(bytes.Buffer)
	for _, v := range []string{"depends", "makedepends"} {
		dbuf.Write(parseBashArray(pb.Bytes(), v))
		dbuf.WriteString(" ")
	}
	wg := new(sync.WaitGroup)
	for _, b := range bytes.Split(dbuf.Bytes(), []byte(" "), -1) {
		if len(b) == 0 {
			continue
		}
		depend := strings.Trim(string(b), " ")
		repo, pr := whichRepo(depend)
		switch pr {
		case "":
			//fprintf(tw, "%s\t%s\t\n", depend, repo)
		default:
			//fprintf(tw, "%s\t%s\t(%s)\n", depend, repo, pr)
		}
		if repo == "aur" {
			wg.Add(1)
			go func() {
				checkDepends(depend)
				doDownload(depend)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

// Calls search rpc and prints results
func doSearch() {
	defer timer.From(timer.Now())
	if len(flag.Args()) == 0 {
		err := os.NewError("no packages specified")
		handleError(err)
	}
	arg := flag.Arg(0)
	aur, _ := NewAur()
	sr, err := aur.Results("search", arg)
	handleError(err)
	for _, i := range sr.RawResults {
		b, err := i.MarshalJSON()
		handleError(err)
		result := new(Result)
		err = json.Unmarshal(b, result)
		handleError(err)
		fprintln(tw, result)
	}
	tw.Flush()
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
