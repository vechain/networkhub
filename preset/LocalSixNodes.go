package preset

import (
	"encoding/json"
	"math/big"

	"github.com/vechain/networkhub/internal/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/networkhub/network/node/genesis"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/thor/v2/thor"

	thorgenesis "github.com/vechain/thor/v2/genesis"
)

var (
	SixNNAccount1 = common.NewAccount("b2c859e115ef4a3f5e4d32228b41de4c661c527a32f723ac37745bf860fd09cb") // 0x5F90f56c7b87E3d1acf9437f0E43E4d687AcEB7e
	SixNNAccount2 = common.NewAccount("4de650ca1c8beae4ed6a4358087f50c01b51f5c0002ae9836c55039ca9818d0c") // 0x5c29518F6a6124a2BeE89253347c8295f604710A
	SixNNAccount3 = common.NewAccount("1b310ea04afd6d14a8f142158873fc70bfd4ba12a19138cc5b309fce7c77105e") // 0x1b1c0055065b3ADee4B9a9e8297142Ba2cD34EfE
	SixNNAccount4 = common.NewAccount("c70dda88e779df10abbc7c5d37fbb3478c5cf8df2a70d6b0bfc551a5a9a17359") // 0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c
	SixNNAccount5 = common.NewAccount("ade54b623a4f4afc38f962a85df07a428204a67cee0c9b43a99ca255fd2fb9a6") // 0x0aeC31606e217895696771961de416Efa185Be66
	SixNNAccount6 = common.NewAccount("92ad65923d6782a43e6a1be01a8e52bce701967d78937e73da746a58f293ba30") // 0x9C2871C411CCe579B987E9b932C484dA8b901075
)

func LocalSixNodesNetwork() *network.Network {
	return LocalSixNodesNetworkWithGenesis(LocalSixNodesNetworkGenesis())
}

func LocalSixNodesNetworkGenesis() *genesis.CustomGenesis {
	return &genesis.CustomGenesis{
		LaunchTime: 1703180212,
		GasLimit:   10000000,
		ExtraData:  "Local Six Nodes Network",
		Accounts: []thorgenesis.Account{
			{
				Address: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"), // dce1443bd2ef0c2631adc1c67e5c93f13dc23a41c18b536effbbdcbcdb96fb65
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount1.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount2.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount3.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount4.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount5.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
			{
				Address: *SixNNAccount6.Address,
				Balance: convToHexOrDecimal256(LargeBigValue),
				Energy:  convToHexOrDecimal256(LargeBigValue),
			},
		},
		Authority: []thorgenesis.Authority{
			{
				MasterAddress:   *SixNNAccount1.Address,
				EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
				Identity:        thor.MustParseBytes32("0x000000000000000068747470733a2f2f636f6e6e65782e76656368612e696e2f"),
			},
			{
				MasterAddress:   thor.MustParseAddress("0x5c29518F6a6124a2BeE89253347c8295f604710A"),
				EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
				Identity:        thor.MustParseBytes32("0x000000000000000068747470733a2f2f656e762e7665636861696e2e6f72672f"),
			},
			{
				MasterAddress:   thor.MustParseAddress("0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c"),
				EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
				Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
			},
			{
				MasterAddress:   thor.MustParseAddress("0x0aeC31606e217895696771961de416Efa185Be66"),
				EndorsorAddress: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
				Identity:        thor.MustParseBytes32("0x0000000000000068747470733a2f2f617070732e7665636861696e2e6f72672f"),
			},
		},
		Params: thorgenesis.Params{
			RewardRatio:         convToHexOrDecimal256(big.NewInt(300000000000000000)),
			BaseGasPrice:        convToHexOrDecimal256(big.NewInt(1000000000000000)),
			ProposerEndorsement: convToHexOrDecimal256(LargeBigValue),
			ExecutorAddress:     &localThreeMasterEndorser,
		},
		Executor: thorgenesis.Executor{
			Approvers: []thorgenesis.Approver{
				{
					Address:  thor.MustParseAddress("0x199b836d8a57365baccd4f371c1fabb7be77d389"),
					Identity: thor.MustParseBytes32("0x00000000000067656e6572616c20707572706f736520626c6f636b636861696e"),
				},
			},
		},
		ForkConfig: &genesis.CustomGenesisForkConfig{
			ForkConfig: thor.ForkConfig{
				VIP191:    0,
				ETH_CONST: 0,
				BLOCKLIST: 0,
				ETH_IST:   0,
				VIP214:    0,
			},
		},
	}
}

func LocalSixNodesNetworkWithGenesis(genesis *genesis.CustomGenesis) *network.Network {
	return &network.Network{
		BaseID:      "sixNodesNetwork",
		Environment: environments.Local,
		Nodes: []node.Config{
			&node.BaseNode{
				ID:            "node1",
				P2PListenPort: 8061,
				APIAddr:       "127.0.0.1:8161",
				Key:           "b2c859e115ef4a3f5e4d32228b41de4c661c527a32f723ac37745bf860fd09cb", // 0x5F90f56c7b87E3d1acf9437f0E43E4d687AcEB7e
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node2",
				P2PListenPort: 8062,
				APIAddr:       "127.0.0.1:8162",
				Key:           "4de650ca1c8beae4ed6a4358087f50c01b51f5c0002ae9836c55039ca9818d0c", // 0x5c29518F6a6124a2BeE89253347c8295f604710A
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node3",
				P2PListenPort: 8063,
				APIAddr:       "127.0.0.1:8163",
				Key:           "1b310ea04afd6d14a8f142158873fc70bfd4ba12a19138cc5b309fce7c77105e", // 0x1b1c0055065b3ADee4B9a9e8297142Ba2cD34EfE
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node4",
				P2PListenPort: 8064,
				APIAddr:       "127.0.0.1:8164",
				Key:           "c70dda88e779df10abbc7c5d37fbb3478c5cf8df2a70d6b0bfc551a5a9a17359", // 0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node5",
				P2PListenPort: 8065,
				APIAddr:       "127.0.0.1:8165",
				Key:           "ade54b623a4f4afc38f962a85df07a428204a67cee0c9b43a99ca255fd2fb9a6", // 0x0aeC31606e217895696771961de416Efa185Be66
				Genesis:       genesis,
			},
			&node.BaseNode{
				ID:            "node6",
				P2PListenPort: 8066,
				APIAddr:       "127.0.0.1:8166",
				Key:           "92ad65923d6782a43e6a1be01a8e52bce701967d78937e73da746a58f293ba30", // 0x9C2871C411CCe579B987E9b932C484dA8b901075
				Genesis:       genesis,
			},
		},
	}
}

func LocalSixNodesNetworkCustomGenesis(customGenesisJson string) (*genesis.CustomGenesis, error) {
	var customGenesis *genesis.CustomGenesis
	err := json.Unmarshal([]byte(customGenesisJson), &customGenesis)
	return customGenesis, err
}
