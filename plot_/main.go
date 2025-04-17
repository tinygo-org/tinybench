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

var colors = map[string]color.Color{
	"go":      plotutil.Color(2), // Blue: Gophers.
	"tinygo":  plotutil.Color(1), // Green: for the color of the PCBs this runs on.
	"C gcc":   plotutil.Color(0), // Red: for blood of developers spilt.
	"C clang": plotutil.Color(4), // Violet: for "roses are red, violets are blue, clobbered register #32".
	"zig":     plotutil.Color(3), // Orange: for Zig go brrr.
}

func main() {
	var output, baseLang string
	flag.StringVar(&output, "o", "benchmark.png", "Output file")
	flag.StringVar(&baseLang, "base", "tinygo", "Language benchmark to normalize other benchmark timings with")
	flag.Parse()
	langs, err := parsebench(os.Stdin, baseLang)
	if err != nil {
		log.Fatal(err)
	}
	err = drawBenchmark(langs, output, baseLang)
	if err != nil {
		log.Fatal(err)
	}
}

func drawBenchmark(langs []langBench, savefile, baseLang string) error {

	var nBenchs int
	for i := range langs {
		if langs[i].Langname == baseLang {
			nBenchs = len(langs[i].Results)
		}
	}
	if nBenchs == 0 {
		return fmt.Errorf("base language %q not found among %v", baseLang, langs)
	}
	var (
		maxBenchs  = nBenchs * len(langs)
		plotHeight = 10 * vg.Inch
		plotWidth  = plotHeight * vg.Length(maxBenchs) / 20
		benchWidth = plotWidth / vg.Length(nBenchs)

		fontsize = plotHeight / 25
		barwidth = plotHeight / 35
	)
	p_binsize := plot.New()
	p_binsize.Title.Text = "Compiler/language binary size benchmark (lower is better)"
	p_binsize.Y.Label.Text = "binary size wrt " + baseLang + " (percent)"
	p_binsize.Y.Tick.Label.Font.Size = fontsize
	var binsizeplotters []plot.Plotter
	type binsizeResult struct {
		BinSize   int
		Benchmark string
	}
	for i := range langs {
		bar, err := plotter.NewBarChart(&langs[i], barwidth)
		if err != nil {
			return err
		}

		bar.Width = barwidth / vg.Length(maxBenchs)
		bar.LineStyle.Width = vg.Length(0)
		bar.Color = colors[langs[i].Langname]
		bar.Offset = barwidth * vg.Length(i)
		binsizeplotters = append(binsizeplotters, bar)
	}

	p_time := plot.New()

	p_time.Title.Text = "Compiler/language performance benchmark (lower is better)"
	p_time.Y.Label.Text = "Average runtime normalized wrt " + baseLang + " (percent)"
	p_time.Y.Tick.Label.Font.Size = fontsize

	var plotters []plot.Plotter

	for i := range langs {
		bar, err := plotter.NewBarChart(&langs[i], barwidth)
		if err != nil {
			return err
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

	p_time.Add(plotters...)
	for i := range langs {
		x := plotters[i].(*plotter.BarChart)
		p_time.Legend.Add(langs[i].DisplayName(), x)
	}
	p_time.Legend.Top = true
	p_time.NominalX(nominals...)
	p_time.X.Tick.Label.Font.Size = benchWidth / 12 // Nominal size.

	p_time.Title.TextStyle.Font.Size = fontsize
	p_time.Legend.TextStyle.Font.Size = fontsize
	p_time.Y.Label.TextStyle.Font.Size = fontsize
	p_time.X.Label.TextStyle.Font.Size = fontsize
	p_time.X.Label.Text = "Benchmark name"
	if err := p_time.Save(plotWidth, plotHeight, savefile); err != nil {
		return err
	}
	return nil
}

type langBench struct {
	Langname string
	Compiler string
	Results  []benchResult
}

func (lb *langBench) DisplayName() string {
	if lb.Langname == lb.Compiler {
		return lb.Langname
	} else if lb.Langname == "c" {
		return "C " + lb.Compiler
	} else if lb.Compiler == "tinygo" {
		return "tinygo"
	}
	return lb.Langname + " " + lb.Compiler
}

func (lb *langBench) Value(i int) float64 { return lb.Results[i].PerOpNormalized * 100 }
func (lb *langBench) Len() int            { return len(lb.Results) }

type benchResult struct {
	Name       string
	BinarySize int
	Args       string
	N          int
	PerOp      time.Duration
	// PerOpNormalized is calculated as PerOp/PerOp_baseLang
	PerOpNormalized float64
}

func (br benchResult) ID() string {
	return br.Name + " " + br.Args
}

func parsebench(r io.Reader, baseCompiler string) (langs []langBench, err error) {
	br := bufio.NewReader(r)
	var base *langBench
	binarySize := -1
	binaryCompiler := ""
	binaryBench := ""
	for {
		orig, rderr := br.ReadString('\n')
		isRecord := strings.HasPrefix(orig, "BenchmarkAll/") && strings.HasSuffix(orig, "ns/op\n")
		if !isRecord {
			_, remaining, _ := strings.Cut(orig, "name=\"")
			benchname, remaining, _ := strings.Cut(remaining, "\" compiler=\"")
			name, strbinsize, _ := strings.Cut(remaining, "\" binarysize=")
			bsize, err := strconv.Atoi(strings.TrimSpace(strbinsize))
			if err == nil {
				// Found registry of binary size.
				binaryBench = benchname
				binaryCompiler = name
				binarySize = bsize
			}
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
		langAndCompiler, _, _ := strings.Cut(splits[0], "-")
		language, compiler, ok := strings.Cut(langAndCompiler, "/")
		if !ok {
			return langs, fmt.Errorf("line %q unexpected formatting of lang/compiler tuple", orig)
		}
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
		if benchname == binaryBench && compiler == binaryCompiler {
			result.BinarySize = binarySize
		}
		added := false
		for i, lang := range langs {
			if lang.Langname == language && compiler == lang.Compiler {
				langs[i].Results = append(langs[i].Results, result)
				added = true
				break
			}
		}
		if !added {
			langs = append(langs, langBench{
				Langname: language,
				Compiler: compiler,
				Results:  []benchResult{result},
			})
		}
	}
	for i := range langs {
		if baseCompiler == langs[i].Compiler {
			base = &langs[i]
		}
	}
	if base == nil {
		return langs, fmt.Errorf("language %q not found", baseCompiler)
	}
	for i := range langs {
		for j := range langs[i].Results {
			var baseOp float64 = -1
			id := langs[i].Results[j].ID()
			for jj := range base.Results {
				if base.Results[jj].ID() == id {
					baseOp = base.Results[jj].PerOp.Seconds()
					break
				}
			}
			if baseOp < 0 {
				return langs, fmt.Errorf("failed to find %q benchmark among base language's benchmarks: %+v", id, base.Results)
			}
			langs[i].Results[j].PerOpNormalized = langs[i].Results[j].PerOp.Seconds() / baseOp
		}
	}
	return langs, nil
}
