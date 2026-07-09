package main

func main() {
	wal, err := NewWal("/home/pianodessources/Projects/distributed-kv-store/leader.wal")
	if err != nil {
		println("Could not initialize write ahead log")
		println(err.Error())
	}
	store, err := NewStore(1024, wal)

	if err != nil {
		println("Could not initialize store")
		println(err.Error())
	}

	server := NewServer(store)

	server.ListenAndServe()

}
