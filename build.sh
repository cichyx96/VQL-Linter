cp -r vql_linter velociraptor/
cd velociraptor
#export PATH=$PATH:~/go/bin

#go run make.go -v linux
go build -o ../vql-linter vql_linter/*.go


