package gifting

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"io"
)

// collector of calculating statistics
//
// offered : number of items looked at
// alreadyInSolution : offered recipient already in solution
// reciprocalInSolution : reciprocal pair already in solution (if A=>B is in solution B=>A cannot be)
// scoreTooHigh : adding this pair would push the partial solution past the current lowest solution
// accepted : passed everything else, move on
// offered = alreadyInSolution + reciprocalInSolution + scoreTooHigh + accepted
type statistics struct {
	offered              int64
	alreadyInSolution    int64
	reciprocalInSolution int64
	scoreTooHigh         int64
	accepted             int64
}

func (s *statistics) add(o *statistics) {
	s.offered += o.offered
	s.alreadyInSolution += o.alreadyInSolution
	s.reciprocalInSolution += o.reciprocalInSolution
	s.scoreTooHigh += o.scoreTooHigh
	s.accepted += o.accepted
}

func (s *statistics) oneLine() string {
	return fmt.Sprintf("%12s  %12s  %12s  %12s  %12s",
		humanize.Comma(s.offered),
		humanize.Comma(s.alreadyInSolution),
		humanize.Comma(s.reciprocalInSolution),
		humanize.Comma(s.scoreTooHigh),
		humanize.Comma(s.accepted))
}

func (s *statistics) report(lenTitle, lenValue int, out io.Writer) {
	fmt.Fprintf(out, "%-*s : %*s\n", lenTitle, "Offered", lenValue, humanize.Comma(s.offered))
	fmt.Fprintf(out, "%-*s : %*s\n", lenTitle, "Already In Solution", lenValue, humanize.Comma(s.alreadyInSolution))
	fmt.Fprintf(out, "%-*s : %*s\n", lenTitle, "Reciprocal", lenValue, humanize.Comma(s.reciprocalInSolution))
	fmt.Fprintf(out, "%-*s : %*s\n", lenTitle, "Score Too High", lenValue, humanize.Comma(s.scoreTooHigh))
	fmt.Fprintf(out, "%-*s : %*s\n", lenTitle, "Accepted", lenValue, humanize.Comma(s.accepted))
}
