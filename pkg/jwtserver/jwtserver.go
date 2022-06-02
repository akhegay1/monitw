package jwtserver

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

var Logger *log.Logger

var Appkey = "jgF5Op__dwdqw"

type JwtServer struct {
}

func init() {
	Logger = log.New()
}

func (s *JwtServer) GetToken(ctx context.Context, in *Reqtoken) (*Tokenstring, error) {
	Logger.Printf("Receive message body from client: %s %s", in.User)
	vTokenString := generateToken(in.User)
	return &Tokenstring{TokenString: vTokenString}, nil
}

func (s *JwtServer) CheckToken(ctx context.Context, in *CheckAuth) (*AuthRslt, error) {
	Logger.Printf("Receive message body from client 111: %s %s", in.TokenString)
	vtokenvalid, vuser := checkTokenFnc(in.TokenString)
	return &AuthRslt{Tokenvalid: vtokenvalid, User: vuser}, nil
}

func (s *JwtServer) mustEmbedUnimplementedJwtServerServiceServer() {
}

func generateToken(username string) string {
	// We are happy with the credentials, so build a token. We've given it
	// an expiry of 1 hour.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": username,
		"exp":  time.Now().Add(time.Hour * time.Duration(10000)).Unix(),
		"iat":  time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(Appkey))
	if err != nil {
		Logger.Println("err", err)
		return err.Error()
	}
	Logger.Println("tokenString", tokenString)
	return tokenString
}

func checkTokenFnc(reqToken string) (bool, string) {
	Logger.Println("in CheckToken")
	var mySigningKey = []byte(Appkey)

	//Logger.Println(reqToken, Appkey)

	token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
		Logger.Println("in anonym fnc")
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(("Invalid Signing Method"))
		}
		if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
			return nil, fmt.Errorf(("Expired token"))
		}

		return mySigningKey, nil
	})

	Logger.Println("token.Valid", token.Valid)

	if err != nil {
		Logger.Println("AuthMiddleware bef next", err.Error())
		return false, ""
	}

	/////////////////////////////////////////////////
	///////////////////Parse token to claims////////////////////////////
	/////////////////////////////////////////////////
	claims := jwt.MapClaims{}
	tokenA, err := jwt.ParseWithClaims(reqToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(Appkey), nil
	})
	Logger.Println("tokenA", tokenA)
	return true, claims["user"].(string)

}
