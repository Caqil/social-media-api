// utils/jwt.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"social-media-api/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey            string
	RefreshSecretKey     string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	TokenType             string    `json:"token_type"`
	ExpiresIn             int64     `json:"expires_in"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

// Claims represents JWT token claims
type Claims struct {
	UserID     string          `json:"user_id"`
	Username   string          `json:"username"`
	Email      string          `json:"email"`
	Role       models.UserRole `json:"role"`
	SessionID  string          `json:"session_id"`
	TokenType  string          `json:"token_type"` // "access" or "refresh"
	DeviceInfo string          `json:"device_info,omitempty"`
	IPAddress  string          `json:"ip_address,omitempty"`
	IssuedAt   int64           `json:"iat"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	config JWTConfig
}

// NewJWTService creates a new JWT service instance
func NewJWTService() *JWTService {
	config := JWTConfig{
		SecretKey:            getEnvOrDefault("JWT_SECRET", "your-secret-key"),
		RefreshSecretKey:     getEnvOrDefault("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
		AccessTokenDuration:  AccessTokenExpiry,
		RefreshTokenDuration: RefreshTokenExpiry,
		Issuer:               getEnvOrDefault("JWT_ISSUER", AppName),
	}

	return &JWTService{config: config}
}

// GenerateTokenPair generates access and refresh token pair
func (j *JWTService) GenerateTokenPair(user *models.User, deviceInfo, ipAddress string) (*TokenPair, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	time.Now()

	// Generate access token
	accessToken, accessExpiry, err := j.generateToken(user, sessionID, "access", deviceInfo, ipAddress, j.config.AccessTokenDuration, j.config.SecretKey)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, refreshExpiry, err := j.generateToken(user, sessionID, "refresh", deviceInfo, ipAddress, j.config.RefreshTokenDuration, j.config.RefreshSecretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		ExpiresIn:             int64(j.config.AccessTokenDuration.Seconds()),
		AccessTokenExpiresAt:  accessExpiry,
		RefreshTokenExpiresAt: refreshExpiry,
	}, nil
}

// generateToken generates a JWT token
func (j *JWTService) generateToken(user *models.User, sessionID, tokenType, deviceInfo, ipAddress string, duration time.Duration, secret string) (string, time.Time, error) {
	now := time.Now()
	expiryTime := now.Add(duration)

	claims := Claims{
		UserID:     user.ID.Hex(),
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		SessionID:  sessionID,
		TokenType:  tokenType,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		IssuedAt:   now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   user.ID.Hex(),
			Issuer:    j.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiryTime, nil
}

// ValidateAccessToken validates an access token
func (j *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, j.config.SecretKey, "access")
}

// ValidateRefreshToken validates a refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, j.config.RefreshSecretKey, "refresh")
}

// validateToken validates a JWT token
func (j *JWTService) validateToken(tokenString, secret, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check token type
		if claims.TokenType != expectedType {
			return nil, fmt.Errorf("invalid token type")
		}

		// Check if token is expired
		if time.Now().Unix() > claims.ExpiresAt.Unix() {
			return nil, fmt.Errorf("token is expired")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshTokens refreshes access token using refresh token
func (j *JWTService) RefreshTokens(refreshTokenString string, user *models.User, deviceInfo, ipAddress string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Verify user matches token
	if claims.UserID != user.ID.Hex() {
		return nil, fmt.Errorf("token user mismatch")
	}

	// Generate new token pair
	return j.GenerateTokenPair(user, deviceInfo, ipAddress)
}

// ExtractClaimsFromToken extracts claims without validation (for debugging)
func (j *JWTService) ExtractClaimsFromToken(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid claims")
}

// IsTokenExpired checks if a token is expired without full validation
func (j *JWTService) IsTokenExpired(tokenString string) (bool, error) {
	claims, err := j.ExtractClaimsFromToken(tokenString)
	if err != nil {
		return true, err
	}

	return time.Now().Unix() > claims.ExpiresAt.Unix(), nil
}

// GetTokenTTL returns time to live for a token
func (j *JWTService) GetTokenTTL(tokenString string) (time.Duration, error) {
	claims, err := j.ExtractClaimsFromToken(tokenString)
	if err != nil {
		return 0, err
	}

	expiryTime := time.Unix(claims.ExpiresAt.Unix(), 0)
	ttl := time.Until(expiryTime)

	if ttl < 0 {
		return 0, nil // Token is expired
	}

	return ttl, nil
}

// RevokeToken adds token to blacklist (requires Redis or similar)
func (j *JWTService) RevokeToken(tokenString string) error {
	// Extract token ID for blacklisting
	claims, err := j.ExtractClaimsFromToken(tokenString)
	if err != nil {
		return err
	}

	// In a real implementation, you would add the token ID to a blacklist
	// stored in Redis or another fast storage system
	_ = claims.ID // Token ID for blacklisting

	// For now, just return success
	// TODO: Implement actual token blacklisting
	return nil
}

// IsTokenBlacklisted checks if token is blacklisted
func (j *JWTService) IsTokenBlacklisted(tokenString string) (bool, error) {
	// Extract token ID
	claims, err := j.ExtractClaimsFromToken(tokenString)
	if err != nil {
		return true, err
	}

	// In a real implementation, check if token ID is in blacklist
	_ = claims.ID // Token ID to check

	// For now, return false (not blacklisted)
	// TODO: Implement actual blacklist checking
	return false, nil
}

// GeneratePasswordResetToken generates a token for password reset
func (j *JWTService) GeneratePasswordResetToken(userID primitive.ObjectID, email string) (string, error) {
	now := time.Now()
	expiryTime := now.Add(PasswordResetExpiry)

	claims := Claims{
		UserID:   userID.Hex(),
		Email:    email,
		IssuedAt: now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.Hex(),
			Issuer:    j.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// ValidatePasswordResetToken validates a password reset token
func (j *JWTService) ValidatePasswordResetToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateEmailVerificationToken generates a token for email verification
func (j *JWTService) GenerateEmailVerificationToken(userID primitive.ObjectID, email string) (string, error) {
	now := time.Now()
	expiryTime := now.Add(EmailVerificationExpiry)

	claims := Claims{
		UserID:   userID.Hex(),
		Email:    email,
		IssuedAt: now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.Hex(),
			Issuer:    j.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// ValidateEmailVerificationToken validates an email verification token
func (j *JWTService) ValidateEmailVerificationToken(tokenString string) (*Claims, error) {
	return j.ValidatePasswordResetToken(tokenString) // Same validation logic
}

// GenerateAPIToken generates a long-lived API token
func (j *JWTService) GenerateAPIToken(userID primitive.ObjectID, apiKeyName string) (string, error) {
	now := time.Now()
	// API tokens are long-lived (1 year)
	expiryTime := now.Add(365 * 24 * time.Hour)

	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID:    userID.Hex(),
		SessionID: sessionID,
		TokenType: "api",
		IssuedAt:  now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   userID.Hex(),
			Issuer:    j.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// ValidateAPIToken validates an API token
func (j *JWTService) ValidateAPIToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, j.config.SecretKey, "api")
}

// Utility functions

// generateSessionID generates a unique session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetUserIDFromClaims extracts user ID from claims
func GetUserIDFromClaims(claims *Claims) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(claims.UserID)
}

// GetSessionIDFromClaims extracts session ID from claims
func GetSessionIDFromClaims(claims *Claims) string {
	return claims.SessionID
}

// IsAccessToken checks if claims represent an access token
func IsAccessToken(claims *Claims) bool {
	return claims.TokenType == "access"
}

// IsRefreshToken checks if claims represent a refresh token
func IsRefreshToken(claims *Claims) bool {
	return claims.TokenType == "refresh"
}

// IsAPIToken checks if claims represent an API token
func IsAPIToken(claims *Claims) bool {
	return claims.TokenType == "api"
}

// GetTokenAge returns the age of a token
func GetTokenAge(claims *Claims) time.Duration {
	issuedAt := time.Unix(claims.IssuedAt, 0)
	return time.Since(issuedAt)
}

// GetTokenExpiry returns when the token expires
func GetTokenExpiry(claims *Claims) time.Time {
	return time.Unix(claims.ExpiresAt.Unix(), 0)
}

// IsTokenNearExpiry checks if token expires within the given duration
func IsTokenNearExpiry(claims *Claims, threshold time.Duration) bool {
	expiryTime := GetTokenExpiry(claims)
	return time.Until(expiryTime) <= threshold
}
