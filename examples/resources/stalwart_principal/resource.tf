resource "stalwart_principal" "example_domain" {
  type        = "domain"
  name        = "example.org"
  description = "Example domain"
  quota       = 0
}

resource "stalwart_principal" "example_user" {
  type        = "individual"
  name        = "john"
  description = "John Doe"
  quota       = 10737418240 # 10GB
  emails      = ["john@example.org"]
  roles       = ["user"]
  lists       = ["all"]
}

resource "stalwart_principal" "example_group" {
  type        = "group"
  name        = "engineering"
  description = "Engineering Team"
  members     = ["john", "jane"]
}
