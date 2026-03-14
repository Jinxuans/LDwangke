package order

var orderStatusNotifier = func(oid int, newStatus string, newProcess string, remarks string) {}

func SetOrderStatusNotifier(fn func(oid int, newStatus string, newProcess string, remarks string)) {
	if fn == nil {
		orderStatusNotifier = func(int, string, string, string) {}
		return
	}
	orderStatusNotifier = fn
}
