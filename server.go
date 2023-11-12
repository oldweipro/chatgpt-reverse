package main

type server interface {
	ListenAndServe() error
}
