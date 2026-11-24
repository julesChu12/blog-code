package main

func main() {
	ch := make(chan string)
	go func() {
		ch <- "user"
		ch <- "order"
		close(ch)
	}()

	for i := range ch {
		println(i)
	}
}
