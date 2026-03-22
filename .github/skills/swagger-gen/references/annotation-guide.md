# Swaggo Annotation Guide

## General API Annotations (main.go)
| Annotation | Description |
|:---|:---|
| `@title` | API title |
| `@version` | API version |
| `@description` | API description |
| `@host` | Host (e.g., `localhost:8080`) |
| `@BasePath` | Base path (e.g., `/api/v1`) |
| `@securityDefinitions.apikey` | Security scheme name |

## Handler Annotations
| Annotation | Description | Example |
|:---|:---|:---|
| `@Summary` | Short description | `@Summary Get user by ID` |
| `@Description` | Detailed description | `@Description Returns a single user` |
| `@Tags` | Grouping tag | `@Tags users` |
| `@Accept` | Input MIME type | `@Accept json` |
| `@Produce` | Output MIME type | `@Produce json` |
| `@Param` | Parameter | `@Param id path string true "User ID"` |
| `@Success` | Success response | `@Success 200 {object} models.User` |
| `@Failure` | Error response | `@Failure 404 {object} models.ErrorResponse` |
| `@Security` | Auth requirement | `@Security BearerAuth` |
| `@Router` | Route + method | `@Router /users/{id} [get]` |

## @Param Format
```
@Param <name> <in> <type> <required> "<description>" [format/enums]
```
- `in`: `path`, `query`, `header`, `body`, `formData`
- `type`: `string`, `integer`, `boolean`, `object`
- `required`: `true` or `false`

## Body Parameter
```go
// @Param request body models.CreateUserRequest true "Create user request"
```

## Array Response
```go
// @Success 200 {array} models.User
```

## Pagination Query Params
```go
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
```
