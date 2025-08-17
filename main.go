package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

func main() {
	// 配置服务
	port := "30012"

	serviceName := "go-service"
	serviceID := fmt.Sprintf("%s-%s", serviceName, uuid.New().String())
	consulHost := os.Getenv("CONSUL_HOST")
	if consulHost == "" {
		consulHost = "consul"
	}
	consulPort := os.Getenv("CONSUL_PORT")
	if consulPort == "" {
		consulPort = "8500"
	}

	// 初始化 Consul 客户端
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", consulHost, consulPort)
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	// 注册服务到 Consul
	err = client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: getHostname(),
		Port:    parsePort(port),
		Tags:    []string{fmt.Sprintf("urlprefix-/%s", serviceName)},
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%s/health", getHostname(), port),
			Interval: "10s",
			Timeout:  "5s",
		},
	})
	if err != nil {
		log.Fatalf("Failed to register service with Consul: %v", err)
	}
	log.Printf("Service %s registered with Consul", serviceID)

	// 设置 HTTP 路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message": "Hello from %s on %s"}`, serviceName, getHostname())
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// 启动 HTTP 服务
	addr := fmt.Sprintf(":%s", port)
	go func() {
		log.Printf("Server running on port %s", port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 注销服务
	err = client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		log.Printf("Failed to deregister service: %v", err)
	}
	log.Printf("Service %s deregistered from Consul", serviceID)
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func parsePort(port string) int {
	var p int
	fmt.Sscanf(port, "%d", &p)
	return p
}
