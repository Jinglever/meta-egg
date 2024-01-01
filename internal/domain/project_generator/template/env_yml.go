package template

var TplEnvYml = `project:
  # 工程代码根目录的绝对路径
  root: "%%PROJECT-ROOT%%"
manifest:
  # 工程的manifest目录的绝对路径
  root: "%%MANIFEST-ROOT%%"
  # 指定工程的manifest文件的绝对路径
  file: "%%MANIFEST-FILE%%"

# 数据库信息
db:
  host: ""
  port: ""
  user: ""
  password: ""
  db_name: "" # 数据库名, 优先级高于_manifest/xxx.xml文件里的database.name

# 被忽略的文件, 在update --uncertain操作里不会询问是否覆盖, 而会直接跳过
ignore_files:
  - ".gitignore"
  - "go.mod"
  - "go.sum"
  - "Makefile"
  - "README.md"
  - "*_gen.go"

# 被忽略的表, 不会因为xml里面没有这个表而生成删除该表的inc sql
ignore_tables:
  - "_xxx_"
`
