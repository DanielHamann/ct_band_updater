package main

// BuildID is injected at build time via: -ldflags "-X 'main.BuildID=...'"
var BuildID = "dev"
