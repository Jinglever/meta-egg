# protobuf规范

在 `proto` 目录里编写grpc的接口协议。

参考文档：https://developers.google.com/protocol-buffers/docs/proto3

- 关键词
    - syntax = "proto3"
        - 约定使用proto3的语法，如果没有这行，会默认当proto2看待
    - message
        - 定义一个结构体
        - 代码样例
            
            ```protobuf
            message SearchRequest {
              string query = 1;
              int32 page_number = 2;
              int32 result_per_page = 3;
            }
            ```
            
        - 结构体属性的格式一般为：<type> <name> = <number>;
            - type跟各种语言的对应关系，参考文档：`https://developers.google.com/protocol-buffers/docs/proto3#scalar`
            - number是一个unique number，唯一标识属性，用在Protocol Buffer Encoding，不要重用它
        - 一些常用的type：
            - double\float\int32\int64\uint32\uint64\bool\string\bytes
        - required\optional
            - 定义了optional的字段，拿Golang举例，生成的字段是一个指针类型
        - enum枚举类型
            - 代码样例
                
                ```protobuf
                message SearchRequest {
                  string query = 1;
                  int32 page_number = 2;
                  int32 result_per_page = 3;
                  enum Corpus {
                    UNIVERSAL = 0;
                    WEB = 1;
                    IMAGES = 2;
                    LOCAL = 3;
                    NEWS = 4;
                    PRODUCTS = 5;
                    VIDEO = 6;
                  }
                  Corpus corpus = 4;
                }
                ```
                
        - repeated数组类型
            - 代码样例
                
                ```protobuf
                message Result {
                  string url = 1;
                  string title = 2;
                  repeated string snippets = 3; // 代表[]string数组
                }
                ```
                
        - Oneof：两个或多个字段里只选其一
            - 代码样例
                
                ```protobuf
                message SampleMessage {
                  oneof test_oneof {
                    string name = 4;
                    SubMessage sub_message = 9;
                  }
                }
                ```
                
        - 支持message嵌套
            - 代码样例
                
                ```protobuf
                message SearchResponse {
                  message Result {
                    string url = 1;
                    string title = 2;
                    repeated string snippets = 3;
                  }
                  repeated Result results = 1;
                }
                message SomeOtherMessage {
                  SearchResponse.Result result = 1;
                }
                ```
                
    - service
        - 定义一个服务接口
        - 代码样例
            
            ```protobuf
            service SearchService {
              rpc Search(SearchRequest) returns (SearchResponse);
            }
            ```
            
    - Package
        - 为了避开命名冲突，用package实现命名空间管理，如果生成golang的pb，会生成go的package
        - 代码样例
            
            ```protobuf
            package foo.bar;
            message Open { ... }
            
            message Foo {
              ...
              foo.bar.Open open = 1;
              ...
            }
            ```
            
    - 引入别的proto
        - 代码样例
            
            ```protobuf
            import "myproject/other_protos.proto";
            ```