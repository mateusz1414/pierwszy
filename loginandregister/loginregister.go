package loginandregister

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"students/user"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Outs struct {
	Message       string `json:"message"`
	ErrorCode     string `json:"errorCode"`
	ActivationURL string `json:"activationURL"`
}

type Credentials struct {
	Google Platform `json:"google"`
}

type Platform struct {
	Cid           string   `json:"cid"`
	Csecret       string   `json:"csecret"`
	Redirect      string   `json:"redirect"`
	Scopes        []string `json:"scopes"`
	GetJwtAddress string   `json:"getjwtaddress"`
}

func outFunc(status int, mess string, errc string, c *gin.Context) {
	outs := Outs{
		Message:   mess,
		ErrorCode: errc,
	}
	c.JSON(status, outs)
}

func Login(c *gin.Context) {
	userData := user.User{}
	err := c.ShouldBindJSON(&userData)
	if err != nil {
		outFunc(400, "Invalid data format", err.Error(), c)
		return
	}
	db, DBbool := c.Get("db")
	if DBbool == false {
		outFunc(500, "Database not found", "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	err = userData.Authentication(database)
	if err != nil {
		outFunc(400, "Invalid login or password", err.Error(), c)
	} else {
		if userData.Active == 0 {
			c.JSON(400, gin.H{
				"message":       "Accont is not active",
				"errorCode":     "Not active",
				"activationURL": user.APIAdress + "user/activation/" + userData.Code,
			})
			return
		}
		token, err := user.CreateJWTToken(jwt.MapClaims{
			"userid":      userData.UserID,
			"permissions": userData.Permissions,
			"time":        time.Now().Unix(),
		})
		if err != nil {
			outFunc(500, "Server error", err.Error(), c)
		} else {
			c.JSON(200, gin.H{
				"message":     "Logged",
				"errorCode":   "",
				"permissions": userData.Permissions,
				"email":       userData.Email,
				"userID":      userData.UserID,
				"authToken":   token,
			})
		}
	}
}

func Register(c *gin.Context) {
	userData := user.User{}
	err := c.ShouldBindJSON(&userData)
	if err != nil {
		outFunc(400, "Niepoprawny format danych", err.Error(), c)
		return
	}
	db, DBbool := c.Get("db")
	if DBbool == false {
		outFunc(500, "Nie znaleziono bazy danych", "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	err = userData.RegisterValidate(database)
	if err != nil {
		outFunc(400, "Register failed", err.Error(), c)
	} else {
		result := Outs{}
		result.Message = "Registered"
		result.ActivationURL = user.APIAdress + "user/activation/" + userData.Code
		c.JSON(200, result)
	}

}

func getOauthConfig(credentials interface{}, provider string) oauth2.Config {
	platform, endpoint := getData(provider, credentials)
	return oauth2.Config{
		ClientID:     platform.Cid,
		ClientSecret: platform.Csecret,
		RedirectURL:  platform.Redirect,
		Scopes:       platform.Scopes,
		Endpoint:     endpoint,
	}
}

func OauthAuthorize(c *gin.Context) {
	credentials, _ := c.Get("credentials")
	config := getOauthConfig(credentials, c.Param("provider"))
	url := config.AuthCodeURL(generateSaveToken(c))
	c.Redirect(302, url)
}

func OauthLogin(c *gin.Context) {
	credentialsFile, _ := c.Get("credentials")
	db, _ := c.Get("db")
	database := db.(*gorm.DB)
	config := getOauthConfig(credentialsFile, c.Param("provider"))
	platform, _ := getData(c.Param("provider"), credentialsFile)
	code := c.Query("code")
	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		c.Redirect(302, user.ServerAdress+"/oauth/null")
		return
	}
	client := config.Client(oauth2.NoContext, token)
	response, err := client.Get(platform.GetJwtAddress)
	if err != nil {
		c.Redirect(302, user.ServerAdress+"/oauth/null")
		return
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	person := user.OauthData{}
	json.Unmarshal(data, &person)
	apiUser, err := person.OauthLogin(database)
	if err != nil {
		c.Redirect(302, user.ServerAdress+"/oauth/null")
		return
	}
	jwtToken, err := user.CreateJWTToken(jwt.MapClaims{
		"email":       apiUser.Email,
		"userid":      apiUser.UserID,
		"permissions": apiUser.Permissions,
		"time":        time.Now().Unix(),
	})
	if err != nil {
		c.Redirect(302, user.ServerAdress+"/oauth/null")
		return
	}
	c.Redirect(302, user.ServerAdress+"oauth/"+jwtToken)

}

func getData(provider string, credentialsFile interface{}) (platform Platform, endpoint oauth2.Endpoint) {
	file := credentialsFile.([]byte)
	var credentials Credentials
	json.Unmarshal(file, &credentials)
	switch provider {
	case "google":
		platform = credentials.Google
		endpoint = google.Endpoint
	}
	return platform, endpoint
}

func generateSaveToken(c *gin.Context) string {
	token := make([]byte, 32)
	rand.Read(token)
	state := base64.StdEncoding.EncodeToString(token)
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	return state
}
