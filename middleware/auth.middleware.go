package middleware

import (
	"fmt"
	"gin-goinc-api/utils"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	Bearer = "Bearer"
)

func AuthMiddlelware(ctx *gin.Context) {
	bearerToken := ctx.GetHeader("Authorization")
	if !strings.Contains(bearerToken, Bearer) {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "token is required",
		})
		return
	}

	token := strings.Replace(bearerToken, "Bearer ", "", -1)
	//fmt.Println(token)

	if token == "" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "token is required",
		})
		return
	}

	claimsData, err := utils.DecodeToken(token)
	fmt.Println(claimsData)
	fmt.Println(claimsData["id"])

	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "token is invalid",
		})
		return
	}

	var userID int
	switch v := claimsData["id"].(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	default:
		ctx.JSON(500, gin.H{
			"message": "Tipo de ID não suportado",
		})
		return
	}
	fmt.Println("conveeeeeeerteu", reflect.TypeOf(userID))

	ctx.Set("claimsData", claimsData)
	ctx.Set("user_id", claimsData["id"])

	ctx.Set("user_name", claimsData["name"])
	ctx.Set("user_email", claimsData["email"])

	ctx.Next()
}

// ExtractUserIDFromContext extrai o ID do usuário do contexto Gin.
func ExtractUserIDFromContext(ctx *gin.Context) int8 {
	const Bearer = "Bearer"

	// Obter o cabeçalho Authorization do contexto
	bearerToken := ctx.GetHeader("Authorization")

	// Verificar se o token está presente
	if !strings.Contains(bearerToken, Bearer) {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "Token is required",
		})
		return 0
	}

	// Remover o prefixo "Bearer " do token
	token := strings.Replace(bearerToken, Bearer+" ", "", -1)

	// Verificar se o token está vazio
	if token == "" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "Token is required",
		})
		return 0
	}

	// Decodificar o token para obter os dados das reivindicações
	claimsData, err := utils.DecodeToken(token)
	fmt.Println(claimsData)
	fmt.Println(err)

	// Verificar se ocorreu um erro ao decodificar o token
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "Token is invalid",
		})
		return 0
	}

	// Tentar converter o valor do ID para int8
	userIDFloat, ok := claimsData["id"].(float64)
	fmt.Println("Values ==", userIDFloat, ok)

	// Verificar se a conversão para int8 foi bem-sucedida
	userID := int8(userIDFloat)
	fmt.Println("Values ==", userID)

	return userID
}

func TokenMiddleware(ctx *gin.Context) {
	token := ctx.GetHeader("X-Token")

	if token == "" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "token is required",
		})
		return
	}

	if token != "123" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "token is invalid",
		})
	}

	ctx.Next()
}
