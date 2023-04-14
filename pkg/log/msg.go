package log

func Success(msg string) {
	Sugar.Infof("SUCCESS: %v", msg)
}

func Failure(msg string) {
	Sugar.Errorf("FAILED: %v", msg)
}
