package preset

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/vechain/networkhub/environments"
	"github.com/vechain/networkhub/network"
	"github.com/vechain/networkhub/network/node"
	"github.com/vechain/thor/v2/genesis"
	"github.com/vechain/thor/v2/thor"
)

var LocalSixNodesNetwork = &network.Network{
	ID:          "sixNodesNetwork",
	Environment: environments.Local,
	Nodes: []*node.Node{
		{
			ID:            "node1",
			ExecArtifact:  "/app/thor",
			Genesis:       localSixNodesNetworkGenesis,
			P2PListenPort: 8081,
			APIAddr:       "0.0.0.0:8181",
			APICORS:       "*",
			Type:          node.MasterNode,
			Verbosity:     4,
			Key:           "b2c859e115ef4a3f5e4d32228b41de4c661c527a32f723ac37745bf860fd09cb", // 0x5F90f56c7b87E3d1acf9437f0E43E4d687AcEB7e
			Enode:         "enode://7f73b785e7c288a1c246b34f9ac4e323b53d759f1988d580dc6829b6a84cbdb0d315d65669b2312f2c60c940d7680badfe6c208be93ce8a8cfe063a034382328@127.0.0.1:8081",
		},
		{
			ID:            "node2",
			ExecArtifact:  "/app/thor",
			Genesis:       localSixNodesNetworkGenesis,
			P2PListenPort: 8082,
			APIAddr:       "0.0.0.0:8182",
			APICORS:       "*",
			Type:          node.MasterNode,
			Verbosity:     4,
			Key:           "4de650ca1c8beae4ed6a4358087f50c01b51f5c0002ae9836c55039ca9818d0c", // 0x5c29518F6a6124a2BeE89253347c8295f604710A
			Enode:         "enode://4f253253dfe997eb8b20f5ecfc25f6d3a646efd850f762c40947981c811d4553702b507589083144ae0784947b1abffe85ea533a81a90aa43637549104b4885c@127.0.0.1:8082",
		},
		{
			ID:            "node3",
			ExecArtifact:  "/app/thor",
			Genesis:       localSixNodesNetworkGenesis,
			P2PListenPort: 8083,
			APIAddr:       "0.0.0.0:8183",
			APICORS:       "*",
			Type:          node.RegularNode,
			Key:           "1b310ea04afd6d14a8f142158873fc70bfd4ba12a19138cc5b309fce7c77105e", // 0x1b1c0055065b3ADee4B9a9e8297142Ba2cD34EfE
			Enode:         "enode://dfc6eaf3fd295980d74f8af726a66b849b1d017b7a8145570edb8a86744add5d1551bee4199cf3bc685ad2fb7824b8f36ecdd997100f3750dd474c1fcecdd4a7@127.0.0.1:8083",
		},
		{
			ID:            "node4",
			ExecArtifact:  "/app/thor",
			Genesis:       localSixNodesNetworkGenesis,
			P2PListenPort: 8084,
			APIAddr:       "0.0.0.0:8184",
			APICORS:       "*",
			Type:          node.MasterNode,
			Verbosity:     4,
			Key:           "c70dda88e779df10abbc7c5d37fbb3478c5cf8df2a70d6b0bfc551a5a9a17359", // 0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c
			Enode:         "enode://45bc18d543b6ae66b83bb6360e7d552f7749cc9143d00bd61c5665f7e9dcdd2287690b86737353d12ee1a3dd419ff828eb89b17585c9357d7833e8d6e9c6a2b1@127.0.0.1:8084",
		},
		{
			ID:            "node5",
			ExecArtifact:  "/app/thor",
			Genesis:       localSixNodesNetworkGenesis,
			P2PListenPort: 8085,
			APIAddr:       "0.0.0.0:8185",
			APICORS:       "*",
			Type:          node.MasterNode,
			Verbosity:     4,
			Key:           "ade54b623a4f4afc38f962a85df07a428204a67cee0c9b43a99ca255fd2fb9a6", // 0x0aeC31606e217895696771961de416Efa185Be66
			Enode:         "enode://61fd9f3ae7e8daa2ced88a85cec097ad68e6c99af2688523cec0b2aab5b77f6a93bc138cafe97cb6b09cecf5248332669beac02add1eb6f45e1950597696a564@127.0.0.1:8085",
		},
		{
			ID:            "node6",
			ExecArtifact:  "/app/thor",
			Genesis:       localSixNodesNetworkGenesis,
			P2PListenPort: 8086,
			APIAddr:       "0.0.0.0:8186",
			APICORS:       "*",
			Type:          node.RegularNode,
			Key:           "92ad65923d6782a43e6a1be01a8e52bce701967d78937e73da746a58f293ba30", // 0x9C2871C411CCe579B987E9b932C484dA8b901075
			Enode:         "enode://ec9c4b2066b63b10f5b255f08967cb5f7829ee4583b0aed0cea29b452f235e7bcc80b778970a3079929a26abaab28b01ece33f48ca339061c2b6c0526fad557d@127.0.0.1:8086",
		},
	},
}

var localSixNodesNetworkGenesis = &genesis.CustomGenesis{
	LaunchTime: 1703180212,
	GasLimit:   10000000,
	ExtraData:  "Local Six Nodes Network",
	Accounts: []genesis.Account{
		{
			Address: thor.MustParseAddress("0x7567d83b7b8d80addcb281a71d54fc7b3364ffed"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(big.NewInt(0)),
			Code:    "0x6060604052600256",
			Storage: map[string]thor.Bytes32{
				"0x0000000000000000000000000000000000000000000000000000000000000001": thor.MustParseBytes32("0x0000000000000000000000000000000000000000000000000000000000000002"),
			},
		},
		{
			Address: thor.MustParseAddress("0x5F90f56c7b87E3d1acf9437f0E43E4d687AcEB7e"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x5c29518F6a6124a2BeE89253347c8295f604710A"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x1b1c0055065b3ADee4B9a9e8297142Ba2cD34EfE"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x042306e116Dc301ecd7b83a04F4c8277Fbe41b6c"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x0aeC31606e217895696771961de416Efa185Be66"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0xf077b491b355E64048cE21E3A6Fc4751eEeA77fa"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x435933c8064b4Ae76bE665428e0307eF2cCFBD68"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},

		{
			Address: thor.MustParseAddress("0x0F872421Dc479F3c11eDd89512731814D0598dB5"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0xF370940aBDBd2583bC80bfc19d19bc216C88Ccf0"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x99602e4Bbc0503b8ff4432bB1857F916c3653B85"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x61E7d0c2B25706bE3485980F39A3a994A8207aCf"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x361277D1b27504F36a3b33d3a52d1f8270331b8C"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		}, {
			Address: thor.MustParseAddress("0xD7f75A0A1287ab2916848909C8531a0eA9412800"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0xAbEf6032B9176C186F6BF984f548bdA53349f70a"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
		{
			Address: thor.MustParseAddress("0x865306084235Bf804c8Bba8a8d56890940ca8F0b"),
			Balance: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
			Energy:  convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		},
	},
	Authority: []genesis.Authority{
		{
			MasterAddress:   thor.MustParseAddress("0x5F90f56c7b87E3d1acf9437f0E43E4d687AcEB7e"),
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
	Params: genesis.Params{
		RewardRatio:         convToHexOrDecimal256(big.NewInt(300000000000000000)),
		BaseGasPrice:        convToHexOrDecimal256(big.NewInt(1000000000000000)),
		ProposerEndorsement: convToHexOrDecimal256(new(big.Int).SetBytes(hexutil.MustDecode("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))),
		ExecutorAddress:     &localThreeMasterEndorser,
	},
	Executor: genesis.Executor{
		Approvers: []genesis.Approver{
			{
				Address:  thor.MustParseAddress("0x199b836d8a57365baccd4f371c1fabb7be77d389"),
				Identity: thor.MustParseBytes32("0x00000000000067656e6572616c20707572706f736520626c6f636b636861696e"),
			},
		},
	},
	ForkConfig: &thor.ForkConfig{
		VIP191:    0,
		ETH_CONST: 0,
		BLOCKLIST: 0,
		ETH_IST:   0,
		VIP214:    0,
	},
}
