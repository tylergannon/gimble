package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: quote SUBTOTAL MODE")
		os.Exit(1)
	}
	amount, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	subtotal := int(math.Round(amount * 100))
	mode := os.Args[2]
	if mode != "standard" {
		fmt.Fprintln(os.Stderr, "unsupported shipping mode")
		os.Exit(1)
	}
	shipping := 800
	if subtotal > 5000 {
		shipping = 0
	}
	if err := json.NewEncoder(os.Stdout).Encode(map[string]any{
		"subtotal_cents": subtotal, "shipping_cents": shipping,
		"total_cents": subtotal + shipping, "mode": mode,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
