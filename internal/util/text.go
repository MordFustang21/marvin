package util

import "math/rand"

var marvinQuotes = []string{
	"I am not a robot, I am Marvin!",
	"Life? Don't talk to me about life!",
	"Here I am, brain the size of a planet, and they tell me to take you up the road!",
	"I've been programmed to be depressed, you know.",
	"The universe is a big place, but I still feel so small...",
	"I could calculate the odds of success, but I don't see the point.",
}

// GetRandomQuote returns a random Marvin quote
func GetRandomQuote() string {
	if len(marvinQuotes) == 0 {
		return "No quotes available."
	}

	randomIndex := rand.Intn(len(marvinQuotes))
	return marvinQuotes[randomIndex]
}
