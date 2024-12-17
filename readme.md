go test -v -coverpkg=./... ./test/ -coverprofile=profile.cov
go tool cover -func profile.cov