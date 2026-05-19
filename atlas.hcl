variable "database_url" {
  type    = string
  default = "postgres://pulse:pulse@localhost:5432/pulse?sslmode=disable"
}

env "local" {
  src = "file://schema.hcl"
  url = var.database_url
  dev = "docker://postgres/16-alpine/dev"
  migration {
    dir = "file://migrations"
  }
}
