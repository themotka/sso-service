package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ssov1 "github.com/themotka/proto/gen/go/sso"
	"github.com/themotka/sso-service/tests/suite"
	"strconv"
	"testing"
	"time"
)

const (
	emptyAppID = -1
	appID      = 1
	appSecret  = "secret"

	passLen = 8
)

func TestAuthRegisterLogin_Happy(t *testing.T) {
	ctx, s := suite.NewSuite(t)

	mail := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := s.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    mail,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := s.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    mail,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetJwtToken()
	require.NotEmpty(t, token)
	parse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	claims := parse.Claims.(jwt.MapClaims)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, mail, claims["email"].(string))
	appId, err := strconv.Atoi(claims["app_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, appID, appId)

	// Проверка iat с точностью до секунды
	const delta = 1
	assert.InDelta(t, loginTime.Add(s.Config.TokenExpire).Unix(), claims["exp"].(float64), delta)
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passLen)
}

func TestAuthRegisterLogin_DoubleRegistration(t *testing.T) {
	ctx, s := suite.NewSuite(t)

	mail := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := s.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    mail,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = s.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    mail,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, s := suite.NewSuite(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Short Password",
			email:       gofakeit.Email(),
			password:    "123",
			expectedErr: "password must be at least 8 characters",
		},
		{
			name:        "Register with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "password should not be empty",
		},
		{
			name:        "Register with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			expectedErr: "email is not valid",
		},
		{
			name:        "Register with Both Empty",
			email:       "",
			password:    "",
			expectedErr: "email is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, s := suite.NewSuite(t)

	tests := []struct {
		name        string
		appId       int32
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Login with Empty AppId",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appId:       emptyAppID,
			expectedErr: "app id should not be empty",
		},
		{
			name:        "Login with Short Password",
			email:       gofakeit.Email(),
			password:    "123",
			expectedErr: "password must be at least 8 characters",
		},
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       appID,
			expectedErr: "password should not be empty",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			appId:       appID,
			expectedErr: "email is not valid",
		},
		{
			name:        "Login with Both Empty",
			email:       "",
			password:    "",
			appId:       appID,
			expectedErr: "email is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appId,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
