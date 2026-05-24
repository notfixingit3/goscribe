package hello

// Greeter builds greeting messages.
type Greeter struct {
	prefix string
}

// NewGreeter creates a Greeter with the given prefix.
func NewGreeter(prefix string) *Greeter {
	return &Greeter{prefix: prefix}
}

// Greet returns a greeting for the named person.
func (g *Greeter) Greet(name string) string {
	return g.prefix + name
}

// Farewell returns a farewell message.
func (g *Greeter) Farewell(name string) string {
	return "Goodbye, " + name
}
