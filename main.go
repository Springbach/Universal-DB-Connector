package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("docker-compose", "up", "--build")
		r, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err = cmd.Start(); err != nil {
			log.Fatal(err)
		}

		if _, err := io.Copy(os.Stdout, r); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	//wait some time for docker-compose up before DB connection
	time.Sleep(5 * time.Second)
	db := NewDB("psql")
	err := db.Connect(&PSQLconnector{})
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()

}
