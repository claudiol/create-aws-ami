package cmd

import (
	"flag"
	"fmt"
	"log"
	"time"
)

const (
	maxbars  int           = 100
	interval time.Duration = 500 * time.Millisecond
	thebars  string        = "========================================================================================================"
)

var src string

// Progbar Struct
type Progbar struct {
	total int
}

// PrintProg function
func (p *Progbar) PrintProg(portion int) {
	bars := p.calcBars(portion)
	//spaces := maxbars - bars - 1
	percent := 100 * (float32(portion) / float32(p.total))

	fmt.Print("\033[G\033[K")
	fmt.Print("Progress [")

	fmt.Print(thebars[:bars])
	fmt.Print(">")
	//fmt.Printf(" %3.2f%% (%d/%d) ]", percent, portion, p.total)
	fmt.Printf(" ] %3.2f%% ", percent) //, portion, p.total)
}

// PrintComplete function
func (p *Progbar) PrintComplete() {
	p.PrintProg(p.total)
	fmt.Print("\n")
}

func (p *Progbar) calcBars(portion int) int {
	if portion == 0 {
		return portion
	}

	return int(float32(maxbars) / (float32(p.total) / float32(portion)))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	flag.Parse()
	src = flag.Args()[0]
}

// func main() {
// 	fmt.Printf("Downloading %s\n", src)

// 	res, err := http.Get(src)
// 	check(err)
// 	defer res.Body.Close()

// 	u, err := url.Parse(src)
// 	check(err)

// 	out, err := os.Create(path.Base(u.Path))
// 	defer out.Close()

// 	size := res.ContentLength
// 	bar := &Progbar{total: int(size)}
// 	written := make(chan int, 500)

// 	go func() {
// 		copied := 0
// 		c := 0
// 		tick := time.Tick(interval)

// 		for {
// 			select {
// 			case c = <-written:
// 				copied += c
// 			case <-tick:
// 				bar.PrintProg(copied)
// 			}
// 		}
// 	}()

// 	buf := make([]byte, 32*1024)
// 	for {
// 		rc, re := res.Body.Read(buf)
// 		if rc > 0 {
// 			wc, we := out.Write(buf[0:rc])
// 			check(we)

// 			if wc != rc {
// 				log.Fatal("Read and Write count mismatch")
// 			}

// 			if wc > 0 {
// 				written <- wc
// 			}
// 		}
// 		if re == io.EOF {
// 			break
// 		}
// 		check(re)
// 	}
// 	bar.PrintComplete()

// }
