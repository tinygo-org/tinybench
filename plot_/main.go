package main

import (
	"bufio"
	"flag"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	var output string
	flag.StringVar(&output, "o", "benchmark.png", "Output file")
	flag.Parse()
	langs, err := parsebench(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	var (
		size     = 5 * vg.Inch * vg.Length(len(langs))
		fontsize = size / 50
		barwidth = size / 60
	)

	p := plot.New()

	p.Title.Text = "Compiler/language benchmark (lower is better)"
	p.Y.Label.Text = "Average runtime (milliseconds)"
	p.Y.Tick.Label.Font.Size = fontsize

	var plotters []plot.Plotter
	var colors = map[string]color.Color{
		"go":      plotutil.Color(2),
		"tinygo":  plotutil.Color(1),
		"C clang": plotutil.Color(4),
		"C gcc":   plotutil.Color(0),
		"zig":     plotutil.Color(3), // hopefully added someday.
	}
	for i := range langs {
		bar, err := plotter.NewBarChart(&langs[i], barwidth)
		if err != nil {
			panic(err)
		}

		bar.Width = barwidth
		bar.LineStyle.Width = vg.Length(0)
		bar.Color = colors[langs[i].Langname]
		bar.Offset = barwidth * vg.Length(i)
		plotters = append(plotters, bar)
	}

	var nominals []string
	for i := range langs[0].Results {
		nominal := langs[0].Results[i].Name + " " + langs[0].Results[i].Args
		nominals = append(nominals, nominal)
	}

	p.Add(plotters...)
	for i := range langs {
		x := plotters[i].(*plotter.BarChart)
		p.Legend.Add(langs[i].Langname, x)
	}
	p.Legend.Top = true
	p.NominalX(nominals...)

	p.Title.TextStyle.Font.Size = fontsize
	p.Legend.TextStyle.Font.Size = fontsize
	p.Y.Label.TextStyle.Font.Size = fontsize
	p.X.Label.TextStyle.Font.Size = fontsize
	p.X.Label.Text = "Benchmark name"
	if err := p.Save(size, size/2, output); err != nil {
		panic(err)
	}
}

type langBench struct {
	Langname string
	Results  []benchResult
}

func (lb *langBench) Value(i int) float64 { return lb.Results[i].PerOp.Seconds() * 1000 }
func (lb *langBench) Len() int            { return len(lb.Results) }

type benchResult struct {
	Name  string
	Args  string
	N     int
	PerOp time.Duration
}

func parsebench(r io.Reader) (langs []langBench, err error) {
	br := bufio.NewReader(r)
	for {
		orig, rderr := br.ReadString('\n')
		if !strings.HasPrefix(orig, "BenchmarkAll") {
			if rderr != nil {
				break
			}
			continue
		}
		_, line, _ := strings.Cut(orig, "/")
		benchname, line, _ := strings.Cut(line, "/")
		benchname, args, _ := strings.Cut(benchname, ":")
		splits := strings.Fields(line)
		if len(splits) != 4 {
			return langs, fmt.Errorf("line %q unexpected formatting", orig)
		}
		langname, _, _ := strings.Cut(splits[0], "-")
		langname = strings.ReplaceAll(langname, "_", " ")
		N, err := strconv.Atoi(splits[1])
		if err != nil {
			return langs, err
		}
		ns, err := strconv.Atoi(splits[2])
		if err != nil {
			return langs, err
		}
		result := benchResult{
			Name:  benchname,
			Args:  args,
			N:     N,
			PerOp: time.Duration(ns) * time.Nanosecond,
		}
		added := false
		for i, lang := range langs {
			if lang.Langname == langname {
				langs[i].Results = append(langs[i].Results, result)
				added = true
				break
			}
		}
		if !added {
			langs = append(langs, langBench{
				Langname: langname,
				Results:  []benchResult{result},
			})
		}
	}
	return langs, nil
}
