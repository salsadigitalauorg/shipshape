package fact

func GatherAllFacts(factPlugins []Facter) {
	for _, p := range factPlugins {
		p.Gather()
	}
}
