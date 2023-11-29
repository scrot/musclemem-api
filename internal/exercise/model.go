package exercise

// Exercise contains details of a single workout exercise
// an exercise is a node in a linked list, determining the order
type Exercise struct {
	ID          int
	Workout     int     `json:"workout-id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Weight      float64 `json:"weight" validate:"required"`
	Repetitions int     `json:"repetitions" validate:"required"`
	NextID      int
	PreviousID  int
}

// Next returns next Exercise node in linked list
// returns empty exercise if already at the tail
func (e Exercise) Next(r Retreiver) Exercise {
	x, err := r.WithID(e.NextID)
	if err != nil {
		return Exercise{}
	}
	return x
}

// Previous returns previous Exercise node in linked list
// returns empty exercise if already at the head
func (e Exercise) Previous(r Retreiver) Exercise {
	x, err := r.WithID(e.PreviousID)
	if err != nil {
		return Exercise{}
	}
	return x
}
