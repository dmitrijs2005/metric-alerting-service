// Command keygen is the entry point for the key generator binary.
// It provides helpers to generate an RSA key pair and export
// the private/public keys in PEM-encoded strings or files.
package main

import (
	"flag"
	"fmt"
)

// main initializes the application:
// - loads configuration (key storage path)
// - creates and stores private and public keys
func main() {

	var outputDir string
	flag.StringVar(&outputDir, "d", ".", "keys output dir")
	flag.Parse()

	// Generate a 2048-bits key
	privateKey, publicKey := generateKeyPair(2048)

	privKeyString := exportPrivKeyAsPEMStr(privateKey)
	writePEMFile(fmt.Sprintf("%s/%s", outputDir, "private.pem"), []byte(privKeyString), 0400)

	pubKeyString := exportPubKeyAsPEMStr(publicKey)
	writePEMFile(fmt.Sprintf("%s/%s", outputDir, "public.pem"), []byte(pubKeyString), 0400)

}
