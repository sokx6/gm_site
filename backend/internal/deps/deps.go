// Package deps ensures all required dependencies are pinned in go.mod.
// This file is removed once the actual code imports these packages.
package deps

import (
	_ "github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/gorilla/websocket"
	_ "github.com/spf13/viper"
	_ "github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	_ "golang.org/x/crypto/bcrypt"
	_ "gopkg.in/gomail.v2"
)
