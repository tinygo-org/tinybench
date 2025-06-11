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

var compilerColors = map[string]color.Color{
	"go":     plotutil.Color(2), // Blue: Gophers.
	"tinygo": plotutil.Color(1), // Green: for the color of the PCBs this runs on.
	"gcc":    plotutil.Color(0), // Red: for blood of developers spilt.
	"clang":  plotutil.Color(4), // Violet: for "roses are red, violets are blue, clobbered register #32".
	"zig":    plotutil.Color(3), // Orange: for Zig go brrr.
	"rustc":  plotutil.Color(5), // How serendipitous, next color looks like rust.
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

func drawBenchmark(langs []langBench, savefile, baseCompiler string) error {
	var nBenchs int
	for i := range langs {
		if langs[i].Compiler == baseCompiler {
			nBenchs = len(langs[i].Results)
		}
	}
	if nBenchs == 0 {
		return fmt.Errorf("base compiler %q not found among %v", baseCompiler, langs)
	}
	var (
		maxBenchs  = nBenchs * len(langs)
		plotHeight = 10 * vg.Inch
		plotWidth  = plotHeight * vg.Length(maxBenchs) / 20
		benchWidth = plotWidth / vg.Length(nBenchs)

		fontsize = plotHeight / 25
		barwidth = plotHeight / 35
	)
	// p_binsize := plot.New()
	// p_binsize.Title.Text = "Compiler/language binary size benchmark (lower is better)"
	// p_binsize.Y.Label.Text = "binary size wrt " + baseCompiler + " (percent)"
	// p_binsize.Y.Tick.Label.Font.Size = fontsize
	// var binsizeplotters []plot.Plotter
	// type binsizeResult struct {
	// 	BinSize   int
	// 	Benchmark string
	// }
	// for i := range langs {
	// 	bar, err := plotter.NewBarChart(&langs[i], barwidth)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	bar.Width = barwidth / vg.Length(maxBenchs)
	// 	bar.LineStyle.Width = vg.Length(0)
	// 	bar.Color = colors[langs[i].Langname]
	// 	bar.Offset = barwidth * vg.Length(i)
	// 	binsizeplotters = append(binsizeplotters, bar)
	// }

	p_time := plot.New()

	p_time.Title.Text = "Compiler/language performance benchmark (lower is better)"
	p_time.Y.Label.Text = "Average runtime normalized wrt " + baseCompiler + " (percent)"
	p_time.Y.Tick.Label.Font.Size = fontsize

	var plotters []plot.Plotter

	for i := range langs {
		bar, err := plotter.NewBarChart(&langs[i], barwidth)
		if err != nil {
			return err
		}

		bar.Width = barwidth
		bar.LineStyle.Width = vg.Length(0)
		bar.Color = compilerColors[langs[i].Compiler]
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

type rawCompilerResult struct {
	Benchmark string
	Compiler  string
	Size      int
	Runs      []rawCompilerBench
}

type rawCompilerBench struct {
	Benchmark string
	Arg       string
	LangName  string
	Compiler  string
	N         int
	PerOp     time.Duration
}

type langBench struct {
	Langname string
	Compiler string
	Results  []benchResult
}

func (cr *rawCompilerResult) Language() (langname string) {
	if len(cr.Runs) > 0 {
		langname = cr.Runs[0].LangName
	} else {
		langname = cr.Compiler
	}
	return langname
}

func (lb *langBench) DisplayName() string {
	if lb.Langname == lb.Compiler {
		return lb.Langname
	} else if lb.Langname == "c" {
		return "C " + lb.Compiler
	} else if lb.Compiler == "tinygo" {
		return "tinygo"
	} else if lb.Compiler == "rustc" {
		return "rust"
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

func parseKeyValue(data string, key string) string {
	data = strings.TrimSpace(data)
	_, vstart, ok1 := strings.Cut(data, key+"=")
	hasQuote := ok1 && vstart[0] == '"'
	if hasQuote {
		value, _, _ := strings.Cut(vstart[1:], "\"")
		return value
	}
	value, _, _ := strings.Cut(vstart, " ")
	return value
}

func parsebench(r io.Reader, baseCompiler string) (langs []langBench, err error) {
	rawbench, err := parseBenchRaw(r)
	if err != nil {
		return langs, err
	}
	langs = make([]langBench, 0, 256) // Make sure not resized since we are using a pointer to it.
	var base *langBench
	for _, bench := range rawbench {
		binsize := bench.Size
		lang := bench.Language()
		compiler := bench.Compiler
		benchname := bench.Benchmark
		for _, run := range bench.Runs {
			added := false
			result := benchResult{
				Name:       benchname,
				BinarySize: binsize,
				Args:       run.Arg,
				N:          run.N,
				PerOp:      run.PerOp,
			}
			for i := range langs {
				target := &langs[i]
				if target.Compiler == compiler && target.Langname == lang {
					target.Results = append(target.Results, result)
					added = true
					break
				}
			}
			if !added {
				langs = append(langs, langBench{
					Langname: lang,
					Compiler: compiler,
					Results:  []benchResult{result},
				})
				if baseCompiler == compiler {
					base = &langs[len(langs)-1]
				}
			}
		}
	}
	if base == nil {
		return langs, fmt.Errorf("base compiler %q result not found", baseCompiler)
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
				return langs, fmt.Errorf("failed to find %q benchmark among base compiler's %q benchmarks: %+v", id, baseCompiler, base.Results)
			}
			langs[i].Results[j].PerOpNormalized = langs[i].Results[j].PerOp.Seconds() / baseOp
		}
	}
	return langs, nil
}

func parseBenchRaw(r io.Reader) (benchs []rawCompilerResult, err error) {
	br := bufio.NewReader(r)
	var currentBench *rawCompilerResult
	line := 0
	for {
		orig, rderr := br.ReadString('\n')
		if rderr != nil {
			break
		}
		line++
		newBench := strings.Contains(orig, " compiler=") && strings.Contains(orig, " binarysize=")
		if newBench {
			if currentBench != nil {
				benchs = append(benchs, *currentBench)
			}
			sz, _ := strconv.Atoi(parseKeyValue(orig, "binarysize"))
			currentBench = &rawCompilerResult{
				Benchmark: parseKeyValue(orig, "name"),
				Compiler:  parseKeyValue(orig, "compiler"),
				Size:      sz,
			}
			continue
		}
		isRecord := strings.HasPrefix(orig, "BenchmarkAll/") && strings.HasSuffix(orig, "ns/op\n")
		if !isRecord {
			continue
		} else if currentBench == nil {
			return benchs, fmt.Errorf("invalid line %d ordering %q", line, orig)
		}
		data := strings.TrimPrefix(orig, "BenchmarkAll/")
		benchmark, data, _ := strings.Cut(data, ":")
		args, data, _ := strings.Cut(data, "/")
		language, data, _ := strings.Cut(data, "/")
		compiler, data, _ := strings.Cut(data, " ")
		if strings.Contains(compiler, "-") {
			compiler, _, _ = strings.Cut(compiler, "-")
		}
		numbers := strings.Fields(data)
		if len(numbers) != 3 {
			return benchs, fmt.Errorf("malformed benchmark data line %d: %q", line, orig)
		}
		N, _ := strconv.Atoi(numbers[0])
		ns, _ := strconv.Atoi(numbers[1])
		if currentBench.Compiler != compiler {
			return benchs, fmt.Errorf("mismatch compiler's name %q!=%q", currentBench.Compiler, compiler)
		}
		if currentBench.Benchmark != benchmark {
			return benchs, fmt.Errorf("mismatch compiler's benchmark %q!=%q", currentBench.Benchmark, benchmark)
		}
		currentBench.Runs = append(currentBench.Runs, rawCompilerBench{
			Benchmark: benchmark,
			Arg:       args,
			LangName:  language,
			Compiler:  compiler,
			N:         N,
			PerOp:     time.Duration(ns) * time.Nanosecond,
		})
	}
	if currentBench != nil {
		benchs = append(benchs, *currentBench)
	}
	return benchs, nil
}
