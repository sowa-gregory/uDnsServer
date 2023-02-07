package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

const TTL = 3600
const PORT = 54

var ipmap map[string]string
var port int = PORT

const HOSTS_FILE = "hosts.txt"

func readHostsFile(fileName string) (map[string]string, error) {
	fmt.Printf("Reading hosts file:%s\n", fileName)
	ipmap := make(map[string]string)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("readHostsFile: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		// skip comment lines
		if line[0][0] == '#' {
			continue
		}
		if len(line) != 2 {
			return nil, fmt.Errorf("readHostsFile parse error in line: %s", line)
		}
		ipmap[line[0]] = line[1]
	}
	return ipmap, nil
}

func readArgHosts(startInd int) {
	ipmap = make(map[string]string)
	for ind := startInd; ind < len(os.Args); ind++ {
		host := os.Args[ind]
		arr := strings.Split(host, ":")
		if len(arr) != 2 {
			panic(fmt.Sprintf("invalid host format:%s", host))
		}
		ipmap[arr[0]] = arr[1]
	}
}

func main() {
	fmt.Println("uDNSServer v0.93")
	if len(os.Args) < 2 {
		panic("Usage: udnsserver [-p port] -h dns:ip dns:ip")
	}
	var err error
	startind := 1
	if os.Args[startind] == "-p" {
		if len(os.Args) < 3 {
			panic("no port number provided")
		}
		if port, err = strconv.Atoi(os.Args[2]); err != nil {
			panic("port must be a number")
		}
		startind = 3
	}

	if os.Args[startind] == "-h" {
		readArgHosts(startind + 1)
	}

	for domain, ip := range ipmap {
		fmt.Println(domain, ip)
	}

	dns.HandleFunc(".", handleRequest)
	srv := &dns.Server{Addr: ":" + strconv.Itoa(PORT), Net: "udp"}
	fmt.Printf("Starting dns server on port: %d\n", PORT)
	err = srv.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("Failed to set udp listener %s\n", err.Error()))
	}
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	domain := r.Question[0].Name

	ip, _, _ := net.SplitHostPort(w.RemoteAddr().String())
	fmt.Printf("Query: src_ip: %s domain: %s\n", ip, domain)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	rr1 := new(dns.A)
	rr1.Hdr = dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(TTL)}

	if ip, ok := ipmap[domain]; ok {
		rr1.A = net.ParseIP(ip)
		m.Answer = []dns.RR{rr1}
		fmt.Printf("Answer:%s\n", ip)
	} else {
		fmt.Printf("Answer: -")
	}

	w.WriteMsg(m)
}
