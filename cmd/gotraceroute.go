package main

import (
	"flag"
	"fmt"
	//	"github.com/aeden/traceroute"
	"../traceroute"
	"github.com/nsf/termbox-go"
	"net"
	"time"
)

func printHop(hop traceroute.TracerouteHop, i int) {
	addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
	hostOrAddr := addr
	if hop.Host != "" {
		hostOrAddr = hop.Host
	}
	if hop.Success {
		//fmt.Printf("%-3d %v (%v)  %v\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime)
		drawLine(1, i, fmt.Sprintf("%-3d %v (%v)  %v\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime))
	} else {
		//fmt.Printf("%-3d *\n", hop.TTL)
		drawLine(1, i, fmt.Sprintf("%-3d *\n", hop.TTL))
	}
}

/*
func address(address [4]byte) string {
	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
}
*/

func drawLine(x, y int, str string) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault
	runes := []rune(str)

	for n := 0; n < len(runes); n += 1 {
		termbox.SetCell(x+n, y, runes[n], color, backgroundColor)
	}
}

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func keyEventLoop(killKey chan termbox.Key, chanMaxX, chanMaxY chan int) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			killKey <- ev.Key
		case termbox.EventResize:
			maxX, maxY := termbox.Size()
			chanMaxX <- maxX
			chanMaxY <- maxY
		default:
		}
	}
}

func traceLoop(chanMaxX, chanMaxY chan int) {
	var m = flag.Int("m", traceroute.DEFAULT_MAX_HOPS, `Set the max time-to-live (max number of hops) used in outgoing probe packets (default is 64)`)
	var q = flag.Int("q", 1, `Set the number of probes per "ttl" to nqueries (default is one probe).`)
	maxX, maxY := termbox.Size()

	flag.Parse()
	host := flag.Arg(0)

	options := traceroute.TracerouteOptions{}
	options.SetRetries(*q - 1)
	options.SetMaxHops(*m + 1)

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return
	}

	//fmt.Printf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, ipAddr, options.MaxHops(), options.PacketSize())
	drawLine(1, 0, fmt.Sprintf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, ipAddr, options.MaxHops(), options.PacketSize()))

	go func () {
		for {
			select {
			case <- chanMaxX:
				//maxX, maxY = termbox.Size()
				maxX = <- chanMaxX
				maxY = <- chanMaxY
			default:
			}
		}
	}()


	for {
		i := 1
		c := make(chan traceroute.TracerouteHop, 0)
		done := make(chan struct{}, 0)
		go func() {
			for {
				hop, ok := <-c
				if !ok {
					done <- struct{}{}
				}
				printHop(hop, i)
				drawLine(100, 0, fmt.Sprintf("%v:%v", maxX, maxY))
				termbox.Flush()
				i++
			}
		}()

		_, err = traceroute.Traceroute(host, &options, c)
		if err != nil {
			fmt.Printf("Error: ", err)
		}
		<-done
		fill(0, 1, 80, maxY-1, termbox.Cell{Ch: ' '})
		time.Sleep(2000 * time.Millisecond)
	}
}

func main() {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	defer termbox.Close()

	maxX, maxY := termbox.Size()
	chanMaxX, chanMaxY := make(chan int, maxX), make(chan int, maxY)

	//terch := make(chan struct{})

	killKey := make(chan termbox.Key)

	go traceLoop(chanMaxX, chanMaxY)
	go keyEventLoop(killKey, chanMaxX, chanMaxY)

	for {
		select {
		case wait := <-killKey:
			switch wait {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return
			}
		//case <-terch:
		//	maxX, maxY = termbox.Size()
		//	chanMaxX <- maxX
		//	chanMaxY <- maxY
		}
	}
}
