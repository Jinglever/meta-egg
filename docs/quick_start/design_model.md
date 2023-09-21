# 数据模型设计

### 设计一个基本的实体表
    
```xml
<!-- 示例 -->
<table name="user" type="DATA" comment="用户" handler="true">
    <column name="id" type="_ID" order="true" list="true" />
    <column name="name" type="VARCHAR" length="64" comment="用户名" unique="true" list="true" alter="true" />
    <column name="gender" type="_FID" required="true" comment="性别" filter="true" list="true" alter="true">
        <foreign_key table="gender" />
    </column>
    <column name="age" type="TINYINT" unsigned="true" required="true" comment="年龄" alter="true" />
    <column name="is_on_job" type="_BOOL" required="true" insert_default="0" comment="是否在职" alter="true" filter="true" />
    <column name="birthday" type="DATE" required="true" comment="生日" alter="true" />
    <column name="created_by" type="_ME_CREATE" required="false" comment="创建者">
        <foreign_key table="user" />
    </column>
    <column name="created_at" type="_TIME_CREATE" required="true" comment="创建时间" />
    <column name="updated_by" type="_ME_UPDATE" required="false" comment="更新者">
        <foreign_key table="user" />
    </column>
    <column name="updated_at" type="_TIME_UPDATE" required="true" comment="更新时间" />
    <column name="deleted_by" type="_ME_DELETE" comment="删除者">
        <foreign_key table="user" />
    </column>
    <column name="deleted_at" type="_TIME_DELETE" comment="删除时间" />

    <index>
        <index_column name="is_on_job" />
        <index_column name="age" />
    </index>
</table>
```
    
- `table.type` 为 `DATA` 时，代表该表属于实体表，支持增删改查。

- `table.handler` 为true时，将会为该实体生成biz和handler代码，包括5个接口： 创建，获取详情，获取列表，更新，删除。

- `column.order` 为true时，生成的列表接口，支持按本字段排序，建议为这个字段创建索引。

- `column.list` 为true时，生成的列表接口，返回的实体字段里将包含本字段。

- `column.alter` 为true时，生成的创建和更新接口，支持传入该字段的值。

- `column.filter` 为true时，生成的列表接口，支持按本字段进行筛选。

- `column.type` 为 `_TIME_DELETE` ，将会为实体增加一个datetime字段，用于逻辑删除标记，字段值为null代表数据有效，不为null代表已逻辑删除，值为删除的时间。此字段在代码中的类型是 `gorm.DeletedAt` ，所以常规的gorm查询操作，无需特意判断该字段，gorm框架会自动补充判断逻辑，但特殊情况，比如 `join` 查询会导致逻辑删除判断失效，要自行添加其判断条件。（*ps. 建议尽可能不使用 join 查询*）

- `column.required` 为false时，代表字段default null，对应到代码中，该字段将会是一个指针。如非必要，建议尽量选择not null，在代码上操作起来会更加方便。但是在零值不适合代表不存在的场景下，也请大方使用default null。

- `column.comment` 强烈建议填入正确的字段含义，它将会出现在sql及代码注释里，这是一个良好的工程习惯。



### 设计一个常规元数据表
    
```xml
<!-- 示例 -->
<table name="gender" type="META" comment="性别">
    <column name="id" type="_ID" />
    <column name="sematic" type="_SMT2" length="8" required="true" comment="语义" />
    <column name="desc" type="_DESC" length="64" required="false" comment="描述" />
    <column name="deleted_at" type="_TIME_DELETE" comment="删除时间" />
</table>
```
    
- `table.type` 为 `META` 时，代表该表属于元数据表，仅支持查。数据的增删改，需要手工操作数据库，然后借助工具相关常量定义及注释。

- `column.type` 为 `_SMT2` 时，代表该字段值是元数据的唯一标识，请务必使用恰当的英文作为标识，它将应用到生成代码的元数据常量名中。

- `column.type` 为 `_DESC` 时，代表该字段值是元数据的描述，会应用到生成代码的元数据常量注释中。



### 数据模型设计的一些建议

- 首先要吃透业务逻辑，基于实体的本质去思考它们的数据模型，然后才是数据结构的设计技巧。


- 数据库设计存在三大范式：
    - 保证表中每个属性都保持原子性
    - 保证表中的非主属性与主键完全依赖
    - 保证表中的非主属性与主键不存在传递依赖


- 同时，对反范式设计保持开放和辩证的心态。比如，当冗余信息有价值或者能大幅提高查询效率时，我们可以适当采用反范式设计进行优化。