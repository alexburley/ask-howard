variable "database_url" {
  type    = string
  default = "postgres://ask-howard:ask-howard@localhost:5432/ask-howard?sslmode=disable"
}

env "local" {
  src = "file://schema.hcl"
  url = var.database_url
  dev = "docker://postgres/16-alpine/dev"
  migration {
    dir = "file://migrations"
  }
}
