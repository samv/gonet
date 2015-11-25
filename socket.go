package network

type Socket struct {
	bind func() int
}
