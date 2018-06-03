package fixtures

import (
	"bytes"
	"fmt"
	"sort"
)

type Match struct {
	team1 string
	team2 string
}

func (m *Match) String() string {
	return fmt.Sprintf("%s v %s\n", m.team1, m.team2)
}

func NewMatch(team1 string, team2 string) *Match {
	return &Match{
		team1: team1,
		team2: team2,
	}
}

type ScheduledMatch struct {
	Match
	date     string
	timeslot int
	court    string
}

func (m *ScheduledMatch) String() string {
	return fmt.Sprintf("%s, %d%s, %s: %s v %s\n", m.date, m.timeslot, ".15", m.court, m.team1, m.team2)
}

func NewScheduledMatch(m *Match, date string, timeslot int, court string) *ScheduledMatch {
	return &ScheduledMatch{
		Match: Match{
			m.team1,
			m.team2,
		},
		date:     date,
		timeslot: timeslot,
		court:    court,
	}
}

type Week struct {
	date             string
	timeslots        []int
	matches          []*Match
	combinationCount int
}

func (w *Week) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%v: %v\n", w.date, w.timeslots))
	for _, m := range w.matches {
		buffer.WriteString(m.String())
	}
	return buffer.String()
}

func (w *Week) combination(comb int) Schedule {
	matchCount := len(w.matches)
	answer := make(Schedule, 0, matchCount)
	remainingMatches := w.matches
	c := comb
	for i := 0; i < matchCount; i++ {
		var mi int
		c, mi = divmod(c, len(remainingMatches))
		answer = append(answer, NewScheduledMatch(remainingMatches[mi], w.date, w.timeslots[i], w.court(i)))
		remainingMatches = copyWithoutItemAt(remainingMatches, mi)
	}
	return answer[:]
}

func (w *Week) matchCount() int {
	return len(w.matches)
}

func (w *Week) court(index int) string {
	if index == 0 || w.timeslots[index] != w.timeslots[index-1] {
		return "A"
	}
	return "B"
}

func NewWeek(date string, startTime int, endTime int, firstTimeSingle bool, matches ...*Match) *Week {
	matchCount := len(matches)
	ts := make([]int, 0, matchCount)
	for t := startTime; t <= endTime && len(ts) < matchCount; t++ {
		ts = append(ts, t)
		if (t != startTime || !firstTimeSingle) && len(ts) < matchCount {
			ts = append(ts, t)
		}
	}
	return &Week{
		date:             date,
		timeslots:        ts,
		matches:          matches,
		combinationCount: combinations(len(matches)),
	}
}

type FixtureWeekList []*Week

func (fl *FixtureWeekList) Iterator(startIndices ...int) *FixtureListIterator {
	listLength := len(*fl)
	nextIndices := mapSlice(copy(startIndices, listLength), func(i, v int) int {
		if v >= fl.combinationCount(i) {
			return 0
		}
		return v
	})
	matchCount := 0
	for _, w := range *fl {
		matchCount += len(w.matches)
	}
	it := &FixtureListIterator{
		list:        fl,
		nextIndices: nextIndices,
		matchCount:  matchCount,
	}
	return it
}

func (fl *FixtureWeekList) combinationCount(w int) int {
	return (*fl)[w].combinationCount
}

type FixtureListIterator struct {
	list        *FixtureWeekList
	nextIndices []int
	matchCount  int
}

func (it *FixtureListIterator) Next() (Schedule, bool) {
	if it.nextIndices[0] == it.list.combinationCount(0) {
		return nil, false
	}
	matches := make(Schedule, 0, it.matchCount)
	for i, v := range it.nextIndices {
		for _, m := range (*it.list)[i].combination(v) {
			matches = append(matches, m)
		}
	}
	it.increment()
	return matches, true
}

func (it *FixtureListIterator) NextIndices() []int {
	return it.nextIndices
}

func (it *FixtureListIterator) increment() {
	for i := len(it.nextIndices) - 1; i >= 0; i-- {
		it.nextIndices[i]++
		if i == 0 || it.nextIndices[i] < it.list.combinationCount(i) {
			break
		}
		it.nextIndices[i] = 0
	}
}

type Schedule []*ScheduledMatch

func (s *Schedule) String() string {
	var buffer bytes.Buffer
	for _, m := range *s {
		buffer.WriteString(m.String())
	}
	return buffer.String()
}

func (s *Schedule) teamSchedules() []*TeamSchedule {
	matchesByTeam := make(map[string]*TeamSchedule)
	for _, m := range *s {
		for _, t := range []string{m.team1, m.team2} {
			ts, found := matchesByTeam[t]
			if !found {
				ts = &TeamSchedule{
					team:    t,
					matches: make([]*ScheduledMatch, 0, 10),
				}
				matchesByTeam[t] = ts
			}
			ts.matches = append(ts.matches, m)
		}
	}
	answer := make([]*TeamSchedule, 0, len(matchesByTeam))
	for _, ts := range matchesByTeam {
		answer = append(answer, ts)
	}
	sort.Slice(answer, func(i, j int) bool {
		return answer[i].team < answer[j].team
	})
	return answer
}

func (s *Schedule) Evaluate() int {
	answer := 0
	for _, ts := range s.teamSchedules() {
		if score := ts.evaluate(); score > answer {
			answer = score
		}
	}
	return answer
}

type TeamSchedule struct {
	team    string
	matches []*ScheduledMatch
}

func (ts *TeamSchedule) evaluate() int {
	timeCounts := make(map[interface{}]int)
	for _, i := range []int{6, 7, 8, 9} {
		timeCounts[i] = 0
	}
	courtCounts := make(map[interface{}]int)
	for _, i := range []string{"A", "B"} {
		courtCounts[i] = 0
	}
	updateCount := func(counts map[interface{}]int, value interface{}) {
		_, found := counts[value]
		if !found {
			counts[value] = 0
		}
		counts[value]++
	}
	for _, m := range ts.matches {
		updateCount(timeCounts, m.timeslot)
		updateCount(courtCounts, m.court)
	}
	imbalance := func(counts map[interface{}]int) int {
		min, max := -1, -1
		for _, t := range counts {
			if min == -1 || t < min {
				min = t
			}
			if max == -1 || t > max {
				max = t
			}
		}
		if min == -1 {
			return max
		}
		if max == -1 {
			return min
		}
		return max - min
	}
	answer := 0
	for i, t := range []int{5, 9} {
		if c, ok := timeCounts[t]; ok && c > i+1 {
			answer = 100
			break
		}
	}
	answer += 10*imbalance(timeCounts) + imbalance(courtCounts)
	return answer
}

func findFirst(slice []int, predicate func(int, int) bool) int {
	for i, v := range slice {
		if predicate(i, v) {
			return i
		}
	}
	return -1
}

func findLast(slice []int, predicate func(int, int) bool) int {
	for i := len(slice) - 1; i >= 0; i-- {
		v := slice[i]
		if predicate(i, v) {
			return i
		}
	}
	return -1
}

func BuildFixtureList() FixtureWeekList {
	weeks := FixtureWeekList{
		NewWeek("30 Sep", 6, 8, false,
			NewMatch("25", "26"),
			NewMatch("21", "24"),
			NewMatch("23", "22"),
			NewMatch("15", "16"),
			NewMatch("51", "52"),
			NewMatch("41", "42")),
		NewWeek("14 Oct", 6, 9, false,
			NewMatch("26", "21"),
			NewMatch("22", "25"),
			NewMatch("11", "14"),
			NewMatch("13", "12"),
			NewMatch("31", "32"),
			NewMatch("53", "510"),
			NewMatch("43", "410"),
			NewMatch("33", "310")),
		NewWeek("21 Oct", 6, 9, false,
			NewMatch("24", "23"),
			NewMatch("59", "54"),
			NewMatch("49", "44"),
			NewMatch("39", "34"),
			NewMatch("55", "58"),
			NewMatch("45", "48"),
			NewMatch("35", "38"),
			NewMatch("57", "56")),
		NewWeek("28 Oct", 6, 9, false,
			NewMatch("22", "26"),
			NewMatch("16", "11"),
			NewMatch("12", "15"),
			NewMatch("47", "46"),
			NewMatch("37", "36"),
			NewMatch("510", "51"),
			NewMatch("410", "41"),
			NewMatch("310", "31")),
		NewWeek("4 Nov", 6, 9, false,
			NewMatch("23", "21"),
			NewMatch("25", "24"),
			NewMatch("14", "13"),
			NewMatch("52", "59"),
			NewMatch("42", "49"),
			NewMatch("32", "39"),
			NewMatch("58", "53"),
			NewMatch("48", "43")),
		NewWeek("11 Nov", 6, 9, false,
			NewMatch("12", "16"),
			NewMatch("38", "33"),
			NewMatch("54", "57"),
			NewMatch("44", "47"),
			NewMatch("34", "37"),
			NewMatch("56", "55"),
			NewMatch("46", "45"),
			NewMatch("36", "35")),
		NewWeek("18 Nov", 5, 9, false,
			NewMatch("51", "59"),
			NewMatch("41", "49"),
			NewMatch("26", "23"),
			NewMatch("24", "22"),
			NewMatch("13", "11"),
			NewMatch("15", "14"),
			NewMatch("31", "39"),
			NewMatch("510", "58"),
			NewMatch("410", "48"),
			NewMatch("310", "38")),
		NewWeek("25 Nov", 6, 9, false,
			NewMatch("25", "21"),
			NewMatch("57", "52"),
			NewMatch("47", "42"),
			NewMatch("37", "32"),
			NewMatch("53", "56"),
			NewMatch("43", "46"),
			NewMatch("33", "36"),
			NewMatch("55", "54")),
		NewWeek("2 Dec", 5, 9, true,
			NewMatch("45", "44"),
			NewMatch("24", "26"),
			NewMatch("16", "13"),
			NewMatch("14", "12"),
			NewMatch("35", "34"),
			NewMatch("58", "51"),
			NewMatch("48", "41"),
			NewMatch("38", "31"),
			NewMatch("59", "57")),
		NewWeek("9 Dec", 6, 9, false,
			NewMatch("25", "23"),
			NewMatch("22", "21"),
			NewMatch("15", "11"),
			NewMatch("49", "47"),
			NewMatch("39", "37"),
			NewMatch("56", "510"),
			NewMatch("46", "410"),
			NewMatch("36", "310")),
		NewWeek("16 Dec", 6, 9, false,
			NewMatch("14", "16"),
			NewMatch("52", "55"),
			NewMatch("42", "45"),
			NewMatch("32", "35"),
			NewMatch("54", "53"),
			NewMatch("44", "43"),
			NewMatch("34", "33"),
			NewMatch("51", "57")),
		NewWeek("6 Jan", 6, 9, false,
			NewMatch("26", "25"),
			NewMatch("24", "21"),
			NewMatch("15", "13"),
			NewMatch("12", "11"),
			NewMatch("41", "47"),
			NewMatch("31", "37"),
			NewMatch("58", "56"),
			NewMatch("48", "46")),
		NewWeek("13 Jan", 6, 9, false,
			NewMatch("23", "22"),
			NewMatch("38", "36"),
			NewMatch("55", "59"),
			NewMatch("45", "49"),
			NewMatch("35", "39"),
			NewMatch("510", "54"),
			NewMatch("410", "44"),
			NewMatch("310", "34")),
		NewWeek("20 Jan", 6, 9, false,
			NewMatch("21", "26"),
			NewMatch("16", "15"),
			NewMatch("14", "11"),
			NewMatch("53", "52"),
			NewMatch("43", "42"),
			NewMatch("33", "32"),
			NewMatch("56", "51"),
			NewMatch("46", "41")),
		NewWeek("27 Jan", 6, 9, false,
			NewMatch("25", "22"),
			NewMatch("23", "24"),
			NewMatch("13", "12"),
			NewMatch("36", "31"),
			NewMatch("57", "55"),
			NewMatch("47", "45"),
			NewMatch("37", "35"),
			NewMatch("54", "58")),
		NewWeek("3 Feb", 6, 9, false,
			NewMatch("11", "16"),
			NewMatch("44", "48"),
			NewMatch("34", "38"),
			NewMatch("59", "53"),
			NewMatch("49", "43"),
			NewMatch("39", "33"),
			NewMatch("52", "510"),
			NewMatch("42", "410")),
		NewWeek("10 Feb", 5, 9, false,
			NewMatch("32", "310"),
			NewMatch("51", "55"),
			NewMatch("26", "22"),
			NewMatch("21", "23"),
			NewMatch("15", "12"),
			NewMatch("13", "14"),
			NewMatch("41", "45"),
			NewMatch("31", "35"),
			NewMatch("56", "54"),
			NewMatch("46", "44")),
		NewWeek("17 Feb", 6, 9, false,
			NewMatch("25", "24"),
			NewMatch("36", "34"),
			NewMatch("53", "57"),
			NewMatch("43", "47"),
			NewMatch("33", "37"),
			NewMatch("58", "52"),
			NewMatch("48", "42"),
			NewMatch("38", "32")),
		NewWeek("24 Feb", 6, 9, false,
			NewMatch("23", "26"),
			NewMatch("16", "12"),
			NewMatch("11", "13"),
			NewMatch("510", "59"),
			NewMatch("410", "49"),
			NewMatch("310", "39"),
			NewMatch("54", "51"),
			NewMatch("44", "41")),
		NewWeek("3 Mar", 6, 9, false,
			NewMatch("22", "24"),
			NewMatch("15", "14"),
			NewMatch("34", "31"),
			NewMatch("55", "53"),
			NewMatch("45", "43"),
			NewMatch("52", "56"),
			NewMatch("57", "510"),
			NewMatch("59", "58")),
		NewWeek("10 Mar", 6, 9, false,
			NewMatch("25", "21"),
			NewMatch("13", "16"),
			NewMatch("35", "33"),
			NewMatch("42", "46"),
			NewMatch("32", "36"),
			NewMatch("47", "410"),
			NewMatch("37", "310"),
			NewMatch("49", "48")),
		NewWeek("17 Mar", 6, 7, false,
			NewMatch("12", "14"),
			NewMatch("15", "11"),
			NewMatch("39", "38")),
		//NewMatch("51", "53"),
		//NewMatch("54", "52"),
		//NewMatch("510", "55"),
		//NewMatch("56", "59"),
		//NewMatch("58", "57")),
		//NewWeek("24 Mar", 6, 9, false,
		//	NewMatch("26", "24"),
		//	NewMatch("23", "25"),
		//	NewMatch("21", "22"),
		//	NewMatch("41", "43"),
		//	NewMatch("44", "42"),
		//	NewMatch("410", "45"),
		//	NewMatch("46", "49"),
		//	NewMatch("48", "47")),
		//NewWeek("31 Mar", 6, 9, false,
		//	NewMatch("16", "14"),
		//	NewMatch("13", "15"),
		//	NewMatch("11", "12"),
		//	NewMatch("31", "33"),
		//	NewMatch("34", "32"),
		//	NewMatch("310", "35"),
		//	NewMatch("36", "39"),
		//	NewMatch("38", "37")),
	}
	return weeks
}

func combinations(itemCount int) int {
	if itemCount == 2 {
		return 2
	}
	return itemCount * combinations(itemCount-1)
}

func copyWithoutItemAt(list []*Match, index int) []*Match {
	count := len(list) - 1
	answer := make([]*Match, 0, count)
	for sourceIndex, m := range list {
		if sourceIndex != index {
			answer = append(answer, m)
		}
	}
	return answer
}

func divmod(dividend int, divisor int) (quotient int, remainder int) {
	quotient = dividend / divisor
	remainder = dividend % divisor
	return
}

func copy(slice []int, length int) []int {
	answer := make([]int, length)
	for i, v := range slice {
		if i >= length {
			break
		}
		answer[i] = v
	}
	return answer
}

func mapSlice(slice []int, f func(int, int) int) []int {
	answer := make([]int, len(slice))
	for i, v := range slice {
		answer[i] = f(i, v)
	}
	return answer
}
