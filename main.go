package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"golang.org/x/sync/errgroup"
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
	suffix = ""
)

func main() {
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
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
	cmd := exec.Command(fmt.Sprintf("./%s/frpc%s", runtime.GOOS, suffix), "-c", "./frpc.ini")
	go func() {
		<-make(chan os.Signal)
		cmd.Process.Kill()
	}()
	go runExec(cmd)
}

func execFrps(config *Config) {
	cmd := exec.Command(fmt.Sprintf("./%s/frps%s", runtime.GOOS, suffix), "-c", "./frps.ini")
	g.Go(func() error {
		<-make(chan os.Signal)
		cmd.Process.Kill()
		return nil
	})
	go runExec(cmd)
}

func GetOutput(reader *bufio.Reader) {
	var sumOutput string
	outputBytes := make([]byte, 200)
	for {
		n, err := reader.Read(outputBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			sumOutput += err.Error()
		}
		output := string(outputBytes[:n])
		fmt.Print(output) //输出屏幕内容
		sumOutput += output
	}
}

func runExec(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("exec frp stdout error: %s\n", err)
	}
	readout := bufio.NewReader(stdout)
	go GetOutput(readout)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("exec frp stderr error: %s\n", err)
	}
	readerr := bufio.NewReader(stderr)
	go GetOutput(readerr)
	cmd.Run()
}
