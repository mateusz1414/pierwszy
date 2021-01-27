package loginandregister

import (
	"encoding/json"
	"io/ioutil"
	"students/user"
	"time"

	"github.com/dgrijalva/jwt-go"

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
	Cid      string   `json:"cid"`
	Csecret  string   `json:"csecret"`
	Redirect string   `json:"redirect"`
	Scopes   []string `json:"scopes"`
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

func OauthLogin(c *gin.Context) {
	config, _ := c.Get("credentials")
	file := config.([]byte)
	db, ok := c.Get("db")
	if !ok {
		c.Redirect(302, user.ServerAdress)
		return
	}
	database := db.(*gorm.DB)
	var cred Credentials
	var oaperson user.OauthData
	json.Unmarshal(file, &cred)
	platform, endpoint := getData("google", cred)
	conf := oauth2.Config{
		ClientID:     platform.Cid,
		ClientSecret: platform.Csecret,
		RedirectURL:  platform.Redirect,
		Scopes:       platform.Scopes,
		Endpoint:     endpoint,
	}
	tok, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.Redirect(302, user.ServerAdress)
		return
	}
	client := conf.Client(oauth2.NoContext, tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.Redirect(302, user.ServerAdress)
		return
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(data, &oaperson)
	person, err := oaperson.OauthLogin(database)
	if err != nil {
		c.Redirect(302, user.ServerAdress)
		return
	}
	c.JSON(200, person)
}

func getData(provider string, cred Credentials) (platform Platform, endpoint oauth2.Endpoint) {
	switch provider {
	case "google":
		platform = cred.Google
		endpoint = google.Endpoint
	}
	return platform, endpoint
}
