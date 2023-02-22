// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"fmt"
	"strings"

	"github.com/dalzilio/hue/pkg/formula"
	"github.com/dalzilio/hue/pkg/pnml"
)

// State is the type of High-Level Nets (hlnet) markings. It is a slice
// containing the Marking of each places in the same order than in Places. We
// also use hashmaps to have the "compound" marking for each place (useful for
// computing tokens-count) and the list of enabled transitions (for
// is-fireable), which helps simplify the evaluation of formulas. When we find
// that a transition is enabled, we keep the result of firing this transition in
// Next.
type State struct {
	COL     pnml.Marking            // Colored marking
	PT      map[string]int          // Compound P/T marking
	Enabled map[string]formula.Bool // list of enabled transition (after ComputeEnabled)
	After   []pnml.Marking
}

// NewState returns a fresh State value whose colored marking is m, but without
// any information on enabled transitions.
func (s *Stepper) newState(m pnml.Marking) *State {
	res := State{
		COL: m,
		PT:  make(map[string]int),
	}
	for k, v := range s.Places {
		res.PT[v.Name] = m[k].Sum()
	}
	return &res
}

// ----------------------------------------------------------------------

func (s *Stepper) PrintEnabled() string {
	if s.Enabled == nil {
		return "(unknown)"
	}
	res := ""
	for k, t := range s.Trans {
		if _, ok := s.forbidEnabled[k]; ok {
			res += fmt.Sprintf("%s(?) ", t.Name)
			continue
		}
		res += fmt.Sprintf("%s(%s) ", t.Name, s.Enabled[t.Name].PrintShort())
	}
	return res + "\n"
}

func (net *Net) printMarkingAligned(m State, left int, trunca, truncb int) string {
	s := ""
	for k, v := range net.Places {
		mm := []rune(net.PrintHue(m.COL[k]))
		if len(mm) > trunca+truncb+10 {
			ma, mb := mm[:trunca], mm[len(mm)-truncb:]
			mm = append(append(ma, []rune(".....")...), mb...)
		}
		s += fmt.Sprintf("%s%s : %s\n", v.Name, strings.Repeat(" ", left-len(v.Name)), string(mm))
	}
	return s
}
