package rfutils

func GetTimeout(timeout int) int {
	if timeout <= 0 {
		return 2 * 60
	}
	return timeout
}
