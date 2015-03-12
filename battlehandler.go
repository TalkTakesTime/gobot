package main

import (
	"encoding/json"
)

var (
	MoveDex   map[string]Move
	TypeChart map[string]Type
)

type Move struct {
	Name      string
	Id        string
	Type      string
	Category  string
	BasePower int
	Status    string
}

type Pokemon struct {
	Name      string
	Id        string
	Type      []string
	BaseStats map[string]int
	Stats     map[string]int
}
