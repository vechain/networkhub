package genesis

// PostCoefForkGenesis is user customized genesis
type PostCoefForkGenesis struct {
	PreCoefForkGenesis
}

type PostCoefForkConfig struct {
	PreCoefForkConfig
	VIPGASCOEF uint32
}
