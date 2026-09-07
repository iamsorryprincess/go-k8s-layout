package main

import (
	"fmt"

	"github.com/iamsorryprincess/go-k8s-layout/internal/config"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/env"
)

func main() {
	_, err := env.Parse[config.ApiConfig]()
	if err != nil {
		fmt.Println(err)
		return
	}
}
