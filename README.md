JWT Authentication in Go

A full JWT-based authentication system in Go with login, access tokens, refresh tokens, rotation, and logout designed for stateless, scalable applications. 
Includes secure cookie handling for refresh tokens.

View the GitHub Repository

Features
-Login with email and password
-Short-lived access tokens (15 min)
-Long-lived refresh tokens (7 days) stored securely in DB
-Refresh token rotation to prevent reuse
-Max refresh token lifetime enforcement (30 days)
-Logout that revokes refresh tokens
-Cookie-based refresh tokens (HttpOnly + SameSite=Lax)

Full stateless access token validation
Tech Stack
Language: Go
Database: PostgreSQL (with SQLC)
Hashing: bcrypt
JWT Library: github.com/golang-jwt/jwt
Cookie Security: HttpOnly, SameSite=Lax
