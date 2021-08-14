package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-yaml/yaml"
)

const (
	casesDir = "."
	tempdir  = "./temp"
	tool     = "../bin/scrapen"
	wsAddr   = "127.0.0.1:8811"
)

func main() {
	log.Print("Run integration tests")

	stop := make(chan interface{}, 1)
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt)
	go func() {
		// wait for SIGINT, then tell server to stop
		<-sig
		stop <- true
	}()

	go webserver(casesDir, wsAddr, stop)

	errors := make([]error, 0)
	var errMx sync.Mutex
	var wg sync.WaitGroup

	err := filepath.Walk(casesDir, func(path string, info os.FileInfo, err error) error {
		wg.Add(1)
		go func() {
			e := runCase(path, info, err)
			if e != nil {
				errMx.Lock()
				errors = append(errors, e)
				errMx.Unlock()
			}
			wg.Done()
		}()
		return nil
	})

	if err != nil {
		log.Fatalf("Integration tests failed: %v", err)
	}

	wg.Wait()

	errMx.Lock()
	defer errMx.Unlock()
	if len(errors) != 0 {
		for n, e := range errors {
			log.Printf("%v) %v", n+1, e)
		}
		log.Fatal("Integration tests failed")
	}
}

func webserver(dir, addr string, stop chan interface{}) {
	http.Handle("/", http.FileServer(http.Dir(dir)))

	log.Printf("Starting webserver, dir=%q, addr=%q...", dir, addr)
	go http.ListenAndServe(addr, nil)

	// block until stopped
	<-stop
	log.Println("Stopped webserver...")
}

func runCase(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	if ext != ".yaml" {
		return nil
	}

	base := strings.TrimSuffix(path, ext)
	html := base + ".html"

	// check if the corresponding HTML file exists
	_, err = os.Stat(html)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	p := params{}
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&p)
	if err != nil {
		return err
	}

	log.Printf("Check %q", base)
	err = check(html, p)

	if err != nil {
		return fmt.Errorf("case %q, %v", base, err)
	}

	return nil
}

func check(html string, p params) error {
	url := "http://" + wsAddr + "/" + html
	base := strings.TrimSuffix(filepath.Base(html), filepath.Ext(html))
	outfile := filepath.Join(tempdir, base+".output")

	cmd := exec.Command(tool, url, outfile)
	output, err := cmd.Output()
	if err != nil {
		log.Print("scrapen output:")
		log.Print(string(output))
		return err
	}

	s, d, err := readResult(outfile)
	if err != nil {
		return err
	}

	err = find(s, p)
	if err != nil {
		return err
	}

	err = findNot(s, p)
	if err != nil {
		return err
	}

	err = queryNodes(d, p)
	if err != nil {
		return err
	}

	return nil
}

func find(s string, p params) error {
	for _, q := range p.Find {
		if !strings.Contains(s, q) {
			return fmt.Errorf("Missing substring %q", q)
		}
	}
	return nil
}

func findNot(s string, p params) error {
	for _, q := range p.FindNot {
		if strings.Contains(s, q) {
			return fmt.Errorf("Unexpected content %q", q)
		}
	}
	return nil
}

func queryNodes(d *goquery.Document, p params) error {
	var err error
	for _, q := range p.Query {
		if q.T == "" {
			found := d.Find(q.Q).Length()
			if found != q.N {
				return fmt.Errorf("Query %q should return %v elements, found %v", q.Q, q.N, found)
			}
		} else {
			count := 0
			d.Find(q.Q).Each(func(i int, s *goquery.Selection) {
				if s.Text() == q.T {
					count++
				}
			})
			if count != q.N {
				return fmt.Errorf("Query %q should return %v elements with text %q, found %v", q.Q, q.N, q.T, count)
			}
		}

	}
	return err
}

func readResult(p string) (string, *goquery.Document, error) {
	data, err := os.ReadFile(p + ".html")
	if err != nil {
		return "", nil, err
	}

	d, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return "", nil, err
	}

	return string(data), d, nil
}

type params struct {
	URL     string
	Find    []string
	FindNot []string
	Query   []query
}

type query struct {
	Q string
	T string
	N int
}
