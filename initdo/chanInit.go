package initdo

var (
	//非阻塞
	Admin = make(chan int, 2)
	Proxy = make(chan int, 2)
	All   = make(chan int, 2)
)
