package gifting

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"math/big"
	"os"
)

func (s *Solver) PossibleSolutions() *big.Int {
	possibleSolutions := big.NewInt(1)
	for _, g := range s.everyone {
		possibleSolutions = possibleSolutions.Mul(possibleSolutions, big.NewInt(int64(len(g.possibleRecipients))))
	}
	return possibleSolutions
}

func (s *Solver) RepeatPenaltyReport() {
	for y := 2015; y > 1996; y-- {
		fmt.Printf("%4d  :  %7d\n", y, CalcPenalty(y))
	}
}

func (s *Solver) BaggageReport() {
	fmt.Printf("%-10s  %s\n", "Name", "baggage")
	for _, g := range s.everyone {
		fmt.Printf("%-10s  %8.2f\n", g.person.Name, g.baggage)
	}
}

const (
	lenTitle = 25
	lenValue = 40
)

func (s *Solver) Statistics() {

	fmt.Println()
	fmt.Println()

	histories := 0
	years := 0
	stats := statistics{}
	for _, g := range s.everyone {
		stats.add(&g.statistics)
		h := len(g.person.Histories)
		histories += h
		if years < h {
			years = h
		}
	}

	fmt.Println("Some Hiistory")
	fmt.Printf("%-*s : %*s\n", lenTitle, "Year", lenValue, humanize.Ordinal(years+1))
	fmt.Printf("%-*s : %*d\n", lenTitle, "Total Gifts", lenValue, int64(histories))
	fmt.Println()

	fmt.Println("Calculating this year")
	fmt.Printf("%-*s : %*s\n", lenTitle, "Time to calculate", lenValue, s.elapsed.String())
	fmt.Printf("%-*s : %*d\n", lenTitle, "Folks", lenValue, int64(len(s.everyone)))
	fmt.Printf("%-*s : %*d\n", lenTitle, "Families", lenValue, int64(len(s.dataStore.Families)))
	fmt.Printf("%-*s : %*d\n", lenTitle, "Possible Pairs", lenValue, int64(s.numberPairs))
	fmt.Printf("%-*s : %*s\n", lenTitle, "Possible Solutions", lenValue, humanize.BigComma(s.PossibleSolutions()))
	fmt.Println()

	fmt.Println("Statistics")
	stats.report(lenTitle, lenValue, os.Stdout)
	fmt.Printf("%-*s : %*d\n", lenTitle, "Solutions considered", lenValue, int64(s.solutions))
	fmt.Println()
	fmt.Println("Details")
	fmt.Printf("    %12s  %12s  %12s  %12s  %12s\n", "Offered", "In Solution", "Reciprocal", "Score", "Accepted")

	for idx, g := range s.everyone {
		fmt.Printf("%2d  %s\n", idx, g.statistics.oneLine())
	}
}
