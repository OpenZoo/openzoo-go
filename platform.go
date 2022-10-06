package main

type IdleMode int

const (
	IdleUntilPit IdleMode = iota
	IdleUntilFrame
	IdleMinimal
)
