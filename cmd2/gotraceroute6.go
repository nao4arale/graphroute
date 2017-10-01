package main

import (
	"flag"
	"fmt"
//	"github.com/aeden/traceroute"
	"../traceroute6"
	"net"
)

func printHop(hop traceroute6.TracerouteHop) {
	//addr := fmt.Sprintf("%v", hop.Address)
	addr := fmt.Sprintf("%v", traceroute6.IPv6Conv(hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3],hop.Address[4], hop.Address[5], hop.Address[6], hop.Address[7],hop.Address[8], hop.Address[9], hop.Address[10], hop.Address[11],hop.Address[12], hop.Address[13], hop.Address[14], hop.Address[15]))
	hostOrAddr := addr
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}
	if hop.Success {
		fmt.Printf("%-3d %v (%v)  %v\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime)
	} else {
		fmt.Printf("%-3d *\n", hop.TTL)
	}
}

func main() {
	var m = flag.Int("m", traceroute6.DEFAULT_MAX_HOPS, `Set the max time-to-live (max number of hops) used in outgoing probe packets (default is 64)`)
	var q = flag.Int("q", 1, `Set the number of probes per "ttl" to nqueries (default is one probe).`)

	flag.Parse()
	fmt.Println("test1")
	host := flag.Arg(0)
	options := traceroute6.TracerouteOptions{}
	fmt.Println("test2")
	options.SetRetries(*q - 1)
	options.SetMaxHops(*m + 1)

	ipAddr, err := net.ResolveIPAddr("ip6", host)
	if err != nil {
		fmt.Println("test")
		return
	}

	fmt.Printf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, ipAddr, options.MaxHops(), options.PacketSize())

	c := make(chan traceroute6.TracerouteHop, 0)

	fmt.Println("test3")

	go func() {
		for {
			hop, ok := <-c
			if !ok {
				fmt.Println("test")
				fmt.Println()
				return
			}
			fmt.Println("test4")
			printHop(hop)
		}
	}()

	_, err = traceroute6.Traceroute(host, &options, c)
	if err != nil {
		fmt.Printf("Error: ", err)
	}
}
