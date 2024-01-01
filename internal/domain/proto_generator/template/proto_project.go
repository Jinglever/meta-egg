package template

var TplProtoDataTableMessage = `// %%TABLE-NAME-STRUCT%%Detail %%TABLE-COMMENT%%详情
message %%TABLE-NAME-STRUCT%%Detail {
%%COL-LIST-IN-VO%%}

// Create%%TABLE-NAME-STRUCT%%Request 创建%%TABLE-COMMENT%%请求
message Create%%TABLE-NAME-STRUCT%%Request {
%%COL-LIST-FOR-CREATE%%}

// Get%%TABLE-NAME-STRUCT%%DetailRequest 获取%%TABLE-COMMENT%%详情请求
message Get%%TABLE-NAME-STRUCT%%DetailRequest {
    // %%TABLE-COMMENT%%ID
    uint64 id = 1 [(validate.rules).uint64 = {gte: 1}];
}

// %%TABLE-NAME-STRUCT%%ListInfo %%TABLE-COMMENT%%列表信息
message %%TABLE-NAME-STRUCT%%ListInfo {
%%COL-LIST-FOR-LIST%%}

// Get%%TABLE-NAME-STRUCT%%ListRequest 获取%%TABLE-COMMENT%%列表请求
message Get%%TABLE-NAME-STRUCT%%ListRequest {
    // 分页请求（可选, 不传则不分页）
    optional Pagination pagination = 1;
%%COL-LIST-FOR-FILTER%%%%COL-LIST-FOR-ORDER%%}

// Get%%TABLE-NAME-STRUCT%%ListResponse 获取%%TABLE-COMMENT%%列表响应
message Get%%TABLE-NAME-STRUCT%%ListResponse {
    repeated %%TABLE-NAME-STRUCT%%ListInfo list = 1; // 列表数据
    int64 total = 2; // 结果集总数
}

// Update%%TABLE-NAME-STRUCT%%Request 更新%%TABLE-COMMENT%%请求
message Update%%TABLE-NAME-STRUCT%%Request {
    // %%TABLE-COMMENT%%ID
    uint64 id = 1 [(validate.rules).uint64 = {gte: 1}];
%%COL-LIST-FOR-UPDATE%%}

// Delete%%TABLE-NAME-STRUCT%%Request 删除%%TABLE-COMMENT%%请求
message Delete%%TABLE-NAME-STRUCT%%Request {
    // %%TABLE-COMMENT%%ID
    uint64 id = 1 [(validate.rules).uint64 = {gte: 1}];
}
`

var TplProtoMetaTableMessage = `// %%TABLE-NAME-STRUCT%%Detail %%TABLE-COMMENT%%详情
message %%TABLE-NAME-STRUCT%%Detail {
%%COL-LIST-IN-VO%%}

// Get%%TABLE-NAME-STRUCT%%DetailRequest 获取%%TABLE-COMMENT%%详情请求
message Get%%TABLE-NAME-STRUCT%%DetailRequest {
    // %%TABLE-COMMENT%%ID
    uint64 id = 1 [(validate.rules).uint64 = {gte: 1}];
}

// %%TABLE-NAME-STRUCT%%ListInfo %%TABLE-COMMENT%%列表信息
message %%TABLE-NAME-STRUCT%%ListInfo {
%%COL-LIST-FOR-LIST%%}

// Get%%TABLE-NAME-STRUCT%%ListRequest 获取%%TABLE-COMMENT%%列表请求
message Get%%TABLE-NAME-STRUCT%%ListRequest {
    // 分页请求（可选, 不传则不分页）
    optional Pagination pagination = 1;
%%COL-LIST-FOR-FILTER%%%%COL-LIST-FOR-ORDER%%}

// Get%%TABLE-NAME-STRUCT%%ListResponse 获取%%TABLE-COMMENT%%列表响应
message Get%%TABLE-NAME-STRUCT%%ListResponse {
    repeated %%TABLE-NAME-STRUCT%%ListInfo list = 1; // 列表数据
    int64 total = 2; // 结果集总数
}
`

var TplProtoDataTableHandlerFuncs = `    // 创建%%TABLE-COMMENT%%
    rpc Create%%TABLE-NAME-STRUCT%% (Create%%TABLE-NAME-STRUCT%%Request) returns (%%TABLE-NAME-STRUCT%%Detail) {}
    // 获取%%TABLE-COMMENT%%详情
    rpc Get%%TABLE-NAME-STRUCT%%Detail (Get%%TABLE-NAME-STRUCT%%DetailRequest) returns (%%TABLE-NAME-STRUCT%%Detail) {}
    // 获取%%TABLE-COMMENT%%列表
    rpc Get%%TABLE-NAME-STRUCT%%List (Get%%TABLE-NAME-STRUCT%%ListRequest) returns (Get%%TABLE-NAME-STRUCT%%ListResponse) {}
    // 更新%%TABLE-COMMENT%%
    rpc Update%%TABLE-NAME-STRUCT%% (Update%%TABLE-NAME-STRUCT%%Request) returns (google.protobuf.Empty) {}
    // 删除%%TABLE-COMMENT%%
    rpc Delete%%TABLE-NAME-STRUCT%% (Delete%%TABLE-NAME-STRUCT%%Request) returns (google.protobuf.Empty) {}
`

var TplProtoMetaTableHandlerFuncs = `    // 获取%%TABLE-COMMENT%%详情
    rpc Get%%TABLE-NAME-STRUCT%%Detail (Get%%TABLE-NAME-STRUCT%%DetailRequest) returns (%%TABLE-NAME-STRUCT%%Detail) {}
    // 获取%%TABLE-COMMENT%%列表
    rpc Get%%TABLE-NAME-STRUCT%%List (Get%%TABLE-NAME-STRUCT%%ListRequest) returns (Get%%TABLE-NAME-STRUCT%%ListResponse) {}
`

var TplProtoProject string = `syntax = "proto3";
package %%PROJECT-NAME-PKG%%;
import "google/protobuf/empty.proto";
%%IMPORT-TABLE-PROTO%%%%IMPORT-FOR-MESSAGE%%option go_package = "./%%PROJECT-NAME-DIR%%";

%%HANDLER-MESSAGE%%

service %%PROJECT-NAME-STRUCT%% {
%%HANDLER-FUNCTION%%
    // TODO: add your service here
}
`

var TplProtoCommon string = `syntax = "proto3";
package %%PROJECT-NAME-PKG%%;
import "github.com/envoyproxy/protoc-gen-validate/validate/validate.proto";
option go_package = "./%%PROJECT-NAME-DIR%%";

// 分页请求
message Pagination {
    int32 page = 1 [(validate.rules).int32 = {gt: 0}];
    int32 page_size = 2 [(validate.rules).int32 = {gt: 0, lt: 100}];
}
`

var TplProtoTable string = `syntax = "proto3";
package %%PROJECT-NAME-PKG%%;
%%IMPORT-FOR-MESSAGE%%option go_package = "./%%PROJECT-NAME-DIR%%";

%%HANDLER-MESSAGE%%
`

var TplProtoImportForMessage string = `import "github.com/envoyproxy/protoc-gen-validate/validate/validate.proto";
import "%%PROJECT-NAME-DIR%%_common.proto";
`
