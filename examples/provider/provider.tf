terraform {
  required_providers {
    stalwart = {
      source = "f0reachARR/stalwart"
    }
  }
}

provider "stalwart" {
  endpoint = "https://mail.example.org/api"
  username = "admin"
  password = "your-password"
}
