/*

   ____     _____            _     _
  / ___| __|_   _|   _ _ __ (_)___| |_
 | |  _ / _ \| || | | | '_ \| / __| __|
 | |_| | (_) | || |_| | |_) | \__ \ |_
  \____|\___/|_| \__, | .__/|_|___/\__|
                 |___/|_|

	Date of Creation: October 11, 2026
	Author: KERNL | Terence Montecillo
	(https://github.com/kxrnl)

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

// Embed the wordlist
//
//go:embed wordlist.txt
var embeddedWords string

const maxBoxWidth = 60

var (
	horizontalPosition = lipgloss.Center
	verticalPosition   = lipgloss.Center
)

// Avaialble difficulties
var (
	wordCounts = []int{10, 25, 50, 100}
	poolSizes  = []int{200, 1000, 5000}
	poolNames  = []string{"easy", "normal", "hard"}
)

// AAAAAAA
var (
	correctStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#44ff33"))
	wrongStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0048"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#ababac"))
	cursorStyle   = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#cf9a28"))
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#377cec"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#56b2fd"))

	// Main box
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#585b70")).
			Padding(1, 2)

	actionKey = lipgloss.NewStyle().Underline(true)

	test = lipgloss.NewStyle().Blink(true)
)

// View
func helpLine(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		parts = append(parts, actionKey.Render(pairs[i])+dimStyle.Render(" "+pairs[i+1]))
	}
	return strings.Join(parts, dimStyle.Render("  ·  "))
}

func centeredView(content string, width, height int) tea.View {
	if width > 0 {
		content = lipgloss.Place(
			width, height,
			horizontalPosition, verticalPosition,
			content,
		)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func wrap(number, size int) int {
	return (number%size + size) % size
}

// Word list fetcher

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

// Words generation
type settings struct {
	wordCount int
	poolSize  int
}

// Adding items here also adds counts on the menu selection!
// Thats cool!
const (
	rowStart = iota
	rowWords
	rowDifficulty
	rowSeperator
	rowAbout
	rowQuit
	rowCount
)

type menuModel struct {
	row         int
	countChoice int
	poolChoice  int
	start       bool

	termWidth  int
	termHeight int
}

func newMenuModel() menuModel {
	return menuModel{poolChoice: 1}
}

func (m menuModel) settings() settings {
	return settings{
		wordCount: wordCounts[m.countChoice],
		poolSize:  poolSizes[m.poolChoice],
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
	switch k.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "up", "w":
		m.row = wrap(m.row-1, rowCount)
	case "down", "s":
		m.row = wrap(m.row+1, rowCount)
	case "left", "a":
		m = m.changeSetting(-1)
	case "right", "d":
		m = m.changeSetting(1)
	case "enter":
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
		// TODO: Display about\
		return m, tea.Quit

	case rowQuit:
		return m, tea.Quit
	}
	return m.changeSetting(1), nil
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

func (m menuModel) ViewSettings() tea.View {
	aboutContent := []string{
		"Test",
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("GoTypist // By KERNL"),
		"",
		boxStyle.Width(32).Render(strings.Join(aboutContent, "\n")),
		"",
		helpLine("[↑/↓]", "move", "[←/→]", "change", "enter", "select"),
	)

	return centeredView(content, m.termWidth, m.termHeight)
}

type aboutView struct {
	row int

	termWidth  int
	termHeight int
}

func (m aboutView) View() tea.View {
	labels := []string{
		"Test",
	}

	// var rows []string
	// for i, label := range labels {
	// 	if i == m.row {
	// 		rows = append(rows, selectedStyle.Render("> "+label))
	// 	} else {
	// 		rows = append(rows, dimStyle.Render(" "+label))
	// 		// Double spaces is matched alignment
	// 		// One space has a select effect (better imo)
	// 	}
	// }

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("GoTypist // By KERNL"),
		"",
		boxStyle.Width(32).Render(labels[0]),
		"",
		helpLine("[↑/↓]", "move", "[←/→]", "change", "enter", "select"),
	)
	return centeredView(content, m.termWidth, m.termHeight)
}

// Menu
func (m menuModel) View() tea.View {
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
		if i == m.row {
			rows = append(rows, selectedStyle.Render("> "+label))
		} else {
			rows = append(rows, dimStyle.Render(" "+label))
			// Double spaces is matched alignment
			// One space has a select effect (better imo)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("GoTypist // By KERNL"),
		"",
		boxStyle.Width(32).Render(strings.Join(rows, "\n")),
		"",
		helpLine("[↑/↓]", "move", "[←/→]", "change", "enter", "select"),
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

// Main typing test
type typingModel struct {
	allWords  []string
	wordCount int

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

func newTypingModel(allWords []string, wordCount int) typingModel {
	return typingModel{
		allWords:  allWords,
		wordCount: wordCount,
		target:    makeTarget(allWords, wordCount),
	}
}

func (m typingModel) finished() bool {
	// TYPED to LENGTH OF TEST
	return len(m.typed) == len(m.target)
}

func (m typingModel) restart() typingModel {
	fresh := newTypingModel(m.allWords, m.wordCount)
	fresh.termWidth = m.termWidth
	fresh.termHeight = m.termHeight
	return fresh
}

func (m typingModel) Init() tea.Cmd {
	// Nothing
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

func (m typingModel) handleKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	name := k.String()

	if name == "ctrl+c" {
		m.quitAll = true
		return m, tea.Quit
	}
	if name == "esc" {
		return m, tea.Quit
	}
	if name == "tab" {
		return m.restart(), nil
	}
	if m.finished() {
		return m, nil
	}
	if name == "backspace" {
		return m.deleteLetter(), nil
	}
	return m.typeLetter(k.Text), nil
}

func (m typingModel) typeLetter(text string) typingModel {
	letters := []rune(text)
	if len(letters) != 1 {
		return m
	}
	typedLetter := letters[0]

	if m.startTime.IsZero() {
		m.startTime = time.Now()
	}

	m.totalKeystrokes++
	expectedLetter := m.target[len(m.typed)]
	if typedLetter == expectedLetter {
		m.correctKeystrokes++
	} else {
		// Count errors
	}
	m.typed = append(m.typed, typedLetter)

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
	var stringBuilder strings.Builder
	for i, letter := range m.target {
		show := string(letter)
		alreadyTyped := i < len(m.typed)
		isCursor := i == len(m.typed)

		switch {
		case alreadyTyped && m.typed[i] == letter:
			stringBuilder.WriteString(correctStyle.Render(show))
		case alreadyTyped:
			if letter == ' ' {
				show = "_"
			}
			stringBuilder.WriteString(wrongStyle.Render(show))
		case isCursor:
			stringBuilder.WriteString(cursorStyle.Render(show))
		default:
			stringBuilder.WriteString(dimStyle.Render(show))
		}
	}
	return stringBuilder.String()
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

	wordPerMinute := float64(correctLetters) / 5 / minutes
	accuracy := float64(m.correctKeystrokes) / float64(m.totalKeystrokes) * 100
	return fmt.Sprintf("Words per Minute: %.0f  Accuracy: %.1f%%", wordPerMinute, accuracy)
}

// Typing View
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
	lines = append(lines, "", helpLine("[tab]", "new test / restart", "[esc]", "back to menu"))

	content := lipgloss.JoinVertical(lipgloss.Center, lines...)
	return centeredView(content, m.termWidth, m.termHeight)
}

// runTyping runs one test and aks if the user wants to quit
func runTyping(allWords []string, s settings) (bool, error) {
	pool := allWords[:min(s.poolSize, len(allWords))]
	final, err := tea.NewProgram(newTypingModel(pool, s.wordCount)).Run()
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

// You made it down here
func main() {
	// path := flag.String("wl", "wordlist.txt", "custom wordlist file")
	path := flag.String("words", "", "word list file (default: built-in list)")
	flag.Usage = printHelp
	flag.Parse()

	// Help
	if flag.Arg(0) == "help" {
		printHelp()
		return
	}

	// No wordlist check
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
			fmt.Printf("YOU QUIT???")
			return
		}
	}
}
