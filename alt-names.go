package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sync"
)

type Params struct {
	fn   string
	nGor int
}

func getParams(args []string) Params {

	if len(args) < 3 {
		fmt.Printf("Usage %s -f <filename> -t <threads>\n", args[0])
		os.Exit(1)
	}

	par := Params{}
	flag.StringVar(&par.fn, "f", "", "File with hosts in format <host>:<port>. One per line")
	flag.IntVar(&par.nGor, "t", 10, "Number of paralell goroutines")
	flag.Parse()
	return par
}

func getCertData(theChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	addr := <-theChan
	if match, _ := regexp.MatchString(".+:.*", addr); !match {
		addr = fmt.Sprintf("%s:443", addr)
	}
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.Close()

	re := regexp.MustCompile(`^\*`)

	for _, pc := range conn.ConnectionState().PeerCertificates {
		for _, dns := range pc.DNSNames {
			if !re.MatchString(dns) {
				fmt.Printf("%s\t%s\n", addr, dns)
			}
		}
	}
}

func main() {
	PARAMS := getParams(os.Args)

	WG := new(sync.WaitGroup)
	gorChan := make(chan string, PARAMS.nGor)

	f, err := os.Open(PARAMS.fn)
	if err != nil {
		panic("Can't open file")
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		gorChan <- scanner.Text()
		WG.Add(1)
		go getCertData(gorChan, WG)
	}
	WG.Wait()

	fmt.Println("Done!")
}
