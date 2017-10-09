package main

import (
	"flag"
	"fmt"
	//	"github.com/aeden/traceroute"
	"../traceroute"
	"github.com/nsf/termbox-go"
	"net"
	"os"
	"time"
)

func printHop(hop traceroute.TracerouteHop, i int, kill chan struct{}) {
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

func drawLine(x, y int, str string) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault
	runes := []rune(str)

	for n := 0; n < len(runes); n += 1 {
		termbox.SetCell(x+n, y, runes[n], color, backgroundColor)
	}
}

func drawLineFull(x, y int, str string, lineAttr, backAttr termbox.Attribute) {
	color := lineAttr
	backgroundColor := backAttr
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

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

var m = flag.Int("m", traceroute.DEFAULT_MAX_HOPS, `Set the max time-to-live (max number of hops) used in outgoing probe packets (default is 64)`)
var q = flag.Int("q", 1, `Set the number of probes per "ttl" to nqueries (default is one probe).`)

func traceLoop(host string, maxX, maxY int, skip, received chan struct{}) {
//	var m = flag.Int("m", traceroute.DEFAULT_MAX_HOPS, `Set the max time-to-live (max number of hops) used in outgoing probe packets (default is 64)`)
//	var q = flag.Int("q", 1, `Set the number of probes per "ttl" to nqueries (default is one probe).`)
	//maxX, maxY := termbox.Size()

//	flag.Parse()
	//host := flag.Arg(0)

	options := traceroute.TracerouteOptions{}
	options.SetRetries(*q - 1)
	options.SetMaxHops(*m + 1)

	ipAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return
	}

	drawLine(1, 0, fmt.Sprintf("traceroute to %v (%v), %v hops max, %v byte packets\n", host, ipAddr, options.MaxHops(), options.PacketSize()))

	/*
		go func() {
			for {
				select {
				case <-chanMaxX:
					//maxX, maxY = termbox.Size()
					maxX = <-chanMaxX
					maxY = <-chanMaxY
				default:
				}
			}
		}()
	*/

	//done := make(chan struct{}, 0)
	kill := make(chan struct{}, 0)

			i := 1
			c := make(chan traceroute.TracerouteHop, 0)
			var hop traceroute.TracerouteHop
			var ok bool
			go func() {
				for {
					select {
					case hop, ok = <-c:
						switch ok {
						case false:
							//done <- struct{}{}
							return
						case true:
							printHop(hop, i, kill)
							termbox.Flush()
							i++
						}
					//case <-kill:
					//	return
					default:
					}
				}
			}()

			_, err = traceroute.Traceroute(host, &options, kill, c)
			if err != nil {
				termbox.Close()
				fmt.Println("Error: ", err)
				os.Exit(1)
			}
			//<-done
			time.Sleep(1000 * time.Millisecond)
			fill(0, 1, 80, maxY-1, termbox.Cell{Ch: ' '})
		}

//var edit_box EditBox

//const edit_box_width = 30

func main() {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	flag.Parse()

	defer termbox.Close()

	text := make([]string, 0, 30)
	//tmp := make([]string, 0, 30)
	maxX, maxY := termbox.Size()
	//chanMaxX, chanMaxY := make(chan int, maxX), make(chan int, maxY)

	skip := make(chan struct{}, 0)
	received := make(chan struct{}, 0)
	host := flag.Arg(0)
	huff := host

	cursX := 80
	termbox.SetCursor(cursX+1, 2)

	go func() {
		for {
			host = huff
			traceLoop(host, maxX, maxY, skip, received)
		}
	}()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return
			case termbox.KeyBackspace:
				if cursX > 80 {
					cursX--
					termbox.SetCursor(cursX+1, 2)
					drawLine(cursX+1, 2, " ")

					//tmp := make([]string, 0, len(text)-1)
					text = text[:len(text)-1]
					//copy(tmp, text)
					termbox.Flush()
				}
			case termbox.KeyEnter:
				x := 80
				huff = ""
				fill(x, 2, maxX-x, 2, termbox.Cell{Ch: ' '})
				for _, s := range text {
					drawLineFull(x+1, 3, s, termbox.ColorRed, termbox.ColorDefault)
					huff = huff + s
					x++
				}
				text = make([]string, 0, 30)
				cursX = 80
				termbox.SetCursor(cursX+1, 2)
				termbox.Flush()
			default:
				if cursX < maxX-1 {
					cursX++
					termbox.SetCursor(cursX+1, 2)
					drawLine(cursX, 2, fmt.Sprintf("%c", ev.Ch))
					termbox.Flush()
					text = append(text, fmt.Sprintf("%c", ev.Ch))
				}
			}
		}
	}
}
