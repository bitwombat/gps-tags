package main

func logIfErr(err error) {
	if err != nil {
		errorLogger.Printf("error sending notification: %v", err)
	}
}
