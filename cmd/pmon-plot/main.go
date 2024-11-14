// Copyright 2020 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/sbinet/pmon"
	"go-hep.org/x/hep/hbook"
	"go-hep.org/x/hep/hplot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func main() {
	log.SetPrefix("pmon-plot: ")
	log.SetFlags(0)

	oname := flag.String("o", "out.png", "path to output plot file")

	flag.Usage = func() {
		usage := `pmon-plot reads monitor data from pmon and creates plots.

Usage:

 $ pmon-plot [options] file.data

Examples:

 $ pmon-plot -o foo.png pmon.data
 $ pmon-plot -o foo.svg pmon.data
 $ pmon-plot -o foo.pdf pmon.data

Options:
`
		fmt.Fprintf(os.Stderr, "%s", usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	switch flag.NArg() {
	case 0:
		flag.Usage()
		log.Printf("missing input file")
		os.Exit(1)
	case 1:
		// ok.
	default:
		flag.Usage()
		log.Printf("too many input files")
		os.Exit(1)
	}

	run(*oname, flag.Arg(0))
}

func run(oname, fname string) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("could not open input pmon-data file %q: %+v", fname, err)
	}
	defer f.Close()

	meta, err := pmon.Parse(f)
	if err != nil {
		log.Fatalf("could not parse pmon-data file %q: %+v", fname, err)
	}

	const (
		layout = time.RFC3339Nano
	)

	log.Printf("cmd:   %s", meta.Cmd)
	log.Printf("freq:  %v", meta.Freq)
	log.Printf("start: %v", meta.Start.Format(layout))
	log.Printf("delta: %v", meta.Elapsed)
	log.Printf("start: %v", meta.Stop.Format(layout))

	tp := hplot.NewTiledPlot(draw.Tiles{Cols: 1, Rows: 2})
	tp.Align = true
	doVMem(tp.Plots[0], meta)
	doRSS(tp.Plots[1], meta)

	err = tp.Save(20*vg.Centimeter, -1, oname)
	if err != nil {
		log.Fatalf("could not save output plot: %+v", err)
	}
}

func doVMem(p *hplot.Plot, meta pmon.Meta) {
	const MB = 1 / 1024.0

	xs := make([]float64, len(meta.Infos))
	ys := make([]float64, len(meta.Infos))
	for i, v := range meta.Infos {
		xs[i] = float64(i) * meta.Freq.Seconds()
		ys[i] = float64(v.VMem) * MB
	}

	s2 := hplot.NewS2D(hbook.NewS2DFrom(xs, ys))
	s2.LineStyle.Color = color.RGBA{R: 255, A: 255}
	s2.LineStyle.Width = vg.Points(2)

	p.Title.Text = "VMem [MB]"
	p.X.Label.Text = "Time [s]"
	p.Y.Label.Text = "VMem [MB]"

	p.Add(s2, hplot.NewGrid())
}

func doRSS(p *hplot.Plot, meta pmon.Meta) {
	const MB = 1 / 1024.0

	xs := make([]float64, len(meta.Infos))
	ys := make([]float64, len(meta.Infos))
	for i, v := range meta.Infos {
		xs[i] = float64(i) * meta.Freq.Seconds()
		ys[i] = float64(v.RSS) * MB
	}

	s2 := hplot.NewS2D(hbook.NewS2DFrom(xs, ys))
	s2.LineStyle.Color = color.RGBA{R: 255, A: 255}
	s2.LineStyle.Width = vg.Points(2)

	p.Title.Text = "RSS [MB]"
	p.X.Label.Text = "Time [s]"
	p.Y.Label.Text = "RSS [MB]"

	p.Add(s2, hplot.NewGrid())
}
