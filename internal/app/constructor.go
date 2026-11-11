package app

type BoardBuilder struct {
	Unused        int
	Rows          []string
	Cols          []int
	Sides         []string
	Board         map[string]int
	BoardTemplate map[string]map[int]string
}

var (
	Player2 = 2
	Player1 = 1
)

func NewBoardBuilder() *BoardBuilder {
	return &BoardBuilder{
		0,
		[]string{"a", "b", "c"},
		[]int{1, 2, 3},
		[]string{"X", "O"},
		map[string]int{
			"a1": 0, "a2": 0, "a3": 0,
			"b1": 0, "b2": 0, "b3": 0,
			"c1": 0, "c2": 0, "c3": 0,
		},
		map[string]map[int]string{
			"a": {1: "_", 2: "_", 3: "_"},
			"b": {1: "_", 2: "_", 3: "_"},
			"c": {1: "_", 2: "_", 3: "_"},
		},
	}
}
