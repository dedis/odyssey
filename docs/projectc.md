# Projectc 

This is the project contract

## Purpose

Holds the project contract, which is instantiated each time a data
scientist makes a request. The project instance holds all the
informations about a project.

## Set up

We need a local version of the cothority repository and do two things:

1. In the `/conode/conode.go` add this import directive:  
`_ "github.com/dedis/odyssey/projectc"`
2. In the `go.mod` of the cothority, add this directive at the very end
   (adapt to your path):  
`replace github.com/dedis/odyssey v0.0.0 => /Users/nkocher/GitHub/odyssey`
