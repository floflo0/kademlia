package entities

import "fmt"

type Address struct {
	IP   string
	Port int
}

func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}
