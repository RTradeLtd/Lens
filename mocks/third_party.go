package mocks

// Mock generation of third-party interfaces should be tracked here.

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o ./manager.mock.go -fake-name FakeRTFSManager github.com/RTradeLtd/rtfs/v2.Manager
