package androidutil

import (
	"fmt"
	"os/exec"
	"strings"
)

func StartActivity(activity string) error {
	res, err := exec.Command("am", "start", "--activity-single-top", activity).Output()

	if err != nil {
		return err
	}

	output := string(res)

	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		if strings.HasPrefix(strings.ToLower(line), "error:") {
			return fmt.Errorf("failed to start activity: %s", line)
		}
	}

	return nil
}

func GetForegroundWindow() (window string, err error) {

	res, err := exec.Command("dumpsys", "window", "windows").Output()
	if err != nil {
		return "", err
	}

	output := string(res)

	for _, line := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {

		line = strings.TrimPrefix(line, " ")
		if strings.HasPrefix(line, "mCurrentFocus=") {
			line = strings.TrimSuffix(line, "}")
			elements := strings.Split(line, " ")
			if len(elements) != 3 {
				return "", fmt.Errorf("failed to parse dumpsys output: %s", line)
			}
			return elements[2], nil
		}
	}
	return "", fmt.Errorf("failed to find focused window in dumpsys output")
}
