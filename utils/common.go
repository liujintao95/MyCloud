package utils

func ErrCheck(err error, msg string, risk bool)  {
	if err != nil{
		Logging.Error(msg)
		if risk {
			panic("")
		}
	}
}
