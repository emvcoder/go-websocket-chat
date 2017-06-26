package main

import "regexp"

func LookForMatch(url string) string {
	match, _ := regexp.MatchString("templates", url)
	if match == true {
		return url
	}
	return "/error"
}