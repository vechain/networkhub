package genesis

// GalacticaGenesis is user customized genesis
type GalacticaGenesis struct {
	PreCoefForkGenesis
}

type GalacticaGenesisConfig struct {
	PreCoefForkConfig
	GALACTICA uint32
}
