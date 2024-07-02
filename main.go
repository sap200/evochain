package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/sap200/evochain/blockchain"
	"github.com/sap200/evochain/blockchainserver"
	"github.com/sap200/evochain/constants"
	"github.com/sap200/evochain/walletserver"
)

func init() {
	log.SetPrefix(constants.BLOCKCHAIN_NAME + ":")
}

func main() {

	chainCmdSet := flag.NewFlagSet("chain", flag.ExitOnError)
	walletCmdSet := flag.NewFlagSet("wallet", flag.ExitOnError)

	chainPort := chainCmdSet.Uint64("port", 5000, "HTTP port to launch our blockchain server")
	chainMiner := chainCmdSet.String("miners_address", "", "Miners address to credit mining reward")
	remoteNode := chainCmdSet.String("remote_node", "", "Remote Node from where the blockchain will be synced")

	walletPort := walletCmdSet.Uint64("port", 8080, "HTTP port to launch our wallet server")
	blockchainNodeAddress := walletCmdSet.String("node_address", "http://127.0.0.1:5000", "Blockchain node address for the wallet gateway")

	if len(os.Args) < 2 {
		fmt.Println("Error:Expected chain or wallet subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "chain":
		var wg sync.WaitGroup

		chainCmdSet.Parse(os.Args[2:])
		if chainCmdSet.Parsed() {
			if *chainMiner == "" || chainCmdSet.NFlag() == 0 {
				fmt.Println("Usage of chain subcommand: ")
				chainCmdSet.PrintDefaults()
				os.Exit(1)
			}

			if *remoteNode == "" {

				genesisBlock := blockchain.NewBlock("0x0", 0, 0)
				blockchain1 := blockchain.NewBlockchain(*genesisBlock, "http://127.0.0.1:"+strconv.Itoa(int(*chainPort)))
				blockchain1.Peers[blockchain1.Address] = true
				bcs := blockchainserver.NewBlockchainServer(*chainPort, blockchain1)
				wg.Add(4)
				go bcs.Start()
				go bcs.BlockchainPtr.ProofOfWorkMining(*chainMiner)
				go bcs.BlockchainPtr.DialAndUpdatePeers()
				go bcs.BlockchainPtr.RunConsensus()
				wg.Wait()
			} else {
				blockchain1, err := blockchain.SyncBlockchain(*remoteNode)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				blockchain2 := blockchain.NewBlockchainFromSync(blockchain1, "http://127.0.0.1:"+strconv.Itoa(int(*chainPort)))
				blockchain2.Peers[blockchain2.Address] = true
				bcs := blockchainserver.NewBlockchainServer(*chainPort, blockchain2)
				wg.Add(4)
				go bcs.Start()
				go bcs.BlockchainPtr.ProofOfWorkMining(*chainMiner)
				go bcs.BlockchainPtr.DialAndUpdatePeers()
				go bcs.BlockchainPtr.RunConsensus()
				wg.Wait()
			}

		}
	case "wallet":
		walletCmdSet.Parse(os.Args[2:])
		if walletCmdSet.Parsed() {
			if walletCmdSet.NFlag() == 0 {
				fmt.Println("Usage of wallet subcommand: ")
				walletCmdSet.PrintDefaults()
				os.Exit(1)
			}

			ws := walletserver.NewWalletServer(*walletPort, *blockchainNodeAddress)
			ws.Start()
		}
	default:
		fmt.Println("Error:Expected chain or wallet subcommand")
		os.Exit(1)
	}
}
