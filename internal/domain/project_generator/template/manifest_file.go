package template

var TplManifestTableDemo = `
        <!-- 实体表示例 -->
        <table name="user" type="DATA" comment="用户" handler="true">
            <column name="id" type="_ID" order="true" list="true" />
            <column name="name" type="VARCHAR" length="64" comment="用户名" unique="true" list="true" alter="true" />
            <column name="gender" type="_FID" required="true" comment="性别" filter="true" list="true" alter="true">
                <foreign_key table="mt_gender" />
            </column>
            <column name="age" type="TINYINT" unsigned="true" required="true" comment="年龄" alter="true" />
            <column name="is_on_job" type="_BOOL" required="true" insert_default="0" comment="是否在职" alter="true" filter="true" />
            <column name="birthday" type="DATE" required="false" comment="生日" alter="true" />
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
            <column name="deleted_at" type="_TIME_DELETE2" comment="删除时间" />

            <index>
                <index_column name="is_on_job" />
                <index_column name="age" />
            </index>
        </table>
        <!-- 元数据表示例 -->
        <table name="mt_gender" type="META" comment="性别" handler="true">
            <column name="id" type="_ID" />
            <column name="semantic" type="_SMT2" length="8" required="true" comment="语义" />
            <column name="desc" type="_DESC" length="64" required="false" comment="描述" />
            <column name="deleted_at" type="_TIME_DELETE2" comment="删除时间" />
        </table>
`

var TplManifestDatabase = `
    <database name="%%PROJECT-NAME%%" type="%%DB-TYPE%%" charset="%%DB-CHARSET%%">
%%TPL-MANIFEST-TABLE-DEMO%%
    </database>
`

var TplManifestFile = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE project SYSTEM "meta_egg.dtd">

<project name="%%PROJECT-NAME%%" desc="%%PROJECT-DESC%%" go_module="%%GO-MODULE%%" go_version="%%GO-VERSION%%" server_type="%%SERVER-TYPE%%" no_auth="%%NO-AUTH%%">
    <usecases>
        <usecase name="login" desc="登录模块" />
    </usecases>
%%TPL-MANIFEST-DATABASE%%
</project>
`
