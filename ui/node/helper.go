package node

func PasswordLoginOrder(in []string) []string {
	if len(in) == 0 {
		return []string{"password"}
	}
	if len(in) == 1 {
		return append(in, "password")
	}
	return append([]string{in[0], "password"}, in[1:]...)
}
