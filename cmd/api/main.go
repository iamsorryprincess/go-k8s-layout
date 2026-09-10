package main

import (
	"os"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/background"
)

func main() {
	os.Exit(background.Run(NewApp))
}
