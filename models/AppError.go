package models

import "log"

func AppError(method string) {
	if r := recover(); r != nil {
		log.Println(method + " - ", r)
	}


}
