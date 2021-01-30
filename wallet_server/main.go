package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("WalletServer: ")
}

func main() {
	port := flag.Uint("port", 8080, "TCP Number for Wallet Server")
	gateway := flag.String("gateway", "http://127.0.0.1:5001", "Blockchain Gateway")
	flag.Parse()

	app := NewWalletServer(uint16(*port), string(*gateway))
	app.Run()
}
