package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sevenc-nanashi/pjsekai-overlay/pkg/pjsekaioverlay"
	"github.com/srinathh/gokilo/rawmode"
	"golang.org/x/sys/windows"
)

func origMain(isOptionSpecified bool) {
	Title()

	var skipAviutlInstall bool
	flag.BoolVar(&skipAviutlInstall, "no-aviutl-install", false, "Skip installing AviUtl objects.")

	var outDir string
	flag.StringVar(&outDir, "out-dir", "./dist/_chartId_", "Specify the output path. _chartId_ will be replaced with the chart ID.")

	var teamPower int
	flag.IntVar(&teamPower, "team-power", 250000, "Specify the team's power.")

	var apCombo bool
	flag.BoolVar(&apCombo, "ap-combo", true, "Enable AP indicator.")

	var enUI bool
	flag.BoolVar(&enUI, "en-ui", false, "Use English for the intro.")

	flag.Usage = func() {
		fmt.Println("Usage: pjsekai-overlay [chart ID] [arguments]")
		flag.PrintDefaults()
	}

	flag.Parse()

	if !skipAviutlInstall {
		success := pjsekaioverlay.TryInstallObject()
		if success {
			fmt.Println("Successfully installed AviUtl object.")
		}
	}

	var chartId string
	if flag.Arg(0) != "" {
		chartId = flag.Arg(0)
		fmt.Printf("Chart ID: %s\n", color.GreenString(chartId))
	} else {
		fmt.Print("Enter the chart ID including the prefix ('chcy-' for Chart Cyanvas and 'ptlv-' for Potato Leaves).\n> ")
		fmt.Scanln(&chartId)
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartId))
	}

	chartSource, err := pjsekaioverlay.DetectChartSource(chartId)
	if err != nil {
		fmt.Println(color.RedString("The specified chart doesn't exist. Please enter the correct chart ID including the prefix."))
		return
	}
	fmt.Printf("Getting chart from %s%s%s... ", RgbColorEscape(chartSource.Color), chartSource.Name, ResetEscape())
	chart, err := pjsekaioverlay.FetchChart(chartSource, chartId)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}
	if chart.Engine.Version != 10 {
		fmt.Println(color.RedString(fmt.Sprintf("Failed. This engine is not supported. （version %d）", chart.Engine.Version)))
		return
	}

	fmt.Println(color.GreenString("Success"))
	fmt.Printf("  %s / %s - %s (Lv. %s)\n",
		color.CyanString(chart.Title),
		color.CyanString(chart.Artists),
		color.CyanString(chart.Author),
		color.MagentaString(strconv.Itoa(chart.Rating)),
	)

	fmt.Printf("Getting exe path... ")
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("Success"))

	cwd, err := os.Getwd()

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}

	formattedOutDir := filepath.Join(cwd, strings.Replace(outDir, "_chartId_", chartId, -1))
	fmt.Printf("Output path: %s\n", color.CyanString(filepath.Dir(formattedOutDir)))

	fmt.Print("Downloading cover... ")
	err = pjsekaioverlay.DownloadCover(chartSource, chart, formattedOutDir)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("Success"))

	fmt.Print("Downloading background... ")
	err = pjsekaioverlay.DownloadBackground(chartSource, chart, formattedOutDir)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("Success"))

	fmt.Print("Analyzing chart... ")
	levelData, err := pjsekaioverlay.FetchLevelData(chartSource, chart)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("Success"))

	if !isOptionSpecified {
		fmt.Print("Input your team's power.\n> ")
		var tmpTeamPower string
		fmt.Scanln(&tmpTeamPower)
		teamPower, err = strconv.Atoi(tmpTeamPower)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
			return
		}
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(tmpTeamPower))

	}

	fmt.Print("Calculating score... ")
	scoreData := pjsekaioverlay.CalculateScore(chart, levelData, teamPower)

	fmt.Println(color.GreenString("Success"))

	if !isOptionSpecified {
		fmt.Print("Would you like to use English for the intro? (Y/n)\n> ")
		before, _ := rawmode.Enable()
		tmpEnableENByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpEnableEN := string(tmpEnableENByte)
		rawmode.Restore(before)
		fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpEnableEN))
		if tmpEnableEN == "Y" || tmpEnableEN == "y" || tmpEnableEN == "" {
			enUI = true
		} else {
			enUI = false
		}
	}

	if !isOptionSpecified {
		fmt.Print("Would you like to enable AP indicator? (Y/n)\n> ")
		before, _ := rawmode.Enable()
		tmpEnableComboApByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpEnableComboAp := string(tmpEnableComboApByte)
		rawmode.Restore(before)
		fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpEnableComboAp))
		if tmpEnableComboAp == "Y" || tmpEnableComboAp == "y" || tmpEnableComboAp == "" {
			apCombo = true
		} else {
			apCombo = false
		}
	}
	executableDir := filepath.Dir(executablePath)
	assets := filepath.Join(executableDir, "assets")

	fmt.Print("Generating ped file... ")

	err = pjsekaioverlay.WritePedFile(scoreData, assets, apCombo, filepath.Join(formattedOutDir, "data.ped"))

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("Success"))

	fmt.Print("Generating exo file... ")

	composerAndVocals := []string{chart.Artists, "？"}
	if separateAttempt := strings.Split(chart.Artists, " / "); chartSource.Id == "chart_cyanvas" && len(separateAttempt) <= 2 {
		composerAndVocals = separateAttempt
	}

	artists := fmt.Sprintf("作詞：？    作曲：%s    編曲：？\r\nVo：%s   譜面作成：%s", composerAndVocals[0], composerAndVocals[1], chart.Author)
	if enUI {
		artists = fmt.Sprintf("Lyrics: ?    Music: %s    Arrangement: ?\r\nVocals: %s   Charter: %s", composerAndVocals[0], composerAndVocals[1], chart.Author)
	}

	err = pjsekaioverlay.WriteExoFiles(assets, formattedOutDir, chart.Title, artists)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("Failed：%s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("Success"))

	fmt.Println(color.GreenString("\nExecution complete. Import the exo file into AviUtl."))
}

func main() {
	isOptionSpecified := len(os.Args) > 1
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	origMain(isOptionSpecified)

	if !isOptionSpecified {
		fmt.Print(color.CyanString("\nPress any key to exit..."))

		before, _ := rawmode.Enable()
		bufio.NewReader(os.Stdin).ReadByte()
		rawmode.Restore(before)
	}
}
