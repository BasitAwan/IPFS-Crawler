package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var IPToKey map[string]string
var keyToIP map[string]string

type keyIP struct {
	key string
	ip  string
}

func crawl(jobs chan string, keyIPMapChan, ipKeyMapChan chan *keyIP) {
	for {

		key := <-jobs
		// fmt.Println(key, "lol")
		cmd1 := exec.Command("ipfs", "dht", "findpeer", key)
		output1, err := cmd1.Output()
		if err != nil {
			// fmt.Println(err)
			continue
		}
		err = cmd1.Process.Kill()
		keyIpPair := new(keyIP)
		keyIpPair.key = key
		keyIpPair.ip = string(output1)
		keyIPMapChan <- keyIpPair
		ipList := strings.Split(string(output1), "\n")
		for _, x := range ipList {
			keyIpPair.ip = x
			ipKeyMapChan <- keyIpPair
		}
		cmd := exec.Command("timeout", "2", "ipfs", "dht", "query", key)
		output, err := cmd.Output()

		stKeys := strings.Split(string(output), "\n")
		for _, x := range stKeys {
			jobs <- x

		}
	}
}

//listens to the channel for results and checks if key already present in the map, if it is then doesn't write into csv file
func mapHandler(keyIPMapChan chan *keyIP) {
	f, _ := os.Create("KeyToMetaData.csv")
	writer := csv.NewWriter(f)
	titles := make([]string, 2)
	titles[0] = "Key"
	titles[1] = "MetaData"
	writer.Flush()

	writer.Write(titles)
	for {
		newEntry := <-keyIPMapChan
		if len(keyToIP[newEntry.key]) > 0 {
			continue
		} else {
			if len(newEntry.ip) < 2 {
				continue
			}
			keyToIP[newEntry.key] = newEntry.ip
			titles[0] = newEntry.key
			titles[1] = newEntry.ip
			titles[1] = strings.ReplaceAll(titles[1], "\n", " ")
			writer.Write(titles)
			writer.Flush()
			fmt.Println(len(keyToIP), "Key Count")
		}
	}

}
func mapHandler1(keyIPMapChan chan *keyIP) {
	f, _ := os.Create("IPtoKeyData.csv")
	writer := csv.NewWriter(f)
	titles := make([]string, 2)
	titles[0] = "Ip"
	titles[1] = "Key"
	writer.Flush()

	writer.Write(titles)
	for {
		newEntry := <-keyIPMapChan
		if len(IPToKey[newEntry.ip]) > 0 {
			continue
		} else {

			IPToKey[newEntry.ip] = newEntry.key
			titles[0] = newEntry.ip
			titles[1] = newEntry.key
			writer.Write(titles)
			writer.Flush()
			fmt.Println(len(IPToKey), "Ip Count")
		}
	}

}

func main() {
	// runtime.GOMAXPROCS(2)
	keyToIP = make(map[string]string)
	IPToKey = make(map[string]string)
	cmd := exec.Command("ipfs", "swarm", "peers")
	output, err := cmd.CombinedOutput()
	cmd.Process.Kill()
	stkeys := strings.Split(string(output), "\n")
	fmt.Println(stkeys[0])
	firstKeys := make([]string, 0)
	jobs := make(chan string, 1000000)
	keyIPMapChan := make(chan *keyIP)
	ipKeyMapChan := make(chan *keyIP)
	go mapHandler(keyIPMapChan)
	go mapHandler1(ipKeyMapChan)
	for _, x := range stkeys {
		fmt.Println(x)
		y := strings.Split(x, "/")
		if len(y) < 7 {
			continue
		}
		extractedKey := y[6]
		if extractedKey == "ipfs" {
			continue
		}
		firstKeys = append(firstKeys, extractedKey)

	}
	for i := 0; i < 5000; i++ {
		go crawl(jobs, keyIPMapChan, ipKeyMapChan)
	}
	for _, x := range firstKeys {
		jobs <- x
	}
	crawl(jobs, keyIPMapChan, ipKeyMapChan)
	fmt.Println(firstKeys)

	if err != nil {
		fmt.Println(err)
	}

}
