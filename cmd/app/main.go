package main

import "dosage_drugs/internal/api"

// @title           Dosage Drugs API
// @version         1.0
// @description     Веб-сервис дозирования лекарств для детей. Чтение доступно гостю, методы создателя — по JWT.

// @host      localhost:8090
// @BasePath  /api
// @schemes   http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите: Bearer <access_token из /users/login>

func main() {
	api.StartServer()
}
