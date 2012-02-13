package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"json"
	"os"
	"strings"
	"tabwriter"
	"sync"
)

// Constants
const (
	program = "gur"
	version = "0.0.1"
)

// Global vars
var (
	printf     = fmt.Printf
	println    = fmt.Println
	sprintf    = fmt.Sprintf
	fprintln   = fmt.Fprintln
	fprintf    = fmt.Fprintf
	tw         = tabwriter.NewWriter(os.Stderr, 1, 4, 1, ' ', 0)
	bufout     = bufio.NewWriter(os.Stdout)
	isSearch   = flag.Bool("s", true, "search aur for packages")
	isHelp     = flag.Bool("h", false, "displays usage")
	isQuiet    = flag.Bool("q", false, "only output package names")
	isTest     = flag.Bool("t", false, "run tests")
	isForce    = flag.Bool("f", false, "force overwrite")
	isDownload = flag.Bool("d", false, "download and extract tarball into working path")
	visited    = map[string]bool{}
	//aur      *Aur
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
	if *isHelp {
		flag.Usage()
		os.Exit(0)
	}
	if *isTest {
		test()
		return
	}
	if *isDownload {
		if len(flag.Args()) == 0 {
			err := os.NewError("no packages specified")
			handleError(err)
		}
		readCache()
		checkDepends(flag.Arg(0))
		download(flag.Arg(0))
		return
	}
	if *isSearch {
		search()
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

func test() {
}

//TODO: fix all the crazy err handling
func download(name string) {
	if fileExists(name) && !*isForce {
		fmt.Println(name, "exists", "skipping")
		return
	}
	aur, _ := NewAur()
	reader, err := aur.Tarball(name)
	if err != nil {
		return
	}
	tar := NewTar()
	err = tar.Untar("./", reader)
	handleError(err)
	os.Stderr.WriteString(sprintf("./%v\n", name))
}

func checkDepends(name string) os.Error {
	aur, _ := NewAur()
	b, err := aur.Pkgbuild(name)
	if err != nil {
		return err
	}
	pb := bytes.NewBuffer(b)
	dbuf := new(bytes.Buffer)
	for _, v := range []string{"depends", "makedepends"} {
		dbuf.Write(parseBashArray(pb.Bytes(), v))
		dbuf.WriteString(" ")
	}
	wg := new(sync.WaitGroup)
	for _, b := range bytes.Split(dbuf.Bytes(), []byte(" ")) {
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
				_, ok := visited[depend]
				if !ok {
					visited[depend] = true
					checkDepends(depend)
					download(depend)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	return nil
}

// Calls search rpc and prints results
func search() {
	if len(flag.Args()) == 0 {
		err := os.NewError("no packages specified")
		handleError(err)
	}
	arg := flag.Arg(0)
	aur, err := NewAur()
	handleError(err)
	sr, err := aur.Results("search", arg)
	handleError(err)
	for _, i := range sr.RawResults {
		b, err := i.MarshalJSON()
		handleError(err)
		result := new(Result)
		err = json.Unmarshal(b, result)
		handleError(err)
		bufout.WriteString(result.String() + "\n")
	}
	bufout.Flush()
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
	for _, d := range bytes.Split(b, []byte(" ")) {
		if len(d) == 0 {
			continue
		}
		d = bytes.Replace(d, []byte("'"), nil, -1)
		d = bytes.Replace(d, []byte("\n"), nil, -1)
		d = bytes.Replace(d, []byte(" "), nil, -1)
		switch {
		case strings.Contains(string(d), ">"):
			s := bytes.Split(d, []byte(">"))
			depends.WriteString(string(s[0]) + " ")
		case strings.Contains(string(d), "<"):
			s := bytes.Split(d, []byte("<"))
			depends.WriteString(string(s[0]) + " ")
		case strings.Contains(string(d), "="):
			s := bytes.Split(d, []byte("="))
			depends.WriteString(string(s[0]) + " ")
		default:
			depends.WriteString(string(d) + " ")
		}
	}
	return depends.Bytes()[0 : len(depends.Bytes())-1]
}
