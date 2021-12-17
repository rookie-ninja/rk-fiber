// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberjwt

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"reflect"
	"strings"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	ctx := fiber.New().AcquireCtx(&fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	})

	// without options
	set := newOptionSet()
	assert.NotEmpty(t, set.EntryName)
	assert.NotEmpty(t, set.EntryType)
	assert.False(t, set.Skipper(ctx))
	assert.Empty(t, set.SigningKeys)
	assert.Nil(t, set.SigningKey)
	assert.Equal(t, set.SigningAlgorithm, AlgorithmHS256)
	assert.NotNil(t, set.Claims)
	assert.Equal(t, set.TokenLookup, "header:"+headerAuthorization)
	assert.Equal(t, set.AuthScheme, "Bearer")
	assert.Equal(t, reflect.ValueOf(set.KeyFunc).Pointer(), reflect.ValueOf(set.defaultKeyFunc).Pointer())
	assert.Equal(t, reflect.ValueOf(set.ParseTokenFunc).Pointer(), reflect.ValueOf(set.defaultParseToken).Pointer())

	// with options
	skipper := func(*fiber.Ctx) bool {
		return false
	}
	claims := &fakeClaims{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	}
	parseToken := func(string, *fiber.Ctx) (*jwt.Token, error) { return nil, nil }
	tokenLookups := strings.Join([]string{
		"query:ut-query",
		"param:ut-param",
		"cookie:ut-cookie",
		"form:ut-form",
		"header:ut-header",
	}, ",")

	set = newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithSkipper(skipper),
		WithSigningKey("ut-signing-key"),
		WithSigningKeys("ut-key", "ut-value"),
		WithSigningAlgorithm("ut-signing-algorithm"),
		WithClaims(claims),
		WithTokenLookup(tokenLookups),
		WithAuthScheme("ut-auth-scheme"),
		WithKeyFunc(keyFunc),
		WithParseTokenFunc(parseToken),
		WithIgnorePrefix("/ut"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
	assert.False(t, set.Skipper(ctx))
	assert.Equal(t, "ut-signing-key", set.SigningKey)
	assert.NotEmpty(t, set.SigningKeys)
	assert.Equal(t, "ut-signing-algorithm", set.SigningAlgorithm)
	assert.Equal(t, claims, set.Claims)
	assert.Equal(t, tokenLookups, set.TokenLookup)
	assert.Len(t, set.extractors, 3)
	assert.Equal(t, "ut-auth-scheme", set.AuthScheme)
	assert.Equal(t, reflect.ValueOf(set.KeyFunc).Pointer(), reflect.ValueOf(keyFunc).Pointer())
	assert.Equal(t, reflect.ValueOf(set.ParseTokenFunc).Pointer(), reflect.ValueOf(parseToken).Pointer())
}

func TestJwtFromHeader(t *testing.T) {
	headerKey := "ut-header"
	authScheme := "ut-auth-scheme"
	jwtValue := "ut-jwt"
	extractor := jwtFromHeader(headerKey, authScheme)

	// happy case
	ctx := fiber.New().AcquireCtx(&fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	})

	ctx.Request().Header.Set(headerKey, strings.Join([]string{authScheme, jwtValue}, " "))
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx.Request().Header.Set(headerKey, strings.Join([]string{"invalid", jwtValue}, " "))
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func TestJwtFromQuery(t *testing.T) {
	queryKey := "ut-query"
	jwtValue := "ut-jwt"
	extractor := jwtFromQuery(queryKey)
	ctx := fiber.New().AcquireCtx(&fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	})

	// happy case
	ctx.Context().QueryArgs().Set(queryKey, jwtValue)
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx.Context().QueryArgs().Reset()
	ctx.Context().QueryArgs().Set("invalid", jwtValue)
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func TestJwtFromCookie(t *testing.T) {
	cookieKey := "ut-cookie"
	jwtValue := "ut-jwt"
	extractor := jwtFromCookie(cookieKey)
	ctx := fiber.New().AcquireCtx(&fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	})

	// happy case
	ctx.Request().Header.SetCookie(cookieKey, jwtValue)
	res, err := extractor(ctx)
	assert.Equal(t, jwtValue, res)
	assert.Nil(t, err)

	// invalid auth
	ctx.Request().Header.DelAllCookies()
	ctx.Request().Header.SetCookie("invalid", jwtValue)
	res, err = extractor(ctx)
	assert.Empty(t, res)
	assert.Nil(t, err)
}

type fakeClaims struct{}

func (c *fakeClaims) Valid() error {
	return nil
}
