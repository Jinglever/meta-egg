package template

var TplReadme = `# %%PROJECT-NAME%%
%%PROJECT-DESC%%

## Usage

### Prepare
Install protoc:
` + "```" + `bash
wget https://github.com/protocolbuffers/protobuf/releases/download/v21.8/protoc-21.8-linux-x86_64.zip
mkdir -p /usr/local/protoc
sudo unzip protoc-21.8-linux-x86_64.zip -d /usr/local/protoc
sudo ln -s /usr/local/protoc/bin/protoc /usr/local/bin/protoc
` + "```" + `

### Init
` + "```" + `bash
make init
` + "```" + `

### Generate pb/swag/wire/mock
` + "```" + `bash
make generate
` + "```" + `
P.S. Maybe you need to run ` + "`" + `go mod tidy` + "`" + ` several times.

### Build
` + "```" + `bash
make build
` + "```" + `

### Run
` + "```" + `bash
make run
` + "```" + `
P.S. You need to copy and modify ` + "`" + `configs/<project-name>.yml` + "`" + ` to ` + "`" + `configs/<project-name>-local.yml` + "`" + ` before run.

## Project Structure
` + "```" + `sql
.
├── api  --------------------------  API Generated by protoc
├── build
│   └── bin  ----------------------  Binary Generated by go build
├── cmd  --------------------------  Main Entry
├── configs  ----------------------  Configs such as <project-name>.yml
├── docs  -------------------------  Documents such as swagger
├── gen
│   ├── model  --------------------  Model Generated by meta-egg
│   └── repo  ---------------------  Repo Generated by meta-egg
├── go.mod
├── go.sum
├── internal
│   ├── biz  ----------------------  Business Logic, file name is basically the same as table name
│   ├── common
│   │   ├── cerror  ---------------  Custom Error, feel free to define your own error
│   │   ├── constraint  -----------  Constraint, such as timezone, time format, etc.
│   │   ├── contexts  -------------  Contexts, such as session, trace, etc.
│   │   └── resource  -------------  Resource, such as db, redis, grpc client, etc.
│   ├── config  -------------------  golang struct for config file, function to load config file, etc.
│   ├── handler
│   │   ├── grpc  -----------------  GRPC Handler, file name is basically the same as table name
│   │   └── http  -----------------  HTTP Handler, file name is basically the same as table name
│   ├── repo  ---------------------  Repo Interface, file name is basically the same as table name
│   ├── server
│   │   ├── grpc  -----------------  GRPC Server, with middleware
│   │   └── http  -----------------  HTTP Server, with middleware, router, etc.
│   └── usecase  ------------------  UseCase, something quite like biz, but more complex so that it needs to be separated
├── LICENSE
├── Makefile
├── _manifest  --------------------  Manifest for meta-egg
├── pkg  --------------------------  Common Packages
├── proto  ------------------------  Protobuf Files, including xx_error.proto for custom error
├── README.md
├── sql  --------------------------  SQL Files, including schema.sql for table schema
└── third_party  ------------------  Third Party, such as proto files, sdk for grpc client, etc.
` + "```" + `

## Constraint
layers from top to bottom:
- cmd
- config
- server
- handler
- usecase
- biz
- repo
- model
- common

You should not import packages from bottom to top, such as import ` + "`" + `repo` + "`" + ` in ` + "`" + `model` + "`" + `, import ` + "`" + `biz` + "`" + ` in ` + "`" + `repo` + "`" + `, etc.

## Development
### Modify DB Schema
1. Modify manifest file ` + "`" + `_manifest/<project-name>.xml` + "`" + `
2. Run ` + "`" + `meta-egg db` + "`" + ` to generate ` + "`" + `_manifest/sql/schema.sql` + "`" + ` and ` + "`" + `_manifest/sql/inc.sql` + "`" + `. If ` + "`" + `inc.sql` + "`" + ` is not empty, you need to run it manually in your database. Then you should copy ` + "`" + `schema.sql` + "`" + ` to ` + "`" + `sql/schema.sql` + "`" + ` to maintain the latest schema.
3. Run ` + "`" + `meta-egg update` + "`" + ` to update ` + "`" + `gen/model` + "`" + ` and ` + "`" + `gen/repo` + "`" + `
4. if new file in ` + "`" + `repo` + "`" + `, ` + "`" + `biz` + "`" + `, ` + "`" + `usecase` + "`" + `, maybe you need to modify ` + "`" + `ProviderSet` + "`" + ` in ` + "`" + `base.go` + "`" + ` in their folder.
5. Run ` + "`" + `make generate` + "`" + ` to update ` + "`" + `api` + "`" + `, ` + "`" + `docs` + "`" + `, ` + "`" + `wire_gen.go` + "`" + `, ` + "`" + `repo/mock` + "`" + `, etc.

### Modify API
1. Modify proto file ` + "`" + `proto/<project-name>.proto` + "`" + `
2. Run ` + "`" + `make pb` + "`" + ` to update ` + "`" + `api` + "`" + `
3. Or, for http, you can just modify ` + "`" + `handler/http/<table-name>.go` + "`" + ` and ` + "`" + `server/http/router.go` + "`" + ` to add new api.
4. Then Run ` + "`" + `make swag` + "`" + ` to update ` + "`" + `docs` + "`" + `

<sub><sup>Special thanks to @ZZH @WZQ @GTJ @HQ for their contributions to the framework design.</sup></sub>
`
