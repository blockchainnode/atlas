package main

import (
	"fmt"
	"github.com/mapprotocol/atlas/cmd/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
)

func withdrawMock(ctx *cli.Context) error {
	debugInfo := debugInfo{}
	debugInfo.preWork(ctx, []int{1, 2, 3}, true)
	debugInfo.withdrawMock(ctx) //change this
	return nil
}

func (d *debugInfo) withdrawMock(ctx *cli.Context) {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			switch currentEpoch {
			case 1:
				d.queck(QUERY_RELAYERINFO)
				d.doWithdraw()
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
				return
			}
		}
	}
}

func (d *debugInfo) doWithdraw() {
	fmt.Println("=================DO Withdraw========================")
	conn := d.client
	for k, _ := range d.relayerData {
		err := d.relayerData[k].withdraw11(conn)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (d *debugInfo) withdrawAtDifferentEpoch12() {
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			switch currentEpoch {
			case 1:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.doWithdraw()
				d.atlasBackendCh <- NEXT_STEP
			case 2:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				fmt.Println("=====================================================")
				d.doWithdraw()
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
			case 3:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.atlasBackendCh <- NEXT_STEP
				return
			}
		}
	}
}

func (d *debugInfo) withdrawAccordingToDifferentBalance12() {
	go d.atlasBackend()
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("CURRENT EPOCH ========>", currentEpoch)
			switch currentEpoch {
			case 1:
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.changeAllImpawnValue(500)
				d.doWithdraw()
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.changeAllImpawnValue(300)
				d.doWithdraw()
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.changeAllImpawnValue(100)
				d.doWithdraw()
				d.queck(QUERY_RELAYERINFO)
				d.queck(BALANCE)
				d.queck(IMPAWN_BALANCE)
				d.changeAllImpawnValue(1000000)
				d.doWithdraw()
				d.atlasBackendCh <- NEXT_STEP
			}
		}
	}
}
func (r *relayerInfo) withdraw11(conn *ethclient.Client) error {

	if int(r.impawnValue) <= 0 {
		log.Fatal("Value must bigger than 0")
	}
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	value := new(big.Int).Mul(big.NewInt(r.impawnValue), baseUnit)

	input := packInput("withdraw", r.from, value)

	sendContractTransaction(conn, r.from, RelayerAddress, new(big.Int).SetInt64(0), r.priKey, input)

	return nil
}