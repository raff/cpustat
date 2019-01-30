package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
	"github.com/shirou/gopsutil/cpu"
)

var (
	all_data bool
)

func GetTimesStat() cpu.TimesStat {
	stat, err := cpu.Times(false)
	if err != nil {
		log.Fatalf("stat read fail: %v", err)
	}

	return stat[0]
}

func percent(n, total float64) float64 {
	return n * 100.0 / total
}

func getStats(prev, stat cpu.TimesStat) []float64 {
	ptotal := prev.Total()
	// prev.User + prev.Nice + prev.System + prev.Idle + prev.IowaiT +
	// prev.Irq + prev.Softirq + prev.Steal

	total := stat.Total() - ptotal
	// stat.User + stat.Nice + stat.System + stat.Idle + stat.Iowait +
	// stat.Irq + stat.Softirq + stat.Steal - ptotal

	if all_data {
		irq := (stat.Irq + stat.Softirq) - (prev.Irq + prev.Softirq)

		return []float64{
			percent(stat.User-prev.User, total),
			percent(stat.Nice-prev.Nice, total),
			percent(stat.System-prev.System, total),
			percent(stat.Idle-prev.Idle, total),
			percent(stat.Iowait-prev.Iowait, total),
			percent(irq, total),
			percent(stat.Steal-prev.Steal, total),
		}
	} else {
		cpu := (stat.User + stat.Nice + stat.System) - (prev.User + prev.Nice + prev.System)
		steal := stat.Steal - prev.Steal
		idle := total - steal - cpu

		return []float64{percent(cpu, total), percent(steal, total), percent(idle, total)}
	}
}

func main() {
	flag.BoolVar(&all_data, "all", all_data, "display all stat categories (user, nice, system, idle, iowait, irq, steal)")
	waitTime := flag.Duration("wait", 5*time.Second, "wait between runs")
	flag.Parse()

	previous := GetTimesStat()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}

	defer ui.Close()

	chart := widgets.NewStackedBarChart()
	chart.Data = [][]float64{}
	chart.MaxVal = 100.0
	chart.BarWidth = 3
	chart.NumFmt = func(n float64) string {
		if n < 0.6 {
			return ""
		}
		return fmt.Sprintf("%v", int(n+0.4))
	}

	if all_data {
		chart.Title = "CPU Usage (GREEN:User, CYAN:Nice, YELLOW:System, BLUE:Idle, WHITE:IOwait, MAGENTA:IRQ, RED:Steal)"
		chart.BarColors = []ui.Color{ui.ColorGreen,
			ui.ColorCyan,
			ui.ColorYellow,
			ui.ColorBlue,
			ui.ColorWhite,
			ui.ColorMagenta,
			ui.ColorRed}
	} else {
		chart.Title = "CPU Usage (GREEN:Work, RED:Steal, BLUE:Idle)"
		chart.BarColors = []ui.Color{ui.ColorGreen, ui.ColorRed, ui.ColorBlue}
	}

	termWidth, termHeight := ui.TerminalDimensions()
	chart.SetRect(0, 0, termWidth, termHeight)

	ui.Render(chart)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(*waitTime).C

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>", "<esc>":
				return

			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				chart.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(chart)
			}

		case <-ticker:
			curr := GetTimesStat()
			stats := getStats(previous, curr)

			data := append(chart.Data, stats)
			maxbars := chart.Inner.Size().X / (chart.BarWidth + 1)
			if len(data) > maxbars {
				data = data[len(data)-maxbars:]
			}
			chart.Data = data

			ui.Render(chart)

			previous = curr
		}
	}
}
