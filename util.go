package main

func getStatusCode(s string) int {
	// use icinga like status codes
	switch s {
	case "GREEN":
		return 0
	case "YELLOW":
		return 1
	case "RED":
		return 2
	case "GRAY":
		return 3
	default:
		// should not happen
		return -1
	}
}
