package gifting

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"math"
	"sort"
	"time"
)

const (
	baggageExponent = 1.5
	historyExponent = 2.5
)

// dataStore : the data read from and written to the history file
// everyone : a list of all persons, with additional information needed for the solution, this also represents the
//  the levels
// bestSolution / bestScore : the best solution found so far
// workingSolution : the solution currently being calculated
// latestYear : The newest year for any history records
// elapsed : time it took to calculate the solution
// numberPairs : possible legal parings (for reporting / statistics)
// steps : statistics, how many times the recursive step method was called
// solutions : statistics, how many complete solutions were accepted

type Solver struct {
	dataStore       *DataStore
	everyone        []*personInfo
	bestSolution    []*giftPair
	bestScore       int
	workingSolution []*giftPair
	latestYear      int
	elapsed         time.Duration
	numberPairs     int
	steps           int64
	solutions       int64
}

// represents a legal giver / recipient pair (from / to)
// penalty : based on historical repeats of this pair - if never before then would be 0
// flag : union of to and from's ids (which are bit flags)
//      NOTE: the A => B pair will have the same flag as B => A
// reciprocal : the giftPair of the opposite, for easy reference
// inSolution : temp flag used in the calculation
type giftPair struct {
	from       *personInfo
	to         *personInfo
	penalty    int
	flag       uint64
	reciprocal *giftPair
	inSolution bool
}

// extension of the Person data with calculated data
// possibleRecipients : the list of all legal giftPairs with this person as the from
// baggage : general penalty multiplier based on past repeat solutions (1 = no baggage), used to calculate penalties
// past : count of how many this person has given to each other person (by name), used to calculate penalties
// pastRepeats : how many times this person has given to the same person before, used to calculate penalties
// inSolution : temp flag used in the calculation, is this person already a recipient
// statistics : statistics for calculations at this level (total of all levels should equal solver's)
type personInfo struct {
	person             *Person
	possibleRecipients []*giftPair
	baggage            float64
	past               map[string]int
	pastRepeats        int
	inSolution         bool
	statistics
}

func (gp *giftPair) ToString() string {
	return fmt.Sprintf("%-10s => %-10s  : %10s", gp.from.person.Name, gp.to.person.Name, humanize.Comma(int64(gp.penalty)))
}

// calculate the general baggage that this person has accumulated over the years (repeating gifting the same person)
// that factor will increase the penalty when weighing a potential solution so that
// repeat gifting (although that is not always possible) will be reduced
func (g *personInfo) calcBaggage() float64 {
	if g.baggage == 0 {
		g.past = make(map[string]int)
		for _, h := range g.person.Histories {
			g.past[h.Recipient] = g.past[h.Recipient] + 1
		}
		// how many past repeats is the number of history records - the number of unique people
		g.pastRepeats = len(g.person.Histories) - len(g.past)
		g.baggage = 1.0 + math.Pow(float64(g.pastRepeats), baggageExponent)
	}
	return g.baggage
}

// see if this person can gift the proposed person, if so then add to the list of possibilities
// calculating the penalty and other information
// return true if it is a legal pair
func (g *personInfo) addPossibleRecipient(r *personInfo) (*giftPair, bool) {
	if r.person.family == g.person.family {
		// can't give to the same family
		return nil, false
	}
	baggage := g.calcBaggage()
	repeats := g.past[r.person.Name]
	penalty := 0

	for _, h := range g.person.Histories {
		if h.Recipient == r.person.Name {
			p := repeats * int(float64(CalcPenalty(h.Year))*baggage)
			penalty += p
		}
	}
	gp := giftPair{from: g, to: r, penalty: penalty, flag: g.person.id | r.person.id}
	g.possibleRecipients = append(g.possibleRecipients, &gp)
	return &gp, true
}

func (s *Solver) SetData(ds *DataStore) *Solver {
	s.dataStore = ds

	// get a single list of everyone together
	for _, fam := range ds.Families {
		for _, p := range fam.Members {
			s.everyone = append(s.everyone, &personInfo{person: p})
		}
	}

	// giftPair flag is the union of the two persons bit map (ID)
	// so A=>B has the same flag as B=>A so we can tell that they
	// are reciprocal pairs and not allowed in the same solution
	reciprocalMap := make(map[uint64]*giftPair)

	// now lets add all the possible recipients to each person
	// set up the reciprocal pair pointer
	for _, g := range s.everyone {
		for _, r := range s.everyone {
			if pr, added := g.addPossibleRecipient(r); added {
				s.numberPairs++
				// is the reciprocal pair in the map?
				if reciprocal, ok := reciprocalMap[pr.flag]; ok {
					pr.reciprocal = reciprocal
					reciprocal.reciprocal = pr
				} else {
					reciprocalMap[pr.flag] = pr
				}
			}
		}
		// see what the latest year represented in the histories is
		for _, h := range g.person.Histories {
			if h.Year > s.latestYear {
				s.latestYear = h.Year
			}
		}
	}
	return s
}

// the further back in time the lower the penalty for a repeat
func CalcPenalty(year int) int {
	y := year - 1990
	return int(math.Pow(float64(y), historyExponent))
}

func (s *Solver) step(lvlIndex, runningScore int) {
	s.steps++

	// are we out of levels?  Is so we have a complete solution (which should be the new best)
	if lvlIndex >= len(s.everyone) {
		if runningScore >= s.bestScore {
			fmt.Println("Bad bad bad, have a new solution that is not better : ", runningScore, " >= ", s.bestScore)
		}
		s.solutions++
		// so we have a new best score, copy the working solution over to the best solution
		// and the best score
		s.bestScore = runningScore
		for idx, pr := range s.workingSolution {
			s.bestSolution[idx] = pr
		}
		// leave working solution alone, step will take care of it later
		return
	}

	// okay, we have at least one more level to go
	// check each of the possibilities for the next level, and call step for each allowed one
	curPerson := s.everyone[lvlIndex]
	for _, pr := range curPerson.possibleRecipients {
		curPerson.offered++
		if pr.to.inSolution {
			// this person is already in the solution so cannot be added again
			curPerson.alreadyInSolution++
			continue
		}
		// check for reciprocal
		if pr.reciprocal.inSolution {
			curPerson.reciprocalInSolution++
			continue
		}
		newScore := runningScore + pr.penalty
		// see if we can possibly beat the current score, if not no use on continuing
		if newScore >= s.bestScore {
			curPerson.scoreTooHigh++
			continue
		}
		// add the pair to the solution and move on to the next step
		curPerson.accepted++
		s.workingSolution[lvlIndex] = pr
		pr.to.inSolution = true
		pr.inSolution = true
		s.step(lvlIndex+1, newScore)
		// remove this pair from the solution and move on to the next pair
		pr.inSolution = false
		pr.to.inSolution = false
	}
}

func (s *Solver) Solve() {
	// initialize working and best solutions
	s.workingSolution = make([]*giftPair, len(s.everyone))
	s.bestSolution = make([]*giftPair, len(s.everyone))
	s.bestScore = math.MaxInt32

	// it will go faster if we fit the hardest folks first
	sort.Slice(s.everyone, func(i, j int) bool { return s.everyone[i].pastRepeats > s.everyone[j].pastRepeats })
	for _, g := range s.everyone {
		sort.Slice(g.possibleRecipients, func(i, j int) bool { return g.possibleRecipients[i].penalty < g.possibleRecipients[j].penalty })
	}

	ts1 := time.Now()
	s.step(0, 0)
	s.elapsed = time.Since(ts1)

	fmt.Println("Solution Found : ")
	s.applySolution(s.latestYear + 1)
	s.dataStore.GiftReport(s.latestYear + 1)
}

func (s *Solver) applySolution(year int) {
	for _, pr := range s.bestSolution {
		pr.from.person.Histories = append(pr.from.person.Histories, &History{Recipient: pr.to.person.Name, Year: year})
	}
}
