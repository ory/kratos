package recovery

type State string

const (
	StateChooseMethod    State = "choose_method"
	StateEmailSent       State = "sent_email"
	StatePassedChallenge State = "passed_challenge"
)

var states = []State{StateChooseMethod, StateEmailSent, StatePassedChallenge}

func indexOf(current State) int {
	for k, s := range states {
		if s == current {
			return k
		}
	}
	return 0
}

func HasReachedState(expected, actual State) bool {
	return indexOf(actual) >= indexOf(expected)
}

func NextState(current State) State {
	if current == StatePassedChallenge {
		return StatePassedChallenge
	}

	return states[indexOf(current)+1]
}
