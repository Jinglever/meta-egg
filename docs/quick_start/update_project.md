# 工程迭代

随着产品的迭代，当需要变更数据库结构的时候，可参考下面的步骤，借助工具快速完成部分开发工作。

1. 编辑数据模型定义文件 `_manifest/<project-name>.xml` 
    
    ```bash
    # 这个文件里的语法，可参见 _manifest/meta_egg.dtd
    # ps. 如果给IDE装上了xml的插件，编写xml时会自动根据dtd的语法提供代码补全功能，比如对于vscode，推荐redhat的XML插件。
    
    # 在数据模型定义文件 _manifest/<project-name>.xml 里，可以增删改数据库表及字段，也可以变更个别跟业务层代码相关的属性，比如是否创建handler，字段是否可见，字段是否可筛选等（建议参考dtd进行探索~）
    ```
    
2. 生成增量sql
    
    ```bash
    meta-egg db -e _manifest/env-local.yml
    
    # 如果数据模型需要变更的，应该可以看到生成的 _manifest/sql/inc.sql 文件里有了增量sql
    
    # 请人工审查增量sql，慎重地拷贝到数据库内执行
    
    # 然后再次执行上面的 db 命令，直至 inc.sql 内不再出现操作数据库变更的sql，代表数据库结构已经跟数据模型定义保持了一致
    ```
    
3. 生成代码
    
    ```bash
    meta-egg update -e _manifest/env-local.yml --uncertain
    
    # 其中选项 --uncertain，会告诉工具自动对比gen目录之外的代码，将存在差异的文件向你展示，由你来选择是否采用生成的文件来新增或覆盖到当前工程里。
    
    # ps. 如果删除了数据表，假设原先存在关于这个数据表的文件，出于谨慎考虑，工具不会主动去删除它们，需要你自行判断和手工删除多余的文件
    ```
    
4. 习惯性执行Makefile里的 `generate` 命令，刷新框架里的生成代码
    
    ```bash
    make generate
    ```