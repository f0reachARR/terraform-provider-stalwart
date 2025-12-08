resource "stalwart_dkim_signature" "example" {
  algorithm = "Ed25519"
  domain    = "example.org"
  selector  = "default"
}
