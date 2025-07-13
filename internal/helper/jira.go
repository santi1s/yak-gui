package helper

import "fmt"

func JiraLink(ticket string) string {
	return fmt.Sprintf("https://doctolib.atlassian.net/browse/%s", ticket)
}
