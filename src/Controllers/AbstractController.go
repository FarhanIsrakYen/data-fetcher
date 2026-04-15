package Controller

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
	"data-fetcher-api/src/Helper"
	"data-fetcher-api/src/Lib"
)

type UserDetails struct {
	ID int `json:"id"`
}

func GetUser(c *gin.Context) (UserDetails, error) {

	authorization := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authorization) == 2 && authorization[0] == "Bearer" {
		token := authorization[1]
		parameter, _ := Helper.GetParameter()
		key := parameter.Parameters.DfSessionPrefix + token
		sessionId, _ := Lib.GetValue(key)
		if sessionId != "" {
			user, _ := Lib.GetValue(parameter.Parameters.DfSessionPrefix + sessionId)

			start := strings.Index(user, "{")
			end := strings.LastIndex(user, "}")
			secondStart := strings.Index(user[start+1:], "{") + start
			secondEnd := strings.LastIndex(user[:end], "}")

			jsonStr := user[secondStart+1 : secondEnd-2]

			var userData UserDetails

			err := json.Unmarshal([]byte(jsonStr), &userData)
			if err != nil {
				return userData, err
			} else {
				return userData, err

			}

			return userData, nil
		}
	}
	return UserDetails{}, errors.New("user not found")

}
