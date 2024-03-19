package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

type Result struct {
	privateKey string
	address    string
}

func generateVanityAddress(prefix string, suffix string, wg *sync.WaitGroup, results chan<- Result, quit chan struct{}) {
	defer wg.Done()

	prefix = strings.ToLower(prefix)
	suffix = strings.ToLower(suffix)
	counter := 0

	for {
		select {
		case <-quit:
			return
		default:
			privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
			if err != nil {
				fmt.Println("Error generating key:", err)
				continue
			}

			publicKey := privateKey.PublicKey

			publicKeyBytes := elliptic.Marshal(crypto.S256(), publicKey.X, publicKey.Y)[1:]
			d := sha3.NewLegacyKeccak256()
			d.Write(publicKeyBytes)
			address := d.Sum(nil)[12:]

			value := fmt.Sprintf("%x", address)
			if strings.HasPrefix(value, prefix) && strings.HasSuffix(value, suffix) {
				results <- Result{fmt.Sprintf("%x", privateKey.D), "0x" + value}
				return
			} else {
				counter++
				if counter%100000 == 0 {
					fmt.Printf("\r[-] Generated Addresses: %d", counter)
				}
			}
		}
	}
}

func main() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\t💎 ETH Vanity Address Generator\n\n")

	fmt.Print("[-] Prefix: ")
	scanner.Scan()
	vanityPrefix := scanner.Text()

	fmt.Print("[-] Suffix: ")
	scanner.Scan()
	vanitySuffix := scanner.Text()

	fmt.Print("[-] Threads: (4)")
	scanner.Scan()
	threadsStr := scanner.Text()

	threadsAmount, err := strconv.Atoi(threadsStr)
	if err != nil {
		threadsAmount = 4
	}

	results := make(chan Result)
	quit := make(chan struct{})
	var wg sync.WaitGroup

	start := time.Now()

	numGoroutines := threadsAmount
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go generateVanityAddress(vanityPrefix, vanitySuffix, &wg, results, quit)
	}

	result := <-results
	close(quit)
	wg.Wait()

	elapsed := time.Since(start)
	output := fmt.Sprintf("\n\nPrivate Key: %s\nAddress: %s\n\nElapsed time: %s", result.privateKey, result.address, elapsed)
	fmt.Println(output)
	scanner.Scan()
}
