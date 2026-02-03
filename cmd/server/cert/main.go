package main

import (
	"flag"
	"fmt"
	"gopthkeeper/internal/tls"
	"log"
	"os"
)

func main() {
	certFile := flag.String("cert", "server.crt", "Output certificate file")
	keyFile := flag.String("key", "server.key", "Output private key file")
	flag.Parse()

	log.Printf("Generating self-signed certificate...")
	log.Printf("Certificate file: %s", *certFile)
	log.Printf("Private key file: %s", *keyFile)

	if err := tls.GenerateSelfSignedCert(*certFile, *keyFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Successfully generated self-signed certificate!\nThis certificate is for development only")
}
