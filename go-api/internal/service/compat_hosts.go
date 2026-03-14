package service

// Compat receiver hosts that remain only because sibling files still attach
// methods to them during the compatibility-host phase.
type AuthService struct{}

type UserCenterService struct{}

var userCenterService = &UserCenterService{}
