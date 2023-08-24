package main

import (
	"fmt"
	"net/http"
	"crypto/tls"
	"io/ioutil"  // Import ioutil package for reading files
	"encoding/json"
	"strings"
)

const (
	apiURL = "https://192.168.1.201:6443/api/v1/services"
	tokenFile = "token.txt"
)

type ServiceList struct {
	Items []Service `json:"items"`
}

type Service struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
	Status   Status   `json:"status"`
}

type Metadata struct {
	Name string `json:"name"`
}

type Spec struct {
	Type  string `json:"type"`
	Ports []Port `json:"ports"`
}

type Port struct {
	Port int `json:"port"`
}

type Status struct {
	LoadBalancer LoadBalancer `json:"loadBalancer"`
}

type LoadBalancer struct {
	Ingress []Ingress `json:"ingress"`
}

type Ingress struct {
	IP string `json:"ip"`
}

func readAPITokenFromFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func main() {
	token, err := readAPITokenFromFile(tokenFile)
	if err != nil {
		fmt.Println("Error reading token:", err)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	var serviceList ServiceList
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&serviceList); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	for _, service := range serviceList.Items {
		if service.Spec.Type == "LoadBalancer" && len(service.Status.LoadBalancer.Ingress) > 0 {
			fmt.Printf("Service: %s\n", service.Metadata.Name)
			fmt.Printf("External IP(s): %s\n", getExternalIPs(service.Status.LoadBalancer.Ingress))
			for _, port := range service.Spec.Ports {
				fmt.Printf("Port: %d\n", port.Port)
			}
			fmt.Println("---")
		}
	}
}

func getExternalIPs(ingress []Ingress) string {
	var ips []string
	for _, i := range ingress {
		ips = append(ips, i.IP)
	}
	return strings.Join(ips, ", ")
}
