# 工程初始化

### 创建项目
    
```bash
# 执行new命令
meta-egg new

# 之后根据提示设置基础信息，如下：
Please input project EN name: # 输入工程的英文名（多个单词推荐以中横线连接）
Please input project description: # 输入工程的描述，会出现在README中
Please input go module name: # 相当于于go mod init "xxx"
Please input go version [1.19]: # 输入go的版本，直接回车则默认1.19
Please select server type [GRPC/HTTP/ALL default:ALL]: # 选择服务类型
Do you need database? [y/n default:y]: # 选择工程是否用到数据库
Please select database type [MySQL/PostgreSQL default:MySQL]: # 选择数据库类型
Do you need table demo? [y/n default:y]: # 是否生成demo数据库表供参考

# 执行成功后会提示
Project generated successfully
```
    
### 环境准备

- 当前环境有版本匹配的 go，确保开启GO111MODULE
        
    ```bash
    go env -w GO111MODULE=on
    ```
        
- 安装protoc，比如对于x86_64的linux环境：
        
    ```bash
    wget https://github.com/protocolbuffers/protobuf/releases/download/v21.8/protoc-21.8-linux-x86_64.zip
    mkdir -p /usr/local/protoc
    sudo unzip protoc-21.8-linux-x86_64.zip -d /usr/local/protoc
    sudo ln -s /usr/local/protoc/bin/protoc /usr/local/bin/protoc
    ```
        
### 初始化工程

1. 安装依赖
        
    ```bash
    make init # 将安装protoc-gen-go, wire, mockgen等依赖
    ```
        
2. 根据proto生成api
        
    ```bash
    make pb
    ```
        
3. 拉取go依赖包
    
    ```bash
    go mod tidy
    ```
        
4. 生成wire和mock代码
    
    ```bash
    make generate # ps. 如果执行报错，那么先再执行go mod tidy，然后重新执行本命令
    ```
    
5. 创建本地配置文件
    
    ```bash
    cp _manifest/env.yml _manifest/env-local.yml
    
    cp configs/conf.yml configs/conf-local.yml
    ```
    
6. 假设前面选择了需要数据库，并且需要demo数据库表，那么接下来准备数据库服务
    1. 生成数据库schema
        
        ```bash
        meta-egg db -e _manifest/env-local.yml
        # 由于还没有真正创建数据库，所以这个命令会提示一个ERR表示数据库不存在，
        # 可以忽略此报错，我们需要的schema.sql已经被生成出来了，
        # 在_manifest/sql/下可以找到。
        ```
        
    2. 在数据库中使用上一步生成的 `schema.sql` 创建数据库
        
        ```bash
        # ps. 推荐使用GUI工具接入数据库，在GUI工具中执行schema.sql去创建。
        # 因为有些docker起的mysql，登入容器用cli来导入sql，会由于容器环境缺少中文支持，导致创建的数据库表comment和字段comment未能正确写入。
        ```
        
    3. 将数据库地址、账号密码及数据库表名，填入 `_manifest/env-local.yml` 和 `configs/conf-local.yml` 的相应配置项内
    4. 由于demo数据库表 `gener` 是元数据库表，我们应该在数据库里手工添加元数据
        
        ```bash
        INSERT INTO `gender` (`deleted_at`,`desc`,`id`,`sematic`) VALUES
        (NULL, '男性', '1', 'MALE'),
        (NULL, '女性', '2', 'FEMALE');
        ```
        
    5. 再次执行 `db` 命令，检验数据库结构是否对齐，以及生成元数据初始化sql `meta-data.sql` 
        
        ```bash
        meta-egg db -e _manifest/env-local.yml
        
        # 这里应该不会再有报错，结果应该如下所示
        database exists, will generate inc.sql
        generate db sql success
        copy schema.sql and meta-data.sql to <proj_root>/sql/
            schema.sql [✓]
            meta-data.sql [✓]
        
        # 其中inc.sql里，应该不存在操作数据库变更的sql
        ```
        
    6. 执行 `update` 命令，根据前面创建的元数据生成代码，增加对元数据的常量定义
        
        ```bash
        meta-egg update -e _manifest/env-local.yml
        
        # 结果应该会更新gender的model和repo，如下：
        replaced files in gen/model
            gender.go [✓]
        replaced files in gen/repo
            gender.go [✓]
        ```
        
7. 编译和运行工程
    
    ```bash
    make build; make run
    
    # 应该看到服务被正常启动
    ```