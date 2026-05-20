variable "database_url" {
  type    = string
  default = "postgres://ask-howard:ask-howard@localhost:5432/ask-howard?sslmode=disable"
}

env "local" {
  src = "file://internal/adapter/outbound/postgres/schema.hcl"
  url = var.database_url
  dev = "postgres://ask-howard:ask-howard@postgres:5432/ask-howard-dev?sslmode=disable"
  migration {
    dir = "file://internal/adapter/outbound/postgres/migrations"
  }
}
