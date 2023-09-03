package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/exec"
	"time"
)

type Config struct {
	ParentAddr  string
	ParentPort  string
	ParentToken string
	Name        string
	BindPort    string
	HostPort    string
	ChildToken  string
}

var (
	g      errgroup.Group
	osType = "darwin" // linux darwin
)

func main() {
	if os.Args[1] == "" {
		log.Fatal("config path not found")
	}
	var configPath = os.Args[1]
	config := readConfig(configPath)
	if config.ParentAddr != "" {
		createFrpcIni(&config)
		execFrpc(&config)
		app := fiber.New()
		app.Use(logger.New())
		app.Listen(fmt.Sprintf(":%s", config.HostPort))
	}

	createFrpsIni(&config)
	execFrps(&config)

	g.Wait()
}

func readConfig(path string) Config {
	var config Config
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(file, &config)
	return config
}

func createFrpcIni(config *Config) {
	file, err := os.Create("./frpc.ini")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString("[common]\n")
	file.WriteString(fmt.Sprintf("server_addr = %s\n", config.ParentAddr))
	file.WriteString(fmt.Sprintf("server_port = %s\n", config.ParentPort))
	file.WriteString(fmt.Sprintf("token = %s\n", config.ParentToken))
	file.WriteString("\n")
	file.WriteString(fmt.Sprintf("[%s]\n", config.Name))
	file.WriteString(fmt.Sprintf("type = http\n"))
	file.WriteString(fmt.Sprintf("local_port = %s\n", config.HostPort))
	file.WriteString(fmt.Sprintf("custom_domains = %s\n", "0.0.0.0"))
	file.WriteString(fmt.Sprintf("locations = /%s\n", config.Name))
}

func createFrpsIni(config *Config) {
	file, err := os.Create("./frps.ini")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString("[common]\n")
	file.WriteString(fmt.Sprintf("bind_port = %s\n", config.BindPort))
	file.WriteString(fmt.Sprintf("token = %s\n", config.ChildToken))
	file.WriteString(fmt.Sprintf("vhost_http_port = %s\n", config.HostPort))
}

func execFrpc(config *Config) {
	cmd := exec.Command(fmt.Sprintf("./%s/frpc", osType), "-c", "./frpc.ini")
	go func() {
		<-make(chan os.Signal)
		cmd.Process.Kill()
	}()
	go func() {
		time.Sleep(2 * time.Second)
		log.Printf("link success: %s parent addr %s bind port %s\n", config.Name, config.ParentAddr, config.ParentPort)
	}()
	go func() {
		output, err := cmd.Output()
		if err != nil {
			log.Fatalf("exec frpc error: %s\n", err)
		}
		log.Printf("%s\n", output)
		os.Exit(0)
	}()
}

func execFrps(config *Config) {
	cmd := exec.Command(fmt.Sprintf("./%s/frps", osType), "-c", "./frps.ini")
	g.Go(func() error {
		<-make(chan os.Signal)
		cmd.Process.Kill()
		return nil
	})
	go func() {
		time.Sleep(2 * time.Second)
		log.Printf("start frps success: %s bind %s host %s\n", config.Name, config.BindPort, config.HostPort)
	}()
	go func() {
		output, err := cmd.Output()
		if err != nil {
			log.Fatalf("exec frps error: %s\n", err)
		}
		log.Printf("%s\n", output)
		os.Exit(0)
	}()
}
