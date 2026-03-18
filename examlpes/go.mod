module example

go 1.24.0

replace ritual => ../

require (
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/webview/webview_go v0.0.0-20240831120633-6173450d4dd6
	golang.org/x/crypto v0.48.0
	ritual v0.0.0-00010101000000-000000000000
)

require golang.org/x/sys v0.41.0 // indirect
