Ecom API

A production-ready e-commerce REST API built in Go with a clean layered architecture — handler, service, and repository layers with full unit test coverage.

Features
Auth
- JWT-based authentication with short-lived access tokens (15 min)
- Refresh token rotation with max lifetime enforcement (30 days)
- HttpOnly cookie-based refresh tokens
- Email verification on signup
- Password reset flow with one-time tokens
- Logout with token revocation

E-commerce
- Product and variant management(pending)
- Order and order item processing(pending)
- Shipping rules engine(pending)
- Brand and category management(pending)

Tech Stack
- Language: Go
- Database: PostgreSQL via SQLC
- Auth: bcrypt, github.com/golang-jwt/jwt
- Testing: testify/mock

Architecture

Clean separation of concerns across three layers — handlers handle HTTP, services own business logic, repositories own database access. Interfaces between layers enable full unit testability without a live database.
