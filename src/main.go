/*

   ____     _____            _     _
  / ___| __|_   _|   _ _ __ (_)___| |_
 | |  _ / _ \| || | | | '_ \| / __| __|
 | |_| | (_) | || |_| | |_) | \__ \ |_
  \____|\___/|_| \__, | .__/|_|___/\__|
                 |___/|_|

	Date of Creation: October 11, 2026
	Author: KERNL | Terence Montecillo
	[https://github.com/kxrnl]

*/

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	// Thanks bubbletea & lipgloss <3

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//go:embed wordlist.txt
var embeddedWords string

const maxBoxWidth = 60

// Menu choices
var (
	wordCounts = []int{10, 25, 50, 100}
	poolSizes  = []int{200, 1000, 5000}
	poolNames  = []string{"easy", "normal", "hard"}
	textColors = []string{"#ababac", "#FFFFFF", "#000000"}
)

var (
	correctStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#44ff33"))
	wrongStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0048"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#ababac"))
	cursorStyle   = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#cf9a28"))
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#377cec"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#56b2fd"))
	actionKey     = lipgloss.NewStyle().Underline(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#585b70")).
			Padding(1, 2)
)

// Helpers
func keyBindsLine(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, actionKey.Render(pairs[i])+dimStyle.Render(" "+pairs[i+1]))
	}
	return strings.Join(parts, dimStyle.Render("  ·  "))
}

func centeredView(content string, width, height int) tea.View {
	if width > 0 {
		content = lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func wrap(number, size int) int {
	return (number%size + size) % size
}

func loadWords(path string) ([]string, error) {
	text := embeddedWords
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		text = string(data)
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil, fmt.Errorf("no words found in wordlist")
	}
	return words, nil
}

func makeTarget(allWords []string, count int) []rune {
	picked := make([]string, count)
	for i := range picked {
		picked[i] = allWords[rand.IntN(len(allWords))]
	}
	return []rune(strings.Join(picked, " "))
}

// Menu
type typingSettings struct {
	wordCount int
	poolSize  int
	textColor string
}

// Each row in the main menu adding one here adds it to the count
const (
	rowStart = iota
	rowWords
	rowDifficulty
	rowSeparator
	rowAbout
	rowQuit
	rowCount
)

type menuModel struct {
	row             int
	countChoice     int
	poolChoice      int
	textColorChoice int

	start bool
	about bool

	termWidth  int
	termHeight int
}

func newMenuModel() menuModel {
	return menuModel{poolChoice: 1}
}

func (m menuModel) settings() typingSettings {
	return typingSettings{
		wordCount: wordCounts[m.countChoice],
		poolSize:  poolSizes[m.poolChoice],
		textColor: textColors[m.textColorChoice],
	}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m menuModel) handleKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := k.String()

	if m.about {
		if key == "esc" || key == "ctrl+c" || key == "space" {
			m.about = false
		}
		return m, nil
	}

	// Main menu
	switch key {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "up", "w":
		// m.row = wrap(m.row-1, rowCount)
		m.row = wrap(m.row-1, rowCount)
		if m.row == rowSeparator {
			m.row = wrap(m.row-1, rowCount)
		}
	case "down", "s":
		m.row = wrap(m.row+1, rowCount)
		if m.row == rowSeparator {
			m.row = wrap(m.row+1, rowCount)
		}
	case "left", "a":
		m = m.changeSetting(-1)
	case "right", "d":
		m = m.changeSetting(1)
	case "enter", "space":
		return m.selectRow()
	}
	return m, nil
}

func (m menuModel) selectRow() (tea.Model, tea.Cmd) {
	switch m.row {
	case rowStart:
		m.start = true
		return m, tea.Quit
	case rowAbout:
		m.about = true
	case rowQuit:
		return m, tea.Quit
	default:
		m = m.changeSetting(1)
	}
	return m, nil
}

// changeSetting moves the selected row to its next or previous choice
func (m menuModel) changeSetting(step int) menuModel {
	switch m.row {
	case rowWords:
		m.countChoice = wrap(m.countChoice+step, len(wordCounts))

	case rowDifficulty:
		m.poolChoice = wrap(m.poolChoice+step, len(poolSizes))
	}
	return m
}

func (m menuModel) View() tea.View {
	if m.about {
		return m.pageView("About", keyBindsLine("[esc]", "back to menu"), []string{
			"GoTypist",
			" typing test built and shipped for GoShip.",
			"",
			" Practice your speed & accuracy",
			" right from your command line!",
			"@kernl | Written in Go [https://go.dev]",
		})
	}

	labels := []string{
		"Start",
		fmt.Sprintf("Words       < %d >", wordCounts[m.countChoice]),
		fmt.Sprintf("Difficulty  < %s >", poolNames[m.poolChoice]),
		"",
		"About",
		"Quit",
	}

	var rows []string
	for i, label := range labels {
		// Skip if rowSeperator then in keyHandler skip also
		if i == rowSeparator {
			rows = append(rows, selectedStyle.Render(""))
			continue
		}

		if i == m.row {
			rows = append(rows, selectedStyle.Render("> "+label))
		} else {
			rows = append(rows, dimStyle.Render(" "+label))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("GoTypist // By KERNL"),
		"",
		boxStyle.Width(32).Render(strings.Join(rows, "\n")),
		"",
		keyBindsLine("[↑/↓]", "move", "[←/→]", "change", "enter", "select"),
	)
	return centeredView(content, m.termWidth, m.termHeight)
}

// pageView draws a sub page about and settings with a title, box and footer
func (m menuModel) pageView(title, footer string, lines []string) tea.View {
	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("GoTypist // "+title),
		"",
		boxStyle.Width(maxBoxWidth).Render(strings.Join(lines, "\n")),
		"",
		footer,
	)
	return centeredView(content, m.termWidth, m.termHeight)
}

// runMenu shows the menu and returns the choices when it closes
func runMenu(menu menuModel) (menuModel, error) {
	menu.start = false
	final, err := tea.NewProgram(menu).Run()
	if err != nil {
		return menu, err
	}
	return final.(menuModel), nil
}

// Typing test
type typingModel struct {
	allWords  []string
	wordCount int
	textColor string

	target []rune
	typed  []rune

	startTime         time.Time
	endTime           time.Time
	totalKeystrokes   int
	correctKeystrokes int

	quitAll bool

	termWidth  int
	termHeight int
}

func newTypingModel(allWords []string, wordCount int, textColor string) typingModel {
	return typingModel{
		allWords:  allWords,
		wordCount: wordCount,
		textColor: textColor,
		target:    makeTarget(allWords, wordCount),
	}
}

func (m typingModel) finished() bool {
	return len(m.typed) == len(m.target)
}

func (m typingModel) restart() typingModel {
	fresh := newTypingModel(m.allWords, m.wordCount, m.textColor)
	fresh.termWidth = m.termWidth
	fresh.termHeight = m.termHeight
	return fresh
}

func (m typingModel) Init() tea.Cmd {
	return nil
}

func (m typingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m typingModel) handleKey(key tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c":
		m.quitAll = true
		return m, tea.Quit
	case "esc":
		return m, tea.Quit
	case "tab":
		return m.restart(), nil
	}

	if m.finished() {
		return m, nil
	}
	if key.String() == "backspace" {
		return m.deleteLetter(), nil
	}
	return m.typeLetter(key.Text), nil
}

func (m typingModel) typeLetter(text string) typingModel {
	letters := []rune(text)
	if len(letters) != 1 {
		return m
	}

	if m.startTime.IsZero() {
		m.startTime = time.Now()
	}

	m.totalKeystrokes++
	if letters[0] == m.target[len(m.typed)] {
		m.correctKeystrokes++
	}
	m.typed = append(m.typed, letters[0])

	if m.finished() {
		m.endTime = time.Now()
	}
	return m
}

func (m typingModel) deleteLetter() typingModel {
	if len(m.typed) > 0 {
		m.typed = m.typed[:len(m.typed)-1]
	}
	return m
}

func (m typingModel) renderText() string {
	untypedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.textColor))

	var sb strings.Builder
	for i, letter := range m.target {
		show := string(letter)
		alreadyTyped := i < len(m.typed)

		switch {
		case alreadyTyped && m.typed[i] == letter:
			sb.WriteString(correctStyle.Render(show))
		case alreadyTyped:
			if letter == ' ' {
				show = "_"
			}
			sb.WriteString(wrongStyle.Render(show))
		case i == len(m.typed):
			sb.WriteString(cursorStyle.Render(show))
		default:
			sb.WriteString(untypedStyle.Render(show))
		}
	}
	return sb.String()
}

func (m typingModel) statsText() string {
	correctLetters := 0
	for i, letter := range m.typed {
		if letter == m.target[i] {
			correctLetters++
		}
	}

	minutes := m.endTime.Sub(m.startTime).Minutes()
	if minutes <= 0 || m.totalKeystrokes == 0 {
		return ""
	}

	// 5 letters is 1 word
	wpm := float64(correctLetters) / 5 / minutes
	accuracy := float64(m.correctKeystrokes) / float64(m.totalKeystrokes) * 100
	return fmt.Sprintf("Words per Minute: %.0f  Accuracy: %.1f%%", wpm, accuracy)
}

func (m typingModel) View() tea.View {
	boxWidth := maxBoxWidth
	if m.termWidth > 0 && m.termWidth-2 < boxWidth {
		boxWidth = m.termWidth - 2
	}

	lines := []string{
		titleStyle.Render("GoTypist // By KERNL"),
		"",
		boxStyle.Width(boxWidth).Render(m.renderText()),
	}
	if m.finished() {
		lines = append(lines, "", m.statsText())
	}
	lines = append(lines, "", keyBindsLine("[tab]", "new test / restart", "[esc]", "back to menu"))

	content := lipgloss.JoinVertical(lipgloss.Center, lines...)
	return centeredView(content, m.termWidth, m.termHeight)
}

// runTyping runs one test and reports if the user wants to quit everything
func runTyping(allWords []string, s typingSettings) (bool, error) {
	pool := allWords[:min(s.poolSize, len(allWords))]
	final, err := tea.NewProgram(newTypingModel(pool, s.wordCount, s.textColor)).Run()
	if err != nil {
		return false, err
	}
	return final.(typingModel).quitAll, nil
}

func printHelp() {
	out := flag.CommandLine.Output()
	fmt.Fprintln(out, "gotypist - a terminal typing test submitted for GoShip (Hackclub)")
	fmt.Fprintln(out, "Author: KERNL (https://github.com/kxrnl)")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  gotypist [flags]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	flag.PrintDefaults()
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Keys:")
	fmt.Fprintln(out, "  menu    up/down move, left/right change, enter select")
	fmt.Fprintln(out, "  typing  tab new/restart test, esc menu, ctrl+c quit")
}

func main() {
	path := flag.String("words", "", "word list file (default: built-in list)")
	flag.Usage = printHelp
	flag.Parse()

	if flag.Arg(0) == "help" {
		printHelp()
		return
	}

	allWords, err := loadWords(*path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	menu := newMenuModel()
	for {
		menu, err = runMenu(menu)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if !menu.start {
			return
		}

		quitAll, err := runTyping(allWords, menu.settings())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if quitAll {
			// fmt.Println("YOU QUIT???") This is useless
			return
		}
	}
}
