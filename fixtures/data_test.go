package fixtures

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"bytes"
	"regexp"
	"strconv"
)

func TestBuildFixtureList(t *testing.T) {
	fl := BuildFixtureList()
	assert.Equal(t, 22, len(fl))
	for _, w := range fl {
		assert.Equal(t, len(w.timeslots), len(w.matches))
	}
}

func TestCombinations(t *testing.T) {
	assert.Equal(t, 720, combinations(6))
	assert.Equal(t, 5040, combinations(7))
	assert.Equal(t, 40320, combinations(8))
	assert.Equal(t, 362880, combinations(9))
	assert.Equal(t, 3628800, combinations(10))
}

func TestWeek_Combination_0(t *testing.T) {
	week := BuildFixtureList()[0]
	fixtures := week.combination(0)
	assert.Equal(t, 6, len(fixtures))
	index := Incrementable(0)
	checkExpected(t, "25", "26", "30 Sep", 6, "A", fixtures[index.postInc()])
	checkExpected(t, "21", "24", "30 Sep", 6, "B", fixtures[index.postInc()])
	checkExpected(t, "23", "22", "30 Sep", 7, "A", fixtures[index.postInc()])
	checkExpected(t, "15", "16", "30 Sep", 7, "B", fixtures[index.postInc()])
	checkExpected(t, "51", "52", "30 Sep", 8, "A", fixtures[index.postInc()])
	checkExpected(t, "41", "42", "30 Sep", 8, "B", fixtures[index.postInc()])
}

func TestWeek_Combination_500(t *testing.T) {
	week := BuildFixtureList()[0]
	fixtures := week.combination(500)
	assert.Equal(t, 6, len(fixtures))
	index := Incrementable(0)
	checkExpected(t, "23", "22", "30 Sep", 6, "A", fixtures[index.postInc()])
	checkExpected(t, "51", "52", "30 Sep", 6, "B", fixtures[index.postInc()])
	checkExpected(t, "25", "26", "30 Sep", 7, "A", fixtures[index.postInc()])
	checkExpected(t, "15", "16", "30 Sep", 7, "B", fixtures[index.postInc()])
	checkExpected(t, "41", "42", "30 Sep", 8, "A", fixtures[index.postInc()])
	checkExpected(t, "21", "24", "30 Sep", 8, "B", fixtures[index.postInc()])
}

func checkExpected(t *testing.T, team1 string, team2 string, date string, timeslot int, court string, sm *ScheduledMatch) {
	assert.Equal(t, team1, sm.team1)
	assert.Equal(t, team2, sm.team2)
	assert.Equal(t, date, sm.date)
	assert.Equal(t, timeslot, sm.timeslot)
	assert.Equal(t, court, sm.court)
}

type Incrementable int

func (i *Incrementable) preInc() int {
	*i++
	return i.get()
}

func (i *Incrementable) postInc() int {
	answer := i.get()
	*i++
	return answer
}

func (i *Incrementable) get() int {
	return int(*i)
}

func TestIncrementable(t *testing.T) {
	i := Incrementable(42)
	assert.Equal(t, 43, i.preInc())
	assert.Equal(t, 43, i.get())
	assert.Equal(t, 43, i.postInc())
	assert.Equal(t, 44, i.get())
}

func TestInitialIterator(t *testing.T) {
	list := BuildFixtureList()
	it := list.Iterator()
	assert.Equal(t, len(list), len(it.nextIndices))
	for _, v := range it.nextIndices {
		assert.Zero(t, v)
	}
}

func TestIteratorWithStartPosition(t *testing.T) {
	list := BuildFixtureList()
	it := list.Iterator(1, 2, 3, 4)
	assert.Equal(t, len(list), len(it.nextIndices))
	for i, v := range it.nextIndices {
		if i < 4 {
			assert.Equal(t, i+1, v)
		} else {
			assert.Zero(t, v)
		}
	}
}

func TestIteratorWithTooManyStartPositions(t *testing.T) {
	list := FixtureWeekList{
		NewWeek("31 May", 1, 2, true,
			NewMatch("11", "12"),
			NewMatch("13", "14"),
			NewMatch("15", "16")),
		NewWeek("1 Jun", 3, 4, true,
			NewMatch("21", "22"),
			NewMatch("23", "24"),
			NewMatch("25", "26")),
		NewWeek("2 Jun", 5, 6, true,
			NewMatch("31", "32"),
			NewMatch("33", "34"),
			NewMatch("35", "36")),
	}
	it := list.Iterator(1, 2, 6, 4)
	assert.Equal(t, []int{1, 2, 0}, it.nextIndices)
}

func TestIteratorNextFirstIteration(t *testing.T) {
	weeks := FixtureWeekList{
		NewWeek("31 May", 1, 2, true,
			NewMatch("11", "12"),
			NewMatch("13", "14"),
			NewMatch("15", "16")),
		NewWeek("1 Jun", 3, 4, true,
			NewMatch("21", "22"),
			NewMatch("23", "24"),
			NewMatch("25", "26")),
		NewWeek("2 Jun", 5, 6, true,
			NewMatch("31", "32"),
			NewMatch("33", "34"),
			NewMatch("35", "36")),
	}
	it := weeks.Iterator()
	matches, hasNext := it.Next()
	assert.True(t, hasNext)
	assert.Equal(t, 9, len(matches))
	index := Incrementable(0)
	checkExpected(t, "11", "12", "31 May", 1, "A", matches[index.postInc()])
	checkExpected(t, "13", "14", "31 May", 2, "A", matches[index.postInc()])
	checkExpected(t, "15", "16", "31 May", 2, "B", matches[index.postInc()])
	checkExpected(t, "21", "22", "1 Jun", 3, "A", matches[index.postInc()])
	checkExpected(t, "23", "24", "1 Jun", 4, "A", matches[index.postInc()])
	checkExpected(t, "25", "26", "1 Jun", 4, "B", matches[index.postInc()])
	checkExpected(t, "31", "32", "2 Jun", 5, "A", matches[index.postInc()])
	checkExpected(t, "33", "34", "2 Jun", 6, "A", matches[index.postInc()])
	checkExpected(t, "35", "36", "2 Jun", 6, "B", matches[index.postInc()])
}


func TestIteratorAllIterations(t *testing.T) {
	weeks := FixtureWeekList{
		NewWeek("31 May", 1, 2, true,
			NewMatch("11", "12"),
			NewMatch("13", "14"),
			NewMatch("15", "16")),
		NewWeek("1 Jun", 3, 4, true,
			NewMatch("21", "22"),
			NewMatch("23", "24"),
			NewMatch("25", "26")),
		NewWeek("2 Jun", 5, 6, true,
			NewMatch("31", "32"),
			NewMatch("33", "34"),
			NewMatch("35", "36")),
	}
	it := weeks.Iterator()
	for i := 0; i <= 216; i++ {
		if _, ok := it.Next(); i == 216 {
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}
	}
}

func TestTeamSchedules(t *testing.T) {
	weeks := FixtureWeekList{
		NewWeek("31 May", 1, 2, true,
			NewMatch("11", "12"),
			NewMatch("13", "14"),
			NewMatch("15", "16")),
		NewWeek("1 Jun", 3, 4, true,
			NewMatch("11", "12"),
			NewMatch("13", "14"),
			NewMatch("15", "16")),
		NewWeek("2 Jun", 5, 6, true,
			NewMatch("31", "32"),
			NewMatch("33", "34"),
			NewMatch("35", "36")),
	}
	s, ok := weeks.Iterator().Next()
	assert.True(t, ok)
	ts := s.teamSchedules()
	assert.Equal(t, 12, len(ts))
	index := Incrementable(0)
	checkItem := func(ts *TeamSchedule, team string, matches ...*Match) {
		assert.Equal(t, team, ts.team)
		assert.Equal(t, len(matches), len(ts.matches))
		for i, m := range matches {
			assert.Equal(t, m.team1, ts.matches[i].team1)
			assert.Equal(t, m.team2, ts.matches[i].team2)
		}
	}
	checkItem(ts[index.postInc()], "11", weeks[0].matches[0], weeks[1].matches[0])
	checkItem(ts[index.postInc()], "12", weeks[0].matches[0], weeks[1].matches[0])
	checkItem(ts[index.postInc()], "13", weeks[0].matches[1], weeks[1].matches[1])
	checkItem(ts[index.postInc()], "14", weeks[0].matches[1], weeks[1].matches[1])
	checkItem(ts[index.postInc()], "15", weeks[0].matches[2], weeks[1].matches[2])
	checkItem(ts[index.postInc()], "16", weeks[0].matches[2], weeks[1].matches[2])
	checkItem(ts[index.postInc()], "31", weeks[2].matches[0])
	checkItem(ts[index.postInc()], "32", weeks[2].matches[0])
	checkItem(ts[index.postInc()], "33", weeks[2].matches[1])
	checkItem(ts[index.postInc()], "34", weeks[2].matches[1])
	checkItem(ts[index.postInc()], "35", weeks[2].matches[2])
	checkItem(ts[index.postInc()], "36", weeks[2].matches[2])
}

func TestTeamScheduleEvaluate(t *testing.T) {
	list := BuildFixtureList()
	fl, ok := list.Iterator().Next()
	assert.True(t, ok)
	checkItem := func(ts *TeamSchedule, length int, score int) {
		assert.Equal(t, length, len(ts.matches))
		assert.Equal(t, score, ts.evaluate())
	}
	checkItem(fl.teamSchedules()[0], 9, 63)
	checkItem(fl.teamSchedules()[32], 8, 132)
}

func TestScheduleEvaluate(t *testing.T) {
	list := BuildFixtureList()
	fl, ok := list.Iterator().Next()
	assert.True(t, ok)
	assert.Equal(t, 164, fl.Evaluate())
}

func TestRegexp(t *testing.T) {
	var buffer bytes.Buffer
	buffer.WriteString("143 114 2 8")
	r, _ := regexp.Compile("\\d+")
	data := r.FindAllString(buffer.String(), -1)
	islice := make([]int, len(data))
	for i, v := range data {
		islice[i], _ = strconv.Atoi(v)
	}
	assert.Equal(t, []int{143, 114, 2, 8}, islice)
}