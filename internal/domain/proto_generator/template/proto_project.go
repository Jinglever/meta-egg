package template

var TplProtoDataTableMessage = `// %%TABLE-NAME-STRUCT%%Detail %%TABLE-COMMENT%%详情
message %%TABLE-NAME-STRUCT%%Detail {
%%COL-LIST-IN-VO%%%%RL-FIELDS-IN-DETAIL%%}

// Create%%TABLE-NAME-STRUCT%%Request 创建%%TABLE-COMMENT%%请求
message Create%%TABLE-NAME-STRUCT%%Request {
%%COL-LIST-FOR-CREATE%%%%RL-FIELDS-IN-CREATE%%}

// Get%%TABLE-NAME-STRUCT%%DetailRequest 获取%%TABLE-COMMENT%%详情请求
message Get%%TABLE-NAME-STRUCT%%DetailRequest {
    // %%TABLE-COMMENT%%ID
    uint64 id = 1 [(validate.rules).uint64 = {gte: 1}];
}

// %%TABLE-NAME-STRUCT%%ListInfo %%TABLE-COMMENT%%列表信息
message %%TABLE-NAME-STRUCT%%ListInfo {
%%COL-LIST-FOR-LIST%%%%RL-FIELDS-IN-LIST%%}

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

%%RL-MESSAGES%%`

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
%%RL-HANDLER-FUNCTIONS%%`

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

// RL表相关的消息模板
var TplRLTableMessage = `
// %%RL-TABLE-NAME-STRUCT%%Detail %%RL-TABLE-COMMENT%%详情
message %%RL-TABLE-NAME-STRUCT%%Detail {
%%COL-LIST-IN-VO%%}

// %%RL-TABLE-NAME-STRUCT%%ListInfo %%RL-TABLE-COMMENT%%列表信息
message %%RL-TABLE-NAME-STRUCT%%ListInfo {
%%COL-LIST-FOR-LIST%%}

// %%RL-TABLE-NAME-STRUCT%%CreateData %%RL-TABLE-COMMENT%%创建数据
message %%RL-TABLE-NAME-STRUCT%%CreateData {
%%COL-LIST-FOR-CREATE%%}

// Add%%RL-TABLE-NAME-STRUCT%%Request 添加%%RL-TABLE-COMMENT%%请求
message Add%%RL-TABLE-NAME-STRUCT%%Request {
    // %%MAIN-TABLE-COMMENT%%ID
    uint64 %%MAIN-TABLE-NAME-LOWER%%_id = 1 [(validate.rules).uint64 = {gte: 1}];
%%COL-LIST-FOR-CREATE%%}

// Remove%%RL-TABLE-NAME-STRUCT%%Request 删除%%RL-TABLE-COMMENT%%请求
message Remove%%RL-TABLE-NAME-STRUCT%%Request {
    // %%MAIN-TABLE-COMMENT%%ID
    uint64 %%MAIN-TABLE-NAME-LOWER%%_id = 1 [(validate.rules).uint64 = {gte: 1}];
    // %%RL-TABLE-COMMENT%%ID
    uint64 %%RL-TABLE-NAME-LOWER%%_id = 2 [(validate.rules).uint64 = {gte: 1}];
}

// GetAll%%RL-TABLE-NAME-STRUCT%%Request 获取所有%%RL-TABLE-COMMENT%%请求
message GetAll%%RL-TABLE-NAME-STRUCT%%Request {
    // %%MAIN-TABLE-COMMENT%%ID
    uint64 %%MAIN-TABLE-NAME-LOWER%%_id = 1 [(validate.rules).uint64 = {gte: 1}];
}

// GetAll%%RL-TABLE-NAME-STRUCT%%Response 获取所有%%RL-TABLE-COMMENT%%响应
message GetAll%%RL-TABLE-NAME-STRUCT%%Response {
    repeated %%RL-TABLE-NAME-STRUCT%%Detail list = 1; // %%RL-TABLE-COMMENT%%列表
}
`

// RL表相关的gRPC方法模板
var TplRLTableHandlerFuncs = `    // 添加%%RL-TABLE-COMMENT%%
    rpc Add%%RL-TABLE-NAME-STRUCT%% (Add%%RL-TABLE-NAME-STRUCT%%Request) returns (%%RL-TABLE-NAME-STRUCT%%Detail) {}
    // 删除%%RL-TABLE-COMMENT%%
    rpc Remove%%RL-TABLE-NAME-STRUCT%% (Remove%%RL-TABLE-NAME-STRUCT%%Request) returns (google.protobuf.Empty) {}
    // 获取所有%%RL-TABLE-COMMENT%%
    rpc GetAll%%RL-TABLE-NAME-STRUCT%% (GetAll%%RL-TABLE-NAME-STRUCT%%Request) returns (GetAll%%RL-TABLE-NAME-STRUCT%%Response) {}
`
