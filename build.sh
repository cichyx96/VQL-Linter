cp -r vql_linter velociraptor/
cp -r to_replace/file_store velociraptor/
cd velociraptor
mkdir velociraptor/vql_linter/definitions
cp -r artifacts/definitions vql_linter/definitions
export PATH=$PATH:~/go/bin

go run make.go -v linux
go build -o ../vql-linter vql_linter/*.go
GOOS=windows GOARCH=amd64 go build -o ../vql-linter.exe vql_linter/*.go
