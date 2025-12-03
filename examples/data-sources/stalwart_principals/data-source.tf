data "stalwart_principals" "all" {
}

data "stalwart_principals" "users" {
  types = "individual"
}

output "all_principals" {
  value = data.stalwart_principals.all.principals
}

output "users_only" {
  value = data.stalwart_principals.users.principals
}
