package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
 
	"github.com/fgsoftware1/go-servers/FTP/sftp"
)

func main() {
	pk, err := ioutil.ReadFile("./ssh/id_rsa") // required only if private key authentication is to be used
	if err != nil {
		log.Fatalln(err)
	}
 
	config := sftp.Config{
		Username:     "inanzzz",
		Password:     "password", // required only if password authentication is to be used
		PrivateKey:   string(pk), // required only if private key authentication is to be used
		Server:       "0.0.0.0:2022",
		KeyExchanges: []string{"diffie-hellman-group-exchange-sha256", "diffie-hellman-group14-sha256"}, // optional
		Timeout:      time.Second * 30,                                                                  // 0 for not timeout
	}
 
	client, err := sftp.New(config)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
 
	// Open local file for reading.
	source, err := os.Open("file.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer source.Close()
 
	// Create remote file for writing.
	destination, err := client.Create("tmp/file.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer destination.Close()
 
	// Upload local file to a remote location as in 1MB (byte) chunks.
	if err := client.Upload(source, destination, 1000000); err != nil {
		log.Fatalln(err)
	}
 
	// Download remote file.
	file, err := client.Download("tmp/file.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
 
	// Read downloaded file.
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(data))
 
	// Get remote file stats.
	info, err := client.Info("tmp/file.txt")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", info)
}