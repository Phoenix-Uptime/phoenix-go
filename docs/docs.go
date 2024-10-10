// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Check if the API is healthy",
                "tags": [
                    "Health"
                ],
                "summary": "Health check",
                "responses": {
                    "200": {
                        "description": "status and message",
                        "schema": {
                            "$ref": "#/definitions/api.HealthCheckResponse"
                        }
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "Logs in a user with username and password, returning an API key if successful.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "User Login",
                "parameters": [
                    {
                        "description": "User login data",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "user successfully logged in",
                        "schema": {
                            "$ref": "#/definitions/api.LoginResponse"
                        }
                    },
                    "400": {
                        "description": "invalid login payload",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "invalid credentials",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "internal server error",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/signup": {
            "post": {
                "description": "Allows a new user to sign up, but only one user can exist.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "User Signup",
                "parameters": [
                    {
                        "description": "User signup data",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.SignupRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "User successfully created",
                        "schema": {
                            "$ref": "#/definitions/api.SignupResponse"
                        }
                    },
                    "400": {
                        "description": "Invalid request payload",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Signup not allowed if a user already exists",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal server error",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.ErrorResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "api.HealthCheckResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "api.LoginRequest": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string",
                    "minLength": 6
                },
                "username": {
                    "type": "string",
                    "maxLength": 20,
                    "minLength": 3
                }
            }
        },
        "api.LoginResponse": {
            "type": "object",
            "properties": {
                "api_key": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "api.SignupRequest": {
            "type": "object",
            "required": [
                "email",
                "password",
                "username"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "minLength": 6
                },
                "username": {
                    "description": "Username must be between 3 to 20 characters long, allowing only alphanumeric characters,\nunderscores, and hyphens. It is required and should be unique.",
                    "type": "string",
                    "maxLength": 20,
                    "minLength": 3
                }
            }
        },
        "api.SignupResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyHeader": {
            "type": "apiKey",
            "name": "x-api-key",
            "in": "header"
        },
        "ApiKeyQuery": {
            "type": "apiKey",
            "name": "api_key",
            "in": "query"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "127.0.0.1:3031",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "PhoenixUptime API",
	Description:      "PhoenixUptime Backend API",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
