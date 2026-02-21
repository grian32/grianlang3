package util

import "fmt"

type Position struct {
	StartLine uint32
	StartCol  uint32
	EndLine   uint32
	EndCol    uint32
}

type PositionError struct {
	Position *Position
	Msg      string
}

func (pe *PositionError) String() string {
	return fmt.Sprintf("%d:%d-%d:%d: %s", pe.Position.StartLine, pe.Position.StartCol, pe.Position.EndLine, pe.Position.EndCol, pe.Msg)
}
