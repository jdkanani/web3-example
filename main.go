package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/privval"

	"github.com/jdkanani/go-web3-example/contracts/rootchain"
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, secp256k1.PubKeyAminoRoute, nil)
	cdc.RegisterConcrete(secp256k1.PrivKeySecp256k1{}, secp256k1.PrivKeyAminoRoute, nil)
}

func main() {
	client, err := ethclient.Dial("https://kovan.infura.io")
	if err != nil {
		log.Fatal(err)
	}

	// with no 0x
	rootchainAddress := "24e01716a6ac34d5f2c4c082f553d86a557543a7"

	// token address
	tokenAddress := "670568761764f53E6C10cd63b71024c31551c9EC"

	// root chain client
	rootchainClient, err := rootchain.NewRootchain(common.HexToAddress(rootchainAddress), client)
	if err != nil {
		panic(err)
	}

	// load default base config
	privVal := privval.LoadFilePV(config.DefaultBaseConfig().PrivValidatorFile())

	// retrieve private key
	var pkObject secp256k1.PrivKeySecp256k1
	cdc.MustUnmarshalBinaryBare(privVal.PrivKey.Bytes(), &pkObject)

	// create ecdsa private key
	ecdsaPrivateKey, err := crypto.ToECDSA(pkObject[:])
	if err != nil {
		panic(err)
	}

	// from address
	fromAddress := common.BytesToAddress(privVal.Address)

	// fetch gas price
	gasprice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}

	// fetch nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		panic(err)
	}

	// create auth
	auth := bind.NewKeyedTransactor(ecdsaPrivateKey)
	auth.GasPrice = gasprice
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)

	// Calling contract method
	var amount big.Int
	amount.SetUint64(0)
	tx, err := rootchainClient.Deposit(auth, common.HexToAddress(tokenAddress), common.BytesToAddress(privVal.Address.Bytes()), &amount)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Pending TX: 0x%x\n", tx.Hash())
}
