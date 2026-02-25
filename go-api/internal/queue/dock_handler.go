package queue

// DockOrderHandler 的实际实现放在 service 层，通过 main.go 注入到 DockQueue
// 这样避免 queue → service → queue 的循环依赖
